package frontmatter

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Frontmatter struct {
	src  string
	data map[string]interface{}
}

func (fm *Frontmatter) Init(src string, data map[string]interface{}) {
	fm.src = src
	fm.data = data
}

func (fm Frontmatter) GetTitle() string {
	t := fm.data["title"]
	if t != nil {
		return fmt.Sprint(t)
	}
	return strings.TrimSuffix(filepath.Base(fm.src), filepath.Ext(fm.src))
}

func (fm Frontmatter) GetCSSs() []string {
	d := filepath.Dir(fm.src)
	var ps []string
	l := fm.data["load"]
	if data, ok := l.([]interface{}); ok {
		for _, s := range data {
			if s, ok := s.(string); ok {
				p := filepath.Join(d, s)
				ps = append(ps, p)
			}
		}
	}
	return ps
}
