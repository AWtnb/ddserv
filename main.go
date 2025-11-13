package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AWtnb/ddserv/domtree"
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

func render(src string, css string) (string, error) {
	var dt domtree.DomTree
	if err := dt.Init(src); err != nil {
		return "", err
	}

	doc := domtree.NewHtmlNode("ja")

	h := domtree.NewHeadNode(dt.Title, css)
	domtree.AppendStyles(h, dt.CssToLoad)
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
			fmt.Println(err)
			return
		}
		defer watcher.Close()

		info, err := os.Stat(src)
		if err != nil {
			fmt.Println(err)
			return
		}
		if info.IsDir() {
			watcher.Add(src)
		} else {
			watcher.Add(filepath.Dir(src))
		}

		debounce := time.NewTimer(0)
		if !debounce.Stop() {
			<-debounce.C
		}

		for {
			select {
			case event := <-watcher.Events:
				ext := strings.ToLower(filepath.Ext(event.Name))
				if ext != ".md" && ext != ".css" {
					continue
				}

				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					if !debounce.Stop() {
						select {
						case <-debounce.C:
						default:
						}
					}
					debounce.Reset(300 * time.Millisecond)

					go func(filename string) {
						<-debounce.C
						websocket.Message.Send(ws, "reload")
					}(event.Name)
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

func htmlHandler(src string, css string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		html, err := render(src, css)
		if err != nil {
			fmt.Println(err)
			return
		}
		html += reloadScript
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, html)
	}
}

func downloadString(url string) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Go-HTTP-Client/1.1")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func run(src string, plain, export bool) int {
	css := ""
	if !plain {
		u := "https://raw.githubusercontent.com/AWtnb/md-stylesheet/refs/heads/main/dist/style.css"
		s, err := downloadString(u)
		if err == nil {
			css = s
		} else {
			fmt.Println(err)
		}
	}
	if export {
		html, err := render(src, css)
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

	root := filepath.Dir(src)

	http.Handle("/socket", websocket.Handler(wsReloadHandler(src)))
	handler := htmlHandler(src, css)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			handler(w, r)
			return
		}

		fp := filepath.Join(root, r.URL.Path)
		if _, err := os.Stat(fp); err == nil {
			http.ServeFile(w, r, fp)
		} else {
			fmt.Printf("'%s' not found\n", fp)
		}
	})

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
	flag.BoolVar(&plain, "plain", false, "prevent loading css from https://github.com/AWtnb/md-stylesheet")
	flag.BoolVar(&export, "export", false, "export as sigle html file")
	flag.Parse()

	info, err := os.Stat(src)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if info.IsDir() {
		fmt.Printf("this is directory path: %s\n", src)
		os.Exit(1)
	}
	if !strings.HasSuffix(src, ".md") {
		fmt.Printf("invalid path: %s\n", src)
		os.Exit(1)
	}
	os.Exit(run(src, plain, export))
}
