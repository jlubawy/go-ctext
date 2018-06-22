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
func (s *Scanner) Err() error {
	if s.currentTT == ErrorToken && s.err == nil {
		panic("token type was error but there was no error")
	}
	return s.err
}

// Next returns the next token type to be processed.
func (s *Scanner) Next() TokenType {
	// Return error right away if one already exists
	if s.err != nil {
		return ErrorToken
	}

	// Reset the buffer length to 0
	s.buf = s.buf[:0]

	// Reset scanner fields
	s.currentTT = ErrorToken
	s.inStringLiteral = false
	s.mlCommentCount = 0
	s.inSLComment = false

	s.startPosition = s.Position
	s.startPosition.Line = 0
	s.startPosition.Column = 0

	for done := false; !done; {
		// Peek one character first so we can skip any chars we don't want
		var bs []byte
		bs, s.err = s.br.Peek(1)
		if s.err != nil {
			if s.err == io.EOF {
				if len(s.buf) > 0 {
					// If EOF but there is data in the buffer then process it first,
					// the EOF will be returned on the next call to this function.
					if s.mlCommentCount > 0 {
						s.err = errors.New("unexpected end of multi-line comment")
						s.currentTT = ErrorToken
					} else if s.inSLComment {
						s.currentTT = TextToken
					} else {
						s.currentTT = TextToken
					}
					return s.currentTT
				}
			}

			return ErrorToken
		}

		if s.startPosition.Line == 0 {
			s.startPosition.Line = s.Position.Line
			if s.Position.Column == 0 {
				s.startPosition.Column = 1
			} else {
				s.startPosition.Column = s.Position.Column
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
					lc, ok := internal.LastChar(s.buf)
					if ok && lc == '/' {
						// Check if this is the start of a comment
						s.inSLComment = true
						s.startPosition.Line, s.startPosition.Column = s.Position.Line, s.Position.Column-1
					} else if len(s.buf) > 0 {
						// If the buffer is not empty then process the text first
						s.currentTT = TextToken
						return s.currentTT
					}
				}

			} else if s.mlCommentCount > 0 {
				// Else if in a multi-line comment
				lc, ok := internal.LastChar(s.buf)
				if ok && lc == '*' {
					s.mlCommentCount -= 1

					if s.mlCommentCount == 0 {
						s.currentTT = CommentToken
						done = true
					}
				}
			} else {
				// Else if in a single-line comment do nothing
			}

		case '*':
			// Possible start or end of multi-line comment
			lc, ok := internal.LastChar(s.buf)
			if ok && lc == '/' {
				s.mlCommentCount += 1
				if s.mlCommentCount == 1 {
					s.startPosition.Line, s.startPosition.Column = s.Position.Line, s.Position.Column-1
				}
			}

		case '\r':
			// Discard and wait for the \n
			_, s.err = s.br.Discard(1)
			if s.err != nil {
				return ErrorToken
			}

			s.Position.Column += 1

			continue

		case '\n':
			// Increment the line and reset the current column
			s.Position.Line += 1
			s.Position.Column = 0

			if s.mlCommentCount > 0 {
				// If in a multi-line comment then continue processing
			} else if s.inSLComment {
				s.inSLComment = false
				s.currentTT = CommentToken
				done = true
			}

		case '"':
			if !s.inSLComment && s.mlCommentCount == 0 {
				lc, ok := internal.LastChar(s.buf)
				if ok && lc != '\\' {
					s.inStringLiteral = !s.inStringLiteral
				}
			}
		}

		b, s.err = s.br.ReadByte()
		if s.err != nil {
			// EOF is not expected since we already peeked successfully above
			return ErrorToken
		}

		s.Position.Column += 1

		s.buf, s.err = internal.AddChar(&s.buf, s.maxBuf, b)
		if s.err != nil {
			return ErrorToken
		}
	}

	return s.currentTT
}

// SetMaxBuf sets the maximum buffer allowed by the scanner. Zero is the default
// and it means an unlimited buffer size.
func (s *Scanner) SetMaxBuf(maxBuf uint) {
	s.maxBuf = int(maxBuf)
}

// Token returns the last token returned by Next.
func (s *Scanner) Token() Token {
	return Token{
		Type:     s.currentTT,
		Position: s.startPosition,
		Data:     string(s.buf[:]),
	}
}

// TokenString returns the last token string returned by Next.
func (s *Scanner) TokenString() string {
	return string(s.buf[:])
}
