// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package docker

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	client "docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/container"
	"github.com/platinasystems/test"
	"github.com/platinasystems/test/netport"
	"gopkg.in/yaml.v2"
)

// initial values filled from yaml, then DevType and Name updated
type Router struct {
	Image    string
	Hostname string // netns
	Cmd      string
	Intfs    []struct {
		DevType  int
		IsBridge bool
		Name     string
		Address  []string // nil for BRIDGE_PORT
		Vlan     string   // PORT or BRIDGE_PORT
		Upper    string   // BRIDGE_PORT
	}
	id string
}

type Config struct {
	Volume  string
	Mapping string
	Routers []Router
	cli     *client.Client
	user    string
}

func Check(t *testing.T) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		t.Fatalf("Unable to get docker client: %v")
		return err
	}

	ver := cli.ClientVersion()
	Comment(t, "Docker client version", ver)
	_, err = cli.Info(context.Background())
	if err != nil {
		return err
	}
	_, err = cli.Ping(context.Background())
	if err != nil {
		t.Fatalf("Docker ping failed: %v", err)
		return err
	}
	if _, err := os.Stat("/var/run/netns"); os.IsNotExist(err) {
		_ = os.Mkdir("/var/run/netns", os.ModeDir)
	}

	return nil
}

// Log args if -test.vv
func Comment(t *testing.T, args ...interface{}) {
	t.Helper()
	if *test.VV {
		t.Log(args...)
	}
}

// Format args if -test.vv
func Commentf(t *testing.T, format string, args ...interface{}) {
	t.Helper()
	if *test.VV {
		t.Logf(format, args...)
	}
}

