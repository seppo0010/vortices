package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/pion/ice"
    "github.com/sparrc/go-ping"
)

func main() {
	http.HandleFunc("/gather-candidates", func(w http.ResponseWriter, r *http.Request) {
		agent, err := ice.NewAgent(&ice.AgentConfig{NetworkTypes: []ice.NetworkType{
			ice.NetworkTypeUDP4,
			ice.NetworkTypeUDP6,
		}})
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		candidates, err := agent.GetLocalCandidates()
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		candidatesMap := make([]interface{}, len(candidates))
		for i, candidate := range candidates {
			c := map[string]interface{}{}
			c["address"] = candidate.Address()
			candidatesMap[i] = c
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"candidates": candidatesMap,
		})
	})
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
pinger, err := ping.NewPinger(r.FormValue("ip"))
if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
}
pinger.Timeout = time.Second * 5
pinger.SetPrivileged(true)
pinger.Count = 3
pinger.Run()
stats := pinger.Statistics()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"times": stats.Rtts,
		})
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
