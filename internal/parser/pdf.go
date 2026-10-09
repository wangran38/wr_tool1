package parser

import "errors"

type PDFParser struct{}

// Placeholder: PDF text/table extraction needs a third-party dependency
// (e.g. github.com/ledongthuc/pdf for text, or an external pdfium binding).
var ErrNotImplemented = errors.New("pdf parser not implemented yet")

func (PDFParser) Parse(path string) (*Document, error) {
	return nil, ErrNotImplemented
}
