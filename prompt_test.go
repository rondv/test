// Copyright © 2015-2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func expect(t *testing.T, want, got string) {
	t.Helper()
	if want != got {
		t.Fatalf("want: %q\ngot: %q", want, got)
	}
	t.Logf("OK, got: %q", got)
}

func TestStep(t *testing.T) {
	in := new(bytes.Buffer)
	got := new(bytes.Buffer)
	promptIn = in
	promptOut = got
	step.set()
	fmt.Fprintln(in)
	if err := step.Prompt("step-1"); err != nil {
		t.Fatal(err)
	}
	if !step.Flag() {
		t.Fatal("not again")
	}
	expect(t, "step-1; step"+promptSuffix, got.String())
	got.Reset()
	fmt.Fprintln(in, "no")
	if err := step.Prompt("step-2"); err != nil {
		t.Fatal(err)
	}
	if step.Flag() {
		t.Fatal("not continued")
	}
	expect(t, "step-2; step"+promptSuffix, got.String())
	got.Reset()
	step.set()
	in.Reset()
	if err := step.Prompt("step-3"); err != io.EOF {
		t.Fatalf("failed to quit, %v", err)
	}
	if step.Flag() {
		t.Fatal("not continued")
	}
	expect(t, "step-3; step"+promptSuffix, got.String())
	got.Reset()
}

func TestPause(t *testing.T) {
	in := new(bytes.Buffer)
	got := new(bytes.Buffer)
	promptIn = in
	promptOut = got
	Pause.set()
	fmt.Fprintln(in)
	if err := Pause.Prompt("pause-1"); err != nil {
		t.Fatal(err)
	}
	if !Pause.Flag() {
		t.Fatal("not again")
	}
	expect(t, "pause-1; pause"+promptSuffix, got.String())
	got.Reset()
	fmt.Fprintln(in, "no")
	if err := Pause.Prompt("pause-2"); err != nil {
		t.Fatal(err)
	}
	if Pause.Flag() {
		t.Fatal("not continued")
	}
	expect(t, "pause-2; pause"+promptSuffix, got.String())
	got.Reset()
	Pause.set()
	in.Reset()
	if err := Pause.Prompt("pause-3"); err != io.EOF {
		t.Fatalf("failed to quit, %v", err)
	}
	if Pause.Flag() {
		t.Fatal("not continued")
	}
	expect(t, "pause-3; pause"+promptSuffix, got.String())
	got.Reset()
}
