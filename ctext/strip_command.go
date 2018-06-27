// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"io"
	"os"

	"github.com/jlubawy/go-ctext"
)

const stripUsage = `usage: strip [-output output] [source file]
Run 'ctext help strip' for details.
`

const stripHelp = `usage: ctext strip [-output output] [source file]

Strips comments from a C source file. If a file is not provided then the source
is read from stdin.

Options:

	-output        file to output the stripped source to, or stdout if empty

`

var stripCommand = Command{
	Name: "strip",
	CmdFn: func(args []string) {
		var flagOutput string

		fs := flag.NewFlagSet("strip", flag.ExitOnError)
		fs.Usage = func() { info(stripUsage) }
		fs.StringVar(&flagOutput, "output", "", "file to output to, stdout if empty")
		fs.Parse(args)

		var r io.Reader
		if fs.NArg() == 0 {
			r = os.Stdin
		} else if fs.NArg() == 1 {
			f, err := os.OpenFile(fs.Arg(0), os.O_RDONLY, 0664)
			if err != nil {
				fatalf("Error opening input file: %v\n", err)
			}
			defer f.Close()
			r = f
		} else {
			fatal("Expected a single input file.\n")
		}

		var w io.Writer
		if flagOutput == "" {
			w = os.Stdout
		} else {
			f, err := os.OpenFile(flagOutput, os.O_CREATE|os.O_WRONLY, 0664)
			if err != nil {
				fatalf("Error opening output file: %v\n", err)
			}
			defer f.Close()
			w = f
		}

		if err := ctext.StripComments(w, r); err != nil {
			fatalf("Error stripping comments: %v\n", err)
		}
	},
	HelpFn: func() { info(stripHelp) },
}
