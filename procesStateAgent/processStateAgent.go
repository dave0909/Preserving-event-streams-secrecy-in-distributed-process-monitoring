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
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
)

// go run processStateAgent.go localhost:6065 localhost:1234 false
// CGO_CFLAGS=-I/opt/ego/include CGO_LDFLAGS=-L/opt/ego/lib ego-go run processStateAgent.go localhost:6065 localhost:1234 false
func main() {
	psaServer := os.Args[1]
	esgAddress := os.Args[2]
	skippAttestation := os.Args[3]
	testMode := os.Args[4]
	skippAttestationBool := skippAttestation == "true"
	testModeBool := testMode == "true"
	if skippAttestationBool {
		fmt.Println("Running process state agent without remote attestation")
	}

	psa := initProcessStateAgent(psaServer, esgAddress, skippAttestationBool, testModeBool)
	//TODO: UNCOMMENT HERE TO CONNECT EVENT STREAM GENERATOR
	//psa.connectToEventSteamGenerator()
	go psa.StartRPCServer(psaServer)
	go psa.checkTimeouts()
	psa.connectToEventSteamGenerator()
}

func (psa *ProcessStateAgent) connectToEventSteamGenerator() {
	for {
		conn, err := net.Dial("tcp", psa.EventStreamGenerator)
		if err != nil {
			fmt.Println("Error connecting:", err)
			//os.Exit(1)
		} else {
			defer conn.Close()
			fmt.Println("Connected to Event Stream Generator at " + psa.EventStreamGenerator)
			psa.readEventStream(conn)
			break
		}
	}
}

// Struct type ProcessStateAgent
type ProcessStateAgent struct {
	//Map ProcessVault reference->connection
	//ProcessVaultConnections map[string]rpc.Client
	//Event Stream Generator address
	EventStreamGenerator string
	//Subscription list
	Subscriptions []attestation.Subscription
	//Address
	Address string
	// Publib key
	PublicKey []byte
	// Mutex
	mu sync.Mutex
	//Subcounter
	SubCounter      int
	skipAttestation bool
	submittedEvents int
	mapArrivals     map[string]int
	testMode        bool
	delayHubClient  *rpc.Client
}

func initProcessStateAgent(address string, eventStreamGenerator string, skipAttestation bool, testMode bool) ProcessStateAgent {
	if testMode {
		fmt.Println("Connecting to delay hub at localhost:8388")
		client, err := rpc.Dial("tcp", "localhost:8388")
		if err != nil {
			panic(err)
		}
		return ProcessStateAgent{EventStreamGenerator: eventStreamGenerator, Address: address, PublicKey: []byte("publicKey"), mu: sync.Mutex{}, skipAttestation: skipAttestation, mapArrivals: make(map[string]int), submittedEvents: 0, testMode: testMode, delayHubClient: client}

	}
	return ProcessStateAgent{EventStreamGenerator: eventStreamGenerator, Address: address, PublicKey: []byte("publicKey"), mu: sync.Mutex{}, skipAttestation: skipAttestation, mapArrivals: make(map[string]int), submittedEvents: 0, testMode: testMode, delayHubClient: nil}
}

func (psa *ProcessStateAgent) readEventStream(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		event, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from server:", err)
			break
		}
		if err != nil {
			log.Fatalf("Failed to parse XES data: %v", err)
		}
		//Register the event timestamp in the mapArrivals
		psa.mapArrivals[event] = int(time.Now().UnixMilli())
		psa.broadcastEvent(event)
	}
}

// Send the event to an RPC server with the sendEvent rpc call
func (psa *ProcessStateAgent) sendEvent(eventString string, client rpc.Client) {
	// Connect to the RPC server
	// Call the SendEvent method
	var reply string
	eventSubmission := eventsubmission.EventSubmission{EncryptedEvent: eventString, AgentReference: psa.Address}
	psa.mu.Lock()
	err := client.Call("EventDispatcher.SendEvent", eventSubmission, &reply)
	if err != nil {
		log.Println("Error calling SendEvent: %v", err)
	} else {
		if psa.testMode {
			psa.submittedEvents++
			var arrivalReply bool
			arrivalArgs := &delayargs.ArrivalArgs{
				EventCode:        psa.submittedEvents,
				ArrivalTimestamp: psa.mapArrivals[eventString],
			}
			err = psa.delayHubClient.Call("DelayHub.WriteArrival", arrivalArgs, &arrivalReply)
			if err != nil {
				fmt.Printf("Error calling WriteArrival: %v\n", err)
			} else {
				fmt.Printf("WriteArrival success: %v\n", arrivalReply)
			}
			//TODO: send here the number of the submitted event alongside its timestamp to the delay hub

		}
	}
	psa.mu.Unlock()
}

