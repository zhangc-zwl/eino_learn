package main

import (
	"context"

	"github.com/cloudwego/eino-ext/a2a/extension/eino"
	"github.com/cloudwego/eino-ext/a2a/transport/jsonrpc"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	hertzServer "github.com/cloudwego/hertz/pkg/app/server"
)

func main() {
	ctx := context.Background()
	h := hertzServer.Default()
	r, err := jsonrpc.NewRegistrar(ctx, &jsonrpc.ServerConfig{
		Router:      h,
		HandlerPath: "/a2a",
	})
	if err != nil {
		panic(err)
	}
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     "https://dashscope.aliyuncs.com/compatible-mode/v1",
		Model:       "qwen3-32b",
		APIKey:      "sk-25cfec2f986a4cf1bf4187493a268d11",
		ExtraFields: map[string]any{"enable_thinking": false},
	},
	)
	if err != nil {
		panic(err)
	}
	chatAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "book-recommender",
		Description: "A book recommender agent",
		Instruction: "You are a book recommender agent. You are given a book title and you need to recommend 5 books that are similar to the given book. You need to use the book-recommender tool to get the recommendations. You need to use the book-recommender tool to get the recommendations. You need to use the book-recommender tool to get the recommendations. You need to use the book-recommender tool to get the recommendations. ",
		Model:       chatModel,
	})
	if err != nil {
		panic(err)
	}
	err = eino.RegisterServerHandlers(ctx, chatAgent, &eino.ServerConfig{
		Registrar: r,
	})
	if err != nil {
		panic(err)
	}
	err = h.Run()
	if err != nil {
		panic(err)
	}
}
