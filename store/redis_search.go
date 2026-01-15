package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/cloudwego/eino-ext/components/embedding/dashscope"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

	// --- 1. 连接到 Redis ---
	rdb := redis.NewClient(&redis.Options{
		Addr:          "36.151.150.11:6379", // Redis Stack 服务的地址
		UnstableResp3: true,
		Protocol:      2,
	})
	indexName := "doc_index"
	// 构建 KNN 查询
	// `*=>[KNN 2 @vector_content $blob]` 的含义:
	// - `*`: 匹配所有文档 (我们不过滤元数据)。
	// - `=>`: 表示这是一个混合查询，我们主要关心右边的向量部分。
	// - `[KNN 2 @vector_content $blob]`: 在 `vector_content` 字段上执行一个 K-最近邻查询，
	//   查找 2 个最近邻。`$blob` 是一个参数，我们将把查询向量的二进制数据传递给它。
	// DIALECT 2 是必须的，用于支持这种现代的查询语法。
	k := 2
	query := fmt.Sprintf("*=>[KNN %d @vector_content $blob AS score]", k)
	searchContent := "That is a happy person"
	dim := 1024
	embedder, err := dashscope.NewEmbedder(ctx, &dashscope.EmbeddingConfig{
		APIKey:     "sk-25cfec2f986a4cf1bf4187493a268d11",
		Model:      "text-embedding-v4",
		Dimensions: &dim,
	})
	if err != nil {
		panic(err)
	}
	embeddings, err := embedder.EmbedStrings(ctx, []string{searchContent})
	if err != nil {
		panic(err)
	}

	// 先检查 Redis 中是否有数据
	fmt.Println("=== 检查 Redis 数据 ===")
	keys, err := rdb.Keys(ctx, "eino_doc:*").Result()
	if err != nil {
		panic(err)
	}
	fmt.Printf("找到 %d 个 keys\n", len(keys))
	for _, key := range keys {
		fmt.Printf("  - %s\n", key)
	}

	// 检查索引信息
	fmt.Println("\n=== 检查索引信息 ===")
	indexInfo, err := rdb.FTInfo(ctx, indexName).Result()
	if err != nil {
		panic(err)
	}
	fmt.Printf("索引信息: %v\n", indexInfo)

	fmt.Println("\n=== 执行向量查询 ===")
	vectorBytes := vector2Bytes(embeddings[0])
	fmt.Printf("查询向量长度: %d bytes (应该是 %d)\n", len(vectorBytes), 1024*4)

	// 使用 redis.NewSearch 来构建带参数的查询
	searchResult, err := rdb.FTSearchWithArgs(ctx, indexName, query, &redis.FTSearchOptions{
		Params: map[string]interface{}{
			"blob": vectorBytes,
		},
		DialectVersion: 2,
		Return: []redis.FTSearchReturn{
			{
				FieldName: "content",
			},
			{
				FieldName: "score",
			},
		},
		SortBy: []redis.FTSearchSortBy{
			{
				FieldName: "score",
				Asc:       true,
			},
		},
	}).Result()
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n查询结果: 总共找到 %d 个文档\n", searchResult.Total)
	if searchResult.Total == 0 {
		fmt.Println("⚠️ 没有找到任何文档！")
	} else {
		fmt.Println("\n=== 搜索结果 ===")
		for i, v := range searchResult.Docs {
			fmt.Printf("[%d] Content: %v, Score: %v\n", i+1, v.Fields["content"], v.Fields["score"])
		}
	}
}

func vector2Bytes(vector []float64) []byte {
	float32Arr := make([]float32, len(vector))
	for i, v := range vector {
		float32Arr[i] = float32(v)
	}
	bytes := make([]byte, len(float32Arr)*4)
	for i, v := range float32Arr {
		binary.LittleEndian.PutUint32(bytes[i*4:], math.Float32bits(v))
	}
	return bytes
}
