package complianceCheckingLogic


import (
    "context"
    "fmt"
    "github.com/open-policy-agent/opa/rego"
    "sync"
    "log"
)
type ConstraintState int
const (
	Init ConstraintState = iota
	Pending
	Violated
	Satisfied
	TemporarySatisfied
	TemporaryViolated
)

// Custom FSM struct using an integer matrix for transitions
type CustomFSM struct {
	Transitions [][]int
}

// Generated process constraints code

var constraintNames = []string{
"appeal_notcollect", "balance_notcollect", "case_start", "delay_no_collect", "dismissal_notinsert", "insert_and_notify_notcollect", "no_insert_no_collect", "two_create_send_notinsertfine"}

var constraints = []string{

`package appeal_notcollect

import rego.v1


# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]


#Temporary satisfied if the last event is the last event is "Appeal to Judge"
temporary_satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Appeal to Judge"
}

#Temporary satisfied if the last event is or "Send Appeal to Prefecture"
temporary_satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Send Appeal to Prefecture"
}

#Violated if "Send for Credit Collection" activity exists in the trace
violated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    some e1; e1 = input.events[_]; e1.trace_concept_name == trace_id; e1.concept_name == "Send for Credit Collection"
}
`,

`package balance_notcollect

import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

#Satisfied if the sum af the paymentAmount attribute of all the "Payment" activities is greater than 10
satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    sum([e.paymentAmount | e := input.events; e.concept_name == "Payment"]) > 10
}

#Temporary satisfied if the sum af the paymentAmount attribute of all the "Payment" activities is less or equal than 10
temporary_satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    sum([e.paymentAmount | e := input.events; e.concept_name == "Payment"]) <= 10
}

#Violated if the "Send for Credit Collection" activity exists in the trace
temporary_violated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    some e1; e1 = input.events[_]; e1.trace_concept_name == trace_id; e1.concept_name == "Send for Credit Collection"
}`,

`package case_start

import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]
#Init a Timestamp variable of the first event in the log is 2000-01-01 00:00:00+00:00
first_event_timestamp := time.parse_rfc3339_ns("2000-01-01T00:00:00Z")

temporary_satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name != "Send for Credit Collection"
    #The timestamp of the event is after 4481 days from the first event
    (time.parse_rfc3339_ns(most_recent_event.timestamp) / 1000000000) - (first_event_timestamp / 1000000000) >= 4481 * 24 * 60 * 60
}

#Check among the pasts events if there is a "Send for Credit Collection" event
violated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    some e1; e1 = input.events[_]; e1.trace_concept_name == trace_id; e1.concept_name == "Send for Credit Collection"
}

satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    #The timestamp of the event is after 4481 days from the first event
    (time.parse_rfc3339_ns(most_recent_event.timestamp) / 1000000000) - (first_event_timestamp / 1000000000) < 4481 * 24 * 60 * 60
}
`,

`package delay_no_collect
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    some e1; e1 = input.events[_]; e1.trace_concept_name == trace_id; e1.concept_name == "Add penalty"
    some e2; e2 = input.events[_]; e2.trace_concept_name == trace_id; e2.concept_name == "Payment"
    (time.parse_rfc3339_ns(e2.timestamp) / 1000000000) - (time.parse_rfc3339_ns(e1.timestamp) / 1000000000) > 3 * 24 * 60 * 60
}

# Violated if the last event is "Send for Credit Collection"
violated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Send for Credit Collection"
}

temporary_satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    some e3; e3 = input.events[_]; e3.trace_concept_name == trace_id; e3.concept_name == "Add penalty"
    some e4; e4 = input.events[_]; e4.trace_concept_name == trace_id; e4.concept_name == "Payment"
    (time.parse_rfc3339_ns(e4.timestamp) / 1000000000) - (time.parse_rfc3339_ns(e3.timestamp) / 1000000000) <= 3 * 24 * 60 * 60
}

temporary_violated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name != "Send for Credit Collection"
}`,

`package dismissal_notinsert

import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

#Define a set of strings that represent the events that are considered as "Dismissal"
dismissal_events := {"2", "3", "5", "A", "B", "E", "F", "I", "J", "K", "M", "N", "Q", "R", "T", "U", "V"}

#Temporary satisfied if the last event is the last event is "Create Fine" and the "dismissal" attribute is in the dismissal_events set
temporary_satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Create Fine"
    most_recent_event.dismissal in dismissal_events
}

#Violated if the "Insert Fine" activity exists in the trace
violated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    some e1; e1 = input.events[_]; e1.trace_concept_name == trace_id; e1.concept_name == "Insert Fine Notification"
}`,

`package insert_and_notify_notcollect
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

temporary_satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name != "Insert Date Appeal to Prefecture"
}

temporary_violated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    #Send for Credit Collection extists in the trace
    some e1; e1 = input.events[_]; e1.trace_concept_name == trace_id; e1.concept_name == "Send for Credit Collection"
}

satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Notify Result Appeal to Offender"
}`,

`package no_insert_no_collect
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Insert Fine Notification"
}

temporary_satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name != "Insert Fine Notification"
    most_recent_event.concept_name != "Send for Credit Collection"
}
temporary_violated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Send for Credit Collection"
}


`,

`package two_create_send_notinsertfine
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

# Temporary satisfied if 1) Create Fine is in the past events 2) Send Fine is in the past events 3) There are only two events in the trace_id
temporary_satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    some e1; e1 = input.events[_]; e1.trace_concept_name == trace_id; e1.concept_name == "Create Fine"
    some e2; e2 = input.events[_]; e2.trace_concept_name == trace_id; e2.concept_name == "Send Fine"
    count([e | e := input.events; e.trace_concept_name == trace_id]) == 2
}

# Violated if the "Insert Fine" is the last event of the trace
violated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Insert Fine Notification"
}

# Satisfied if the last event is not "Insert Fine Notification"
satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name != "Insert Fine Notification"
}`,

}


