package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andrebq/exiftool2json/exif"
)

func TestTagsEndpoint(t *testing.T) {
	tool, err := exif.Open("")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(NewAPI(tool))
	response, err := http.Get(fmt.Sprintf("%v/tags", server.URL))
	if err != nil {
		t.Fatalf("Unable to execute request to server: %v", err)
	}
	defer response.Body.Close()
	decodedBody := struct {
		Tags []TagInfo
	}{}
	dec := json.NewDecoder(response.Body)
	err = dec.Decode(&decodedBody)
	if err != nil {
		t.Errorf("Unable to decode body from HTTP response: %v", err)
	}
	if len(decodedBody.Tags) == 0 {
		t.Errorf("Should have at least one tag")
	}
}
