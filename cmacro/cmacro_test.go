// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmacro

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jlubawy/go-ctext"
)

func TestIsMacroDef(t *testing.T) {
	var cases = []struct {
		Input string
		IsDef bool
	}{
		{
			Input: `TEST_FUNC( a, b, c )  (a, b, c)`,
			IsDef: false,
		},
		{
			Input: `#define TEST_FUNC( a, b, c )  (a, b, c)`,
			IsDef: true,
		},
		{
			Input: `  #  define   TEST_FUNC( a, b, c )  (a, b, c)`,
			IsDef: true,
		},
	}

	for i, tc := range cases {
		t.Logf("Test Case: %d", i)

		ni := strings.Index(tc.Input, "TEST_FUNC")
		if ni == -1 {
			t.Fatal("expected to find TEST_FUNC but didn't")
		}

		if isDef := isMacroDef(tc.Input, ni); isDef != tc.IsDef {
			t.Errorf("expected %t but got %t", tc.IsDef, isDef)
		}
	}
}

func newTextToken(s string) *ctext.Token {
	return &ctext.Token{
		Type: ctext.TextToken,
		Data: s,
		Position: ctext.Position{
			Line:   1,
			Column: 1,
		},
	}
}

func TestFindMacroFuncs(t *testing.T) {
	var cases = []struct {
		Input      *ctext.Token
		Names      []string
		MacroFuncs []MacroFunc
		ExpErr     bool
	}{
		{
			Input:      newTextToken(`#define TEST_FUNC( a, b, c )  (a, b, c)`),
			MacroFuncs: []MacroFunc{},
			ExpErr:     false,
		},
		{
			Input: newTextToken(`TEST_FUNC ( );`),
			Names: []string{"TEST_FUNC"},
			MacroFuncs: []MacroFunc{
				{
					Name:      "TEST_FUNC",
					Args:      []string{},
					LineStart: 1,
					LineEnd:   1,
				},
			},
			ExpErr: false,
		},
		{
			Input: newTextToken(`TEST_FUNC( INNER_TEST_FUNC() );`),
			Names: []string{"TEST_FUNC"},
			MacroFuncs: []MacroFunc{
				{
					Name:      "TEST_FUNC",
					Args:      []string{"INNER_TEST_FUNC()"},
					LineStart: 1,
					LineEnd:   1,
				},
			},
			ExpErr: false,
		},
		{
			Input: newTextToken(`TEST_FUNC ( "Format string: %d %s %d", a, "b \\ string", c );`),
			Names: []string{"TEST_FUNC"},
			MacroFuncs: []MacroFunc{
				{
					Name:      "TEST_FUNC",
					Args:      []string{"\"Format string: %d %s %d\"", "a", "\"b \\\\ string\"", "c"},
					LineStart: 1,
					LineEnd:   1,
				},
			},
			ExpErr: false,
		},
		{
			Input: newTextToken(`TEST_FUNC(
								    "Format string: %d %s %d",
								    a,
								    "b \\ string",
								    c
								);`),
			Names: []string{"TEST_FUNC"},
			MacroFuncs: []MacroFunc{
				{
					Name:      "TEST_FUNC",
					Args:      []string{"\"Format string: %d %s %d\"", "a", "\"b \\\\ string\"", "c"},
					LineStart: 1,
					LineEnd:   6,
				},
			},
			ExpErr: false,
		},
		{
			Input: newTextToken(`TEST_FUNC ( "Format string 1: %d %s %d", a, "b \\ string", c );
  								 TEST_FUNC( "Format string 2: %d %s %d", "d \\ string", e , f);`),
			Names: []string{"TEST_FUNC"},
			MacroFuncs: []MacroFunc{
				{
					Name:      "TEST_FUNC",
					Args:      []string{"\"Format string 1: %d %s %d\"", "a", "\"b \\\\ string\"", "c"},
					LineStart: 1,
					LineEnd:   1,
				},
				{
					Name:      "TEST_FUNC",
					Args:      []string{"\"Format string 2: %d %s %d\"", "\"d \\\\ string\"", "e", "f"},
					LineStart: 2,
					LineEnd:   2,
				},
			},
			ExpErr: false,
		},
		{
			Input: newTextToken(`#define TEST_FUNC( _fmt, ... )  func( _fmt, __VA_ARGS__ )
						  		 TEST_FUNC ( "Format string 1: %d %s %d", a, "b \\ string", c );
						         TEST_FUNC( "Format string 2: %d %s %d", "d \\ string", e , f);`),
			Names: []string{"TEST_FUNC"},
			MacroFuncs: []MacroFunc{
				{
					Name:      "TEST_FUNC",
					Args:      []string{"\"Format string 1: %d %s %d\"", "a", "\"b \\\\ string\"", "c"},
					LineStart: 2,
					LineEnd:   2,
				},
				{
					Name:      "TEST_FUNC",
					Args:      []string{"\"Format string 2: %d %s %d\"", "\"d \\\\ string\"", "e", "f"},
					LineStart: 3,
					LineEnd:   3,
				},
			},
			ExpErr: false,
		},
		{
			Input: newTextToken(`TEST_FUNC_A( "Format string 1: %d %s %d", a, "b \\ string", c );
								 TEST_FUNC_B( "Format string 2: %d %s %d", "d \\ string", e , f);`),
			Names: []string{"TEST_FUNC_A", "TEST_FUNC_B"},
			MacroFuncs: []MacroFunc{
				{
					Name:      "TEST_FUNC_A",
					Args:      []string{"\"Format string 1: %d %s %d\"", "a", "\"b \\\\ string\"", "c"},
					LineStart: 1,
					LineEnd:   1,
				},
				{
					Name:      "TEST_FUNC_B",
					Args:      []string{"\"Format string 2: %d %s %d\"", "\"d \\\\ string\"", "e", "f"},
					LineStart: 2,
					LineEnd:   2,
				},
			},
			ExpErr: false,
		},
		{
			Input: newTextToken(`TEST_FUNC_A( "Format string 1: %d %s %d", a, "b \\ string", c );  // comment 1
								 TEST_FUNC_B( "Format string 2: %d %s %d",
									  		  "d \\ string",
										   	  e,
											  f );`),
			Names: []string{"TEST_FUNC_A", "TEST_FUNC_B"},
			MacroFuncs: []MacroFunc{
				{
					Name:      "TEST_FUNC_A",
					Args:      []string{"\"Format string 1: %d %s %d\"", "a", "\"b \\\\ string\"", "c"},
					LineStart: 1,
					LineEnd:   1,
				},
				{
					Name:      "TEST_FUNC_B",
					Args:      []string{"\"Format string 2: %d %s %d\"", "\"d \\\\ string\"", "e", "f"},
					LineStart: 2,
					LineEnd:   5,
				},
			},
			ExpErr: false,
		},
		{
			Input: newTextToken(`TEST_FUNC( "%d",
					                       	1,
					                       	INNER_TEST_FUNC( 123, "Test" )
		                        );`),
			Names: []string{"TEST_FUNC"},
			MacroFuncs: []MacroFunc{
				{
					Name:      "TEST_FUNC",
					Args:      []string{"\"%d\"", "1", "INNER_TEST_FUNC( 123, \"Test\" )"},
					LineStart: 1,
					LineEnd:   4,
				},
			},
			ExpErr: false,
		},

		// Errors
		{
			Input:      newTextToken(`TEST_FUNC  );`),
			Names:      []string{"TEST_FUNC"},
			MacroFuncs: []MacroFunc{},
			ExpErr:     true,
		},
	}

	for i, tc := range cases {
		t.Logf("Test Case: %d", i)
		mfs, err := FindMacroFuncs(tc.Input, tc.Names...)
		if err != nil {
			if !tc.ExpErr {
				t.Errorf("unexpected error: %v", err)
			}
		} else {
			if tc.ExpErr {
				t.Error("expected error")
			} else {
				if !reflect.DeepEqual(tc.MacroFuncs, mfs) {
					t.Error("data mismatch")
					t.Errorf("%+v", tc.MacroFuncs)
					t.Errorf("%+v", mfs)
				}
			}
		}
	}
}
