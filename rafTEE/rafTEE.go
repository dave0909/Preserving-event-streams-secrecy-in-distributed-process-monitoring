package rafTEE

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path"
	"sync"
	//sync "github.com/sasha-s/go-deadlock"
	"time"
)

func Assert[T comparable](msg string, a, b T) {
	if a != b {
		panic(fmt.Sprintf("%s. Got a = %#v, b = %#v", msg, a, b))
	}
}

type RafTEEserver struct {
	// These variables for shutting down.
	done   bool
	server *http.Server

	Debug bool

	mu sync.Mutex
	// ----------- PERSISTENT STATE -----------

	// The current term
	currentTerm uint64

	log []Entry

	// votedFor is stored in `cluster []ClusterMember` below,
	// mapped by `clusterIndex` below

	// ----------- READONLY STATE -----------

	// Unique identifier for this RafTEEserver
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
	state ServerState

	// Servers in the cluster, including this one
	cluster []ClusterMember

	// Index of this server
	clusterIndex int
}

func min[T ~int | ~uint64](a, b T) T {
	if a < b {
		return a
	}

	return b
}

func max[T ~int | ~uint64](a, b T) T {
	if a > b {
		return a
	}

	return b
}

func (s *RafTEEserver) debugmsg(msg string) string {
	return fmt.Sprintf("%s [Id: %d, Term: %d] %s", time.Now().Format(time.RFC3339Nano), s.id, s.currentTerm, msg)
}

func (s *RafTEEserver) debug(msg string) {
	if !s.Debug {
		return
	}
	fmt.Println(s.debugmsg(msg))
}

func (s *RafTEEserver) debugf(msg string, args ...any) {
	if !s.Debug {
		return
	}

	s.debug(fmt.Sprintf(msg, args...))
}

func (s *RafTEEserver) warn(msg string) {
	fmt.Println("[WARN] " + s.debugmsg(msg))
}

func (s *RafTEEserver) warnf(msg string, args ...any) {
	fmt.Println(fmt.Sprintf(msg, args...))
}

func Server_assert[T comparable](s *RafTEEserver, msg string, a, b T) {
	Assert(s.debugmsg(msg), a, b)
}

// Instantiate a new RafTEEserver
func NewRafTEEServer(
	clusterConfig []ClusterMember,
	statemachine StateMachine,
	metadataDir string,
	clusterIndex int,
) *RafTEEserver {
	//sync.Opts.DeadlockTimeout = 2000 * time.Millisecond
	// Explicitly make a copy of the cluster because we'll be
	// modifying it in this server.
	var cluster []ClusterMember
	//Check that no server has the 0 identifier
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
		heartbeatMs:  2000,
		mu:           sync.Mutex{},
	}
}

const PAGE_SIZE = 4096
const ENTRY_HEADER = 16
const ENTRY_SIZE = 128

