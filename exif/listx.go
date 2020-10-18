package exif

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os/exec"
)

type (
	// TagFunc is used by WalkTags to iterate over all tags,
	// users can return ErrStop error to indicate that they don't want
	// to process further tags.
	TagFunc func(Tag) error

	// Tag contains details about an ExIF Tag
	Tag struct {
		Table       TableInfo     `xml:"-"`
		Writable    bool          `xml:"writable,attr"`
		Type        string        `xml:"type,attr"`
		Name        string        `xml:"name,attr"`
		Description []Description `xml:"desc"`
	}

	// Description of a given tag
	Description struct {
		Lang string `xml:"lang,attr"`
		Text string `xml:",innerxml"`
	}

	// TableInfo contains information about an ExIF tag table
	TableInfo struct {
		// Name of this table
		Name string `xml:"name,attr"`
	}
)

var (
	// ErrStop is used by TagFunc implementations to indicate that
	// the Walk process is completed
	ErrStop = errors.New("stop")
)

// WalkTags from -listx calling handler for each tag processed. See TagFunc
// for more information.
//
// The function will abort as soon as the first error is found, if the first
// error is ErrStop the error is silenced, otherwise it is passed as-is,
// to callers.
//
// The operation works as a stream, so it is possible that a mal-formed XML
// document is partially decoded, usually this is not a concern since the
// underlying tool will always generate valid XMLs, but if the Context is
// Done before the whole XML is streamed, partial data might have been processed
// by handler.
func (t *Tool) WalkTags(ctx context.Context, handler TagFunc) error {
	var earlyStop context.CancelFunc
	ctx, earlyStop = context.WithCancel(ctx)
	defer earlyStop()

	cmd := exec.CommandContext(ctx, t.binary, "-listx")
	xmlBytes, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	go func() {
		cmd.Wait()
	}()

	dec := xml.NewDecoder(xmlBytes)
	var streamCompleted bool
	for !streamCompleted {
		tk, err := dec.RawToken()
		if err != nil {
			streamCompleted = true
			if errors.Is(err, io.EOF) {
				// EOF is expected here
				err = nil
			}
		}
		switch tk := tk.(type) {
		case xml.StartElement:
			if tk.Name.Local == "taginfo" {
				err = processTagInfo(dec, handler)
			}
		default:
			// ignore any other tag
		}
		if err == ErrStop {
			err = nil
			streamCompleted = true
		}
	}

	return err
}

func processTagInfo(dec *xml.Decoder, handle TagFunc) error {
	for {
		tk, err := dec.RawToken()
		if err != nil {
			return err
		}
		switch tk := tk.(type) {
		case xml.EndElement:
			if tk.Name.Local == "taginfo" {
				return nil
			}
		case xml.StartElement:
			if tk.Name.Local == "table" {
				err = processTable(dec, tk, handle)
				if err != nil {
					return err
				}
			}
		default:
			// ignore any other tag
		}
	}
}

func processTable(dec *xml.Decoder, tableStart xml.StartElement, handle TagFunc) error {
	var table TableInfo

	for _, a := range tableStart.Attr {
		if a.Name.Local == "name" {
			table.Name = a.Value
			break
		}
	}

	for {
		tk, err := dec.RawToken()
		if err != nil {
			return err
		}
		switch tk := tk.(type) {
		case xml.StartElement:
			if tk.Name.Local == "tag" {
				err = processTag(dec, table, tk, handle)
				if err != nil {
					return err
				}
			}
		case xml.EndElement:
			if tk.Name.Local == "table" {
				return nil
			}
		default:
		}
	}
}

func processTag(dec *xml.Decoder, table TableInfo, tagStart xml.StartElement, handle TagFunc) error {
	var tag Tag
	tag.Table = table
	for _, a := range tagStart.Attr {
		switch a.Name.Local {
		case "name":
			tag.Name = a.Value
		case "type":
			tag.Type = a.Value
		case "writable":
			tag.Writable = a.Value != "false"
		}
	}
	for {
		tk, err := dec.RawToken()
		if err != nil {
			return err
		}
		switch tk := tk.(type) {
		case xml.StartElement:
			if tk.Name.Local == "desc" {
				err = processDescription(dec, tk, &tag)
			}
		case xml.EndElement:
			if tk.Name.Local == "tag" {
				return handle(tag)
			}
		default:
		}
	}
	return fmt.Errorf("mal-formed XML missing table tag at offset %v", dec.InputOffset())
}

func processDescription(dec *xml.Decoder, description xml.StartElement, tag *Tag) error {
	var desc Description
	for _, a := range description.Attr {
		if a.Name.Local == "lang" {
			desc.Lang = a.Value
		}
	}
	for {
		tk, err := dec.RawToken()
		if err != nil {
			return err
		}
		switch tk := tk.(type) {
		case xml.CharData:
			desc.Text = string(tk)
		case xml.EndElement:
			if tk.Name.Local == "desc" {
				tag.Description = append(tag.Description, desc)
				return nil
			}
		default:
		}
	}
}
