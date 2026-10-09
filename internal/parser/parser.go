package parser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Table struct {
	Rows [][]string `json:"rows"`
}

func (t Table) ColCount() int {
	max := 0
	for _, r := range t.Rows {
		if len(r) > max {
			max = len(r)
		}
	}
	return max
}

type Document struct {
	Path       string   `json:"path"`
	Paragraphs []string `json:"paragraphs"`
	Tables     []Table  `json:"tables"`
}

type Parser interface {
	Parse(path string) (*Document, error)
}

func Open(path string) (*Document, error) {
	var p Parser
	switch strings.ToLower(filepath.Ext(path)) {
	case ".docx":
		p = &DocxParser{}
	case ".pdf":
		p = &PDFParser{}
	case ".txt":
		p = &PlainTextParser{}
	default:
		return nil, fmt.Errorf("unsupported file type: %s", path)
	}

	doc, err := p.Parse(path)
	if err != nil {
		return nil, err
	}
	doc.Path = path
	return doc, nil
}

type PlainTextParser struct{}

func (PlainTextParser) Parse(path string) (*Document, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}
	return &Document{Paragraphs: lines}, nil
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}
