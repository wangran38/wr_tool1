package parser

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ledongthuc/pdf"
)

// PDFParser rebuilds text from glyph positions: a PDF stores no paragraph or table
// structure, so lines are grouped from spans and merged back into paragraphs.
// Document.Tables always stays empty for PDF input; table rows arrive as plain lines.
type PDFParser struct{}

// ErrPDFNoText covers scans and image-only PDFs, which carry no text layer at all.
var ErrPDFNoText = errors.New("PDF 中没有可提取的文字，可能是扫描件或图片型 PDF，请改用 docx/txt 或先做 OCR")

const (
	spaceGapFactor     = 0.25 // horizontal gap, in font sizes, that counts as a word space
	paragraphGapFactor = 1.6  // vertical gap, in font sizes, that counts as a new paragraph
	indentFactor       = 1.2  // line start this far right of the body margin counts as a first line
)

// A clause heading always begins a new paragraph, whatever the spacing says.
var clauseHeading = regexp.MustCompile(`^(?:第[0-9一二三四五六七八九十百千]+[条章節节款项部分]|[（(][0-9一二三四五六七八九十]+[)）]|[0-9]+[.．、)）]|[一二三四五六七八九十]+、)`)

type pdfLine struct {
	y    float64
	x    float64
	size float64
	text string
}

func (PDFParser) Parse(path string) (*Document, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return nil, fmt.Errorf("读取 PDF 失败（文件可能已加密或损坏）: %v", err)
	}
	defer f.Close()

	var paragraphs []string
	for num := 1; num <= r.NumPage(); num++ {
		paragraphs = append(paragraphs, mergeParagraphs(groupLines(readSpans(r.Page(num))))...)
	}
	if len(paragraphs) == 0 {
		return nil, ErrPDFNoText
	}
	return &Document{Paragraphs: paragraphs}, nil
}

// readSpans swallows the library's panics: one malformed content stream must not
// take down the whole desktop app.
func readSpans(p pdf.Page) (spans []pdf.Text) {
	defer func() {
		if recover() != nil {
			spans = nil
		}
	}()
	return p.Content().Text
}

func groupLines(spans []pdf.Text) []pdfLine {
	if len(spans) == 0 {
		return nil
	}

	topFirst := make([]pdf.Text, len(spans))
	copy(topFirst, spans)
	sort.SliceStable(topFirst, func(i, j int) bool { return topFirst[i].Y > topFirst[j].Y })

	var lines []pdfLine
	for i := 0; i < len(topFirst); {
		size := topFirst[i].FontSize
		if size <= 0 {
			size = 10
		}
		tolerance := math.Max(2, size*0.4)
		j := i + 1
		for j < len(topFirst) && math.Abs(topFirst[j].Y-topFirst[i].Y) <= tolerance {
			j++
		}
		run := topFirst[i:j]
		sort.SliceStable(run, func(a, b int) bool { return run[a].X < run[b].X })
		lines = append(lines, pdfLine{y: run[0].Y, x: run[0].X, size: size, text: joinRun(run, size)})
		i = j
	}
	return lines
}

func joinRun(spans []pdf.Text, size float64) string {
	var b strings.Builder
	b.WriteString(spans[0].S)
	end := spans[0].X + spans[0].W

	for _, s := range spans[1:] {
		if s.X-end > size*spaceGapFactor && betweenWords(b.String(), s.S) {
			b.WriteString(" ")
		}
		b.WriteString(s.S)
		end = s.X + s.W
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// betweenWords reports whether a wide gap really splits two Latin words. Letterspaced
// digits and dashes are laid out with the same gaps, and a space there invents a change.
func betweenWords(prev, next string) bool {
	a, okA := lastRune(prev)
	c, okC := firstRune(next)
	if !okA || !okC {
		return false
	}
	return unicode.IsLetter(a) && unicode.IsLetter(c) && !isHan(a) && !isHan(c)
}

func mergeParagraphs(lines []pdfLine) []string {
	margin := bodyMargin(lines)

	var out []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}

	var prev pdfLine
	for _, l := range lines {
		if l.text == "" {
			continue
		}
		if cur.Len() > 0 && startsParagraph(l, prev, margin) {
			flush()
		}
		if cur.Len() > 0 {
			appendContinuation(&cur, l.text)
		} else {
			cur.WriteString(l.text)
		}
		prev = l
	}
	flush()
	return out
}

func startsParagraph(l, prev pdfLine, margin float64) bool {
	switch {
	case clauseHeading.MatchString(l.text):
		return true
	case l.x > margin+l.size*indentFactor:
		return true
	case prev.y-l.y > math.Max(prev.size, l.size)*paragraphGapFactor:
		return true
	default:
		return false
	}
}

// bodyMargin is the most common line start, that is the flush-left edge of the body text.
// Ties break left so two parses of the same file cannot disagree about indentation.
func bodyMargin(lines []pdfLine) float64 {
	counts := make(map[int64]int)
	for _, l := range lines {
		if l.text == "" {
			continue
		}
		counts[int64(math.Round(l.x))]++
	}
	best, bestCount := 0.0, -1
	for bucket, count := range counts {
		x := float64(bucket)
		if count > bestCount || (count == bestCount && x < best) {
			best, bestCount = x, count
		}
	}
	return best
}

// A wrapped line of a Chinese paragraph continues without a space; a Latin one needs it back.
func appendContinuation(b *strings.Builder, next string) {
	if last, ok := lastRune(b.String()); ok {
		if head, okC := firstRune(next); okC && (isHan(last) || isHan(head)) {
			b.WriteString(next)
			return
		}
	}
	b.WriteString(" ")
	b.WriteString(next)
}

func lastRune(s string) (rune, bool) {
	for len(s) > 0 {
		r, size := utf8.DecodeLastRuneInString(s)
		if !unicode.IsSpace(r) {
			return r, true
		}
		s = s[:len(s)-size]
	}
	return 0, false
}

func firstRune(s string) (rune, bool) {
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return r, true
		}
	}
	return 0, false
}

func isHan(r rune) bool {
	return unicode.In(r, unicode.Han)
}
