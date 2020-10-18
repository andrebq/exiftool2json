package exif

import (
	"context"
	"testing"
	"time"
)

func TestListX(t *testing.T) {
	tool, err := Open("")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()

	var tagsProcessed int
	acc := func(tag Tag) error {
		tagsProcessed++
		return nil
	}

	err = tool.WalkTags(ctx, acc)
	if err != nil {
		t.Errorf("Listing of items should have succeded without issues but %v", err)
	}

	if tagsProcessed == 0 {
		t.Errorf("Unexpected outcome... Zero tags found...")
	}
}

func TestEarlyStop(t *testing.T) {
	tool, err := Open("")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()

	var tagsProcessed int
	acc := func(tag Tag) error {
		tagsProcessed++
		return ErrStop
	}

	err = tool.WalkTags(ctx, acc)
	if err != nil {
		t.Errorf("Listing of items should have succeded without issues but %v", err)
	}

	if tagsProcessed != 1 {
		t.Errorf("Unexpected outcome... Should have stoped at the first tag but processed %v", tagsProcessed)
	}
}
