package main

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/html"
	"github.com/cloudwego/eino/schema"
)

var commonSuccessHTML = `<!DOCTYPE html>
<html>
<body>
    <div>
        <h1>H1</h1>
        <p>H1 content1</p>
        <div>
            <h2>H2.1</h2>
            <p>H2.1 content</p>
            <h3>H3.1</h3>
            <p>H3.1 content</p>
            <h3>H3.2</h3>
            <p>H3.2 content</p>
            <h2>H2.2</h2>
            <p>H2.2 content</p>
        </div>
        <div>
            <h2>H2.3</h2>
            <p>H2.3 content</p>
        </div>
		<div>
			<p>H1 content2</p>
		</div>
        <br>
        <p>H1 content3</p>
    </div>
	<div>
		<h2>H2.4</h2>
		<p>H2.4 content</p>
	</div>
	<div>
		<p>content</p>
	</div>
</body>
</html>`

func main() {
	ctx := context.Background()
	splitter, err := html.NewHeaderSplitter(ctx, &html.HeaderConfig{
		Headers: map[string]string{
			"h1": "h1",
			"h2": "h2",
			"h3": "h3",
			"h4": "h4",
			"h5": "h5",
			"h6": "h6",
		},
	})
	if err != nil {
		panic(err)
	}
	docs, err := splitter.Transform(ctx, []*schema.Document{
		{
			ID:      "1",
			Content: commonSuccessHTML,
		},
	})
	if err != nil {
		panic(err)
	}
	for _, doc := range docs {
		println(doc.String())
	}
}
