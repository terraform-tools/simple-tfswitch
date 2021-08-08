package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"

	lib "github.com/jb-abbadie/simple-tfswitch/lib"
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

	tfBinaryPath := installTFProvidedModule(dir, mirrorURL)

	exitCode := runTerraform(tfBinaryPath, args...)
	os.Exit(exitCode)
}

/* Helper functions */

// Print invalid TF version
func printInvalidTFVersion() {
	fmt.Println("Invalid terraform version format. Format should be #.#.# or #.#.#-@# where # are numbers and @ are word characters. For example, 0.11.7 and 0.11.9-beta1 are valid versions")
}

// install when tf file is provided
func installTFProvidedModule(dir string, mirrorURL string) string {
	module, _ := tfconfig.LoadModule(dir)
	tfconstraint := module.RequiredCore[0] //we skip duplicated definitions and use only first one
	return installFromConstraint(&tfconstraint, mirrorURL)
}

// install using a version constraint
func installFromConstraint(tfconstraint *string, mirrorURL string) string {
	tfversion := ""
	listAll := true                                //set list all true - all versions including beta and rc will be displayed
	tflist, _ := lib.GetTFList(mirrorURL, listAll) //get list of versions

	constrains, err := semver.NewConstraint(*tfconstraint) //NewConstraint returns a Constraints instance that a Version instance can be checked against
	if err != nil {
		fmt.Printf("Error parsing constraint: %s\nPlease check constrain syntax on terraform file.\n", err)
		fmt.Println()
		os.Exit(1)
	}
	versions := make([]*semver.Version, len(tflist))
	for i, tfvals := range tflist {
		version, err := semver.NewVersion(tfvals) //NewVersion parses a given version and returns an instance of Version or an error if unable to parse the version.
		if err != nil {
			fmt.Printf("Error parsing version: %s", err)
			os.Exit(1)
		}

		versions[i] = version
	}

	sort.Sort(sort.Reverse(semver.Collection(versions)))

	for _, element := range versions {
		if constrains.Check(element) { // Validate a version against a constraint
			tfversion = element.String()
			if lib.ValidVersionFormat(tfversion) { //check if version format is correct
				return Install(tfversion, mirrorURL)
			} else {
				printInvalidTFVersion()
				os.Exit(1)
			}
		}
	}

	fmt.Println("No version found to match constraint. Follow the README.md instructions for setup. https://github.com/jb-abbadie/simple-tfswitch/blob/master/README.md")
	os.Exit(1)
	return ""
}

func Install(tfversion string, mirrorURL string) string {
	return lib.Install(tfversion, mirrorURL)
}

func runTerraform(tfBinaryPath string, args ...string) int {
	cmd := exec.Command(tfBinaryPath, args...)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode()
		}
	}
	return 0
}
