package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func gitRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("git root: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func changedGoFiles(root, base, head string) ([]string, error) {
	rangeRef := base + "..." + head
	cmd := exec.Command("git", "diff", "--name-only", rangeRef, "--", "*.go")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff %s: %w", rangeRef, err)
	}
	var files []string
	for line := range strings.SplitSeq(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, filepath.ToSlash(line))
		}
	}
	sort.Strings(files)
	return files, nil
}

func gitShow(root, ref, rel string) (string, bool) {
	spec := ref + ":" + rel
	cmd := exec.Command("git", "show", spec)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}
	return string(out), true
}

func readWorktree(root, rel string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		return "", false
	}
	return string(data), true
}

func packageDir(rel string) string {
	dir := filepath.ToSlash(filepath.Dir(rel))
	if dir == "." {
		return "(root)"
	}
	return dir
}

func listTrackedGoFiles(root, ref string, includeInternal bool, packagePrefixes []string) ([]string, error) {
	var cmd *exec.Cmd
	if ref == "" || ref == "HEAD" {
		cmd = exec.Command("git", "ls-files", "--", "*.go")
	} else {
		cmd = exec.Command("git", "ls-tree", "-r", "--name-only", ref)
	}
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list go files at %s: %w", ref, err)
	}
	var files []string
	for line := range strings.SplitSeq(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasSuffix(line, ".go") {
			continue
		}
		line = filepath.ToSlash(line)
		if strings.HasSuffix(line, "_test.go") || strings.HasSuffix(line, "_gen.go") {
			continue
		}
		if !shouldScan(line, packagePrefixes, includeInternal) {
			continue
		}
		files = append(files, line)
	}
	sort.Strings(files)
	return files, nil
}

func scanRepoAtRef(root, ref string, includeInternal bool, packagePrefixes []string) (map[string]*packageExports, error) {
	module, err := readModule(root)
	if err != nil {
		return nil, err
	}

	files, err := listTrackedGoFiles(root, ref, includeInternal, packagePrefixes)
	if err != nil {
		return nil, err
	}

	out := map[string]*packageExports{}
	useWorktree := ref == "" || ref == "HEAD"
	for _, rel := range files {
		var src []byte
		if useWorktree {
			data, err := os.ReadFile(filepath.Join(root, rel))
			if err != nil {
				continue
			}
			src = data
		} else {
			text, ok := gitShow(root, ref, rel)
			if !ok {
				continue
			}
			src = []byte(text)
		}

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, rel, src, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		if f.Name.Name == "main" {
			continue
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
	}
	return out, nil
}
