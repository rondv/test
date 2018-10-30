// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"log"
	"os"
)

var cachedLogger *log.Logger

// Log provides a short file style logger to Stdout use in TestMain
func Log() *log.Logger {
	if cachedLogger == nil {
		cachedLogger = log.New(os.Stdout, "    ", log.Lshortfile)
	}
	return cachedLogger
}
