package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strconv"

	"golang.org/x/tools/go/ast/inspector"
)

func main() {
	// if len(os.Args) < 2 {
	// 	log.Fatalf("usage: %s FILE", os.Args[0])
	// }
	// filename := os.Args[1]

	src := `package p

import (
	"fmt"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("group 1", func() {
	Context("context 1", func() {
		It("should test 1.1", func() {
		})
		It("should test 1.2", func() {
		})
	})
	Context("context 2", func() {
		It("should test 2.1", func() {
		})
		It(fmt.Sprintf("should test %d.%d", 2, 2), func() {
		})
	})
})

`
	filename := "src.go"

	fset := token.NewFileSet()

	var f *ast.File
	var err error

	f, err = parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	g, err := parser.ParseFile(fset, "src_2.go", src, 0)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	ispr := inspector.New([]*ast.File{f, g})

	type GinkgoMetadata struct {
		SpecType     string
		CodeLocation string
		Text         string
	}

	type GinkgoNode struct {
		GinkgoMetadata
		Children []*GinkgoNode
	}

	root := GinkgoNode{
		GinkgoMetadata: GinkgoMetadata{
			SpecType: "root",
		},
	}

	stack := []*GinkgoNode{&root}

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
					child.Text = "[could not determine]"
					if text, ok := c.Args[0].(*ast.BasicLit); ok {
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

	fmt.Println("input:")
	fmt.Print(src)

	fmt.Println("output:")
	b, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
