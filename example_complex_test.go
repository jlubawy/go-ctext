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

const ComplexProgram = `
#include <stdio.h>
/*******************************************************************************
 * hello_world.c
 */ #include <stdio.h>

/* PUTS
 * multi-line comment
 */
#define PUTS( _s ) { \
    fputs( _s, stdout ); \
}

// Main function
int
main( void )
{
    PUTS( "Hello World 1\n" ); // comment 1
    PUTS( "Hello World 2\n" ); // comment 2
    PUTS(
        "Hello // World 2\n"
    ); // prints "Hello // World 2 \n"
    return 0;
}

#if 0
/* Allow nested /* comments even though not supported by most compilers */ */
#endif
`

func Example_complex() {
	s := ctext.NewScanner(strings.NewReader(ComplexProgram))
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
	// <text>    hello_world.c:1:1: "\n#include <stdio.h>\n"
	// <comment> hello_world.c:3:1: "/*******************************************************************************\n * hello_world.c\n */"
	// <text>    hello_world.c:5:4: " #include <stdio.h>\n\n"
	// <comment> hello_world.c:7:1: "/* PUTS\n * multi-line comment\n */"
	// <text>    hello_world.c:9:4: "\n#define PUTS( _s ) { \\\n    fputs( _s, stdout ); \\\n}\n\n"
	// <comment> hello_world.c:14:1: "// Main function\n"
	// <text>    hello_world.c:15:1: "int\nmain( void )\n{\n    PUTS( \"Hello World 1\\n\" ); "
	// <comment> hello_world.c:18:32: "// comment 1\n"
	// <text>    hello_world.c:19:1: "    PUTS( \"Hello World 2\\n\" ); "
	// <comment> hello_world.c:19:32: "// comment 2\n"
	// <text>    hello_world.c:20:1: "    PUTS(\n        \"Hello // World 2\\n\"\n    ); "
	// <comment> hello_world.c:22:8: "// prints \"Hello // World 2 \\n\"\n"
	// <text>    hello_world.c:23:1: "    return 0;\n}\n\n#if 0\n"
	// <comment> hello_world.c:27:1: "/* Allow nested /* comments even though not supported by most compilers */ */"
	// <text>    hello_world.c:27:78: "\n#endif\n"
}
