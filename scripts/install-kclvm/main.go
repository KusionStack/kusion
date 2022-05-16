package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"kusionstack.io/kclvm-go/scripts"
)

var (
	flagTriple = flag.String("triple", "", "set kclvm triple")
	flagOutdir = flag.String("outdir", "", "set output dir")

	flagMirrors = flag.String("mirrors", "", "set mirror url")
)

func main() {
	flag.Parse()
	if *flagTriple == "" || *flagOutdir == "" {
		flag.Usage()
		os.Exit(1)
	}

	scripts.DefaultKclvmTriple = *flagTriple
	if s := *flagMirrors; s != "" {
		for _, s := range strings.Split(s, ",") {
			s := strings.TrimSpace(s)
			if s != "" {
				scripts.KclvmDownloadUrlBase_mirrors = append(scripts.KclvmDownloadUrlBase_mirrors, s)
			}
		}
	}

	err := scripts.SetupKclvm(*flagOutdir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
