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
)

type docSymbol struct {
	Name    string
	Key     string // export-set key: Type.Method or bare name
	Kind    string // const, var, func, type, struct, interface, method
	Type    string // parent type for methods / interface methods
	Doc     string
}

type docPackageContent struct {
	ImportPath string
	ShortName  string
	Constants  []docSymbol
	Variables  []docSymbol
	Functions  []docSymbol
	Types      []docTypeContent
}

type docTypeContent struct {
	docSymbol
	Methods []docSymbol
}

func runWrite(args []string) int {
	fs := flag.NewFlagSet("write", flag.ExitOnError)
	registerConfigFlags(fs)
	outPath := fs.String("o", "types.md", "output path")
	legacy := fs.Bool("legacy", false, "legacy flat list (kind suffix) instead of structured sections")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	rc := parseRuntimeConfig(fs)

	repoRoot, err := gitRoot()
	if err != nil {
		exitErr(err)
	}

	pkgs, err := scanDocumented(repoRoot, rc.includeInternal, rc.packagePrefixes)
	if err != nil {
		exitErr(err)
	}

	var b strings.Builder
	b.WriteString("# Project exported API\n\n")
	if *legacy {
		writeLegacyInventory(&b, pkgs)
	} else {
		writeStructuredInventory(&b, pkgs)
	}

	target := *outPath
	if !filepath.IsAbs(target) {
		target = filepath.Join(repoRoot, target)
	}
	if err := os.WriteFile(target, []byte(b.String()), 0o644); err != nil {
		exitErr(err)
	}
	fmt.Printf("Wrote %s (%d packages)\n", *outPath, len(pkgs))
	return 0
}

func scanDocumented(root string, includeInternal bool, packagePrefixes []string) ([]*docPackageContent, error) {
	raw, err := scanRepo(root, includeInternal, packagePrefixes)
	if err != nil {
		return nil, err
	}

	byFile := map[string][]string{}
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".agents", "vendor", "tools", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, "_gen.go") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !shouldScan(rel, packagePrefixes, includeInternal) {
			return nil
		}
		dir := filepath.ToSlash(filepath.Dir(rel))
		byFile[dir] = append(byFile[dir], rel)
		return nil
	})
	if err != nil {
		return nil, err
	}

	out := make([]*docPackageContent, 0, len(raw))
	module, _ := readModule(root)
	for _, pe := range sortedPackages(raw) {
		dir := pe.ImportPath
		if module != "" {
			dir = strings.TrimPrefix(dir, module+"/")
		}
		if dir == pe.ImportPath {
			dir = "."
		}
		files := byFile[dir]
		sort.Strings(files)
		content := buildPackageDoc(root, pe, files)
		out = append(out, content)
	}
	return out, nil
}

func buildPackageDoc(root string, pe *packageExports, files []string) *docPackageContent {
	content := &docPackageContent{
		ImportPath: pe.ImportPath,
		ShortName:  pe.ShortName,
	}
	typeMethods := map[string][]docSymbol{}
	typeDocs := map[string]docSymbol{}
	ifaceMethods := map[string][]docSymbol{}

	for _, rel := range files {
		src, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			continue
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, rel, src, parser.ParseComments)
		if err != nil || f.Name.Name == "main" {
			continue
		}
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Name == nil || !ast.IsExported(d.Name.Name) {
					continue
				}
				doc := firstSentence(docText(d.Doc))
				if d.Recv != nil && len(d.Recv.List) > 0 {
					recv := normalizeReceiver(exprString(d.Recv.List[0].Type))
					if !ast.IsExported(recv) {
						continue
					}
					typeMethods[recv] = append(typeMethods[recv], docSymbol{
						Name: d.Name.Name,
						Key:  recv + "." + d.Name.Name,
						Kind: "method",
						Type: recv,
						Doc:  doc,
					})
					continue
				}
				content.Functions = append(content.Functions, docSymbol{
					Name: d.Name.Name,
					Key:  d.Name.Name,
					Kind: "func",
					Doc:  doc,
				})
			case *ast.GenDecl:
				switch d.Tok {
				case token.CONST:
					for _, spec := range d.Specs {
						vs, ok := spec.(*ast.ValueSpec)
						if !ok {
							continue
						}
						doc := valueSpecDoc(vs, d)
						for _, n := range vs.Names {
							if !ast.IsExported(n.Name) {
								continue
							}
							content.Constants = append(content.Constants, docSymbol{
								Name: n.Name, Key: n.Name, Kind: "const", Doc: doc,
							})
						}
					}
				case token.VAR:
					for _, spec := range d.Specs {
						vs, ok := spec.(*ast.ValueSpec)
						if !ok {
							continue
						}
						doc := valueSpecDoc(vs, d)
						for _, n := range vs.Names {
							if !ast.IsExported(n.Name) {
								continue
							}
							content.Variables = append(content.Variables, docSymbol{
								Name: n.Name, Key: n.Name, Kind: "var", Doc: doc,
							})
						}
					}
				case token.TYPE:
					for _, spec := range d.Specs {
						ts, ok := spec.(*ast.TypeSpec)
						if !ok || ts.Name == nil || !ast.IsExported(ts.Name.Name) {
							continue
						}
						kind := "type"
						switch ts.Type.(type) {
						case *ast.StructType:
							kind = "struct"
						case *ast.InterfaceType:
							kind = "interface"
						}
						doc := firstSentence(docText(ts.Doc))
						if doc == "" {
							doc = firstSentence(docText(d.Doc))
						}
						sym := docSymbol{Name: ts.Name.Name, Key: ts.Name.Name, Kind: kind, Doc: doc}
						typeDocs[ts.Name.Name] = sym
						if it, ok := ts.Type.(*ast.InterfaceType); ok {
							for _, m := range interfaceMethodNames(it) {
								ifaceMethods[ts.Name.Name] = append(ifaceMethods[ts.Name.Name], docSymbol{
									Name: m, Key: ts.Name.Name + "." + m, Kind: "method", Type: ts.Name.Name,
								})
							}
						}
					}
				}
			}
		}
	}

	for name, sym := range typeDocs {
		methods := append(typeMethods[name], ifaceMethods[name]...)
		sort.Slice(methods, func(i, j int) bool { return methods[i].Name < methods[j].Name })
		content.Types = append(content.Types, docTypeContent{docSymbol: sym, Methods: methods})
	}
	sort.Slice(content.Types, func(i, j int) bool { return content.Types[i].Name < content.Types[j].Name })
	sort.Slice(content.Constants, func(i, j int) bool { return content.Constants[i].Name < content.Constants[j].Name })
	sort.Slice(content.Variables, func(i, j int) bool { return content.Variables[i].Name < content.Variables[j].Name })
	sort.Slice(content.Functions, func(i, j int) bool { return content.Functions[i].Name < content.Functions[j].Name })
	return content
}

