// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package docker

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/platinasystems/test"
	"github.com/platinasystems/test/netport"
)

type Docket struct {
	Tmpl string
	*Config
}

func (d *Docket) ExecCmd(t *testing.T, ID string,
	cmd ...string) (string, error) {
	t.Helper()
	return ExecCmd(t, ID, d.Config, cmd)
}

func (d *Docket) PingCmd(t *testing.T, ID string, target string) error {
	return PingCmd(t, ID, d.Config, target)
}

// Docket.Test, unlike netport.NetDevs.Test, doesn't descend given tests during
// dryruns.
func (d *Docket) Test(t *testing.T, tests ...test.Tester) {
	if *test.DryRun {
		t.SkipNow()
	}
	if err := Check(t); err != nil {
		t.Skip(err)
	}
	assert := test.Assert{t}
	assert.Helper()
	text, err := ioutil.ReadFile(d.Tmpl)
	assert.Nil(err)
	name := strings.TrimSuffix(d.Tmpl, ".tmpl")
	tmpl, err := template.New(name).Parse(string(text))
	assert.Nil(err)
	buf := new(bytes.Buffer)
	assert.Nil(tmpl.Execute(buf, netport.PortByNetPort))
	d.Config, err = LaunchContainers(t, buf.Bytes())
	assert.Nil(err)
	defer TearDownContainers(t, d.Config)
	test.Tests(tests).Test(t)
}
