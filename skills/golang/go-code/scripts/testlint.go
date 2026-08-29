// testlint audits Go test naming and file layout.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type testFunc struct {
	name  string
	file  string
	line  int
	words int
}

type fileIssue struct {
	path   string
	reason string
}

func main() {
	root := flag.String("root", ".", "repository root to scan")
	maxWords := flag.Int("max-words", 3, "max camelCase words after Test prefix")
	maxLen := flag.Int("max-len", 40, "max rune length of full test function name")
	showAll := flag.Bool("all", false, "print every test name, not just violations")
	strictPair := flag.Bool("strict-pair", true, "require foo_test.go to have foo.go in the same directory")
	flag.Parse()

	absRoot, err := filepath.Abs(*root)
	if err != nil {
		fatal(err)
	}

	tests, orphans, unpaired, untested, manyFuncs, err := scan(absRoot, *strictPair)
	if err != nil {
		fatal(err)
	}

	sort.Slice(tests, func(i, j int) bool {
		if tests[i].file != tests[j].file {
			return tests[i].file < tests[j].file
		}
		return tests[i].line < tests[j].line
	})

	var violations []testFunc
	for _, tf := range tests {
		if tf.words > *maxWords || len(tf.name) > *maxLen {
			violations = append(violations, tf)
		}
	}

	fmt.Printf("scanned %d test functions in %s\n\n", len(tests), absRoot)

	if *showAll {
		fmt.Println("=== all tests ===")
		for _, tf := range tests {
			fmt.Printf("%4d words  %3d chars  %s:%d  %s\n", tf.words, len(tf.name), tf.file, tf.line, tf.name)
		}
		fmt.Println()
	}

	if len(violations) > 0 {
		fmt.Printf("=== long test names (%d, want Test + <=%d words, <=%d chars) ===\n", len(violations), *maxWords, *maxLen)
		for _, tf := range violations {
			fmt.Printf("  %4d words  %3d chars  %s:%d  %s\n", tf.words, len(tf.name), tf.file, tf.line, tf.name)
		}
		fmt.Println()
	} else {
		fmt.Println("=== long test names: none ===")
		fmt.Println()
	}

	if len(unpaired) > 0 {
		fmt.Printf("=== unpaired _test.go files (%d, missing matching .go stem) ===\n", len(unpaired))
		for _, issue := range unpaired {
			fmt.Printf("  %s\n    %s\n", issue.path, issue.reason)
		}
		fmt.Println()
	} else if *strictPair {
		fmt.Println("=== unpaired _test.go files: none ===")
		fmt.Println()
	}

	if len(untested) > 0 {
		fmt.Printf("=== unpaired .go files (%d, missing matching _test.go stem) ===\n", len(untested))
		for _, issue := range untested {
			fmt.Printf("  %s\n    %s\n", issue.path, issue.reason)
		}
		fmt.Println()
	} else {
		fmt.Println("=== unpaired .go files: none ===")
		fmt.Println()
	}

	if len(orphans) > 0 {
		fmt.Printf("=== orphan test directories (%d, no non-test .go files) ===\n", len(orphans))
		for _, issue := range orphans {
			fmt.Printf("  %s\n    %s\n", issue.path, issue.reason)
		}
		fmt.Println()
	} else {
		fmt.Println("=== orphan test directories: none ===")
		fmt.Println()
	}

	if len(manyFuncs) > 0 {
		fmt.Println("=== candidates for table tests (>=5 top-level Test funcs, no t.Run) ===")
		for _, issue := range manyFuncs {
			fmt.Printf("  %s\n    %s\n", issue.path, issue.reason)
		}
		fmt.Println()
	}

	exit := 0
	if len(violations) > 0 || len(unpaired) > 0 || len(untested) > 0 || len(orphans) > 0 {
		exit = 1
	}
	os.Exit(exit)
}

