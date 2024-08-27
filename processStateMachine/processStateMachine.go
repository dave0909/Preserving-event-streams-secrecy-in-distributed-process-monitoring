package processStateMachine

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
)

type ProcessStateMachine struct {
	//TODO here we should have the process state (eg., a petri net)
	Db *sync.Map
	//Identifier of the server
	Server int
}

type commandKind uint8

const (
	//Set commands means that the client wants to set a key-value pair
	SetCommand commandKind = iota
	GetCommand
)

type Command struct {
	Kind  commandKind
	Key   string
	Value string
}

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

func DecodeCommand(msg []byte) Command {
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

// Function invoked by the advanceCommitIndex() method of the rafTEE component apply a command
func (s *ProcessStateMachine) Apply(cmd []byte) ([]byte, error) {
	c := DecodeCommand(cmd)
	switch c.Kind {
	case SetCommand:
		s.Db.Store(c.Key, c.Value)
	case GetCommand:
		value, ok := s.Db.Load(c.Key)
		if !ok {
			return nil, fmt.Errorf("Key not found")
		}
		return []byte(value.(string)), nil
	default:
		return nil, fmt.Errorf("Unknown command: %x", cmd)
	}

	return nil, nil
}
