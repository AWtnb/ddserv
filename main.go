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

func wsReloadHandler(src string) websocket.Handler {
	return func(ws *websocket.Conn) {
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
					if filepath.Ext(event.Name) == ".md" || filepath.Ext(event.Name) == ".css" {
						websocket.Message.Send(ws, "reload")
					}
				}
			case err := <-watcher.Errors:
				fmt.Println("watcher error:", err)
				return
			}
		}
	}
}

const reloadScript = `
<script>
let ws = new WebSocket("ws://" + location.host + "/socket");
ws.onmessage = function(evt) { if(evt.data === "reload") location.reload(); }
</script>`

func htmlHandler(src string, plain bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		html, err := render(src, plain)
		if err != nil {
			return
		}
		html += reloadScript
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, html)
	}
}

func run(src string, plain, export bool) int {
	if export {
		html, err := render(src, plain)
		if err != nil {
			fmt.Println(err)
			return 1
		}
		o := strings.TrimSuffix(src, filepath.Ext(src)) + ".html"
		if err := writeFile(html, o); err != nil {
			fmt.Println(err)
			return 1
		}
		return 0
	}

	http.Handle("/socket", websocket.Handler(wsReloadHandler(src)))
	http.HandleFunc("/", htmlHandler(src, plain))
	fmt.Println("Serving at http://localhost:8080")
	if http.ListenAndServe(":8080", nil) != nil {
		return 1
	}
	return 0
}

func main() {
	var (
		src    string
		plain  bool
		export bool
	)
	flag.StringVar(&src, "src", "", "markdown path")
	flag.BoolVar(&plain, "plain", false, "prevent loading css from https://github.com/AWtnb/md-less")
	flag.BoolVar(&export, "export", false, "export as sigle html file")
	flag.Parse()

	if !strings.HasSuffix(src, ".md") {
		fmt.Printf("invalid path: %s\n", src)
		os.Exit(1)
	}
	os.Exit(run(src, plain, export))
}
