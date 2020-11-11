package main

/*
* Version 0.6.0
* Compatible with Mac OS X ONLY
 */

/*** OPERATION WORKFLOW ***/
/*
* 1- Create /usr/local/terraform directory if does not exist
* 2- Download zip file from url to /usr/local/terraform
* 3- Unzip the file to /usr/local/terraform
* 4- Rename the file from `terraform` to `terraform_version`
* 5- Remove the downloaded zip file
* 6- Read the existing symlink for terraform (Check if it's a homebrew symlink)
* 7- Remove that symlink (Check if it's a homebrew symlink)
* 8- Create new symlink to binary  `terraform_version`
 */

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

	// original hashicorp upstream have broken dependencies, so using fork as workaround
	// TODO: move back to upstream
	"github.com/Masterminds/semver"
	"github.com/kiranjthomas/terraform-config-inspect/tfconfig"

	//	"github.com/hashicorp/terraform-config-inspect/tfconfig"

	"github.com/manifoldco/promptui"
	"github.com/pborman/getopt"
	"github.com/spf13/viper"

	lib "github.com/warrensbox/terraform-switcher/lib"
)

const (
	hashiURL     = "https://releases.hashicorp.com/terraform/"
	defaultBin   = "/usr/local/bin/terraform" //default bin installation dir
	tfvFilename  = ".terraform-version"
	rcFilename   = ".tfswitchrc"
	tomlFilename = ".tfswitch.toml"
)

var version = "0.9.0\n"

func main() {

	custBinPath := getopt.StringLong("bin", 'b', defaultBin, "Custom binary path. For example: /Users/username/bin/terraform")
	listAllFlag := getopt.BoolLong("list-all", 'l', "List all versions of terraform - including beta and rc")
	versionFlag := getopt.BoolLong("version", 'v', "Displays the version of tfswitch")
	helpFlag := getopt.BoolLong("help", 'h', "Displays help message")
	_ = versionFlag

	getopt.Parse()
	args := getopt.Args()

	dir, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to get current directory %v\n", err)
		os.Exit(1)
	}

	tfvfile := dir + fmt.Sprintf("/%s", tfvFilename)         //settings for .terraform-version file in current directory (tfenv compatible)
	rcfile := dir + fmt.Sprintf("/%s", rcFilename)           //settings for .tfswitchrc file in current directory (backward compatible purpose)
	tomlconfigfile := dir + fmt.Sprintf("/%s", tomlFilename) //settings for .tfswitch.toml file in current directory (option to specify bin directory)

	switch {
	case *versionFlag:
		//if *versionFlag {
		fmt.Printf("\nVersion: %v\n", version)
	case *helpFlag:
		//} else if *helpFlag {
		usageMessage()
	/* Checks if the .tfswitch.toml file exist */
	/* This block checks to see if the tfswitch toml file is provided in the current path.
	 * If the .tfswitch.toml file exist, it has a higher precedence than the .tfswitchrc file
	 * You can specify the custom binary path and the version you desire
	 * If you provide a custom binary path with the -b option, this will override the bin value in the toml file
	 * If you provide a version on the command line, this will override the version value in the toml file
	 */
	case fileExists(tomlconfigfile):

		readingFileMsg(tomlFilename)
		version, binPath := getParamsTOML(custBinPath, dir)

		switch {
		case version != "":
			installVersion(version, &binPath)
		case len(args) == 1:
			installVersion(args[0], &binPath)
		case *listAllFlag:
			listAll := true //set list all true - all versions including beta and rc will be displayed
			installOption(listAll, &binPath)
		case checkTFModuleFileExist(dir):
			installTFProvidedModule(dir, &binPath)
		default:
			listAll := false //set list all false - only official release will be displayed
			installOption(listAll, &binPath)
		}

		// if len(args) == 1 { //if the version is passed in the command line
		// 	if lib.ValidVersionFormat(args[0]) {
		// 		requestedVersion := args[0]
		// 		listAll := true                                     //set list all true - all versions including beta and rc will be displayed
		// 		tflist, _ := lib.GetTFList(hashiURL, listAll)       //get list of versions
		// 		exist := lib.VersionExist(requestedVersion, tflist) //check if version exist before downloading it

		// 		if exist {
		// 			tfversion = requestedVersion // set tfversion = the version needed
		// 		}
		// 	} else if version != "" { // if the required version in the toml file is provided (use it)
		// 		tfversion = version
		// 	}
		// }

		// if *listAllFlag { //show all terraform version including betas and RCs
		// 	listAll := true //set list all true - all versions including beta and rc will be displayed
		// 	installOption(listAll, &binPath)
		// } else if tfversion == "" { // if no version is provided, show a dropdown of available release versions
		// 	listAll := false //set list all false - only official release will be displayed
		// 	installOption(listAll, &binPath)
		// } else {
		//if lib.ValidVersionFormat(tfversion) { //check if version is correct
		// 		lib.Install(tfversion, binPath)
		// 	} else {
		// 		printInvalidTFVersion()
		// 		os.Exit(1)
		// 	}
		// }
	/* list all versions, //show all terraform version including betas and RCs*/
	case *listAllFlag:
		installWithListAll(custBinPath)

	/* version provided on command line as arg */
	case len(args) == 1:
		installVersion(args[0], custBinPath)

	/* provide an tfswitchrc file */
	case fileExists(rcfile) && len(args) == 0:
		readingFileMsg(rcfile)
		tfversion := retrieveFileContents(rcfile)
		installVersion(tfversion, custBinPath)

	/* if .terraform-version file found */
	case fileExists(tfvfile) && len(args) == 0:
		readingFileMsg(tfvFilename)
		tfversion := retrieveFileContents(tfvfile)
		installVersion(tfversion, custBinPath)

	/* if versions.tf file found */
	case checkTFModuleFileExist(dir) && len(args) == 0:
		installTFProvidedModule(dir, custBinPath)

	// if no arg is provided
	default:
		listAll := false //set list all false - only official release will be displayed
		installOption(listAll, custBinPath)
	}
}

