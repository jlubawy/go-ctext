// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ctext

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"
)

type expected struct {
	TokenType
	Position
}

func TestOffsets(t *testing.T) {
	var cases = []struct {
		Lines    []string
		Expected []expected
	}{
		{
			Lines: []string{
				``,
				`#include <stdio.h>`,
				`/*******************************************************************************`,
				` * hello_world.c`,
				` */ #include <stdio.h>`,
				``,
				`/* PUTS`,
				` * multi-line comment`,
				` */`,
				`#define PUTS( _s ) { \`,
				`    fputs( _s, stdout ); \`,
				`}`,
				``,
				`// Main function`,
				`int`,
				`main( void )`,
				`{`,
				`    PUTS( "Hello World 1\n" ); // comment 1`,
				`    PUTS( "Hello World 2\n" ); // comment 2`,
				`    PUTS(`,
				`        "Hello // World 2\n"`,
				`    ); // prints "Hello // World 2 \n"`,
				`    return 0;`,
				`}`,
				``,
				`#if 0`,
				`/* Allow nested /* comments even though not supported by most compilers */ */`,
				`#endif`,
			},
			Expected: []expected{
				{TextToken, Position{"", 0, 1, 1}},
				{CommentToken, Position{"", 20, 3, 1}},
				{TextToken, Position{"", 121, 5, 4}},
				{CommentToken, Position{"", 142, 7, 1}},
				{TextToken, Position{"", 175, 9, 4}},
				{CommentToken, Position{"", 229, 14, 1}},
				{TextToken, Position{"", 246, 15, 1}},
				{CommentToken, Position{"", 296, 18, 32}},
				{TextToken, Position{"", 309, 19, 1}},
				{CommentToken, Position{"", 340, 19, 32}},
				{TextToken, Position{"", 353, 20, 1}},
				{CommentToken, Position{"", 399, 22, 8}},
				{TextToken, Position{"", 431, 23, 1}},
				{CommentToken, Position{"", 454, 27, 1}},
				{TextToken, Position{"", 531, 27, 78}},
			},
		},
	}

	for i, tc := range cases {
		t.Logf("Test case %d", i)
		actual := make([]expected, 0)

		s := NewScanner(strings.NewReader(strings.Join(tc.Lines, "\n")))
		for {
			tt := s.Next()
			switch tt {
			case ErrorToken:
				err := s.Err()
				if err == io.EOF {
					goto DONE
				}
				t.Fatal(err)

			default:
				actual = append(actual, expected{tt, s.Position})
			}
		}

	DONE:

		if len(tc.Expected) != len(actual) {
			t.Errorf("expected %d results but got %d", len(tc.Expected), len(actual))
		} else {
			for i := 0; i < len(actual); i++ {
				if !reflect.DeepEqual(actual[i], tc.Expected[i]) {
					t.Errorf("Data mismatch on line %d", i)
					t.Errorf("%d '%s':%d:%d:%d", tc.Expected[i].TokenType, tc.Expected[i].Filename, tc.Expected[i].Offset, tc.Expected[i].Line, tc.Expected[i].Column)
					t.Errorf("%d '%s':%d:%d:%d", actual[i].TokenType, actual[i].Filename, actual[i].Offset, actual[i].Line, actual[i].Column)
				}
			}
		}
	}
}

func TestStripComments(t *testing.T) {
	var input = strings.Join([]string{
		`/**`,
		` * hello_world.c`,
		` */`,
		``,
		`// Comment `,
		`#include <stdio.h> `,
		``,
		`// Comment`,
		`int`,
		`main( void )`,
		`{`,
		`    fputs( "Hello world\n", stdout ); // Comment`,
		`    return 0;`,
		`}`,
		``,
	}, "\n")

	var expected = strings.Join([]string{
		``,
		``,
		``,
		``,
		``,
		`#include <stdio.h> `,
		``,
		``,
		`int`,
		`main( void )`,
		`{`,
		`    fputs( "Hello world\n", stdout ); `, // notice space at end of line
		`    return 0;`,
		`}`,
		``,
	}, "\n")

	buf := &bytes.Buffer{}
	if err := StripComments(buf, strings.NewReader(input)); err != nil {
		t.Fatal(err)
	}
	actual := buf.String()
	if actual != expected {
		t.Errorf("%q", expected)
		t.Errorf("%q", actual)
	}
}
