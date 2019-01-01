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

type DummyIf struct {
	Ifname string
	Ifa    string
}

const (
	NETPORT_DEVTYPE_PORT = iota
	NETPORT_DEVTYPE_PORT_VLAN
	NETPORT_DEVTYPE_BRIDGE
	NETPORT_DEVTYPE_BRIDGE_PORT
)

// NetDev describes a network interface configuration.
// DevType is determined by IsBridge, Vlan, and Upper
// default PORT sets ifa for ifname derived from NetPort 
// BRIDGE adds a linux bridge device with ifname and ifa
// PORT_VLAN adds a linux vlan device to ifname derived from NetPort with ifa
// BRIDGE_PORT adds a linux vlan device and sets upper to named bridge (no ifa)
type NetDev struct {
	DevType  int
	IsBridge bool
	Vlan     int // for PORT_VLAN or BRIDGE_PORT
	NetPort  string
	Netns    string
	Ifname   string
	Upper    string // only for BRIDGE_PORT

	Ifa      string // no ifa for BRIDGE_PORT
	DummyIfs []DummyIf
	Routes   []Route
	Remotes  []string
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
		nd.DevType = NETPORT_DEVTYPE_PORT
		if nd.IsBridge {
			nd.DevType = NETPORT_DEVTYPE_BRIDGE
		} else if nd.Vlan != 0 {
			if nd.Upper != "" {
				nd.DevType = NETPORT_DEVTYPE_BRIDGE_PORT
			} else {
				nd.DevType = NETPORT_DEVTYPE_PORT_VLAN
			}
		}
		if *test.VVV {
			t.Logf("nd %+v\n", nd)
		}
		if nd.DevType == NETPORT_DEVTYPE_BRIDGE || nd.DevType == NETPORT_DEVTYPE_BRIDGE_PORT {
			t.Log("bridge ignored")
			continue
		}

		ns := nd.Netns
		_, err := os.Stat(filepath.Join("/var/run/netns", ns))
		if err != nil {
			assert.Program("ip", "netns", "add", ns)
			defer cleanup.Program("ip", "netns", "del", ns)
		}

		ifname := PortByNetPort[nd.NetPort]
		if nd.DevType == NETPORT_DEVTYPE_PORT_VLAN ||
			nd.DevType == NETPORT_DEVTYPE_BRIDGE_PORT {
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

		if nd.DevType != NETPORT_DEVTYPE_BRIDGE_PORT {
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
	}
	test.Tests(tests).Test(t)
}
