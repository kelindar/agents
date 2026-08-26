package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"sort"
)

func runEval(args []string) int {
	fs := flag.NewFlagSet("eval", flag.ExitOnError)
	registerConfigFlags(fs)

	path := "types.md"
	if len(args) > 0 && args[0][0] != '-' {
		path = args[0]
		args = args[1:]
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	rc := parseRuntimeConfig(fs)

	repoRoot, err := gitRoot()
	if err != nil {
		exitErr(err)
	}

	code, err := scanRepo(repoRoot, rc.includeInternal, rc.packagePrefixes)
	if err != nil {
		exitErr(err)
	}

	docPath := path
	if !filepath.IsAbs(docPath) {
		docPath = filepath.Join(repoRoot, path)
	}
	doc, err := parseTypesMD(docPath)
	if err != nil {
		exitErr(fmt.Errorf("parse %s: %w", path, err))
	}

	var missingAll, staleAll, noDoc, noCode []string
	for _, pair := range matchPackages(code, doc) {
		c, d := pair[0], pair[1]
		if c == nil && d != nil {
			noCode = append(noCode, fmt.Sprintf("`%s` (%s)", d.ShortName, d.ImportPath))
			continue
		}
		if c != nil && d == nil {
			noDoc = append(noDoc, fmt.Sprintf("`%s` (%s)", c.ShortName, c.ImportPath))
			missingAll = append(missingAll, formatSymbolList(c, setKeys(c.Symbols))...)
			continue
		}
		missing, stale := diffSets(c.Symbols, d.Symbols)
		missingAll = append(missingAll, formatSymbolList(c, missing)...)
		staleAll = append(staleAll, formatSymbolList(c, stale)...)
	}

	ok := len(missingAll) == 0 && len(staleAll) == 0 && len(noDoc) == 0 && len(noCode) == 0
	printEvalReport(path, missingAll, staleAll, noDoc, noCode)
	if ok {
		fmt.Println("\nOK — types.md matches exported API.")
		return 0
	}
	fmt.Println("\nFAIL — update types.md.")
	return 1
}

func printEvalReport(path string, missing, stale, noDoc, noCode []string) {
	fmt.Printf("# types.md eval: %s\n\n", path)
	printEvalSection("Missing from types.md", missing)
	printEvalSection("Stale in types.md (not in code)", stale)
	printEvalSection("Packages in code, no section in types.md", noDoc)
	printEvalSection("Packages in types.md, not found in code", noCode)
}

func printEvalSection(title string, items []string) {
	fmt.Printf("## %s\n\n", title)
	if len(items) == 0 {
		fmt.Println("_None._\n")
		return
	}
	for _, item := range items {
		fmt.Printf("- %s\n", item)
	}
	fmt.Println()
}

func setKeys(set exportSet) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
