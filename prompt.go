// Copyright © 2015-2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
)

const (
	promptSuffix = "; continue? [y/(n)] "
)

var (
	promptOnce    sync.Once
	promptIn      = io.Reader(os.Stdin)
	promptOut     = io.Writer(os.Stdout)
	promptHeading = `
At following continue prompts, press EOF to quit;
enter y[es] to continue remaining tests w/o pause;
or just enter to proceed to next step.`[1:]
)

// Prompt returns io.EOF to skip all remaining tests
func prompt(args ...interface{}) error {
	if !*MustPause && !*MustStep {
		return nil
	}
	promptOnce.Do(func() {
		fmt.Fprintln(promptOut, promptHeading)
	})
	if len(args) > 0 {
		fmt.Fprint(promptOut, args...)
		fmt.Fprint(promptOut, promptSuffix)
	}
	buf, err := bufio.NewReader(promptIn).ReadBytes('\n')
	if err != nil {
		*MustPause = false
		*MustStep = false
		return err
	}
	switch string(buf) {
	case "y\n", "yes\n":
		if *MustStep {
			*MustStep = false
		} else {
			*MustPause = false
		}
	case "q\n", "quit\n":
		return io.EOF
	case "\n", "n\n", "no\n":
	default:
		fmt.Fprintf(promptOut, "%q ignored", buf)
	}
	return nil
}
