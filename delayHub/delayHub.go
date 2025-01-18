package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"main/utils/delayargs"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"sync"
)

type DelayHub struct {
	mu         sync.Mutex
	eventStore map[int]int
	csvFile    *os.File
	csvWriter  *csv.Writer
}

func NewDelayHub() (*DelayHub, error) {
	//Check if the file exists, if not create it
	outputDir := "../data/output"
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		log.Fatalf("Error creating directory: %v", err)
	}
	// Open the CSV file for writing
	filePath := filepath.Join(outputDir, "delay_result.csv")
	fmt.Println("Creating file at: ", filePath)
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file: ", err)
		return nil, err
	}
	writer := csv.NewWriter(file)
	return &DelayHub{
		eventStore: make(map[int]int),
		csvFile:    file,
		csvWriter:  writer,
	}, nil
}

func (d *DelayHub) WriteArrival(args *delayargs.ArrivalArgs, reply *bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, exists := d.eventStore[args.EventCode]; exists {
		return errors.New("event code already exists")
	}
	fmt.Println("Writing arrival for event code: ", args.EventCode, " with timestamp: ", args.ArrivalTimestamp)
	d.eventStore[args.EventCode] = args.ArrivalTimestamp
	*reply = true
	return nil
}

func (d *DelayHub) WriteCompletion(args *delayargs.CompletionArgs, reply *bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	arrivalTimestamp, exists := d.eventStore[args.EventCode]
	if !exists {
		return errors.New("event code not found")
	}

	record := []string{
		fmt.Sprintf("%d", args.EventCode),
		fmt.Sprintf("%d", arrivalTimestamp),
		fmt.Sprintf("%d", args.CompletionTimestamp),
	}

	if err := d.csvWriter.Write(record); err != nil {
		return err
	}
	d.csvWriter.Flush()
	if err := d.csvWriter.Error(); err != nil {
		return err
	}

	delete(d.eventStore, args.EventCode)
	fmt.Println("Writing completion for event code: ", args.EventCode, " with timestamp: ", args.CompletionTimestamp)
	*reply = true
	return nil
}

func main() {
	addr := os.Args[1]
	delayHub, err := NewDelayHub()
	if err != nil {
		panic(err)
	}
	defer delayHub.csvFile.Close()
	err = rpc.Register(delayHub)
	if err != nil {
		panic(err)
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	fmt.Println("DelayHub server started on port ", addr)
	rpc.Accept(listener)
	fmt.Println("DelayHub server stopped")
}
