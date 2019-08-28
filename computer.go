package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	dc "github.com/seppo0010/vortices/dockercompose"
)

type Computer struct {
	*dc.Computer
	IPAddresses []string
}

type Candidate struct {
	Address string `json:"address"`
}

func (c *Computer) GatherCandidates() ([]*Candidate, error) {
	res, err := http.Get(fmt.Sprintf("http://%s:8080/gather-candidates", c.IPAddresses[0]))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	target := struct {
		Candidates []*Candidate `json:"candidates"`
	}{}
	err = json.NewDecoder(res.Body).Decode(&target)
	return target.Candidates, err
}

func (c *Computer) Ping(ip string) ([]float64, error) {
	res, err := http.PostForm(fmt.Sprintf("http://%s:8080/ping", c.IPAddresses[0]), url.Values{"ip": {ip}, "times": {"3"}})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	target := struct {
		Times []float64 `json:"times"`
	}{}
	err = json.NewDecoder(res.Body).Decode(&target)
	return target.Times, err
}
