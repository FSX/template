package template

import (
	"path/filepath"
	"sync"
)

type NodeMap struct {
	sync.RWMutex
	m map[string]Node
}

func (n *nodeMap) Get(name string) (Node, bool) {
	n.Lock()
	defer n.Unlock()

	if p, ok := n.m[name]; ok {
		return p, true
	} else {
		return nil, false
	}
}

func (n *nodeMap) Set(name string, p Node) {
	n.Lock()
	n.m[name] = p
	n.Unlock()
}

func ParseFiles(basedir string, filenames ...string) (*Template, error) {
	for _, n := range filenames {
		//
	}

	return nil, nil
}
