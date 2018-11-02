// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package netport parses testdata/netport.yaml for interface assingments along
// with utilities to build and test virtual networks configured from these
// assignments.
package netport

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/platinasystems/test"
	"gopkg.in/yaml.v2"
)

const NetPortFile = "testdata/netport.yaml"

var PortByNetPort, NetPortByPort map[string]string

func Init() {
	b, err := ioutil.ReadFile(NetPortFile)
	if err != nil {
		panic(err)
	}
	PortByNetPort = make(map[string]string)
	NetPortByPort = make(map[string]string)
	if err = yaml.Unmarshal(b, PortByNetPort); err != nil {
		panic(fmt.Errorf("%s: %v", NetPortFile, err))
	}
	for netport, port := range PortByNetPort {
		sysport := filepath.Join("/sys/class/net", port)
		if _, err = os.Stat(sysport); err != nil {
			panic(err)
		}
		NetPortByPort[port] = netport
	}
}

type Route struct {
	Prefix string
	GW     string
}

// NetDev describes a network interface configuration.  A Vlan > 0 adds a linux
// vlan device to the named NetPort with the given address; otherwise, the
// configuration applies to the referenced NetPort.
type NetDev struct {
	Vlan    int
	NetPort string
	Netns   string
	Ifname  string
	Ifa     string
	Routes  []Route
	Remotes []string
}

// NetDevs describe all of the interfaces in the virtual network under test.
type NetDevs []NetDev

// netdevs list the interface configurations of the network under test
func (netdevs NetDevs) Test(t *testing.T, tests ...test.Tester) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	cleanup := test.Cleanup{t}
	for i := range netdevs {
		nd := &netdevs[i]
		ns := nd.Netns
		_, err := os.Stat(filepath.Join("/var/run/netns", ns))
		if err != nil {
			assert.Program("ip", "netns", "add", ns)
			defer cleanup.Program("ip", "netns", "del", ns)
		}
		ifname := PortByNetPort[nd.NetPort]
		if nd.Vlan > 0 {
			link := ifname
			ifname += fmt.Sprint(".", nd.Vlan)
			assert.Program("ip", "link", "set", link, "up")
			assert.Program("ip", "link", "add", ifname,
				"link", link, "type", "vlan",
				"id", nd.Vlan)
			defer cleanup.Program("ip", "link", "del",
				ifname)
		}
		nd.Ifname = ifname
		assert.Program("ip", "link", "set", ifname, "up",
			"netns", ns)
		defer cleanup.Program("ip", "netns", "exec", ns,
			"ip", "link", "set", ifname, "down",
			"netns", 1)
		assert.Program("ip", "netns", "exec", ns,
			"ip", "address", "add", nd.Ifa,
			"dev", ifname)
		defer cleanup.Program("ip", "netns", "exec", ns,
			"ip", "address", "del", nd.Ifa,
			"dev", ifname)
		for _, route := range nd.Routes {
			prefix := route.Prefix
			gw := route.GW
			assert.Program("ip", "netns", "exec", ns,
				"ip", "route", "add", prefix,
				"via", gw)
		}
	}
	test.Tests(tests).Test(t)
}
