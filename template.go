package template

import (
	"fmt"
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

func (t *Template) Execute(wr io.Writer, name string, data interface{}) error {
	node, ok := t.nodes.Get(name)
	if !ok {
		return fmt.Errorf("template not available: %s", name)
	}

	return t.execute(wr, node, data)
}

func (t *Template) execute(wr io.Writer, node Node, data interface{}) error {
	// PrintNodes(node, 0)

	switch n := node.(type) {
	case (*listNode):
		for _, n := range n.Children() {
			t.execute(wr, n, data)
		}
	case (*textNode):
		wr.Write([]byte(n.Text)) // Store text as bytes in nodes?
	case (*partialNode):
		if partial, ok := t.nodes.Get(n.Name()); !ok {
			return fmt.Errorf("template not available: %s", n.Name())
		} else {
			t.execute(wr, partial, data)
		}
	default:
		panic("unknown node")
	}

	return nil
}
