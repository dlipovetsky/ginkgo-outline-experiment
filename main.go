package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"

	"golang.org/x/tools/go/ast/inspector"
)

func main() {
	// if len(os.Args) < 2 {
	// 	log.Fatalf("usage: %s FILE", os.Args[0])
	// }
	// filename := os.Args[1]

	src := `
package p

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("group 1", func() {
	Context("context 1", func() {
		It("should test 1.1", func() {
		})
		It("should test 1.2", func() {
		})
	})
	Context("group 2", func() {
		It("should test 2.1", func() {
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

	ispr := inspector.New([]*ast.File{f})

	ginkgoNodes := []ast.Node{
		(*ast.CallExpr)(nil),
	}
	ispr.Nodes(ginkgoNodes, func(n ast.Node, push bool) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if i, ok := c.Fun.(*ast.Ident); ok {
				fmt.Println(i.Name)
			}
			// Figuring out the "name" of each Spec is a pain, because the it's not necessarily a string literal...
			// I could handle string literals, and error out on everything else, though.
		}
		return true
	})

	// ast.Inspect(f, func(n ast.Node) bool {
	// 	switch x := n.(type) {
	// 	case *ast.CallExpr:
	// 		switch c := x.Fun.(type) {
	// 		case *ast.Ident:
	// 			switch c.Name {
	// 			case "It":
	// 				fmt.Println(c.Pos())
	// 				return false
	// 			// TODO handle other ginkgo test expressions
	// 			default:
	// 				return true
	// 			}
	// 		default:
	// 			return true
	// 		}
	// 	}
	// 	return true
	// })

}
