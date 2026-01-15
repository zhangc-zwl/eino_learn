package main

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino-ext/components/embedding/dashscope"
	ri "github.com/cloudwego/eino-ext/components/indexer/redis"
	"github.com/cloudwego/eino/schema"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr:          "36.151.150.11:6379",
		UnstableResp3: true,
		Protocol:      2,
	})
	keyPrefix := "eino_doc:"
	indexName := "doc_index"
	slice, err2 := client.Do(ctx, "FT._LIST").StringSlice()
	if err2 != nil {
		panic(err2)
	}
	indexExists := false
	for _, v := range slice {
		if v == indexName {
			indexExists = true
			break
		}
	}
	if indexExists {
		result, err2 := client.FTDropIndex(ctx, indexName).Result()
		if err2 != nil {
			panic(err2)
		}
		fmt.Println("🗑️ 删除旧索引:", result)
	}
	if !indexExists {
		//创建索引
		result, err := client.FTCreate(ctx, indexName, &redis.FTCreateOptions{
			OnHash: true,
			Prefix: []any{keyPrefix},
		}, &redis.FieldSchema{
			FieldName: "content",
			FieldType: redis.SearchFieldTypeText,
			Weight:    1,
		}, &redis.FieldSchema{
			FieldName: "vector_content",
			FieldType: redis.SearchFieldTypeVector,
			VectorArgs: &redis.FTVectorArgs{
				FlatOptions: &redis.FTFlatOptions{
					Type:           "FLOAT32", // BFLOAT16 / FLOAT16 / FLOAT32 / FLOAT64. BFLOAT16 and FLOAT16 require v2.10 or later.
					Dim:            1024,      // keeps same with dimensions of Embedding (text-embedding-v4 default)
					DistanceMetric: "COSINE",  // L2 / IP / COSINE
				},
			},
		}).Result()
		if err != nil {
			panic(err)
		}
		fmt.Println(result)
	}

	dim := 1024
	embedder, err := dashscope.NewEmbedder(ctx, &dashscope.EmbeddingConfig{
		APIKey:     "sk-25cfec2f986a4cf1bf4187493a268d11",
		Model:      "text-embedding-v4",
		Dimensions: &dim,
	})
	if err != nil {
		panic(err)
	}
	splitter, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   10,                            // 必需：目标片段大小
		OverlapSize: 2,                             // 可选：片段重叠大小
		Separators:  []string{"\n", ".", "?", "！"}, // 可选：分隔符列表
		LenFunc:     nil,                           // 可选：自定义长度计算函数
		KeepType:    recursive.KeepTypeNone,        // 可选：分隔符保留策略
		IDGenerator: func(ctx context.Context, originalID string, splitIndex int) string {
			return fmt.Sprintf("%s_%d", originalID, splitIndex)
		},
	})
	if err != nil {
		panic(err)
	}
	docs, err := splitter.Transform(ctx, []*schema.Document{
		{
			ID: "testDoc",
			Content: `
			That is a very happy person。
            That is a happy dog。
            Today is a sunny day。`,
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("✅ Split into %d documents\n", len(docs))
	for i, doc := range docs {
		fmt.Printf("  Doc %d: ID=%s, Content=%q\n", i, doc.ID, doc.Content)
	}
	indexer, err := ri.NewIndexer(ctx, &ri.IndexerConfig{
		Client:    client,
		KeyPrefix: keyPrefix,
		BatchSize: len(docs),
		Embedding: embedder,
	})
	if err != nil {
		panic(err)
	}
	ids, err := indexer.Store(ctx, docs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n✅ 成功存储文档，IDs: %v\n", ids)

	// 验证数据是否真的写入了
	fmt.Println("\n=== 验证数据写入 ===")
	keys, err := client.Keys(ctx, keyPrefix+"*").Result()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Redis 中有 %d 个文档\n", len(keys))

	// 查看第一个文档的内容
	if len(keys) > 0 {
		fmt.Printf("\n第一个文档的 key: %s\n", keys[0])
		result, err := client.HGetAll(ctx, keys[0]).Result()
		if err != nil {
			panic(err)
		}
		fmt.Println("文档字段:")
		for k, v := range result {
			if k == "vector_content" {
				fmt.Printf("  %s: <binary data, length=%d bytes>\n", k, len(v))
			} else {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}
	}
}
