package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/andrebq/exiftool2json/exif"
)

type (
	// Tags implements the /tags endpoint
	Tags struct {
		tool *exif.Tool
	}

	// TagInfo contains the information about an exif tag
	TagInfo struct {
		Writable    bool              `json:"writable"`
		Path        string            `json:"path"`
		Group       string            `json:"group"`
		Description map[string]string `json:"description"`
		Type        string            `json:"type"`
	}

	tagWriter struct {
		w        http.ResponseWriter
		enc      *json.Encoder
		primed   bool
		err      error
		putComma bool
	}
)

// ServeHTTP returns the JSON encoded data from exif tags
func (t *Tags) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx, cancelOnError := context.WithCancel(req.Context())
	// use cancelOnError to stop the underlying process in case there
	// is an error writing data to the client...
	//
	// no need to waste resources with an output that nobody will
	// consume
	defer cancelOnError()

	errCh := make(chan error, 1)
	// tradeoff between adding more memory consumption to Go runtime
	// so that the exiftool process can exit faster
	//
	// might not make sense if the number of concurrent requests is
	// high
	tags := make(chan exif.Tag, 1000)

	go func() {
		err := t.tool.WalkTags(req.Context(), func(t exif.Tag) error {
			select {
			case <-ctx.Done():
				return exif.ErrStop
			case tags <- t:
			}
			return nil
		})
		if err != nil {
			errCh <- err
		}
		close(tags)
	}()
	tw := &tagWriter{
		w:   w,
		enc: json.NewEncoder(w),
	}
	for {
		select {
		case err := <-errCh:
			// this logging call should be sampled to avoid overloading the log output
			log.Printf("Unexpected error while processing the request: %v", err)
			http.Error(w, "Unable to process your request.... try again later", http.StatusServiceUnavailable)
			return
		case t, open := <-tags:
			if !open {
				tw.close()
				return
			}
			if !tw.primed {
				tw.prime()
			}
			tw.write(t)

			if tw.err != nil {
				return
			}
		}
	}
}

// NewAPI returns a muxer with all the HTTP routes pre-configured
func NewAPI(tool *exif.Tool) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/tags", &Tags{tool: tool})
	return mux
}

func (tw *tagWriter) write(t exif.Tag) {
	if tw.err != nil {
		return
	}
	if tw.putComma {
		tw.writeComma()
	}
	if tw.err != nil {
		return
	}
	tw.err = tw.enc.Encode(TagInfo{
		Writable:    t.Wriatable,
		Path:        fmt.Sprintf("%v:%v", t.Table.Name, t.Name),
		Group:       t.Table.Name,
		Type:        t.Type,
		Description: descriptionsToMap(t.Description),
	})
	tw.putComma = true
}

func (tw *tagWriter) writeComma() {
	if tw.err != nil {
		return
	}
	_, tw.err = fmt.Fprint(tw.w, ",")
}

func descriptionsToMap(d []exif.Description) map[string]string {
	m := make(map[string]string)
	for _, v := range d {
		m[v.Lang] = v.Text
	}
	return m
}

func (tw *tagWriter) close() {
	if tw.err != nil {
		return
	}
	_, tw.err = fmt.Fprintf(tw.w, "]}")
	tw.logErr()
}

func (tw *tagWriter) prime() {
	tw.w.WriteHeader(http.StatusOK)
	_, tw.err = fmt.Fprintf(tw.w, `{"tags": [`)
	tw.primed = true
}

func (tw *tagWriter) logErr() {
	if tw.err != nil {
		// this logging call should be sampled to avoid overloading the log output
		log.Printf("Client error while sending data: %v", tw.err)
	}
}
