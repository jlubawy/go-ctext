// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package internal

import (
	"testing"
)

func TestAddChar(t *testing.T) {
	var cases = []struct {
		Buf    []byte
		MaxBuf int
		ExpErr bool
	}{
		// No errors
		{
			Buf:    make([]byte, 0, 0),
			MaxBuf: 0,
			ExpErr: false,
		},
		{
			Buf:    make([]byte, 0, 0),
			MaxBuf: 1,
			ExpErr: false,
		},
		{
			Buf:    make([]byte, 0, 1),
			MaxBuf: 1,
			ExpErr: false,
		},

		// Errors
		{
			Buf:    make([]byte, 1, 1),
			MaxBuf: 1,
			ExpErr: true,
		},
	}

	for _, tc := range cases {
		buf, err := AddChar(&tc.Buf, tc.MaxBuf, 'c')
		if err != nil {
			if !tc.ExpErr {
				t.Errorf("unexpected error: %v", err)
			}
		} else {
			if tc.ExpErr {
				t.Error("expected error")
			} else {
				if buf[len(buf)-1] != 'c' {
					t.Error("unexpected character")
				}
			}
		}
	}
}
