// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netport

var BridgeNets0 = NetDevs{
	{
		Netns:   "h1",
		Vlan:    100,
		NetPort: "net0port0",
		Ifa:     "10.1.0.1/24",
		Remotes: []string{"10.1.0.2", "10.1.0.3"},
	},

	{
		Netns:   "r",
		Kind:    "bridge",
		Ifname:  "tb1",
		Ifa:     "10.1.0.2/24",
		Remotes: []string{"10.1.0.1", "10.1.0.3"},
		Lowers: []PortVlan{
			{NetPort: "net0port1", Vlan: 100},
			{NetPort: "net1port1", Vlan: 0},
		},
	},
	{
		Netns:   "h2",
		NetPort: "net1port0",
		Ifa:     "10.1.0.3/24",
		Remotes: []string{"10.1.0.1", "10.1.0.2"},
	},
}

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
		Netns:   "r",
		Kind:    "bridge",
		Ifname:  "tb1",
		Ifa:     "10.1.0.1/24",
		Remotes: []string{"10.1.0.2", "10.2.0.2"},
		Lowers: []PortVlan{
			{NetPort: "net0port1", Vlan: 100},
		},
	},
	{
		Netns:   "r",
		Kind:    "bridge",
		Ifname:  "tb3",
		Ifa:     "10.2.0.1/24",
		Remotes: []string{"10.1.0.2", "10.2.0.2"},
		Lowers: []PortVlan{
			{NetPort: "net1port1", Vlan: 200},
		},
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

var _BridgeNets2 = NetDevs{ // adjacent bridges, not supported
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
		Netns:  "b1",
		Kind:   "bridge",
		Ifname: "tb1",
		Ifa:    "10.1.0.20/24",
		Routes: []Route{
			{"default", "10.1.0.1"},
		},
		Lowers: []PortVlan{
			{NetPort: "net0port0", Vlan: 100},
			{NetPort: "net1port0", Vlan: 200},
		},
	},

	// L3 bridge
	{
		Netns:   "r2",
		Kind:    "bridge",
		Ifname:  "tb2",
		Ifa:     "10.1.0.1/24",
		Remotes: []string{"10.1.0.2", "10.2.0.2"},
		Lowers: []PortVlan{
			{NetPort: "net1port1", Vlan: 200},
		},
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

var BridgeNets1u = NetDevs{
	{
		Netns:   "h1",
		NetPort: "net0port0",
		Ifa:     "10.1.0.2/24",
		Routes: []Route{
			{"10.2.0.0/24", "10.1.0.1"},
		},
		Remotes: []string{"10.1.0.1", "10.2.0.1", "10.2.0.2"},
	},

	{
		Netns:   "r",
		Kind:    "bridge",
		Ifname:  "tb1",
		Ifa:     "10.1.0.1/24",
		Remotes: []string{"10.1.0.2", "10.2.0.2"},
		Lowers: []PortVlan{
			{NetPort: "net0port1", Vlan: 0},
		},
	},
	{
		Netns:   "r",
		Kind:    "bridge",
		Ifname:  "tb3",
		Ifa:     "10.2.0.1/24",
		Remotes: []string{"10.1.0.2", "10.2.0.2"},
		Lowers: []PortVlan{
			{NetPort: "net1port1", Vlan: 0},
		},
	},

	{
		Netns:   "h2",
		NetPort: "net1port0",
		Ifa:     "10.2.0.2/24",
		Routes: []Route{
			{"10.1.0.0/24", "10.2.0.1"},
		},
		Remotes: []string{"10.2.0.1", "10.1.0.1", "10.1.0.2"},
	},
}

var _BridgeNets2u = NetDevs{ // adjacent bridges, not supported
	{
		Netns:   "h1",
		NetPort: "net0port1",
		Ifa:     "10.1.0.2/24",
		Routes: []Route{
			{"10.2.0.0/24", "10.1.0.1"},
		},
		Remotes: []string{"10.1.0.1", "10.2.0.1", "10.2.0.2"},
	},

	// L2 bridge
	{
		Netns:  "b1",
		Kind:   "bridge",
		Ifname: "tb1",
		Ifa:    "10.1.0.20/24",
		Routes: []Route{
			{"default", "10.1.0.1"},
		},
		Lowers: []PortVlan{
			{NetPort: "net0port0", Vlan: 0},
			{NetPort: "net1port0", Vlan: 0},
		},
	},

	// L3 bridge
	{
		Netns:   "r2",
		Kind:    "bridge",
		Ifname:  "tb2",
		Ifa:     "10.1.0.1/24",
		Remotes: []string{"10.1.0.2", "10.2.0.2"},
		Lowers: []PortVlan{
			{NetPort: "net1port1", Vlan: 0},
		},
	},
	{
		Netns:   "r2",
		NetPort: "net2port0",
		Ifa:     "10.2.0.1/24",
		Remotes: []string{"10.1.0.2", "10.2.0.2"},
	},

	{
		Netns:   "h3",
		NetPort: "net2port1",
		Ifa:     "10.2.0.2/24",
		Routes: []Route{
			{"10.1.0.0/24", "10.2.0.1"},
		},
		Remotes: []string{"10.2.0.1", "10.1.0.2", "10.1.0.20"},
	},
}
