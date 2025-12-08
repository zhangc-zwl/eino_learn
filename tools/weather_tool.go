package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type WeatherTool struct {
	apiKey string
}

func NewWeatherTool(apiKey string) *WeatherTool {
	return &WeatherTool{apiKey: apiKey}
}

func (w *WeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_weather",
		Desc: "获取指定城市和日期天气信息，例如：get_weather(city='上海'，extensions='base')",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"city": {
				Type:     schema.String,
				Required: true,
				Desc:     "城市名称",
			},
			"extensions": {
				Desc: "气象类型: base(实况天气) / all(预报天气)",
				Type: schema.String,
				Enum: []string{"base", "all"},
			},
		}),
	}, nil
}

func (w *WeatherTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params map[string]any
	if err := sonic.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}
	city, ok := params["city"].(string)
	if !ok || city == "" {
		return "", fmt.Errorf("city is required")
	}

	baseURL := "https://restapi.amap.com/v3/weather/weatherInfo"
	queryParams := url.Values{}
	queryParams.Set("city", city)
	queryParams.Set("key", w.apiKey)

	if extensions, ok := params["extensions"].(string); ok {
		queryParams.Set("extensions", extensions)
	} else {
		queryParams.Set("extensions", "base")
	}
	queryParams.Set("output", "JSON")

	fullUrl := baseURL + "?" + queryParams.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullUrl, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	return string(body), nil
}
