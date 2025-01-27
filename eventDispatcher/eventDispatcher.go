package eventDispatcher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/edgelesssys/ego/ecrypto"
	"github.com/edgelesssys/ego/enclave"
	"log"
	"net"
	"net/rpc"
	"time"

	attestation "main/utils/attestation"
	"main/utils/eventsubmission"
	"main/utils/xes"
)

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
	newEvidence := attestation.Evidence{Report: encodedReport, Timestamp: time.Now().Unix(), ProvisioningKey: provisioningKey}
	*reply = newEvidence
	return nil
}

// SendEvent method to handle incoming events in string format
func (ed *EventDispatcher) SendEvent(eventSubmission eventsubmission.EventSubmission, reply *string) error {
	var subscriptionHeartbeats []attestation.Evidence
	for _, subscription := range ed.Subscriptions[eventSubmission.AgentReference] {
		if len(subscription.Heartbeats) > 0 {
			subscriptionHeartbeats = subscription.Heartbeats
			break
		}
	}

	if len(subscriptionHeartbeats) == 0 {
		return fmt.Errorf("No heartbeats found for the subscription")
	}

	decryptedEvent, err := ed.decryptEvent(eventSubmission.EncryptedEvent, subscriptionHeartbeats)
	if err != nil {
		fmt.Println("Error decrypting event: ", err)
		return errors.New("Error decrypting event: " + err.Error())
	}

	event, err := xes.ParseXes(decryptedEvent)
	if err != nil {
		fmt.Println("Error parsing event: ", err)
		return errors.New("Error parsing event:  " + err.Error())
	}

	parsedEvent := xes.Event{
		ActivityID: event.ActivityID,
		CaseID:     event.CaseID,
		Timestamp:  event.Timestamp,
		Attributes: make(map[string]interface{}),
	}

	if extractors, ok := ed.AttributeExtractors[event.ActivityID]; ok {
		for _, attrName := range extractors.([]interface{}) {
			if val, exists := event.Attributes[attrName.(string)]; exists {
				parsedEvent.Attributes[attrName.(string)] = val
			}
		}
	}

	if _, testCounterExists := event.Attributes["ESG_test_counter"]; testCounterExists {
		parsedEvent.Attributes["ESG_test_counter"] = event.Attributes["ESG_test_counter"]
	}

	if ed.ExternalQueryClient != nil {
		eventString, err := xes.Stringify(parsedEvent)
		if err != nil {
			return errors.New("Error parsing event to be put in the external queue: " + err.Error())
		}

		byteDecryptedEvent := []byte(eventString)
		sealedEvent, err := ecrypto.SealWithUniqueKey(byteDecryptedEvent, []byte(""))
		if err != nil {
			fmt.Println("Error sealing event: ", err)
		}

		var externalQueryReply string
		err = ed.ExternalQueryClient.Call("Queue.AddEvent", sealedEvent, &externalQueryReply)
		if err != nil {
			fmt.Println("Error calling external query server: ", err)
		}
	} else {
		ed.EventChannel <- parsedEvent
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

// SubscribeTo method to subscribe to a process state agent
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
	}

	subscription := attestation.Subscription{}
	err = json.Unmarshal([]byte(reply), &subscription)
	if err != nil {
		log.Fatalf("Error unmarshalling subscription: %v", err)
	}

	subscription.ClientConnection = client

	if _, ok := ed.Subscriptions[subscription.AgentAddress]; !ok {
		ed.Subscriptions[subscription.AgentAddress] = make([]attestation.Subscription, 0)
	}

	ed.Subscriptions[subscription.AgentAddress] = append(ed.Subscriptions[subscription.AgentAddress], subscription)
	go ed.sendHeartbeat(subscription.TimeInterval, subscription)
}

// sendHeartbeat sends periodic heartbeats for a subscription
func (ed *EventDispatcher) sendHeartbeat(interval int, subscription attestation.Subscription) {
	for {
		time.Sleep((time.Duration(interval) * time.Second) - (time.Duration(2) * time.Second))
		fmt.Println("Sending attestation heartbeat after ", interval-2, "seconds")

		provisioningKey, err := ed.generateProvisioningKey()
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		var report []byte
		if !ed.IsInSimulation {
			report, err = enclave.GetRemoteReport(provisioningKey)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}

		encodedReport := base64.StdEncoding.EncodeToString(report[:])
		evidence := attestation.Evidence{
			Report:          encodedReport,
			Timestamp:       time.Now().Unix(),
			SubscriptionId:  subscription.Id,
			ProvisioningKey: provisioningKey,
		}

		client := subscription.ClientConnection
		var reply string
		err = client.Call("ProcessStateAgent.ReceiveHeartbeat", evidence, &reply)
		if err != nil {
			log.Fatalf("Error calling ReceiveHeartbeat: %v", err)
		}

		if reply == "heartbeat received" {
			// Modify heartbeat recording to keep at most two heartbeats
			for i, sub := range ed.Subscriptions[subscription.AgentAddress] {
				if sub.Id == subscription.Id {
					if len(sub.Heartbeats) >= 2 {
						// Remove the first (oldest) heartbeat
						sub.Heartbeats = sub.Heartbeats[1:]
					}
					// Append the new heartbeat
					sub.Heartbeats = append(sub.Heartbeats, evidence)

					// Update the subscription in the slice
					ed.Subscriptions[subscription.AgentAddress][i] = sub
					break
				}
			}
		}
	}
}

// generateProvisioningKey creates a random symmetric AES key
func (ed *EventDispatcher) generateProvisioningKey() ([]byte, error) {
	key := make([]byte, 32) // AES-256
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// decryptEvent attempts to decrypt an event using available heartbeat keys
func (ed *EventDispatcher) decryptEvent(eventString string, subscriptionHeartbeats []attestation.Evidence) (string, error) {
	if len(subscriptionHeartbeats) == 0 {
		return "", fmt.Errorf("no heartbeats available for decryption")
	}

	// Try decryption with the most recent heartbeat key first
	lastHeartbeatIndex := len(subscriptionHeartbeats) - 1
	lastKey := subscriptionHeartbeats[lastHeartbeatIndex].ProvisioningKey

	decryptedEvent, err := ed.tryDecryptWithKey(eventString, lastKey)
	if err == nil {
		return decryptedEvent, nil
	}

	// If first decryption fails and we have more than one heartbeat, try the previous one
	if len(subscriptionHeartbeats) > 1 {
		previousKey := subscriptionHeartbeats[lastHeartbeatIndex-1].ProvisioningKey

		decryptedEvent, err := ed.tryDecryptWithKey(eventString, previousKey)
		if err == nil {
			return decryptedEvent, nil
		}
	}

	// If all decryption attempts fail, return the last error
	return "", fmt.Errorf("failed to decrypt event with available keys: %v", err)
}

// tryDecryptWithKey attempts to decrypt using a specific key
func (ed *EventDispatcher) tryDecryptWithKey(eventString string, key []byte) (string, error) {
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