func scan(root string, strictPair bool) ([]testFunc, []fileIssue, []fileIssue, []fileIssue, []fileIssue, error) {
	var tests []testFunc
	var orphans, unpaired, untested, manyFuncs []fileIssue

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "vendor", "node_modules", "dist", "testdata", ".agents", "tools":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.Contains(path, string(filepath.Separator)+"scripts"+string(filepath.Separator)) {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		dir := filepath.Dir(path)
		if !strings.HasSuffix(path, "_test.go") {
			stem := strings.TrimSuffix(filepath.Base(path), ".go")
			if !strings.Contains(stem, "_") {
				paired := filepath.Join(dir, stem+"_test.go")
				if _, err := os.Stat(paired); err != nil {
					untested = append(untested, fileIssue{
						path:   rel,
						reason: fmt.Sprintf("missing paired test file %s", filepath.ToSlash(filepath.Join(filepath.Dir(rel), stem+"_test.go"))),
					})
				}
			}
			return nil
		}

		if !dirHasSourceGo(dir) {
			orphans = append(orphans, fileIssue{
				path:   rel,
				reason: "directory has no non-test .go source files",
			})
		}

		if strictPair {
			stem := strings.TrimSuffix(filepath.Base(path), "_test.go")
			paired := filepath.Join(dir, stem+".go")
			if _, err := os.Stat(paired); err != nil {
				unpaired = append(unpaired, fileIssue{
					path:   rel,
					reason: fmt.Sprintf("missing paired source file %s", filepath.ToSlash(filepath.Join(filepath.Dir(rel), stem+".go"))),
				})
			}
		}

		fileTests, subtests, err := parseTests(path)
		if err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		for _, tf := range fileTests {
			tf.file = rel
			tests = append(tests, tf)
		}
		if len(fileTests) >= 5 && subtests == 0 {
			manyFuncs = append(manyFuncs, fileIssue{
				path:   rel,
				reason: fmt.Sprintf("%d top-level Test funcs, 0 t.Run subtests; consider a table test", len(fileTests)),
			})
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	sort.Slice(orphans, func(i, j int) bool { return orphans[i].path < orphans[j].path })
	sort.Slice(unpaired, func(i, j int) bool { return unpaired[i].path < unpaired[j].path })
	sort.Slice(untested, func(i, j int) bool { return untested[i].path < untested[j].path })
	sort.Slice(manyFuncs, func(i, j int) bool { return manyFuncs[i].path < manyFuncs[j].path })
	return tests, orphans, unpaired, untested, manyFuncs, nil
}

func dirHasSourceGo(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
			return true
		}
	}
	return false
}

func parseTests(path string) ([]testFunc, int, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, 0, err
	}

	var tests []testFunc
	subtests := 0

	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel == nil || sel.Sel.Name != "Run" {
			return true
		}
		id, ok := sel.X.(*ast.Ident)
		if !ok || id.Name != "t" {
			return true
		}
		subtests++
		return true
	})

	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name == nil || fn.Recv != nil {
			continue
		}
		name := fn.Name.Name
		if !strings.HasPrefix(name, "Test") || name == "Test" {
			continue
		}
		suffix := strings.TrimPrefix(name, "Test")
		words := len(camelWords(suffix))
		tests = append(tests, testFunc{
			name:  name,
			line:  fset.Position(fn.Pos()).Line,
			words: words,
		})
	}
	return tests, subtests, nil
}

func camelWords(s string) []string {
	if s == "" {
		return nil
	}
	runes := []rune(s)
	var words []string
	start := 0
	for i := 1; i < len(runes); i++ {
		if !unicode.IsUpper(runes[i]) {
			continue
		}
		if unicode.IsLower(runes[i-1]) {
			words = append(words, string(runes[start:i]))
			start = i
			continue
		}
		if i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
			words = append(words, string(runes[start:i]))
			start = i
		}
	}
	words = append(words, string(runes[start:]))
	return words
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "testlint:", err)
	os.Exit(2)
}