func LaunchContainers(t *testing.T, source []byte) (config *Config, err error) {
	lc := test.Cleanup{t}
	lc.Helper()

	cli, xerr := client.NewEnvClient()
	if xerr != nil {
		err = fmt.Errorf("Unable to get docker client: %v", xerr)
		return
	}

	config = &Config{}
	if xerr := yaml.Unmarshal(source, config); xerr != nil {
		err = fmt.Errorf("Unable to unmarshal yamlclient: %v", xerr)
		return
	}

	config.cli = cli

	path := "PATH=/usr/local/sbin"
	path += ":/usr/local/bin"
	path += ":/usr/sbin"
	path += ":/usr/bin"
	path += ":/sbin"
	path += ":/bin"
	path += ":/root"
	env := []string{path}

	// Common container config
	cc := &container.Config{}
	cc.Tty = true
	cc.Env = env

	var vdir string
	if config.Volume != "" && config.Mapping != "" {
		var pwd string
		pwd, err = syscall.Getwd()
		if err != nil {
			return
		}
		vdir = pwd + config.Volume
		cc.Volumes = map[string]struct{}{config.Mapping: {}}
		config.user = os.Getenv("SUDO_USER")
	}
	_ = vdir // make compiler happy

	// Common host config
	ch := &container.HostConfig{}
	ch.Privileged = true
	ch.NetworkMode = "none"

	// router specific cc & ch config
	for i, router := range config.Routers {
		if !isImageLocal(t, config.cli, router) {
			Comment(t, "no local container, trying to pull  remote")
			err = pullImage(t, config.cli, router)
			if err != nil {
				return
			}
			Comment(t, router.Image, "pulled from remote")
		}

		cc.Image = router.Image
		cc.Hostname = router.Hostname
		cc.Cmd = []string{router.Cmd}

		if vdir != "" {
			bind := vdir + "volumes/" + router.Hostname +
				":" + config.Mapping
			ch.Binds = []string{bind}
		}

		cresp, err2 := startContainer(t, config, cc, ch)
		if err2 != nil {
			err = err2
			return
		}
		config.Routers[i].id = cresp.ID
		// wait time for routing daemon before adding interfaces
		time.Sleep(2 * time.Second)

		// set rp_filter off, need to do this again later per interface
		lc.Program("ip", "netns", "exec", router.Hostname,
			"sysctl", "-w", "net/ipv4/conf/all/rp_filter=0")

		lc.Program("ip", "netns", "exec", router.Hostname,
			"sysctl", "-w", "net/ipv6/conf/all/disable_ipv6=0")

		lc.Program("ip", "netns", "exec", router.Hostname,
			"sysctl", "-w", "net/ipv6/conf/all/keep_addr_on_down=1")

		for _, intf := range router.Intfs {
			if intf.IsBridge {
				intf.DevType = netport.NETPORT_DEVTYPE_BRIDGE
			} else if intf.Upper != "" {
				intf.DevType = netport.NETPORT_DEVTYPE_BRIDGE_PORT
			} else {
				intf.DevType = netport.NETPORT_DEVTYPE_PORT
			}
			ns := router.Hostname
			if strings.Contains(intf.Name, "dummy") {
				lc.Program("ip", "link", "add", intf.Name,
					"type", "dummy")
				lc.Program("ip", "link", "set", intf.Name, "up")
			} else if intf.Vlan != "" {
				newIntf := intf.Name + "." + intf.Vlan
				lc.Program("ip", "link", "set", intf.Name, "up")
				lc.Program("ip", "link", "add", newIntf,
					"link", intf.Name, "type", "xeth-vlan")
				lc.Program("ip", "link", "set", newIntf, "up")
				intf.Name = newIntf
			} else if intf.DevType == netport.NETPORT_DEVTYPE_BRIDGE {
				lc.Program("ip", "netns", "exec", ns,
					"ip", "link", "add", intf.Name, "type", "xeth-bridge")
				lc.Program("ip", "netns", "exec", ns,
					"ip", "addr", "add", intf.Address[0], "dev", intf.Name)
				lc.Program("ip", "netns", "exec", ns,
					"ip", "link", "set", intf.Name, "up")
			}
			if intf.DevType != netport.NETPORT_DEVTYPE_BRIDGE {
				moveIntfContainer(t, ns, intf.Name, intf.Address)
			}
			if intf.DevType == netport.NETPORT_DEVTYPE_BRIDGE_PORT {
				lc.Program("ip", "netns", "exec", ns, "ip", "link", "set", "up", intf.Name)
				lc.Program("ip", "netns", "exec", ns,
					"ip", "link", "set", intf.Name, "master", intf.Upper)
			}
			lc.Program("ip", "netns", "exec", ns,
				"sysctl", "-w",
				"net/ipv4/conf/"+intf.Name+"/rp_filter=0")

			if *test.VVV {
				t.Logf("intf %+v\n", intf)
			}
		}
	}
	time.Sleep(1 * time.Second)
	return
}

func FindHost(config *Config, host string) (router Router, err error) {
	for _, r := range config.Routers {
		if r.Hostname == host {
			router = r
			return
		}
	}
	return
}

func ExecCmd(t *testing.T, ID string, config *Config,
	cmd []string) (out string, err error) {
	t.Helper()

	execOpts := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Detach:       false,
	}

	cli := config.cli
	ctx := context.Background()

	if *test.VVV {
		t.Log(ID, cmd)
	}

	execResp, err := cli.ContainerExecCreate(ctx, ID, execOpts)
	if err != nil {
		t.Logf("Error creating exec: %v", err)
		return
	}

	hresp, err := cli.ContainerExecAttach(ctx, execResp.ID, execOpts)
	if err != nil {
		t.Logf("Error attaching exec: %v", err)
		return
	}
	defer hresp.Close()

	content, err := ioutil.ReadAll(hresp.Reader)
	if err != nil {
		t.Logf("Error reading output: %v", err)
		return
	}
	out = strings.TrimSpace(string(content))

	ei, err := cli.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		t.Logf("Error exec Inspect: %v", err)
		return
	}
	if ei.Running {
		t.Logf("exec still running", ei)
	}
	if ei.ExitCode != 0 {
		err = fmt.Errorf("[%v] exit code %v", cmd, ei.ExitCode)
		return
	}
	if *test.VV && len(out) > 0 {
		t.Log(out)
	}
	return
}

