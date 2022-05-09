package io

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/build"
)

// Generate yaml with kustomize build
func ReadKustomizeInput(kustomizeDir string) (string, error) {
	fSys := filesys.MakeFsOnDisk()
	k := krusty.MakeKustomizer(
		build.HonorKustomizeFlags(krusty.MakeDefaultOptions()),
	)
	m, err := k.Run(fSys, kustomizeDir)
	if err != nil {
		return "", err
	}
	yml, err := m.AsYaml()
	if err != nil {
		return "", err
	}
	return string(yml), nil
}

// Read stdin content as string
func ReadStdinInput() (string, error) {
	// validate
	info, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeCharDevice != 0 {
		return "", fmt.Errorf("no data read from stdin")
	}

	// read content from stdin until EOF is encountered
	input := bufio.NewReader(os.Stdin)
	output := bytes.NewBuffer([]byte{})
	for {
		b, err := input.ReadByte()
		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}
		output.WriteByte(b)
	}
	return output.String(), nil
}
