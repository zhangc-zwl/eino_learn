package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// sk-25cfec2f986a4cf1bf4187493a268d11
func main() {
	fmt.Println("start:", time.Now())
	ctx := context.Background()
	model, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
		Model:   "qwen3-32b",
		APIKey:  "sk-25cfec2f986a4cf1bf4187493a268d11",
	},
	)
	if err != nil {
		panic(err)
	}
	//创建message，这个就相当于给大模型的提示词，分为系统输入和用户输入
	//系统输入就是预设的提示词，用户输入就是用户输入的内容
	//message := []*schema.Message{
	//	schema.SystemMessage("你是一个乐于助人的助手"),
	//	schema.UserMessage("请介绍一下Go语言的特点"),
	//}

	template := prompt.FromMessages(
		schema.GoTemplate,
		schema.SystemMessage("你是一个{role}, 请用{tone}的语气回答问题"),
		schema.UserMessage("{question}"),
	)

	vars := map[string]any{
		"role":     "技术专家",
		"tone":     "专业严谨",
		"question": "如何优化数据库性能",
	}
	message, err := template.Format(ctx, vars)

	//调用大模型并生成响应
	//response, err := model.Generate(ctx, message)
	//if err != nil {
	//	panic(err)
	//}
	//println("model: ", response.Content)
	stream, err := model.Stream(ctx, message)
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	for {
		chunk, err := stream.Recv()
		if err != nil {
			break
		}
		print(chunk.Content)
	}
	fmt.Println("\nend:", time.Now())
}
