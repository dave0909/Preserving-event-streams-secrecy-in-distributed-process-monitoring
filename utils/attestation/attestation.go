package attestation

import (
	"net/rpc"
)

// Subscription struct
type Subscription struct {
	Id               int
	Nonce            string
	AgentAddress     string
	AgentPublicKey   []byte
	Heartbeats       []Evidence
	IsActive         bool
	TimeInterval     int
	ClientConnection *rpc.Client
}

// Evidence struct
type Evidence struct {
	//Report of the attestation
	Report string
	//Timestamp of the evidence. It is also part of the unsigned report
	Timestamp int64
	//SubscriptionId
	SubscriptionId int
	//Provisioning key
	ProvisioningKey []byte
}
