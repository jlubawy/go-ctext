# jlubawy/go-ctext

[![GoDoc](https://godoc.org/github.com/jlubawy/go-ctext?status.svg)](https://godoc.org/github.com/jlubawy/go-ctext)
[![Build Status](https://travis-ci.org/jlubawy/go-ctext.svg?branch=master)](https://travis-ci.org/jlubawy/go-ctext)

Package ctext provides a scanner for the C programming language that separates
source code into comment and text tokens. This may be useful to other programs
that need to scan a program for specific text, while ignoring comments.

For example, consider the following source for a file named ```hello_world.c```:

```c
/**
 * hello_world.c
 */

#include <stdio.h>

int
main( void )
{
    fputs( "Hello world\n", stdout );
    return 0;
}
```

We could write the following program to process each comment and text token:

```go
s := ctext.NewScanner(f)
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
```

The resulting output would look like this:

    <comment> hello_world.c:1:1: "/**\n * hello_world.c\n */"
    <text>    hello_world.c:3:4: "\n\n#include <stdio.h>\n\nint\nmain( void )\n{\n    fputs( \"Hello world\\n\", stdout );\n    return 0;\n}\n"

# Ctext Command

## Installation

Install the latest: ```go get -u github.com/jlubawy/go-ctext/ctext```

## Usage

```
Usage: ctext command [options]

Available commands:

    strip           strip comments from a C source file

Use "ctext help [command]" for more information about that command.
```

## Example

To strip the comments from a C source file:

    curl -sG https://raw.githubusercontent.com/mattn/go-sqlite3/master/sqlite3-binding.c | ctext strip
