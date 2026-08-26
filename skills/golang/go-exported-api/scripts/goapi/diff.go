package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"sort"
	"strings"
)

type symbol struct {
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Signature  string `json:"signature"`
	File       string `json:"file"`
	Package    string `json:"package"`
	Deprecated bool   `json:"deprecated,omitempty"`
}

type bridgeEntry struct {
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	File    string `json:"file"`
	Change  string `json:"change"`
	OldLine string `json:"old_line,omitempty"`
	NewLine string `json:"new_line,omitempty"`
}

type report struct {
	Base      string        `json:"base"`
	Head      string        `json:"head"`
	Added     []symbol      `json:"added"`
	Removed   []symbol      `json:"removed"`
	Changed   []changeEntry `json:"changed"`
	WASM      []bridgeEntry `json:"wasm"`
	HTTP      []bridgeEntry `json:"http"`
	Unchanged int           `json:"unchanged_count"`
	Scanned   []string      `json:"scanned_packages"`
}

type changeEntry struct {
	symbol
	OldSignature string `json:"old_signature"`
}

type diffMethodInfo struct {
	Name       string
	Signature  string
	Deprecated bool
}

func runDiff(args []string) int {
	fs := flag.NewFlagSet("diff", flag.ExitOnError)
	registerConfigFlags(fs)
	base := fs.String("base", "main", "git base ref")
	head := fs.String("head", "HEAD", "git head ref")
	full := fs.Bool("full", false, "compare full exported API inventory (not just changed files)")
	skipCmdStructs := fs.Bool("skip-cmd-structs", true, "omit cmd/ struct field changes")
	markdown := fs.Bool("markdown", false, "markdown output")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	rc := parseRuntimeConfig(fs)
	if flagPassed("skip-cmd-structs") {
		rc.skipCmdStructs = *skipCmdStructs
	}

	repoRoot, err := gitRoot()
	if err != nil {
		exitErr(err)
	}

	if *full {
		return runFullDiff(repoRoot, *base, *head, rc, *markdown)
	}

	files, err := changedGoFiles(repoRoot, *base, *head)
	if err != nil {
		exitErr(err)
	}
	if len(files) == 0 {
		fmt.Println("No changed Go files between refs.")
		return 0
	}

	oldSyms := map[string]symbol{}
	newSyms := map[string]symbol{}
	scanned := map[string]struct{}{}

	for _, rel := range files {
		if strings.HasSuffix(rel, "_test.go") || strings.HasSuffix(rel, "_gen.go") {
			continue
		}
		if !shouldScan(rel, rc.packagePrefixes, rc.includeInternal) {
			continue
		}
		pkg := packageDir(rel)
		scanned[pkg] = struct{}{}

		if oldSrc, ok := gitShow(repoRoot, *base, rel); ok {
			for k, v := range extractDetailedSymbols(oldSrc, rel, pkg) {
				oldSyms[k] = v
			}
		}
		if newSrc, ok := readWorktree(repoRoot, rel); ok {
			for k, v := range extractDetailedSymbols(newSrc, rel, pkg) {
				newSyms[k] = v
			}
		}
	}

	rep := report{
		Base:    *base,
		Head:    *head,
		Added:   diffAdded(oldSyms, newSyms),
		Removed: diffRemoved(oldSyms, newSyms),
		Changed: diffChanged(oldSyms, newSyms),
	}
	if rc.skipCmdStructs {
		rep.Changed = filterCmdStructChanges(rep.Changed)
	}
	for k := range oldSyms {
		if nv, ok := newSyms[k]; ok && oldSyms[k].Signature == nv.Signature {
			rep.Unchanged++
		}
	}
	for k := range scanned {
		rep.Scanned = append(rep.Scanned, k)
	}
	sort.Strings(rep.Scanned)

	for _, rel := range rc.extra {
		rel = strings.TrimSpace(rel)
		if rel == "" {
			continue
		}
		oldSrc, oldOK := gitShow(repoRoot, *base, rel)
		newSrc, newOK := readWorktree(repoRoot, rel)
		if oldOK || newOK {
			rep.WASM = append(rep.WASM, diffWASM(oldSrc, newSrc, rel)...)
			rep.HTTP = append(rep.HTTP, diffHTTP(oldSrc, newSrc, rel)...)
		}
	}

	sort.Slice(rep.Added, func(i, j int) bool { return rep.Added[i].Name < rep.Added[j].Name })
	sort.Slice(rep.Removed, func(i, j int) bool { return rep.Removed[i].Name < rep.Removed[j].Name })
	sort.Slice(rep.Changed, func(i, j int) bool { return rep.Changed[i].Name < rep.Changed[j].Name })

	if *markdown {
		printDiffMarkdown(rep)
		return 0
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rep); err != nil {
		exitErr(err)
	}
	return 0
}

