package pkg

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// DownloadFromURL : Downloads the binary from the source url
func DownloadFromURL(installLocation string, url string) (string, error) {
	tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]
	log.Debugf("Downloading to: %s", installLocation)

	response, err := HTTPClient().Get(url)
	if err != nil {
		log.Errorln("Error while downloading", url, "-", err)

		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		// Sometimes hashicorp terraform file names are not consistent
		// For example 0.12.0-alpha4 naming convention in the release repo is not consistent
		return "", fmt.Errorf("unable to download from %s", url)
	}

	zipFile := filepath.Join(installLocation, fileName)
	output, err := os.Create(zipFile)
	if err != nil {
		log.Errorln("Error while creating", zipFile, "-", err)

		return "", err
	}
	defer output.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		log.Errorln("Error while downloading", url, "-", err)

		return "", err
	}

	log.Debugln(n, "bytes downloaded")

	return zipFile, nil
}
