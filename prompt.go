// Copyright © 2015-2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
        "bufio"
        "fmt"
        "os"
        "runtime"
        "strings"
        "sync"
)

const prompt = "Press enter or enter b[reak], c[ontinue], or q[uit]."

var once sync.Once

func Prompt(f *bool, args ...interface{}) {
        if !*f {
                return
        }
        once.Do(func() {
                fmt.Print(prompt)
        })
        if len(args) > 0 {
                fmt.Print(args...)
                fmt.Print("; ")
        }
        buf, _ := bufio.NewReader(os.Stdin).ReadBytes('\n')
        if strings.ContainsAny(string(buf[:]), "q") {
                panic("quitting")
        }
        if strings.ContainsAny(string(buf[:]), "b") {
                runtime.Breakpoint()
        }
        if strings.ContainsAny(string(buf[:]), "c") {
                *f = false
        }
}