func extractDetailedSymbols(src, file, pkg string) map[string]symbol {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, src, parser.ParseComments)
	if err != nil {
		return nil
	}
	out := make(map[string]symbol)
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name == nil || !ast.IsExported(d.Name.Name) {
				continue
			}
			if d.Recv != nil && len(d.Recv.List) > 0 {
				recv := normalizeReceiver(exprString(d.Recv.List[0].Type))
				if !ast.IsExported(recv) {
					continue
				}
			}
			key := pkg + "." + diffFuncKey(d)
			out[key] = symbol{
				Kind:       "func",
				Name:       key,
				Signature:  formatNode(fset, d.Type),
				File:       file,
				Package:    pkg,
				Deprecated: hasDeprecatedComment(d.Doc),
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || ts.Name == nil || !ast.IsExported(ts.Name.Name) {
					continue
				}
				key := pkg + "." + ts.Name.Name
				out[key] = symbol{
					Kind:       diffTypeKind(ts),
					Name:       key,
					Signature:  formatTypeSpec(fset, ts),
					File:       file,
					Package:    pkg,
					Deprecated: hasDeprecatedComment(d.Doc) || hasDeprecatedComment(ts.Doc),
				}
				if it, ok := ts.Type.(*ast.InterfaceType); ok {
					for _, method := range diffInterfaceMethods(it) {
						mkey := key + "." + method.Name
						out[mkey] = symbol{
							Kind:       "method",
							Name:       mkey,
							Signature:  method.Signature,
							File:       file,
							Package:    pkg,
							Deprecated: method.Deprecated,
						}
					}
				}
			}
			if d.Tok == token.CONST || d.Tok == token.VAR {
				names, sig := constVarSignature(d)
				for _, name := range names {
					if !ast.IsExported(name) {
						continue
					}
					key := pkg + "." + name
					out[key] = symbol{
						Kind:       strings.ToLower(d.Tok.String()),
						Name:       key,
						Signature:  sig,
						File:       file,
						Package:    pkg,
						Deprecated: hasDeprecatedComment(d.Doc),
					}
				}
			}
		}
	}
	return out
}

func diffInterfaceMethods(it *ast.InterfaceType) []diffMethodInfo {
	if it.Methods == nil {
		return nil
	}
	var out []diffMethodInfo
	for _, field := range it.Methods.List {
		for _, n := range field.Names {
			if !ast.IsExported(n.Name) {
				continue
			}
			out = append(out, diffMethodInfo{
				Name:       n.Name,
				Signature:  exprString(field.Type),
				Deprecated: hasDeprecatedComment(field.Doc),
			})
		}
	}
	return out
}

func diffFuncKey(d *ast.FuncDecl) string {
	if d.Recv != nil && len(d.Recv.List) > 0 {
		return exprString(d.Recv.List[0].Type) + "." + d.Name.Name
	}
	return d.Name.Name
}

func diffTypeKind(ts *ast.TypeSpec) string {
	switch ts.Type.(type) {
	case *ast.InterfaceType:
		return "interface"
	case *ast.StructType:
		return "struct"
	default:
		return "type"
	}
}

func formatTypeSpec(fset *token.FileSet, ts *ast.TypeSpec) string {
	if ts.TypeParams != nil {
		return formatNode(fset, ts)
	}
	switch t := ts.Type.(type) {
	case *ast.StructType:
		return formatStruct(t)
	case *ast.InterfaceType:
		return formatInterface(t)
	default:
		return exprString(ts.Type)
	}
}

func formatStruct(st *ast.StructType) string {
	if st.Fields == nil {
		return "struct{}"
	}
	var fields []string
	for _, f := range st.Fields.List {
		names := fieldNames(f.Names)
		if len(names) == 0 {
			names = []string{"_"}
		}
		for _, name := range names {
			fields = append(fields, fmt.Sprintf("%s %s", name, exprString(f.Type)))
		}
	}
	sort.Strings(fields)
	return "struct { " + strings.Join(fields, "; ") + " }"
}

func formatInterface(it *ast.InterfaceType) string {
	if it.Methods == nil || len(it.Methods.List) == 0 {
		return "interface{}"
	}
	var methods []string
	for _, f := range it.Methods.List {
		for _, n := range f.Names {
			methods = append(methods, n.Name+exprString(f.Type))
		}
	}
	sort.Strings(methods)
	return "interface { " + strings.Join(methods, "; ") + " }"
}

