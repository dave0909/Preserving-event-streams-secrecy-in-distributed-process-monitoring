package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/edgelesssys/ego/eclient"
	"io"
	"log"
	"main/utils/attestation"
	"main/utils/delayargs"
	"main/utils/eventsubmission"
	"main/utils/xes"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type ProcessStateAgent struct {
	EventStreamGenerator string
	Subscriptions        []attestation.Subscription
	Address              string
	PublicKey            []byte
	subscriptionsMu      sync.RWMutex
	SubCounter           int
	skipAttestation      bool
	submittedEvents      int
	mapArrivals          map[string]int
	testMode             bool
	delayHubClient       *rpc.Client
	rpcClientMu          sync.Mutex
}

func initProcessStateAgent(address string, eventStreamGenerator string, skipAttestation bool, testMode bool) ProcessStateAgent {
	var delayClient *rpc.Client
	if testMode {
		client, err := rpc.Dial("tcp", "localhost:8388")
		if err != nil {
			panic(err)
		}
		delayClient = client
	}

	return ProcessStateAgent{
		EventStreamGenerator: eventStreamGenerator,
		Address:              address,
		PublicKey:            []byte("publicKey"),
		subscriptionsMu:      sync.RWMutex{},
		rpcClientMu:          sync.Mutex{},
		skipAttestation:      skipAttestation,
		mapArrivals:          make(map[string]int),
		submittedEvents:      0,
		testMode:             testMode,
		delayHubClient:       delayClient,
	}
}

func (psa *ProcessStateAgent) readEventStream(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		event, err := reader.ReadString('\n')
		if err != nil {
			continue
		}

		if psa.testMode {
			parsedXes, err := xes.ParseXes(event)
			if err != nil {
				panic(err)
			}

			event_counter := parsedXes.Attributes["ESG_test_counter"].(int)
			arrivalArgs := &delayargs.ArrivalArgs{EventCode: event_counter}
			var arrivalReply bool

			psa.rpcClientMu.Lock()
			arrErr := psa.delayHubClient.Call("DelayHub.WriteArrival", arrivalArgs, &arrivalReply)
			psa.rpcClientMu.Unlock()

			if arrErr != nil {
				fmt.Printf("Error calling WriteArrival: %v\n", arrErr)
			}
		}
		psa.broadcastEvent(event)
	}
}

func (psa *ProcessStateAgent) sendEvent(eventString string, subInd int) {
	const maxRetries = 10
	var reply string

	psa.subscriptionsMu.RLock()
	if subInd >= len(psa.Subscriptions) {
		psa.subscriptionsMu.RUnlock()
		return
	}
	sub := psa.Subscriptions[subInd]
	lastHeartbeat := sub.Heartbeats[len(sub.Heartbeats)-1]
	client := sub.ClientConnection
	provisioningKey := lastHeartbeat.ProvisioningKey
	psa.subscriptionsMu.RUnlock()

	encryptedEvent, err := psa.encryptEvent(eventString, provisioningKey)
	if err != nil {
		fmt.Printf("Error encrypting event: %v\n", err)
		return
	}

	eventSubmission := eventsubmission.EventSubmission{
		EncryptedEvent: encryptedEvent,
		AgentReference: psa.Address,
	}

	for retries := 0; retries <= maxRetries; retries++ {
		psa.rpcClientMu.Lock()
		err = client.Call("EventDispatcher.SendEvent", eventSubmission, &reply)
		psa.rpcClientMu.Unlock()
		if err == nil {
			if psa.testMode {
				psa.subscriptionsMu.Lock()
				psa.submittedEvents++
				psa.subscriptionsMu.Unlock()
			}
			return
		}
		if retries < maxRetries {
			//fmt.Printf("Attempt %d failed: %v. Retrying...\n", retries+1, err)
			//time.Sleep(10 * time.Millisecond)
		} else {
			//fmt.Printf("Failed to send event after %d attempts: %v\n", maxRetries+1, err)
			return
		}
	}
}

func (psa *ProcessStateAgent) broadcastEvent(eventString string) {
	psa.subscriptionsMu.RLock()
	subCount := len(psa.Subscriptions)
	psa.subscriptionsMu.RUnlock()
	// Use an unbuffered channel to control the number of concurrent goroutines
	semaphore := make(chan struct{}, 10) // Limit to 10 concurrent sends

	var wg sync.WaitGroup
	for subI := 0; subI < subCount; subI++ {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire a semaphore slot
		go func(index int) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release the semaphore slot
			psa.sendEvent(eventString, index)
		}(subI)
	}
	wg.Wait()
}

func (psa *ProcessStateAgent) StartRPCServer(addr string) {
	rpc.Register(psa)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Listener error:", err)
	}
	log.Println("Serving RPC server on port ", addr)
	defer func() {
		if psa.delayHubClient != nil {
			psa.delayHubClient.Close()
		}
	}()
	rpc.Accept(listener)
}

