package processStateManager

import (
	"encoding/csv"
	"fmt"
	"github.com/edgelesssys/ego/ecrypto"
	"log"
	"main/complianceCheckingLogic"
	"main/utils/xes"
	"main/workflowLogic"
	"math"
	"net/rpc"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"time"
)

// TODO: we assume that no activity is named "source" or "sink"
// Struct Process State
type ProcessState struct {
	//Maps (set) of all the processed Cases
	Cases map[string]bool
	//Current state of the process
	WfState WorkflowState
	//List of all the events
	EventLog []map[string]interface{}
	//Compliance checking violations
	//ComplianceCheckingViolations map[string]map[string]bool
	//Compliance checking violations
	ComplianceCheckingViolations map[string][]ComplianceCheckingViolation
	//TODO: Remove this field
	Counter int
}

// Struct Event
type Event struct {
	//Event id
	Activity string
	//Case id
	CaseId string
	//Timestamp
	Timestamp string
	//Data
	Data map[string]interface{}
}

// WorkflowState struct
type WorkflowState struct {
	//Map case names to an array of next possible activities
	expectNextActivities map[string][]string
	//Map case names to an array of token ids
	CaseTokenId map[string]int
	//Map case names to an array of violations
	//WorkflowViolations map[string]WorkflowViolation
	//1 for pending, 2 for violated, 3 for satisfied
	WorkflowStatus map[string]int
}

// Get the next possible activities for a case
func (wf *WorkflowState) GetNextActivities(caseId string) []string {
	return wf.expectNextActivities[caseId]
}

// Get the workflow violation
//func (wf *WorkflowState) GetViolations() map[string]WorkflowViolation {
//return wf.WorkflowViolations
//}

func (wf *WorkflowState) GetStatus(caseId string) int { return wf.WorkflowStatus[caseId] }

// ProcessStateManager struct
type ProcessStateManager struct {
	WorkflowLogic           workflowLogic.WorkflowLogic
	ComplianceCheckingLogic complianceCheckingLogic.ComplianceCheckingLogic
	ProcessState            ProcessState
	EventChannel            chan xes.Event
	ExtractionManifest      map[string]interface{}
	minDuration             float64
	maxDuration             float64
	totalDuration           float64
	runStarted              time.Time
	mean                    float64
	m2                      float64
	stopEventNumebr         int
	TotalCounter            int
	ExternalQueueClient     *rpc.Client
	FirstInit               bool
	slidingWindowSize       int
}

// Init a new ProcessStateManager
func InitProcessStateManager(eventChannel chan xes.Event, extractionManifest map[string]interface{}, stopEventNumber int, externalQueueClient *rpc.Client, slidingWindowSize int) ProcessStateManager {
	//ccLogic, cNames := complianceCheckingLogic.InitComplianceCheckingLogic()
	ccLogic, _ := complianceCheckingLogic.InitComplianceCheckingLogic()
	psm := ProcessStateManager{
		WorkflowLogic:           workflowLogic.InitWorkflowLogic(),
		ComplianceCheckingLogic: ccLogic,
		EventChannel:            eventChannel,
		ExtractionManifest:      extractionManifest,
		minDuration:             math.MaxFloat64,
		runStarted:              time.Now(),
		stopEventNumebr:         stopEventNumber,
		TotalCounter:            0,
		ExternalQueueClient:     externalQueueClient,
		FirstInit:               false,
		slidingWindowSize:       slidingWindowSize,
	}
	//ccViolation := map[string]map[string]bool{}
	//ccViolation := map[string]map[string]ComplianceCheckingViolation{}

	ccViolation := map[string][]ComplianceCheckingViolation{}
	//for _, name := range cNames {
	//	ccViolation[name] = map[string]ComplianceCheckingViolation{}
	//}
	ps := ProcessState{
		Cases: map[string]bool{},
		WfState: WorkflowState{
			expectNextActivities: map[string][]string{},
			//WorkflowViolations:   map[string]WorkflowViolation{},
			WorkflowStatus: map[string]int{},
			CaseTokenId:    map[string]int{},
		},
		ComplianceCheckingViolations: ccViolation,
	}
	psm.ProcessState = ps
	fmt.Println("Init PSM with sliding window set to", slidingWindowSize)
	return psm

}

