// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package cmacro provides a scanner for the C programming language that returns
function-like macro invocations. This may be useful to other programs
that need to scan a program for specific macro invocation
*/
package cmacro

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/jlubawy/go-ctext"
	"github.com/jlubawy/go-ctext/internal"
)

// An Invocation is an invocation of a function-like macro within C source code.
type Invocation struct {
	Name       string   // name of the macro invocation
	Start, End int      // lines that the macro invocation starts and ends on
	Args       []string // arguments to the macro invocation if any
}

func (inv Invocation) String() string {
	return fmt.Sprintf("%s( %s );", inv.Name, strings.Join(inv.Args, ", "))
}

// ScanInvocations scans the provided io.Reader for macro invocations that match
// the given names, returning any via the provided callback.
func ScanInvocations(r io.Reader, scanFunc func(inv Invocation), names ...string) (err error) {
	s := ctext.NewScanner(r)
	for {
		tt := s.Next()
		switch tt {
		case ctext.ErrorToken:
			err = s.Err()
			if err == io.EOF {
				err = nil
			}
			return

		case ctext.TextToken:
			err = scanInvocationsTextToken(s.TokenText(), s.Position, scanFunc, names...)
			if err != nil {
				return
			}
		}
	}
	return
}

// ScanInvocationsString scans the provided string for macro invocations that
// match the given names, returning any via the provided callback.
func ScanInvocationsString(s string, scanFunc func(inv Invocation), names ...string) (err error) {
	return ScanInvocations(strings.NewReader(s), scanFunc, names...)
}

// scanInvocationsTextToken scans the provided text token for macro invocations
// that match the given names, returning any via the provided callback.
func scanInvocationsTextToken(s string, pos ctext.Position, scanFunc func(inv Invocation), names ...string) (err error) {
	var re *regexp.Regexp
	re, err = compileNamesRegexp(names...)
	if err != nil {
		return
	}

	var (
		lineCurr = pos.Line
		inv      Invocation
	)
	for {
		// Find the next instance of the macro name
		loc := re.FindStringIndex(s)
		if loc == nil {
			goto DONE
		}

		var (
			ni   = loc[0]
			name = s[loc[0]:loc[1]]
		)

		// Count all line-endings before this name
		for i := 0; i < ni; i++ {
			if s[i] == '\n' {
				lineCurr += 1
			}
		}

		// Check if this is a macro definition and not an invocation
		isDef := isMacroDefinition(s, ni)

		// Shorten the string length to look after the name
		s = s[ni+len(name):]

		if isDef {
			continue // skip if this was a definition
		}

		// Reset the local invocation
		inv.Name = name
		inv.Args = make([]string, 0)
		inv.Start = lineCurr

		// Find the opening parentheses
		opi := strings.Index(s, "(")
		if opi == -1 {
			err = errors.New("macro function missing opening parentheses")
			return
		}

		// Parse each character after the opening parentheses
		var (
			done            bool
			inStringLiteral bool
			parenCount      int
			buf             = &bytes.Buffer{}
		)

		// Iterate over the rest of the characters
		i := opi + 1
		for ; (i < len(s)) && !done; i++ {
			b := s[i]

			switch b {
			case ' ':
				if inStringLiteral || parenCount > 0 {
					// If in a string literal then add the space
					err = buf.WriteByte(b)
					if err != nil {
						return
					}
				} else if parenCount == 0 {
					// Else it's probably the end of an argument
					arg, ok := parseInvocationArg(buf)
					if ok {
						inv.Args = append(inv.Args, strings.TrimSpace(string(arg)))
					}
				}

			case ',':
				if inStringLiteral || parenCount > 0 {
					// If in a string literal add the comma
					err = buf.WriteByte(b)
					if err != nil {
						return
					}
				} else if parenCount == 0 {
					// Else it's probably the end of an argument
					arg, ok := parseInvocationArg(buf)
					if ok {
						inv.Args = append(inv.Args, strings.TrimSpace(string(arg)))
					}
				}

			case '"':
				if inStringLiteral {
					lb, ok := internal.LastByte(buf)
					if ok && lb == '\\' {
						// If in a string literal, but this quote was escaped
						// then add it to the buffer
						err = buf.WriteByte(b)
						if err != nil {
							return
						}
					} else {
						// Else leaving a string literal, which has to be the end
						// of an argument
						inStringLiteral = false
						err = buf.WriteByte(b)
						if err != nil {
							return
						}
						if parenCount == 0 {
							arg, ok := parseInvocationArg(buf)
							if ok {
								inv.Args = append(inv.Args, strings.TrimSpace(string(arg)))
							}
						}
					}
				} else {
					// Else not in a string literal, so we are now
					inStringLiteral = true
					err = buf.WriteByte(b)
					if err != nil {
						return
					}
				}

			case '(':
				// Add any opening paren
				err = buf.WriteByte(b)
				if err != nil {
					return
				}

				if !inStringLiteral {
					// Probably an invocation of a macro/func within an invocation
					parenCount += 1
				}

			case ')':
				if inStringLiteral {
					// If in a string literal add the closing paren
					err = buf.WriteByte(b)
					if err != nil {
						return
					}
				} else {
					if parenCount > 0 {
						// If inside another invocation add the closing paren
						err = buf.WriteByte(b)
						if err != nil {
							return
						}

						// Only decrement if > 0
						parenCount -= 1
					}

					if parenCount == 0 {
						arg, ok := parseInvocationArg(buf)
						if ok {
							inv.Args = append(inv.Args, strings.TrimSpace(string(arg)))
						}
					}
				}

			case ';':
				if inStringLiteral {
					// If in a string literal add the semi-colon
					err = buf.WriteByte(b)
					if err != nil {
						return
					}
				} else {
					// Else if not in a string literal, close out the invocation
					// and find the next one.
					inv.End = lineCurr
					scanFunc(inv)
					done = true
				}

			case '\r':
				// discard carriage returns, wait for newline

			case '\n':
				lineCurr += 1

			default:
				err = buf.WriteByte(b)
				if err != nil {
					return
				}
			}
		}

		// If we've reached the end of the string break out of the outer loop
		if i >= len(s) {
			break
		}
	}

DONE:
	return
}

func isMacroDefinition(s string, ni int) (isDef bool) {
	if ni >= 8 {
		i := ni - 1
		for ; i >= 0; i-- {
			if s[i] != ' ' {
				break
			}
		}
		if i >= 6 {
			if s[i-5:i+1] == "define" {
				for j := i - 6; j >= 0; j-- {
					switch s[j] {
					case ' ':
						// skip spaces
					case '#':
						isDef = true
						return
					default:
						return
					}
				}
			}
		}
	}
	return
}

// parseInvocationArg returns an argument string and shortens the buffer length to zero if
// the provided buffer isn't empty.
func parseInvocationArg(buf *bytes.Buffer) (arg string, ok bool) {
	arg = strings.TrimSpace(buf.String())
	if len(arg) > 0 {
		ok = true
		buf.Reset()
	}
	return
}

// compileNamesRegexp takes a list of macro names and compiles a regexp to match
// all of them.
func compileNamesRegexp(names ...string) (re *regexp.Regexp, err error) {
	ns := make([]string, len(names))

	// Quote any meta characters since the names are searched as is
	for i := 0; i < len(names); i++ {
		ns[i] = "(\\b" + regexp.QuoteMeta(names[i]) + ")"
	}
	return regexp.Compile(strings.Join(ns, "|"))
}
