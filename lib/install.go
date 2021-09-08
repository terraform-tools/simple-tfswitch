package lib

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rogpeppe/go-internal/lockedfile"
)

const (
	installFile    = "terraform"
	installVersion = "terraform_"
	installPath    = ".terraform.versions"
	lockFilePath   = "/tmp/simple-tfswitch.lock"
)

var (
	installLocation = "/tmp"
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
	installLocation = filepath.Join(userCommon, installPath)

	/* Create local installation directory if it does not exist */
	CreateDirIfNotExist(installLocation)

	return installLocation

}

func WaitForLockFile() (unlock func()) {
	m := lockedfile.MutexAt(lockFilePath)
	unlock, err := m.Lock()
	if err != nil {
		fmt.Printf("there was a problem while trying to aquire lockfile %v", lockFilePath)
		os.Exit(1)
	}
	return unlock
}

//Install : Install the provided version in the argument
func Install(tfversion string, mirrorURL string) string {

	if !ValidVersionFormat(tfversion) {
		fmt.Printf("The provided terraform version format does not exist - %s. Try `tfswitch -l` to see all available versions.\n", tfversion)
		os.Exit(1)
	}

	// version install lockfile
	unlock := WaitForLockFile()
	defer unlock()

	installLocation = getInstallLocation() //get installation location -  this is where we will put our terraform binary file

	goarch := runtime.GOARCH
	goos := runtime.GOOS

	// TODO: Workaround for macos arm64 since terraform doesn't have a binary for it yet
	if goos == "darwin" && goarch == "arm64" {
		goarch = "amd64"
	}

	/* check if selected version already downloaded */
	installFileVersionPath := ConvertExecutableExt(filepath.Join(installLocation, installVersion+tfversion))
	fileExist := CheckFileExist(installFileVersionPath)

	/* if selected version already exist, */
	if fileExist {
		return installFileVersionPath
	}

	//if does not have slash - append slash
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
		fmt.Println(errDownload)
		os.Exit(1)
	}

	/* unzip the downloaded zipfile */
	_, errUnzip := Unzip(zipFile, installLocation)
	if errUnzip != nil {
		fmt.Println("[Error] : Unable to unzip downloaded zip file")
		log.Fatal(errUnzip)
		os.Exit(1)
	}

	/* rename unzipped file to terraform version name - terraform_x.x.x */
	installFilePath := ConvertExecutableExt(filepath.Join(installLocation, installFile))
	RenameFile(installFilePath, installFileVersionPath)

	/* remove zipped file to clear clutter */
	RemoveFiles(zipFile)

	return installFileVersionPath
}

//ConvertExecutableExt : convert excutable with local OS extension
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
