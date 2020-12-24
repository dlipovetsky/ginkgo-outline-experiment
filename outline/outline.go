package outline

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"strconv"

	"golang.org/x/tools/go/ast/inspector"
)

const (
	UndefinedTextAlt = "undefined"
)

type GinkgoMetadata struct {
	Name         string
	CodeLocation string
	Text         string

	Spec    bool
	Focused bool
	Pending bool
}

type GinkgoNode struct {
	GinkgoMetadata
	Children []*GinkgoNode
}

func GinkgoNodeFromCallExpr(ce *ast.CallExpr, fset *token.FileSet) (*GinkgoNode, bool) {
	id, ok := ce.Fun.(*ast.Ident)
	if !ok {
		return nil, false
	}

	n := GinkgoNode{}
	n.Name = id.Name
	n.CodeLocation = fset.Position(ce.Pos()).String()
	switch id.Name {
	case "It", "Measure", "Specify":
		n.Spec = true
		n.Text = GinkgoTextOrAltFromCallExpr(ce, UndefinedTextAlt)
	case "FIt", "FMeasure", "FSpecify":
		n.Spec = true
		n.Focused = true
		n.Text = GinkgoTextOrAltFromCallExpr(ce, UndefinedTextAlt)
	case "PIt", "PMeasure", "PSpecify", "XIt", "XMeasure", "XSpecify":
		n.Spec = true
		n.Pending = true
		n.Text = GinkgoTextOrAltFromCallExpr(ce, UndefinedTextAlt)
	case "Context", "Describe", "When":
		n.Text = GinkgoTextOrAltFromCallExpr(ce, UndefinedTextAlt)
	case "FContext", "FDescribe", "FWhen":
		n.Focused = true
		n.Text = GinkgoTextOrAltFromCallExpr(ce, UndefinedTextAlt)
	case "PContext", "PDescribe", "PWhen", "XContext", "XDescribe", "XWhen":
		n.Pending = true
		n.Text = GinkgoTextOrAltFromCallExpr(ce, UndefinedTextAlt)
	case "By":
	case "AfterEach", "BeforeEach":
	case "JustAfterEach", "JustBeforeEach":
	case "AfterSuite", "BeforeSuite":
	case "SynchronizedAfterSuite", "SynchronizedBeforeSuite":
	default:
		return nil, false
	}
	return &n, true
}

func GinkgoTextOrAltFromCallExpr(ce *ast.CallExpr, alt string) string {
	text, defined := TextFromCallExpr(ce)
	if !defined {
		return alt
	}
	return text
}

func TextFromCallExpr(ce *ast.CallExpr) (string, bool) {
	if len(ce.Args) < 1 {
		return "", false
	}
	text, ok := ce.Args[0].(*ast.BasicLit)
	if !ok {
		return "", false
	}
	switch text.Kind {
	case token.CHAR, token.STRING:
		// For token.CHAR and token.STRING, Value is quoted
		unquoted, err := strconv.Unquote(text.Value)
		if err != nil {
			// If unquoting fails, just use the raw Value
			return text.Value, true
		}
		return unquoted, true
	default:
		return text.Value, true
	}
}

func FromASTFiles(fset *token.FileSet, src ...*ast.File) (*Outline, error) {
	ispr := inspector.New(src)

	outline := Outline{
		root: &GinkgoNode{},
	}
	stack := []*GinkgoNode{outline.root}
	ispr.Nodes([]ast.Node{(*ast.CallExpr)(nil)}, func(node ast.Node, push bool) bool {
		ce, ok := node.(*ast.CallExpr)
		if !ok {
			panic(fmt.Errorf("node is not an *ast.CallExpr: %s", fset.Position(node.Pos())))
		}
		gn, ok := GinkgoNodeFromCallExpr(ce, fset)
		if !ok {
			// Not a Ginkgo call, continue
			return true
		}

		// Visiting this node on the way down
		if push {
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, gn)

			stack = append(stack, gn)
			return true
		}
		// Visiting node on the way up
		stack = stack[0 : len(stack)-1]
		return true
	})
	return &outline, nil
}

type Outline struct {
	root *GinkgoNode
}

func (o *Outline) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.root.Children)
}