func (psa *ProcessStateAgent) Subscribe(receiverAddress string, reply *string) error {
	client, err := rpc.Dial("tcp", receiverAddress)
	if err != nil {
		log.Fatalf("Error connecting to RPC server: %v", err)
	}

	var evidence attestation.Evidence
	nonce := "*ThisIsAnonce*"

	psa.rpcClientMu.Lock()
	err = client.Call("EventDispatcher.GetEvidence", nonce, &evidence)
	psa.rpcClientMu.Unlock()

	isVerified, provisioningKey := psa.verifyEvidence(evidence, psa.skipAttestation)
	if !isVerified {
		*reply = "subscription denied"
		return nil
	}

	psa.subscriptionsMu.Lock()
	defer psa.subscriptionsMu.Unlock()

	if provisioningKey != nil {
		evidence.ProvisioningKey = provisioningKey[:32]
	}

	evidenceTime := evidence.Timestamp
	psa.SubCounter++
	newSubscription := attestation.Subscription{
		Nonce:            nonce,
		AgentAddress:     psa.Address,
		AgentPublicKey:   psa.PublicKey,
		Heartbeats:       []attestation.Evidence{},
		IsActive:         true,
		TimeInterval:     5,
		ClientConnection: client,
		Id:               psa.SubCounter,
	}

	newSubscription.Heartbeats = append(newSubscription.Heartbeats, attestation.Evidence{
		Report:          evidence.Report,
		Timestamp:       evidenceTime,
		SubscriptionId:  newSubscription.Id,
		ProvisioningKey: evidence.ProvisioningKey,
	})

	subscriptionJSON, err := json.Marshal(newSubscription)
	if err != nil {
		log.Fatalf("Error marshalling subscription to JSON: %v", err)
	}

	psa.Subscriptions = append(psa.Subscriptions, newSubscription)
	*reply = string(subscriptionJSON)
	return nil
}

func (psa *ProcessStateAgent) checkTimeouts() {
	if !psa.skipAttestation {
		for {
			time.Sleep(1 * time.Second)
			now := time.Now().Unix()
			psa.subscriptionsMu.Lock()
			activeSubs := []attestation.Subscription{}
			for _, sub := range psa.Subscriptions {
				if sub.IsActive && len(sub.Heartbeats) > 0 {
					lastHeartbeatTime := sub.Heartbeats[len(sub.Heartbeats)-1].Timestamp
					if now-lastHeartbeatTime > int64(sub.TimeInterval)+3 {
						sub.IsActive = false
						fmt.Println("Deactivate", sub)
					} else {
						activeSubs = append(activeSubs, sub)
					}
				}
			}
			psa.Subscriptions = activeSubs
			psa.subscriptionsMu.Unlock()
		}
	}
}

func (psa *ProcessStateAgent) ReceiveHeartbeat(evidence *attestation.Evidence, reply *string) error {
	isVerified, provisioningKey := psa.verifyEvidence(*evidence, psa.skipAttestation)
	if !isVerified {
		*reply = "Heartbeat verification failed"
		return nil
	}

	if provisioningKey != nil {
		evidence.ProvisioningKey = provisioningKey[:32]
	}

	psa.subscriptionsMu.Lock()
	defer psa.subscriptionsMu.Unlock()

	for i, sub := range psa.Subscriptions {
		if sub.Id == evidence.SubscriptionId {
			psa.Subscriptions[i].Heartbeats = append(psa.Subscriptions[i].Heartbeats, *evidence)
			*reply = "heartbeat received"
			return nil
		}
	}

	*reply = "Subscription not found"
	return nil
}

func (psa *ProcessStateAgent) verifyEvidence(evidence attestation.Evidence, skippAttestation bool) (bool, []byte) {
	if skippAttestation {
		return true, nil
	}
	encryptedReport, err := base64.StdEncoding.DecodeString(evidence.Report)
	if err != nil {
		fmt.Println(err)
		return false, nil
	}

	decryptedReport, err := eclient.VerifyRemoteReport(encryptedReport)
	if err != nil {
		fmt.Println("Attestation failed. Cannot decrypt the report.")
		fmt.Println(err)
		return false, nil
	}
	return true, decryptedReport.Data
}

func (psa *ProcessStateAgent) encryptEvent(eventString string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(eventString), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (psa *ProcessStateAgent) connectToEventSteamGenerator() {
	fmt.Println("Connecting to Event Stream Generator at", psa.EventStreamGenerator)
	for {
		conn, err := net.Dial("tcp", psa.EventStreamGenerator)
		if err == nil {
			defer conn.Close()
			fmt.Println("Connected to Event Stream Generator")
			psa.readEventStream(conn)
			break
		}
	}
}

func main() {
	psaServer := os.Args[1]
	esgAddress := os.Args[2]
	skippAttestation := os.Args[3]
	testMode := os.Args[4]
	skippAttestationBool := skippAttestation == "true"
	testModeBool := testMode == "true"

	psa := initProcessStateAgent(psaServer, esgAddress, skippAttestationBool, testModeBool)
	go psa.StartRPCServer(psaServer)
	go psa.checkTimeouts()
	psa.connectToEventSteamGenerator()
}
