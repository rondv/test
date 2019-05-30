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

func TestPrompt(t *testing.T) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	promptIn = in
	promptOut = out
	expect := func(want string) {
		t.Helper()
		got := out.String()
		if want != got {
			t.Fatalf("want: %q\ngot: %q", want, got)
		}
		t.Logf("OK, got: %q", got)
	}
	for _, p := range []*prompt{&step, &Pause} {
		t.Run(p.String(), func(t *testing.T) {
			for i, f := range []func(*testing.T){
				func(t *testing.T) {
					const t0 = "test-0"
					in.Reset()
					out.Reset()
					p.reset()
					if err := p.Prompt(t0); err != nil {
						t.Fatal(err)
					}
					expect("")
				},
				func(t *testing.T) {
					const t1 = "test-1"
					in.Reset()
					out.Reset()
					p.set()
					fmt.Fprintln(in)
					if err := p.Prompt(t1); err != nil {
						t.Fatal(err)
					}
					if !p.Flag() {
						t.Fatal("not again")
					}
					expect(t1 + "; " + p.s + promptSuffix)
				},
				func(t *testing.T) {
					const t2 = "test-2"
					in.Reset()
					out.Reset()
					p.set()
					fmt.Fprintln(in, "no")
					if err := p.Prompt(t2); err != nil {
						t.Fatal(err)
					}
					if p.Flag() {
						t.Fatal("not continued")
					}
					expect(t2 + "; " + p.s + promptSuffix)
				},
				func(t *testing.T) {
					const t3 = "test-3"
					in.Reset()
					out.Reset()
					p.set()
					if err := p.Prompt(t3); err != io.EOF {
						t.Fatalf("quit, %v", err)
					}
					if p.Flag() {
						t.Fatal("not continued")
					}
					expect(t3 + "; " + p.s + promptSuffix)
				},
				func(t *testing.T) {
					const t4 = "test-4"
					in.Reset()
					out.Reset()
					p.set()
					fmt.Fprintln(in, "quit")
					if err := p.Prompt(t4); err != io.EOF {
						t.Fatalf("quit, %v", err)
					}
					expect(t4 + "; " + p.s + promptSuffix)
				},
			} {
				t.Run(fmt.Sprint(i), f)
			}
		})
	}
}
