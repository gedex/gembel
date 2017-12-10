package main

import (
	"path/filepath"
	"testing"
)

func Test_ReadConfig(t *testing.T) {
	// Path does not exit.
	if _, err := ReadConfig("path-does-not-exist"); err == nil {
		t.Error("expect ReadConfig to return error when reading inexistence file")
	}

	// Invalid JSON config file
	files, err := filepath.Glob("testdata/error-*.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if _, err := ReadConfig(f); err == nil {
			t.Errorf("expect ReadConfig to return error when reading %s", f)
		}
	}

	// Valid JSON config file.
	files, err = filepath.Glob("testdata/valid-*.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if _, err := ReadConfig(f); err != nil {
			t.Errorf("expect ReadConfig to NOT return error when reading %s", f)
		}
	}

}
