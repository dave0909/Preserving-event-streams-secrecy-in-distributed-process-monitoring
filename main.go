package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"main/eventDispatcher"
	"main/processStateManager"
	"main/utils/attestation"
	"main/utils/delayargs"
	"main/utils/xes"
	"net/http"
	_ "net/http/pprof"
	"net/rpc"
	"os"
	"runtime"
	"strconv"
	"time"
)

func main() {
	//debug.SetGCPercent(-1)
	//debug.SetGCPercent(5)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	addr := os.Args[1]
	manifestFileName := os.Args[2]
	simulationMode := os.Args[3]
	testMode := os.Args[4]
	nEvents := os.Args[5]
	withExternalQuery := os.Args[6]
	slidingWindowSize := os.Args[7]
	//Parse the boolean value of the simulation mode
	if nEvents == "" {
		nEvents = "0"
	}
	//fmt.Println("Number of processed events: ", nEvents)
	n, err := strconv.Atoi(nEvents)
	if err != nil {
		// ... handle error
		panic(err)
	}
	simulationModeBool, err := strconv.ParseBool(simulationMode)
	if err != nil {
		panic(err)
	}
	testModeBool, err := strconv.ParseBool(testMode)
	if err != nil {
		panic(err)
	}
	attribute_extractors := make(map[string]interface{})
	if manifestFileName != "" {
		attribute_extractors = parseExtractionManifest(manifestFileName)
	} else {
		attribute_extractors = nil
	}
	withExternalQueryBool, err := strconv.ParseBool(withExternalQuery)
	if err != nil {
		panic(err)
	}
	var queueClient *rpc.Client
	if withExternalQueryBool {
		// Connect to the RPC server
		queueClient, err = rpc.Dial("tcp", "localhost:8387")
		if err != nil {
			log.Fatal("Dialing:", err)
		}
	}
	slidingWindowInt, err := strconv.Atoi(slidingWindowSize)
	if err != nil {
		// ... handle error
		panic(err)
	}
	eventChannel := make(chan xes.Event)
	psm := processStateManager.InitProcessStateManager(eventChannel, attribute_extractors, n, queueClient, slidingWindowInt)
	eventDispatcher := &eventDispatcher.EventDispatcher{EventChannel: eventChannel, Address: addr, Subscriptions: make(map[string][]attestation.Subscription), AttributeExtractors: attribute_extractors, IsInSimulation: simulationModeBool, ExternalQueryClient: queueClient}
	go eventDispatcher.StartRPCServer(addr)
	time.Sleep(2 * time.Second)
	eventDispatcher.SubscribeTo("localhost:6065")
	// Start recording memory usage
	if testModeBool {
		go recordMemoryUsage(10*time.Millisecond, 20*time.Millisecond, "data/output/memory_usage.csv", n, &psm)
	}
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
	fmt.Println("End of the test after ", nEvents, "events")
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
	/**
	go func() {
		delaycli, err := rpc.Dial("tcp", "localhost:8388")
		if err != nil {
			log.Fatal("Cannot connect to Delay Hub:", err)
		}
		for range ticker.C {
			if psm.FirstInit && psm.TotalCounter < nEvents {
				go sendMemoryUsage(delaycli)
			}
		}
	}()**/

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

	fmt.Println("End of the test after", nEvents, "events")
}

func sendMemoryUsage(delaycli *rpc.Client) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	currentTime := time.Now().UnixMilli()
	alloc := m.HeapAlloc
	args := delayargs.MemoryData{
		Timestamp:   int(currentTime),
		MemoryUsage: int(alloc),
	}
	var reply bool
	go delaycli.Call("DelayHub.RegisterMemoryUsage", args, &reply)
}
