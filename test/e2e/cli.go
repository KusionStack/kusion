package e2e

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega/gexec"
)

// GetWrokDir return work directory
func GetWorkDir() string {
	dir, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}
	return dir
}

// GetKusionCLIBin return kusion binary path in e2e test
func GetKusionCLIBin() string {
	dir, _ := os.Getwd()
	binPath := filepath.Join(dir, "../..", "bin")
	return binPath
}

// Exec execute common command
func Exec(cli string) (string, error) {
	var output []byte
	c := strings.Fields(cli)
	command := exec.Command(c[0], c[1:]...)
	session, err := gexec.Start(command, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	if err != nil {
		return string(output), err
	}
	s := session.Wait(300 * time.Second)
	return string(s.Out.Contents()) + string(s.Err.Contents()), nil
}

// Exec execute common command
func ExecWithWorkDir(cli, dir string) (string, error) {
	var output []byte
	c := strings.Fields(cli)
	command := exec.Command(c[0], c[1:]...)
	command.Dir = dir
	session, err := gexec.Start(command, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	if err != nil {
		return string(output), err
	}
	s := session.Wait(300 * time.Second)
	return string(s.Out.Contents()) + string(s.Err.Contents()), nil
}

// ExecKusion executes kusion command
func ExecKusion(cli string) (string, error) {
	var output []byte
	c := strings.Fields(cli)
	commandName := filepath.Join(GetKusionCLIBin(), c[0])
	command := exec.Command(commandName, c[1:]...)
	session, err := gexec.Start(command, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	if err != nil {
		return string(output), err
	}
	s := session.Wait(300 * time.Second)
	return string(s.Out.Contents()) + string(s.Err.Contents()), nil
}

// ExecKusionWithWorkDir executes kusion command in work directory
func ExecKusionWithWorkDir(cli, dir string) (string, error) {
	var output []byte
	c := strings.Fields(cli)
	commandName := filepath.Join(GetKusionCLIBin(), c[0])
	command := exec.Command(commandName, c[1:]...)
	command.Dir = dir
	session, err := gexec.Start(command, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	if err != nil {
		return string(output), err
	}
	s := session.Wait(300 * time.Second)
	return string(s.Out.Contents()) + string(s.Err.Contents()), nil
}

// ExecKusionWithStdin executes kusion command in work directory with stdin
func ExecKusionWithStdin(cli, dir, input string) (string, error) {
	var output []byte
	c := strings.Fields(cli)
	commandName := filepath.Join(GetKusionCLIBin(), c[0])
	command := exec.Command(commandName, c[1:]...)
	command.Dir = dir
	subStdin, err := command.StdinPipe()
	if err != nil {
		return string(output), err
	}
	io.WriteString(subStdin, input)
	session, err := gexec.Start(command, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	if err != nil {
		return string(output), err
	}
	s := session.Wait(300 * time.Second)
	return string(s.Out.Contents()) + string(s.Err.Contents()), nil
}
