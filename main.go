package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AWtnb/m2h/dom"
	"github.com/AWtnb/m2h/frontmatter"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func writeFile(t, out string) error {
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(t)
	if err != nil {
		return err
	}
	return nil
}

func render(src, css, suffix string) error {
	bs, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, meta.Meta),
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
	if err := md.Convert(bs, &buf, parser.WithContext(context)); err != nil {
		return err
	}

	var fm frontmatter.Frontmatter
	fm.Init(src, meta.Get(context))

	doc := dom.NewHtmlNode("ja")

	h := dom.NewHeadNode(fm.GetTitle(), css)
	dom.AppendStyles(h, fm.GetCSSs())
	doc.AppendChild(h)

	var cont dom.MainContainer
	if err := cont.Init(buf.String()); err != nil {
		return err
	}

	b := dom.NewBodyNode()
	b.AppendChild(dom.NewTimestampNode(src))
	b.AppendChild(cont.GetTOC())
	b.AppendChild(cont.GetTree())

	doc.AppendChild(b)

	o := strings.TrimSuffix(src, filepath.Ext(src)) + suffix + ".html"
	writeFile(dom.Decode(doc), o)

	return nil
}

func run(src, css, suffix string) int {
	err := render(src, css, suffix)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	return 0
}

func main() {
	var (
		src    string
		css    string
		suffix string
	)
	flag.StringVar(&src, "src", "", "markdown path")
	flag.StringVar(&css, "css", "https://cdn.jsdelivr.net/gh/Awtnb/md-less/style.less", "css path or url")
	flag.StringVar(&suffix, "suffix", "", "suffix of result html")
	flag.Parse()

	if !strings.HasPrefix(src, ".md") {
		fmt.Printf("invalid path: %s\n", src)
		os.Exit(1)
	}
	os.Exit(run(src, css, suffix))
}
