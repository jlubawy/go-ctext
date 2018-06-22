// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ctext

import (
	"encoding/json"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestScanner(t *testing.T) {
	cF, err := os.Open("testdata/hello_world.c")
	if err != nil {
		t.Fatalf("could not open C file: %v", err)
	}
	defer cF.Close()

	jsonF, err := os.Open("testdata/hello_world.json")
	if err != nil {
		t.Fatalf("could not open JSON file: %v", err)
	}
	defer jsonF.Close()

	var expTokens []Token
	if err := json.NewDecoder(jsonF).Decode(&expTokens); err != nil {
		t.Fatalf("error decoding JSON file: %v", err)
	}

	tokens := make([]Token, 0)
	s := NewScanner(cF)
	s.Filename = "hello_world.c"
	for {
		tt := s.Next()

		switch tt {
		case ErrorToken:
			err := s.Err()
			if err == io.EOF {
				goto DONE
			}
			t.Fatalf("Error scanning file: %v\n", err)

		case CommentToken:
			tokens = append(tokens, s.Token())

		case TextToken:
			tokens = append(tokens, s.Token())
		}
	}

DONE:
	if !reflect.DeepEqual(tokens, expTokens) {
		t.Error("tokens do not match")
		t.Logf("%+v", expTokens)
		t.Logf("%+v", tokens)
	}
}
