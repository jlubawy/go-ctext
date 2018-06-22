// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package internal

import (
	"errors"
)

var ErrMaxBufferReached = errors.New("maximum buffer reached")

// AddChar adds a character to the provided slice, expanding it up to maxBuf if
// needed, returning an error if the maximum buffer size if reached. If maxBuf is
// zero then there is no limit.
func AddChar(bufP *[]byte, maxBuf int, b byte) (buf []byte, err error) {
	if maxBuf < 0 {
		maxBuf = 0
	}

	buf = *bufP

	// Check if there is room to add a new char
	if len(buf)+1 > cap(buf) {
		if maxBuf > 0 && cap(buf) >= maxBuf {
			// If the capacity is already at max then return an error
			err = ErrMaxBufferReached
			return
		}

		// Try to double the buffer capacity, limiting to the maximum
		dc := 2 * (1 + cap(buf))
		if maxBuf > 0 && dc > maxBuf {
			dc = maxBuf
		}

		newBuf := make([]byte, 0, dc)
		buf = append(newBuf, buf...)
	}

	buf = buf[:len(buf)+1]
	buf[len(buf)-1] = b
	return
}

// DiscardChar discards the last character
func DiscardChar(bufP *[]byte) (buf []byte) {
	buf = *bufP
	if len(buf) > 0 {
		buf = buf[:len(buf)-1]
	}
	return
}

// LastChar returns the last character in a buffer, ok is false if the length
// of the buffer is zero.
func LastChar(buf []byte) (b byte, ok bool) {
	if len(buf) == 0 {
		return
	}
	ok = true
	b = buf[len(buf)-1]
	return
}
