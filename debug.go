package template

import (
	"fmt"
	"strings"
)

func printNodes(node Node, level int) {
	s := strings.Repeat("    ", level)

	switch t := node.(type) {
	case (*listNode):
		fmt.Printf("%s(listNode)\n", s)
		for _, n := range t.Children() {
			printNodes(n, level+1)
		}
	case (*inheritNode):
		fmt.Printf("%s(inheritNode: %s)\n", s, t.Name())

		for _, n := range t.Children() {
			printNodes(n, level+1)
		}
	case (*defineNode):
		fmt.Printf("%s(defineNode: %s)\n", s, t.Name())

		for _, n := range t.Children() {
			printNodes(n, level+1)
		}
	case (*sectionNode):
		fmt.Printf("%s(sectionNode inverted=%t: %s)\n", s, t.Inverted, t.Name())

		printPipe(level + 1)
		printNodes(t.Head, 0)
		for _, n := range t.Tail {
			printPipe(level + 1)
			printNodes(n, 0)
		}
		for _, n := range t.Children() {
			printNodes(n, level+1)
		}
	case (*textNode):
		if len(t.Text) > 10 {
			fmt.Printf("%s(textNode: %.10q...)\n", s, t.Text)
		} else {
			fmt.Printf("%s(textNode: %q)\n", s, t.Text)
		}
	case (*commentNode):
		if len(t.Text) > 10 {
			fmt.Printf("%s(commentNode: %.10q...)\n", s, t.Text)
		} else {
			fmt.Printf("%s(commentNode: %q)\n", s, t.Text)
		}
	case (*variableNode):
		fmt.Printf("%s(variableNode)\n", s)

		printPipe(level + 1)
		printNodes(t.Head, 0)
		for _, n := range t.Tail {
			printPipe(level + 1)
			printNodes(n, 0)
		}
	case (*identifierNode):
		fmt.Printf("%s(identifierNode: %s)\n", s, t.Name())
	case (*stringNode):
		if len(t.Text) > 10 {
			fmt.Printf("%s(stringNode: %.10q...)\n", s, t.Text)
		} else {
			fmt.Printf("%s(stringNode: %q)\n", s, t.Text)
		}
	case (*numberNode):
		fmt.Printf("%s(numberNode: %s)\n", s, t.Text)
	case (*partialNode):
		fmt.Printf("%s(partialNode: %s)\n", s, t.Name())
	}
}

func printPipe(level int) {
	s := strings.Repeat("    ", level)
	fmt.Printf("%s| ", s)
}
