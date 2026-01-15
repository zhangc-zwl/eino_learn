package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino-ext/components/embedding/dashscope"
	ri "github.com/cloudwego/eino-ext/components/indexer/redis"
	"github.com/cloudwego/eino-ext/components/model/openai"
	rr "github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/redis/go-redis/v9"
)

const (
	redisIndexName     = "test_index"
	redisKeyPrefix     = "test_doc:"
	knowledgeFilePath  = "./knowledge_base.txt"
	embeddingModelName = "text-embedding-v4"
	modelKey           = "sk-25cfec2f986a4cf1bf4187493a268d11"
	modelBaseUrl       = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	chatModelName      = "qwen3-32b"
)

var dim = 1024

func main() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:          "36.151.150.11:6379",
		UnstableResp3: true,
		Protocol:      2,
	})
	if err := prepareKnowledge(ctx, rdb); err != nil {
		panic(err)
	}

	userQuestion := "Eino 框架是什么？它的吉祥物是什么？"
	if err := answerQuestion(ctx, rdb, userQuestion); err != nil {
		panic(err)
	}

}

func prepareKnowledge(ctx context.Context, client *redis.Client) error {
	if err := createRedisIndexIfNotExist(ctx, client); err != nil {
		return err
	}
	loader, err := file.NewFileLoader(ctx, &file.FileLoaderConfig{
		UseNameAsID: true,
	})
	if err != nil {
		return err
	}
	docs, err := loader.Load(ctx, document.Source{
		URI: knowledgeFilePath,
	})
	if err != nil {
		return err
	}
	splitter, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   200,
		OverlapSize: 20,
		IDGenerator: func(ctx context.Context, originalID string, splitIndex int) string {
			return fmt.Sprintf("%s_%d", originalID, splitIndex)
		},
	})
	if err != nil {
		return err
	}
	splitDocs, err := splitter.Transform(ctx, docs)
	if err != nil {
		return err
	}

	embedder, err := dashscope.NewEmbedder(ctx, &dashscope.EmbeddingConfig{
		APIKey:     modelKey,
		Model:      embeddingModelName,
		Dimensions: &dim,
	})
	if err != nil {
		return err
	}
	indexer, err := ri.NewIndexer(ctx, &ri.IndexerConfig{
		Client:    client,
		KeyPrefix: redisKeyPrefix,
		Embedding: embedder,
	})
	if err != nil {
		return err
	}
	for _, v := range splitDocs {
		fmt.Println(v.Content)
		fmt.Println("*************************")
	}
	ids, err := indexer.Store(ctx, splitDocs)
	if err != nil {
		return err
	}
	fmt.Printf("✅ Indexed %d documents\n", len(ids))
	fmt.Printf("%v", ids)
	return nil
}

func createRedisIndexIfNotExist(ctx context.Context, client *redis.Client) error {
	index, err := client.Do(ctx, "FT._LIST").StringSlice()
	if err != nil {
		return err
	}
	for _, v := range index {
		if v == redisIndexName {
			return nil
		}
	}
	_, err = client.FTCreate(ctx, redisIndexName, &redis.FTCreateOptions{
		OnHash: true,
		Prefix: []any{redisKeyPrefix},
	}, &redis.FieldSchema{FieldName: "content", FieldType: redis.SearchFieldTypeText},
		&redis.FieldSchema{FieldName: "vector_content",
			FieldType: redis.SearchFieldTypeVector,
			VectorArgs: &redis.FTVectorArgs{
				FlatOptions: &redis.FTFlatOptions{
					Type:           "FLOAT32", // BFLOAT16 / FLOAT16 / FLOAT32 / FLOAT64. BFLOAT16 and FLOAT16 require v2.10 or later.
					Dim:            1024,      // keeps same with dimensions of Embedding (text-embedding-v4 default)
					DistanceMetric: "COSINE",  // L2 / IP / COSINE
				},
			}}).Result()
	return err
}

func answerQuestion(ctx context.Context, rdb *redis.Client, question string) error {
	embedder, err := dashscope.NewEmbedder(ctx, &dashscope.EmbeddingConfig{
		APIKey:     modelKey,
		Model:      embeddingModelName,
		Dimensions: &dim,
	})
	if err != nil {
		return err
	}
	retriever, err := rr.NewRetriever(ctx, &rr.RetrieverConfig{
		Client:    rdb,
		Index:     redisIndexName,
		Embedding: embedder,
		TopK:      3,
	})
	if err != nil {
		return err
	}
	retrievedDocs, err := retriever.Retrieve(ctx, question)
	if err != nil {
		return err
	}
	if len(retrievedDocs) == 0 {
		log.Println("未能从知识库中找到相关信息。")
		return nil
	}
	var contextBuilder strings.Builder
	for i, doc := range retrievedDocs {
		log.Printf("  - 相关片段 %d: %s\n", i+1, strings.ReplaceAll(doc.Content, "\n", " "))
		contextBuilder.WriteString(doc.Content)
		contextBuilder.WriteString("\n\n")
	}
	ragTemplate := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(`
			你是一个智能助手。请根据下面提供的上下文信息来回答用户的问题。请确保你的回答完全基于所给的上下文，不要使用任何外部知识。如果上下文中没有足够信息来回答问题，请直接说“根据所提供的信息，我无法回答该问题。
			--- 上下文 ---
			{context}
		`),
		schema.UserMessage("{question}"),
	)
	var vars = map[string]any{
		"question": question,
		"context":  contextBuilder.String(),
	}
	messages, err := ragTemplate.Format(ctx, vars)
	if err != nil {
		return err
	}
	log.Println("--- 最终发送给 LLM 的提示词 ---")
	for _, msg := range messages {
		log.Printf("[%s]: %s\n", msg.Role, msg.Content)
	}
	log.Println("---------------------------------")
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: modelBaseUrl,
		Model:   chatModelName,
		APIKey:  modelKey,
		//ExtraFields: map[string]any{"enable_thinking": false},
	},
	)
	if err != nil {
		return err
	}

	stream, err := chatModel.Stream(ctx, messages)
	if err != nil {
		return err
	}
	defer stream.Close()

	log.Println("\n--- AI 回答 ---")
	for {
		chunk, err := stream.Recv()
		if err != nil {
			break // 流结束或发生错误
		}
		print(chunk.Content)
	}
	println() // 换行

	return nil
}
