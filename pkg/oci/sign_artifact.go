package oci

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"

	"kusionstack.io/kusion/pkg/log"
)

// Original copy from timoni with minimal change.

// SignArtifact signs an OpenContainers artifact using the specified provider.
func SignArtifact(provider, registryURL, keyRef string) error {
	ref, err := ParseArtifactRef(registryURL)
	if err != nil {
		return err
	}

	switch provider {
	case "cosign":
		if err := SignCosign(ref.String(), keyRef); err != nil {
			return err
		}
	default:
		return fmt.Errorf("signer not supported: %s", provider)
	}
	return nil
}

// SignCosign signs an image (`imageRef`) using a cosign private key (`keyRef`)
func SignCosign(imageRef, keyRef string) error {
	cosignExecutable, err := exec.LookPath("cosign")
	if err != nil {
		return fmt.Errorf("executing cosign failed: %w", err)
	}

	cosignCmd := exec.Command(cosignExecutable, []string{"sign"}...)
	cosignCmd.Env = os.Environ()

	// if key is empty, use keyless mode
	if keyRef != "" {
		cosignCmd.Args = append(cosignCmd.Args, "--key", keyRef)
	}

	cosignCmd.Args = append(cosignCmd.Args, "--yes")
	cosignCmd.Args = append(cosignCmd.Args, imageRef)

	err = processCosignIO(cosignCmd)
	if err != nil {
		return err
	}

	return cosignCmd.Wait()
}

func processCosignIO(cosignCmd *exec.Cmd) error {
	stdout, err := cosignCmd.StdoutPipe()
	if err != nil {
		log.Error(err, "cosign stdout pipe failed")
	}
	stderr, err := cosignCmd.StderrPipe()
	if err != nil {
		log.Error(err, "cosign stderr pipe failed")
	}

	merged := io.MultiReader(stdout, stderr)
	scanner := bufio.NewScanner(merged)

	if err := cosignCmd.Start(); err != nil {
		return fmt.Errorf("executing cosign failed: %w", err)
	}

	for scanner.Scan() {
		log.Info("cosign: " + scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Error(err, "cosign stdout/stderr scanner failed")
	}

	return nil
}
