package main

import (
	"math/rand"
	"os"
	"time"

	_ "kusionstack.io/kcl_plugins"
	"kusionstack.io/kusion/pkg/kusionctl/cmd"
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
