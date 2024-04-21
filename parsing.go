package clf

import (
	"fmt"
	"log"
)

type Program struct {
	Info     Meta
	Commands *Commands
}

func (pr *Program) EvalOnly(args []string, flagNames []string) (rest []string) {
	flags := []*Flag{}
	for _, name := range flagNames {
		for _, flag := range pr.Commands.Flags {
			if flag.Name == name {
				flags = append(flags, flag)
			}
		}
	}

	opt, err := Parse(args, flags)
	if err != nil {
		panic(err)
	}

	return opt.Rest
}

func (pr *Program) PrintedHelp() bool { return printedHelp }

func (pr *Program) Help() {
	fmt.Printf("Usage: %s [commands]\n", pr.Info.Name)
	fmt.Printf("  %s\n", pr.Info.Description)
	fmt.Printf("\n")
	fmt.Printf("Options:\n")
	var groupname string
	var grouped [][][4]string
	for _, flag := range pr.Commands.Flags {
		a := fmt.Sprintf("%v", flag.Keys)
		b := flag.Name
		d := flag.Description
		if groupname != flag.Group {
			grouped = append(grouped, [][4]string{})
			groupname = flag.Group
		}

		grouped[len(grouped)-1] = append(grouped[len(grouped)-1], [4]string{a, b, d, flag.Group})
	}

	for _, group := range grouped {
		fmt.Printf("=== %s ===\n", group[0][3])
		var sar []string
		for _, f := range group {
			sar = append(sar, f[0], f[1], f[2])
		}
		printSLn(sar...)
		fmt.Print("\n")
	}

	printedHelp = true
}

func (pr *Program) Eval(args []string) (rest []string, err error) {
	var opt *Options
	opt, err = Parse(args, pr.Commands.Flags)
	if err != nil {
		panic(err)
	}
	rest = opt.Rest
	return
}

type Option func(*Program) *Program

func Register(opts ...Option) *Program {
	pr := &Program{
		Commands: &Commands{
			Flags: []*Flag{},
		},
		Info: Meta{},
	}
	help = pr.Help
	for _, opt := range opts {
		pr = opt(pr)
	}

	return pr
}

type Commands struct {
	Flags []*Flag
}

func Flags(flags []*Flag) Option {
	return func(pr *Program) *Program {
		pr.Commands.Flags = flags
		_, _ = ParseNames(flags)
		return pr
	}
}

type Meta struct {
	Name        string
	Description string
	Author      string
	Source      string
	Copyright   string
	Version     string
}

func Info(opts Meta) Option {
	return func(pr *Program) *Program {
		pr.Info = opts
		return pr
	}
}

func Group(name string, flags ...*Flag) []*Flag {
	for _, flag := range flags {
		flag.Group = name
	}
	return flags
}

func DefaultHelp(handler func()) *Flag {
	return &Flag{
		Keys:        []string{"help", "-h"},
		Name:        "help",
		Description: "Prints this help message.",
		Exactly:     -1,
		Handler: func([]string) {
			help()
			if handler != nil {
				handler()
			}
		},
	}
}

var printedHelp = false
var help = func() {
	printedHelp = true
}

// printSLn -- print structured line
func printSLn(pairs ...string) {
	if len(pairs)%3 != 0 {
		log.Fatalln("input must be even length, got", len(pairs), "instead")
	}

	maxL1 := 0
	maxL2 := 0
	for i := 0; i < len(pairs); i += 3 {
		if len(pairs[i]) > maxL1 {
			maxL1 = len(pairs[i])
		}
		if len(pairs[i+1]) > maxL2 {
			maxL2 = len(pairs[i+1])
		}
	}

	for i := 0; i < len(pairs); i += 3 {
		a := pairs[i]
		b := pairs[i+1]
		c := pairs[i+2]
		fmt.Printf("  %-*s\t%-*s: %s\n", maxL1, a, maxL2, b, c)
	}
}
