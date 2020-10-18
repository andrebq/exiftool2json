package exif

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestOpenTool(t *testing.T) {
	_, err := Open("")
	if err != nil {
		t.Fatal(err)
	}
}

func TestFailedOpen(t *testing.T) {
	version := "this-version-doesnt-exist"
	_, err := Open(version)
	var notfound ErrToolNotFound
	if !errors.As(err, &notfound) {
		t.Errorf("When the version is invalid should have returned a ErrToolNotFound but got %#v", err)
	} else if !strings.Contains(notfound.Error(), version) {
		t.Errorf("When the version is invalid, the error message should contain the desired version but got %v", notfound.Error())
	}

	path := os.Getenv("PATH")
	defer os.Setenv("PATH", path)
	os.Setenv("PATH", "")
	_, err = Open("")
	if !errors.As(err, &notfound) {
		t.Errorf("When the binary is not in the PATH, open should return ErrToolNotFound but got %#v", err)
	}
}
