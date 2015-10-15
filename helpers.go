package template

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
)

type Options struct {
	LeftDelim, RightDelim string
	StripExtension        bool
}

type NodeMap struct {
	sync.RWMutex
	m map[string]Node
}

func (n *NodeMap) Get(name string) (Node, bool) {
	n.Lock()
	defer n.Unlock()

	if p, ok := n.m[name]; ok {
		return p, true
	} else {
		return nil, false
	}
}

func (n *NodeMap) Set(name string, p Node) {
	n.Lock()
	n.m[name] = p
	n.Unlock()
}

func ParseFiles(options *Options, basedir string, filenames ...string) (*Template, error) {
	if len(filenames) == 0 {
		return nil, fmt.Errorf("template: no files named in call to ParseFiles")
	}

	m := make(map[string]Node)
	if options == nil {
		options = &Options{"", "", false}
	}

	for _, fn := range filenames {
		p := filepath.Join(basedir, fn)

		b, err := ioutil.ReadFile(p)
		if err != nil {
			return nil, err
		}

		n, err := Parse(fn, options.LeftDelim, options.RightDelim, string(b))
		if err != nil {
			return nil, err
		}

		if options.StripExtension {
			m[stripExt(fn)] = n
		} else {
			m[fn] = n
		}

	}

	return New(&NodeMap{m: m}), nil
}

func stripExt(filename string) string {
	a := 0
	if r := strings.LastIndex(filename, "/"); r > -1 {
		a = r + 1
	}

	if b := strings.IndexRune(filename[a:], '.'); b > -1 {
		return filename[:a+b]
	} else {
		return filename
	}
}
