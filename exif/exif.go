package exif

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type (
	// ErrToolNotFound is returned if the system doesn't have the
	// exif tool installed of if the tool is installed but the version
	// is not valid.
	ErrToolNotFound struct {
		error
	}

	// Tool is the wrapper for exif
	Tool struct {
		binary string
	}
)

const (
	// AnyVersion is used to indicate that the caller doesn't care about
	// which version of exiftool is installed
	AnyVersion = ""
)

// Open the tool, the tool must be available in the OS PATH.
//
// expectedVersion will control which version of the tool is used, callers
// can use AnyVersion to indicate that version information is not relevant
func Open(expectedVersion string) (*Tool, error) {
	binary, err := exec.LookPath("exiftool")
	if err != nil {
		return nil, ErrToolNotFound{err}
	}
	if !filepath.IsAbs(binary) {
		// use absolute path to avoid problems if later
		// the working dir changes
		binary, err = filepath.Abs(binary)
		if err != nil {
			return nil, err
		}
	}
	tool := &Tool{
		binary: binary,
	}
	if err := tool.sanityCheck(expectedVersion); err != nil {
		return nil, err
	}
	return tool, nil
}

// sanityCheck verifies if the configured too is valid.
func (t *Tool) sanityCheck(expectedVersion string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, t.binary, "-ver")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	if expectedVersion == AnyVersion {
		// any version will do, just check if the version is not empty
		if len(buf) == 0 {
			return ErrToolNotFound{errors.New("calling exiftool -ver results in empty string")}
		}
		return nil
	}
	actualVersion := strings.TrimSpace(string(buf))
	if expectedVersion != actualVersion {
		return ErrToolNotFound{
			fmt.Errorf("version of exiftool should be %v but got %v", expectedVersion, actualVersion),
		}
	}
	return nil
}
