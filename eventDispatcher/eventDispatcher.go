package eventDispatcher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"main/processStateManager"
	attestation "main/utils/attestation"
	"main/utils/eventsubmission"
	"main/utils/xes"
	"net"
	"net/rpc"
	"os"
	"time"
)

// go run eventDispatcher.go localhost:6969
// EventDispatcher struct
type EventDispatcher struct {
	EventChannel        chan xes.Event
	Subscriptions       map[string][]attestation.Subscription
	Address             string
	AttributeExtractors map[string]interface{}
}

// GetEvidence method to get ProcessVault's evidence
func (ed *EventDispatcher) GetEvidence(nonce string, reply *attestation.Evidence) error {
	//report, err := enclave.GetRemoteReport([]byte(nonce))
	//if err != nil {
	//	return err
	//}
	//*reply = string(report)
	//return nil
	provisioningKey, err := ed.generateProvisioningKey()
	if err != nil {
		return err
	}
	newEvidence := attestation.Evidence{Report: []byte(nonce), Timestamp: time.Now().Unix(), ProvisioningKey: provisioningKey}
	*reply = newEvidence
	return nil
}

// SendEvent method to handle incoming events in string format
func (ed *EventDispatcher) SendEvent(eventSubmission eventsubmission.EventSubmission, reply *string) error {
	//The reveiveing event should arrive as an encrypted message with the Provisioning generated for the last heartbeat
	// Parse the event string
	//fmt.Println("Received event: ", eventSubmission.EncryptedEvent)
	//Find the key for the subscription that sent the event
	var key string
	for _, subscription := range ed.Subscriptions[eventSubmission.AgentReference] {
		if len(subscription.Heartbeats) > 0 {
			key = subscription.Heartbeats[len(subscription.Heartbeats)-1].ProvisioningKey
			break
		}
	}
	if key == "" {
		return fmt.Errorf("No key found for the subscription")
	}
	decryptedEvent, err := ed.decryptEvent(eventSubmission.EncryptedEvent, key)
	if err != nil {
		fmt.Println("Error decrypting event: ", err)
	}
	event, err := xes.ParseXes(decryptedEvent)
	if err != nil {
		fmt.Println("Error parsing event: ", err)
		return err
	}
	fmt.Println("Event received: ", event)
	//TODO:here we should extract the attributes from the event according to the manifest
	*reply = "Event processed successfully"
	ed.EventChannel <- *event
	return nil
}

// StartRPCServer starts the RPC server
func (ed *EventDispatcher) StartRPCServer(addr string) {
	rpc.Register(ed)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Listener error:", err)
	}
	log.Println("Serving RPC server on port ", addr)
	rpc.Accept(listener)
	ed.SubscribeTo("localhost:6068")
}

func main() {
	addr := os.Args[1]
	eventChannel := make(chan xes.Event)
	psm := processStateManager.InitProcessStateManager(eventChannel)
	eventDispatcher := &EventDispatcher{EventChannel: eventChannel, Address: addr, Subscriptions: make(map[string][]attestation.Subscription)}
	go eventDispatcher.StartRPCServer(addr)
	eventDispatcher.SubscribeTo("localhost:6869")
	psm.WaitForEvents()
}

// Check for subscription timeouts
func (ed *EventDispatcher) checkTimeouts() {

}

func (ed *EventDispatcher) SubscribeTo(address string) {
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		log.Fatalf(err.Error())
	}
	var reply string
	err = client.Call("ProcessStateAgent.Subscribe", ed.Address, &reply)
	if err != nil {
		log.Fatalf("Error calling RequestSubscription: %v", err)
	}
	if reply == "subscription denied" {
		log.Fatalf("Subscription failed: %v", reply)
	} else {

		//parse the response into a subscription object
		subscription := attestation.Subscription{}
		err = json.Unmarshal([]byte(reply), &subscription)
		if err != nil {
			log.Fatalf("Error unmarshalling subscription: %v", err)
		}
		subscription.ClientConnection = client
		//chek if the subscription exists
		if _, ok := ed.Subscriptions[subscription.AgentAddress]; !ok {
			ed.Subscriptions[subscription.AgentAddress] = make([]attestation.Subscription, 0)
		}
		//add the subscription to the list of subscriptions
		ed.Subscriptions[subscription.AgentAddress] = append(ed.Subscriptions[subscription.AgentAddress], subscription)
		go ed.sendHeartbeat(subscription.TimeInterval, subscription)
	}

}

// Golang routine that sleeps for a given time interval and then sends a heartbeat
func (ed *EventDispatcher) sendHeartbeat(interval int, subscription attestation.Subscription) {
	for {
		time.Sleep(time.Duration(interval) * time.Second)
		// Create a new evidence
		//TODO: THE CLOCK SHOULD BE THE SAME OF THE REPORT
		//Generate a symetric key for the evidence
		provisioningKey, err := ed.generateProvisioningKey()
		if err != nil {
			log.Fatalf("Error generating provisioning key: %v", err)
		}
		evidence := attestation.Evidence{Report: []byte("heartbeat"), Timestamp: time.Now().Unix(), SubscriptionId: subscription.Id, ProvisioningKey: provisioningKey}
		client := subscription.ClientConnection
		var reply string
		err = client.Call("ProcessStateAgent.ReceiveHeartbeat", evidence, &reply)
		if err != nil {
			log.Fatalf("Error calling ReceiveHeartbeat: %v", err)
		}
		if reply == "heartbeat received" {
			//Add the evidence to the subscription
			subscription.Heartbeats = append(subscription.Heartbeats, evidence)
			//Update the subscription in the slice
			for i, sub := range ed.Subscriptions[subscription.AgentAddress] {
				if sub.Id == subscription.Id {
					ed.Subscriptions[subscription.AgentAddress][i] = subscription
					break
				}
			}
		}
	}

}

// Function to generate a random symmetric AES key
func (ed *EventDispatcher) generateProvisioningKey() (string, error) {
	key := make([]byte, 32) // AES-256
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

func (ed *EventDispatcher) decryptEvent(eventString, key string) (string, error) {
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
	ciphertext, err := base64.StdEncoding.DecodeString(eventString)
	if err != nil {
		return "", err
	}
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

//Function that given a the address of the agent retrieve the provisioning key from the last heartbeat
