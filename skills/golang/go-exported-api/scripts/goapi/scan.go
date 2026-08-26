package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type exportSet map[string]struct{}

type packageExports struct {
	ImportPath string
	ShortName  string
	Symbols    exportSet
}

func scanRepo(root string, includeInternal bool, packagePrefixes []string) (map[string]*packageExports, error) {
	module, err := readModule(root)
	if err != nil {
		return nil, err
	}

	out := map[string]*packageExports{}
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := d.Name()
			switch base {
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

		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, rel, src, parser.ParseComments)
		if err != nil {
			return nil
		}
		if f.Name.Name == "main" {
			return nil
		}

		dir := filepath.ToSlash(filepath.Dir(rel))
		if dir == "." {
			dir = module
		} else {
			dir = module + "/" + dir
		}

		pe := out[dir]
		if pe == nil {
			pe = &packageExports{
				ImportPath: dir,
				ShortName:  filepath.Base(dir),
				Symbols:    exportSet{},
			}
			out[dir] = pe
		}
		for sym := range extractSymbolNames(f) {
			pe.Symbols[sym] = struct{}{}
		}
		return nil
	})
	return out, err
}

func readModule(root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("module directive not found in go.mod")
}

func extractSymbolNames(f *ast.File) exportSet {
	out := exportSet{}
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
				out[recv+"."+d.Name.Name] = struct{}{}
				continue
			}
			out[d.Name.Name] = struct{}{}
		case *ast.GenDecl:
			switch d.Tok {
			case token.TYPE:
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok || ts.Name == nil || !ast.IsExported(ts.Name.Name) {
						continue
					}
					out[ts.Name.Name] = struct{}{}
					if it, ok := ts.Type.(*ast.InterfaceType); ok {
						for _, m := range interfaceMethodNames(it) {
							out[ts.Name.Name+"."+m] = struct{}{}
						}
					}
				}
			case token.CONST, token.VAR:
				for _, spec := range d.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for _, n := range vs.Names {
						if ast.IsExported(n.Name) {
							out[n.Name] = struct{}{}
						}
					}
				}
			}
		}
	}
	return out
}

func interfaceMethodNames(it *ast.InterfaceType) []string {
	if it.Methods == nil {
		return nil
	}
	var out []string
	for _, field := range it.Methods.List {
		for _, n := range field.Names {
			if ast.IsExported(n.Name) {
				out = append(out, n.Name)
			}
		}
	}
	sort.Strings(out)
	return out
}

func normalizeReceiver(recv string) string {
	return strings.TrimPrefix(recv, "*")
}

func exprString(expr ast.Expr) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, token.NewFileSet(), expr); err != nil {
		return "<unformatted>"
	}
	return strings.TrimSpace(buf.String())
}

func shouldScan(rel string, prefixes []string, includeInternal bool) bool {
	if len(prefixes) > 0 {
		for _, p := range prefixes {
			p = strings.TrimSuffix(p, "/")
			if rel == p || strings.HasPrefix(rel, p+"/") {
				return true
			}
		}
		return false
	}
	if !includeInternal && strings.HasPrefix(rel, "internal/") {
		return false
	}
	return true
}

// --- types.md parsing ---

type docPackage struct {
	ImportPath string
	ShortName  string
	Symbols    exportSet
}

var (
	pkgHeadingRe  = regexp.MustCompile(`^##\s+` + "\x60" + `([^\x60]+)\x60(?:\s+\(\x60([^\x60]+)\x60\))?`)
	legacyLineRe  = regexp.MustCompile(`^-\s+` + "\x60" + `([^\x60]+)\x60\s+-`)
	modernLineRe  = regexp.MustCompile(`^-\s+` + "\x60" + `([^\x60]+)\x60\s+—`)
	typeHeadingRe = regexp.MustCompile(`^###\s+` + "\x60" + `([^\x60]+)\x60\s+—`)
	sectionHeading = map[string]struct{}{
		"constants": {}, "variables": {}, "functions": {},
	}
)

func parseTypesMD(path string) (map[string]*docPackage, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := map[string]*docPackage{}
	var current *docPackage
	var currentType string

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "## Branch delta") {
			break
		}

		if m := pkgHeadingRe.FindStringSubmatch(line); len(m) >= 2 {
			short := m[1]
			importPath := short
			if len(m) >= 3 && m[2] != "" {
				importPath = m[2]
			}
			key := importPath
			current = out[key]
			if current == nil {
				current = &docPackage{
					ImportPath: importPath,
					ShortName:  short,
					Symbols:    exportSet{},
				}
				out[key] = current
			}
			currentType = ""
			continue
		}

		if current == nil {
			continue
		}

		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "### ") {
			if _, ok := sectionHeading[strings.TrimPrefix(lower, "### ")]; ok {
				currentType = ""
				continue
			}
			if m := typeHeadingRe.FindStringSubmatch(line); len(m) == 2 {
				currentType = m[1]
				current.Symbols[currentType] = struct{}{}
				continue
			}
		}

		var sym string
		switch {
		case modernLineRe.MatchString(line):
			sym = modernLineRe.FindStringSubmatch(line)[1]
		case legacyLineRe.MatchString(line):
			sym = legacyLineRe.FindStringSubmatch(line)[1]
		default:
			continue
		}

		sym = strings.TrimPrefix(sym, "*")
		if currentType != "" && !strings.Contains(sym, ".") {
			sym = currentType + "." + sym
		}
		current.Symbols[sym] = struct{}{}
	}
	return out, sc.Err()
}

func matchPackages(code map[string]*packageExports, doc map[string]*docPackage) [][2]*packageExports {
	type pair struct {
		code *packageExports
		doc  *docPackage
	}
	byImport := map[string]*docPackage{}
	byShort := map[string]*docPackage{}
	for _, d := range doc {
		byImport[d.ImportPath] = d
		byShort[d.ShortName] = d
	}

	var pairs []pair
	seenDoc := map[*docPackage]bool{}
	for _, c := range sortedPackages(code) {
		d := byImport[c.ImportPath]
		if d == nil {
			d = byShort[c.ShortName]
		}
		if d != nil {
			seenDoc[d] = true
		}
		pairs = append(pairs, pair{code: c, doc: d})
	}
	for _, d := range doc {
		if !seenDoc[d] {
			pairs = append(pairs, pair{doc: d})
		}
	}
	out := make([][2]*packageExports, 0, len(pairs))
	for _, p := range pairs {
		var docPE *packageExports
		if p.doc != nil {
			docPE = &packageExports{
				ImportPath: p.doc.ImportPath,
				ShortName:  p.doc.ShortName,
				Symbols:    p.doc.Symbols,
			}
		}
		out = append(out, [2]*packageExports{p.code, docPE})
	}
	return out
}

func sortedPackages(m map[string]*packageExports) []*packageExports {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]*packageExports, 0, len(keys))
	for _, k := range keys {
		out = append(out, m[k])
	}
	return out
}

func diffSets(code, doc exportSet) (missing, stale []string) {
	for s := range code {
		if _, ok := doc[s]; !ok {
			missing = append(missing, s)
		}
	}
	for s := range doc {
		if _, ok := code[s]; !ok {
			stale = append(stale, s)
		}
	}
	sort.Strings(missing)
	sort.Strings(stale)
	return missing, stale
}

func formatSymbolList(pkg *packageExports, symbols []string) []string {
	out := make([]string, 0, len(symbols))
	for _, s := range symbols {
		out = append(out, fmt.Sprintf("`%s.%s`", pkg.ShortName, s))
	}
	return out
}