// Weird thing to note is that writing to a deleted disk is not an
// error on Linux. So if these files are deleted, you won't know about
// that until the process restarts.
//
// Must be called within s.mu.Lock()
// Function used to save the persistent infor of the state in the local disk
func (s *RafTEEserver) persist(writeLog bool, nNewEntries int) {
	//Get the current time
	t := time.Now()
	//If there are no entries and the writeLog is true...
	if nNewEntries == 0 && writeLog {
		//newEntries is the length of the log
		nNewEntries = len(s.log)
	}

	s.fd.Seek(0, 0)
	s.Debug = true
	var page [PAGE_SIZE]byte
	// Bytes 0  - 8:   Current term
	// Bytes 8  - 16:  Voted for
	// Bytes 16 - 24:  Log length
	// Bytes 4096 - N: Log

	binary.LittleEndian.PutUint64(page[:8], s.currentTerm)
	binary.LittleEndian.PutUint64(page[8:16], s.getVotedFor())
	binary.LittleEndian.PutUint64(page[16:24], uint64(len(s.log)))
	n, err := s.fd.Write(page[:])
	if err != nil {
		panic(err)
	}
	Server_assert(s, "Wrote full page", n, PAGE_SIZE)
	//If write log is set to true and nNewEntries is > 0
	if writeLog && nNewEntries > 0 {
		//Write the persistent part of the state in fd
		newLogOffset := max(len(s.log)-nNewEntries, 0)

		s.fd.Seek(int64(PAGE_SIZE+ENTRY_SIZE*newLogOffset), 0)
		bw := bufio.NewWriter(s.fd)

		var entryBytes [ENTRY_SIZE]byte
		for i := newLogOffset; i < len(s.log); i++ {
			// Bytes 0 - 8:    Entry term
			// Bytes 8 - 16:   Entry command length
			// Bytes 16 - ENTRY_SIZE: Entry command

			if len(s.log[i].Command) > ENTRY_SIZE-ENTRY_HEADER {
				panic(fmt.Sprintf("Command is too large (%d). Must be at most %d bytes.", len(s.log[i].Command), ENTRY_SIZE-ENTRY_HEADER))
			}

			binary.LittleEndian.PutUint64(entryBytes[:8], s.log[i].Term)
			binary.LittleEndian.PutUint64(entryBytes[8:16], uint64(len(s.log[i].Command)))
			copy(entryBytes[16:], []byte(s.log[i].Command))

			n, err := bw.Write(entryBytes[:])
			if err != nil {
				panic(err)
			}
			Server_assert(s, "Wrote full page", n, ENTRY_SIZE)
		}

		err = bw.Flush()
		if err != nil {
			panic(err)
		}
	}

	if err = s.fd.Sync(); err != nil {
		panic(err)
	}
	s.debugf("Persisted in %s. Term: %d. Log Len: %d (%d new). Voted For: %d.", time.Now().Sub(t), s.currentTerm, len(s.log), nNewEntries, s.getVotedFor())
}

// If the log is empty, append an empty entry
func (s *RafTEEserver) ensureLog() {
	if len(s.log) == 0 {
		// Always has at least one log entry.
		s.log = append(s.log, Entry{})
	}
}

// Must be called within s.mu.Lock()
// Set the votedFor field of the server to the id passed as parameter
func (s *RafTEEserver) setVotedFor(id uint64) {
	for i := range s.cluster {
		if i == s.clusterIndex {
			s.cluster[i].votedFor = id
			return
		}
	}

	Server_assert(s, "Invalid cluster", true, false)
}

// Must be called within s.mu.Lock()
// Get the votedFor field of the server
func (s *RafTEEserver) getVotedFor() uint64 {
	for i := range s.cluster {
		if i == s.clusterIndex {
			return s.cluster[i].votedFor
		}
	}

	Server_assert(s, "Invalid cluster", true, false)
	return 0
}

// Get the name of the metadata file
func (s *RafTEEserver) Metadata() string {
	return fmt.Sprintf("md_%d.dat", s.id)
}

