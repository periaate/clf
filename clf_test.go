package clf

import (
	"testing"
)

var help = Flag{
	Toggle:      true,
	Keys:        []string{"help", "-h", "--help"},
	Name:        "help",
	Description: "Prints this help message",
}

var include = Flag{
	Keys:        []string{"-i", "--include"},
	Name:        "include",
	Description: "Flag will add all values to the include list, until another flag is found",
}

var file = Flag{
	Keys: []string{"f", "-f"},
	Name: "file",
}

var recurse = Flag{
	Toggle: true,
	Keys:   []string{"r", "-r"},
	Name:   "recurse",
}

type Expect struct {
	Case     string
	Name     string
	Flag     func() []*Flag
	Args     []string
	Values   []string
	Present  int
	ValCount int
}

func WrapFlag(f ...Flag) func() []*Flag {
	return func() (fl []*Flag) {
		for _, v := range f {
			otherFlag := v
			fl = append(fl, &otherFlag)
		}
		return fl
	}
}

func TestParse(t *testing.T) {
	cases := []Expect{
		{
			Case:    "help flag",
			Name:    "help",
			Flag:    WrapFlag(help, include, file),
			Args:    []string{"--help", "help", "-h"},
			Present: 3,
		},
		{
			Case:     "eager flag",
			Name:     "include",
			Flag:     WrapFlag(include, file),
			Args:     []string{"--include", "foo", "bar", "-f", "baz"},
			Present:  1,
			ValCount: 2,
		},
		{
			Case:     "test",
			Name:     "file",
			Flag:     WrapFlag(recurse, file),
			Args:     []string{"./path/to/something/", "r", "-f", "search", "term"},
			Present:  1,
			ValCount: 2,
		},
	}
	for _, c := range cases {
		t.Run(c.Case, func(t *testing.T) {
			opts, err := Parse(c.Args, c.Flag())
			if err != nil {
				t.Error(err)
			}
			g := opts.Get(c.Name)
			p := g.Present
			cn := len(g.Values)
			if cn != c.ValCount {
				t.Errorf("expected %d, got %d", c.ValCount, cn)
			}
			if p != c.Present {
				t.Errorf("expected %d, got %d", c.Present, p)
			}
		})
	}
}

func TestYield(t *testing.T)
