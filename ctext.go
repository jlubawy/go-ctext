// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package ctext implements a naive C source scanner that can be used to separate
comments from code.
*/
package ctext

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"github.com/jlubawy/go-ctext/internal"
)

// A TokenType is the type of token.
type TokenType int

const (
	// ErrorToken is an error token type. The error can be retrieved by
	// calling the scanner.Err() method.
	ErrorToken TokenType = iota
	// CommentToken is a comment token type.
	CommentToken
	// TextToken is a text token.
	TextToken
)

// A Position is the position within a file.
type Position struct {
	Filename string // the name of the file
	Line     int    // the line within the file starting at 1
	Column   int    // the column within the line starting at 1
}

func (pos Position) IsValid() bool {
	return pos.Line > 0
}

func (pos Position) String() string {
	s := pos.Filename
	if s == "" {
		s = "<input>"
	}
	if pos.IsValid() {
		s += fmt.Sprintf(":%d:%d", pos.Line, pos.Column)
	}
	return s
}

// A Token is either of type Comment or Text.
type Token struct {
	// Type is the token type.
	Type TokenType

	// Position is the position within a file that the token was found.
	Position

	// Data is the token content.
	Data string
}

// A Scanner is used to split a C source file into comment and text tokens for
// further processing.
type Scanner struct {
	// Position is the current position of the scanner within a file.
	Position

	br     *bufio.Reader
	maxBuf int
	buf    []byte

	err error

	// Track lines and positions
	startPosition Position

	// Should be reset every invocation of Next
	currentTT       TokenType
	inStringLiteral bool
	mlCommentCount  int
	inSLComment     bool
}

// NewScanner returns a pointer to a new C source scanner.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		Position: Position{
			Line:   1,
			Column: 1,
		},

		br:  bufio.NewReader(r),
		buf: make([]byte, 0, 4096),
	}
}

// Err returns the error associated with the most recent ErrorToken token.
// This is typically io.EOF, meaning the end of tokenization.
func (z *Scanner) Err() error {
	if z.currentTT == ErrorToken && z.err == nil {
		panic("token type was error but there was no error")
	}
	return z.err
}

// Next returns the next token type to be processed.
func (z *Scanner) Next() TokenType {
	// Return error right away if one already exists
	if z.err != nil {
		return ErrorToken
	}

	// Reset the buffer length to 0
	z.buf = z.buf[:0]

	// Reset scanner fields
	z.currentTT = ErrorToken
	z.inStringLiteral = false
	z.mlCommentCount = 0
	z.inSLComment = false

	z.startPosition = z.Position
	z.startPosition.Line = 0
	z.startPosition.Column = 0

	for done := false; !done; {
		// Peek one character first so we can skip any chars we don't want
		var bs []byte
		bs, z.err = z.br.Peek(1)
		if z.err != nil {
			if z.err == io.EOF {
				if len(z.buf) > 0 {
					// If EOF but there is data in the buffer then process it first,
					// the EOF will be returned on the next call to this function.
					if z.mlCommentCount > 0 {
						z.err = errors.New("unexpected end of multi-line comment")
						z.currentTT = ErrorToken
					} else if z.inSLComment {
						z.currentTT = TextToken
					} else {
						z.currentTT = TextToken
					}
					return z.currentTT
				}
			}

			return ErrorToken
		}

		if z.startPosition.Line == 0 {
			z.startPosition.Line = z.Position.Line
			if z.Position.Column == 0 {
				z.startPosition.Column = 1
			} else {
				z.startPosition.Column = z.Position.Column
			}
		}

		b := bs[0]
		switch b {
		case '/':
			if !z.inSLComment && z.mlCommentCount == 0 {
				// If not in a comment

				if !z.inStringLiteral {
					// If not in a string literal check if this is the start
					// of a single-line comment.
					lc, ok := internal.LastChar(z.buf)
					if ok && lc == '/' {
						// Check if this is the start of a comment
						z.inSLComment = true
						z.startPosition.Line, z.startPosition.Column = z.Position.Line, z.Position.Column-1
					} else if len(z.buf) > 0 {
						// If the buffer is not empty then process the text first
						z.currentTT = TextToken
						return z.currentTT
					}
				}

			} else if z.mlCommentCount > 0 {
				// Else if in a multi-line comment
				lc, ok := internal.LastChar(z.buf)
				if ok && lc == '*' {
					z.mlCommentCount -= 1

					if z.mlCommentCount == 0 {
						z.currentTT = CommentToken
						done = true
					}
				}
			} else {
				// Else if in a single-line comment do nothing
			}

		case '*':
			// Possible start or end of multi-line comment
			lc, ok := internal.LastChar(z.buf)
			if ok && lc == '/' {
				z.mlCommentCount += 1
				if z.mlCommentCount == 1 {
					z.startPosition.Line, z.startPosition.Column = z.Position.Line, z.Position.Column-1
				}
			}

		case '\r':
			// Discard and wait for the \n
			_, z.err = z.br.Discard(1)
			if z.err != nil {
				return ErrorToken
			}

			z.Position.Column += 1

			continue

		case '\n':
			// Increment the line and reset the current column
			z.Position.Line += 1
			z.Position.Column = 0

			if z.mlCommentCount > 0 {
				// If in a multi-line comment then continue processing
			} else if z.inSLComment {
				z.inSLComment = false
				z.currentTT = CommentToken
				done = true
			}

		case '"':
			if !z.inSLComment && z.mlCommentCount == 0 {
				lc, ok := internal.LastChar(z.buf)
				if ok && lc != '\\' {
					z.inStringLiteral = !z.inStringLiteral
				}
			}
		}

		b, z.err = z.br.ReadByte()
		if z.err != nil {
			// EOF is not expected since we already peeked successfully above
			return ErrorToken
		}

		z.Position.Column += 1

		z.buf, z.err = internal.AddChar(&z.buf, z.maxBuf, b)
		if z.err != nil {
			return ErrorToken
		}
	}

	return z.currentTT
}

// SetMaxBuf sets the maximum buffer allowed by the scanner. Zero is the default
// and it means an unlimited buffer size.
func (z *Scanner) SetMaxBuf(maxBuf uint) {
	z.maxBuf = int(maxBuf)
}

// Token returns the last token returned by Next.
func (z *Scanner) Token() Token {
	return Token{
		Type:     z.currentTT,
		Position: z.startPosition,
		Data:     string(z.buf[:]),
	}
}

// TokenString returns the last token string returned by Next.
func (z *Scanner) TokenString() string {
	return string(z.buf[:])
}
