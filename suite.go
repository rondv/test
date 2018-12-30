// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import "testing"

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
	Pause("before tests")
	for _, v := range tests {
		if t.Failed() {
			break
		}
		name := v.String()
		if suite, ok := v.(Suite); ok {
			Pause("before ", name)
			t.Run(name, suite.Test)
			Pause("after ", name)
		} else {
			t.Run(name, func(t *testing.T) {
				t.Helper()
				if !*DryRun {
					Pause("before ", v)
					v.Test(t)
					Pause("after ", v)
				}
			})
		}
	}
	Pause("after tests")
}
