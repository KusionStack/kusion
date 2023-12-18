package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra/doc"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"kusionstack.io/kusion/pkg/cmd"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util/i18n"
	uio "kusionstack.io/kusion/pkg/util/io"
)

const defaultCmdDocsDir = "docs/cmd/"

var docsMatrix = map[string]string{
	i18n.LangEnUS: i18n.LangValueEnUS,
	i18n.LangZhCN: i18n.LangValueZhCN,
}

func main() {
	// use os.Args instead of "flags" because "flags" will mess up the man pages!
	path := defaultCmdDocsDir
	langEnvKey := i18n.EnvKeyLang
	if len(os.Args) >= 2 {
		path = os.Args[1]
	}
	if len(os.Args) == 3 {
		langEnvKey = os.Args[2]
		if langEnvKey != i18n.EnvKeyLanguage && langEnvKey != i18n.EnvKeyLcAll && langEnvKey != i18n.EnvKeyLcMessages && langEnvKey != i18n.EnvKeyLang {
			log.Fatal("lang env key must be %s, %s, %s or %s", i18n.EnvKeyLanguage, i18n.EnvKeyLcAll, i18n.EnvKeyLcMessages, i18n.EnvKeyLang)
		}
	} else if len(os.Args) > 3 {
		log.Fatal("usage: %s [output directory] [lang env key]\n", os.Args[0])
	}

	outDir, err := uio.OutDir(path)
	if err != nil {
		log.Fatal(os.Stderr, "failed to get output directory: %v\n", err)
	}

	genDocs(outDir, langEnvKey)
}

func genDocs(rootDir, langEnvKey string) {
	for langDir, lang := range docsMatrix {
		targetDir := filepath.Join(rootDir, langDir)
		if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
			log.Fatal("make target dir '%s' failed: %v", targetDir, err)
		}
		if err := os.Setenv(langEnvKey, lang); err != nil {
			log.Fatal("set env 'LANG' to '%s' failed: %v", lang, err)
		}
		command := cmd.NewKusionctlCmd(cmd.KusionctlOptions{
			IOStreams: genericclioptions.IOStreams{},
		})
		if err := doc.GenMarkdownTree(command, targetDir); err != nil {
			log.Fatal("generate markdown tree failed: %v", err)
		}
	}
}
