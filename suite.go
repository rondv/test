// Copyright © 2015-2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"io"
	"testing"
)

// A Suite is a named set of tests
type Suite struct {
	Name string
	Tests
}

func (suite Suite) String() string {
	return suite.Name
}

func (suite Suite) Test(t *testing.T) {
	suite.Tests.Test(t)
}

type Tester interface {
	String() string
	Test(*testing.T)
}

type Tests []Tester

func (tests Tests) Test(t *testing.T) {
	t.Helper()
	for _, v := range tests {
		if t.Failed() {
			break
		}
		name := v.String()
		if suite, ok := v.(Suite); ok {
			t.Run(name, suite.Test)
		} else {
			t.Run(name, func(t *testing.T) {
				t.Helper()
				if *DryRun {
					return
				}
				terr(t, step.Prompt(v))
				if t.Skipped() || t.Failed() {
					return
				}
				v.Test(t)
				if t.Failed() {
					terr(t, Pause.Prompt(v, " FAILED"))
				}
			})
		}
	}
}

func terr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		if err == io.EOF {
			t.SkipNow()
		} else {
			t.Fatal(err)
		}
	}
}
