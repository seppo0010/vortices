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
	"strings"

	dc "github.com/seppo0010/vortices/dockercompose"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <path to target>", os.Args[0])
	}
	path := os.Args[1]
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalf("path %s does not exist", path)
	}

	var out bytes.Buffer
	cmd := exec.Command("docker", "build", path)
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to build docker image at path %s: %s", path, err.Error())
	}
	submatches := regexp.MustCompile(`Successfully built ([a-fA-F0-9]*)`).FindStringSubmatch(string(out.Bytes()))
	if len(submatches) == 0 {
		log.Fatalf("could not find docker image tag. Full output:\n%s", out.Bytes)
	}

	if !runTests(submatches[1]) {
		os.Exit(1)
	}
}

func runTests(image string) bool {
	passed := true
	for _, test := range []func(s string) error{testICECandidatesGather} {
		testName := runtime.FuncForPC(reflect.ValueOf(test).Pointer()).Name()
		log.Printf("running test %v", testName)
		err := test(image)
		if err != nil {
			log.Printf("test %v failed: %s", testName, err.Error())
			passed = false
		} else {
			log.Printf("finished OK test %v", testName)
		}
	}
	return passed
}

func startSetup(setup *dc.Setup) ([]*Computer, error) {
	f, err := os.Create("docker-compose.yml")
	if err != nil {
		return nil, err
	}
	_, err = f.WriteString(setup.ToYML())
	if err != nil {
		return nil, err
	}
	f.Close()

	var stderr bytes.Buffer
	cmd := exec.Command("docker-compose", "up", "-d")
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalf("failed to start docker-compose: %s", err.Error())
	}

	computers := make([]*Computer, len(setup.Computers))
	for i, computer := range setup.Computers {
		cmd = exec.Command("docker", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}} {{end}}", computer.Name)
		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		err = cmd.Run()
		if err != nil {
			log.Fatalf("failed to run docker inspect: %s", err.Error())
		}
		computers[i] = &Computer{
			Computer:    computer,
			IPAddresses: strings.Split(strings.Trim(string(stdout.Bytes()), " \n"), " "),
		}
	}

	return computers, nil
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

func testICECandidatesGather(image string) error {
	setup := dc.NewSetup()
	network1 := setup.NewNetwork("network1")
	network2 := setup.NewNetwork("network2")
	setup.NewComputer("computer", image, "", []*dc.Network{network1, network2})
	setup.NewComputer("computer2", image, "", []*dc.Network{network1})
	computers, err := startSetup(setup)
	if err != nil {
		return err
	}
	for _, computer := range computers {
		candidates, err := computer.GatherCandidates()
		if err != nil {
			return err
		}
		err = checkCandidatesMatch(candidates, computer.IPAddresses)
		if err != nil {
			return err
		}
	}
	return nil
}
