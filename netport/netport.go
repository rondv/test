// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package netport parses testdata/netport.yaml for interface assingments.
package netport

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/platinasystems/test"
	"gopkg.in/yaml.v2"
)

const NetPortFile = "testdata/netport.yaml"

var PortByNetPort, NetPortByPort map[string]string

func Init(assert test.Assert) {
	assert.Helper()
	b, err := ioutil.ReadFile(NetPortFile)
	assert.Nil(err)
	PortByNetPort = make(map[string]string)
	NetPortByPort = make(map[string]string)
	err = yaml.Unmarshal(b, PortByNetPort)
	assert.Nil(err)
	for netport, port := range PortByNetPort {
		NetPortByPort[port] = netport
		_, err = os.Stat(filepath.Join("/sys/class/net", port))
		if err != nil {
			assert.Fatal(err)
		}
	}
}
