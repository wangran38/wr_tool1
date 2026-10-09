package differ

import (
	"unicode/utf8"

	"github.com/sergi/go-diff/diffmatchpatch"

	"wr_tool/internal/cleaner"
)

var dmp = diffmatchpatch.New()

type Kind string

const (
	Equal   Kind = "equal"
	Insert  Kind = "insert"
	Delete  Kind = "delete"
	Replace Kind = "replace"
)

// Segment is a character-level span inside a changed line, used for inline highlighting.
type Segment struct {
	Kind Kind   `json:"kind"`
	Text string `json:"text"`
}

type Pair struct {
	Kind    Kind      `json:"kind"`
	Old     string    `json:"old"`
	New     string    `json:"new"`
	OldSegs []Segment `json:"oldSegs,omitempty"`
	NewSegs []Segment `json:"newSegs,omitempty"`
}

type Stats struct {
	Equal      int     `json:"equal"`
	Insert     int     `json:"insert"`
	Delete     int     `json:"delete"`
	Replace    int     `json:"replace"`
	Similarity float64 `json:"similarity"`
}

type Result struct {
	Pairs []Pair `json:"pairs"`
	Stats Stats  `json:"stats"`
}

// Diff aligns two cleaned documents with the Myers algorithm. Units are matched on their
// cleaned Key but reported with their original Raw text.
func Diff(oldUnits, newUnits []cleaner.Unit) Result {
	oldTokens, newTokens := tokenize(oldUnits, newUnits)
	diffs := dmp.DiffMainRunes(oldTokens, newTokens, false)

	pairs := toPairs(diffs, oldUnits, newUnits)
	for i := range pairs {
		if pairs[i].Kind == Replace {
			pairs[i].OldSegs, pairs[i].NewSegs = inlineSegments(pairs[i].Old, pairs[i].New)
		}
	}

	return Result{Pairs: pairs, Stats: summarize(pairs, len(oldUnits), len(newUnits))}
}

// tokenize maps each distinct cleaned key to one rune so the Myers diff runs on a token
// stream and every returned span decodes back to exact unit positions.
func tokenize(oldUnits, newUnits []cleaner.Unit) ([]rune, []rune) {
	ids := make(map[string]rune, len(oldUnits)+len(newUnits))
	next := rune(0x1000)

	idOf := func(key string) rune {
		if r, ok := ids[key]; ok {
			return r
		}
		// Surrogates would be mangled when the token string is re-encoded.
		if next >= 0xD800 && next <= 0xDFFF {
			next = 0xE000
		}
		ids[key] = next
		next++
		return next - 1
	}

	oldTokens := make([]rune, 0, len(oldUnits))
	for _, u := range oldUnits {
		oldTokens = append(oldTokens, idOf(u.Key))
	}
	newTokens := make([]rune, 0, len(newUnits))
	for _, u := range newUnits {
		newTokens = append(newTokens, idOf(u.Key))
	}
	return oldTokens, newTokens
}

func toPairs(diffs []diffmatchpatch.Diff, oldUnits, newUnits []cleaner.Unit) []Pair {
	pairs := make([]Pair, 0, len(oldUnits)+len(newUnits))
	var dels, inss []int
	oi, ni := 0, 0

	flush := func() {
		replaces := len(dels)
		if len(inss) < replaces {
			replaces = len(inss)
		}
		for k := 0; k < replaces; k++ {
			pairs = append(pairs, Pair{
				Kind: Replace,
				Old:  oldUnits[dels[k]].Raw,
				New:  newUnits[inss[k]].Raw,
			})
		}
		for k := replaces; k < len(dels); k++ {
			pairs = append(pairs, Pair{Kind: Delete, Old: oldUnits[dels[k]].Raw})
		}
		for k := replaces; k < len(inss); k++ {
			pairs = append(pairs, Pair{Kind: Insert, New: newUnits[inss[k]].Raw})
		}
		dels, inss = dels[:0], inss[:0]
	}

	for _, d := range diffs {
		count := utf8.RuneCountInString(d.Text)
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			flush()
			for k := 0; k < count; k++ {
				pairs = append(pairs, Pair{Kind: Equal, Old: oldUnits[oi+k].Raw, New: newUnits[ni+k].Raw})
			}
			oi += count
			ni += count
		case diffmatchpatch.DiffDelete:
			for k := 0; k < count; k++ {
				dels = append(dels, oi+k)
			}
			oi += count
		case diffmatchpatch.DiffInsert:
			for k := 0; k < count; k++ {
				inss = append(inss, ni+k)
			}
			ni += count
		}
	}
	flush()

	return pairs
}

func inlineSegments(oldText, newText string) ([]Segment, []Segment) {
	diffs := dmp.DiffCleanupSemantic(dmp.DiffMain(oldText, newText, false))

	var oldSegs, newSegs []Segment
	for _, d := range diffs {
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			oldSegs = append(oldSegs, Segment{Kind: Equal, Text: d.Text})
			newSegs = append(newSegs, Segment{Kind: Equal, Text: d.Text})
		case diffmatchpatch.DiffDelete:
			oldSegs = append(oldSegs, Segment{Kind: Delete, Text: d.Text})
		case diffmatchpatch.DiffInsert:
			newSegs = append(newSegs, Segment{Kind: Insert, Text: d.Text})
		}
	}
	return oldSegs, newSegs
}

func summarize(pairs []Pair, oldCount, newCount int) Stats {
	s := Stats{}
	for _, p := range pairs {
		switch p.Kind {
		case Equal:
			s.Equal++
		case Insert:
			s.Insert++
		case Delete:
			s.Delete++
		case Replace:
			s.Replace++
		}
	}
	total := oldCount
	if newCount > total {
		total = newCount
	}
	if total > 0 {
		s.Similarity = float64(s.Equal) / float64(total)
	}
	return s
}