func PingCmd(t *testing.T, ID string, config *Config, target string) error {
	t.Helper()

	is_v6 := test.IsIPv6(target)
	cmd := []string{"/bin/ping", "-c1", "-W1", target}
	if is_v6 {
		cmd[0] = "/bin/ping6"
	}
	execOpts := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Detach:       false,
	}

	cli := config.cli
	ctx := context.Background()

	count := 10
	for i := 0; i < count; i++ {

		execResp, err := cli.ContainerExecCreate(ctx, ID, execOpts)
		if err != nil {
			t.Logf("Error creating exec: %v", err)
			return err
		}

		hresp, err := cli.ContainerExecAttach(ctx, execResp.ID,
			execOpts)
		if err != nil {
			t.Logf("Error attaching exec: %v", err)
			return err
		}
		defer hresp.Close()

		content, err := ioutil.ReadAll(hresp.Reader)
		if err != nil {
			t.Logf("Error reading output: %v", err)
			return err
		}
		out := string(content)

		ei, err := cli.ContainerExecInspect(ctx, execResp.ID)
		if err != nil {
			t.Logf("Error exec Inspect: %v", err)
			return err
		}
		if ei.Running {
			t.Logf("exec still running", ei)
			return err
		}
		if ei.ExitCode == 0 {
			return nil
		}
		Commentf(t, "%v\nping count %v", out, i)
		if ei.ExitCode != 0 {
			Commentf(t, "[%v] exit code %v", cmd, ei.ExitCode)
		}
		time.Sleep(1 * time.Second)
	}

	err := fmt.Errorf("ping timeout %v -> %v", ID, target)
	if err != nil && false {
		test.Pause.Prompt("ping timeout ", ID, " -> ", target)
	}
	return err
}

func TearDownContainers(t *testing.T, config *Config) {
	t.Helper()
	td := test.Cleanup{t}
	for _, r := range config.Routers {
		for _, intf := range r.Intfs {
			if intf.IsBridge {
				intf.DevType = netport.NETPORT_DEVTYPE_BRIDGE
			} else if intf.Upper != "" {
				intf.DevType = netport.NETPORT_DEVTYPE_BRIDGE_PORT
			} else {
				intf.DevType = netport.NETPORT_DEVTYPE_PORT
			}
			if intf.DevType == netport.NETPORT_DEVTYPE_BRIDGE {
				continue
			}
			if intf.Vlan != "" {
				newIntf := intf.Name + "." + intf.Vlan
				moveIntfDefault(t, r.Hostname, newIntf)
				td.Program("ip", "link", "del", newIntf)
			} else if strings.Contains(intf.Name, "dummy") {
				moveIntfDefault(t, r.Hostname, intf.Name)
				td.Program("ip", "link", "del", intf.Name)
			} else {
				moveIntfDefault(t, r.Hostname, intf.Name)
			}
		}
		// delete bridge after members moved to default and deleted
		for _, intf := range r.Intfs {
			if intf.DevType == netport.NETPORT_DEVTYPE_BRIDGE {
				td.Program("ip", "netns", "exec", r.Hostname,
					"ip", "link", "del", intf.Name)
			}
		}
		err := stopContainer(t, config, r.Hostname, r.id)
		if err != nil {
			t.Logf("Error: stopping %v: %v", r.Hostname, err)
		}

	}
	if config.user != "" {
		user := config.user + ":" + config.user
		cmd := []string{"chown", "-R", user, "testdata"}
		exec.Command(cmd[0], cmd[1:]...).Run()
	}
	config.cli.Close()
}

func isImageLocal(t *testing.T, cli *client.Client, router Router) bool {

	images, err := cli.ImageList(context.Background(),
		types.ImageListOptions{})
	if err != nil {
		t.Error("failed to get docker image list")
		return false
	}

	for _, i := range images {
		for _, tag := range i.RepoTags {
			if tag == router.Image {
				return true
			}
		}
	}
	return false
}

