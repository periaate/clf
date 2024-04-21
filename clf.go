package clf

import (
	"fmt"
)

type Flag struct {
	Values []string

	Toggle bool // Same as Exactly = -1

	Present int // number of times the flag was present

	Exactly int // exact number of values needed for yield, 0 is undefined, -1 is flag only
	AtLeast int
	AtMost  int

	Handler func([]string)

	Keys        []string
	Name        string
	Description string

	Group string
}

type Options struct {
	Names  map[string]*Flag
	Flags  map[string]*Flag
	Rest   []string
	Errors bool
}

// Get returns a deferenced flag value
func (o *Options) Get(key string) (f Flag) {
	if v, ok := o.Flags[key]; ok {
		val := *v
		return val
	}
	return
}

func ParseNames(flags []*Flag) (opts *Options, err error) {
	opts = &Options{
		Names: map[string]*Flag{},
		Flags: map[string]*Flag{},
		Rest:  make([]string, 0),
	}
	for _, flag := range flags {
		if flag.Toggle {
			flag.Exactly = -1
		}
		if flag.Name == "" {
			if len(flag.Keys) == 0 {
				err = fmt.Errorf("flag must have a name or key")
				return
			}
			flag.Name = flag.Keys[0]
		}
		if v, ok := opts.Flags[flag.Name]; ok {
			err = fmt.Errorf("names must be unique! name %s used by both %v and %v", flag.Name, flag, v)
			return
		}
		opts.Flags[flag.Name] = flag

		if len(flag.Keys) == 0 {
			opts.Names[flag.Name] = flag
			continue
		}

		for _, key := range flag.Keys {
			if v, ok := opts.Names[key]; ok {
				err = fmt.Errorf("keys must be unique! %s collides with %s with key %s", flag.Name, v.Name, key)
				return
			}
			opts.Names[key] = flag
		}
	}
	return
}

func Parse(args []string, flags []*Flag) (opts *Options, err error) {
	opts, err = ParseNames(flags)
	var cur *Flag
	var i int
	for _, arg := range args {
		if cur != nil && i > 1 {
			if cur.AtMost < i {
				cur.Handler(cur.Values)
				cur.Values = nil
				cur = nil
				i = 0
			}
		}

		f, ok := opts.Names[arg]
		if !ok {
			if cur == nil || cur.Exactly == -1 {
				opts.Rest = append(opts.Rest, arg)
			} else {
				i++
				cur.Values = append(cur.Values, arg)
			}
			continue
		}
		if cur != nil {
			cur.Handler(cur.Values)
			cur.Values = nil
		}
		i = 0
		f.Present++

		cur = f
	}
	if cur != nil {
		if cur.Handler != nil {
			cur.Handler(cur.Values)
		}
	}

	return
}
