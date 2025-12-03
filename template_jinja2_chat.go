package main

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

func main() {
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
		schema.Jinja2,
		schema.SystemMessage(`{% if level == 'expert' %}你是一个专家级顾问。{% else %}你是一个初级助手。{% endif %}
{% if domain %}你专长于{{ domain }}领域。{% endif %}
请用{% if formal %}正式{% else %}友好{% endif %}的语气回答问题。`),
		schema.UserMessage("{{question}}"),
	)
	//vars := map[string]any{
	//	"isExpert": true,
	//	"domain":   "数据库",
	//	"isFormal": true,
	//	"task":     "提供专业的数据库优化建议",
	//	"question": "如何优化大型数据库的查询性能？",
	//}
	vars := map[string]any{
		"level":    "expert",
		"domain":   "人工智能",
		"formal":   true,
		"question": "请解释Transformer模型的工作原理。",
	}
	message, err := template.Format(ctx, vars)
	for _, v := range message {
		println(v.Content)
	}
	//获取流式回复
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
	//换行
	println()
}