func isContainerRunning(t *testing.T, config *Config, name string) bool {

	conts, err := config.cli.ContainerList(context.Background(),
		types.ContainerListOptions{All: true})
	if err != nil {
		t.Error("failed to get docker container list")
		return false
	}

	for _, cont := range conts {
		for _, imagename := range cont.Names {
			if imagename[1:] == name {
				return true
			}
		}
	}
	return false
}

func pullImage(t *testing.T, cli *client.Client, router Router) error {
	repo := "docker.io/" + router.Image
	out, err := cli.ImagePull(context.Background(), repo,
		types.ImagePullOptions{})
	if err != nil {
		t.Error("failed to pull remote image")
		return err
	}
	defer out.Close()
	io.Copy(os.Stdout, out)
	return nil
}

func startContainer(t *testing.T, config *Config, cc *container.Config,
	ch *container.HostConfig) (cresp container.ContainerCreateCreatedBody,
	err error) {

	assert := test.Assert{t}
	cli := config.cli

	if isContainerRunning(t, config, cc.Hostname) {
		err = fmt.Errorf("Container %v already running", cc.Hostname)
		return
	}
	Comment(t, "Starting container", cc.Hostname)

	ctx := context.Background()

	cresp, err = cli.ContainerCreate(ctx, cc, ch, nil, cc.Hostname)
	if err != nil {
		t.Errorf("Error creating container: %v", err)
		return
	}

	err = cli.ContainerStart(ctx, cresp.ID, types.ContainerStartOptions{})
	if err != nil {
		t.Errorf("Error starting container: %v", err)
		return
	}

	pid, err := getPid(cc.Hostname)
	if err != nil {
		t.Errorf("Error getting pid for %v: %v", cc.Hostname, err)
	}
	src := "/proc/" + pid + "/ns/net"
	dst := "/var/run/netns/" + cc.Hostname
	assert.Program("ln", "-s", src, dst)
	return
}

func stopContainer(t *testing.T, config *Config, name string,
	ID string) error {

	t.Helper()
	Comment(t, "Stopping container", name)

	cli := config.cli
	ctx := context.Background()

	err := cli.ContainerStop(ctx, ID, nil)
	if err != nil {
		t.Errorf("Error stoping %v %v: %v", name, ID, err)
		return err
	}

	err = cli.ContainerRemove(ctx, ID,
		types.ContainerRemoveOptions{RemoveVolumes: true})
	if err != nil {
		t.Errorf("Error removing volume %v: %v", name, err)
		return err
	}
	link := "/var/run/netns/" + name
	test.Cleanup{t}.Program("rm", link)

	return nil
}

func getPid(ID string) (pid string, err error) {

	cmd := []string{"/usr/bin/docker", "inspect", "-f", "'{{.State.Pid}}'",
		ID}
	bytes, err := exec.Command(cmd[0], cmd[1:]...).Output()
	if err != nil {
		return
	}
	pid = string(bytes)
	pid = strings.Replace(pid, "\n", "", -1)
	pid = strings.Replace(pid, "'", "", -1)
	return
}

func moveIntfContainer(t *testing.T, container string, intf string,
	addr []string) error {

	t.Helper()
	assert := test.Assert{t}

	Comment(t, "moving", intf, "to container", container,
		"with address", addr)

	assert.Program("ip", "link", "set", intf, "netns", container)
	assert.Program("ip", "-n", container, "link", "set", "up", "lo")
	assert.Program("ip", "-n", container, "link", "set", "down", intf)
	assert.Program("ip", "-n", container, "link", "set", "up", intf)
	for _, a := range addr {
		assert.Program("ip", "-n", container, "addr", "add", a, "dev", intf)
	}
	return nil
}

func moveIntfDefault(t *testing.T, container string, intf string) error {
	t.Helper()
	Comment(t, "moving", intf, "from", container, "to default")
	mv := test.Cleanup{t}
	mv.Program("ip", "-n", container, "link", "set", "down", intf)
	mv.Program("ip", "-n", container, "link", "set", intf, "netns", "1")
	mv.Program("ip", "link", "set", intf, "up")
	return nil
}
