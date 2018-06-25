// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package internal

import (
	"bytes"
	"testing"
)

func TestLastByte(t *testing.T) {
	var cases = []struct {
		Input    string
		LastByte byte
		Ok       bool
	}{
		{
			Input:    "",
			LastByte: 0,
			Ok:       false,
		},
		{
			Input:    "a",
			LastByte: 'a',
			Ok:       true,
		},
		{
			Input:    "ab",
			LastByte: 'b',
			Ok:       true,
		},
	}

	for _, tc := range cases {
		buf := bytes.NewBufferString(tc.Input)
		lb, ok := LastByte(buf)
		if ok != tc.Ok {
			t.Errorf("expected %t but got %t", tc.Ok, ok)
		} else {
			if ok && lb != tc.LastByte {
				t.Errorf("expected byte 0x%02X but got 0x%02X", tc.LastByte, lb)
			}
		}
	}
}
