// compute the md5 of the new binary
package main

import (
	"crypto/md5"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "-h" {
		fmt.Fprintln(os.Stderr, "Usage: md5file [file]")
		fmt.Fprintln(os.Stderr, "       md5file -h")
		os.Exit(0)
	}

	name := os.Args[1]
	sum, err := MD5File(name)
	if err != nil {
		fmt.Printf("ERR: %s: %v", name, err)
		os.Exit(1)
	}
	fmt.Printf("%x *%s\n", sum, name)
}

func MD5File(filename string) (sum [md5.Size]byte, err error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return
	}
	sum = md5.Sum(data)
	return
}
