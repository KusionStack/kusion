package main

import (
	"math/rand"
	"os"
	"time"

	_ "kcl-lang.io/kcl-plugin"

	"kusionstack.io/kusion/pkg/cmd"
	"kusionstack.io/kusion/pkg/util/pretty"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	command := cmd.NewDefaultKusionctlCommand()

	if err := command.Execute(); err != nil {
		pretty.Error.Println(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
