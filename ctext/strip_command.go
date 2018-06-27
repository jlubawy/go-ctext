// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"io"
	"os"

	"github.com/jlubawy/go-cli"
	"github.com/jlubawy/go-ctext"
)

type StripOptions struct {
	Output string
}

var stripOptions StripOptions

var stripCommand = cli.Command{
	Name:             "strip",
	ShortDescription: "strip comments from a C source file",
	Description: `Strips comments from a C source file. If a file is not provided then the source
is read from stdin.`,
	ShortUsage: "[-output output] [source file]",
	SetupFlags: func(fs *flag.FlagSet) {
		fs.StringVar(&stripOptions.Output, "output", "", "output file or stdout if empty")
	},
	Run: func(args []string) {
		var r io.Reader
		if len(args) == 0 {
			r = os.Stdin
		} else if len(args) == 1 {
			f, err := os.OpenFile(args[0], os.O_RDONLY, 0664)
			if err != nil {
				cli.Fatalf("Error opening input file: %v\n", err)
			}
			defer f.Close()
			r = f
		} else {
			cli.Fatal("Expected a single input file.\n")
		}

		var w io.Writer
		if stripOptions.Output == "" {
			w = os.Stdout
		} else {
			f, err := os.OpenFile(stripOptions.Output, os.O_CREATE|os.O_WRONLY, 0664)
			if err != nil {
				cli.Fatalf("Error opening output file: %v\n", err)
			}
			defer f.Close()
			w = f
		}

		if err := ctext.StripComments(w, r); err != nil {
			cli.Fatalf("Error stripping comments: %v\n", err)
		}
	},
}
