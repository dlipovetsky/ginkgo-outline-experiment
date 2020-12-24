package main

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"log"

	"daniel.lipovetsky.me/ginkgo-outline/outline"
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
	PContext("context 2", func() {
		It("should test 2.1", func() {
		})
		It(fmt.Sprintf("should test %d.%d", 2, 2), func() {
		})
	})
})

`
	filename := "src.go"

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		log.Fatalf("error parsing source: %s", err)
	}

	g, err := parser.ParseFile(fset, "src_2.go", src, 0)
	if err != nil {
		log.Fatalf("error parsing source: %s", err)
	}

	o, err := outline.FromASTFiles(fset, f, g)
	if err != nil {
		log.Fatalf("error building outline: %s", err)
	}

	// o, err := outline.FromASTFile(f, fset)
	// if err != nil {
	// 	log.Fatalf("error building outline: %s", err)
	// }

	fmt.Println("sources:")
	fmt.Print(src)

	fmt.Println("outline:")
	b, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		log.Fatalf("error marshalling outline to json: %s", err)
	}
	fmt.Println(string(b))
}
