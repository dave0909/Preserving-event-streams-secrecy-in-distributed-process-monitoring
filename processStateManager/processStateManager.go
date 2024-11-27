package processStateManager

import (
	"fmt"
	"main/complianceCheckingLogic"
	"main/utils/xes"
	"main/workflowLogic"
	"math"
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
	WorkflowViolations map[string]WorkflowViolation
}

// Get the next possible activities for a case
func (wf *WorkflowState) GetNextActivities(caseId string) []string {
	return wf.expectNextActivities[caseId]
}

// Get the workflow violation
func (wf *WorkflowState) GetViolations() map[string]WorkflowViolation {
	return wf.WorkflowViolations
}

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
}

// Init a new ProcessStateManager
func InitProcessStateManager(eventChannel chan xes.Event, extractionManifest map[string]interface{}) ProcessStateManager {
	//ccLogic, cNames := complianceCheckingLogic.InitComplianceCheckingLogic()
	ccLogic, _ := complianceCheckingLogic.InitComplianceCheckingLogic()
	psm := ProcessStateManager{
		WorkflowLogic:           workflowLogic.InitWorkflowLogic(),
		ComplianceCheckingLogic: ccLogic,
		EventChannel:            eventChannel,
		ExtractionManifest:      extractionManifest,
		minDuration:             math.MaxFloat64,
		runStarted:              time.Now(),
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
			WorkflowViolations:   map[string]WorkflowViolation{},
			CaseTokenId:          map[string]int{},
		},
		ComplianceCheckingViolations: ccViolation,
	}
	psm.ProcessState = ps
	return psm

}

// Init a new case
func (psm *ProcessStateManager) initNewCase(caseId string) {
	psm.ProcessState.Cases[caseId] = true
	indexOfSource, _ := psm.WorkflowLogic.GetSourceAndSinkIndices()
	psm.WorkflowLogic.Petrinet.State[indexOfSource] += 1
	///psm.WorkflowLogic.Petrinet.TokenIds[indexOfSource] = append(psm.WorkflowLogic.Petrinet.TokenIds[indexOfSource], 1)
	psm.WorkflowLogic.Petrinet.Init()
	//Map the case name to the token id
	psm.ProcessState.WfState.CaseTokenId[caseId] = psm.WorkflowLogic.Petrinet.TokenId
	//Find the start event transition checking the input matrix
	for i, tr := range psm.WorkflowLogic.Petrinet.InputMatrix {
		if tr[indexOfSource] == 1 {
			//Start event found!
			//Get the name of the start event from its index
			start_event := psm.WorkflowLogic.Transitions[i]
			//Fire the start event
			psm.WorkflowLogic.FireTokenIdWithTransitionName(start_event, psm.ProcessState.WfState.CaseTokenId[caseId])
		}
	}
}

