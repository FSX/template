// The lexer is based on Go's template lexer in text/template.
// All lexing functions have been copied without modifications and
// some of the state functions have been modified.
//
// Most of this code:
//
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file in
// the Go source.

package template

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Pos int

func (p Pos) Position() Pos {
	return p
}

// item represents a token or text string returned from the scanner.
type item struct {
	typ itemType // The type of this item.
	pos Pos      // The starting position, in bytes, of this item in the input string.
	val string   // The value of this item.
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}

	return fmt.Sprintf("%q", i.val)
}

func (i item) Type() string {
	var s string

	switch i.typ {
	case itemError:
		s = "itemError"
	case itemEOF:
		s = "itemEOF"
	case itemLeftDelim:
		s = "itemLeftDelim"
	case itemRightDelim:
		s = "itemRightDelim"
	case itemText:
		s = "itemText"
	case itemTagType:
		s = "itemTagType"
	case itemIdentifier:
		s = "itemIdentifier"
	case itemDot:
		s = "itemDot"
	case itemName:
		s = "itemName"
	case itemSpace:
		s = "itemSpace"
	case itemString:
		s = "itemString"
	case itemComplex:
		s = "itemComplex"
	case itemNumber:
		s = "itemNumber"
	default:
		s = "Unknown"
	}

	return s
}

// itemType identifies the type of lex items.
type itemType int

const (
	itemError      itemType = iota // Error occurred; value is text of error
	itemEOF                        // End of file
	itemLeftDelim                  // Left action delimiter
	itemRightDelim                 // Right action delimiter
	itemText                       // Plain text
	itemTagType                    // Defines the type of a tag
	itemIdentifier                 // Alphanumeric identifier
	itemDot                        // A dot
	itemName                       // A name
	itemSpace                      // Run of spaces separating arguments
	itemString                     // A text string
	itemComplex                    // complex constant (1+2i); imaginary is just a number
	itemNumber                     // simple number, including imaginary
)

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name       string    // the name of the input; used only for error reports
	input      string    // the string being scanned
	leftDelim  string    // start of action
	rightDelim string    // end of action
	state      stateFn   // the next lexing function to enter
	pos        Pos       // current position in the input
	start      Pos       // start position of this item
	width      Pos       // width of last rune read from input
	lastPos    Pos       // position of most recent item returned by nextItem
	items      chan item // channel of scanned items
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// lineNumber reports which line we're on, based on the position of
// the previous item returned by nextItem. Doing it this way
// means we don't have to worry about peek double counting.
func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.lastPos], "\n")
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next item from the input.
func (l *lexer) nextItem() item {
	item := <-l.items
	l.lastPos = item.pos
	return item
}

// scanNumber scans a number.
func (l *lexer) scanNumber() bool {
	// Optional leading sign.
	l.accept("+-")

	// Is it hex?
	digits := "0123456789"
	if l.accept("0") && l.accept("xX") {
		digits = "0123456789abcdefABCDEF"
	}

	l.acceptRun(digits)

	if l.accept(".") {
		l.acceptRun(digits)
	}

	if l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789")
	}

	// Is it imaginary?
	l.accept("i")

	// Next thing mustn't be alphanumeric.
	if isAlphaNumeric(l.peek()) {
		l.next()
		return false
	}
	return true
}

// lex creates a new scanner for the input string.
func lex(name, input, left, right string) *lexer {
	if left == "" {
		left = leftDelim
	}
	if right == "" {
		right = rightDelim
	}
	l := &lexer{
		name:       name,
		input:      input,
		leftDelim:  left,
		rightDelim: right,
		items:      make(chan item),
	}
	go l.run()
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for l.state = lexText; l.state != nil; {
		l.state = l.state(l)
	}
}

// State functions

const (
	leftDelim  = "(("
	rightDelim = "))"
)

var (
	lexSpaceExpr stateFn
	lexSpaceName stateFn
)

func init() {
	// NOTE: Functions initialized here to avoid an initialization loop.
	lexSpaceExpr = makeLexSpace(lexExpressionTag)
	lexSpaceName = makeLexSpace(lexNameTag)
}

func makeLexSpace(nextState stateFn) stateFn {
	return func(l *lexer) stateFn {
		for isSpace(l.peek()) {
			l.next()
		}

		l.emit(itemSpace)
		return nextState
	}
}

func lexText(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], l.leftDelim) {
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexLeftDelim
		}
		if l.next() == eof {
			break
		}
	}

	// Correctly reached EOF.
	if l.pos > l.start {
		l.emit(itemText)
	}
	l.emit(itemEOF)

	return nil
}

