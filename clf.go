package clf

import (
	"fmt"
)

type Flag struct {
	Values  []string
	Greedy  bool
	Toggle  bool
	Present int // number of times the flag was present
	Exactly int // exact number of values needed for yield, 0 == undefined
	AtLeast int
	AtMost  int

	Keys        []string
	Name        string
	Description string
}

type Options struct {
	Flags map[string]*Flag
	Rest  []string
}

// Get returns a deferenced flag value
func (o *Options) Get(key string) (f Flag) {
	if v, ok := o.Flags[key]; ok {
		val := *v
		return val
	}
	return
}

func Parse(args []string, flags []*Flag) (opts *Options, err error) {
	opts = &Options{
		Flags: map[string]*Flag{},
		Rest:  make([]string, 0),
	}
	lookup := map[string]*Flag{}
	for _, flag := range flags {
		if v, ok := opts.Flags[flag.Name]; ok {
			err = fmt.Errorf("names must be unique! name %s used by both %v and %v", flag.Name, flag, v)
			return
		}
		opts.Flags[flag.Name] = flag
		for _, key := range flag.Keys {
			if v, ok := lookup[key]; ok {
				err = fmt.Errorf("keys must be unique! %s collides with %s with key %s", flag.Name, v.Name, key)
				return
			}
			lookup[key] = flag
		}
	}

	var cur *Flag
	var i int
	for _, arg := range args {
		if cur != nil && i > 1 {
			if !cur.Greedy {
				cur = nil
			}
		}

		f, ok := lookup[arg]
		if !ok {
			if cur == nil {
				opts.Rest = append(opts.Rest, arg)
			} else {
				cur.Values = append(cur.Values, arg)
			}
			continue
		}
		f.Present++

		if !f.Toggle {
			i = 0
			cur = f
		}
		i++
	}

	return
}

func (o *Options) Yield() (flags []Flag) {
	for _, fl := range o.Flags {
		if fl.Present > 0 {
			if fl.Exactly > 0 && fl.Exactly != len(fl.Values) {
				continue
			} else {
				if fl.AtLeast > 0 && fl.AtLeast > len(fl.Values) {
					continue
				}
				if fl.AtMost > 0 && fl.AtMost < len(fl.Values) {
					continue
				}
			}
			v := *fl
			flags = append(flags, v)
		}
	}
	return
}
