// Copyright © 2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netport

var BridgeNets1 = NetDevs{
	{
		Netns:   "h1",
		Vlan:    1,
		NetPort: "net0port0",
		Ifa:     "10.1.0.2/24",
		Routes: []Route{
			{"10.2.0.0/24", "10.1.0.1"},
		},
		Remotes: []string{"10.1.0.1", "10.2.0.1", "10.2.0.2"},
	},

	{
		Netns:   "r",
		IsBridge: true,
		Ifname: "tb1",
		Ifa:     "10.1.0.1/24",
		Remotes: []string{"10.1.0.2", "10.2.0.2"},
	},
	{
		Netns:   "r",
		Vlan:    1,
		NetPort: "net0port1",
		Upper: "tb1",
	},
	{
		Netns:   "r",
		IsBridge: true,
		Ifname: "tb3",
		Ifa:     "10.2.0.1/24",
		Remotes: []string{"10.1.0.2", "10.2.0.2"},
	},
	{
		Netns:   "r",
		Vlan:    2,
		NetPort: "net1port1",
		Upper: "tb3",
	},

	{
		Netns:   "h2",
		Vlan:    2,
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
		Vlan:    10,
		NetPort: "net0port1",
		Ifa:     "10.1.0.2/24",
		Routes: []Route{
			{"10.2.0.0/24", "10.1.0.1"},
		},
		Remotes: []string{"10.1.0.1", "10.2.0.1", "10.2.0.2"},
	},

	// L2 bridge
	{
		Netns:   "b1",
		IsBridge: true,
		Ifname: "tb1",
	},
	{
		Netns:   "b1",
		Vlan:    10,
		NetPort: "net0port0",
		Upper: "tb1",
	},
	{
		Netns:   "b1",
		Vlan:    20,
		NetPort: "net1port0",
		Upper: "tb1",
	},

	// L3 bridge
	{
		Netns:   "r2",
		IsBridge: true,
		Ifname: "tb2",
		Ifa:     "10.1.0.1/24",
		Remotes: []string{"10.1.0.2", "10.2.0.2"},
	},
	{
		Netns:   "r2",
		Vlan:    20,
		NetPort: "net1port1",
		Upper: "tb2",
	},
	{
		Netns:   "r2",
		Vlan:    30,
		NetPort: "net2port0",
		Ifa:     "10.2.0.1/24",
		Remotes: []string{"10.1.0.2", "10.2.0.2"},
	},

	{
		Netns:   "h3",
		Vlan:    30,
		NetPort: "net2port1",
		Ifa:     "10.2.0.2/24",
		Routes: []Route{
			{"10.1.0.0/24", "10.2.0.1"},
		},
		Remotes: []string{"10.2.0.1", "10.1.0.2", "10.1.0.20"},
	},
}



