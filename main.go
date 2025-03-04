package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"main/processVault/eventDispatcher"
	"main/processVault/processStateManager"
	"main/utils/attestation"
	"main/utils/xes"
	"net/http"
	_ "net/http/pprof"
	"net/rpc"
	"os"
	"runtime"
	"strconv"
	"time"
)

// main is the entry point of the application. It initializes various components
// and starts the event processing and memory usage recording.
func main() {
	// Start a goroutine to enable pprof for profiling
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Parse command-line arguments
	addr := os.Args[1]
	manifestFileName := os.Args[2]
	simulationMode := os.Args[3]
	testMode := os.Args[4]
	nEvents := os.Args[5]
	withExternalQuery := os.Args[6]
	slidingWindowSize := os.Args[7]

	// Set default value for nEvents if not provided
	if nEvents == "" {
		nEvents = "0"
	}

	// Convert nEvents to an integer
	n, err := strconv.Atoi(nEvents)
	if err != nil {
		panic(err)
	}

	// Convert simulationMode to a boolean
	simulationModeBool, err := strconv.ParseBool(simulationMode)
	if err != nil {
		panic(err)
	}

	// Convert testMode to a boolean
	testModeBool, err := strconv.ParseBool(testMode)
	if err != nil {
		panic(err)
	}

	// Parse the extraction manifest file if provided
	attribute_extractors := make(map[string]interface{})
	if manifestFileName != "" {
		attribute_extractors = parseExtractionManifest(manifestFileName)
	} else {
		attribute_extractors = nil
	}

	// Convert withExternalQuery to a boolean
	withExternalQueryBool, err := strconv.ParseBool(withExternalQuery)
	if err != nil {
		panic(err)
	}

	// Initialize the RPC client if external query is enabled
	var queueClient *rpc.Client
	if withExternalQueryBool {
		queueClient, err = rpc.Dial("tcp", "localhost:8387")
		if err != nil {
			log.Fatal("Dialing:", err)
		}
	}

	// Convert slidingWindowSize to an integer
	slidingWindowInt, err := strconv.Atoi(slidingWindowSize)
	if err != nil {
		panic(err)
	}

	// Create a channel for events
	eventChannel := make(chan xes.Event)

	// Initialize the process state manager
	psm := processStateManager.InitProcessStateManager(eventChannel, attribute_extractors, n, queueClient, slidingWindowInt)

	// Initialize the event dispatcher
	eventDispatcher := &eventDispatcher.EventDispatcher{
		EventChannel:        eventChannel,
		Address:             addr,
		Subscriptions:       make(map[string][]attestation.Subscription),
		AttributeExtractors: attribute_extractors,
		IsInSimulation:      simulationModeBool,
		ExternalQueryClient: queueClient,
	}

	// Start the RPC server for the event dispatcher
	go eventDispatcher.StartRPCServer(addr)

	// Wait for the RPC server to start
	time.Sleep(2 * time.Second)

	// Subscribe to events from a specific address
	eventDispatcher.SubscribeTo("localhost:6065")

	// Start recording memory usage if testConfigurations mode is enabled
	if testModeBool {
		go recordMemoryUsage(10*time.Millisecond, 0*time.Millisecond, "data/output/memory_usage.csv", n, &psm)
	}

	// Wait for all events to be processed
	psm.WaitForEvents()
}

func parseExtractionManifest(manifestFileName string) map[string]interface{} {
	manifestFile, err := os.Open(manifestFileName)
	if err != nil {
		fmt.Println("Error opening manifest file:", err)
		return nil
	}
	defer manifestFile.Close()
	byteValue, err := ioutil.ReadAll(manifestFile)
	if err != nil {
		fmt.Println("Error reading manifest file:", err)
		return nil
	}
	var manifest map[string]interface{}
	err = json.Unmarshal(byteValue, &manifest)
	if err != nil {
		fmt.Println("Error unmarshalling manifest file:", err)
		return nil
	}

	//fmt.Println("Parsed manifest:", manifest)
	attribute_extractors := manifest["attribute_extraction"].(map[string]interface{})
	return attribute_extractors
}

// Function to record memory usage
/**
func recordMemoryUsage(interval time.Duration, gcInterval time.Duration, fileName string, nEvents int, psm *processStateManager.ProcessStateManager) {
	// Open the file in append mode (create if it doesn't exist)
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()
	// Write the header only if the file is empty
	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatalf("Failed to get file info: %v", err)
	}
	if fileInfo.Size() == 0 {
		_, _ = file.WriteString("Timestamp,Memory Usage\n")
	}
	ticker := time.NewTicker(interval)
	gcTicker := time.NewTicker(gcInterval)
	defer ticker.Stop()
	for {
		if psm.FirstInit {
			if psm.TotalCounter == nEvents {
				break
			}
			select {
			case <-ticker.C:
				//runtime.GC()
				// Capture memory statistics
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				// Format memory stats
				currentTime := time.Now().Unix()
				alloc := m.HeapAlloc
				// Write to file
				data := fmt.Sprintf("%d,%d\n", currentTime, alloc)
				_, _ = file.WriteString(data)
				// Also print to the console
				//fmt.Printf("%s - Memory Usage: %d MiB", currentTime, alloc)
			case <-gcTicker.C:
				if gcInterval != 0 {
					go runtime.GC()
				}
			}
		}
	}
	fmt.Println("End of the testConfigurations after ", nEvents, "events")
}
**/

func recordMemoryUsage(interval time.Duration, gcInterval time.Duration, fileName string, nEvents int, psm *processStateManager.ProcessStateManager) {
	// File setup
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatalf("Failed to get file info: %v", err)
	}
	if fileInfo.Size() == 0 {
		_, _ = file.WriteString("Timestamp,Memory Usage\n")
	}
	ticker := time.NewTicker(interval)

	defer ticker.Stop()
	//Garbage collection
	if gcInterval != 0 {
		gcTicker := time.NewTicker(gcInterval)
		defer gcTicker.Stop()
		// Garbage collection goroutine
		go func() {
			for range gcTicker.C {
				if psm.FirstInit && psm.TotalCounter < nEvents && gcInterval != 0 {
					go runtime.GC()
				}
			}
		}()
	}
	// Memory usage recording goroutine
	go func() {

		for range ticker.C {
			if psm.FirstInit && psm.TotalCounter < nEvents {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				currentTime := time.Now().UnixMilli()
				alloc := m.HeapAlloc
				data := fmt.Sprintf("%d,%d\n", currentTime, alloc)
				go file.WriteString(data)
			}
		}
	}()

	// Wait for all events to be processed
	for psm.TotalCounter < nEvents {
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("End of the testConfigurations after", nEvents, "events")
}
