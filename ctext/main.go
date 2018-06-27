// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Command ctext exposes some of the functions provided by the ctext package.

Run `ctext help` for usage.
*/
package main

import (
	"flag"
	"fmt"
	"os"
)

type Command struct {
	Name   string
	CmdFn  func(args []string)
	HelpFn func()
}

var commands = []Command{
	stripCommand,
}

const mainUsage = `Usage: ctext command [options]

Available commands:

    strip           strip comments from a C source file

Use "ctext help [command]" for more information about that command.
`

func main() {
	flag.Usage = func() { info(mainUsage) }
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	if flag.NArg() == 1 && flag.Arg(0) == "help" {
		flag.Usage()
		os.Exit(1)
	}

	for _, cmd := range commands {
		if flag.Arg(0) == "help" {
			if flag.Arg(1) == cmd.Name {
				cmd.HelpFn()
				os.Exit(1)
			}
		} else if flag.Arg(0) == cmd.Name {
			cmd.CmdFn(flag.Args()[1:])
			os.Exit(0)
		}
	}

	fatalf(`ctext: unknown command "%s"
Run 'ctext help' for usage.
`, flag.Arg(1))
}

func info(s string) {
	fmt.Fprint(os.Stderr, s)
}

func infof(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func fatal(s string) {
	info(s)
	os.Exit(1)
}

func fatalf(format string, args ...interface{}) {
	infof(format, args...)
	os.Exit(1)
}
