package dockercompose

import (
	"fmt"
	"os"
	"os/exec"
	"path"
)

type Router struct {
	*BaseComputer
	LAN      *Network
	Internet *Network
	commands []*exec.Cmd
}

func newRouter(setup *Setup, name, image string, lan, internet *Network) *Router {
	router := &Router{
		BaseComputer: newBaseComputer(setup, name, image, []*Network{lan, internet}),
		LAN:          lan,
		Internet:     internet,
		commands:     []*exec.Cmd{},
	}
	return router
}

func (router *Router) Start() error {
	if err := router.BaseComputer.Start(); err != nil {
		return err
	}

	for i, _ := range []*Network{router.LAN, router.Internet} {
		cmd := exec.Command("docker", "exec", "--privileged", router.Name, "tcpdump", "-i", fmt.Sprintf("eth%d", i), "-xx", "-vv", "-n")
		f, err := os.Create(path.Join(router.setup.tmpDir, fmt.Sprintf("%s.%d.tcpdump", router.Name, i)))
		if err != nil {
			return err
		}
		cmd.Stdout = f
		cmd.Stderr = os.Stderr
		err = cmd.Start()
		if err != nil {
			return err
		}
		router.commands = append(router.commands, cmd)
	}

	/*
		internetIP, err := router.GetIPAddressForNetwork(router.Internet)
		if err != nil {
			return err
		}
		lanIP, err := router.GetIPAddressForNetwork(router.LAN)
		if err != nil {
			return err
		}
	*/
	for _, rr := range []runRequest{
		runRequest{args: []string{"docker", "exec", "--privileged", router.Name, "iptables", "-A", "FORWARD", "-i", "eth0", "-o", "eth1", "-j", "NFQUEUE", "--queue-num", "1"}},
		runRequest{args: []string{"docker", "exec", "--privileged", router.Name, "iptables", "-A", "FORWARD", "-i", "eth1", "-o", "eth0", "-j", "NFQUEUE", "--queue-num", "2"}},
		runRequest{args: []string{"docker", "exec", "--privileged", router.Name, "iptables", "-A", "FORWARD", "-i", "eth0", "-o", "eth1", "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT"}},
		runRequest{args: []string{"docker", "exec", "--privileged", router.Name, "iptables", "-A", "FORWARD", "-i", "eth1", "-o", "eth0", "-j", "ACCEPT"}},
		runRequest{args: []string{"docker", "exec", "--privileged", router.Name, "iptables", "-t", "nat", "-A", "POSTROUTING", "-o", "eth0", "-j", "MASQUERADE"}},
	} {
		cmd := router.setup.exec(rr)
		if cmd.err != nil {
			return cmd.err
		}
	}
	return nil
}

func (router *Router) Stop() error {
	for _, cmd := range router.commands {
		if err := cmd.Process.Kill(); err != nil {
			return err
		}
		cmd.Wait()
	}
	return nil
}
