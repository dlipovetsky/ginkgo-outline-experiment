package outline

import (
	"encoding/json"
	"go/ast"
	"go/token"
	"strconv"

	"golang.org/x/tools/go/ast/inspector"
)

type GinkgoMetadata struct {
	SpecType     string
	CodeLocation string
	Text         string
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
			GinkgoMetadata: GinkgoMetadata{
				SpecType: "root",
			},
		},
	}
}

func (o *Outline) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.root)
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
						SpecType:     i.Name,
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
