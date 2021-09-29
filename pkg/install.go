package pkg

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/rogpeppe/go-internal/lockedfile"
	log "github.com/sirupsen/logrus"
)

const (
	installFile    = "terraform"
	installVersion = "terraform_"
	installPath    = ".terraform.versions"
	lockFilePath   = "/tmp/simple-tfswitch.lock"
)

// getInstallLocation : get location where the terraform binary will be installed,
// will create a directory in the home location if it does not exist
func getInstallLocation() string {
	/* get current user */
	usr, errCurr := user.Current()
	if errCurr != nil {
		log.Fatal(errCurr)
	}

	userCommon := usr.HomeDir

	/* For snapcraft users, SNAP_USER_COMMON environment variable is set by default.
	 * tfswitch does not have permission to save to $HOME/.terraform.versions for snapcraft users
	 * tfswitch will save binaries into $SNAP_USER_COMMON/.terraform.versions */
	if os.Getenv("SNAP_USER_COMMON") != "" {
		userCommon = os.Getenv("SNAP_USER_COMMON")
	}

	/* set installation location */
	installLocation := filepath.Join(userCommon, installPath)

	/* Create local installation directory if it does not exist */
	CreateDirIfNotExist(installLocation)

	return installLocation
}

func WaitForLockFile() (unlock func()) {
	m := lockedfile.MutexAt(lockFilePath)
	unlock, err := m.Lock()
	if err != nil {
		log.Errorf("there was a problem while trying to acquire lockfile %v", lockFilePath)
		os.Exit(1)
	}

	return unlock
}

// Install : Install the provided version in the argument
func Install(tfversion string, mirrorURL string) (string, error) {
	if !ValidVersionFormat(tfversion) {
		log.Errorf("The provided terraform version format does not exist - %s.", tfversion)
		os.Exit(1)
	}

	// version install lockfile
	unlock := WaitForLockFile()
	defer unlock()

	installLocation := getInstallLocation() // get installation location -  this is where we will put our terraform binary file

	goarch := runtime.GOARCH
	goos := runtime.GOOS

	/* check if selected version already downloaded */
	installFileVersionPath := ConvertExecutableExt(filepath.Join(installLocation, installVersion+tfversion))
	fileExist := CheckFileExist(installFileVersionPath)

	/* if selected version already exist, */
	if fileExist {
		return installFileVersionPath, nil
	}

	// if does not have slash - append slash
	hasSlash := strings.HasSuffix(mirrorURL, "/")
	if !hasSlash {
		mirrorURL = fmt.Sprintf("%s/", mirrorURL)
	}

	/* if selected version already exist, */
	/* proceed to download it from the hashicorp release page */
	url := mirrorURL + tfversion + "/" + installVersion + tfversion + "_" + goos + "_" + goarch + ".zip"
	zipFile, errDownload := DownloadFromURL(installLocation, url)

	/* If unable to download file from url, exit(1) immediately */
	if errDownload != nil {
		return "", errDownload
	}

	/* unzip the downloaded zipfile */
	errUnzip := Unzip(zipFile, installLocation)
	if errUnzip != nil {
		log.Error("Unable to unzip downloaded zip file")

		return "", errUnzip
	}

	/* rename unzipped file to terraform version name - terraform_x.x.x */
	installFilePath := ConvertExecutableExt(filepath.Join(installLocation, installFile))
	RenameFile(installFilePath, installFileVersionPath)

	/* remove zipped file to clear clutter */
	RemoveFiles(zipFile)

	return installFileVersionPath, nil
}

// ConvertExecutableExt : convert excutable with local OS extension
func ConvertExecutableExt(fpath string) string {
	switch runtime.GOOS {
	case "windows":
		if filepath.Ext(fpath) == ".exe" {
			return fpath
		}

		return fpath + ".exe"
	default:
		return fpath
	}
}

// install when tf file is provided
func InstallTFProvidedModule(dir string, mirrorURL string) (string, error) {
	module, _ := tfconfig.LoadModule(dir)

	if len(module.RequiredCore) == 0 {
		return "", fmt.Errorf("no required_versions found")
	}
	tfconstraint := module.RequiredCore[0] // we skip duplicated definitions and use only first one

	return installFromConstraint(&tfconstraint, mirrorURL), nil
}

// install using a version constraint
func installFromConstraint(tfconstraint *string, mirrorURL string) string {
	listAll := true                            // set list all true - all versions including beta and rc will be displayed
	tflist, _ := GetTFList(mirrorURL, listAll) // get list of versions

	constrains, err := semver.NewConstraint(*tfconstraint) // NewConstraint returns a Constraints instance that a Version instance can be checked against
	if err != nil {
		log.Errorf("Error parsing constraint: %s, Please check constraint syntax on terraform file.", err)
		os.Exit(1)
	}
	versions := make([]*semver.Version, len(tflist))
	for i, tfvals := range tflist {
		version, err := semver.NewVersion(tfvals) // NewVersion parses a given version and returns an instance of Version or an error if unable to parse the version.
		if err != nil {
			log.Errorf("Error parsing version: %s", err)
			os.Exit(1)
		}

		versions[i] = version
	}

	sort.Sort(sort.Reverse(semver.Collection(versions)))

	for _, element := range versions {
		if constrains.Check(element) { // Validate a version against a constraint
			tfversion := element.String()
			if ValidVersionFormat(tfversion) { // check if version format is correct
				out, err := Install(tfversion, mirrorURL)
				if err != nil {
					log.Printf("Error during install %v", err)
					os.Exit(1)
				}

				return out
			}
			log.Errorf("Invalid terraform version format.")
			os.Exit(1)
		}
	}

	log.Errorf("No version found to match constraint. Follow the README.md instructions for setup. https://github.com/terraform-tools/simple-tfswitch/blob/main/README.md")
	os.Exit(1)

	return ""
}
