package dockercompose

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

type runRequest struct {
	args []string
}

type runResponse struct {
	stdout []byte
	stderr []byte
	err    error
}

func (s *Setup) exec(r runRequest) runResponse {
	var rr runResponse
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(r.args[0], r.args[1:]...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = s.tmpDir
	rr.err = cmd.Run()
	rr.stdout = stdout.Bytes()
	rr.stderr = stderr.Bytes()
	if rr.err != nil {
		log.Printf("command failed, creating logs in %s", s.tmpDir)

		f, err := os.Create(path.Join(s.tmpDir, "argv"))
		if err != nil {
			log.Printf("error creating argv file (%s): %s", s.tmpDir, err.Error())
		}
		_, err = f.WriteString(strings.Join(r.args, " "))
		if err != nil {
			log.Printf("error writing argv (%s): %s", s.tmpDir, err.Error())
		}

		f, err = os.Create(path.Join(s.tmpDir, "stdout"))
		if err != nil {
			log.Printf("error creating stdout file (%s): %s", s.tmpDir, err.Error())
		}
		_, err = f.Write(rr.stdout)
		if err != nil {
			log.Printf("error writing stdout (%s): %s", s.tmpDir, err.Error())
		}

		f, err = os.Create(path.Join(s.tmpDir, "stderr"))
		if err != nil {
			log.Printf("error creating stderr file (%s): %s", s.tmpDir, err.Error())
		}
		_, err = f.Write(rr.stderr)
		if err != nil {
			log.Printf("error writing stderr (%s): %s", s.tmpDir, err.Error())
		}
	}
	return rr
}
