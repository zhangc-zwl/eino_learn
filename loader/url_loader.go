package main

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/loader/url"
	"github.com/cloudwego/eino/components/document"
)

func main() {
	ctx := context.Background()
	loader, err := url.NewLoader(ctx, &url.LoaderConfig{})
	if err != nil {
		panic(err)
	}
	path := "https://www.cloudwego.io/zh/docs/eino/core_modules/components/document_loader_guide"
	docs, err := loader.Load(ctx, document.Source{
		URI: path,
	})
	if err != nil {
		panic(err)
	}

	for _, doc := range docs {
		println(doc.String())
	}
}
