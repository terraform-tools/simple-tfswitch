package pkg

import "fmt"

// Print invalid TF version
func printInvalidTFVersion() {
	fmt.Println("Invalid terraform version format. Format should be #.#.# or #.#.#-@# where # are numbers and @ are word characters. For example, 0.11.7 and 0.11.9-beta1 are valid versions")
}
