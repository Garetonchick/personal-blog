package utils

import (
	"html/template"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	"github.com/microcosm-cc/bluemonday"
)

func MD2SafeHTML(md []byte) template.HTML {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	// Render markdown to HTML
	rendered := markdown.Render(doc, renderer)

	// Sanitize potentially malicious HTML
	safeHTML := bluemonday.UGCPolicy().SanitizeBytes(rendered)

	return template.HTML(safeHTML)
}
