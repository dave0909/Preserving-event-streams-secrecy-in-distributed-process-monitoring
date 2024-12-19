package eventDispatcher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/edgelesssys/ego/ecrypto"
	"github.com/edgelesssys/ego/enclave"
	"net"

	"log"
	attestation "main/utils/attestation"
	"main/utils/eventsubmission"
	"main/utils/xes"
	"net/rpc"
	"time"
)

// go run eventDispatcher.go localhost:6969
// EventDispatcher struct
type EventDispatcher struct {
	EventChannel        chan xes.Event
	Subscriptions       map[string][]attestation.Subscription
	Address             string
	AttributeExtractors map[string]interface{}
	IsInSimulation      bool
	ExternalQueryClient *rpc.Client
}

// GetEvidence method to get ProcessVault's evidence
func (ed *EventDispatcher) GetEvidence(nonce string, reply *attestation.Evidence) error {
	provisioningKey, err := ed.generateProvisioningKey()
	if err != nil {
		return err
	}
	var report []byte
	if !ed.IsInSimulation {
		report, err = enclave.GetRemoteReport(provisioningKey)
		if err != nil {
			return err
		}
	}
	encodedReport := base64.StdEncoding.EncodeToString(report[:])
	if err != nil {
		return err
	}
	//TODO: REMOVE THE PROVISIONING KEY FROM THE TRANSMITTED REPORT
	newEvidence := attestation.Evidence{Report: encodedReport, Timestamp: time.Now().Unix(), ProvisioningKey: provisioningKey}
	*reply = newEvidence
	return nil
}

// SendEvent method to handle incoming events in string format
func (ed *EventDispatcher) SendEvent(eventSubmission eventsubmission.EventSubmission, reply *string) error {
	//The reveiveing event should arrive as an encrypted message with the Provisioning generated for the last heartbeat
	// Parse the event string
	//fmt.Println("Received event: ", eventSubmission.EncryptedEvent)
	//Find the key for the subscription that sent the event
	var key []byte
	for _, subscription := range ed.Subscriptions[eventSubmission.AgentReference] {
		if len(subscription.Heartbeats) > 0 {
			key = subscription.Heartbeats[len(subscription.Heartbeats)-1].ProvisioningKey
			break
		}
	}
	if key == nil {
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
	if ed.ExternalQueryClient != nil {
		// Send the event to the external query server
		var externalQueryReply string
		//Convert decrypted event to a byte array
		byteDecryptedEvent := []byte(decryptedEvent)
		sealedEvent, err := ecrypto.SealWithUniqueKey(byteDecryptedEvent, []byte(""))
		if err != nil {
			fmt.Println("Error sealing event: ", err)
		}
		err = ed.ExternalQueryClient.Call("Queue.AddEvent", sealedEvent, &externalQueryReply)
		if err != nil {
			fmt.Println("Error calling external query server: ", err)
		}
	} else {
		ed.EventChannel <- *event
	}
	*reply = "Event processed successfully"
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
}

//func main() {
//	addr := os.Args[1]
//	eventChannel := make(chan xes.Event)
//	psm := processStateManager.InitProcessStateManager(eventChannel)
//	eventDispatcher := &EventDispatcher{EventChannel: eventChannel, Address: addr, Subscriptions: make(map[string][]attestation.Subscription)}
//	go eventDispatcher.StartRPCServer(addr)
//	eventDispatcher.SubscribeTo("localhost:6869")
//	psm.WaitForEvents()
//}

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
			fmt.Println(err.Error())
			return
		}
		//Maybe we need the byte version of the key, not formatted in this way
		var report []byte
		if !ed.IsInSimulation {
			report, err = enclave.GetRemoteReport(provisioningKey)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}
		encodedReport := base64.StdEncoding.EncodeToString(report[:])
		if err != nil {
			fmt.Println(err.Error())
		}
		//TODO: REMOVE THE PROVISIONING KEY FROM THE TRANSMITTED REPORT
		evidence := attestation.Evidence{Report: encodedReport, Timestamp: time.Now().Unix(), SubscriptionId: subscription.Id, ProvisioningKey: provisioningKey}
		client := subscription.ClientConnection
		var reply string
		err = client.Call("ProcessStateAgent.ReceiveHeartbeat", evidence, &reply)
		if err != nil {
			log.Fatalf("Error calling ReceiveHeartbeat: %v", err)
		}
		if reply == "heartbeat received" {
			//Add the evidence to the subscription
			//subscription.Heartbeats = subscription.Heartbeats[:0]
			subscription.Heartbeats = nil
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
func (ed *EventDispatcher) generateProvisioningKey() ([]byte, error) {
	key := make([]byte, 32) // AES-256
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (ed *EventDispatcher) decryptEvent(eventString string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
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
