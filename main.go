package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"main/eventDispatcher"
	"main/processStateManager"
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

func main() {

	addr := os.Args[1]
	manifestFileName := os.Args[2]
	simulationMode := os.Args[3]
	testMode := os.Args[4]
	nEvents := os.Args[5]
	withExternalQuery := os.Args[6]
	//if testMode == "true" {
	//	test.TEST_MODE = true
	//	_, cancel := context.WithCancel(context.Background())
	//	go test.PrintRamUsage(cancel)
	//}
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
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	eventChannel := make(chan xes.Event)
	psm := processStateManager.InitProcessStateManager(eventChannel, attribute_extractors, n, queueClient)
	eventDispatcher := &eventDispatcher.EventDispatcher{EventChannel: eventChannel, Address: addr, Subscriptions: make(map[string][]attestation.Subscription), AttributeExtractors: attribute_extractors, IsInSimulation: simulationModeBool, ExternalQueryClient: queueClient}
	go eventDispatcher.StartRPCServer(addr)
	time.Sleep(2 * time.Second)
	eventDispatcher.SubscribeTo("localhost:6065")
	// Start recording memory usage
	if testModeBool {
		go recordMemoryUsage(10*time.Millisecond, "data/output/memory_usage.csv", n, &psm)
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

//func readEventStream(psm processStateManager.ProcessStateManager, conn net.Conn) {
//	reader := bufio.NewReader(conn)
//	for {
//		event, err := reader.ReadString('\n')
//		if err != nil {
//			fmt.Println("Error reading from server:", err)
//			return
//		}
//		parsed_event, err := xes.ParseXes(event)
//		if err != nil {
//			log.Fatalf("Failed to parse XES data: %v", err)
//		}
//		psm.HandleEvent(parsed_event.ActivityID, parsed_event.CaseID, parsed_event.Timestamp, parsed_event.Attributes)
//	}
//}

/**
<org.deckfour.xes.model.impl.XTraceImpl><log openxes.version="1.0RC7" xes.features="nested-attributes" xes.version="1.0" xmlns="http://www.xes-standard.org/"><trace><string key="concept:name" value="case_3"/><event><string key="concept:name" value="Select truck"/> <date key="time:timestamp" value="2024-10-03T16:27:33.682+02:00"/> <int key="tempo" value="14109"/> <int key="license_first_issue" value="14107"/> </event> </trace></log></org.deckfour.xes.model.impl.XTraceImpl>
**/

/**
TEST truck policy with violation
<org.deckfour.xes.model.impl.XTraceImpl><log openxes.version="1.0RC7" xes.features="nested-attributes" xes.version="1.0" xmlns="http://www.xes-standard.org/"> <trace> <string key="concept:name" value="3"/> <event> <string key="concept:name" value="Select truck"/> <date key="time:timestamp" value="2024-10-03T16:27:33.682+02:00"/> <date key="license_first_issue" value="2023-10-03T16:27:33.682+02:00"/></event></trace></log></org.deckfour.xes.model.impl.XTraceImpl>
**/
/**
TEST truck policy with no violation
<org.deckfour.xes.model.impl.XTraceImpl><log openxes.version="1.0RC7" xes.features="nested-attributes" xes.version="1.0" xmlns="http://www.xes-standard.org/"> <trace> <string key="concept:name" value="3"/> <event> <string key="concept:name" value="Select truck"/> <date key="time:timestamp" value="2024-10-03T16:27:33.682+02:00"/> <date key="license_first_issue" value="2017-10-03T16:27:33.682+02:00"/></event></trace></log></org.deckfour.xes.model.impl.XTraceImpl>
**/

/**
<org.deckfour.xes.model.impl.XTraceImpl><log openxes.version="1.0RC7" xes.features="nested-attributes" xes.version="1.0" xmlns="http://www.xes-standard.org/"> <trace> <string key="concept:name" value="1"/> <event> <string key="concept:name" value="Reserve shipment"/> <date key="time:timestamp" value="2024-10-03T16:27:33.682+02:00"/> <float key="cost" value="500.00"/> <string key="trace_concept_name" value="1"/> </event></trace></log></org.deckfour.xes.model.impl.XTraceImpl>
**/

/**
<org.deckfour.xes.model.impl.XTraceImpl> <log openxes.version="1.0RC7" xes.features="nested-attributes" xes.version="1.0" xmlns="http://www.xes-standard.org/"> <trace> <string key="concept:name" value="2"/> <event> <string key="concept:name" value="Patient hospitalized (PH)"/> <date key="time:timestamp" value="2024-10-03T16:27:33.682+02:00"/> <string key="organization" value="organization_A"/> </event> </trace> </log> </org.deckfour.xes.model.impl.XTraceImpl>
<org.deckfour.xes.model.impl.XTraceImpl> <log openxes.version="1.0RC7" xes.features="nested-attributes" xes.version="1.0" xmlns="http://www.xes-standard.org/"> <trace> <string key="concept:name" value="2"/> <event> <string key="concept:name" value="Carry out preliminary analyses(COPA)"/> <date key="time:timestamp" value="2024-10-03T16:27:33.682+02:00"/> <string key="organization" value="organization_A"/> </event> </trace> </log> </org.deckfour.xes.model.impl.XTraceImpl>
<org.deckfour.xes.model.impl.XTraceImpl> <log openxes.version="1.0RC7" xes.features="nested-attributes" xes.version="1.0" xmlns="http://www.xes-standard.org/"> <trace> <string key="concept:name" value="2"/> <event> <string key="concept:name" value="Order drugs(OD)"/> <date key="time:timestamp" value="2024-10-03T16:27:33.682+02:00"/> <string key="organization" value="organization_A"/> </event> </trace> </log> </org.deckfour.xes.model.impl.XTraceImpl>
<org.deckfour.xes.model.impl.XTraceImpl> <log openxes.version="1.0RC7" xes.features="nested-attributes" xes.version="1.0" xmlns="http://www.xes-standard.org/"> <trace> <string key="concept:name" value="2"/> <event> <string key="concept:name" value="Drugs order received (DOR)"/> <date key="time:timestamp" value="2024-10-03T16:27:33.682+02:00"/> <string key="organization" value="organization_A"/> </event> </trace> </log> </org.deckfour.xes.model.impl.XTraceImpl>
<org.deckfour.xes.model.impl.XTraceImpl> <log openxes.version="1.0RC7" xes.features="nested-attributes" xes.version="1.0" xmlns="http://www.xes-standard.org/"> <trace> <string key="concept:name" value="2"/> <event> <string key="concept:name" value="Patient  hospitalized (PH)"/> <date key="time:timestamp" value="2024-10-03T16:27:33.682+02:00"/> <string key="organization" value="organization_A"/> </event> </trace> </log> </org.deckfour.xes.model.impl.XTraceImpl>
<org.deckfour.xes.model.impl.XTraceImpl> <log openxes.version="1.0RC7" xes.features="nested-attributes" xes.version="1.0" xmlns="http://www.xes-standard.org/"> <trace> <string key="concept:name" value="2"/> <event> <string key="concept:name" value="Drugs order received (DOR)"/> <date key="time:timestamp" value="2024-10-03T16:27:33.682+02:00"/> <string key="organization" value="organization_A"/> </event> </trace> </log> </org.deckfour.xes.model.impl.XTraceImpl>
**/

// <org.deckfour.xes.model.impl.XTraceImpl> <log openxes.version="1.0RC7" xes.features="nested-attributes" xes.version="1.0" xmlns="http://www.xes-standard.org/"> <trace> <string key="concept:name" value="2"/> <event> <string key="concept:name" value="Activity G"/> <date key="time:timestamp" value="2024-10-03T16:27:33.682+02:00"/> <string key="organization" value="organization_A"/> </event> </trace> </log> </org.deckfour.xes.model.impl.XTraceImpl>
// Function to record memory usage
func recordMemoryUsage(interval time.Duration, fileName string, nEvents int, psm *processStateManager.ProcessStateManager) {
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
	defer ticker.Stop()

	for {
		fmt.Println(psm.TotalCounter, nEvents)
		if psm.TotalCounter == nEvents {
			break
		}
		select {
		case <-ticker.C:
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
			fmt.Printf("%s - Memory Usage: %d MiB", currentTime, alloc)
		}
	}
	fmt.Println("End of the test after ", nEvents, "events")
}
