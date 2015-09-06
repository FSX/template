package template

import (
	"io"
)

type NodeStorage interface {
	Get(string, bool) Node
	Set(string, Node)
}

type Template struct {
	nodes NodeStorage
}

func New(n NodeStorage) *Template {
	t := &Template{n}
	return t
}

func (t *Template) Execute(wr io.Writer, name string, data interface{}) {
	// TODO
}
