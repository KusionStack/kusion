package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"kusionstack.io/kusion/pkg/kusionctl/cmd"
	"kusionstack.io/kusion/pkg/util/io"

	"github.com/spf13/cobra/doc"
)

func main() {
	// use os.Args instead of "flags" because "flags" will mess up the man pages!
	path := "docs/"
	if len(os.Args) == 2 {
		path = os.Args[1]
	} else if len(os.Args) > 2 {
		fmt.Fprintf(os.Stderr, "usage: %s [output directory]\n", os.Args[0])
		os.Exit(1)
	}

	outDir, err := io.OutDir(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get output directory: %v\n", err)
		os.Exit(1)
	}

	// Set environment variables used by kusionctl so the output is consistent,
	// regardless of where we run.
	os.Setenv("HOME", "/home/username")
	kusionctl := cmd.NewKusionctlCmd(bytes.NewReader(nil), ioutil.Discard, ioutil.Discard)
	doc.GenMarkdownTree(kusionctl, outDir)
}
