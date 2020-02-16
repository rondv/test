// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Timeout is the default duration on the Program Wait timer.
const Timeout = 3 * time.Second

type Quiet struct{}

// Begin a Program; type options:
//
//	Quiet
//		don't log output even if err is !nil
//	io.Reader
//		use reader as Stdin instead of the default, /dev/null
//
//	*regexp.Regexp
//		match Stdout with compiled regex pattern
//
//	time.Duration
//		wait up to the given duration for the program to finish instead
//		of the default Timeout
func Begin(tb testing.TB, options ...interface{}) (*Program, error) {
	var (
		stdin io.Reader
		args  []string
	)
	p := &Program{
		tb:   tb,
		obuf: new(bytes.Buffer),
		ebuf: new(bytes.Buffer),
		dur:  Timeout,
	}
	for _, opt := range options {
		switch t := opt.(type) {
		case Quiet:
			p.quiet = true
		case io.Reader:
			stdin = t
		case *regexp.Regexp:
			p.exp = t
		case string:
			args = append(args, t)
		case []string:
			args = append(args, t...)
		case time.Duration:
			p.dur = t
		default:
			args = append(args, fmt.Sprint(t))
		}
	}
	if len(args) == 0 {
		return p, errors.New("missing command args")
	}
	// preface output with newline for pretty logging
	p.obuf.WriteRune('\n')
	p.cmd = exec.Command(args[0], args[1:]...)
	p.cmd.Stdin = stdin
	p.cmd.Stdout = p.obuf
	p.cmd.Stderr = p.ebuf
	if *VVV {
		tb.Helper()
		tb.Log(args)
	}
	return p, p.cmd.Start()
}

// Program is an exec.Cmd wrapper
type Program struct {
	cmd   *exec.Cmd
	tb    testing.TB
	obuf  *bytes.Buffer
	ebuf  *bytes.Buffer
	dur   time.Duration
	exp   *regexp.Regexp
	quiet bool
}

// Quit will SIGTERM the Program then End and Log any error.
func (p *Program) Quit() {
	p.tb.Helper()
	p.cmd.Process.Signal(syscall.SIGTERM)
	if err := p.End(); err != nil {
		p.tb.Log(err)
	}
}

// End will wait for Program to finish or timeout then match and log output.
func (p *Program) End() (err error) {
	p.tb.Helper()
	tm := time.NewTimer(p.dur)
	done := make(chan error)
	sig := syscall.SIGTERM
	go func() { done <- p.cmd.Wait() }()
again:
	select {
	case err = <-done:
		tm.Stop()
		if s := strings.TrimSpace(p.ebuf.String()); len(s) > 0 {
			err = errors.New(p.ebuf.String())
			p.ebuf.Reset()
		}
		if err == nil && p.exp != nil && !p.exp.Match(p.obuf.Bytes()) {
			err = fmt.Errorf("mismatch %q", p.exp)
		}
	case <-tm.C:
		err = syscall.ETIME
		if *VV || !p.quiet {
			p.tb.Log(sig, "process", p.cmd.Process.Pid, p.cmd.Args)
		}
		p.cmd.Process.Signal(sig)
		tm.Reset(3 * time.Second)
		sig = syscall.SIGKILL
		goto again
	}
	if !p.quiet && (*VV || err != nil) {
		s := strings.TrimRight(p.obuf.String(), "\n")
		if len(s) > 0 {
			p.tb.Log(s)
		}
	}
	p.obuf.Reset()
	return
}

// Pid returns the program process identifier.
func (p *Program) Pid() int {
	return p.cmd.Process.Pid
}

// A Daemon is a background program started and defer stopped from a TestMain.
type Daemon struct {
	cmd    *exec.Cmd
	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

func (d *Daemon) Pid() int {
	return d.cmd.Process.Pid
}

// Start the daemon program and panic if error.
func (d *Daemon) Start(args ...string) {
	d.cmd = exec.Command(args[0], args[1:]...)
	d.stdout = new(bytes.Buffer)
	d.stderr = new(bytes.Buffer)
	d.cmd.Stdout = d.stdout
	d.cmd.Stderr = d.stderr
	if *VVV {
		Log().Output(2, strings.TrimSpace(fmt.Sprintln(args)))
	}
	if err := d.cmd.Start(); err != nil {
		panic(fmt.Errorf("%v: %v", args, err))
	}
}

// Stop the running daemon with a TERM, INT, then KILL signal
func (d *Daemon) Stop() {
	var err error
	done := make(chan error)
	sig := syscall.SIGTERM
	timeout := 3 * time.Second
	tm := time.NewTimer(timeout)
	d.cmd.Process.Signal(sig)
	go func() { done <- d.cmd.Wait() }()
again:
	select {
	case err = <-done:
	case <-tm.C:
		switch sig {
		case syscall.SIGKILL:
			Log().Output(2, "won't die!")
		case syscall.SIGINT:
			err = syscall.ETIME
			sig = syscall.SIGKILL
			timeout *= 2
			d.cmd.Process.Signal(sig)
			tm.Reset(timeout)
			goto again
		case syscall.SIGTERM:
			sig = syscall.SIGINT
			timeout *= 2
			d.cmd.Process.Signal(sig)
			tm.Reset(timeout)
			goto again
		}
	}
	tm.Stop()
	if err != nil {
		s := err.Error()
		for _, b := range []*bytes.Buffer{
			d.stdout,
			d.stderr,
		} {
			if b.Len() > 0 {
				s += "\n"
				s += b.String()
			}
		}
		Log().Output(2, strings.TrimSpace(s))
	} else if *VV && d.stdout.Len() > 0 {
		Log().Output(2, d.stdout.String())
	}
}

// Run a program - usually from TestMain - and panic if error.
func Run(args ...string) {
	if *VVV {
		Log().Output(2, strings.TrimSpace(fmt.Sprintln(args)))
	}
	output, err := exec.Command(args[0], args[1:]...).Output()
	if *VV && len(output) > 0 {
		Log().Output(2, string(output))
	}
	if err != nil {
		if *VVV {
			panic(err)
		}
		panic(fmt.Errorf("%v:\n\t%v", args, err))
	}
}
