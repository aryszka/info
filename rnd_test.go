package keyval

import (
	"math/rand"
	"strings"
)

const (
	defaultChars             = "abcdefghijklmnopqrstuvwxyz"
	defaultMinStrLength      = 0
	defaultMaxStrLength      = 48
	defaultMinSectionLength  = 0
	defaultMaxSectionLength  = 6
	defaultMinKeyLength      = 0
	defaultMaxKeyLength      = 12
	defaultEntriesPerSection = 9
)

type genOptions struct {
	commentChars      string
	keyChars          string
	valChars          string
	minStrLength      int
	maxStrLength      int
	minSectionLength  int
	maxSectionLength  int
	minKeyLength      int
	maxKeyLength      int
	entriesPerSection int
	seed              int64
}

type gen struct {
	*genOptions
	rnd            *rand.Rand
	sections       []string
	sectionEntries map[string]int
}

func applyDefaults(o *genOptions) {
	if o.commentChars == "" {
		o.commentChars = defaultChars
	}

	if o.keyChars == "" {
		o.keyChars = defaultChars
	}

	if o.valChars == "" {
		o.valChars = defaultChars
	}

	if o.minStrLength == 0 {
		o.minStrLength = defaultMinStrLength
	}

	if o.maxStrLength == 0 {
		o.maxStrLength = defaultMaxStrLength
	}

	if o.minSectionLength == 0 {
		o.minSectionLength = defaultMinSectionLength
	}

	if o.maxSectionLength == 0 {
		o.maxSectionLength = defaultMaxSectionLength
	}

	if o.minKeyLength == 0 {
		o.minKeyLength = defaultMinKeyLength
	}

	if o.maxKeyLength == 0 {
		o.maxKeyLength = defaultMaxKeyLength
	}

	if o.entriesPerSection == 0 {
		o.entriesPerSection = defaultEntriesPerSection
	}
}

func newGen(o genOptions) *gen {
	applyDefaults(&o)
	return &gen{&o, rand.New(rand.NewSource(o.seed)), nil, make(map[string]int)}
}

func (g *gen) between(min, max int) int {
	return min + g.rnd.Intn(max-min)
}

func (g *gen) char(chars string) byte {
	return []byte(chars)[g.rnd.Intn(len(chars))]
}

func (g *gen) str(chars string) string {
	l := g.between(g.minStrLength, g.maxStrLength)
	name := make([]byte, l)
	for i := 0; i < l; i++ {
		name[i] = g.char(chars)
	}

	return string(name)
}

func (g *gen) strs(min, max int, chars string) []string {
	l := g.between(min, max)
	strs := make([]string, l)
	for i := 0; i < l; i++ {
		strs[i] = g.str(chars)
	}

	return strs
}

func (g *gen) next(repeatSectionAfter int) *Entry {
	var section []string
	if len(g.sections) > 0 && g.entriesPerSection > 0 {
		rsa := g.rnd.Intn(repeatSectionAfter)
		if rsa == 0 {
			i := g.rnd.Intn(len(g.sections))
			sectionStr := g.sections[i]
			section = strings.Split(sectionStr, ".")
			g.sectionEntries[sectionStr]++
			if g.rnd.Intn(g.entriesPerSection) == g.sectionEntries[sectionStr] {
				delete(g.sectionEntries, sectionStr)
				g.sections = append(g.sections[:i], g.sections[i+1:]...)
			}
		}
	}

	if section == nil {
		section = g.strs(g.minSectionLength, g.maxSectionLength, g.keyChars)
		if g.entriesPerSection > 0 {
			sectionStr := strings.Join(section, ".")
			g.sections = append(g.sections, sectionStr)
			g.sectionEntries[sectionStr] = 1
		}
	}

	comment := g.str(g.commentChars)
	key := g.strs(g.minKeyLength, g.maxKeyLength, g.keyChars)
	val := g.str(g.valChars)

	return &Entry{
		Comment: comment,
		Key:     append(section, key...),
		Val:     val}
}

func (g *gen) n(n int) []*Entry {
	es := make([]*Entry, n)
	rsa := n / g.entriesPerSection
	for i := 0; i < n; i++ {
		es[i] = g.next(rsa)
	}

	return es
}
