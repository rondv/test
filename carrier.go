// Copyright © 2015-2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"
)

// Assert that named interface has carrier w/in 3sec.
func Carrier(netns, ifname string) error {
	const period = 250 * time.Millisecond
	fn := filepath.Join("/sys/class/net", ifname, "carrier")
	xargs := []string{"cat", fn}
	if len(netns) > 0 && netns != "default" {
		xargs = append([]string{"ip", "netns", "exec", netns},
			xargs...)
	}
	for t := 3 * (time.Second / period); t != 0; t-- {
		output, err := exec.Command(xargs[0], xargs[1:]...).Output()
		if err != nil {
			return err
		}
		if bytes.Equal(output, []byte("1\n")) {
			return nil
		}
		time.Sleep(period)
	}
	return fmt.Errorf("%s no carrier", ifname)
}
