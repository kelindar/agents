package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type fileConfig struct {
	Packages        []string `json:"packages"`
	Extra           []string `json:"extra"`
	IncludeInternal bool     `json:"include_internal"`
	SkipCmdStructs  bool     `json:"skip_cmd_structs"`
}

type runtimeConfig struct {
	packagePrefixes []string
	extra           []string
	includeInternal bool
	skipCmdStructs  bool
}

func registerConfigFlags(fs *flag.FlagSet) {
	fs.String("config", "", "JSON config (default: .goapi.json in repo root)")
	fs.String("packages", "", "comma-separated path prefixes")
	fs.String("extra", "", "comma-separated bridge files (diff only)")
	fs.Bool("internal", false, "include internal/ packages")
}

func parseRuntimeConfig(fs *flag.FlagSet) runtimeConfig {
	root, err := gitRoot()
	if err != nil {
		exitErr(err)
	}

	fileCfg, err := loadFileConfig(resolveConfigPath(root, flagGet(fs, "config")))
	if err != nil {
		exitErr(err)
	}

	rc := runtimeConfig{
		packagePrefixes: splitCSV(flagGet(fs, "packages")),
		includeInternal: flagGetBool(fs, "internal"),
		skipCmdStructs:  fileCfg.SkipCmdStructs,
		extra:           fileCfg.Extra,
	}
	if len(rc.packagePrefixes) == 0 {
		rc.packagePrefixes = fileCfg.Packages
	}
	if !flagPassed("internal") {
		rc.includeInternal = fileCfg.IncludeInternal
	}
	if extra := splitCSV(flagGet(fs, "extra")); len(extra) > 0 {
		rc.extra = extra
	}
	return rc
}

func loadFileConfig(path string) (fileConfig, error) {
	cfg := fileConfig{SkipCmdStructs: true}
	if path == "" {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}

func resolveConfigPath(root, explicit string) string {
	if explicit != "" {
		return explicit
	}
	p := filepath.Join(root, ".goapi.json")
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return ""
}

func flagGet(fs *flag.FlagSet, name string) string {
	f := fs.Lookup(name)
	if f == nil {
		return ""
	}
	return f.Value.String()
}

func flagGetBool(fs *flag.FlagSet, name string) bool {
	return flagGet(fs, name) == "true"
}

func flagPassed(name string) bool {
	for _, arg := range os.Args[1:] {
		if arg == "-"+name || arg == "--"+name || strings.HasPrefix(arg, "-"+name+"=") || strings.HasPrefix(arg, "--"+name+"=") {
			return true
		}
	}
	return false
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var out []string
	for part := range strings.SplitSeq(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
