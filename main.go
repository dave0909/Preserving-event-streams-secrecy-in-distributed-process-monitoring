package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"main/eventDispatcher"
	"main/processStateManager"
	"main/utils/attestation"
	"main/utils/test"
	"main/utils/xes"
	"os"
	"strconv"
)

func main() {
	////TODO. this is the logic of the process state agent
	//psm := processStateManager.InitProcessStateManager()
	//for {
	//	conn, err := net.Dial("tcp", "localhost:8085")
	//	if err != nil {
	//		fmt.Println("Error connecting:", err)
	//		//os.Exit(1)
	//	} else {
	//		defer conn.Close()
	//		fmt.Println("Connected to localhost:8085")
	//		readEventStream(psm, conn)
	//		break
	//	}
	//}
	////----TODO UPPER WORK FOR EVENT STREAMS
	//
	//////Read event from keyboard input
	////psm := processStateManager.InitProcessStateManager()
	////for {
	////	fmt.Print("Enter the text of the event in XES format:")
	////	reader := bufio.NewReader(os.Stdin)
	////	event, _ := reader.ReadString('\n')
	////	parsed_event, err := xes.ParseXes(event)
	////	if err != nil {
	////		log.Fatalf("Failed to parse XES event: %v", err)
	////	}
	////	psm.HandleEvent(parsed_event.ActivityID, parsed_event.CaseID, parsed_event.Timestamp, parsed_event.Attributes)
	////	//fmt.Println("Final Process State:", psm.ProcessState)
	////}
	addr := os.Args[1]
	manifestFileName := os.Args[2]
	simulationMode := os.Args[3]
	testMode := os.Args[4]
	if testMode == "true" {
		test.TEST_MODE = true
		_, cancel := context.WithCancel(context.Background())
		go test.PrintRamUsage(cancel)
	}
	//Parse the boolean value of the simulation mode
	simulationModeBool, err := strconv.ParseBool(simulationMode)
	if err != nil {
		panic(err)
	}
	fmt.Println("Simulation mode:", simulationModeBool)
	attribute_extractors := make(map[string]interface{})
	if manifestFileName != "" {
		attribute_extractors = parseExtractionManifest(manifestFileName)
	} else {
		attribute_extractors = nil
	}
	eventChannel := make(chan xes.Event)
	psm := processStateManager.InitProcessStateManager(eventChannel)
	eventDispatcher := &eventDispatcher.EventDispatcher{EventChannel: eventChannel, Address: addr, Subscriptions: make(map[string][]attestation.Subscription), AttributeExtractors: attribute_extractors}
	go eventDispatcher.StartRPCServer(addr)
	eventDispatcher.SubscribeTo("localhost:6869")
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

//<org.deckfour.xes.model.impl.XTraceImpl> <log openxes.version="1.0RC7" xes.features="nested-attributes" xes.version="1.0" xmlns="http://www.xes-standard.org/"> <trace> <string key="concept:name" value="2"/> <event> <string key="concept:name" value="Activity G"/> <date key="time:timestamp" value="2024-10-03T16:27:33.682+02:00"/> <string key="organization" value="organization_A"/> </event> </trace> </log> </org.deckfour.xes.model.impl.XTraceImpl>
