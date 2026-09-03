package fsutil

import (
	"os"

	"github.com/sabhiram/go-gitignore"
)

type Matcher struct {
	ignore *ignore.GitIgnore
}

func NewMatcher(root string) (*Matcher, error) {
	data, err := os.ReadFile(root + "/.gitignore")
	if err != nil {
		return nil, err
	}

	matcher := ignore.CompileIgnoreLines(string(data))

	return &Matcher{
		ignore: matcher,
	}, nil
}

func (m *Matcher) Match(path string) bool {
	return m.ignore.MatchesPath(path)
}
