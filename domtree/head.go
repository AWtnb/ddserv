package domtree

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/html"
)

func getHeadMarkup(title string) string {
	var buf strings.Builder

	buf.WriteString(`<meta charset="utf-8">`)
	buf.WriteString(`<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">`)

	faviconHex := "1f4dd"
	fv := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><text x="50%%" y="50%%" style="dominant-baseline:central;text-anchor:middle;font-size:90px;">&#x%s;</text></svg>`, faviconHex)
	buf.WriteString(fmt.Sprintf(`<link rel="icon" href="data:image/svg+xml,%s">`, url.PathEscape(fv)))

	buf.WriteString(fmt.Sprintf(`<title>%s</title>`, title))

	return fmt.Sprintf(`<head>%s</head>`, buf.String())
}

func NewHeadNode(title string, plain bool) *html.Node {
	head := newHeadNode()
	m := getHeadMarkup(title)
	h, _ := html.ParseFragment(strings.NewReader(m), newHeadNode())
	for _, n := range h {
		head.AppendChild(n)
	}

	if !plain {
		l := newLinkNode()
		appendAttr(l, "type", "text/css")
		appendAttr(l, "rel", "stylesheet/less")
		appendAttr(l, "href", "https://raw.githubusercontent.com/AWtnb/md-less/refs/heads/main/style.less")
		head.AppendChild(l)

		s := newScriptNode()
		appendAttr(s, "src", "https://cdn.jsdelivr.net/npm/less")
		head.AppendChild(s)
	}

	return head
}

func AppendStyles(node *html.Node, paths []string) {
	for _, p := range paths {
		d, err := os.ReadFile(p)
		if err != nil {
			fmt.Printf("Error:%s\n", err)
		}
		n := newStyleNode()
		n.AppendChild(newTextNode(strings.TrimSpace(string(d))))
		node.AppendChild(n)
	}
}
