package lib

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type tfVersionList struct {
	tflist []string
}

// GetTFList :  Get the list of available terraform version given the hashicorp url
func GetTFList(mirrorURL string, preRelease bool) ([]string, error) {
	result, err := GetTFURLBody(mirrorURL)
	if err != nil {
		return nil, err
	}

	var tfVersionList tfVersionList
	var semver string
	if preRelease {
		// Getting versions from body; should return match /X.X.X-@/ where X is a number,@ is a word character between a-z or A-Z
		semver = `\/(\d+\.\d+\.\d+)(-[a-zA-z]+\d*)?\/`
	} else if !preRelease {
		// Getting versions from body; should return match /X.X.X/ where X is a number
		semver = `\/(\d+\.\d+\.\d+)\/`
	}
	r, _ := regexp.Compile(semver)
	for i := range result {
		if r.MatchString(result[i]) {
			str := r.FindString(result[i])
			trimstr := strings.Trim(str, "/") // remove "/" from /X.X.X/
			tfVersionList.tflist = append(tfVersionList.tflist, trimstr)
		}
	}

	if len(tfVersionList.tflist) == 0 {
		fmt.Printf("Cannot get list from mirror: %s\n", mirrorURL)
	}

	return tfVersionList.tflist, nil
}

// GetTFURLBody : Get list of terraform versions from hashicorp releases
func GetTFURLBody(mirrorURL string) ([]string, error) {
	hasSlash := strings.HasSuffix(mirrorURL, "/")
	if !hasSlash { // if does not have slash - append slash
		mirrorURL = fmt.Sprintf("%s/", mirrorURL)
	}
	resp, errURL := http.Get(mirrorURL)
	if errURL != nil {
		log.Printf("[Error] : Getting url: %v", errURL)
		os.Exit(1)
		return nil, errURL
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("[Error] : Retrieving contents from url: %s", mirrorURL)
		return nil, fmt.Errorf("[Error] : Retrieving contents from url: %s", mirrorURL)
	}

	body, errBody := ioutil.ReadAll(resp.Body)
	if errBody != nil {
		log.Printf("[Error] : reading body: %v", errBody)
		return nil, errBody
	}

	bodyString := string(body)
	result := strings.Split(bodyString, "\n")

	return result, nil
}

// ValidVersionFormat : returns valid version format
/* For example: 0.1.2 = valid
// For example: 0.1.2-beta1 = valid
// For example: 0.1.2-alpha = valid
// For example: a.1.2 = invalid
// For example: 0.1. 2 = invalid
*/
func ValidVersionFormat(version string) bool {
	// Getting versions from body; should return match /X.X.X-@/ where X is a number,@ is a word character between a-z or A-Z
	// Follow https://semver.org/spec/v1.0.0-beta.html
	// Check regular expression at https://rubular.com/r/ju3PxbaSBALpJB
	semverRegex := regexp.MustCompile(`^(\d+\.\d+\.\d+)(-[a-zA-z]+\d*)?$`)

	return semverRegex.MatchString(version)
}
