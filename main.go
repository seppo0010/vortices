package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"runtime"
	"sort"

	dc "github.com/seppo0010/vortices/dockercompose"
)

func buildDockerPath(path string) (string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("path %s does not exist", path)
	}

	var out bytes.Buffer
	cmd := exec.Command("docker", "build", path)
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to build docker image at path %s: %s", path, err.Error())
	}
	submatches := regexp.MustCompile(`Successfully built ([a-fA-F0-9]*)`).FindStringSubmatch(string(out.Bytes()))
	if len(submatches) == 0 {
		return "", fmt.Errorf("could not find docker image tag. Full output:\n%s", string(out.Bytes()))
	}

	return submatches[1], nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <path to target>", os.Args[0])
	}
	router, err := buildDockerPath("./router")
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	image, err := buildDockerPath(os.Args[1])
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	if !runTests(image, router, os.Args[2:]) {
		os.Exit(1)
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
func runTests(image, router string, tests []string) bool {
	passed := true
	for _, test := range []func(image, router string) error{testICECandidatesGather, testGateway, testStun} {
		testName := runtime.FuncForPC(reflect.ValueOf(test).Pointer()).Name()
		if len(tests) > 0 && !contains(tests, testName) {
			log.Printf("skipping test %v", testName)
			continue
		}
		log.Printf("running test %v", testName)
		err := test(image, router)
		if err != nil {
			log.Printf("test %v failed: %s", testName, err.Error())
			passed = false
		} else {
			log.Printf("finished OK test %v", testName)
		}
	}
	return passed
}

func checkCandidatesMatch(candidates []*Candidate, ipaddresses []string) error {
	if len(candidates) != len(ipaddresses) {
		return fmt.Errorf("expected %d candidates, got %d", len(ipaddresses), len(candidates))
	}
	addresses := make([]string, len(candidates))
	for i, candidate := range candidates {
		addresses[i] = candidate.Address
	}
	sort.Strings(addresses)
	sort.Strings(ipaddresses)
	for i, addr1 := range addresses {
		if addr1 != ipaddresses[i] {
			return fmt.Errorf("ip addresses do not match\ncontainer has: %#v\nreceived: %#v", ipaddresses, addresses)
		}
	}
	return nil
}

func testICECandidatesGather(image, router string) error {
	setup := dc.NewSetup()
	network1 := setup.NewNetwork("network1", "172.18.0.0/24")
	network2 := setup.NewNetwork("network2", "172.19.0.0/24")
	computers := []*dc.Computer{
		setup.NewComputer("computer", image, nil, []*dc.Network{network1, network2}),
		setup.NewComputer("computer2", image, nil, []*dc.Network{network1}),
	}
	err := setup.Start()
	if err != nil {
		return err
	}
	defer setup.Stop()
	for _, computer := range computers {
		candidates, err := (&Computer{computer}).GatherCandidates()
		if err != nil {
			return err
		}
		ips, err := computer.GetAllIPAddresses()
		if err != nil {
			return err
		}
		err = checkCandidatesMatch(candidates, ips)
		if err != nil {
			return err
		}
	}
	return nil
}

func testGateway(image, router string) error {
	setup := dc.NewSetup()
	network1 := setup.NewNetwork("network1", "172.20.0.0/24")
	internet := setup.NewNetwork("internet", "172.21.0.0/24")
	gateway := setup.NewRouter("myrouter", router, []*dc.Network{network1, internet})
	computers := []*dc.Computer{
		setup.NewComputer("computer", image, gateway, []*dc.Network{network1}),
		setup.NewComputer("computer2", image, nil, []*dc.Network{internet}),
	}
	err := setup.Start()
	if err != nil {
		return err
	}
	defer setup.Stop()
	ips, err := computers[1].GetAllIPAddresses()
	if err != nil {
		return err
	}
	_, err = (&Computer{computers[0]}).Ping(ips[0])
	if err != nil {
		return err
	}
	return nil
}

func testStun(image, router string) error {
	setup := dc.NewSetup()
	network1 := setup.NewNetwork("network1", "172.20.0.0/24")
	internet := setup.NewNetwork("internet", "172.21.0.0/24")
	routerComputer := setup.NewRouter("myrouter", router, []*dc.Network{network1, internet})
	computer := setup.NewComputer("computer", image, routerComputer, []*dc.Network{network1})
	stun := setup.NewSTUNServer("stun-server", []*dc.Network{internet})
	err := setup.Start()
	if err != nil {
		return err
	}
	defer setup.Stop()
	ips, err := stun.GetAllIPAddresses()
	if err != nil {
		return err
	}
	stunIP, err := (&Computer{computer}).GetIPFromSTUN(ips[0] + ":3478")
	if err != nil {
		return err
	}
	routerIP, err := routerComputer.GetIPAddressForNetwork(internet)
	if err != nil {
		return err
	}
	if stunIP != routerIP {
		return fmt.Errorf("expected stun ip (%s) to match router ip (%s)", stunIP, routerIP)
	}
	return nil
}
