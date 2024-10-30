package processStateManager

import (
	"fmt"
	"log"
	"main/complianceCheckingLogic"
	"main/utils/xes"
	"main/workflowLogic"
)

// TODO: we assume that no activity is named "source" or "sink"
// Struct Process State
type ProcessState struct {
	//Maps (set) of all the processed Cases
	Cases map[string]bool
	//Current state of the process
	WfState WorkflowState
	//List of all the events
	Events []Event
	//Compliance checking violations
	ComplianceCheckingViolations map[string]map[string]bool
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
}

// Init a new ProcessStateManager
func InitProcessStateManager(eventChannel chan xes.Event) ProcessStateManager {
	ccLogic, cNames := complianceCheckingLogic.InitComplianceCheckingLogic()
	psm := ProcessStateManager{
		WorkflowLogic:           workflowLogic.InitWorkflowLogic(),
		ComplianceCheckingLogic: ccLogic,
		EventChannel:            eventChannel,
	}
	ccViolation := map[string]map[string]bool{}
	for _, name := range cNames {
		ccViolation[name] = map[string]bool{}
	}
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
	event := Event{
		Activity:  eventId,
		CaseId:    caseId,
		Timestamp: timestamp,
		Data:      data,
	}
	//Check if the event exists in the workflow logic
	//TODO: this should be moved into the event dispatcher
	if psm.WorkflowLogic.GetTransitionIndexByName(eventId) == -1 {
		fmt.Println("Unknown event")
		return
	}
	//Add the event to the list of events
	psm.ProcessState.Events = append(psm.ProcessState.Events, event)
	//If the case is not in the cases map, add it
	if !psm.ProcessState.Cases[caseId] {
		psm.initNewCase(caseId)
		//fmt.Println("New workflow violation for case: ", caseId, " event: ", eventId)
	}
	//Check if the case is already in errouneous workflow state
	_, wfViolated := psm.ProcessState.WfState.WorkflowViolations[caseId]
	if wfViolated {
		//Append the current event to the erroneous sequence
		wfViolation := psm.ProcessState.WfState.WorkflowViolations[caseId]
		wfViolation.ErroneousSequence = append(wfViolation.ErroneousSequence, eventId)
		psm.ProcessState.WfState.WorkflowViolations[caseId] = wfViolation
		fmt.Println("Erronous sequence updated for case: ", caseId, " event: ", eventId)
		return
	}
	//Fire the transition associated with the event
	error := psm.WorkflowLogic.FireTokenIdWithTransitionName(eventId, psm.ProcessState.WfState.CaseTokenId[caseId])
	if error != nil {
		//If the transition failed, generate a new workflow violation
		psm.initWorkflowViolation(eventId, caseId, timestamp, error)
		fmt.Println("New workflow violation for case: ", caseId, " event: ", eventId)
	} else {
		//If the transition was successful, update the workflow state
		psm.updateWorkflowState(caseId)
		fmt.Println("Succesful state update with case: ", caseId, " event: ", eventId, " next activities: ", psm.ProcessState.WfState.GetNextActivities(caseId))
	}
	elaboratedLog := psm.prepareEventLog()
	violationMap := psm.ComplianceCheckingLogic.EvaluateEventLog(elaboratedLog)
	for constraint, result := range violationMap {
		castedResult := result.(map[string]interface{})
		for caseId, _ := range castedResult {
			//TODO: be carefull here, we are assuming that each case id in the constraint map is violation (case:true)
			//TODO: I'm not sure about this, i'don't know if there may be cases in which case:false (no violation). CHECK.
			if !psm.ProcessState.ComplianceCheckingViolations[constraint][caseId] {
				log.Println("New compliance violation for case: ", caseId, " constraint: ", constraint)
				psm.ProcessState.ComplianceCheckingViolations[constraint][caseId] = true
				//TODO: test all the constraints
			}
		}
	}

}
func (psm *ProcessStateManager) prepareEventLog() map[string]interface{} {
	elaboratedLog := map[string]interface{}{}
	elaboratedLog["events"] = []interface{}{}
	//For each event in the log
	for _, event := range psm.ProcessState.Events {
		singleEvent := map[string]interface{}{}
		singleEvent["trace_concept_name"] = event.CaseId
		singleEvent["concept_name"] = event.Activity
		singleEvent["timestamp"] = event.Timestamp
		for k, v := range event.Data {
			singleEvent[k] = v
		}
		elaboratedLog["events"] = append(elaboratedLog["events"].([]interface{}), singleEvent)
	}
	return elaboratedLog
}

func (psm *ProcessStateManager) initWorkflowViolation(eventId string, caseId string, timestamp string, error error) {
	//Reset expect next activities for the case
	psm.ProcessState.WfState.expectNextActivities[caseId] = []string{}

	//Generate a new workflow violation with the current event
	wfViolation := WorkflowViolation{
		GeneratedByEvent:     eventId,
		GeneratedByCase:      caseId,
		Timestamp:            timestamp,
		ViolationDescription: error.Error(),
		ErroneousSequence:    []string{eventId},
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
