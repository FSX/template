package template

import (
	"errors"
	"fmt"
	"strings"
)

type parser struct {
	name string
	lex  *lexer
	err  error

	token     [3]item
	peekCount int
}

func Parse(name, leftDelim, rightDelim, input string) (ParentNode, error) {
	p := &parser{name: name, lex: lex(name, input, leftDelim, rightDelim)}
	root := newList()

	if !p.parse(root) {
		return nil, p.err
	}

	return root, nil
}

func (p *parser) next() item {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		p.token[0] = p.lex.nextItem()
	}
	return p.token[p.peekCount]
}

func (p *parser) backup() {
	p.peekCount++
}

func (p *parser) peek() item {
	if p.peekCount > 0 {
		return p.token[p.peekCount-1]
	}
	p.peekCount = 1
	p.token[0] = p.lex.nextItem()
	return p.token[0]
}

func (p *parser) nextNonSpace() (token item) {
	for {
		token = p.next()
		if token.typ != itemSpace {
			break
		}
	}

	return token
}

func (p *parser) peekNonSpace() (token item) {
	for {
		token = p.next()
		if token.typ != itemSpace {
			break
		}
	}
	p.backup()
	return token
}

func (p *parser) errorf(format string, args ...interface{}) Node {
	// Give priority to itemError tokens.
	var msg string
	if p.token[0].typ == itemError {
		msg = p.token[0].val
	} else {
		msg = fmt.Sprintf(format, args...)
	}

	pos := int(p.lex.pos)
	if i := strings.LastIndex(p.lex.input[:p.lex.lastPos], "\n"); i > -1 {
		pos = pos - i - 1
	}

	p.err = errors.New(fmt.Sprintf(
		"%s:%d:%d: %s", p.name, p.lex.lineNumber(), pos, msg))

	return nil
}

// Parse functions

func (p *parser) parse(parent ParentNode) bool {
	name := ""
	close := true

	if n, ok := parent.(NamedNode); ok {
		name = n.Name()
		close = false
	}

	for {
		t := p.peek()

		if t.typ == itemEOF {
			if !close {
				p.errorf("tag not closed")
			}

			break
		} else if t.typ == itemError {
			p.errorf(t.val)
			break
		}

		if n := p.textOrTag(); n != nil {
			if c, ok := n.(*closeNode); ok {
				if name != c.Name() {
					p.errorf("unexpected closing tag")
				}

				break
			}

			parent.Append(n)
		} else {
			break
		}
	}

	return p.err == nil
}

func (p *parser) textOrTag() Node {
	t := p.nextNonSpace()

	switch t.typ {
	case itemText:
		return newText(t.val)
	case itemLeftDelim:
		return p.parseTag()
	}

	return p.errorf("unexpected token: %s", t.val)
}

func (p *parser) parseTag() Node {
	t := p.peekNonSpace()

	switch t.typ {
	case itemRightDelim:
		p.nextNonSpace()
		return p.errorf("empty tags are not allowed")
	case itemIdentifier, itemString, itemNumber:
		return p.parseVariable()
	case itemTagType:
		p.nextNonSpace()

		switch t.val {
		case "!":
			return p.parseComment()
		case "#":
			return p.parseSection(false)
		case "^":
			return p.parseSection(true)
		case ">":
			return p.parsePartial()
		case "<":
			return p.parseInherit()
		case "$":
			return p.parseDefine()
		case "/":
			return p.parseClose()
		}
	}

	return p.errorf("unexpected token: %s", t.val)
}

func (p *parser) parseVariable() Node {
	head, tail := p.parseExpression()

	if t := p.nextNonSpace(); t.typ != itemRightDelim {
		return p.errorf("unexpected token: %s", t.val)
	}

	return newVariable(head, tail)
}

func (p *parser) parseComment() Node {
	t := p.nextNonSpace()
	v := ""

	if t.typ == itemString {
		v = t.val
		t = p.nextNonSpace()
	}

	if t.typ == itemRightDelim {
		return newComment(v)
	}

	return p.errorf("unexpected token: %s", t.val)
}

func (p *parser) parseSection(inverted bool) Node {
	temp, tail := p.parseExpression()

	head, ok := temp.(*identifierNode)
	if !ok {
		return p.errorf("expression in section must start with identifier")
	}

	if t := p.nextNonSpace(); t.typ != itemRightDelim {
		return p.errorf("unexpected token: %s", t.val)
	}

	node := newSection(head, tail, inverted)

	if !p.parse(node) {
		return nil
	}

	return node
}

func (p *parser) parsePartial() Node {
	name := p.parseName()
	if name == "" {
		return nil
	}

	if t := p.nextNonSpace(); t.typ != itemRightDelim {
		return p.errorf("expected a delimiter, but got: %s", t.val)
	}

	return newPartial(name)
}

func (p *parser) parseDefine() Node {
	name := p.parseName()
	if name == "" {
		return nil
	}

	if t := p.nextNonSpace(); t.typ != itemRightDelim {
		return p.errorf("expected a delimiter, but got: %s", t.val)
	}

	node := newDefine(name)

	if !p.parse(node) {
		return nil
	}

	return node
}

func (p *parser) parseInherit() Node {
	name := p.parseName()
	if name == "" {
		return nil
	}

	if t := p.nextNonSpace(); t.typ != itemRightDelim {
		return p.errorf("expected a delimiter, but got: %s", t.val)
	}

	node := newInherit(name)

	if !p.parse(node) {
		return nil
	}

	return node
}

func (p *parser) parseClose() Node {
	name := p.parseName()
	if name == "" {
		return nil
	}

	if t := p.nextNonSpace(); t.typ != itemRightDelim {
		return p.errorf("expected a delimiter, but got: %s", t.val)
	}

	return newClose(name)
}

func (p *parser) parseExpression() (head Node, tail []Node) {
	t := p.peekNonSpace()

	switch t.typ {
	case itemIdentifier:
		head = p.parseIdentifier()
	case itemString:
		p.nextNonSpace()
		head = newString(t.val)
	case itemNumber:
		p.nextNonSpace()
		head = newNumber(t.val)
	}

	if _, ok := head.(*identifierNode); ok {
	Loop:
		for {
			t = p.peekNonSpace()

			switch t.typ {
			case itemIdentifier:
				tail = append(tail, p.parseIdentifier())
			case itemString:
				p.nextNonSpace()
				tail = append(tail, newString(t.val))
			case itemNumber:
				p.nextNonSpace()
				tail = append(tail, newNumber(t.val))
			default:
				break Loop
			}
		}
	}

	return
}

func (p *parser) parseIdentifier() *identifierNode {
	var s []string

Loop:
	for {
		t := p.peek()

		switch t.typ {
		case itemIdentifier:
			s = append(s, t.val)
		case itemDot:
			// continue
		default:
			break Loop
		}

		p.next()
	}

	return newIdentifier(s)
}

func (p *parser) parseName() (name string) {
	t := p.nextNonSpace()

	if t.typ != itemName {
		p.errorf("expected a delimiter, but got: %s", t.val)
	} else {
		name = t.val
	}

	return
}
