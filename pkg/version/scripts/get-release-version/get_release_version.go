package main

import (
	"fmt"

	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/version"
)

func main() {
	v, err := version.NewInfo()
	if err != nil {
		log.Warn(err)
	}
	fmt.Println(v.ReleaseVersion)
}
