package parser

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type pdfRun struct {
	x, y float64
	text string
}

// buildPDF writes a one page PDF with the given glyph runs placed by absolute text
// matrices, so a test controls the coordinates the parser reads back.
func buildPDF(t *testing.T, size float64, runs []pdfRun) string {
	t.Helper()

	var content strings.Builder
	fmt.Fprintf(&content, "BT\n/F1 %v Tf\n", size)
	for _, r := range runs {
		if r.text != "" {
			fmt.Fprintf(&content, "1 0 0 1 %v %v Tm\n(%s) Tj\n", r.x, r.y, strings.NewReplacer(`\`, `\\`, `(`, `\(`, `)`, `\)`).Replace(r.text))
		}
	}
	content.WriteString("ET\n")

	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(content.String()), content.String()),
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	}

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int64, 0, len(objects))
	for i, obj := range objects {
		offsets = append(offsets, int64(buf.Len()))
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	startXRef := int64(buf.Len())
	fmt.Fprintf(&buf, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for _, off := range offsets {
		fmt.Fprintf(&buf, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, startXRef)

	return writeTemp(t, "fixture.pdf", buf.String())
}

func writeTemp(t *testing.T, name, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestPDFParserRebuildsParagraphs(t *testing.T) {
	path := buildPDF(t, 12, []pdfRun{
		{72, 780, "1. Total price is 100 CNY."},
		{72, 768, "Payment is monthly."},
		{110, 756, "The buyer accepts delivery."},
		{72, 740, "Warranty is twelve months."},
		{72, 700, "Article 3 Other."},
		{72, 688, "Split"},
		{130, 688, "words"},
	})

	doc, err := Open(path)
	if err != nil {
		t.Fatalf("parse pdf: %v", err)
	}

	want := []string{
		"1. Total price is 100 CNY. Payment is monthly.",
		"The buyer accepts delivery. Warranty is twelve months.",
		"Article 3 Other. Split words",
	}
	if got := doc.Paragraphs; !reflect.DeepEqual(got, want) {
		t.Fatalf("paragraphs mismatch:\ngot  %#v\nwant %#v", got, want)
	}
	if len(doc.Tables) != 0 {
		t.Fatalf("PDF has no table structure, got %d tables", len(doc.Tables))
	}
	if doc.Path != path {
		t.Fatalf("path not recorded: %q", doc.Path)
	}
}

func TestPDFParserIsDeterministicWhenMarginsTie(t *testing.T) {
	path := buildPDF(t, 12, []pdfRun{
		{72, 780, "Alpha one."},
		{110, 768, "Beta two."},
		{72, 756, "Gamma three."},
		{110, 744, "Delta four."},
		{72, 732, "Epsilon five."},
		{110, 720, "Zeta six."},
	})

	var first []string
	for i := 0; i < 20; i++ {
		doc, err := Open(path)
		if err != nil {
			t.Fatalf("parse pdf: %v", err)
		}
		if first == nil {
			first = doc.Paragraphs
			continue
		}
		if !reflect.DeepEqual(doc.Paragraphs, first) {
			t.Fatalf("run %d differs:\n got %#v\nfirst %#v", i, doc.Paragraphs, first)
		}
	}
}

func TestPDFParserRejectsTextlessPage(t *testing.T) {
	path := buildPDF(t, 12, []pdfRun{{72, 780, ""}})

	if _, err := Open(path); err == nil || !strings.Contains(err.Error(), "扫描件") {
		t.Fatalf("textless pdf should report a scan, got %v", err)
	}
}

func TestPDFParserRejectsBrokenFile(t *testing.T) {
	path := writeTemp(t, "broken.pdf", "this is not a pdf at all")

	if _, err := Open(path); err == nil || !strings.Contains(err.Error(), "读取 PDF 失败") {
		t.Fatalf("broken pdf should report a read failure, got %v", err)
	}
}