// Init a new case
func (psm *ProcessStateManager) initNewCase(caseId string) {
	psm.ProcessState.Cases[caseId] = true
	//Set the status of the case to pending
	psm.ProcessState.WfState.WorkflowStatus[caseId] = 1
	indexOfSource, _ := psm.WorkflowLogic.GetSourceAndSinkIndices()
	psm.WorkflowLogic.Petrinet.State[indexOfSource] += 1
	///psm.WorkflowLogic.Petrinet.TokenIds[indexOfSource] = append(psm.WorkflowLogic.Petrinet.TokenIds[indexOfSource], 1)
	psm.WorkflowLogic.Petrinet.Init()
	//Map the case name to the token id
	psm.ProcessState.WfState.CaseTokenId[caseId] = psm.WorkflowLogic.Petrinet.TokenId

	//Find the start event transition checking the input matrix
	//TODO: here should be executed only if the input petrinet is derived from a BPMN diagram
	//for i, tr := range psm.WorkflowLogic.Petrinet.InputMatrix {
	//	if tr[indexOfSource] == 1 {
	//		//Start event found!
	//		//Get the name of the start event from its index
	//		start_event := psm.WorkflowLogic.Transitions[i]
	//Fire the start event
	//		psm.WorkflowLogic.FireTokenIdWithTransitionName(start_event, psm.ProcessState.WfState.CaseTokenId[caseId])
	//	}
	//}
}

