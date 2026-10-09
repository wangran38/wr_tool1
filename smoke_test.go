package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wr_tool/internal/cleaner"
	"wr_tool/internal/differ"
	"wr_tool/internal/exporter"
	"wr_tool/internal/license"
)

func writeDocx(t *testing.T, path, body string) {
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	w, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}
	zw.Close()
}

func TestSmokePipeline(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())

	dir := t.TempDir()
	oldPath := filepath.Join(dir, "old.docx")
	newPath := filepath.Join(dir, "new.docx")

	writeDocx(t, oldPath, `<?xml version="1.0"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`+
		`<w:p><w:r><w:t>第一条 合同总价为100万元。</w:t></w:r></w:p>`+
		`<w:p><w:r><w:t>第二条 付款方式为月结。</w:t></w:r></w:p>`+
		`<w:tbl><w:tr><w:tc><w:p><w:r><w:t>标的</w:t></w:r></w:p></w:tc><w:tc><w:p><w:r><w:t>单价</w:t></w:r></w:p></w:tc></w:tr>`+
		`<w:tr><w:tc><w:p><w:r><w:t>设备A</w:t></w:r></w:p></w:tc><w:tc><w:p><w:r><w:t>1200</w:t></w:r></w:p></w:tc></w:tr></w:tbl>`+
		`</w:body></w:document>`)

	writeDocx(t, newPath, `<?xml version="1.0"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`+
		`<w:p><w:r><w:t>第一条　合同总价为 100 万元。</w:t></w:r></w:p>`+
		`<w:p><w:r><w:t>第二条 付款方式为季结。</w:t></w:r></w:p>`+
		`<w:p><w:r><w:t>第三条 争议提交深圳仲裁。</w:t></w:r></w:p>`+
		`<w:tbl><w:tr><w:tc><w:p><w:r><w:t>标的</w:t></w:r></w:p></w:tc><w:tc><w:p><w:r><w:t>单价</w:t></w:r></w:p></w:tc></w:tr>`+
		`<w:tr><w:tc><w:p><w:r><w:t>设备A</w:t></w:r></w:p></w:tc><w:tc><w:p><w:r><w:t>1350</w:t></w:r></w:p></w:tc></w:tr></w:tbl>`+
		`</w:body></w:document>`)

	a := NewApp()
	opts := cleaner.Options{IgnoreWhitespace: true, IgnoreNumbering: true}
	res, err := a.CompareFiles(oldPath, newPath, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Tables) == 0 {
		t.Fatal("no tables parsed")
	}

	var kinds []string
	for _, p := range res.Text.Pairs {
		kinds = append(kinds, string(p.Kind))
	}
	t.Log("text pairs:", strings.Join(kinds, ","), "stats:", res.Text.Stats)
	t.Log("table rows:", res.Tables[0].Rows)

	replaces := 0
	for _, p := range res.Text.Pairs {
		if p.Kind != differ.Replace {
			continue
		}
		replaces++
		if joinSegs(p.OldSegs) != p.Old || joinSegs(p.NewSegs) != p.New {
			t.Fatalf("inline segments do not reconstruct the line: %+v", p)
		}
		if !hasKind(p.OldSegs, differ.Delete) || !hasKind(p.NewSegs, differ.Insert) {
			t.Fatalf("replace pair has no character-level change: %+v", p)
		}
		t.Logf("inline old=%v new=%v", p.OldSegs, p.NewSegs)
	}
	if replaces == 0 {
		t.Fatal("expected a replaced paragraph")
	}

	out := filepath.Join(dir, "report.html")
	if err := (exporter.Exporter{}).ToHTML(exporter.Report{
		OldFile: "old.docx", NewFile: "new.docx", Text: res.Text, Tables: res.Tables, Generated: "now",
	}, out); err != nil {
		t.Fatal(err)
	}
	html, _ := os.ReadFile(out)
	for _, want := range []string{`<span class="seg-ins">季</span>`, `<span class="seg-del">月</span>`, "深圳仲裁"} {
		if !strings.Contains(string(html), want) {
			t.Fatalf("report missing %q:\n%s", want, html)
		}
	}

	pdfPath := filepath.Join(dir, "report.pdf")
	if err := (exporter.Exporter{}).ToPDF(exporter.Report{
		OldFile: "old.docx", NewFile: "new.docx", Text: res.Text, Tables: res.Tables, Generated: "now",
	}, pdfPath); err != nil {
		t.Fatalf("pdf export: %v", err)
	}
	pdf, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(pdf), "%PDF-") {
		t.Fatalf("not a pdf (%d bytes)", len(pdf))
	}
	t.Logf("pdf ok: %d bytes -> %s", len(pdf), pdfPath)

	code, err := license.MachineCode()
	if err != nil {
		t.Fatal(err)
	}
	if n := len(strings.ReplaceAll(code, "-", "")); n != 16 {
		t.Fatalf("machine code should be 16 characters, got %d (%s)", n, code)
	}
	key := license.Issue(code)
	if !license.Verify(code, key) {
		t.Fatal("license verification failed")
	}
	if license.Verify(code, key[:20]+"XXXXX") {
		t.Fatal("license verification accepted a tampered code")
	}
	t.Log("machine code:", code, "key:", key)
}

func joinSegs(segs []differ.Segment) string {
	var b strings.Builder
	for _, s := range segs {
		b.WriteString(s.Text)
	}
	return b.String()
}

func hasKind(segs []differ.Segment, kind differ.Kind) bool {
	for _, s := range segs {
		if s.Kind == kind {
			return true
		}
	}
	return false
}
