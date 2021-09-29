package pkg_test

import (
	"log"
	"reflect"
	"testing"

	"github.com/terraform-tools/simple-tfswitch/pkg"
)

const (
	hashiURL = "https://releases.hashicorp.com/terraform/"
)

// TestGetTFList : Get list from hashicorp
func TestGetTFList(t *testing.T) {
	listAll := true
	list, _ := pkg.GetTFList(hashiURL, listAll)

	val := "0.1.0"
	var exists bool

	switch reflect.TypeOf(list).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(list)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				exists = true
			}
		}
	}

	if !exists {
		log.Fatalf("Not able to find version: %s\n", val)
	} else {
		t.Log("Write versions exist (expected)")
	}
}

// TestValidVersionFormat : test if func returns valid version format
// more regex testing at https://rubular.com/r/UvWXui7EU2icSb
func TestValidVersionFormat(t *testing.T) {
	var version string
	version = "0.11.8"

	valid := pkg.ValidVersionFormat(version)

	if valid == true {
		t.Logf("Valid version format : %s (expected)", version)
	} else {
		log.Fatalf("Failed to verify version format: %s\n", version)
	}

	version = "1.11.9"

	valid = pkg.ValidVersionFormat(version)

	if valid == true {
		t.Logf("Valid version format : %s (expected)", version)
	} else {
		log.Fatalf("Failed to verify version format: %s\n", version)
	}

	version = "1.11.a"

	valid = pkg.ValidVersionFormat(version)

	if valid == false {
		t.Logf("Invalid version format : %s (expected)", version)
	} else {
		log.Fatalf("Failed to verify version format: %s\n", version)
	}

	version = "22323"

	valid = pkg.ValidVersionFormat(version)

	if valid == false {
		t.Logf("Invalid version format : %s (expected)", version)
	} else {
		log.Fatalf("Failed to verify version format: %s\n", version)
	}

	version = "@^&*!)!"

	valid = pkg.ValidVersionFormat(version)

	if valid == false {
		t.Logf("Invalid version format : %s (expected)", version)
	} else {
		log.Fatalf("Failed to verify version format: %s\n", version)
	}

	version = "1.11.9-beta1"

	valid = pkg.ValidVersionFormat(version)

	if valid == true {
		t.Logf("Valid version format : %s (expected)", version)
	} else {
		log.Fatalf("Failed to verify version format: %s\n", version)
	}

	version = "0.12.0-rc2"

	valid = pkg.ValidVersionFormat(version)

	if valid == true {
		t.Logf("Valid version format : %s (expected)", version)
	} else {
		log.Fatalf("Failed to verify version format: %s\n", version)
	}

	version = "1.11.4-boom"

	valid = pkg.ValidVersionFormat(version)

	if valid == true {
		t.Logf("Valid version format : %s (expected)", version)
	} else {
		log.Fatalf("Failed to verify version format: %s\n", version)
	}

	version = "1.11.4-1"

	valid = pkg.ValidVersionFormat(version)

	if valid == false {
		t.Logf("Invalid version format : %s (expected)", version)
	} else {
		log.Fatalf("Failed to verify version format: %s\n", version)
	}
}