// Handle event by EventDispatcher
func (psm *ProcessStateManager) HandleEvent(eventId string, caseId string, timestamp string, data map[string]interface{}) {
	if psm.stopEventNumebr != 0 {
		if !psm.FirstInit {
			psm.FirstInit = true
		}
	}
	//Check if the event exists in the workflow logic
	firtsTs := time.Now()
	psm.TotalCounter += 1
	//TODO: this should be moved into the event dispatcher
	psm.ProcessState.Counter += 1
	//Add the event to the list of events
	//psm.ProcessState.Events = append(psm.ProcessState.Events, event)
	eventLogEntry := map[string]interface{}{
		"trace_concept_name": caseId,
		"concept_name":       eventId,
		"timestamp":          timestamp,
	}
	addEventFlag := false
	//Check if the eventId variable is a key in the "attribute_extraction" field of the extraction manifest
	if _, ok := psm.ExtractionManifest[eventId]; ok {
		//If the eventId is a key in the "attribute_extraction" field, extract the attributes
		attributeExtractors := psm.ExtractionManifest[eventId].([]interface{})
		for _, attrName := range attributeExtractors {
			//If attrName is a key in data
			if _, ok := data[attrName.(string)]; ok {
				eventLogEntry[attrName.(string)] = data[attrName.(string)]
			}
		}
		//psm.ProcessState.EventLog = append(psm.ProcessState.EventLog, eventLogEntry)
		addEventFlag = true
	}
	//Check if the event is in the workflow logic
	if psm.WorkflowLogic.GetTransitionIndicesByName(eventId) != nil {
		//Check if the case is already in the process state
		if !psm.ProcessState.Cases[caseId] {
			//If the case is not in the process state, initialize a new case
			psm.initNewCase(caseId)
		}
		//Check if the case is already in errouneous workflow state
		//_, cfStatus := psm.ProcessState.WfState.WorkflowViolations[caseId]
		cfStatus := psm.ProcessState.WfState.GetStatus(caseId)
		if cfStatus == 2 {
			//Append the current event to the erroneous sequence
			//TODO: UNCOMMENT THIS TO KEEP TRACK OF THE ERRONEOUS SEQUENCE
			//wfViolation := psm.ProcessState.WfState.WorkflowViolations[caseId]
			//wfViolation.ErroneousSequence = append(wfViolation.ErroneousSequence, eventId)
			//psm.ProcessState.WfState.WorkflowViolations[caseId] = wfViolation
		} else {
			//Fire the transition associated with the event
			error := psm.WorkflowLogic.FireTokenIdWithTransitionName(eventId, psm.ProcessState.WfState.CaseTokenId[caseId])
			if error != nil {
				//If the transition failed, generate a new workflow violation
				psm.initWorkflowViolation(eventId, caseId, timestamp, error)
				fmt.Println("New workflow violation for case: ", caseId, " event: ", eventId)
			} else {
				//If the transition was successful, update the workflow state
				psm.updateWorkflowState(caseId)
				//fmt.Println("Succesful state update with case: ", caseId, " event: ", eventId, " next activities: ", psm.ProcessState.WfState.GetNextActivities(caseId))
			}
		}
	}
	//fmt.Println("Time for workflow monitoring: ", time.Since(firtsTs).Seconds())
	elaboratedLog := map[string][]map[string]interface{}{}
	//Filter the event log to compute only the events with the same case id
	elaboratedLog["events"] = []map[string]interface{}{}
	for _, event := range psm.ProcessState.EventLog {
		if event["trace_concept_name"] == caseId {
			elaboratedLog["events"] = append(elaboratedLog["events"], event)
		}
	}
	//elaboratedLog["events"] = append(psm.ProcessState.EventLog, eventLogEntry)
	elaboratedLog["events"] = append(elaboratedLog["events"], eventLogEntry)
	_ = psm.ComplianceCheckingLogic.EvaluateEventLog(elaboratedLog)
	if addEventFlag {
		psm.ProcessState.EventLog = append(psm.ProcessState.EventLog, eventLogEntry)
	}
	//Clear old events
	//if len(psm.ProcessState.EventLog) == 150 {
	if len(psm.ProcessState.EventLog) == psm.slidingWindowSize {

		//clear the events log by removing the first 100 events
		//psm.ProcessState.EventLog = psm.ProcessState.EventLog[100:]
		psm.ProcessState.EventLog = slices.Delete(psm.ProcessState.EventLog, 0, psm.slidingWindowSize-50)
		runtime.GC()
	}
	//fmt.Println("Event number: ", psm.ProcessState.Counter)
	duration := time.Since(firtsTs).Seconds()
	durationFromStart := time.Since(psm.runStarted)
	psm.totalDuration += duration
	if duration < psm.minDuration && duration != 0 {
		psm.minDuration = duration
	}
	if duration > psm.maxDuration {
		psm.maxDuration = duration
	}
	// Incremental calculation of mean and variance
	delta := duration - psm.mean
	psm.mean += delta / float64(psm.ProcessState.Counter)
	delta2 := duration - psm.mean
	psm.m2 += delta * delta2
	// Calculate standard deviation
	variance := psm.m2 / float64(psm.ProcessState.Counter)
	stdDev := math.Sqrt(variance)
	//fmt.Printf("Time from start of the run:%f, Current mean (s): %f,Min duration (s): %f, Max duration (s): %f, Std Dev (s): %f\n", durationFromStart.Seconds(), psm.mean, psm.minDuration, psm.maxDuration, stdDev)
	fmt.Println("Processed events: ", psm.TotalCounter, " out of: ", psm.stopEventNumebr)
	if psm.TotalCounter == psm.stopEventNumebr && psm.stopEventNumebr != 0 {
		recordDataDuration(durationFromStart, psm, stdDev)
	}
}

func recordDataDuration(durationFromStart time.Duration, psm *ProcessStateManager, stdDev float64) {
	// Ensure the directory exists
	outputDir := "./data/output"
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		log.Fatalf("Error creating directory: %v", err)
	}

	// Open the CSV file for writing
	filePath := filepath.Join(outputDir, "latency.csv")
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Error creating CSV file: %v", err)
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the header
	header := []string{"DurationFromStart", "Mean", "MinDuration", "MaxDuration", "StdDev"}
	if err := writer.Write(header); err != nil {
		log.Fatalf("Error writing header to CSV file: %v", err)
	}

	// Write the duration data
	record := []string{
		fmt.Sprintf("%f", durationFromStart.Seconds()),
		fmt.Sprintf("%f", psm.mean),
		fmt.Sprintf("%f", psm.minDuration),
		fmt.Sprintf("%f", psm.maxDuration),
		fmt.Sprintf("%f", stdDev),
	}
	if err := writer.Write(record); err != nil {
		log.Fatalf("Error writing record to CSV file: %v", err)
	}
	panic("End of the test")
}