func constVarSignature(d *ast.GenDecl) (names []string, sig string) {
	for _, spec := range d.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, n := range vs.Names {
			names = append(names, n.Name)
		}
		if vs.Type != nil {
			sig = exprString(vs.Type)
		}
	}
	return names, sig
}

func fieldNames(names []*ast.Ident) []string {
	out := make([]string, 0, len(names))
	for _, n := range names {
		out = append(out, n.Name)
	}
	return out
}

func formatNode(fset *token.FileSet, node ast.Node) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return "<unformatted>"
	}
	return strings.TrimSpace(buf.String())
}

func hasDeprecatedComment(cg *ast.CommentGroup) bool {
	if cg == nil {
		return false
	}
	for _, c := range cg.List {
		if strings.Contains(c.Text, "Deprecated:") {
			return true
		}
	}
	return false
}

func diffAdded(old, new map[string]symbol) []symbol {
	var out []symbol
	for k, v := range new {
		if _, ok := old[k]; !ok {
			out = append(out, v)
		}
	}
	return out
}

func diffRemoved(old, new map[string]symbol) []symbol {
	var out []symbol
	for k, v := range old {
		if _, ok := new[k]; !ok {
			out = append(out, v)
		}
	}
	return out
}

func diffChanged(old, new map[string]symbol) []changeEntry {
	var out []changeEntry
	for k, nv := range new {
		ov, ok := old[k]
		if !ok || ov.Signature == nv.Signature {
			continue
		}
		out = append(out, changeEntry{symbol: nv, OldSignature: ov.Signature})
	}
	return out
}

func filterCmdStructChanges(in []changeEntry) []changeEntry {
	var out []changeEntry
	for _, c := range in {
		if strings.HasPrefix(c.Package, "cmd/") && c.Kind == "struct" {
			continue
		}
		out = append(out, c)
	}
	return out
}

var (
	wasmRe = regexp.MustCompile(`js\.Global\(\)\.Set\("([^"]+)"`)
	httpRe = regexp.MustCompile(`http\.HandleFunc\("([^"]+)"`)
)

func diffWASM(oldSrc, newSrc, file string) []bridgeEntry {
	return diffBridge(wasmRe, "wasm", oldSrc, newSrc, file)
}

func diffHTTP(oldSrc, newSrc, file string) []bridgeEntry {
	return diffBridge(httpRe, "http", oldSrc, newSrc, file)
}

