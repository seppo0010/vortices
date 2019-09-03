package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/pion/ice"
	"github.com/pion/stun"
	"github.com/pion/webrtc/v2"
	"github.com/sparrc/go-ping"
)

func configToWebrtcConfig(conf string) (webrtc.Configuration, error) {
	config := webrtc.Configuration{}
	if conf == "" {
		return config, nil
	}
	type configuration struct {
		ICEServers []struct {
			URLs []string `json:"urls"`
		} `json:"ice_servers"`
	}
	received := configuration{}
	err := json.Unmarshal([]byte(conf), &received)
	if err != nil {
		return config, err
	}
	if len(received.ICEServers) > 0 {
		config.ICEServers = make([]webrtc.ICEServer, len(received.ICEServers))
		for i, server := range received.ICEServers {
			config.ICEServers[i] = webrtc.ICEServer{
				URLs: server.URLs,
			}
		}
	}
	return config, nil
}
func main() {
	var peerConnection *webrtc.PeerConnection
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
	http.HandleFunc("/get-ip-from-stun", func(w http.ResponseWriter, r *http.Request) {
		// Creating a "connection" to STUN server.
		c, err := stun.Dial("udp", r.FormValue("stun"))
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		// Building binding request with random transaction id.
		message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
		// Sending request to STUN server, waiting for response message.
		if err := c.Do(message, func(res stun.Event) {
			if res.Error != nil {
				w.WriteHeader(500)
				w.Write([]byte(res.Error.Error()))
				return
			}
			// Decoding XOR-MAPPED-ADDRESS attribute from message.
			var xorAddr stun.XORMappedAddress
			if err := xorAddr.GetFrom(res.Message); err != nil {
				w.WriteHeader(500)
				w.Write([]byte(err.Error()))
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"ip": xorAddr.IP,
			})
		}); err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
	})

	http.HandleFunc("/create-offer", func(w http.ResponseWriter, r *http.Request) {
		if peerConnection != nil {
			w.WriteHeader(400)
			w.Write([]byte("cannot create offer if peer connection already exists"))
			return
		}
		config, err := configToWebrtcConfig(r.FormValue("config"))
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}
		pc, err := webrtc.NewPeerConnection(config)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		peerConnection = pc
		offer, err := pc.CreateOffer(nil)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		err = pc.SetLocalDescription(offer)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"sdp": offer.SDP,
		})
	})

	http.HandleFunc("/create-answer", func(w http.ResponseWriter, r *http.Request) {
		if peerConnection != nil {
			w.WriteHeader(400)
			w.Write([]byte("cannot create answer if peer connection already exists"))
			return
		}
		config, err := configToWebrtcConfig(r.FormValue("config"))
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}
		pc, err := webrtc.NewPeerConnection(config)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		peerConnection = pc
		err = pc.SetRemoteDescription(webrtc.SessionDescription{
			SDP:  r.FormValue("offer"),
			Type: webrtc.SDPTypeOffer,
		})
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		answer, err := pc.CreateAnswer(nil)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		err = pc.SetLocalDescription(answer)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"sdp": answer.SDP,
		})
	})

	http.HandleFunc("/received-answer", func(w http.ResponseWriter, r *http.Request) {
		if peerConnection == nil {
			w.WriteHeader(400)
			w.Write([]byte("cannot receive answer if peer connection does not exists"))
			return
		}
		err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
			SDP:  r.FormValue("answer"),
			Type: webrtc.SDPTypeAnswer,
		})
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		json.NewEncoder(w).Encode(nil)
	})

	http.HandleFunc("/get-ice-connection-state", func(w http.ResponseWriter, r *http.Request) {
		if peerConnection == nil {
			w.WriteHeader(400)
			w.Write([]byte("no ICE connection state if peer connection does not exist"))
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"state": peerConnection.ICEConnectionState().String()})
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
