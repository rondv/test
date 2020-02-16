// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package ethtool parses testdata/ethtool.yaml and
// testdata/ethtool_priv_flags.yaml; then issues the respective commands.
package ethtool

import (
	"fmt"
	"io/ioutil"

	"github.com/platinasystems/test"
	"gopkg.in/yaml.v2"
)

const (
	SettingsFile  = "testdata/ethtool.yaml"
	PrivFlagsFile = "testdata/ethtool_priv_flags.yaml"
)

var Settings, PrivFlags map[string][]string

func Init() {
	if b, err := ioutil.ReadFile(SettingsFile); err == nil {
		Settings = make(map[string][]string)
		if err = yaml.Unmarshal(b, Settings); err != nil {
			panic(fmt.Errorf("%s: %v", SettingsFile, err))
		}
	}
	if b, err := ioutil.ReadFile(PrivFlagsFile); err == nil {
		PrivFlags = make(map[string][]string)
		if err = yaml.Unmarshal(b, PrivFlags); err != nil {
			panic(fmt.Errorf("%s: %v", PrivFlagsFile, err))
		}
	}
	for ifname, args := range Settings {
		ethtool := []string{"ethtool", "-s", ifname}
		test.Run(append(ethtool, args...)...)
	}
	option := "--set-priv-flags"
	if opt, ok := PrivFlags["option"]; ok && len(opt) > 0 {
		option = opt[0]
	}
	for ifname, args := range PrivFlags {
		if ifname == "option" {
			continue
		}
		ethtool := []string{"ethtool", option, ifname}
		test.Run(append(ethtool, args...)...)
	}
}
