package domtree

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	alertcallouts "github.com/ZMT-Creative/gm-alert-callouts"
)

func getTitle(yamlData map[string]any) string {
	s, ok := yamlData["title"].(string)
	if ok {
		return s
	}
	return ""
}

func getCssPaths(root string, yamlData map[string]any) (paths []string) {
	loadIface, ok := yamlData["load"].([]any)
	if ok {
		for _, v := range loadIface {
			s, ok := v.(string)
			if ok {
				p := filepath.Join(root, s)
				paths = append(paths, p)
			}
		}
	}
	return
}

func fromFile(src string) (markup, title string, cssPaths []string, err error) {
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
	context := parser.NewContext()
	if err = md.Convert(bs, &buf, parser.WithContext(context)); err != nil {
		return
	}
	markup = buf.String()

	metaData := meta.Get(context)
	if metaData == nil {
		return
	}
	title = getTitle(metaData)
	cssPaths = getCssPaths(filepath.Dir(src), metaData)
	return
}
