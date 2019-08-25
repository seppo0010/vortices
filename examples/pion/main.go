package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/pion/ice"
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

	log.Fatal(http.ListenAndServe(":8080", nil))
}
