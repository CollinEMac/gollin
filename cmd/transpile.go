package cmd

import (
    "log"
    "os"
    "strings"
    "github.com/CollinEMac/gollin/transpiler"
)

func transpileFile(path string) string {
	if !strings.HasSuffix(path, ".gol") {
		path = path + ".gol"
	}

	src, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	goCode := transpiler.Transpile(string(src))

	goPath := strings.TrimSuffix(path, ".gol") + ".go"
	if err := os.WriteFile(goPath, goCode, 0777); err != nil {
		log.Fatal(err)
	}

	return goPath
}
