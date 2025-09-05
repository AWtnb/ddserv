package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/AWtnb/m2h/domtree"
	"github.com/AWtnb/m2h/frontmatter"
	"github.com/fsnotify/fsnotify"
	"golang.org/x/net/websocket"
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

func render(src string, plain bool) (string, error) {
	var dt domtree.DomTree
	if err := dt.Init(src); err != nil {
		return "", err
	}

	var fm frontmatter.Frontmatter
	fm.Init(src, dt.GetMetaData())

	doc := domtree.NewHtmlNode("ja")

	h := domtree.NewHeadNode(fm.GetTitle(), plain)
	domtree.AppendStyles(h, fm.GetCSSs())
	doc.AppendChild(h)

	b := dt.AsBodyNode()
	doc.AppendChild(b)
	return domtree.Decode(doc), nil
}

const reloadScript = `
<script>
let ws = new WebSocket("ws://" + location.host + "/ws");
ws.onmessage = function(evt) { if(evt.data === "reload") location.reload(); }
</script>`

func previewOnLocalhost(src string, plain bool) error {
	http.Handle("/ws", websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return
		}
		defer watcher.Close()

		info, err := os.Stat(src)
		if err != nil {
			return
		}
		if info.IsDir() {
			watcher.Add(src)
		} else {
			watcher.Add(filepath.Dir(src))
		}

		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					websocket.Message.Send(ws, "reload")
				}
			case err := <-watcher.Errors:
				fmt.Println("watcher error:", err)
				return
			}
		}
	}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html, err := render(src, plain)
		if err != nil {
			return
		}
		html += reloadScript
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, html)
	})
	fmt.Println("Serving at http://localhost:8080")
	return http.ListenAndServe(":8080", nil)
}

func run(src, suffix string, plain bool, serve bool) int {
	if serve {
		if previewOnLocalhost(src, plain) != nil {
			return 1
		}
		return 0
	}
	html, err := render(src, plain)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	o := strings.TrimSuffix(src, filepath.Ext(src)) + suffix + ".html"
	if writeFile(html, o) != nil {
		return 1
	}
	return 0
}

func main() {
	var (
		src    string
		suffix string
		plain  bool
		serve  bool
	)
	flag.StringVar(&src, "src", "", "markdown path")
	flag.StringVar(&suffix, "suffix", "", "suffix of result html")
	flag.BoolVar(&plain, "plain", false, "flag to skip loading https://raw.githubusercontent.com/AWtnb/md-less/refs/heads/main/style.less")
	flag.BoolVar(&serve, "serve", false, "flag to start hot-reload localhost server to preview")
	flag.Parse()

	if !strings.HasSuffix(src, ".md") {
		fmt.Printf("invalid path: %s\n", src)
		os.Exit(1)
	}
	os.Exit(run(src, suffix, plain, serve))
}