// Restore the persistent part of the state from the metadata file in the persistent memory
// Called by the Start() function
func (s *RafTEEserver) restore() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.fd == nil {
		var err error
		s.fd, err = os.OpenFile(
			path.Join(s.metadataDir, s.Metadata()),
			os.O_SYNC|os.O_CREATE|os.O_RDWR,
			0755)
		if err != nil {
			panic(err)
		}
	}

	s.fd.Seek(0, 0)

	// Bytes 0  - 8:   Current term
	// Bytes 8  - 16:  Voted for
	// Bytes 16 - 24:  Log length
	// Bytes 4096 - N: Log
	var page [PAGE_SIZE]byte
	n, err := s.fd.Read(page[:])
	if err == io.EOF {
		s.ensureLog()
		return
	} else if err != nil {
		panic(err)
	}
	Server_assert(s, "Read full page", n, PAGE_SIZE)
	s.currentTerm = binary.LittleEndian.Uint64(page[:8])
	s.setVotedFor(binary.LittleEndian.Uint64(page[8:16]))
	lenLog := binary.LittleEndian.Uint64(page[16:24])
	s.log = nil
	if lenLog > 0 {
		s.fd.Seek(int64(PAGE_SIZE), 0)

		var e Entry
		for i := 0; uint64(i) < lenLog; i++ {
			var entryBytes [ENTRY_SIZE]byte
			n, err := s.fd.Read(entryBytes[:])
			if err != nil {
				panic(err)
			}
			Server_assert(s, "Read full entry", n, ENTRY_SIZE)

			// Bytes 0 - 8:    Entry term
			// Bytes 8 - 16:   Entry command length
			// Bytes 16 - ENTRY_SIZE: Entry command
			e.Term = binary.LittleEndian.Uint64(entryBytes[:8])
			lenValue := binary.LittleEndian.Uint64(entryBytes[8:16])
			e.Command = entryBytes[16 : 16+lenValue]
			s.log = append(s.log, e)
		}
	}

	s.ensureLog()
}

// Function used to request votes to the other nddes of the clusterm when the node is in Candidate state
// Called by the timeout() function, once the timeout has occourred
func (s *RafTEEserver) requestVote() {
	//For each node of the cluster
	for i := range s.cluster {
		//If the node is the current node, continue
		if i == s.clusterIndex {
			continue
		}
		//Go routine to request the vote to the node in parallel
		go func(i int) {
			s.mu.Lock()

			s.debugf("Requesting vote from %d.", s.cluster[i].Id)
			//Get the last log index and term
			lastLogIndex := uint64(len(s.log) - 1)
			lastLogTerm := s.log[len(s.log)-1].Term
			//Create the request
			req := RequestVoteRequest{
				RPCMessage: RPCMessage{
					Term: s.currentTerm,
				},
				CandidateId:  s.id,
				LastLogIndex: lastLogIndex,
				LastLogTerm:  lastLogTerm,
			}
			s.mu.Unlock()
			//Create the response
			var rsp RequestVoteResponse
			//Call the rpcCall function to request the vote
			ok := s.rpcCall(i, "RafTEEserver.HandleRequestVoteRequest", req, &rsp)
			if !ok {
				// Will retry later
				return
			}

			s.mu.Lock()
			defer s.mu.Unlock()
			//If the term of the response is greater than the current term of the server, return
			if s.updateTerm(rsp.RPCMessage) {
				return
			}
			//If the term of the response is different from the term of the request, return
			dropStaleResponse := rsp.Term != req.Term
			if dropStaleResponse {
				return
			}
			//If the vote is granted, set the votedFor field of the server to the id of the node
			if rsp.VoteGranted {
				s.debugf("Vote granted by %d.", s.cluster[i].Id)
				//Set the votedFor field of the server to the id of the node
				s.cluster[i].votedFor = s.id
			}
		}(i)
	}
}

