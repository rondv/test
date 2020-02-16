// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

type prompt struct {
	flag *bool
	s    string
}

const promptSuffix = " again? [(y)/n/q/EOF] "

var (
	promptIn  = io.Reader(os.Stdin)
	promptOut = io.Writer(os.Stdout)
	Pause     = prompt{
		flag.Bool("test.pause", false, "enable program pauses"),
		"pause",
	}
	step = prompt{
		flag.Bool("test.step", false, "pause between each test"),
		"step",
	}
)

func (p *prompt) set() {
	*(p.flag) = true
}

func (p *prompt) reset() {
	*(p.flag) = false
}

func (p *prompt) Flag() bool {
	return *(p.flag)
}

func (p *prompt) String() string {
	return p.s
}

// Prompt returns io.EOF to skip all remaining tests
func (p *prompt) Prompt(args ...interface{}) error {
	if !p.Flag() {
		return nil
	}
	if len(args) > 0 {
		fmt.Fprint(promptOut, args...)
		fmt.Fprint(promptOut, "; ", p, promptSuffix)
	}
	buf, err := bufio.NewReader(promptIn).ReadBytes('\n')
	if err != nil {
		p.reset()
		return err
	}
	switch string(buf) {
	case "n\n", "no\n":
		p.reset()
	case "q\n", "quit\n":
		return io.EOF
	case "\n", "y\n", "yes\n":
	default:
		fmt.Fprintf(promptOut, "%q ignored", buf)
	}
	return nil
}
