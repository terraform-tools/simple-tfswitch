package lib_test

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/terraform-tools/simple-tfswitch/lib"
)

// TestRenameFile : Create a file, check filename exist,
// rename file, check new filename exit
func TestRenameFile(t *testing.T) {
	installFile := lib.ConvertExecutableExt("terraform")
	installVersion := "terraform_"
	installPath := "/.terraform.versions_test/"
	version := "0.0.7"

	usr, errCurr := user.Current()
	if errCurr != nil {
		log.Fatal(errCurr)
	}
	installLocation := filepath.Join(usr.HomeDir, installPath)

	createDirIfNotExist(installLocation)

	installFilePath := filepath.Join(installLocation, installFile)

	createFile(installFilePath)

	if exist := checkFileExist(installFilePath); exist {
		t.Logf("File exist %v", installFilePath)
	} else {
		t.Logf("File does not exist %v", installFilePath)
		t.Error("Missing file")
	}

	installVersionFilePath := lib.ConvertExecutableExt(filepath.Join(installLocation, installVersion+version))

	lib.RenameFile(installFilePath, installVersionFilePath)

	if exist := checkFileExist(installVersionFilePath); exist {
		t.Logf("New file exist %v", installVersionFilePath)
	} else {
		t.Logf("New file does not exist %v", installVersionFilePath)
		t.Error("Missing new file")
	}

	if exist := checkFileExist(installFilePath); exist {
		t.Logf("Old file should not exist %v", installFilePath)
		t.Error("Did not rename file")
	} else {
		t.Logf("Old file does not exist %v", installFilePath)
	}

	cleanUp(installLocation)
}

// TestRemoveFiles : Create a file, check file exist,
// remove file, check file does not exist
func TestRemoveFiles(t *testing.T) {
	installFile := lib.ConvertExecutableExt("terraform")
	installPath := "/.terraform.versions_test/"

	usr, errCurr := user.Current()
	if errCurr != nil {
		log.Fatal(errCurr)
	}
	installLocation := filepath.Join(usr.HomeDir, installPath)

	createDirIfNotExist(installLocation)

	installFilePath := filepath.Join(installLocation, installFile)

	createFile(installFilePath)

	if exist := checkFileExist(installFilePath); exist {
		t.Logf("File exist %v", installFilePath)
	} else {
		t.Logf("File does not exist %v", installFilePath)
		t.Error("Missing file")
	}

	lib.RemoveFiles(installFilePath)

	if exist := checkFileExist(installFilePath); exist {
		t.Logf("Old file should not exist %v", installFilePath)
		t.Error("Did not remove file")
	} else {
		t.Logf("Old file does not exist %v", installFilePath)
	}

	cleanUp(installLocation)
}

// TestUnzip : Create a file, check file exist,
// remove file, check file does not exist
func TestUnzip(t *testing.T) {
	installPath := "/.terraform.versions_test/"
	absPath, _ := filepath.Abs("../test-data/test-data.zip")

	fmt.Println(absPath)

	usr, errCurr := user.Current()
	if errCurr != nil {
		log.Fatal(errCurr)
	}
	installLocation := filepath.Join(usr.HomeDir, installPath)

	createDirIfNotExist(installLocation)

	files, errUnzip := lib.Unzip(absPath, installLocation)

	if errUnzip != nil {
		fmt.Println("Unable to unzip zip file")
		log.Fatal(errUnzip)
		os.Exit(1)
	}

	tst := strings.Join(files, "")

	if exist := checkFileExist(tst); exist {
		t.Logf("File exist %v", tst)
	} else {
		t.Logf("File does not exist %v", tst)
		t.Error("Missing file")
	}

	cleanUp(installLocation)
}

// TestCreateDirIfNotExist : Create a directory, check directory exist
func TestCreateDirIfNotExist(t *testing.T) {
	installPath := "/.terraform.versions_test/"

	usr, errCurr := user.Current()
	if errCurr != nil {
		log.Fatal(errCurr)
	}
	installLocation := filepath.Join(usr.HomeDir, installPath)

	cleanUp(installLocation)

	if _, err := os.Stat(installLocation); os.IsNotExist(err) {
		t.Logf("Directory should not exist %v (expected)", installLocation)
	} else {
		t.Logf("Directory already exist %v (unexpected)", installLocation)
		t.Error("Directory should not exist")
	}

	lib.CreateDirIfNotExist(installLocation)
	t.Logf("Creating directory %v", installLocation)

	if _, err := os.Stat(installLocation); err == nil {
		t.Logf("Directory exist %v (expected)", installLocation)
	} else {
		t.Logf("Directory should exist %v (unexpected)", installLocation)
		t.Error("Directory should exist")
	}

	cleanUp(installLocation)
}

// TestPath : create file in directory, check if path exist
func TestPath(t *testing.T) {
	installPath := "/.terraform.versions_test"
	installFile := lib.ConvertExecutableExt("terraform")

	usr, errCurr := user.Current()
	if errCurr != nil {
		log.Fatal(errCurr)
	}
	installLocation := filepath.Join(usr.HomeDir, installPath)

	createDirIfNotExist(installLocation)

	installFilePath := filepath.Join(installLocation, installFile)
	createFile(installFilePath)

	path := lib.Path(installFilePath)

	t.Logf("Path created %s\n", installFilePath)
	t.Logf("Path expected %s\n", installLocation)
	t.Logf("Path from library %s\n", path)
	if path == installLocation {
		t.Logf("Path exist (expected)")
	} else {
		t.Error("Path does not exist (unexpected)")
	}

	cleanUp(installLocation)
}

// TestConvertExecutableExt : convert executable binary with extension
func TestConvertExecutableExt(t *testing.T) {
	usr, errCurr := user.Current()
	if errCurr != nil {
		log.Fatal(errCurr)
	}

	installPath := "/.terraform.versions_test/"
	test_array := []string{
		"terraform",
		"terraform.exe",
		filepath.Join(usr.HomeDir, installPath, "terraform"),
		filepath.Join(usr.HomeDir, installPath, "terraform.exe"),
	}

	for _, fpath := range test_array {
		fpathExt := lib.ConvertExecutableExt((fpath))
		outputMsg := fpath + " converted to " + fpathExt + " on " + runtime.GOOS

		switch runtime.GOOS {
		case "windows":
			if filepath.Ext(fpathExt) != ".exe" {
				t.Error(outputMsg + " (unexpected)")
				continue
			}

			if filepath.Ext(fpath) == ".exe" {
				if fpathExt != fpath {
					t.Error(outputMsg + " (unexpected)")
				} else {
					t.Logf(outputMsg + " (expected)")
				}
				continue
			}

			if fpathExt != fpath+".exe" {
				t.Error(outputMsg + " (unexpected)")
				continue
			}

			t.Logf(outputMsg + " (expected)")
		default:
			if fpath != fpathExt {
				t.Error(outputMsg + " (unexpected)")
				continue
			}

			t.Logf(outputMsg + " (expected)")
		}
	}
}