func writeStructuredInventory(b *strings.Builder, pkgs []*docPackageContent) {
	for _, pkg := range pkgs {
		fmt.Fprintf(b, "## `%s` (`%s`)\n\n", pkg.ShortName, pkg.ImportPath)
		writeSymbolSection(b, "Constants", pkg.Constants)
		writeSymbolSection(b, "Variables", pkg.Variables)
		writeSymbolSection(b, "Functions", pkg.Functions)
		for _, typ := range pkg.Types {
			desc := typ.Doc
			if desc == "" {
				desc = defaultDesc(typ.Kind, typ.Name)
			}
			fmt.Fprintf(b, "### `%s` — %s\n", typ.Name, desc)
			if len(typ.Methods) == 0 {
				b.WriteString("\n")
				continue
			}
			for _, m := range typ.Methods {
				doc := m.Doc
				if doc == "" {
					doc = defaultDesc("method", m.Name)
				}
				fmt.Fprintf(b, "- `%s` — %s\n", m.Name, doc)
			}
			b.WriteString("\n")
		}
	}
}

func writeSymbolSection(b *strings.Builder, title string, syms []docSymbol) {
	if len(syms) == 0 {
		return
	}
	fmt.Fprintf(b, "### %s\n\n", title)
	for _, s := range syms {
		doc := s.Doc
		if doc == "" {
			doc = defaultDesc(s.Kind, s.Name)
		}
		fmt.Fprintf(b, "- `%s` — %s\n", s.Name, doc)
	}
	b.WriteString("\n")
}

func writeLegacyInventory(b *strings.Builder, pkgs []*docPackageContent) {
	b.WriteString("## Exported API\n\n")
	for _, pkg := range pkgs {
		fmt.Fprintf(b, "### `%s` (`%s`)\n\n", pkg.ShortName, pkg.ImportPath)
		var lines []string
		for _, s := range pkg.Constants {
			lines = append(lines, fmt.Sprintf("- `%s` - const", s.Name))
		}
		for _, s := range pkg.Variables {
			lines = append(lines, fmt.Sprintf("- `%s` - var", s.Name))
		}
		for _, s := range pkg.Functions {
			lines = append(lines, fmt.Sprintf("- `%s` - func", s.Name))
		}
		for _, typ := range pkg.Types {
			lines = append(lines, fmt.Sprintf("- `%s` - %s", typ.Name, typ.Kind))
			for _, m := range typ.Methods {
				lines = append(lines, fmt.Sprintf("- `*%s.%s` - method", typ.Name, m.Name))
			}
		}
		sort.Strings(lines)
		for _, line := range lines {
			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
	}
}

func docText(cg *ast.CommentGroup) string {
	if cg == nil {
		return ""
	}
	var lines []string
	for _, c := range cg.List {
		text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
		lines = append(lines, text)
	}
	return strings.Join(lines, " ")
}

func fieldDoc(field *ast.Field, decl *ast.GenDecl) string {
	if doc := firstSentence(docText(field.Comment)); doc != "" {
		return doc
	}
	if doc := firstSentence(docText(field.Doc)); doc != "" {
		return doc
	}
	if decl != nil {
		if doc := firstSentence(docText(decl.Doc)); doc != "" {
			return doc
		}
	}
	return ""
}

func valueSpecDoc(vs *ast.ValueSpec, decl *ast.GenDecl) string {
	if doc := firstSentence(docText(vs.Comment)); doc != "" {
		return doc
	}
	if doc := firstSentence(docText(vs.Doc)); doc != "" {
		return doc
	}
	if decl != nil {
		if doc := firstSentence(docText(decl.Doc)); doc != "" {
			return doc
		}
	}
	return ""
}

func firstSentence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if idx := strings.IndexAny(s, ".!?"); idx >= 0 {
		return strings.TrimSpace(s[:idx+1])
	}
	return s
}

func defaultDesc(kind, name string) string {
	switch kind {
	case "const", "var":
		return fmt.Sprintf("exported %s", kind)
	case "func":
		return "exported function"
	case "method":
		return "exported method"
	case "struct":
		return "exported struct type"
	case "interface":
		return "exported interface type"
	default:
		return fmt.Sprintf("exported %s", kind)
	}
}
