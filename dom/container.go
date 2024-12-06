package dom

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

type MainNode struct {
	root      *html.Node
	timestamp *html.Node
}

func (mn *MainNode) Init(src, markup string) error {
	nodes, err := html.ParseFragment(strings.NewReader(markup), newDivNode())
	if err != nil {
		return err
	}
	d := newDivNode()
	appendClass(d, "main")
	for _, n := range nodes {
		d.AppendChild(n)
	}
	mn.root = d
	mn.timestamp = newTimestampNode(src)
	return nil
}

func (mn MainNode) getTOC() *html.Node {
	d := newDivNode()
	appendClass(d, "toc")

	headers := findElements(mn.root, []string{"h2", "h3", "h4", "h5", "h6"})
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

	return d
}

func (mn *MainNode) renderArrowList() {
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
	dfs(mn.root)
}

func (mn *MainNode) renderBlankList() {
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
	dfs(mn.root)
}

func (mn *MainNode) renderPageBreak() {
	for c := mn.root.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "p" {
			if len(strings.ReplaceAll(c.FirstChild.Data, "=", "")) < 1 {
				c.FirstChild = nil
				appendClass(c, "page-separator")
			}
		}
	}
}

func (mn *MainNode) renderPDFLink() {
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
	dfs(mn.root)
}

func (mn *MainNode) renderCodeblockLabel() {
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
	dfs(mn.root)
}

func (mn *MainNode) fixHeadingSpacing() {
	var dfs func(*html.Node)
	dfs = func(node *html.Node) {
		if node.Type == html.ElementNode && isHeadingElem(node) && node.FirstChild != nil {
			t := getTextContent(node)
			l := len(strings.TrimSpace(t))
			if 2 <= l && l <= 4 {
				c := fmt.Sprintf("spacing-%d", l)
				appendClass(node, c)
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			dfs(c)
		}
	}
	dfs(mn.root)
}

func (mn *MainNode) setLinkTarget() {
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
	dfs(mn.root)
}

func (mn *MainNode) setImageContainer() {
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
	dfs(mn.root)
}

func (mn *MainNode) applyAll() {
	mn.renderArrowList()
	mn.renderBlankList()
	mn.renderPageBreak()
	mn.renderPDFLink()
	mn.renderCodeblockLabel()
	mn.fixHeadingSpacing()
	mn.setLinkTarget()
	mn.setImageContainer()
}

func (mn *MainNode) AsContainerNode() *html.Node {
	mn.applyAll()

	mn.root.InsertBefore(mn.timestamp, mn.root.FirstChild)

	d := newDivNode()
	appendClass(d, "container")
	d.AppendChild(mn.getTOC())
	d.AppendChild(mn.root)

	return d
}
