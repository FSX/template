package template

import (
	"io"
)

type NodeStorage interface {
	Get(string) (Node, bool)
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
