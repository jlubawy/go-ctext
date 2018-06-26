// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ctext_test

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/jlubawy/go-ctext"
)

const BasicProgram = `/**
 * hello_world.c
 */

#include <stdio.h>

int
main( void )
{
    fputs( "Hello world\n", stdout );
    return 0;
}
`

func Example_basic() {
	s := ctext.NewScanner(strings.NewReader(BasicProgram))
	s.Filename = "hello_world.c"
	for {
		tt := s.Next()

		switch tt {
		case ctext.ErrorToken:
			err := s.Err()
			if err == io.EOF {
				return
			}
			log.Fatal(err)

		case ctext.CommentToken:
			fmt.Printf("<comment> %s: %q\n", s.Position, s.TokenText())

		case ctext.TextToken:
			fmt.Printf("<text>    %s: %q\n", s.Position, s.TokenText())
		}
	}

	// Output:
	// <comment> hello_world.c:1:1: "/**\n * hello_world.c\n */"
	// <text>    hello_world.c:3:4: "\n\n#include <stdio.h>\n\nint\nmain( void )\n{\n    fputs( \"Hello world\\n\", stdout );\n    return 0;\n}\n"
}
