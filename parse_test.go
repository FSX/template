package template

import "testing"

type parseTest struct {
	name   string
	input  string
	ok     bool
	result string // what the user would see in an error message.
}

const (
	noError  = true
	hasError = false
)

var parseTests = []parseTest{
	{"empty", "", noError, ""},
	{"spaces", " \t\n", noError, ""},
	{"variable", `((test 1 "two" 3.14))`, noError, ""},
	{"section", `((#test))((/test))`, noError, ""},
	{"inverted-section", `((^test))((/test))`, noError, ""},
	{"inherit", `((<test))((/test))`, noError, ""},
	{"define", `(($test))((/test))`, noError, ""},
	{"comment", `((! comment))`, noError, ""},
	{"partial", `((>partial))`, noError, ""},
	{"incorrect-section", `((^3.14))((/3.14))`, hasError, "incorrect-section:1: expression in section must start with identifier"},
	{"unclosed-section", "((#test))", hasError, "unclosed-section:1: tag not closed"},
	{"close-tag", "((/test))", hasError, "close-tag:1: unexpected closing tag"},
	{"empty-tag", "(())", hasError, "empty-tag:1: empty tags are not allowed"},
	{"unknown", "((%test))", hasError, "unknown:1: unrecognized character in tag: U+0025 '%'"},
	{"unclosed", "((unclosed", hasError, "unclosed:1: unclosed tag"},
}

func TestParse(t *testing.T) {
	for _, test := range parseTests {
		_, err := Parse(test.name, "", "", test.input)

		if err != nil && test.ok {
			t.Errorf("%q: unexpected error: %v", test.name, err)
		} else if err != nil && !test.ok {
			if result := err.Error(); result != test.result {
				t.Errorf("%s=(%q): got\n\t%v\nexpected\n\t%v", test.name, test.input, result, test.result)
			}
		} else if err == nil && !test.ok {
			t.Errorf("%q: expected error; got none", test.name)
		}
	}
}

var benchmarkParseTmpl = `
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

func BenchmarkParse(b *testing.B) {
	_, err := Parse("benchmark", "", "", benchmarkParseTmpl)
	if err != nil {
		b.Error(err)
	}
}