// Method to check possible next states
func (fsm *CustomFSM) PossibleNextStates(currentState int) []int {
	return fsm.Transitions[currentState]
}

// Method to check if there is a transition from state s1 to state s2
func (fsm *CustomFSM) HasTransition(s1, s2 int) bool {
	for _, nextState := range fsm.PossibleNextStates(s1) {
		if nextState == s2 {
			return true
		}
	}
	return false
}
type Constraint struct {
	name              string
	preparedEvalQuery rego.PreparedEvalQuery
	fsm               *CustomFSM
	ConstraintState   map[string]ConstraintState
}

type ComplianceCheckingLogic struct {
	preparedConstraints []Constraint
	ctx                 context.Context
}

// Function that creates a prepared constraint for each constraint
func InitComplianceCheckingLogic() (ComplianceCheckingLogic, []string) {
	ctx := context.TODO()
	ccLogic := ComplianceCheckingLogic{
		preparedConstraints: []Constraint{},
		ctx:                 ctx,
	}
	for i, constraint := range constraints {
		query, err := rego.New(
			rego.Query("data."+constraintNames[i]),
			rego.Module(constraintNames[i], constraint),
		).PrepareForEval(ctx)
		if err != nil {
			log.Fatal(err)
		}
		ccLogic.preparedConstraints = append(ccLogic.preparedConstraints, Constraint{
			name:              constraintNames[i],
			preparedEvalQuery: query,
			fsm:               fsmMap[constraintNames[i]],
			ConstraintState:   make(map[string]ConstraintState),
		})
	}

	return ccLogic, constraintNames
}

