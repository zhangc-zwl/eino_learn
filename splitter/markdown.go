package main

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	"github.com/cloudwego/eino/schema"
)

var mdContent = `
# 这是一级标题
这是一级标题内容
## 这是二级标题
这是二级标题内容
### 这是三级标题
这是三级标题内容
`

func main() {
	ctx := context.Background()
	splitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
		Headers: map[string]string{
			"#":      "h1",
			"##":     "h2",
			"###":    "h3",
			"####":   "h4",
			"#####":  "h5",
			"######": "h6",
		},
		TrimHeaders: false, //是否在输出中保留标题行 true 不保留 false 保留
	})
	if err != nil {
		panic(err)
	}
	docs, err := splitter.Transform(ctx, []*schema.Document{
		{
			ID:      "1",
			Content: mdContent,
		},
	})
	if err != nil {
		panic(err)
	}
	for _, doc := range docs {
		println(doc.String())
		println("=========================")
	}
}
