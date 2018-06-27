// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Command ctext exposes some of the functions provided by the ctext package.

Run `ctext help` for usage.
*/
package main

import (
	"github.com/jlubawy/go-cli"
)

var program = cli.Program{
	Name: "ctext",
	Commands: []cli.Command{
		stripCommand,
	},
}

func main() { program.RunAndExit() }
