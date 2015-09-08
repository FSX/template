package template

import "strings"

type Node interface {
}

type ParentNode interface {
	Node

	Append(Node)
	Children() []Node
}

type NamedNode interface {
	Name() string
}

// listNode holds child nodes.
type listNode struct {
	children []Node
}

func newList() *listNode {
	return &listNode{}
}

func (l *listNode) Append(n Node) {
	l.children = append(l.children, n)
}

func (l *listNode) Children() []Node {
	return l.children
}

// textNode holds plain text.
type textNode struct {
	Text string
}

func newText(text string) *textNode {
	return &textNode{text}
}

// variableNode holds a list of identifiers,
// strings and numbers (i.e. an pexression).
type variableNode struct {
	Head Node
	Tail []Node
}

func newVariable(head Node, tail []Node) *variableNode {
	return &variableNode{head, tail}
}

// commentNode holds a comment.
type commentNode struct {
	Text string
}

func newComment(text string) *commentNode {
	return &commentNode{text}
}

// sectionNode holds an expression and child nodes.
type sectionNode struct {
	Head     *identifierNode
	Tail     []Node
	Inverted bool
	children []Node
}

func newSection(head *identifierNode, tail []Node, inverted bool) *sectionNode {
	return &sectionNode{Head: head, Tail: tail, Inverted: inverted}
}

func (s *sectionNode) Name() string {
	return s.Head.Name()
}

func (s *sectionNode) Append(n Node) {
	s.children = append(s.children, n)
}

func (s *sectionNode) Children() []Node {
	return s.children
}

// partialNode holds a reference to another template.
type partialNode struct {
	name string
}

func newPartial(name string) *partialNode {
	return &partialNode{name}
}

func (p *partialNode) Name() string {
	return p.name
}

// inheritNode holds a reference to an other template and subtemplates.
type inheritNode struct {
	name  string
	tmpls map[string][]Node
}

func newInherit(name string) *inheritNode {
	return &inheritNode{name, make(map[string][]Node)}
}

func (i *inheritNode) Name() string {
	return i.name
}

func (i *inheritNode) Append(n Node) {
	var name string

	if d, ok := n.(*defineNode); ok {
		name = d.Name()
	} else {
		name = "default" // right name?
	}

	i.tmpls[name] = append(i.tmpls[name], n)
}

func (i *inheritNode) Children() []Node {
	var children []Node

	for _, n := range i.tmpls {
		children = append(children, n...)
	}

	return children
}

// defineNode has a name and holds child nodes.
type defineNode struct {
	name     string
	children []Node
}

func newDefine(name string) *defineNode {
	return &defineNode{name: name}
}

func (d *defineNode) Name() string {
	return d.name
}

func (d *defineNode) Append(n Node) {
	d.children = append(d.children, n)
}

func (d *defineNode) Children() []Node {
	return d.children
}

// closeNode represents the closing tag of a section,
// subtemplate or inherit tag. closeNode is not included
// in the final tree of nodes.
type closeNode struct {
	name string
}

func newClose(name string) *closeNode {
	return &closeNode{name}
}

func (c *closeNode) Name() string {
	return c.name
}

// identifierNode holds a reference to an
// identifier (e.g. a variable or function).
type identifierNode struct {
	path []string
}

func newIdentifier(path []string) *identifierNode {
	return &identifierNode{path}
}

func (i *identifierNode) Name() string {
	return strings.Join(i.path, ".")
}

// stringNode holds plain text.
type stringNode struct {
	Text string
}

func newString(text string) *stringNode {
	return &stringNode{text}
}

// numberNode holds a number (e.g. int, uint, float, complex).
//
// TODO: Convert text to the actual number type.
type numberNode struct {
	Text string // Text representation of the number.
}

func newNumber(text string) *numberNode {
	return &numberNode{text}
}