/* Helper functions */

// install with all possible versions, including beta and rc
func installWithListAll(custBinPath *string) {
	listAll := true //set list all true - all versions including beta and rc will be displayed
	installOption(listAll, custBinPath)
}

// install with provided version as argument
func installVersion(arg string, custBinPath *string) {
	if lib.ValidVersionFormat(arg) {
		requestedVersion := arg
		listAll := true                                     //set list all true - all versions including beta and rc will be displayed
		tflist, _ := lib.GetTFList(hashiURL, listAll)       //get list of versions
		exist := lib.VersionExist(requestedVersion, tflist) //check if version exist before downloading it

		if exist {
			lib.Install(requestedVersion, *custBinPath)
		} else {
			fmt.Println("The provided terraform version does not exist. Try `tfswitch -l` to see all available versions.")
		}

	} else {
		printInvalidTFVersion()
		fmt.Println("Args must be a valid terraform version")
		usageMessage()
	}
}

// Print invalid TF version
func printInvalidTFVersion() {
	fmt.Println("Invalid terraform version format. Format should be #.#.# or #.#.#-@# where # is numbers and @ is word characters. For example, 0.11.7 and 0.11.9-beta1 are valid versions")
}

//retrive file content of regular file
func retrieveFileContents(file string) string {
	fileContents, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Failed to read %s file. Follow the README.md instructions for setup. https://github.com/warrensbox/terraform-switcher/blob/master/README.md\n", tfvFilename)
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	tfversion := strings.TrimSuffix(string(fileContents), "\n")
	return tfversion
}

// Print message reading file content of :
func readingFileMsg(filename string) {
	fmt.Printf("Reading required terraform version %s \n", filename)
}

// fileExists checks if a file exists and is not a directory before we try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func checkTFModuleFileExist(dir string) bool {

	module, _ := tfconfig.LoadModule(dir)
	if len(module.RequiredCore) >= 1 {
		return true
	}
	return false
}