// Handle event by EventDispatcher
func (psm *ProcessStateManager) HandleEvent(eventId string, caseId string, timestamp string, data map[string]interface{}) {
	//Check if the event exists in the workflow logic
	firtsTs := time.Now()
	fmt.Println("event number: ", len(psm.ProcessState.EventLog))
	//TODO: this should be moved into the event dispatcher
	if psm.WorkflowLogic.GetTransitionIndexByName(eventId) == -1 {
		fmt.Println("Unknown event")
		return
	}
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

	//for k, v := range data {
	//	eventLogEntry[k] = v
	//	//
	//}
	//psm.ProcessState.EventLog = append(psm.ProcessState.EventLog, eventLogEntry)
	//If the case is not in the cases map, add it
	if !psm.ProcessState.Cases[caseId] {
		psm.initNewCase(caseId)
		//fmt.Println("New workflow violation for case: ", caseId, " event: ", eventId)
	}
	//Check if the case is already in errouneous workflow state
	_, wfViolated := psm.ProcessState.WfState.WorkflowViolations[caseId]
	if wfViolated {
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
	//Print time passed from firstTS
	fmt.Println("Time for workflow monitoring: ", time.Since(firtsTs).Seconds())
	//elaboratedLog := psm.prepareEventLog()
	elaboratedLog := map[string]interface{}{}
	//elaboratedLog["events"] = psm.ProcessState.EventLog
	elaboratedLog["events"] = append(psm.ProcessState.EventLog, eventLogEntry)
	psm.ComplianceCheckingLogic.EvaluateEventLog(elaboratedLog)
	//violationMap := psm.ComplianceCheckingLogic.EvaluateEventLog(elaboratedLog)
	//for constraint, result := range violationMap {
	//	castedResult := result.(map[string]interface{})
	//	for caseId, _ := range castedResult {
	//		if _, ok := psm.ProcessState.ComplianceCheckingViolations[caseId]; !ok {
	//			psm.ProcessState.ComplianceCheckingViolations[caseId] = []ComplianceCheckingViolation{}
	//		}
	//		//Chek if the case has already a violation for the given constraint
	//		violated := false
	//		for _, violation := range psm.ProcessState.ComplianceCheckingViolations[caseId] {
	//			if violation.ViolatedConstraint == constraint {
	//				violated = true
	//				break
	//			}
	//		}
	//		if !violated {
	//			//If the case has not a violation for the given constraint, add it
	//			psm.ProcessState.ComplianceCheckingViolations[caseId] = append(psm.ProcessState.ComplianceCheckingViolations[caseId], ComplianceCheckingViolation{
	//				ViolatedConstraint: constraint,
	//				InvolvedCase:       caseId,
	//				Timestamp:          timestamp,
	//			})
	//			fmt.Println("New compliance violation for case: ", caseId, " constraint: ", constraint)
	//		}
	//	}
	//}
	if addEventFlag {
		psm.ProcessState.EventLog = append(psm.ProcessState.EventLog, eventLogEntry)
	}
	//fmt.Println("Time for compliance checking: ", time.Since(firtsTs).Seconds())
	//Clear old events
	if len(psm.ProcessState.EventLog) == 150 {
		//clear the events log by removing the first 100 events
		psm.ProcessState.EventLog = psm.ProcessState.EventLog[100:]
	}
	fmt.Println("Event number: ", psm.ProcessState.Counter)
	// Incremental computation of duration statistics
	duration := time.Since(firtsTs).Seconds()
	durationFromStart := time.Since(psm.runStarted)
	psm.totalDuration += duration
	if duration < psm.minDuration && duration != 0 {
		psm.minDuration = duration
	}
	if duration > psm.maxDuration {
		psm.maxDuration = duration
	}
	averageDuration := psm.totalDuration / float64(psm.ProcessState.Counter)
	fmt.Printf("Time from start of the run:%f, Current average duration (ms): %f, Min duration (ms): %f, Max duration (ms): %f\n", durationFromStart.Seconds(), averageDuration, psm.minDuration, psm.maxDuration)
}

//func (psm *ProcessStateManager) prepareEventLog() map[string]interface{} {
//	elaboratedLog := map[string]interface{}{}
//	elaboratedLog["events"] = []interface{}{}
//	//For each event in the log
//	for _, event := range psm.ProcessState.Events {
//		singleEvent := map[string]interface{}{}
//		singleEvent["trace_concept_name"] = event.CaseId
//		singleEvent["concept_name"] = event.Activity
//		singleEvent["timestamp"] = event.Timestamp
//		for k, v := range event.Data {
//			singleEvent[k] = v
//		}
//		elaboratedLog["events"] = append(elaboratedLog["events"].([]interface{}), singleEvent)
//	}
//	return elaboratedLog
//}

func (psm *ProcessStateManager) initWorkflowViolation(eventId string, caseId string, timestamp string, error error) {
	//Reset expect next activities for the case
	psm.ProcessState.WfState.expectNextActivities[caseId] = []string{}

	//Generate a new workflow violation with the current event
	wfViolation := WorkflowViolation{
		//GeneratedByEvent:  eventId,
		GeneratedByCase: caseId,
		Timestamp:       timestamp,
		//ErroneousSequence: []string{eventId},
	}
	//Add the new workflow violation to the workflow violations map
	psm.ProcessState.WfState.WorkflowViolations[caseId] = wfViolation
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
		//if the transition is not already in the expect next activities
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
	_, wfViolated := psm.ProcessState.WfState.WorkflowViolations[caseId]
	if wfViolated {
		return fmt.Errorf("Case %v is in an erroneous state", caseId), nil
	}

	return nil, psm.ProcessState.WfState.GetNextActivities(caseId)
}

// Wait for new events from the event channel
func (psm *ProcessStateManager) WaitForEvents() {
	for {
		event := <-psm.EventChannel
		psm.HandleEvent(event.ActivityID, event.CaseID, event.Timestamp, event.Attributes)
	}
}

// function that print the process state nicely
func (psm *ProcessStateManager) PrintProcessState() {
	fmt.Println("Printing process state")
	fmt.Println("Next activities", psm.ProcessState.WfState.expectNextActivities)
	fmt.Println("Control Flow Violations", psm.ProcessState.WfState.WorkflowViolations)
	fmt.Println("Compliance Checking violations", psm.ProcessState.ComplianceCheckingViolations)
}