// Function that handles vote requests from candidate servers
func (s *RafTEEserver) HandleRequestVoteRequest(req RequestVoteRequest, rsp *RequestVoteResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	//Update the term of the server if necessary
	s.updateTerm(req.RPCMessage)

	s.debugf("Received vote request from %d.", req.CandidateId)
	//Set the voteGranted field of the response to false
	rsp.VoteGranted = false
	//Set the term of the response to the current term of the server
	rsp.Term = s.currentTerm
	//If the term of the request is less than the current term of the server, return
	if req.Term < s.currentTerm {
		s.debugf("Not granting vote request from %d.", req.CandidateId)
		Server_assert(s, "VoteGranted = false", rsp.VoteGranted, false)
		return nil
	}
	//Get the last log index and term
	lastLogTerm := s.log[len(s.log)-1].Term
	logLen := uint64(len(s.log) - 1)
	//Check if the last log term of the request is greater than the last log term of the server (checking if the log is up to date)
	logOk := req.LastLogTerm > lastLogTerm || //or
		//If the last log term of the request is equal to the last log term of the server and the last log index of the request is greater than the last log index of the server
		(req.LastLogTerm == lastLogTerm && req.LastLogIndex >= logLen)
	//Grant the vote if 1) the term of the request is equal to the current term of the server
	grant := req.Term == s.currentTerm &&
		//2)the log is ok (see above)
		logOk &&
		//3)the server has not voted for anyone or the server has already voted for the candidate
		(s.getVotedFor() == 0 || s.getVotedFor() == req.CandidateId)
	//If the vote is granted
	if grant {
		s.debugf("Voted for %d.", req.CandidateId)
		//Set the votedFor field of the server to the id of the candidate
		s.setVotedFor(req.CandidateId)
		//Set the voteGranted field of the response to true
		rsp.VoteGranted = true
		//Set the state of the server to Follower
		s.resetElectionTimeout()
		s.state = followerState
		s.persist(false, 0)
	} else {
		s.debugf("Not granting vote request from %d.", +req.CandidateId)
	}

	return nil
}

// Must be called within a s.mu.Lock()
// Function that updates the term of the server if necessary
func (s *RafTEEserver) updateTerm(msg RPCMessage) bool {
	transitioned := false
	//If the term of the message is greater than the current term of the server
	if msg.Term > s.currentTerm {
		//Update the term of the server
		s.currentTerm = msg.Term
		//Set the state of the server to Follower
		s.state = followerState
		s.setVotedFor(0)
		transitioned = true
		s.debug("Transitioned to follower")
		//Reset the election timeout
		s.resetElectionTimeout()
		//Persist the state of the server
		s.persist(false, 0)
	}
	return transitioned
}

// Function invoked by the client to apply the command and replicate it to the other nodes of the cluster
// It only works if the server is the leader
func (s *RafTEEserver) Apply(commands [][]byte) ([]ApplyResult, error) {
	s.mu.Lock()

	if s.state != leaderState {
		s.mu.Unlock()
		return nil, ErrApplyToLeader
	}
	s.debugf("Processing %d new entry!", len(commands))

	resultChans := make([]chan ApplyResult, len(commands))
	for i, command := range commands {
		resultChans[i] = make(chan ApplyResult)
		s.log = append(s.log, Entry{
			Term:    s.currentTerm,
			Command: command,
			result:  resultChans[i],
		})
	}

	s.persist(true, len(commands))

	s.debug("Waiting to be applied!")
	s.mu.Unlock()

	s.appendEntries()

	// TODO: What happens if this takes too long?
	results := make([]ApplyResult, len(commands))
	var wg sync.WaitGroup
	wg.Add(len(commands))
	for i, ch := range resultChans {
		go func(i int, c chan ApplyResult) {
			results[i] = <-c
			wg.Done()
		}(i, ch)
	}

	wg.Wait()

	return results, nil
}
func (s *RafTEEserver) HandleAppendEntriesRequest(req AppendEntriesRequest, rsp *AppendEntriesResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updateTerm(req.RPCMessage)
	fmt.Println(s.log)
	// From Candidates (§5.2) in Figure 2
	// If AppendEntries RPC received from new leader: convert to follower
	if req.Term == s.currentTerm && s.state == candidateState {
		s.state = followerState
	}

	rsp.Term = s.currentTerm
	rsp.Success = false

	if s.state != followerState {
		s.debugf("Non-follower cannot append entries.")
		return nil
	}

	if req.Term < s.currentTerm {
		s.debugf("Dropping request from old leader %d: term %d.", req.LeaderId, req.Term)
		// Not a valid leader.
		return nil
	}

	// Valid leader so reset election.
	s.resetElectionTimeout()

	logLen := uint64(len(s.log))
	validPreviousLog := req.PrevLogIndex == 0 /* This is the induction step */ ||
		(req.PrevLogIndex < logLen &&
			s.log[req.PrevLogIndex].Term == req.PrevLogTerm)
	if !validPreviousLog {
		s.debug("Not a valid log.")
		return nil
	}

	next := req.PrevLogIndex + 1
	nNewEntries := 0

	for i := next; i < next+uint64(len(req.Entries)); i++ {
		e := req.Entries[i-next]
		if i >= uint64(cap(s.log)) {
			newTotal := next + uint64(len(req.Entries))
			// Second argument must actually be `i`
			// not `0` otherwise the copy after this
			// doesn't work.
			// Only copy until `i`, not `newTotal` since
			// we'll continue appending after this.
			newLog := make([]Entry, i, newTotal*2)
			copy(newLog, s.log)
			s.log = newLog
		}

		if i < uint64(len(s.log)) && s.log[i].Term != e.Term {
			prevCap := cap(s.log)
			// If an existing entry conflicts with a new
			// one (same index but different terms),
			// delete the existing entry and all that
			// follow it (§5.3)
			s.log = s.log[:i]
			Server_assert(s, "Capacity remains the same while we truncated.", cap(s.log), prevCap)
		}

		s.debugf("Appending entry: %s. At index: %d.", string(e.Command), len(s.log))

		if i < uint64(len(s.log)) {
			Server_assert(s, "Existing log is the same as new log", s.log[i].Term, e.Term)
		} else {
			s.log = append(s.log, e)
			Server_assert(s, "Length is directly related to the index.", uint64(len(s.log)), i+1)
			nNewEntries++
		}
	}

	if req.LeaderCommit > s.commitIndex {
		s.commitIndex = min(req.LeaderCommit, uint64(len(s.log)-1))
	}

	s.persist(nNewEntries != 0, nNewEntries)

	rsp.Success = true
	return nil
}

