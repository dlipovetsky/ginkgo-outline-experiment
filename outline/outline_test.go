package outline

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFromASTFile(t *testing.T) {
	tcs := []struct {
		srcFile     string
		outlineFile string
	}{
		{
			srcFile:     "normal_test.go",
			outlineFile: "normal_test_outline.json",
		},
		// TODO: enable tests
		// {
		// 	srcFile:     "nodot_test.go",
		// 	outlineFile: "normal_test_outline.json",
		// },
		// {
		// 	srcFile:     "alias_test.go",
		// 	outlineFile: "normal_test_outline.json",
		// },
		{
			srcFile:     "focused_test.go",
			outlineFile: "focused_test_outline.json",
		},
		{
			srcFile:     "pending_test.go",
			outlineFile: "pending_test_outline.json",
		},
		{
			srcFile:     "suite_test.go",
			outlineFile: "suite_test_outline.json",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.srcFile, func(t *testing.T) {
			fset := token.NewFileSet()
			astFile, err := parser.ParseFile(fset, filepath.Join("testdata", tc.srcFile), nil, 0)
			if err != nil {
				log.Fatalf("error parsing source: %s", err)
			}

			o, err := FromASTFile(fset, astFile)
			if err != nil {
				t.Fatalf("error creating outline %s", err)
			}

			got, err := json.MarshalIndent(o, "", "  ")
			if err != nil {
				log.Fatalf("error marshalling outline to json: %s", err)
			}
			want, err := ioutil.ReadFile(filepath.Join("testdata", tc.outlineFile))
			if err != nil {
				log.Fatalf("error reading outline fixture: %s", err)
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Fatalf("output mismatch (-want, +got):\n%s", diff)
			}
		})
	}
}
