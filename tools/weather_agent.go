package main

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// 1a988ad463f95506ce34f6c30ea56603
func main() {
	ctx := context.Background()
	weatherTool := NewWeatherTool("1a988ad463f95506ce34f6c30ea56603")
	//创建ToolsNode
	toolNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{weatherTool},
	})
	if err != nil {
		panic(err)
	}
	model, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     "https://dashscope.aliyuncs.com/compatible-mode/v1",
		Model:       "qwen3-32b",
		APIKey:      "sk-25cfec2f986a4cf1bf4187493a268d11",
		ExtraFields: map[string]any{"enable_thinking": false},
	},
	)
	if err != nil {
		panic(err)
	}
	weatherInfo, err := weatherTool.Info(ctx)
	if err != nil {
		panic(err)
	}
	toolCallingChatModel, err := model.WithTools([]*schema.ToolInfo{
		weatherInfo,
	})
	if err != nil {
		panic(err)
	}
	messages := prompt.FromMessages(schema.GoTemplate,
		schema.SystemMessage("你是一个AI助手，你必须调用工具来获取天气信息"),
		schema.UserMessage("我需要查询北京今天的天气"),
	)
	vars := map[string]any{}
	result, err := messages.Format(ctx, vars)
	input, err := toolCallingChatModel.Generate(ctx, result)
	if err != nil {
		panic(err)
	}
	println(input.Content)
	for _, v := range input.ToolCalls {
		println("=======================")
		println(v.Function.Name)
		println(v.Function.Arguments)
	}
	//模拟一个message，添加tool call信息
	//input := &schema.Message{
	//	Role: schema.Assistant,
	//	ToolCalls: []schema.ToolCall{
	//		{
	//			Function: schema.FunctionCall{
	//				Name: "get_weather",
	//				Arguments: `{
	//					"city": "上海",
	//					"extensions": "base"
	//				}`,
	//			},
	//		},
	//	},
	//}
	toolMessage, err := toolNode.Invoke(ctx, input)
	if err != nil {
		panic(err)
	}
	for _, v := range toolMessage {
		println(v.Content)
	}
}
