// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// This command is used to generate the JSON file used in testing.
// Run the command with 'go run genjson.go <C source file>'
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jlubawy/go-ctext"
)

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		fatalf("Must provide a source file to parse.\n")
	}

	if flag.NArg() > 1 {
		fatalf("Only one source file allowed.\n")
	}

	f, err := os.Open(flag.Arg(0))
	if err != nil {
		fatalf("Error opening file: %v\n", err)
	}
	defer f.Close()

	_, file := filepath.Split(flag.Arg(0))

	tokens := make([]ctext.Token, 0)
	s := ctext.NewScanner(f)
	s.Filename = file
	for {
		tt := s.Next()

		switch tt {
		case ctext.ErrorToken:
			err := s.Err()
			if err == io.EOF {
				goto DONE
			}
			fatalf("Error scanning file: %v\n", err)

		case ctext.CommentToken:
			tokens = append(tokens, s.Token())

		case ctext.TextToken:
			tokens = append(tokens, s.Token())
		}
	}

DONE:
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(&tokens); err != nil {
		fatalf("Error encoding JSON: %v\n", err)
	}
}

func infof(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func fatalf(format string, args ...interface{}) {
	infof(format, args...)
	os.Exit(1)
}
