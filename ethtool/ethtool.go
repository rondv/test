// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package ethtool parses testdata/ethtool.yaml and
// testdata/ethtool_priv_flags.yaml; then issues the respective commands.
package ethtool

import (
	"io/ioutil"
	"time"

	"github.com/platinasystems/test"
	"gopkg.in/yaml.v2"
)

const (
	SettingsFile  = "testdata/ethtool.yaml"
	PrivFlagsFile = "testdata/ethtool_priv_flags.yaml"
	TimeOut       = 2 * time.Second
)

var Settings, PrivFlags map[string][]string

func Init(assert test.Assert) {
	assert.Helper()
	if b, err := ioutil.ReadFile(SettingsFile); err == nil {
		Settings = make(map[string][]string)
		assert.Nil(yaml.Unmarshal(b, Settings))
	}
	if b, err := ioutil.ReadFile(PrivFlagsFile); err == nil {
		PrivFlags = make(map[string][]string)
		assert.Nil(yaml.Unmarshal(b, PrivFlags))
	}
	for k, args := range Settings {
		assert.Program(TimeOut, "ethtool", "-s", k, args)
	}
	for k, args := range PrivFlags {
		assert.Program(TimeOut, "ethtool", "--set-priv-flags", k, args)
	}
}
