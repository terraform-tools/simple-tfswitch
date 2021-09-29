package main

import (
	"fmt"
	"log"
	"os"

	"github.com/terraform-tools/simple-tfswitch/pkg"
)

const (
	mirrorURL = "https://releases.hashicorp.com/terraform"
)

func main() {
	args := os.Args
	dir, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to get current directory %v\n", err)
		os.Exit(1)
	}

	tfBinaryPath, err := pkg.InstallTFProvidedModule(dir, mirrorURL)
	if err != nil {
		fmt.Println("Error occurred:", err)
	}

	exitCode := pkg.RunTerraform(tfBinaryPath, args[1:]...)
	os.Exit(exitCode)
}
