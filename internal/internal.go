// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package internal

import (
	"bytes"
)

// LastByte returns the last byte in a byte.Buffer, ok is false if the length
// of the buffer is zero.
func LastByte(buf *bytes.Buffer) (b byte, ok bool) {
	var (
		ln = buf.Len()
		bs = buf.Bytes()
	)
	if ln == 0 {
		return
	}
	ok = true
	b = bs[ln-1]
	return
}
