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

	"github.com/platinasystems/goes/external/xeth"
	"github.com/platinasystems/test"
	"gopkg.in/yaml.v2"
)

const NetPortFile = "testdata/netport.yaml"

var Goes string
var PortByNetPort, NetPortByPort map[string]string

var DevKindOf = map[string]xeth.DevKind{
	"port":   xeth.DevKindPort,
	"bridge": xeth.DevKindBridge,
}

func Init(goes string) {
	Goes = goes
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

type PortVlan struct {
	NetPort string
	Vlan    int
	Ifname  string // runtime lookup
}

// NetDev describes a network interface configuration.
// default PORT sets ifa for ifname derived from NetPort
// BRIDGE adds a linux bridge device with ifname and ifa, plus its vlan members
// PORT_VLAN adds a linux vlan device to ifname derived from NetPort with ifa
// initial values filled from NetDev[], then DevKind and Ifname updated
type NetDev struct {
	Kind string

	NetPort string // lookup key for NetPortFile to Ifname
	Vlan    int    // for PORT_VLAN
	Ifname  string

	Netns    string
	Ifa      string
	DummyIfs []DummyIf
	Routes   []Route
	Remotes  []string
	Lowers   []PortVlan
}

// NetDevs describe all of the interfaces in the virtual network under test.
type NetDevs []NetDev

// add/del netport in netns, either as dataport or member of bridge
func NetPortConfig(t *testing.T, ns string, netport string, vlan int) (ifname string) {
	assert := test.Assert{t}

	ifname = PortByNetPort[netport]
	if vlan != 0 {
		link := ifname
		ifname += fmt.Sprint(".", vlan)
		assert.Program(Goes, "ip", "link", "set", link,
			"up")
		assert.Program(Goes, "ip", "link", "add",
			ifname, "link", link, "type", "xeth-vlan")
	}
	assert.ProgramRetry(3, Goes, "ip", "link", "set",
		ifname, "up", "netns", ns)
	return
}

func NetPortCleanup(t *testing.T, ns string, ifname string, vlan int) {
	cleanup := test.Cleanup{t}

	cleanup.ProgramRetry(3, Goes, "ip", "-n", ns,
		"link", "set", ifname, "down", "netns", 1)

	if vlan != 0 {
		cleanup.Program(Goes, "ip", "link", "del", ifname)
	} else {
		cleanup.ProgramRetry(3, Goes,
			"ip", "link", "set", ifname, "up")
	}
}

// netdevs list the interface configurations of the network under test
func (netdevs NetDevs) Test(t *testing.T, tests ...test.Tester) {
	var brlink string

	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	cleanup := test.Cleanup{t}
	for i := range netdevs {
		nd := &netdevs[i]
		kind, ok := DevKindOf[nd.Kind]
		if ok != true {
			kind = xeth.DevKindPort
		}

		ns := nd.Netns
		_, err := os.Stat(filepath.Join("/var/run/netns", ns))
		if err != nil {
			assert.Program(Goes, "ip", "netns", "add", ns)
			defer cleanup.Program(Goes, "ip", "netns", "del", ns)
		}

		if kind == xeth.DevKindBridge {
			brlink = ""
			for i, mbr := range nd.Lowers {
				nd.Lowers[i].Ifname = NetPortConfig(t, ns, mbr.NetPort, mbr.Vlan)
				if brlink == "" {
					brlink = nd.Lowers[i].Ifname
				}
			}
			if brlink != "" {
				assert.ProgramRetry(3, Goes, "ip", "netns", "exec", ns, Goes, "ip",
					"link", "add", "name", nd.Ifname,
					"link", brlink,
					"type", "xeth-bridge")
			} else {
				assert.ProgramRetry(3, Goes, "ip", "netns", "exec", ns, Goes, "ip",
					"link", "add", "name", nd.Ifname,
					"type", "xeth-bridge")
			}
			/*
				assert.ProgramRetry(3, Goes, "ip", "-n", ns,
					"link", "add", nd.Ifname,
					brlink,
					"type", "xeth-bridge")
			*/
			assert.ProgramRetry(3, Goes, "ip", "-n", ns,
				"link", "set", nd.Ifname, "up")
			defer cleanup.Program(Goes, "ip", "-n", ns,
				"link", "del", nd.Ifname)

			for _, mbr := range nd.Lowers {
				assert.ProgramRetry(3, GoesIP, "ip", "-n", ns,
					"link", "set", mbr.Ifname, "master", nd.Ifname)
				defer NetPortCleanup(t, ns, mbr.Ifname, mbr.Vlan)
				defer cleanup.ProgramRetry(3, Goes, "ip", "-n", ns,
					"link", "set", mbr.Ifname, "nomaster")
			}
		} else {
			nd.Ifname = NetPortConfig(t, ns, nd.NetPort, nd.Vlan)
			defer NetPortCleanup(t, ns, nd.Ifname, nd.Vlan)
		}

		if nd.Ifa != "" {
			/* ip commands like "ip route" require specific family
			 * (-6) for routes
			 */
			family := test.IpFamily(nd.Ifa)
			assert.ProgramRetry(3, Goes, "ip", "-n", ns, family,
				"address", "add", nd.Ifa, "dev", nd.Ifname)
			defer cleanup.Program(Goes, "ip", "-n", ns, family,
				"address", "del", nd.Ifa, "dev", nd.Ifname)
			for _, route := range nd.Routes {
				prefix := route.Prefix
				gw := route.GW
				assert.ProgramRetry(3, Goes, "ip", "-n", ns, family,
					"route", "add", prefix, "via", gw)
			}
		}
		if *test.VVV {
			t.Logf("nd %+v\n", nd)
		}
	}
	test.Tests(tests).Test(t)
}
