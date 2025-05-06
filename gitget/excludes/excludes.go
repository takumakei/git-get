package excludes

import (
	"fmt"

	"github.com/bmatcuk/doublestar/v4"
)

type Matcher struct {
	patterns []string
}

func New(patterns []string) (*Matcher, error) {
	for _, p := range patterns {
		if !doublestar.ValidatePattern(p) {
			return nil, fmt.Errorf("%q, %w", p, doublestar.ErrBadPattern)
		}
	}
	return &Matcher{patterns: patterns}, nil
}

func (m *Matcher) PathMatch(path string) bool {
	for _, p := range m.patterns {
		if doublestar.PathMatchUnvalidated(p, path) {
			return true
		}
	}
	return false
}
