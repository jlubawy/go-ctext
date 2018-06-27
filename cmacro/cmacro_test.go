// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmacro

import (
	"reflect"
	"strings"
	"testing"
)

func TestIsMacroDefinition(t *testing.T) {
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

		if isDef := isMacroDefinition(tc.Input, ni); isDef != tc.IsDef {
			t.Errorf("expected %t but got %t", tc.IsDef, isDef)
		}
	}
}

func TestScanInvocationsString(t *testing.T) {
	var cases = []struct {
		Input string
		Names []string

		Expected []Invocation
		ExpErr   bool
	}{
		{
			Input:    `#define TEST_FUNC( a, b, c )  (a, b, c)`,
			Expected: []Invocation{},
			ExpErr:   false,
		},
		{
			Input: `TEST_FUNC ( );`,
			Names: []string{"TEST_FUNC"},
			Expected: []Invocation{
				{
					Name:  "TEST_FUNC",
					Args:  []string{},
					Start: 1,
					End:   1,
				},
			},
			ExpErr: false,
		},
		{
			Input: `TEST_FUNC( INNER_TEST_FUNC() );`,
			Names: []string{"TEST_FUNC"},
			Expected: []Invocation{
				{
					Name:  "TEST_FUNC",
					Args:  []string{"INNER_TEST_FUNC()"},
					Start: 1,
					End:   1,
				},
			},
			ExpErr: false,
		},
		{
			Input: `TEST_FUNC ( "Format string: %d %s %d", a, "b \\ string", c );`,
			Names: []string{"TEST_FUNC"},
			Expected: []Invocation{
				{
					Name:  "TEST_FUNC",
					Args:  []string{"\"Format string: %d %s %d\"", "a", "\"b \\\\ string\"", "c"},
					Start: 1,
					End:   1,
				},
			},
			ExpErr: false,
		},
		{
			Input: `TEST_FUNC(
								    "Format string: %d %s %d",
								    a,
								    "b \\ string",
								    c
								);`,
			Names: []string{"TEST_FUNC"},
			Expected: []Invocation{
				{
					Name:  "TEST_FUNC",
					Args:  []string{"\"Format string: %d %s %d\"", "a", "\"b \\\\ string\"", "c"},
					Start: 1,
					End:   6,
				},
			},
			ExpErr: false,
		},
		{
			Input: `TEST_FUNC ( "Format string 1: %d %s %d", a, "b \\ string", c );
  								 TEST_FUNC( "Format string 2: %d %s %d", "d \\ string", e , f);`,
			Names: []string{"TEST_FUNC"},
			Expected: []Invocation{
				{
					Name:  "TEST_FUNC",
					Args:  []string{"\"Format string 1: %d %s %d\"", "a", "\"b \\\\ string\"", "c"},
					Start: 1,
					End:   1,
				},
				{
					Name:  "TEST_FUNC",
					Args:  []string{"\"Format string 2: %d %s %d\"", "\"d \\\\ string\"", "e", "f"},
					Start: 2,
					End:   2,
				},
			},
			ExpErr: false,
		},
		{
			Input: `#define TEST_FUNC( _fmt, ... )  func( _fmt, __VA_ARGS__ )
						  		 TEST_FUNC ( "Format string 1: %d %s %d", a, "b \\ string", c );
						         TEST_FUNC( "Format string 2: %d %s %d", "d \\ string", e , f);`,
			Names: []string{"TEST_FUNC"},
			Expected: []Invocation{
				{
					Name:  "TEST_FUNC",
					Args:  []string{"\"Format string 1: %d %s %d\"", "a", "\"b \\\\ string\"", "c"},
					Start: 2,
					End:   2,
				},
				{
					Name:  "TEST_FUNC",
					Args:  []string{"\"Format string 2: %d %s %d\"", "\"d \\\\ string\"", "e", "f"},
					Start: 3,
					End:   3,
				},
			},
			ExpErr: false,
		},
		{
			Input: `TEST_FUNC_A( "Format string 1: %d %s %d", a, "b \\ string", c );
								 TEST_FUNC_B( "Format string 2: %d %s %d", "d \\ string", e , f);`,
			Names: []string{"TEST_FUNC_A", "TEST_FUNC_B"},
			Expected: []Invocation{
				{
					Name:  "TEST_FUNC_A",
					Args:  []string{"\"Format string 1: %d %s %d\"", "a", "\"b \\\\ string\"", "c"},
					Start: 1,
					End:   1,
				},
				{
					Name:  "TEST_FUNC_B",
					Args:  []string{"\"Format string 2: %d %s %d\"", "\"d \\\\ string\"", "e", "f"},
					Start: 2,
					End:   2,
				},
			},
			ExpErr: false,
		},
		{
			Input: `TEST_FUNC_A( "Format string 1: %d %s %d", a, "b \\ string", c );  // comment 1
								 TEST_FUNC_B( "Format string 2: %d %s %d",
									  		  "d \\ string",
										   	  e,
											  f );`,
			Names: []string{"TEST_FUNC_A", "TEST_FUNC_B"},
			Expected: []Invocation{
				{
					Name:  "TEST_FUNC_A",
					Args:  []string{"\"Format string 1: %d %s %d\"", "a", "\"b \\\\ string\"", "c"},
					Start: 1,
					End:   1,
				},
				{
					Name:  "TEST_FUNC_B",
					Args:  []string{"\"Format string 2: %d %s %d\"", "\"d \\\\ string\"", "e", "f"},
					Start: 2,
					End:   5,
				},
			},
			ExpErr: false,
		},
		{
			Input: `TEST_FUNC( "%d",
					                       	1,
					                       	INNER_TEST_FUNC( 123, "Test" )
		                        );`,
			Names: []string{"TEST_FUNC"},
			Expected: []Invocation{
				{
					Name:  "TEST_FUNC",
					Args:  []string{"\"%d\"", "1", "INNER_TEST_FUNC( 123, \"Test\" )"},
					Start: 1,
					End:   4,
				},
			},
			ExpErr: false,
		},

		// Errors
		{
			Input:    `TEST_FUNC  );`,
			Names:    []string{"TEST_FUNC"},
			Expected: []Invocation{},
			ExpErr:   true,
		},
	}

	for i, tc := range cases {
		t.Logf("Test Case: %d", i)

		actual := make([]Invocation, 0)
		err := ScanInvocationsString(tc.Input, func(i Invocation) { actual = append(actual, i) }, tc.Names...)
		if err != nil {
			if !tc.ExpErr {
				t.Errorf("unexpected error: %v", err)
			}
		} else {
			if tc.ExpErr {
				t.Error("expected error")
			} else {
				if !reflect.DeepEqual(tc.Expected, actual) {
					t.Error("data mismatch")
					t.Errorf("%+v", tc.Expected)
					t.Errorf("%+v", actual)
				}
			}
		}
	}
}
