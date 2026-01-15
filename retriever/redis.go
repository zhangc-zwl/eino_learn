package main

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/embedding/dashscope"
	rr "github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

	client := redis.NewClient(&redis.Options{
		Addr:          "36.151.150.11:6379",
		Protocol:      2,
		UnstableResp3: true,
	})
	dim := 1024
	embedder, err := dashscope.NewEmbedder(ctx, &dashscope.EmbeddingConfig{
		APIKey:     "sk-25cfec2f986a4cf1bf4187493a268d11",
		Model:      "text-embedding-v4",
		Dimensions: &dim,
	})
	if err != nil {
		panic(err)
	}
	r, err := rr.NewRetriever(ctx, &rr.RetrieverConfig{
		Client:    client,
		Index:     "doc_index",
		Embedding: embedder,
	})
	if err != nil {
		panic(err)
	}
	docs, err := r.Retrieve(ctx, "That is a happy person")
	if err != nil {
		panic(err)
	}
	for _, v := range docs {
		fmt.Printf("ID:%s, CONTENT:%v \n", v.ID, v.Content)
	}
}
