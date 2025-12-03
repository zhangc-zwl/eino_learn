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

	template := prompt.FromMessages(
		schema.GoTemplate,
		schema.SystemMessage("{{if .isExpert}}你是一个专家级{{.domain}}顾问。{{else}}你是一个初级{{.domain}}助手。{{end}}\n{{if .isFormal}}请使用正式的语言风格。{{else}}请使用友好的语言风格。{{end}}\n你的任务是{{.task}}。"),
		schema.UserMessage("{{.question}}"),
	)

	vars := map[string]interface{}{
		"isExpert": false,
		"domain":   "编程",
		"isFormal": false,
		"task":     "帮助初学者理解编程概念",
		"question": "什么是变量？",
	}
	message, err := template.Format(ctx, vars)

	for _, v := range message {
		println(v.Content)
	}

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
