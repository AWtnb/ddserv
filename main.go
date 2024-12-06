package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AWtnb/m2h/dom"
	"github.com/AWtnb/m2h/md"
	meta "github.com/yuin/goldmark-meta"
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
	mu, ctx, err := md.FromFile(src)
	if err != nil {
		return err
	}

	var fm md.Frontmatter
	fm.Init(src, meta.Get(ctx))

	doc := dom.NewHtmlNode("ja")

	h := dom.NewHeadNode(fm.GetTitle(), css)
	dom.AppendStyles(h, fm.GetCSSs())
	doc.AppendChild(h)

	var mn dom.MainNode
	if err := mn.Init(src, mu); err != nil {
		return err
	}

	c := mn.AsContainerNode()

	b := dom.NewBodyNode()
	b.AppendChild(c)

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

	if !strings.HasSuffix(src, ".md") {
		fmt.Printf("invalid path: %s\n", src)
		os.Exit(1)
	}
	os.Exit(run(src, css, suffix))
}
