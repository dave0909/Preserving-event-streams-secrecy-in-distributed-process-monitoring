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
	"sort"
	"sync"
	"time"
)

type EventData struct {
	ArrivalTimestamp    int
	CompletionTimestamp int
}

type DelayHub struct {
	mu          sync.Mutex
	eventStore  map[int]*EventData
	memoryStore []delayargs.MemoryData
	csvFile     *os.File
	csvWriter   *csv.Writer
}

func NewDelayHub() (*DelayHub, error) {
	// Check if the file exists, if not create it
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
		eventStore:  make(map[int]*EventData),
		memoryStore: make([]delayargs.MemoryData, 0),
		csvFile:     file,
		csvWriter:   writer,
	}, nil
}

type MemoryUsageArgs struct {
	Timestamp   int
	MemoryUsage int
}

func (d *DelayHub) RegisterMemoryUsage(args *MemoryUsageArgs, reply *bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.memoryStore = append(d.memoryStore, delayargs.MemoryData{
		Timestamp:   args.Timestamp,
		MemoryUsage: args.MemoryUsage,
	})

	fmt.Printf("Registered memory usage: %d at timestamp: %d\n", args.MemoryUsage, args.Timestamp)
	*reply = true
	return nil
}

func (d *DelayHub) WriteMemoryUsage(_ *struct{}, reply *bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Create memory usage CSV file
	outputDir := "../data/output"
	filePath := filepath.Join(outputDir, "memory_usage.csv")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating memory usage file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"Timestamp", "Memory Usage"}); err != nil {
		return fmt.Errorf("error writing header: %v", err)
	}

	// Sort entries by timestamp
	sort.Slice(d.memoryStore, func(i, j int) bool {
		return d.memoryStore[i].Timestamp < d.memoryStore[j].Timestamp
	})

	// Write entries
	for _, entry := range d.memoryStore {
		record := []string{
			fmt.Sprintf("%d", entry.Timestamp),
			fmt.Sprintf("%d", entry.MemoryUsage),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error writing record: %v", err)
		}
	}

	fmt.Println("Successfully wrote memory usage data to CSV")
	*reply = true
	return nil
}

// ... rest of the existing methods (GetCurrentTimestamp, writeRecordIfComplete, WriteArrival, WriteCompletion) remain unchanged ...

// GetCurrentTimestamp returns the current Unix timestamp in nanoseconds
func (d *DelayHub) GetCurrentTimestamp(_ *struct{}, reply *delayargs.TimestampResponse) error {
	*reply = delayargs.TimestampResponse{
		Timestamp: time.Now().UnixNano(),
	}
	return nil
}

func (d *DelayHub) writeRecordIfComplete(eventCode int) error {
	event := d.eventStore[eventCode]
	if event.ArrivalTimestamp != 0 && event.CompletionTimestamp != 0 {
		record := []string{
			fmt.Sprintf("%d", eventCode),
			fmt.Sprintf("%d", event.ArrivalTimestamp),
			fmt.Sprintf("%d", event.CompletionTimestamp),
		}
		if err := d.csvWriter.Write(record); err != nil {
			return err
		}
		d.csvWriter.Flush()
		if err := d.csvWriter.Error(); err != nil {
			return err
		}
		delete(d.eventStore, eventCode)
		//fmt.Println("Recorded complete event for code: ", eventCode, " and wrote to CSV.")
	}
	return nil
}

func (d *DelayHub) WriteArrival(args *delayargs.ArrivalArgs, reply *bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	event, exists := d.eventStore[args.EventCode]
	if !exists {
		event = &EventData{}
		d.eventStore[args.EventCode] = event
	}
	if event.ArrivalTimestamp != 0 {
		return errors.New("arrival time already recorded for this event code")
	}
	event.ArrivalTimestamp = int(time.Now().UnixNano())
	//fmt.Println("Recorded arrival for event code: ", args.EventCode, " with timestamp: ", int(time.Now().UnixNano()))
	err := d.writeRecordIfComplete(args.EventCode)
	if err != nil {
		return err
	}
	*reply = true
	return nil
}

func (d *DelayHub) WriteCompletion(args *delayargs.CompletionArgs, reply *bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	event, exists := d.eventStore[args.EventCode]
	if !exists {
		event = &EventData{}
		d.eventStore[args.EventCode] = event
	}
	if event.CompletionTimestamp != 0 {
		return errors.New("completion time already recorded for this event code")
	}
	event.CompletionTimestamp = int(time.Now().UnixNano())
	//fmt.Println("Recorded completion for event code: ", args.EventCode, " with timestamp: ", int(time.Now().UnixNano()))
	err := d.writeRecordIfComplete(args.EventCode)
	if err != nil {
		return err
	}
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
