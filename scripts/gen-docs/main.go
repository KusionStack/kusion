package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra/doc"

	"kusionstack.io/kusion/pkg/cmd"
	"kusionstack.io/kusion/pkg/log"
	uio "kusionstack.io/kusion/pkg/util/io"
)

var docsMatrix = map[string]string{
	"en": "en_US.UTF-8",
	"zh": "zh_CN.UTF-8",
}

func main() {
	// use os.Args instead of "flags" because "flags" will mess up the man pages!
	path := "docs/cmd/"
	if len(os.Args) == 2 {
		path = os.Args[1]
	} else if len(os.Args) > 2 {
		log.Fatal("usage: %s [output directory]\n", os.Args[0])
	}

	outDir, err := uio.OutDir(path)
	if err != nil {
		log.Fatal(os.Stderr, "failed to get output directory: %v\n", err)
	}

	genDocs(outDir)
}

func genDocs(rootDir string) {
	for langDir, lang := range docsMatrix {
		targetDir := filepath.Join(rootDir, langDir)
		if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
			log.Fatal("make target dir '%s' failed: %v", targetDir, err)
		}
		if err := os.Setenv("LANG", lang); err != nil {
			log.Fatal("set env 'LANG' to '%s' failed: %v", lang, err)
		}
		command := cmd.NewKusionctlCmd(bytes.NewReader(nil), io.Discard, io.Discard)
		if err := doc.GenMarkdownTree(command, targetDir); err != nil {
			log.Fatal("generate markdown tree failed: %v", err)
		}
	}
}
