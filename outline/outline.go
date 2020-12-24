package outline

import (
	"encoding/json"
	"go/ast"
	"go/token"
	"strconv"

	"golang.org/x/tools/go/ast/inspector"
)

const (
	UndefinedAltText = "undefined"
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

type Outline struct {
	root *GinkgoNode
}

func New() *Outline {
	return &Outline{
		root: &GinkgoNode{
			GinkgoMetadata: GinkgoMetadata{},
		},
	}
}

func (o *Outline) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.root)
}

type GinkgoVisitor struct {
	fset    *token.FileSet
	astFile *ast.File
	stack   []*GinkgoNode
}

func (v *GinkgoVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		// pop off stack
		v.stack = v.stack[0 : len(v.stack)-1]
		return v
	}

	// process node
	ce, ok := node.(*ast.CallExpr)
	if !ok {
		return v
	}
	n, ok := v.GinkgoNodeFromCallExpr(ce)
	if !ok {
		return v
	}

	// add to parent
	parent := v.stack[len(v.stack)-1]
	parent.Children = append(parent.Children, n)

	// push onto stack
	v.stack = append(v.stack, n)
	return v
}

func (v *GinkgoVisitor) GinkgoNodeFromCallExpr(ce *ast.CallExpr) (*GinkgoNode, bool) {
	id, ok := ce.Fun.(*ast.Ident)
	if !ok {
		return nil, false
	}

	n := GinkgoNode{}
	n.Name = id.Name
	n.CodeLocation = v.fset.Position(ce.Pos()).String()
	switch id.Name {
	case "It", "Measure", "Specify":
		n.Spec = true
		n.Text = GinkgoTextOrAltFromCallExpr(ce, UndefinedAltText)
	case "FIt", "FMeasure", "FSpecify":
		n.Spec = true
		n.Focused = true
		n.Text = GinkgoTextOrAltFromCallExpr(ce, UndefinedAltText)
	case "PIt", "PMeasure", "PSpecify", "XIt", "XMeasure", "XSpecify":
		n.Spec = true
		n.Pending = true
		n.Text = GinkgoTextOrAltFromCallExpr(ce, UndefinedAltText)
	case "Context", "Describe", "When":
		n.Text = GinkgoTextOrAltFromCallExpr(ce, UndefinedAltText)
	case "FContext", "FDescribe", "FWhen":
		n.Focused = true
		n.Text = GinkgoTextOrAltFromCallExpr(ce, UndefinedAltText)
	case "PContext", "PDescribe", "PWhen", "XContext", "XDescribe", "XWhen":
		n.Pending = true
		n.Text = GinkgoTextOrAltFromCallExpr(ce, UndefinedAltText)
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
	unquoted, err := strconv.Unquote(text.Value)
	if err != nil {
		return text.Value, true
	}
	return unquoted, true
}

func FromASTFile(astFile *ast.File, fset *token.FileSet) (*Outline, error) {
	outline := Outline{
		root: &GinkgoNode{},
	}
	visitor := &GinkgoVisitor{
		astFile: astFile,
		fset:    fset,
		stack:   []*GinkgoNode{outline.root},
	}
	ast.Walk(visitor, visitor.astFile)
	return &outline, nil
}

func FromASTFiles(fset *token.FileSet, src ...*ast.File) (*Outline, error) {
	ispr := inspector.New(src)

	o := New()
	stack := []*GinkgoNode{o.root}
	ispr.Nodes([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node, push bool) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if i, ok := c.Fun.(*ast.Ident); ok {
				// TODO return immediately if identifer is not a ginkgo spec/container
				child := GinkgoNode{
					GinkgoMetadata: GinkgoMetadata{
						Name:         i.Name,
						CodeLocation: fset.Position(i.Pos()).String(),
					},
				}
				if len(c.Args) > 0 {
					if text, ok := c.Args[0].(*ast.BasicLit); ok {
						// TODO: inspect text.Kind, maybe use UnquoteChar() as well?
						unquoted, err := strconv.Unquote(text.Value)
						if err != nil {
							panic(err)
						}
						child.Text = unquoted
					}
				}

				if push {
					// add to parent
					parent := stack[len(stack)-1]
					parent.Children = append(parent.Children, &child)

					// push onto stack
					stack = append(stack, &child)
					return true
				}
				// pop off stack
				stack = stack[0 : len(stack)-1]
				return true
			}
		}
		return true
	})
	return o, nil
}