func lexLeftDelim(l *lexer) stateFn {
	l.pos += Pos(len(l.leftDelim))
	l.emit(itemLeftDelim)

	return lexTag
}

func lexRightDelim(l *lexer) stateFn {
	l.pos += Pos(len(l.rightDelim))
	l.emit(itemRightDelim)
	return lexText
}

func lexTag(l *lexer) stateFn {
	r := l.next()

	switch r {
	case '#', '^':
		l.emit(itemTagType)
		return lexExpressionTag
	case '<', '>', '/', '$':
		l.emit(itemTagType)
		return lexNameTag
	case '!':
		l.emit(itemTagType)
		return lexComment
	}

	// Possibly a normal variable.
	l.backup()
	return lexExpressionTag
}

func lexComment(l *lexer) stateFn {
	n := strings.Index(l.input[l.pos:], l.leftDelim)
	if n < 0 {
		n = len(l.input)
	} else {
		n = int(l.pos) + n
	}

	i := strings.Index(l.input[l.pos:n], l.rightDelim)

	if i < 0 {
		return l.errorf("unclosed comment")
	}

	// MAYBE: Consume leading and trailing whitespace of string?
	l.pos += Pos(i)
	l.emit(itemString)

	return lexRightDelim
}

func lexExpressionTag(l *lexer) stateFn {
	// NOTE: An expression tag (except variable tags) should always start
	// with an identifier, not a string or number (both not implemented yet),
	// but I (FSX) can't figure out a simple way of enforcing this. This will
	// be handled in the parser.
	//
	// The reason for this rule is that closing tags also have a name.
	// It's possible to use a string or number as its name, but it's
	// easier to just use an indentifier. A closing tag will be handled
	// as a identifier tag (lexIdentifierTag).

	if strings.HasPrefix(l.input[l.pos:], l.rightDelim) {
		return lexRightDelim
	}

	r := l.next()

	switch {
	case r == eof || isEndOfLine(r):
		return l.errorf("unclosed tag")
	case isSpace(r):
		return lexSpaceExpr
	case isAlpha(r):
		l.backup()
		return lexIdentifier
	case isNumeric(r), r == '-', r == '+':
		l.backup()
		return lexNumber
	case r == '"':
		l.ignore()
		return lexString
	}

	return l.errorf("unrecognized character in tag: %#U", r)
}

func lexIdentifier(l *lexer) stateFn {
Loop:
	for {
		switch r := l.peek(); {
		case isAlphaNumeric(r):
			l.next()
		case r == '.':
			l.emit(itemIdentifier)
			l.next()
			l.emit(itemDot)

			if s := l.peek(); !isAlphaNumeric(s) {
				return l.errorf("unrecognized character in identifier: %#U", r)
			}
		default:
			l.emit(itemIdentifier)
			break Loop
		}
	}

	return lexExpressionTag
}

func lexNumber(l *lexer) stateFn {
	if !l.scanNumber() {
		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
	}

	if sign := l.peek(); sign == '+' || sign == '-' {
		// Complex: 1+2i. No spaces, must end in 'i'.
		if !l.scanNumber() || l.input[l.pos-1] != 'i' {
			return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
		}

		l.emit(itemComplex)
	} else {
		l.emit(itemNumber)
	}

	return lexExpressionTag
}

func lexString(l *lexer) stateFn {
Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.errorf("unterminated quoted string")
		case '"':
			break Loop
		}
	}

	// This might look a bit weird, but we don't want to have
	// the quote in the string.
	l.backup()
	l.emit(itemString)
	l.next()
	l.ignore()

	return lexExpressionTag
}

func lexNameTag(l *lexer) stateFn {
	if strings.HasPrefix(l.input[l.pos:], l.rightDelim) {
		return lexRightDelim
	}

	r := l.next()

	switch {
	case r == eof || isEndOfLine(r):
		return l.errorf("unclosed tag")
	case isSpace(r):
		return lexSpaceName
	case isAlpha(r):
		l.backup()
		return lexName
	}

	return l.errorf("unrecognized character in tag: %#U", r)
}

func lexName(l *lexer) stateFn {
	if r := l.next(); !isAlpha(r) {
		return l.errorf("name must begin with a letter, but got: %#U", r)
	}

	for {
		r := l.next()
		if !isAlphaNumeric(r) && r != '.' && r != '/' {
			break
		}
	}

	l.backup()
	l.emit(itemName)

	return lexNameTag
}

func isAlpha(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isNumeric(r rune) bool {
	return unicode.IsDigit(r)
}

func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}
