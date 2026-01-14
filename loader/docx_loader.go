package main

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino-ext/components/document/parser/docx"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/document/parser"
)

func main() {
	ctx := context.Background()
	docxParser, err := docx.NewDocxParser(ctx, &docx.Config{})
	if err != nil {
		panic(err)
	}
	extParser, err := parser.NewExtParser(ctx, &parser.ExtParserConfig{
		FallbackParser: parser.TextParser{},
		Parsers: map[string]parser.Parser{
			".docx": docxParser,
		},
	})
	loader, err := file.NewFileLoader(ctx, &file.FileLoaderConfig{
		UseNameAsID: true,
		Parser:      extParser,
	})
	if err != nil {
		panic(err)
	}
	path := "./loader/test.docx"
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
