package main

import (
	"math/rand"
	"os"
	"time"

	"kusionstack.io/kusion/pkg/cmd"
	"kusionstack.io/kusion/pkg/util/pretty"
)

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	command := cmd.NewDefaultKusionctlCommand()

	if err := command.Execute(); err != nil {
		// Pretty-print the error and exit with an error.
		pretty.ErrorT.Println(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
