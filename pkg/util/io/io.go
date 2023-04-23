package io

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
)

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
