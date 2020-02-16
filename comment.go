// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import "testing"

// Log args if -test.vv
func Comment(t *testing.T, args ...interface{}) {
	t.Helper()
	if *VV {
		t.Log(args...)
	}
}

// Format args if -test.vv
func Commentf(t *testing.T, format string, args ...interface{}) {
	t.Helper()
	if *VV {
		t.Logf(format, args...)
	}
}
