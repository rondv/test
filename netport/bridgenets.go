// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netport

// starting tag for vlan interfaces is 100 instead of 1
// allocated stag for bridge must not overlap with tag of configured vlan interface
// otherwise vlan ingress is switched as bridge ingress

// 32b ifindex for bridge will collidge across netns unless allocated manually
// use ridiculously huge number to avoid collision with default namespace
const BridgeIndexBase = 2000000000

var BridgeNets1 = NetDevs{
	{
		Netns:   "h1",
		Vlan:    100,
		NetPort: "net0port0",
		Ifa:     "10.1.0.2/24",
		Routes: []Route{
			{"10.2.0.0/24", "10.1.0.1"},
		},
		Remotes: []string{"10.1.0.1", "10.2.0.1", "10.2.0.2"},
	},

	{
		Netns:         "r",
		IsBridge:      true,
		BridgeIfindex: BridgeIndexBase + 0,
		Ifname:        "tb1",
		BridgeMac:     "00:00:01:b1:b1:b1",
		Ifa:           "10.1.0.1/24",
		Remotes:       []string{"10.1.0.2", "10.2.0.2"},
	},
	{
		Netns:   "r",
		Vlan:    100,
		NetPort: "net0port1",
		Upper:   "tb1",
	},
	{
		Netns:         "r",
		IsBridge:      true,
		BridgeIfindex: BridgeIndexBase + 1,
		Ifname:        "tb3",
		BridgeMac:     "00:00:01:b3:b3:b3",
		Ifa:           "10.2.0.1/24",
		Remotes:       []string{"10.1.0.2", "10.2.0.2"},
	},
	{
		Netns:   "r",
		Vlan:    200,
		NetPort: "net1port1",
		Upper:   "tb3",
	},

	{
		Netns:   "h2",
		Vlan:    200,
		NetPort: "net1port0",
		Ifa:     "10.2.0.2/24",
		Routes: []Route{
			{"10.1.0.0/24", "10.2.0.1"},
		},
		Remotes: []string{"10.2.0.1", "10.1.0.1", "10.1.0.2"},
	},
}

var BridgeNets2 = NetDevs{
	{
		Netns:   "h1",
		Vlan:    100,
		NetPort: "net0port1",
		Ifa:     "10.1.0.2/24",
		Routes: []Route{
			{"10.2.0.0/24", "10.1.0.1"},
		},
		Remotes: []string{"10.1.0.1", "10.2.0.1", "10.2.0.2"},
	},

	// L2 bridge
	{
		Netns:         "b1",
		IsBridge:      true,
		BridgeIfindex: BridgeIndexBase + 0,
		Ifname:        "tb1",
		BridgeMac:     "00:00:02:b1:b1:b1",
		Ifa:           "10.1.0.20/24",
		Routes: []Route{
			{"default", "10.1.0.1"},
		},
	},
	{
		Netns:   "b1",
		Vlan:    100,
		NetPort: "net0port0",
		Upper:   "tb1",
	},
	{
		Netns:   "b1",
		Vlan:    200,
		NetPort: "net1port0",
		Upper:   "tb1",
	},

	// L3 bridge
	{
		Netns:         "r2",
		IsBridge:      true,
		BridgeIfindex: BridgeIndexBase + 1,
		Ifname:        "tb2",
		BridgeMac:     "00:00:02:b2:b2:b2",
		Ifa:           "10.1.0.1/24",
		Remotes:       []string{"10.1.0.2", "10.2.0.2"},
	},
	{
		Netns:   "r2",
		Vlan:    200,
		NetPort: "net1port1",
		Upper:   "tb2",
	},
	{
		Netns:   "r2",
		Vlan:    300,
		NetPort: "net2port0",
		Ifa:     "10.2.0.1/24",
		Remotes: []string{"10.1.0.2", "10.2.0.2"},
	},

	{
		Netns:   "h3",
		Vlan:    300,
		NetPort: "net2port1",
		Ifa:     "10.2.0.2/24",
		Routes: []Route{
			{"10.1.0.0/24", "10.2.0.1"},
		},
		Remotes: []string{"10.2.0.1", "10.1.0.2", "10.1.0.20"},
	},
}
