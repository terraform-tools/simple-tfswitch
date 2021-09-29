package pkg

import (
	"errors"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func RunTerraform(tfBinaryPath string, args ...string) int {
	cmd := exec.Command(tfBinaryPath, args...)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		log.Errorf("Failed launching terraform binary %v", err)

		return -1
	}

	return 0
}