/* valid install */
// func simpleInstall(tfversion string, custBinPath *string) {
// 	if lib.ValidVersionFormat(tfversion) { //check if version is correct
// 		lib.Install(string(tfversion), *custBinPath)
// 	} else {
// 		fmt.Println("Invalid terraform version format. Format should be #.#.# or #.#.#-@# where # is numbers and @ is word characters. For example, 0.11.7 and 0.11.9-beta1 are valid versions")
// 		os.Exit(1)
// 	}
// }

/* parses everything in the toml file, return required version and bin path */
func getParamsTOML(custBinPath *string, dir string) (string, string) {

	fmt.Printf("Reading configuration from %s\n", tomlFilename)
	binPath := *custBinPath                         //takes the default bin (defaultBin) if user does not specify bin path
	configfileName := lib.GetFileName(tomlFilename) //get the config file
	viper.SetConfigType("toml")
	viper.SetConfigName(configfileName)
	viper.AddConfigPath(dir)

	errs := viper.ReadInConfig() // Find and read the config file
	if errs != nil {
		fmt.Printf("Unable to read %s provided\n", tomlFilename) // Handle errors reading the config file
		fmt.Println(errs)
		os.Exit(1) // exit immediately if config file provided but it is unable to read it
	}

	bin := viper.Get("bin")                  // read custom binary location
	if binPath == defaultBin && bin != nil { // if the bin path is the same as the default binary path and if the custom binary is provided in the toml file (use it)
		binPath = os.ExpandEnv(bin.(string))
	}
	version := viper.Get("version") //attempt to get the version if it's provided in the toml

	return version.(string), binPath
}

func usageMessage() {
	fmt.Print("\n\n")
	getopt.PrintUsage(os.Stderr)
	fmt.Println("Supply the terraform version as an argument, or choose from a menu")
}

/* installOption : displays & installs tf version */
/* listAll = true - all versions including beta and rc will be displayed */
/* listAll = false - only official stable release are displayed */
func installOption(listAll bool, custBinPath *string) {

	tflist, _ := lib.GetTFList(hashiURL, listAll) //get list of versions
	recentVersions, _ := lib.GetRecentVersions()  //get recent versions from RECENT file
	tflist = append(recentVersions, tflist...)    //append recent versions to the top of the list
	tflist = lib.RemoveDuplicateVersions(tflist)  //remove duplicate version

	/* prompt user to select version of terraform */
	prompt := promptui.Select{
		Label: "Select Terraform version",
		Items: tflist,
	}

	_, tfversion, errPrompt := prompt.Run()
	tfversion = strings.Trim(tfversion, " *recent") //trim versions with the string " *recent" appended

	if errPrompt != nil {
		log.Printf("Prompt failed %v\n", errPrompt)
		os.Exit(1)
	}

	lib.Install(tfversion, *custBinPath)
	os.Exit(0)
}

// installation when
func installTFProvidedModule(dir string, custBinPath *string) {
	tfversion := ""
	module, _ := tfconfig.LoadModule(dir)
	tfconstraint := module.RequiredCore[0]        //we skip duplicated definitions and use only first one
	listAll := true                               //set list all true - all versions including beta and rc will be displayed
	tflist, _ := lib.GetTFList(hashiURL, listAll) //get list of versions
	fmt.Printf("Reading required version from terraform file, constraint: %s\n", tfconstraint)

	constrains, err := semver.NewConstraint(tfconstraint) //NewConstraint returns a Constraints instance that a Version instance can be checked against
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
			fmt.Printf("Matched version: %s\n", tfversion)
			if lib.ValidVersionFormat(tfversion) { //check if version format is correct
				lib.Install(tfversion, *custBinPath)
			} else {
				printInvalidTFVersion()
				os.Exit(1)
			}
		}
	}

	fmt.Println("No version found to match constraint. Follow the README.md instructions for setup. https://github.com/warrensbox/terraform-switcher/blob/master/README.md")
	os.Exit(1)
}
