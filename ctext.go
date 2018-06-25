// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package ctext implements a simple C source scanner that can be used to separate
comments from code.
*/
package ctext

import (
	"bufio"
	"bytes"
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

// A Scanner is used to split a C source file into comment and text tokens for
// further processing.
type Scanner struct {
	// Position is the current position of the scanner within a file.
	Position

	posCurr Position

	br  *bufio.Reader
	buf *bytes.Buffer

	err error

	// Should be reset every invocation of Next
	inStringLiteral bool
	mlCommentCount  int
	inSLComment     bool
}

// NewScanner returns a pointer to a new C source scanner.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		posCurr: Position{
			Line:   1,
			Column: 1,
		},

		br:  bufio.NewReader(r),
		buf: &bytes.Buffer{},
	}
}

// Err returns the error associated with the most recent ErrorToken token.
// This is typically io.EOF, meaning the end of tokenization.
func (s *Scanner) Err() error {
	return s.err
}

// Next returns the next token type to be processed.
func (s *Scanner) Next() (tt TokenType) {
	if s.err != nil {
		return ErrorToken // return error right away if one already exists
	}

	// Reset the buffer
	s.buf.Reset()

	// Reset and invalidate the current position
	s.Position.Line = 0
	s.Position.Column = 0

	// Reset scanner fields
	s.inStringLiteral = false
	s.mlCommentCount = 0
	s.inSLComment = false

	for done := false; !done; {
		// Peek one character first so we can skip any chars we don't want
		var bs []byte
		bs, s.err = s.br.Peek(1)
		if s.err != nil {
			if s.err == io.EOF {
				if s.buf.Len() > 0 {
					// If EOF but there is data in the buffer then process it first,
					// the EOF will be returned on the next call to this function.
					if s.mlCommentCount > 0 {
						s.err = errors.New("unexpected end of multi-line comment")
						tt = ErrorToken
					} else if s.inSLComment {
						tt = TextToken
					} else {
						tt = TextToken
					}
					return
				}
			}

			return ErrorToken
		}

		if s.Position.Line == 0 {
			s.Position.Line = s.posCurr.Line
			if s.posCurr.Column == 0 {
				s.Position.Column = 1
			} else {
				s.Position.Column = s.posCurr.Column
			}
		}

		b := bs[0]
		switch b {
		case '/':
			if !s.inSLComment && s.mlCommentCount == 0 {
				// If not in a comment

				if !s.inStringLiteral {
					// If not in a string literal check if this is the start
					// of a single-line comment.
					lb, ok := internal.LastByte(s.buf)
					if ok && lb == '/' {
						// Check if this is the start of a comment
						s.inSLComment = true
						s.Position.Line, s.Position.Column = s.posCurr.Line, s.posCurr.Column-1
					} else if s.buf.Len() > 0 {
						// If the buffer is not empty then process the text first
						tt = TextToken
						return
					}
				}

			} else if s.mlCommentCount > 0 {
				// Else if in a multi-line comment
				lb, ok := internal.LastByte(s.buf)
				if ok && lb == '*' {
					s.mlCommentCount -= 1

					if s.mlCommentCount == 0 {
						tt = CommentToken
						done = true
					}
				}
			} else {
				// Else if in a single-line comment do nothing
			}

		case '*':
			// Possible start or end of multi-line comment
			lb, ok := internal.LastByte(s.buf)
			if ok && lb == '/' {
				s.mlCommentCount += 1
				if s.mlCommentCount == 1 {
					s.Position.Line, s.Position.Column = s.posCurr.Line, s.posCurr.Column-1
				}
			}

		case '\r':
			// Discard and wait for the \n
			_, s.err = s.br.Discard(1)
			if s.err != nil {
				return ErrorToken
			}

			s.posCurr.Column += 1

			continue

		case '\n':
			// Increment the line and reset the current column
			s.posCurr.Line += 1
			s.posCurr.Column = 0

			if s.mlCommentCount > 0 {
				// If in a multi-line comment then continue processing
			} else if s.inSLComment {
				s.inSLComment = false
				tt = CommentToken
				done = true
			}

		case '"':
			if !s.inSLComment && s.mlCommentCount == 0 {
				lb, ok := internal.LastByte(s.buf)
				if ok && lb != '\\' {
					s.inStringLiteral = !s.inStringLiteral
				}
			}
		}

		b, s.err = s.br.ReadByte()
		if s.err != nil {
			// EOF is not expected since we already peeked successfully above
			return ErrorToken
		}

		s.posCurr.Column += 1

		s.err = s.buf.WriteByte(b)
		if s.err != nil {
			return ErrorToken
		}
	}

	return
}

// TokenText returns the string corresponding to the most recently scanned
// token. Valid after calling Scan().
func (s *Scanner) TokenText() string {
	return s.buf.String()
}
