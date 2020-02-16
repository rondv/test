// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"flag"
	"testing"
)

var (
	DryRun = flag.Bool("test.dryrun", false,
		"don't run, just print test names")
	VV  = flag.Bool("test.vv", false, "log program output")
	VVV = flag.Bool("test.vvv", false, "log program execution")
)

func SkipIfDryRun(t *testing.T) {
	if *DryRun {
		t.SkipNow()
	}
}
