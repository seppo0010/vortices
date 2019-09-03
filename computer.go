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
}

type Candidate struct {
	Address string `json:"address"`
}

type ICEServer struct {
	URLs []string `json:"urls"`
}

type AgentConfig struct {
	ICEServers []ICEServer `json:"ice_servers"`
}

func (c *Computer) url(path string) string {
	return fmt.Sprintf("http://%s:8080/%s", c.GetIPAddress(), path)
}

func (c *Computer) isGoodResponse(r *http.Response, err error) (*http.Response, error) {
	if err != nil {
		return nil, err
	}
	if r.StatusCode < 400 {
		return r, nil
	}
	data := make([]byte, r.ContentLength)
	r.Body.Read(data)
	r.Body.Close()
	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("error running request: %s", string(data))
}

func (c *Computer) get(path string) (*http.Response, error) {
	r, err := c.isGoodResponse(http.Get(c.url(path)))
	if err != nil {
		return nil, fmt.Errorf("request %s failed: %s", path, err.Error())
	}
	return r, nil
}

func (c *Computer) post(path string, vals url.Values) (*http.Response, error) {
	r, err := c.isGoodResponse(http.PostForm(c.url(path), vals))
	if err != nil {
		return nil, fmt.Errorf("request %s failed: %s", path, err.Error())
	}
	return r, nil
}

func (c *Computer) GatherCandidates() ([]*Candidate, error) {
	res, err := c.get("gather-candidates")
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
	res, err := c.post("ping", url.Values{"ip": {ip}, "times": {"3"}})
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

func (c *Computer) GetIPFromSTUN(stun string) (string, error) {
	res, err := c.post("get-ip-from-stun", url.Values{"stun": {stun}})
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	target := struct {
		IP string `json:"ip"`
	}{}
	err = json.NewDecoder(res.Body).Decode(&target)
	return target.IP, err
}

func (c *Computer) CreateOffer(config *AgentConfig) (string, error) {
	vals := url.Values{}
	if config != nil {
		conf, err := json.Marshal(config)
		if err != nil {
			return "", err
		}
		vals["config"] = []string{string(conf)}
	}
	res, err := c.post("create-offer", vals)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	target := struct {
		SDP string `json:"sdp"`
	}{}
	err = json.NewDecoder(res.Body).Decode(&target)
	return target.SDP, err
}

func (c *Computer) CreateAnswer(offer string, config *AgentConfig) (string, error) {
	vals := url.Values{"offer": {offer}}
	if config != nil {
		conf, err := json.Marshal(config)
		if err != nil {
			return "", err
		}
		vals["config"] = []string{string(conf)}
	}
	res, err := c.post("create-answer", vals)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	target := struct {
		SDP string `json:"sdp"`
	}{}
	err = json.NewDecoder(res.Body).Decode(&target)
	return target.SDP, err
}

func (c *Computer) ReceivedAnswer(answer string) error {
	res, err := c.post("received-answer", url.Values{"answer": {answer}})
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var val interface{}
	return json.NewDecoder(res.Body).Decode(&val)
}

func (c *Computer) GetICEConnectionState() (string, error) {
	res, err := c.get("get-ice-connection-state")
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	target := struct {
		State string `json:"state"`
	}{}
	err = json.NewDecoder(res.Body).Decode(&target)
	return target.State, err
}