func (psa *ProcessStateAgent) broadcastEvent(eventString string) {
	for _, sub := range psa.Subscriptions {
		lastHeartbeat := sub.Heartbeats[len(sub.Heartbeats)-1]
		provisioningKey := lastHeartbeat.ProvisioningKey
		encryptedEvent, err := psa.encryptEvent(eventString, provisioningKey)
		psa.mapArrivals[encryptedEvent] = psa.mapArrivals[eventString]
		go delete(psa.mapArrivals, eventString)
		if err != nil {
			log.Fatalf("Error encrypting event: %v", err)
		}
		//TODO: remove go here below if you have issues
		go psa.sendEvent(encryptedEvent, *sub.ClientConnection)
	}
}

func (psa *ProcessStateAgent) StartRPCServer(addr string) {
	rpc.Register(psa)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Listener error:", err)
	}
	log.Println("Serving RPC server on port ", addr)
	defer psa.delayHubClient.Close()
	rpc.Accept(listener)
}

func (psa *ProcessStateAgent) Subscribe(receiverAddress string, reply *string) error {
	client, err := rpc.Dial("tcp", receiverAddress)
	if err != nil {
		log.Fatalf("Error connecting to RPC server: %v", err)
	}
	var evidence attestation.Evidence
	nonce := "*ThisIsAnonce*"
	err = client.Call("EventDispatcher.GetEvidence", nonce, &evidence)
	isVerified, provisioningKey := psa.verifyEvidence(evidence, psa.skipAttestation)
	if isVerified {
		evidenceTime := evidence.Timestamp
		timeInt := 5
		newSubscription := attestation.Subscription{
			Nonce:            nonce,
			AgentAddress:     psa.Address,
			AgentPublicKey:   psa.PublicKey,
			Heartbeats:       []attestation.Evidence{},
			IsActive:         true,
			TimeInterval:     timeInt, //Seconds
			ClientConnection: client,
			Id:               psa.SubCounter + 1,
		}
		if provisioningKey != nil {
			evidence.ProvisioningKey = provisioningKey
		}
		newSubscription.Heartbeats = append(newSubscription.Heartbeats, attestation.Evidence{
			Report: evidence.Report,
			//TODO: change here after the attestation mechanism is implemented
			Timestamp:      evidenceTime,
			SubscriptionId: newSubscription.Id,
			//ProvisioningKey: evidence.ProvisioningKey,
			ProvisioningKey: evidence.ProvisioningKey[:32],
		})
		// Generate a JSON string from the subscription
		psa.SubCounter++
		subscriptionJSON, err := json.Marshal(newSubscription)
		if err != nil {
			log.Fatalf("Error marshalling subscription to JSON: %v", err)
		}
		psa.Subscriptions = append(psa.Subscriptions, newSubscription)

		*reply = string(subscriptionJSON)
	} else {
		*reply = "subscription denied"
	}
	return nil
}

// Function that continously check if the subscription time is expired
// To do this, we need to check the last heartbeat timestamp and compare it with the current time
func (psa *ProcessStateAgent) checkTimeouts() {
	for {
		time.Sleep(1 * time.Second)
		psa.mu.Lock()
		activeSubs := []attestation.Subscription{}
		currentSubs := psa.Subscriptions
		for _, sub := range currentSubs {
			if sub.IsActive {
				if time.Now().Unix()-sub.Heartbeats[len(sub.Heartbeats)-1].Timestamp > int64(sub.TimeInterval)+2 {
					sub.IsActive = false
					fmt.Println("Deactivate", sub)
				} else {
					activeSubs = append(activeSubs, sub)
				}
			}
		}
		psa.Subscriptions = activeSubs
		psa.mu.Unlock()
	}
}

// ReceiveHeartbeat processes incoming heartbeat evidence and updates the corresponding subscription.
func (psa *ProcessStateAgent) ReceiveHeartbeat(evidence *attestation.Evidence, reply *string) error {
	isVerified, provisioningKey := psa.verifyEvidence(*evidence, psa.skipAttestation)
	if isVerified {
		if provisioningKey != nil {
			evidence.ProvisioningKey = provisioningKey[:32]
		}
		// Find the subscription with the given id in the list of subscriptions
		for i, sub := range psa.Subscriptions {
			if sub.Id == evidence.SubscriptionId {
				sub.Heartbeats = append(sub.Heartbeats, *evidence)
				psa.Subscriptions[i] = sub // Update the subscription in the slice
				*reply = "heartbeat received"
				break
			} else {
				*reply = "Subscription not found"
			}
		}
	} else {
		*reply = "Heartbeat verification failed not verified"
	}
	//fmt.Println(psa.Subscriptions)
	return nil
}

// TODO: change here after the attestation mechanism is implemented
func (psa *ProcessStateAgent) verifyEvidence(evidence attestation.Evidence, skippAttestation bool) (bool, []byte) {
	if skippAttestation {
		return true, nil
	}
	encryptedReport, err := base64.StdEncoding.DecodeString(evidence.Report)
	if err != nil {
		fmt.Println(err)
		return false, nil
	}
	if err != nil {
		log.Fatalf("Error calling Get Evidence from event dispatcher: %v", err)
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
