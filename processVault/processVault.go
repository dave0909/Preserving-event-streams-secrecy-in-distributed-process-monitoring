package main

/**
go build processVault.go && sudo ./processVault --node 0 --http :2020 --cluster "1,:3030;2,:3031;3,:3032"
go build processVault.go && sudo ./processVault --node 1 --http :2021 --cluster "1,:3030;2,:3031;3,:3032"
go build processVault.go && sudo ./processVault --node 2 --http :2022 --cluster "1,:3030;2,:3031;3,:3032"
**/

import (
	crypto "crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	psm "main/processStateMachine"
	"main/rafTEE"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

type processVault struct {
	raft                *rafTEE.RafTEEserver
	processStateMachine *psm.ProcessStateMachine
}
type Config struct {
	cluster []rafTEE.ClusterMember
	index   int
	id      string
	address string
	http    string
}

// Method used by clients to update the process state
//TODO: this should be invoked by the event generator to submit events
// TODO key value should not be sen t in a get request but in a post request
func (hs processVault) setHandler(w http.ResponseWriter, r *http.Request) {
	var c psm.Command
	c.Kind = psm.SetCommand
	c.Key = r.URL.Query().Get("key")
	c.Value = r.URL.Query().Get("value")
	_, err := hs.raft.Apply([][]byte{psm.EncodeCommand(c)})
	if err != nil {
		log.Printf("Could not write key-value: %s", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

// Method used to by clients to read the state of the process
func (hs processVault) getHandler(w http.ResponseWriter, r *http.Request) {
	var c psm.Command
	c.Kind = psm.GetCommand
	c.Key = r.URL.Query().Get("key")

	var value []byte
	var err error
	if r.URL.Query().Get("relaxed") == "true" {
		v, ok := hs.processStateMachine.Db.Load(c.Key)
		if !ok {
			err = fmt.Errorf("Key not found")
		} else {
			value = []byte(v.(string))
		}
	} else {
		var results []rafTEE.ApplyResult
		results, err = hs.raft.Apply([][]byte{psm.EncodeCommand(c)})
		if err == nil {
			if len(results) != 1 {
				err = fmt.Errorf("Expected single response from Raft, got: %d.", len(results))
			} else if results[0].Error != nil {
				err = results[0].Error
			} else {
				value = results[0].Result
			}

		}
	}

	if err != nil {
		log.Printf("Could not encode key-value in http response: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	written := 0
	for written < len(value) {
		n, err := w.Write(value[written:])
		if err != nil {
			log.Printf("Could not encode key-value in http response: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		written += n
	}

}

// Function to get the configuration data from the script command-line
func getConfig() Config {
	cfg := Config{}
	var node string
	for i, arg := range os.Args[1:] {
		//Here you get the node identifier from the parameters
		if arg == "--node" {
			var err error
			node = os.Args[i+2]
			cfg.index, err = strconv.Atoi(node)
			if err != nil {
				log.Fatal("Expected $value to be a valid integer in `--node $value`, got: %s", node)
			}
			i++
			continue
		}
		//Here you get the client address to send the commands
		if arg == "--http" {
			cfg.http = os.Args[i+2]
			i++
			continue
		}
		//Here you get the data of the cluster members
		if arg == "--cluster" {
			cluster := os.Args[i+2]
			var clusterEntry rafTEE.ClusterMember
			//Iterate over the cluster members
			for _, part := range strings.Split(cluster, ";") {
				idAddress := strings.Split(part, ",")
				var err error
				clusterEntry.Id, err = strconv.ParseUint(idAddress[0], 10, 64)
				if err != nil {
					log.Fatal("Expected $id to be a valid integer in `--cluster $id,$ip`, got: %s", idAddress[0])
				}
				clusterEntry.Address = idAddress[1]
				cfg.cluster = append(cfg.cluster, clusterEntry)
			}

			i++
			continue
		}
	}

	if node == "" {
		log.Fatal("Missing required parameter: --node $index")
	}

	if cfg.http == "" {
		log.Fatal("Missing required parameter: --http $address")
	}

	if len(cfg.cluster) == 0 {
		log.Fatal("Missing required parameter: --cluster $node1Id,$node1Address;...;$nodeNId,$nodeNAddress")
	}

	return cfg
}

// TODO: This should be put in a main.go file. for now let's keep it here
func main() {
	var b [8]byte
	_, err := crypto.Read(b[:])
	if err != nil {
		panic("cannot seed math/rand package with cryptographically secure random number generator")
	}
	rand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
	cfg := getConfig()
	BuildAndRun(cfg, err)
}

// Build a PromTEE node and run it
func BuildAndRun(cfg Config, err error) {
	var sm psm.ProcessStateMachine
	var db sync.Map
	sm.Server = cfg.index
	sm.Db = &db
	s := rafTEE.NewRafTEEServer(cfg.cluster, &sm, ".", cfg.index)
	go s.Start()
	hs := processVault{s, &sm}
	http.HandleFunc("/set", hs.setHandler)
	http.HandleFunc("/get", hs.getHandler)
	err = http.ListenAndServe(cfg.http, nil)
	if err != nil {
		panic(err)
	}
}
