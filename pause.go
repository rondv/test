// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
)

func Pause(args ...interface{}) {
	if !*MustPause {
		return
	}
	if len(args) > 0 {
		fmt.Print(args...)
		fmt.Print("; ")
	}
	fmt.Print("Press enter to continue, or 'c'=run, 'q'=quit, 'b'=break ...")
	buf, _ := bufio.NewReader(os.Stdin).ReadBytes('\n')
	if strings.ContainsAny(string(buf[:]), "q") {
		panic("quitting")
	}
	if strings.ContainsAny(string(buf[:]), "b") {
		runtime.Breakpoint()
	}
	if strings.ContainsAny(string(buf[:]), "c") {
		*MustPause = false
	}
}
