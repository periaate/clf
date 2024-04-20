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
	Must    bool

	Cb func([]string)

	Keys        []string
	Name        string
	Description string
}

type Options struct {
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

func Parse(args []string, flags []*Flag) (opts *Options, err error) {
	opts = &Options{
		Flags: map[string]*Flag{},
		Rest:  make([]string, 0),
	}
	lookup := map[string]*Flag{}
	for _, flag := range flags {
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
			lookup[flag.Name] = flag
			continue
		}

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
			if cur == nil || cur.Exactly == -1 {
				opts.Rest = append(opts.Rest, arg)
			} else {
				i++
				cur.Values = append(cur.Values, arg)
			}
			continue
		}
		i = 0
		f.Present++

		cur = f
	}

	return
}

func Cloner(f *Flag) func(*Flag) *Flag {
	return func(v *Flag) *Flag {
		v.AtLeast = DefSet(v.AtLeast, f.AtLeast)
		v.AtMost = DefSet(v.AtMost, f.AtMost)
		v.Exactly = DefSet(v.Exactly, f.Exactly)
		v.Must = DefSet(v.Must, f.Must)
		v.Toggle = DefSet(v.Toggle, f.Toggle)
		v.Greedy = DefSet(v.Greedy, f.Greedy)
		return v
	}
}

func DefSet[T comparable](a, b T) (zero T) {
	if a == zero {
		return b
	}
	return a
}

var helpFlag *Flag

func SetHelp(s string) {
	helpFlag = &Flag{
		Keys:        []string{"-h", "--help"},
		Name:        "help",
		Description: s,
	}
}

const (
	h1 = "-h"
	h2 = "--help"
)

var name string

func SetName(s string) {
	name = s
}

func CheckHelp(args []string) bool {
	if helpFlag == nil {
		return false
	}
	for _, arg := range args {
		if arg == h1 || arg == h2 {
			PrintHelp()
			return true
		}
	}
	return false
}

func PrintHelp() {
	if helpFlag == nil {
		return
	}
	fmt.Printf(helpFlag.Description, name)
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

func AnyOf[C comparable](arr []C) {}

func OneOf[C comparable](arr []C) {}

func (o *Options) Run() bool {
	for _, r := range o.Yield() {
		if r.Cb != nil {
			r.Cb(o.Rest)
			return true
		}
	}
	return false
}

type Runner struct {
	Must bool
}

func (r *Runner) Run(args []string, flags ...*Flag) ([]string, bool) {
	if len(args) == 0 || len(flags) == 0 {
		return args, false
	}

	opts, err := Parse(args, flags)
	if err != nil {
		panic(err)
	}
	b := opts.Run()

	return opts.Rest, b
}

func Run(args []string, flags []*Flag) []string {
	opts, err := Parse(args, flags)
	if err != nil {
		panic(err)
	}
	opts.Run()

	return opts.Rest
}

func RunR(args []string, flags ...*Flag) []string {
	if len(args) == 0 || len(flags) == 0 {
		return args
	}

	opts, err := Parse(args, flags)
	if err != nil {
		panic(err)
	}
	opts.Run()

	return opts.Rest
}

/*
args []string ->

arg[0] ...{
	Command = OneOf
	Flag	= AnyOf
}

Commands
	Lazy and strict;

Flags
	Lossy, can be greedy;



func(
	{"build", "b"}
)

{
	"get": get(Any),
	"set": set(get(Any)),
	"_": get(Any),
}

*/
