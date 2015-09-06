// The tests is based on Go's template lexer tests in text/template.
//
// Most of this code:
//
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file in
// the Go source.

package template

import "testing"

var (
	tEOF   = item{itemEOF, 0, ""}
	tLeft  = item{itemLeftDelim, 0, "(("}
	tRight = item{itemRightDelim, 0, "))"}
	tSpace = item{itemSpace, 0, " "}
	tDot   = item{itemDot, 0, "."}
)

type lexTest struct {
	name  string
	input string
	items []item
}

var lexTests = []lexTest{
	{"empty", "", []item{tEOF}},
	{"spaces", " \t\n", []item{{itemText, 0, " \t\n"}, tEOF}},
	{"text", `now is the time`, []item{{itemText, 0, "now is the time"}, tEOF}},
	{"comment", "((! This is a comment))", []item{
		tLeft,
		{itemTagType, 0, "!"},
		{itemString, 0, " This is a comment"},
		tRight,
		tEOF,
	}},
	{"unclosed-comment", "((! This is a comment", []item{
		tLeft,
		{itemTagType, 0, "!"},
		{itemError, 0, "unclosed comment"},
	}},
	{"variable", "((variable_or_function))", []item{
		tLeft,
		{itemIdentifier, 0, "variable_or_function"},
		tRight,
		tEOF,
	}},
	{"variable-with-fields", "((variable.or.function))", []item{
		tLeft,
		{itemIdentifier, 0, "variable"},
		tDot,
		{itemIdentifier, 0, "or"},
		tDot,
		{itemIdentifier, 0, "function"},
		tRight,
		tEOF,
	}},
	{"unclosed-variable", "((variable", []item{
		tLeft,
		{itemIdentifier, 0, "variable"},
		{itemError, 0, "unclosed tag"},
	}},
	{"section", "((#variable))", []item{
		tLeft,
		{itemTagType, 0, "#"},
		{itemIdentifier, 0, "variable"},
		tRight,
		tEOF,
	}},
	{"inverted-section", "((^variable))", []item{
		tLeft,
		{itemTagType, 0, "^"},
		{itemIdentifier, 0, "variable"},
		tRight,
		tEOF,
	}},
	{"numbers", "((1 02 0x14 -7.2i 1e3 +1.2e-4 4.2i 1+2i))", []item{
		tLeft,
		{itemNumber, 0, "1"},
		tSpace,
		{itemNumber, 0, "02"},
		tSpace,
		{itemNumber, 0, "0x14"},
		tSpace,
		{itemNumber, 0, "-7.2i"},
		tSpace,
		{itemNumber, 0, "1e3"},
		tSpace,
		{itemNumber, 0, "+1.2e-4"},
		tSpace,
		{itemNumber, 0, "4.2i"},
		tSpace,
		{itemComplex, 0, "1+2i"},
		tRight,
		tEOF,
	}},
	{"strings", `((variable "and a \"string\""))`, []item{
		tLeft,
		{itemIdentifier, 0, "variable"},
		tSpace,
		{itemString, 0, `and a \"string\"`},
		tRight,
		tEOF,
	}},
}

func collect(t *lexTest, left, right string) (items []item) {
	l := lex(t.name, t.input, left, right)

	for {
		item := l.nextItem()
		items = append(items, item)
		if item.typ == itemEOF || item.typ == itemError {
			break
		}
	}

	return
}

func equal(i1, i2 []item, checkPos bool) bool {
	if len(i1) != len(i2) {
		return false
	}
	for k := range i1 {
		if i1[k].typ != i2[k].typ {
			return false
		}
		if i1[k].val != i2[k].val {
			return false
		}
		if checkPos && i1[k].pos != i2[k].pos {
			return false
		}
	}
	return true
}

func TestLex(t *testing.T) {
	for _, test := range lexTests {
		items := collect(&test, "", "")
		if !equal(items, test.items, false) {
			t.Errorf("%s: got\n\t%+v\nexpected\n\t%v", test.name, items, test.items)
		}
	}
}

var (
	tLeftDelim  = item{itemLeftDelim, 0, "$$"}
	tRightDelim = item{itemRightDelim, 0, "@@"}
)

var lexDelimTests = []lexTest{
	{"empty-tag", `$$@@`, []item{tLeftDelim, tRightDelim, tEOF}},
	{"variable", `$$variable@@`, []item{
		tLeftDelim,
		{itemIdentifier, 0, "variable"},
		tRightDelim,
		tEOF,
	}},
}

func TestDelims(t *testing.T) {
	for _, test := range lexDelimTests {
		items := collect(&test, "$$", "@@")
		if !equal(items, test.items, false) {
			t.Errorf("%s: got\n\t%v\nexpected\n\t%v", test.name, items, test.items)
		}
	}
}

var lexPosTests = []lexTest{
	{"empty", "", []item{tEOF}},
	{"variable-with-field", "(( variable.field ))", []item{
		{itemLeftDelim, 0, "(("},
		{itemSpace, 2, " "},
		{itemIdentifier, 3, "variable"},
		{itemDot, 11, "."},
		{itemIdentifier, 12, "field"},
		{itemSpace, 17, " "},
		{itemRightDelim, 18, "))"},
		{itemEOF, 20, ""},
	}},
	{"text-and-tag", "0123((hello))xyz", []item{
		{itemText, 0, "0123"},
		{itemLeftDelim, 4, "(("},
		{itemIdentifier, 6, "hello"},
		{itemRightDelim, 11, "))"},
		{itemText, 13, "xyz"},
		{itemEOF, 16, ""},
	}},
}

// The other tests don't check position, to make the test
// cases easier to construct. This one does.
func TestPos(t *testing.T) {
	for _, test := range lexPosTests {
		items := collect(&test, "", "")
		if !equal(items, test.items, true) {
			t.Errorf("%s: got\n\t%v\nexpected\n\t%v", test.name, items, test.items)
			if len(items) == len(test.items) {
				// Detailed print; avoid item.String() to expose the position value.
				for i := range items {
					if !equal(items[i:i+1], test.items[i:i+1], true) {
						i1 := items[i]
						i2 := test.items[i]
						t.Errorf("\t#%d: got {%v %d %q} expected  {%v %d %q}", i, i1.typ, i1.pos, i1.val, i2.typ, i2.pos, i2.val)
					}
				}
			}
		}
	}
}

var benchmarkLexTmpl = `
((<base))
	((one "two" 3))

	((#one "two" 3 ))
		((self))
	((/one))

	((^one "two" 3 ))
		((self))
	((/one))

	((>one.two.three))

	((! This is a comment!))
((/base))
`

func BenchmarkLex(b *testing.B) {
	l := lex("benchmark", benchmarkLexTmpl, "", "")

	for {
		item := l.nextItem()
		if item.typ == itemEOF {
			break
		} else if item.typ == itemError {
			b.Fail()
		}
	}
}
