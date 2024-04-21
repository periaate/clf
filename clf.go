package clf

import (
	"fmt"

	"github.com/periaate/common"
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
		if v, ok := opts.Names[flag.Name]; ok {
			err = fmt.Errorf("names must be unique! name %s used by both %v and %v", flag.Name, flag, v)
			return
		}
		opts.Names[flag.Name] = flag

		if len(flag.Keys) == 0 {
			opts.Flags[flag.Name] = flag
			continue
		}

		for _, key := range flag.Keys {
			if v, ok := opts.Flags[key]; ok {
				err = fmt.Errorf("keys must be unique! %s collides with %s with key %s", flag.Name, v.Name, key)
				return
			}
			opts.Flags[key] = flag
		}
	}
	return
}

var glog common.Logger = common.DummyLogger{}

func SetGlobalLogger(s common.Logger) { glog = s }

func Parse(args []string, flags []*Flag) (opts *Options, err error) {
	opts, err = ParseNames(flags)
	for k, f := range opts.Flags {
		glog.Debug("REGISTERED FLAG", "FLAG", f.Name, "KEYS", f.Keys, "KEY", k)
	}

	var cur *Flag
	var i int
	var capturing bool
	reset := func() {
		if cur != nil {
			glog.Debug("resetting args", "target", cur.Name, "captured", cur.Values)
			if cur.Handler != nil {
				glog.Debug("calling handler", "target", cur.Name)
				cur.Handler(cur.Values)
				cur.Values = nil
			}
		}
		cur = nil
		i = 0
		capturing = false
	}
	canReset := func() bool {
		if cur == nil {
			return true
		}
		if cur.Exactly == -1 {
			glog.Debug("can reset", "cause", "flag isn't capturing", "name", cur.Name)
			return true
		}
		if cur.Exactly != 0 && cur.Exactly == i {
			glog.Debug("can reset", "cause", "EXC fulfilled", "name", cur.Name, "i", i, "EXC", cur.Exactly)
			return true
		}
		r := cur.AtLeast <= i
		if r {
			glog.Debug("can reset", "cause", "ATL fulfilled", "name", cur.Name, "i", i, "ATL", cur.AtLeast)
		} else {
			glog.Debug("can't reset", "cause", "ATL not fulfilled", "name", cur.Name, "i", i, "ATL", cur.AtLeast)
		}
		return r
	}

	shouldReset := func() bool {
		if cur == nil {
			return false
		}
		if cur.Exactly != 0 && (cur.Exactly == i || cur.Exactly == -1) {
			glog.Debug("should reset", "cause", "EXC", "name", cur.Name, "i", i, "EXC", cur.Exactly)
			return true
		}

		if cur.AtMost != 0 {
			if cur.AtLeast != 0 {
				if cur.AtLeast <= i {
					return false
				}
			}
			if cur.AtMost < i {
				glog.Debug("should reset", "cause", "ATM", "name", cur.Name, "i", i, "ATM", cur.AtMost)
				return true
			}
		}
		return false
	}
	begin := func(f *Flag, arg string) {
		if cur != nil {
			reset()
		}
		glog.Debug("begin flag capture", "name", f.Name, "arg", arg, "ATM", f.AtMost, "ATL", f.AtLeast, "EXC", f.Exactly)
		cur = f
		capturing = true
	}
	value := func(arg string) {
		i++
		glog.Debug("FOUND MATCH", "FLAG", cur.Name, "ARG", arg, "INDEX", i)
		cur.Values = append(cur.Values, arg)
	}

	for _, arg := range args {
		if canReset() {
			if shouldReset() {
				reset()
			}

			f, ok := opts.Flags[arg]
			if ok {
				begin(f, arg)
				continue
			}

		}

		switch {
		case capturing:
			value(arg)
			continue
		}

		glog.Debug("not capturing and no flag found, adding to rest", "arg", arg)
		opts.Rest = append(opts.Rest, arg)
	}

	reset()

	return
}
