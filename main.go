package main

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"os"

	"daniel.lipovetsky.me/ginkgo-outline-experiment/outline"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s FILE", os.Args[0])
	}
	filename := os.Args[1]

	fset := token.NewFileSet()

	astFile, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		log.Fatalf("error parsing source: %s", err)
	}

	o, err := outline.FromASTFile(fset, astFile)
	if err != nil {
		log.Fatalf("error building outline: %s", err)
	}

	fmt.Println("json:")
	b, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		log.Fatalf("error marshalling outline to json: %s", err)
	}
	fmt.Println(string(b))

	fmt.Println("csv:")
	fmt.Print(o)
}
