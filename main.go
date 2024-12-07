package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AWtnb/m2h/domtree"
	"github.com/AWtnb/m2h/frontmatter"
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
	var dt domtree.DomTree
	if err := dt.Init(src); err != nil {
		return err
	}

	var fm frontmatter.Frontmatter
	fm.Init(src, dt.GetMetaData())

	doc := domtree.NewHtmlNode("ja")

	h := domtree.NewHeadNode(fm.GetTitle(), css)
	domtree.AppendStyles(h, fm.GetCSSs())
	doc.AppendChild(h)

	b := dt.AsBodyNode()
	doc.AppendChild(b)

	o := strings.TrimSuffix(src, filepath.Ext(src)) + suffix + ".html"
	writeFile(domtree.Decode(doc), o)

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
