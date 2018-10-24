// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import "flag"

var (
	Goes = flag.String("test.goes", "./goes",
		"GO Embedded System")
	DryRun = flag.Bool("test.dryrun", false,
		"don't run, just print test names")
	MustPause = flag.Bool("test.pause", false,
		"pause before and after suite")
	VV  = flag.Bool("test.vv", false, "log test.Program output")
	VVV = flag.Bool("test.vvv", false, "log test.Program execution")
)
