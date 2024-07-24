package processStateMachine

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"main/rafTEE"
	"net/http"
	"sync"
)

// State machine type TODO here we should model the process state instead of db
type ProcessStateMachine struct {
	Db *sync.Map
	//TODO here we should have something like ...  processModel petrinet
	//This is the node identifier
	Server int
}

type commandKind uint8

const (
	//Set the state
	SetCommand commandKind = iota
	//Get the state
	GetCommand
)

// Command type
type Command struct {
	Kind  commandKind
	Key   string
	Value string
}

// It is the function that will be called by the raft rafTEE algorithm
func (s *ProcessStateMachine) Apply(cmd []byte) ([]byte, error) {
	c := decodeCommand(cmd)
	switch c.Kind {
	case SetCommand:
		s.Db.Store(c.Key, c.Value)
		//TODO: Here we should compute the new state of the process using processStateMachine.value. Replace occurrence of db.Store.
	case GetCommand:
		value, ok := s.Db.Load(c.Key)
		if !ok {
			return nil, fmt.Errorf("Key not found")
		}
		return []byte(value.(string)), nil
	default:
		return nil, fmt.Errorf("Unknown Command: %x", cmd)
	}

	return nil, nil
}

// Encode the command to be sent to the raft rafTEE algorithm as a byte array
func EncodeCommand(c Command) []byte {
	msg := bytes.NewBuffer(nil)
	err := msg.WriteByte(uint8(c.Kind))
	if err != nil {
		panic(err)
	}

	err = binary.Write(msg, binary.LittleEndian, uint64(len(c.Key)))
	if err != nil {
		panic(err)
	}

	msg.WriteString(c.Key)

	err = binary.Write(msg, binary.LittleEndian, uint64(len(c.Value)))
	if err != nil {
		panic(err)
	}

	msg.WriteString(c.Value)

	return msg.Bytes()
}

// Decode the command from byte
func decodeCommand(msg []byte) Command {
	var c Command
	c.Kind = commandKind(msg[0])

	keyLen := binary.LittleEndian.Uint64(msg[1:9])
	c.Key = string(msg[9 : 9+keyLen])

	if c.Kind == SetCommand {
		valLen := binary.LittleEndian.Uint64(msg[9+keyLen : 9+keyLen+8])
		c.Value = string(msg[9+keyLen+8 : 9+keyLen+8+valLen])
	}

	return c
}

// RafTEErpcService that handles read amd write requests to the cluster
type httpServer struct {
	raft *rafTEE.RafTEEserver
	db   *sync.Map
}

// Method used by clients to update the process state TODO: this should be invoked by the event generator to submit events
// TODO key value should not be sen t in a get request but in a post request
func (hs httpServer) setHandler(w http.ResponseWriter, r *http.Request) {
	var c Command
	c.Kind = SetCommand
	c.Key = r.URL.Query().Get("key")
	c.Value = r.URL.Query().Get("value")

	_, err := hs.raft.Apply([][]byte{EncodeCommand(c)})
	if err != nil {
		log.Printf("Could not write key-value: %s", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}
