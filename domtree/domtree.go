package domtree

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"

	meta "github.com/yuin/goldmark-meta"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type DomTree struct {
	dir       string
	root      *html.Node
	metaData  map[string]any
	timestamp *html.Node
}

func (dt *DomTree) Init(src string) error {
	dt.dir = filepath.Dir(src)

	mu, ctx, err := fromFile(src)
	if err != nil {
		return err
	}
	dt.metaData = meta.Get(ctx)

	nodes, err := html.ParseFragment(strings.NewReader(mu), newDivNode())
	if err != nil {
		return err
	}

	d := newDivNode()
	setId(d, "main")

	for _, n := range nodes {
		d.AppendChild(n)
	}
	dt.root = d

	dt.timestamp = newTimestampNode(src)

	return nil
}

func (dt DomTree) GetTitle() string {
	m := dt.metaData
	if m == nil {
		return ""
	}
	s, ok := m["title"].(string)
	if ok {
		return s
	}
	return ""
}

func (dt DomTree) GetCSSs() (paths []string) {
	m := dt.metaData
	if m == nil {
		return
	}
	loadIface, ok := m["load"].([]any)
	if ok {
		for _, v := range loadIface {
			s, ok := v.(string)
			if ok {
				p := filepath.Join(dt.dir, s)
				paths = append(paths, p)
			}
		}
	}
	return
}

func (dt DomTree) getTOC() *html.Node {
	d := newDivNode()

	headers := findElements(dt.root, []string{"h2", "h3", "h4", "h5", "h6"})
	if len(headers) > 0 {
		ul := newUlNode()
		for _, header := range headers {
			a := newANode()
			appendAttr(a, "href", "#"+getAttribute(header, "id"))
			a.AppendChild(newTextNode(getTextContent(header)))

			li := newLiNode()
			appendClass(li, "toc-"+header.Data)

			li.AppendChild(a)
			ul.AppendChild(li)
		}
		d.AppendChild(ul)
	}

	setId(d, "toc")
	return d
}

func (dt *DomTree) renderArrowList() {
	s := "=>"
	var dfs func(*html.Node)
	dfs = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "li" && node.FirstChild != nil {
			if strings.HasPrefix(node.FirstChild.Data, s) {
				node.FirstChild.Data = strings.TrimPrefix(node.FirstChild.Data, s)
				appendClass(node, "sub")
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			dfs(c)
		}
	}
	dfs(dt.root)
}

func (dt *DomTree) renderBlankList() {
	var dfs func(*html.Node)
	dfs = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "li" {
			if isBlankNode(node) {
				appendClass(node, "empty")
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			dfs(c)
		}
	}
	dfs(dt.root)
}

func (dt *DomTree) renderTableCellWithCheckbox() {
	var dfs func(*html.Node)
	dfs = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "td" {
			t := strings.TrimSpace(getTextContent(node))
			if strings.HasPrefix(t, "[") && strings.HasSuffix(t, "]") && len(t) == 3 {
				node.RemoveChild(node.FirstChild)
				c := newCheckboxInputNode(true, strings.Index(t, "x") == 1)
				node.AppendChild(c)
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			dfs(c)
		}
	}
	dfs(dt.root)
}

func (dt *DomTree) renderPageBreak() {
	for c := dt.root.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "p" {
			if len(strings.ReplaceAll(c.FirstChild.Data, "=", "")) < 1 {
				c.FirstChild = nil
				appendClass(c, "page-separator")
			}
		}
	}
}

func (dt *DomTree) renderPDFLink() {
	var dfs func(*html.Node)
	dfs = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			if h := getAttribute(node, "href"); strings.HasSuffix(h, ".pdf") {
				appendAttr(node, "filetype", "pdf")
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			dfs(c)
		}
	}
	dfs(dt.root)
}

func (dt *DomTree) renderCodeblockLabel() {
	var dfs func(*html.Node)
	dfs = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "code" {
			if c := getAttribute(node, "class"); strings.HasPrefix(c, "language-") {
				l := strings.TrimPrefix(c, "language-")
				if p := node.Parent; p != nil && p.Data == "pre" {
					appendClass(p, "codeblock-header")
					appendAttr(p, "data-label", l)
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			dfs(c)
		}
	}
	dfs(dt.root)
}

func (dt *DomTree) fixHeadingSpacing() {
	var dfs func(*html.Node)
	dfs = func(node *html.Node) {
		if node.Type == html.ElementNode && isHeadingElem(node) && node.FirstChild != nil {
			t := strings.TrimSpace(getTextContent(node))
			l := utf8.RuneCountInString(t)
			if 2 <= l && l <= 4 {
				c := fmt.Sprintf("spacing-%d", l)
				appendClass(node, c)
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			dfs(c)
		}
	}
	dfs(dt.root)
}

func (dt *DomTree) setLinkTarget() {
	var dfs func(*html.Node)
	dfs = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			h := getAttribute(node, "href")
			if !strings.HasPrefix(h, "#") {
				appendAttr(node, "target", "_blank")
				appendAttr(node, "rel", "noopener noreferrer")
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			dfs(c)
		}
	}
	dfs(dt.root)
}

func (dt *DomTree) setImageContainer() {
	var dfs func(*html.Node)
	dfs = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "p" {
			var imgNode *html.Node
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "img" {
					imgNode = c
					break
				}
			}
			if imgNode != nil {
				container := newDivNode()
				appendClass(container, "img-container")
				wrapper := newDivNode()
				appendClass(wrapper, "img-wrapper")
				if a := getAttribute(node, "alt"); a == "left" || a == "right" {
					appendAttr(container, "pos", a)
				}
				wrapper.AppendChild(imgNode)
				container.AppendChild(wrapper)
				node.Parent.InsertBefore(container, node)
				node.Parent.RemoveChild(node)
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			dfs(c)
		}
	}
	dfs(dt.root)
}

func (dt *DomTree) applyAll() {
	dt.renderArrowList()
	dt.renderBlankList()
	dt.renderTableCellWithCheckbox()
	dt.renderPageBreak()
	dt.renderPDFLink()
	dt.renderCodeblockLabel()
	dt.fixHeadingSpacing()
	dt.setLinkTarget()
	dt.setImageContainer()
}

func (dt *DomTree) AsBodyNode() *html.Node {
	dt.applyAll()
	d := newDivNode()
	d.AppendChild(dt.root)
	setId(d, "container")

	d.InsertBefore(dt.timestamp, d.FirstChild)
	d.InsertBefore(dt.getTOC(), d.FirstChild)

	b := newElementNode("body", atom.Body)
	b.AppendChild(d)

	return b
}
