package parser

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

type DocxParser struct{}

const documentPart = "word/document.xml"

func (DocxParser) Parse(path string) (*Document, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open docx: %w", err)
	}
	defer zr.Close()

	var part *zip.File
	for _, f := range zr.File {
		if f.Name == documentPart {
			part = f
			break
		}
	}
	if part == nil {
		return nil, fmt.Errorf("%s not found, not a valid docx", documentPart)
	}

	rc, err := part.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return decodeBody(rc)
}

// The body is a flat sequence of block elements; tables can nest paragraphs, so
// elements are dispatched on their local name rather than traversed structurally.
func decodeBody(r io.Reader) (*Document, error) {
	dec := xml.NewDecoder(r)
	doc := &Document{}

	var (
		inCell    bool
		inParaRun bool
		cell      []string
		row       []string
		table     [][]string
		para      strings.Builder
	)

	flushPara := func() {
		text := strings.TrimSpace(para.String())
		para.Reset()
		if text == "" {
			return
		}
		if inCell {
			cell = append(cell, text)
		} else {
			doc.Paragraphs = append(doc.Paragraphs, text)
		}
	}

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decode document.xml: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "tbl":
				table = nil
			case "tr":
				row = nil
			case "tc":
				inCell = true
				cell = nil
			case "p":
				inParaRun = true
			case "tab":
				para.WriteString("\t")
			case "br":
				para.WriteString("\n")
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "p":
				flushPara()
				inParaRun = false
			case "tc":
				row = append(row, strings.Join(cell, " "))
				cell = nil
				inCell = false
			case "tr":
				table = append(table, row)
				row = nil
			case "tbl":
				if len(table) > 0 {
					doc.Tables = append(doc.Tables, Table{Rows: table})
				}
				table = nil
			}
		case xml.CharData:
			if inParaRun {
				para.Write([]byte(t))
			}
		}
	}

	return doc, nil
}