func diffBridge(re *regexp.Regexp, kind, oldSrc, newSrc, file string) []bridgeEntry {
	old := extractBridge(re, oldSrc)
	newB := extractBridge(re, newSrc)
	var out []bridgeEntry
	for name, line := range newB {
		if oldLine, ok := old[name]; !ok {
			out = append(out, bridgeEntry{Kind: kind, Name: name, File: file, Change: "added", NewLine: line})
		} else if oldLine != line {
			out = append(out, bridgeEntry{Kind: kind, Name: name, File: file, Change: "changed", OldLine: oldLine, NewLine: line})
		}
	}
	for name, line := range old {
		if _, ok := newB[name]; !ok {
			out = append(out, bridgeEntry{Kind: kind, Name: name, File: file, Change: "removed", OldLine: line})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func extractBridge(re *regexp.Regexp, src string) map[string]string {
	out := make(map[string]string)
	for _, line := range strings.Split(src, "\n") {
		if m := re.FindStringSubmatch(line); len(m) == 2 {
			out[m[1]] = strings.TrimSpace(line)
		}
	}
	return out
}

func printDiffMarkdown(rep report) {
	fmt.Printf("# Exported API diff: %s...%s\n\n", rep.Base, rep.Head)
	if len(rep.Scanned) > 0 {
		fmt.Println("## Scanned packages")
		for _, p := range rep.Scanned {
			fmt.Printf("- `%s`\n", p)
		}
		fmt.Println()
	}
	printDiffSymbolSection("Breaking / changed signatures", rep.Changed)
	printDiffSymbolSection("Removed", rep.Removed)
	printDiffSymbolSection("Added", rep.Added)
	printDiffBridgeSection("WASM JS bridge", rep.WASM)
	printDiffBridgeSection("HTTP routes", rep.HTTP)
	fmt.Printf("\n_%d unchanged exported symbols in changed files._\n", rep.Unchanged)
}

func printDiffSymbolSection(title string, items any) {
	fmt.Printf("## %s\n\n", title)
	switch v := items.(type) {
	case []changeEntry:
		if len(v) == 0 {
			fmt.Println("_None._\n")
			return
		}
		fmt.Println("| Symbol | Change |")
		fmt.Println("|--------|--------|")
		for _, s := range v {
			fmt.Printf("| `%s` | `%s` → `%s` |\n", s.Name, compact(s.OldSignature), compact(s.Signature))
		}
		fmt.Println()
	case []symbol:
		if len(v) == 0 {
			fmt.Println("_None._\n")
			return
		}
		fmt.Println("| Symbol | Kind | Signature |")
		fmt.Println("|--------|------|-----------|")
		for _, s := range v {
			fmt.Printf("| `%s` | %s | `%s` |\n", s.Name, s.Kind, compact(s.Signature))
		}
		fmt.Println()
	}
}

func printDiffBridgeSection(title string, items []bridgeEntry) {
	fmt.Printf("## %s\n\n", title)
	if len(items) == 0 {
		fmt.Println("_None._\n")
		return
	}
	fmt.Println("| Name | Change |")
	fmt.Println("|------|--------|")
	for _, b := range items {
		fmt.Printf("| `%s` | %s |\n", b.Name, b.Change)
	}
	fmt.Println()
}

func compact(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	if len(s) > 120 {
		return s[:117] + "..."
	}
	return s
}

type flatSym struct {
	Pkg  string
	Name string
	Kind string
}

func runFullDiff(root, base, head string, rc runtimeConfig, markdown bool) int {
	oldPkgs, err := scanRepoAtRef(root, base, rc.includeInternal, rc.packagePrefixes)
	if err != nil {
		exitErr(err)
	}
	newPkgs, err := scanRepoAtRef(root, head, rc.includeInternal, rc.packagePrefixes)
	if err != nil {
		exitErr(err)
	}

	oldFlat := map[string]flatSym{}
	newFlat := map[string]flatSym{}

	flatten := func(pkgs map[string]*packageExports, out map[string]flatSym) {
		for _, pe := range sortedPackages(pkgs) {
			for sym := range pe.Symbols {
				kind := "symbol"
				if strings.Contains(sym, ".") {
					kind = "method"
				}
				key := pe.ImportPath + "\x00" + sym
				out[key] = flatSym{Pkg: pe.ShortName, Name: sym, Kind: kind}
			}
		}
	}
	flatten(oldPkgs, oldFlat)
	flatten(newPkgs, newFlat)

	var added, removed []flatSym
	for k, v := range newFlat {
		if _, ok := oldFlat[k]; !ok {
			added = append(added, v)
		}
	}
	for k, v := range oldFlat {
		if _, ok := newFlat[k]; !ok {
			removed = append(removed, v)
		}
	}
	sort.Slice(added, func(i, j int) bool {
		if added[i].Pkg != added[j].Pkg {
			return added[i].Pkg < added[j].Pkg
		}
		return added[i].Name < added[j].Name
	})
	sort.Slice(removed, func(i, j int) bool {
		if removed[i].Pkg != removed[j].Pkg {
			return removed[i].Pkg < removed[j].Pkg
		}
		return removed[i].Name < removed[j].Name
	})

	if markdown {
		printFullDiffMarkdown(base, head, added, removed)
		return 0
	}

	fmt.Printf("full inventory diff: %s...%s\n", base, head)
	fmt.Printf("added=%d removed=%d\n", len(added), len(removed))
	for _, s := range added {
		fmt.Printf("+\t%s\t%s\t%s\n", s.Pkg, s.Name, s.Kind)
	}
	for _, s := range removed {
		fmt.Printf("-\t%s\t%s\t%s\n", s.Pkg, s.Name, s.Kind)
	}
	return 0
}

func printFullDiffMarkdown(base, head string, added, removed []flatSym) {
	fmt.Printf("## Branch delta (`%s` vs `%s`)\n\n", head, base)
	fmt.Printf("Compared with exported API inventory on `%s` (%d new symbols, %d removed).\n\n", base, len(added), len(removed))
	if len(added) > 0 {
		fmt.Println("### New exported symbols")
		fmt.Println()
		fmt.Println("| Symbol | Kind | Package |")
		fmt.Println("|--------|------|---------|")
		for _, s := range added {
			fmt.Printf("| `%s` | %s | `%s` |\n", s.Name, s.Kind, s.Pkg)
		}
		fmt.Println()
	}
	if len(removed) > 0 {
		fmt.Println("### Removed exported symbols")
		fmt.Println()
		fmt.Println("| Symbol | Kind | Package |")
		fmt.Println("|--------|------|---------|")
		for _, s := range removed {
			fmt.Printf("| `%s` | %s | `%s` |\n", s.Name, s.Kind, s.Pkg)
		}
		fmt.Println()
	}
}
