package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCountBranches(t *testing.T) {
	for _, tc := range []struct {
		name  string
		body  string
		count int
		init  int
		note  string
	}{
		{"decoder", `if value, ok, err := decode(); err != nil { return err } else if ok { result = value }`, 2, 1, "; else"},
		{"assignment", `if value < minimum { value = minimum }`, 1, 0, "value < minimum"},
		{"switch", `switch kind { case 1: if ok { return nil } }`, 1, 0, "ok"},
		{"closure", `if ok { run(func() { if nested { return } }) }`, 1, 0, "ok"},
		{"break", `for { if done { break } }`, 1, 0, "break; preserve target"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "sample.go", "package sample\nfunc sample() {\n"+tc.body+"\n}", 0)
			require.NoError(t, err)
			report := &funcReport{}
			countBranches(fset, file.Decls[0].(*ast.FuncDecl).Body, report)
			assert.Equal(t, tc.count, report.Branches)
			assert.Equal(t, tc.init, report.Inits)
			assert.Contains(t, strings.Join(report.Notes, "\n"), tc.note)
			assert.Contains(t, report.Notes[0], "line 3:")
		})
	}
}