var ErrApplyToLeader = errors.New("Cannot apply message to follower, apply to leader.")

func (s *RafTEEserver) rpcCall(i int, name string, req, rsp any) bool {
	s.mu.Lock()
	c := s.cluster[i]
	var err error
	var rpcClient *rpc.Client = c.rpcClient
	if c.rpcClient == nil {
		c.rpcClient, err = rpc.DialHTTP("tcp", c.Address)
		rpcClient = c.rpcClient
	}
	s.mu.Unlock()

	if err == nil {
		err = rpcClient.Call(name, req, rsp)
	}

	if err != nil {
		s.warnf("Error calling %s on %d: %s.", name, c.Id, err)
	}

	return err == nil
}

const MAX_APPEND_ENTRIES_BATCH = 8_000

// Called by the Apply() function to replicate the command to the other nodes of the cluster
func (s *RafTEEserver) appendEntries() {
	fmt.Println(s.log)
	for i := range s.cluster {
		// Don't need to send message to self
		if i == s.clusterIndex {
			continue
		}

		go func(i int) {
			s.mu.Lock()

			next := s.cluster[i].nextIndex
			prevLogIndex := next - 1
			prevLogTerm := s.log[prevLogIndex].Term

			var entries []Entry
			if uint64(len(s.log)-1) >= s.cluster[i].nextIndex {
				s.debugf("len: %d, next: %d, server: %d", len(s.log), next, s.cluster[i].Id)
				entries = s.log[next:]
			}

			// Keep latency down by only applying N at a time.
			if len(entries) > MAX_APPEND_ENTRIES_BATCH {
				entries = entries[:MAX_APPEND_ENTRIES_BATCH]
			}

			lenEntries := uint64(len(entries))
			req := AppendEntriesRequest{
				RPCMessage: RPCMessage{
					Term: s.currentTerm,
				},
				LeaderId:     s.cluster[s.clusterIndex].Id,
				PrevLogIndex: prevLogIndex,
				PrevLogTerm:  prevLogTerm,
				Entries:      entries,
				LeaderCommit: s.commitIndex,
			}

			s.mu.Unlock()

			var rsp AppendEntriesResponse
			s.debugf("Sending %d entries to %d for term %d.", len(entries), s.cluster[i].Id, req.Term)
			ok := s.rpcCall(i, "RafTEEserver.HandleAppendEntriesRequest", req, &rsp)
			if !ok {
				// Will retry next tick
				return
			}

			s.mu.Lock()
			defer s.mu.Unlock()
			if s.updateTerm(rsp.RPCMessage) {
				return
			}

			dropStaleResponse := rsp.Term != req.Term && s.state == leaderState
			if dropStaleResponse {
				return
			}

			if rsp.Success {
				prev := s.cluster[i].nextIndex
				s.cluster[i].nextIndex = max(req.PrevLogIndex+lenEntries+1, 1)
				s.cluster[i].matchIndex = s.cluster[i].nextIndex - 1
				s.debugf("Messages (%d) accepted for %d. Prev Index: %d, Next Index: %d, Match Index: %d.", len(req.Entries), s.cluster[i].Id, prev, s.cluster[i].nextIndex, s.cluster[i].matchIndex)
			} else {
				s.cluster[i].nextIndex = max(s.cluster[i].nextIndex-1, 1)
				s.debugf("Forced to go back to %d for: %d.", s.cluster[i].nextIndex, s.cluster[i].Id)
			}
		}(i)
	}
}

