package dockercompose

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"

	"github.com/google/uuid"
)

func BuildDockerPath(path string) (string, error) {
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

func BuildDocker(script string) (string, error) {
	dirPath := path.Join(os.TempDir(), uuid.New().String())
	err := os.MkdirAll(dirPath, 0744)
	if err != nil {
		return "", err
	}

	filePath := path.Join(dirPath, "Dockerfile")
	f, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	_, err = f.WriteString(script)
	if err != nil {
		return "", err
	}
	f.Close()

	var out bytes.Buffer
	cmd := exec.Command("docker", "build", dirPath)
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to build docker image at path %s: %s", filePath, err.Error())
	}
	submatches := regexp.MustCompile(`Successfully built ([a-fA-F0-9]*)`).FindStringSubmatch(string(out.Bytes()))
	if len(submatches) == 0 {
		return "", fmt.Errorf("could not find docker image tag. Full output:\n%s", string(out.Bytes()))
	}

	return submatches[1], nil
}
