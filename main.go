package main

import (
	"errors"
	"log"

	"github.com/andrebq/exiftool2json/exif"
)

func main() {
	tool, err := exif.Open(exif.AnyVersion)
	var notFound exif.ErrToolNotFound
	if errors.As(err, &notFound) {
		log.Fatalf("exiftool invalid or not found: %v", notFound)
	}
	_ = tool
}
