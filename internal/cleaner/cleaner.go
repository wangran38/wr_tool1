package cleaner

import (
	"regexp"
	"strings"
	"unicode"

	"wr_tool/internal/parser"
)

type Options struct {
	IgnoreWhitespace  bool `json:"ignoreWhitespace"`
	IgnoreCase        bool `json:"ignoreCase"`
	IgnorePunctuation bool `json:"ignorePunctuation"`
	IgnoreNumbering   bool `json:"ignoreNumbering"`
	IgnoreLineBreaks  bool `json:"ignoreLineBreaks"`
}

// Unit keeps the display text separate from the comparison key so the viewer can
// show the original formatting while the differ ignores it.
type Unit struct {
	Raw string `json:"raw"`
	Key string `json:"key"`
}

// A numeric clause marker must carry a separator ("1." / "1.2" / "1）"), otherwise
// plain amounts such as "1200" inside table cells would be stripped as numbering.
var numberingPrefix = regexp.MustCompile(`^\s*(?:第[一二三四五六七八九十百千0-9]+[条章節节部分项]|\d+(?:\.\d+|[.．、)）:：])+[.．、)）:：]?|（[一二三四五六七八九十]+）)\s*`)

func Prepare(doc *parser.Document, o Options) []Unit {
	units := make([]Unit, 0, len(doc.Paragraphs))
	for _, p := range doc.Paragraphs {
		if strings.TrimSpace(p) == "" {
			continue
		}
		units = append(units, Unit{Raw: p, Key: Normalize(p, o)})
	}
	return units
}

func Normalize(text string, o Options) string {
	s := text
	if o.IgnoreNumbering {
		s = numberingPrefix.ReplaceAllString(s, "")
	}
	if o.IgnoreWhitespace {
		s = strings.Join(strings.FieldsFunc(s, func(r rune) bool {
			return unicode.IsSpace(r) || r == '　'
		}), "")
	}
	if o.IgnoreLineBreaks {
		s = strings.ReplaceAll(s, "\n", "")
	}
	if o.IgnorePunctuation {
		s = strings.Map(func(r rune) rune {
			if unicode.IsPunct(r) || r == '、' || r == '。' || r == '，' {
				return -1
			}
			return r
		}, s)
	}
	if o.IgnoreCase {
		s = strings.ToLower(s)
	}
	return s
}
