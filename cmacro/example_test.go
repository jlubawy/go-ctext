// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmacro_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/jlubawy/go-ctext/cmacro"
)

const BasicProgram = `#include <stdio.h>

#define PRINTF( fmt, ... )  printf( _fmt, __VA_ARGS__ )

int
main( void )
{
    PRINTF(
        "Hello World: %d, %d, %d, %d",
        1,
        2,
        3,
        4
    );
    return 0;
}
`

func Example() {
	err := cmacro.ScanInvocations(strings.NewReader(BasicProgram), func(inv cmacro.Invocation) {
		fmt.Println(inv)
		fmt.Printf("start=%d, end=%d\n", inv.Start, inv.End)
	}, "PRINTF")
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// PRINTF( "Hello World: %d, %d, %d, %d", 1, 2, 3, 4 );
	// start=8, end=14
}