// Function that advances the commit index of the server
// Called in the main loop if the server is either the leader or a follower
func (s *RafTEEserver) advanceCommitIndex() {
	s.mu.Lock()
	defer s.mu.Unlock()

	//If the server is the leader, it advances the commit index of the whole cluster
	if s.state == leaderState {
		//Get the last log index of the local server
		lastLogIndex := uint64(len(s.log) - 1)
		//For each uncommited log entry
		for i := lastLogIndex; i > s.commitIndex; i-- {
			//Compute the quorum of the cluster
			quorum := len(s.cluster)/2 + 1
			//For each node of the cluster
			for j := range s.cluster {
				//If the quorum is 0, break
				if quorum == 0 {
					break
				}
				//if the node is the leader
				isLeader := j == s.clusterIndex
				// if the node is the leader or the match index of the node is greater than or equal to i
				//The match index of the node is the highest indexed log entry known to be replicated
				if isLeader || s.cluster[j].matchIndex >= i {
					//Decrement the quorum since the node has replicated the log entry i in its log
					quorum--
				}
			}
			//If a quorum of servers have replicated the log entry i, set the commit index of the server to i
			if quorum == 0 {
				//Set the commit index of the server to i, since it is safe to apply it to the state machine
				s.commitIndex = i
				s.debugf("New commit index: %d.", i)
				break
			}
		}
	}
	// for every state a server might be in, if there are messages committed but not applied, we'll apply one here.
	//And importantly, we'll pass the result back to the message's result channel if it exists, so that s.Apply() can learn about the result.
	//If the last applied entry is less or equal to the commit index of the server
	if s.lastApplied <= s.commitIndex {
		//Get the last applied entry of the log
		log := s.log[s.lastApplied]

		// len(log.Command) == 0 is a noop committed by the leader.
		//If the Command is not a default flag value
		if len(log.Command) > 0 {
			s.debugf("Entry applied: %d.", s.lastApplied)
			// TODO: what if Apply() takes too long?
			//Apply the command to the state machine and get the result (This command is not the omonymous function of the server, but the command state machine)
			res, err := s.statemachine.Apply(log.Command)

			// Will be nil for follower entries and for no-op entries.
			// Not nil for all user submitted messages.
			//If the result channel of the log entry is not nil
			if log.result != nil {
				//Send the result of the application to the result channel of the log entry
				log.result <- ApplyResult{
					Result: res,
					Error:  err,
				}
			}
		}
		//Increment the last applied entry of the server (i.e., the last entry that has been applied to the state machine)
		s.lastApplied++
	}
}

