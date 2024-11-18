package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"main/utils/attestation"
	"main/utils/eventsubmission"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
)

// go run processStateAgent.go localhost:6869 localhost:1234
func main() {
	psaServer := os.Args[1]
	esgAddress := os.Args[2]
	psa := initProcessStateAgent(psaServer, esgAddress)
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
	SubCounter int
}

func initProcessStateAgent(address string, eventStreamGenerator string) ProcessStateAgent {
	return ProcessStateAgent{EventStreamGenerator: eventStreamGenerator, Address: address, PublicKey: []byte("publicKey"), mu: sync.Mutex{}}
}

func (psa *ProcessStateAgent) readEventStream(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		event, err := reader.ReadString('\n')
		fmt.Println("Event received: ", event)
		if err != nil {
			fmt.Println("Error reading from server:", err)
			continue
		}
		if err != nil {
			log.Fatalf("Failed to parse XES data: %v", err)
		}
		psa.broadcastEvent(event)
	}
}

// Send the event to an RPC server with the sendEvent rpc call
func (psa *ProcessStateAgent) sendEvent(eventString string, client rpc.Client) {
	// Connect to the RPC server
	// Call the SendEvent method
	var reply string
	eventSubmission := eventsubmission.EventSubmission{EncryptedEvent: eventString, AgentReference: psa.Address}
	err := client.Call("EventDispatcher.SendEvent", eventSubmission, &reply)
	if err != nil {
		log.Println("Error calling SendEvent: %v", err)
		return
	}
}

func (psa *ProcessStateAgent) broadcastEvent(eventString string) {
	for _, sub := range psa.Subscriptions {
		lastHeartbeat := sub.Heartbeats[len(sub.Heartbeats)-1]
		provisioningKey := lastHeartbeat.ProvisioningKey
		encryptedEvent, err := psa.encryptEvent(eventString, provisioningKey)
		if err != nil {
			log.Fatalf("Error encrypting event: %v", err)
		}
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
	rpc.Accept(listener)
}

// TODO start implementing from here. You must integrate subscription generation and heartbeat meachanism
func (psa *ProcessStateAgent) Subscribe(receiverAddress string, reply *string) error {
	fmt.Println(receiverAddress)
	client, err := rpc.Dial("tcp", receiverAddress)
	if err != nil {
		log.Fatalf("Error connecting to RPC server: %v", err)
	}
	var evidence attestation.Evidence
	nonce := "*ThisIsAnonce*"
	err = client.Call("EventDispatcher.GetEvidence", nonce, &evidence)
	if err != nil {
		log.Fatalf("Error calling Get Evidence from event dispatcher: %v", err)
	}
	if psa.verifyEvidence(string(evidence.Report)) {
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
		newSubscription.Heartbeats = append(newSubscription.Heartbeats, attestation.Evidence{
			Report: []byte(evidence.Report),
			//TODO: change here after the attestation mechanism is implemented
			Timestamp:       evidenceTime,
			SubscriptionId:  newSubscription.Id,
			ProvisioningKey: evidence.ProvisioningKey,
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
	if psa.verifyEvidence(string(evidence.Report)) {
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
	fmt.Println(psa.Subscriptions)
	return nil
}

// TODO: change here after the attestation mechanism is implemented
func (psa *ProcessStateAgent) verifyEvidence(evidence string) bool {
	return true
}

func (psa *ProcessStateAgent) encryptEvent(eventString, key string) (string, error) {
	decodedKey, err := hex.DecodeString(key)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(decodedKey)
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
