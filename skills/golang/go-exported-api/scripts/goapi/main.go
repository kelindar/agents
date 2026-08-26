// Command goapi validates types.md (eval) or diffs exports between git refs (diff).
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var code int
	switch os.Args[1] {
	case "eval":
		code = runEval(os.Args[2:])
	case "diff":
		code = runDiff(os.Args[2:])
	case "write":
		code = runWrite(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
	os.Exit(code)
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage:
  go run ./.agents/skills/go-exported-api/scripts/goapi write [-o types.md] [-internal] [-legacy]
  go run ./.agents/skills/go-exported-api/scripts/goapi eval [types.md] [-internal]
  go run ./.agents/skills/go-exported-api/scripts/goapi diff [-base main] [-internal] [-full] [-markdown]

Flags (eval/write/diff): -config, -packages, -internal
Diff only: -head, -extra, -skip-cmd-structs, -full, -markdown
`)
}

func exitErr(err error) {
	fmt.Fprintf(os.Stderr, "goapi: %v\n", err)
	os.Exit(1)
}
