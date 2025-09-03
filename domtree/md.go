package domtree

import (
	"bytes"
	"os"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	alertcallouts "github.com/ZMT-Creative/gm-alert-callouts"
)

func fromFile(src string) (markup string, context parser.Context, err error) {
	bs, err := os.ReadFile(src)
	if err != nil {
		return
	}

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, meta.Meta, alertcallouts.NewAlertCallouts(
			alertcallouts.UseGFMStrictIcons(),
		)),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	context = parser.NewContext()
	if err = md.Convert(bs, &buf, parser.WithContext(context)); err != nil {
		return
	}
	markup = buf.String()
	return
}