// Must be called within a s.mu.Lock()
func (s *RafTEEserver) resetElectionTimeout() {
	interval := time.Duration(rand.Intn(s.heartbeatMs*2) + s.heartbeatMs*2)
	s.debugf("New interval: %s.", interval*time.Millisecond)
	s.electionTimeout = time.Now().Add(interval * time.Millisecond)
}

func (s *RafTEEserver) timeout() {
	s.mu.Lock()
	defer s.mu.Unlock()

	hasTimedOut := time.Now().After(s.electionTimeout)
	if hasTimedOut {
		s.debug("Timed out, starting new election.")
		s.state = candidateState
		s.currentTerm++
		for i := range s.cluster {
			if i == s.clusterIndex {
				s.cluster[i].votedFor = s.id
			} else {
				s.cluster[i].votedFor = 0
			}
		}

		s.resetElectionTimeout()
		s.persist(false, 0)
		s.requestVote()
	}
}

func (s *RafTEEserver) becomeLeader() {
	s.mu.Lock()
	defer s.mu.Unlock()

	quorum := len(s.cluster)/2 + 1
	for i := range s.cluster {
		if s.cluster[i].votedFor == s.id && quorum > 0 {
			quorum--
		}
	}

	if quorum == 0 {
		// Reset all cluster state
		for i := range s.cluster {
			s.cluster[i].nextIndex = uint64(len(s.log) + 1)
			// Yes, even matchIndex is reset. Figure 2
			// from Raft shows both nextIndex and
			// matchIndex are reset after every election.
			s.cluster[i].matchIndex = 0
		}

		s.debug("New leader.")
		s.state = leaderState

		// From Section 8 Client Interaction:
		// > First, a leader must have the latest information on
		// > which entries are committed. The Leader
		// > Completeness Property guarantees that a leader has
		// > all committed entries, but at the start of its
		// > term, it may not know which those are. To find out,
		// > it needs to commit an entry from its term. Raft
		// > handles this by having each leader commit a blank
		// > no-op entry into the log at the start of its term.
		s.log = append(s.log, Entry{Term: s.currentTerm, Command: nil})
		s.persist(true, 1)

		// Triggers s.appendEntries() in the next tick of the
		// main state loop.
		s.heartbeatTimeout = time.Now()
	}
}

func (s *RafTEEserver) heartbeat() {
	s.mu.Lock()
	defer s.mu.Unlock()
	timeForHeartbeat := time.Now().After(s.heartbeatTimeout)
	if timeForHeartbeat {
		s.heartbeatTimeout = time.Now().Add(time.Duration(s.heartbeatMs) * time.Millisecond)
		s.debug("Sending heartbeat")
		s.appendEntries()
	}
}

// Make sure rand is seeded
func (s *RafTEEserver) Start() {
	s.mu.Lock()
	s.state = followerState
	s.done = false
	s.mu.Unlock()

	s.restore()

	rpcServer := rpc.NewServer()
	rpcServer.Register(s)
	l, err := net.Listen("tcp", s.address)
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.Handle(rpc.DefaultRPCPath, rpcServer)

	s.server = &http.Server{Handler: mux}
	go s.server.Serve(l)

	go func() {
		s.mu.Lock()
		s.resetElectionTimeout()
		s.mu.Unlock()

		for {
			s.mu.Lock()
			if s.done {
				s.mu.Unlock()
				return
			}
			state := s.state
			s.mu.Unlock()

			switch state {
			case leaderState:
				//Check if the server should send and heartbeat and send it if necessary
				s.heartbeat()
				//Advance the commit index of the server and cluster
				s.advanceCommitIndex()
			case followerState:
				s.timeout()
				s.advanceCommitIndex()
			case candidateState:
				s.timeout()
				s.becomeLeader()
			}
		}
	}()
}