// Evaluate the event log with the prepared constraints
func (ccl *ComplianceCheckingLogic) EvaluateEventLog(eventLog map[string]interface{}) map[string]interface{} {
	lastEvent := eventLog["events"].([]map[string]interface{})[len(eventLog["events"].([]map[string]interface{}))-1]
	traceId := lastEvent["trace_concept_name"].(string)
	for _, constraint := range ccl.preparedConstraints {
		if _, ok := constraint.ConstraintState[traceId]; !ok {
			constraint.ConstraintState[traceId] = Init // Init state
		}
	}
	violationMap := map[string]interface{}{}
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, constraint := range ccl.preparedConstraints {
		wg.Add(1)
		go func(constraint Constraint) {
			defer wg.Done()
			res, err := constraint.preparedEvalQuery.Eval(ccl.ctx, rego.EvalInput(eventLog))
			if err != nil {
				fmt.Println(err)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			for constraintName := range constraint.preparedEvalQuery.Modules() {
				resultValue := res[0].Expressions[0].Value
				resultValueMap, ok := resultValue.(map[string]interface{})
				if !ok {
					fmt.Println(res)
					log.Fatalf("Failed to convert result from policy inspection")
				}
				//currentState := constraint.ConstraintState[traceId]
				//nextStates := constraint.fsm.PossibleNextStates(int(currentState))
				violations, ok := resultValueMap["violations"].(map[string]interface{})
				if !ok {
				}
				pending, ok := resultValueMap["pending"].(map[string]interface{})
				if !ok {
				}
				satisfied, ok := resultValueMap["satisfied"].(map[string]interface{})
				if !ok {
				}
				temporarySatisfied, ok := resultValueMap["temporary_satisfied"].(map[string]interface{})
				if !ok {
				}
				temporaryViolated, ok := resultValueMap["temporary_violated"].(map[string]interface{})
				if !ok {
				}
				for caseId := range pending {
					if constraint.fsm.HasTransition(int(constraint.ConstraintState[traceId]), 1) {
						constraint.ConstraintState[caseId] = Pending
						fmt.Println("Constraint ", constraintName, "in pending state for case ", caseId)

					}
				}
				for caseId := range temporarySatisfied {
					if constraint.fsm.HasTransition(int(constraint.ConstraintState[traceId]), 4) {
						constraint.ConstraintState[caseId] = TemporarySatisfied
						fmt.Println("Constraint ", constraintName, "in temporary satisfied state for case ", caseId)

					}
				}
				for caseId := range temporaryViolated {
					if constraint.fsm.HasTransition(int(constraint.ConstraintState[traceId]), 5) {
						constraint.ConstraintState[caseId] = TemporaryViolated
						fmt.Println("Constraint ", constraintName, "in temporary violated state for case ", caseId)

					}
				}
				for caseId := range violations {
					if constraint.fsm.HasTransition(int(constraint.ConstraintState[traceId]), 2) {
						constraint.ConstraintState[caseId] = Violated
						fmt.Println("Constraint ", constraintName, "in violated state for case ", caseId)

					}
				}
				for caseId := range satisfied {
					if constraint.fsm.HasTransition(int(constraint.ConstraintState[traceId]), 3) {
						constraint.ConstraintState[caseId] = Satisfied
						fmt.Println("Constraint ", constraintName, "in satisfied state for case ", caseId)
					}
				}
				violationMap[constraintName] = map[string]interface{}{
					"violations":          resultValueMap["violations"],
					"pending":             resultValueMap["pending"],
					"satisfied":           resultValueMap["satisfied"],
					"temporary_satisfied": resultValueMap["temporary_satisfied"],
					"temporary_violated":  resultValueMap["temporary_violated"],
				}
			}
		}(constraint)
	}
	wg.Wait()
	return violationMap
}


var fsmMap = map[string]*CustomFSM{

"appeal_notcollect": {
    Transitions: [][]int{
        {4},
        {},
        {},
        {},
        {2},
        {},
    },
},

"balance_notcollect": {
    Transitions: [][]int{
        {4, 3},
        {},
        {},
        {},
        {5, 3},
        {3},
    },
},

"case_start": {
    Transitions: [][]int{
        {4, 3},
        {},
        {},
        {},
        {2},
        {},
    },
},

"delay_no_collect": {
    Transitions: [][]int{
        {5, 4, 3},
        {},
        {},
        {},
        {2},
        {4, 3},
    },
},

"dismissal_notinsert": {
    Transitions: [][]int{
        {4},
        {},
        {},
        {},
        {2},
        {},
    },
},

"insert_and_notify_notcollect": {
    Transitions: [][]int{
        {4, 3},
        {},
        {},
        {},
        {3, 5},
        {3},
    },
},

"no_insert_no_collect": {
    Transitions: [][]int{
        {3, 5, 4},
        {},
        {},
        {},
        {3, 5},
        {3},
    },
},

"two_create_send_notinsertfine": {
    Transitions: [][]int{
        {4},
        {},
        {},
        {},
        {2, 3},
        {},
    },
},

}