func (psm *ProcessStateManager) initWorkflowViolation(eventId string, caseId string, timestamp string, error error) {
	//Reset expect next activities for the case
	psm.ProcessState.WfState.expectNextActivities[caseId] = []string{}
	//Generate a new workflow violation with the current event
	//wfViolation := WorkflowViolation{
	//	//GeneratedByEvent:  eventId,
	//	GeneratedByCase: caseId,
	//	Timestamp:       timestamp,
	//	//ErroneousSequence: []string{eventId},
	//}
	//Add the new workflow violation to the workflow violations map
	//psm.ProcessState.WfState.WorkflowViolations[caseId] = wfViolation
	//Set the status of the case to violated
	psm.ProcessState.WfState.WorkflowStatus[caseId] = 2
}

func (psm *ProcessStateManager) updateWorkflowState(caseId string) {
	//Reset the expect next activities for the case
	psm.ProcessState.WfState.expectNextActivities[caseId] = []string{}
	//Init map of processed transitions
	processedTransitions := map[string]bool{}
	//For each token position, get the next possible transitions
	nextTransitions := psm.WorkflowLogic.GetEnabledTransitionsForTokenId(psm.ProcessState.WfState.CaseTokenId[caseId])
	//Iterate over the possible next transitions
	for _, transition := range nextTransitions {
		//If the transition is not already in the expect next activities
		if !processedTransitions[transition] {
			//Add the possible next transition to the expect next activities
			psm.ProcessState.WfState.expectNextActivities[caseId] = append(psm.ProcessState.WfState.expectNextActivities[caseId], transition)
			//Mark the transition as processed
			processedTransitions[transition] = true
		}
	}
}

// Get the expected next activities for a case
func (psm *ProcessStateManager) GetExpectedNextActivities(caseId string) (error, []string) {
	//Check if the case is in errouneous state
	//_, wfViolated := psm.ProcessState.WfState.WorkflowViolations[caseId]
	wfViolated := psm.ProcessState.WfState.GetStatus(caseId) == 2
	if wfViolated {
		return fmt.Errorf("Case %v is in an erroneous state", caseId), nil
	}
	return nil, psm.ProcessState.WfState.GetNextActivities(caseId)
}

// Wait for new events from the event channel
func (psm *ProcessStateManager) WaitForEvents() {
	if psm.ExternalQueueClient == nil {
		for {
			event := <-psm.EventChannel
			psm.HandleEvent(event.ActivityID, event.CaseID, event.Timestamp, event.Attributes)
		}
	} else {
		for {
			var event []byte
			err := psm.ExternalQueueClient.Call("Queue.DequeueEvent", event, &event)
			if err != nil {
				continue
			}
			unsealedByteEvent, err := ecrypto.Unseal(event, []byte(""))
			if err != nil {
				fmt.Println("Error unsealing event: ", err)
				panic(err)
			}
			//Convert the byte array to the string
			stringEvent := string(unsealedByteEvent)
			//Parse the event
			parsedEvent, err := xes.ParseXes(stringEvent)
			if err != nil {
				fmt.Println("Error parsing event: ", err)
				panic(err)
			}
			psm.HandleEvent(parsedEvent.ActivityID, parsedEvent.CaseID, parsedEvent.Timestamp, parsedEvent.Attributes)
		}
	}
}

// function that print the process state nicely
func (psm *ProcessStateManager) PrintProcessState() {
	fmt.Println("Printing process state")
	fmt.Println("Next activities", psm.ProcessState.WfState.expectNextActivities)
	fmt.Println("Control Flow Violations", psm.ProcessState.WfState.WorkflowStatus)
	fmt.Println("Compliance Checking violations", psm.ProcessState.ComplianceCheckingViolations)
}
