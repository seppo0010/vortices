package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"runtime"

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
		err := test(image)
		if err != nil {
			log.Printf("test %v failed: %s", runtime.FuncForPC(reflect.ValueOf(test).Pointer()).Name(), err.Error())
			passed = false
		}
	}
	return passed
}

func startImage(image string) ([]string, error) {
	setup := dc.NewSetup()
	network1 := setup.NewNetwork("network1")
	network2 := setup.NewNetwork("network2")
	setup.NewComputer("computer", image, []*dc.Network{network1, network2})
	setup.NewComputer("computer2", image, []*dc.Network{network1})
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

	submatches := regexp.MustCompile(`(?:Starting|Creating) ([A-Za-z0-9_]*)`).FindAllStringSubmatch(string(stderr.Bytes()), -1)
	if len(submatches) == 0 {
		log.Fatalf("could not find docker Container ID. Full output:\n%s", string(stderr.Bytes()))
	}

	return nil, nil
}

func testICECandidatesGather(image string) error {
	_, err := startImage(image)
	if err != nil {
		return err
	}
	return nil
}
