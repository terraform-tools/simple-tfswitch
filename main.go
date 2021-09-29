package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/terraform-tools/simple-tfswitch/pkg"
	"github.com/terraform-tools/simple-tfswitch/pkg/logger"
)

const (
	mirrorURL = "https://releases.hashicorp.com/terraform"
)

func main() {
	args := os.Args
	dir, err := os.Getwd()

	logger.Setup()

	if err != nil {
		log.Errorf("Failed to get current directory %v", err)
		os.Exit(1)
	}

	tfBinaryPath, err := pkg.InstallTFProvidedModule(dir, mirrorURL)
	if err != nil {
		log.Errorln("Error occurred:", err)
	}

	exitCode := pkg.RunTerraform(tfBinaryPath, args[1:]...)
	os.Exit(exitCode)
}
