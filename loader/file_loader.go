package main

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino/components/document"
)

func main() {
	ctx := context.Background()
	loader, err := file.NewFileLoader(ctx, &file.FileLoaderConfig{
		UseNameAsID: true,
	})
	if err != nil {
		panic(err)
	}
	path := "./loader/test.md"
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
