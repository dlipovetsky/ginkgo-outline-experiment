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
	Name         string `json:"name"`
	CodeLocation string `json:"codeLocation"`
	Text         string `json:"text"`

	Spec    bool `json:"spec"`
	Focused bool `json:"focused"`
	Pending bool `json:"pending"`
}

type GinkgoNode struct {
	GinkgoMetadata
	Children []*GinkgoNode `json:"children,omitempty"`
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

func FromASTFile(fset *token.FileSet, src *ast.File) (*Outline, error) {
	ispr := inspector.New([]*ast.File{src})

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
			if parent.Pending {
				gn.Pending = true
			}
			// TODO: Update focused based on ginkgo behavior:
			// > Nested programmatically focused specs follow a simple rule: if
			// > a leaf-node is marked focused, any of its ancestor nodes that
			// > are marked focus will be unfocused.
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
	if o.root != nil {
		return json.Marshal(o.root.Children)
	}
	return nil, nil
}
