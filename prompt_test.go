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

func TestPrompt(t *testing.T) {
	in := new(bytes.Buffer)
	got := new(bytes.Buffer)
	promptIn = in
	promptOut = got
	*MustStep = true
	fmt.Fprintln(in)
	if err := prompt("step-1"); err != nil {
		t.Fatal(err)
	}
	if !*MustStep {
		t.Fatal("misstep")
	}
	expect(t, promptHeading+"\nstep-1"+promptSuffix, got.String())
	got.Reset()
	fmt.Fprintln(in, "yes")
	if err := prompt("step-2"); err != nil {
		t.Fatal(err)
	}
	if *MustStep {
		t.Fatal("not continued")
	}
	expect(t, "step-2"+promptSuffix, got.String())
	got.Reset()
	*MustPause = true
	fmt.Fprintln(in)
	if err := prompt("pause-1"); err != nil {
		t.Fatal(err)
	}
	if !*MustPause {
		t.Fatal("unpaused")
	}
	expect(t, "pause-1"+promptSuffix, got.String())
	got.Reset()
	fmt.Fprintln(in, "yes")
	if err := prompt("pause-2"); err != nil {
		t.Fatal(err)
	}
	if *MustPause {
		t.Fatal("still paused")
	}
	expect(t, "pause-2"+promptSuffix, got.String())
	got.Reset()
	*MustPause = true
	in.Reset()
	if err := prompt("pause-3"); err != io.EOF {
		t.Fatalf("failed to quit, %v", err)
	}
	if *MustPause {
		t.Fatal("still paused")
	}
	if *MustStep {
		t.Fatal("not continued")
	}
	expect(t, "pause-3"+promptSuffix, got.String())
	got.Reset()
}
