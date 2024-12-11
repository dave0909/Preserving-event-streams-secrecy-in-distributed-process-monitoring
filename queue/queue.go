package main

import (
	"errors"
	"net"
	"net/rpc"
	"os"
)

//Go run queue.go localhost:8387

// Queue struct with a queue field
type Queue struct {
	Queue [][]byte
}

// AddEvent method to add a string into the queue
func (q *Queue) AddEvent(event []byte, reply *string) error {
	q.Queue = append(q.Queue, event)
	*reply = "Event added successfully"
	return nil
}

// DequeueEvent method to dequeue the event
func (q *Queue) DequeueEvent(_ []byte, reply *[]byte) error {
	if len(q.Queue) == 0 {
		return errors.New("queue is empty")
	}
	*reply = q.Queue[0]
	q.Queue = q.Queue[1:]
	return nil
}

// StartRPCServer starts the RPC server
func StartRPCServer(addr string) error {
	queue := new(Queue)
	err := rpc.Register(queue)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()
	rpc.Accept(listener)
	return nil
}

// main function to start the server
func main() {
	addr := os.Args[1]
	StartRPCServer(addr)
}

//func main() {
//	// Connect to the RPC server
//	client, err := rpc.Dial("tcp", "localhost:1234")
//	if err != nil {
//		log.Fatal("Dialing:", err)
//	}
//
//	// Add events to the queue
//	var reply string
//	events := []string{"event1", "event2", "event3"}
//	for _, event := range events {
//		err = client.Call("Queue.AddEvent", event, &reply)
//		if err != nil {
//			log.Fatal("AddEvent error:", err)
//		}
//		fmt.Println(reply)
//	}
//
//	// Dequeue events from the queue
//	for range events {
//		var dequeuedEvent string
//		err = client.Call("Queue.DequeueEvent", "", &dequeuedEvent)
//		if err != nil {
//			log.Fatal("DequeueEvent error:", err)
//		}
//		fmt.Println("Dequeued event:", dequeuedEvent)
//	}
//}
