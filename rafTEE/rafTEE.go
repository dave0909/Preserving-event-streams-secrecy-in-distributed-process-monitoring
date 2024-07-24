package rafTEE

import (
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

// Interface for a general state machine
type StateMachine interface {
	Apply(cmd []byte) ([]byte, error)
}

// Result application
type ApplyResult struct {
	Result []byte
	Error  error
}

type Entry struct {
	//Term of the entry
	Term uint64
	//Comand of the entry
	Command []byte
	//Set by the primary so it can learn about the result of applying this command to the state machine
	result chan ApplyResult
}

// A member of the rafTEE cluster
type ClusterMember struct {
	Id      uint64
	Address string
	// Index of the next log entry to send
	nextIndex uint64
	// Highest log entry known to be replicated
	matchIndex uint64
	// Who was voted for in the most recent term
	votedFor uint64
	// TCP connection
	rpcClient *rpc.Client
}

type RafTEEserverState string

const (
	leaderState    RafTEEserverState = "leader"
	followerState                    = "follower"
	candidateState                   = "candidate"
)

type RafTEEserver struct {
	// These variables for shutting down.
	done   bool
	server *http.Server
	Debug  bool
	mu     sync.Mutex
	// ----------- PERSISTENT STATE -----------
	// The current term
	currentTerm uint64
	log         []Entry
	// votedFor is stored in `cluster []ClusterMember` below,
	// mapped by `clusterIndex` below

	// ----------- READONLY STATE -----------

	// Unique identifier for this RafTEErpcService
	id uint64

	// The TCP address for RPC
	address string

	// When to start elections after no append entry messages
	electionTimeout time.Time

	// How often to send empty messages
	heartbeatMs int

	// When to next send empty message
	heartbeatTimeout time.Time

	// User-provided state machine
	statemachine StateMachine

	// Metadata directory
	metadataDir string

	// Metadata store
	fd *os.File

	// ----------- VOLATILE STATE -----------

	// Index of highest log entry known to be committed
	commitIndex uint64

	// Index of highest log entry applied to state machine
	lastApplied uint64

	// Candidate, follower, or leader
	state RafTEEserverState

	// Servers in the cluster, including this one
	cluster []ClusterMember

	// Index of this server
	clusterIndex int
}

// Constructor for the initial state
func NewRafTEEserver(
	//This is the cluster config passed through command line
	clusterConfig []ClusterMember,
	statemachine StateMachine,
	metadataDir string,
	clusterIndex int,
) *RafTEEserver {
	// Explicitly make a copy of the cluster because we'll be
	// modifying it in this server.
	var cluster []ClusterMember
	for _, c := range clusterConfig {
		if c.Id == 0 {
			panic("Id must not be 0.")
		}
		cluster = append(cluster, c)
	}
	return &RafTEEserver{
		id:           cluster[clusterIndex].Id,
		address:      cluster[clusterIndex].Address,
		cluster:      cluster,
		statemachine: statemachine,
		metadataDir:  metadataDir,
		clusterIndex: clusterIndex,
		heartbeatMs:  300,
		mu:           sync.Mutex{},
	}
}
