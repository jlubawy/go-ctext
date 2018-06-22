// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package cmacro finds all function-like macros called within a C source file.
*/
package cmacro

import (
	"errors"
	"regexp"
	"strings"

	"github.com/jlubawy/go-ctext"
	"github.com/jlubawy/go-ctext/internal"
)

// A MacroFunc is an invocation of a function-like macro within C source code.
type MacroFunc struct {
	// Name is the name of the macro function.
	Name string

	// Args is a slice of strings that are the arguments for a given function.
	Args []string

	// LineStart is the line that the macro invocation starts on.
	LineStart uint32

	// LineEnd is the line that the macro invocation end on.
	LineEnd uint32
}

func isMacroDef(s string, ni int) (isDef bool) {
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

// FindMacroFuncs finds all invocations of the macro functions matching
// the provided names in a given string.
func FindMacroFuncs(tok *ctext.Token, names ...string) (mfs []MacroFunc, err error) {
	var re *regexp.Regexp
	re, err = compileNamesRegexp(names...)
	if err != nil {
		return
	}

	return FindMacroFuncsRegexp(tok, re)
}

// FindMacroFuncsRegexp finds all invocations of the macro functions matching
// the provided regexp in a given string.
func FindMacroFuncsRegexp(tok *ctext.Token, re *regexp.Regexp) (mfs []MacroFunc, err error) {
	mfs = make([]MacroFunc, 0)

	var (
		s        = tok.Data
		lineCurr = tok.LineStart
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
		isDef := isMacroDef(s, ni)

		// Shorten the string length to look after the name
		s = s[ni+len(name):]

		if isDef {
			continue // skip if this was a definition
		}

		mf := MacroFunc{
			Name:      name,
			Args:      make([]string, 0),
			LineStart: lineCurr,
		}

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
			maxBuf          int
			buf             = make([]byte, 0, 4096)
		)

		// Iterate over the rest of the characters
		i := opi + 1
		for ; (i < len(s)) && !done; i++ {
			b := s[i]

			switch b {
			case ' ':
				if inStringLiteral || parenCount > 0 {
					// If in a string literal then add the space
					buf, err = internal.AddChar(&buf, maxBuf, b)
					if err != nil {
						return
					}
				} else if parenCount == 0 {
					// Else it's probably the end of an argument
					arg, ok := parseArg(&buf)
					if ok {
						mf.Args = append(mf.Args, strings.TrimSpace(string(arg)))
					}
				}

			case ',':
				if inStringLiteral || parenCount > 0 {
					// If in a string literal add the comma
					buf, err = internal.AddChar(&buf, maxBuf, b)
					if err != nil {
						return
					}
				} else if parenCount == 0 {
					// Else it's probably the end of an argument
					arg, ok := parseArg(&buf)
					if ok {
						mf.Args = append(mf.Args, strings.TrimSpace(string(arg)))
					}
				}

			case '"':
				if inStringLiteral {
					lc, ok := internal.LastChar(buf)
					if ok && lc == '\\' {
						// If in a string literal, but this quote was escaped
						// then add it to the buffer
						buf, err = internal.AddChar(&buf, maxBuf, b)
						if err != nil {
							return
						}
					} else {
						// Else leaving a string literal, which has to be the end
						// of an argument
						inStringLiteral = false
						buf, err = internal.AddChar(&buf, maxBuf, b)
						if err != nil {
							return
						}
						if parenCount == 0 {
							arg, ok := parseArg(&buf)
							if ok {
								mf.Args = append(mf.Args, strings.TrimSpace(string(arg)))
							}
						}
					}
				} else {
					// Else not in a string literal, so we are now
					inStringLiteral = true
					buf, err = internal.AddChar(&buf, maxBuf, b)
					if err != nil {
						return
					}
				}

			case '(':
				// Add any opening paren
				buf, err = internal.AddChar(&buf, maxBuf, b)
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
					buf, err = internal.AddChar(&buf, maxBuf, b)
					if err != nil {
						return
					}
				} else {
					if parenCount > 0 {
						// If inside another invocation add the closing paren
						buf, err = internal.AddChar(&buf, maxBuf, b)
						if err != nil {
							return
						}

						// Only decrement if > 0
						parenCount -= 1
					}

					if parenCount == 0 {
						arg, ok := parseArg(&buf)
						if ok {
							mf.Args = append(mf.Args, strings.TrimSpace(string(arg)))
						}
					}
				}

			case ';':
				if inStringLiteral {
					// If in a string literal add the semi-colon
					buf, err = internal.AddChar(&buf, maxBuf, b)
					if err != nil {
						return
					}
				} else {
					// Else if not in a string literal, close out the invocation
					// and find the next one.
					mf.LineEnd = lineCurr
					mfs = append(mfs, mf)
					done = true
				}

			case '\r':
				// discard carriage returns, wait for newline

			case '\n':
				lineCurr += 1

			default:
				buf, err = internal.AddChar(&buf, maxBuf, b)
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

// parseArg returns an argument string and shortens the buffer length to zero if
// the provided buffer isn't empty.
func parseArg(bufP *[]byte) (buf []byte, ok bool) {
	buf = *bufP
	buf = []byte(strings.TrimSpace(string(buf)))
	if len(buf) > 0 {
		ok = true
		*bufP = buf[:0]
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
