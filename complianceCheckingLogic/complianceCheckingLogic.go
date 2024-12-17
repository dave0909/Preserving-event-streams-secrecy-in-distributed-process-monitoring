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
"iv_antibiotics_within_onehour", "lactic_acid_within_onehour"}

var constraints = []string{

`package iv_antibiotics_within_onehour
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

temporary_violated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "ER Sepsis Triage"
}
#temporary satisfied condition if the last event is "Inspect goods (IG)" and the difference with the older Truck reached costumer (TRC) is less than one hour
temporary_satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "IV Antibiotics"
    #Get the older Truck reached costumer (TRC) event
    reached_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "ER Sepsis Triage"]
    reached := min(reached_events) # This will be 0 if reached_events is empty
    #check if the fime difference is less than one hour
    time.parse_rfc3339_ns(most_recent_event.timestamp) - reached <= 3600000000000
}
#Violation condition if the last event is "Inspect goods (IG)" and the difference with the older Truck reached costumer (TRC) is more than one hour
violations[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "IV Antibiotics"
    #Get the older Truck reached costumer (TRC) event
    reached_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "ER Sepsis Triage"]
    reached := min(reached_events) # This will be 0 if reached_events is empty
    #check if the fime difference is less than one hour
    time.parse_rfc3339_ns(most_recent_event.timestamp) - reached > 3600000000000
}
`,

`package lactic_acid_within_onehour
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]


temporary_violated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "ER Sepsis Triage"
}
#temporary satisfied condition if the last event is "LacticAcid" and the difference with the older ER Sepsis Triage is less than one hour
temporary_satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "LacticAcid"
    #Get the older Truck reached costumer (TRC) event
    reached_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "ER Sepsis Triage"]
    reached := min(reached_events) # This will be 0 if reached_events is empty
    #check if the fime difference is less than one hour
    time.parse_rfc3339_ns(most_recent_event.timestamp) - reached <= 10800000000000
}
#Violation condition if the last event is "Inspect goods (IG)" and the difference with the older Truck reached costumer (TRC) is more than one hour
violations[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "LacticAcid"
    #Get the older Truck reached costumer (TRC) event
    reached_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "ER Sepsis Triage"]
    reached := min(reached_events) # This will be 0 if reached_events is empty
    #check if the fime difference is less than one hour
    time.parse_rfc3339_ns(most_recent_event.timestamp) - reached > 10800000000000
}
`,

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
func (ccl *ComplianceCheckingLogic) EvaluateEventLog(eventLog map[string][]map[string]interface{}) map[string]interface{} {
	lastEvent := eventLog["events"][len(eventLog["events"])-1]
	traceId := lastEvent["trace_concept_name"].(string)
	for _, constraint := range ccl.preparedConstraints {
		if _, ok := constraint.ConstraintState[traceId]; !ok {
			constraint.ConstraintState[traceId] = Init //Init state
		}
	}
	resultMap := map[string]interface{}{}
	var wg sync.WaitGroup
	var mu sync.Mutex
    evInputLog:=rego.EvalInput(eventLog)
    for _, constraint := range ccl.preparedConstraints {
        wg.Add(1)
        go func(constraint Constraint, evInputLog rego.EvalOption) {
            defer wg.Done()
            res, err := constraint.preparedEvalQuery.Eval(ccl.ctx, evInputLog)
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
                        resultMap[traceId] = Pending
                    }
                }
                for caseId := range temporarySatisfied {
                    if constraint.fsm.HasTransition(int(constraint.ConstraintState[traceId]), 4) {
                        constraint.ConstraintState[caseId] = TemporarySatisfied
                        fmt.Println("Constraint ", constraintName, "in temporary satisfied state for case ", caseId)
                        resultMap[traceId] = TemporarySatisfied

                    }
                }
                for caseId := range temporaryViolated {
                    if constraint.fsm.HasTransition(int(constraint.ConstraintState[traceId]), 5) {
                        constraint.ConstraintState[caseId] = TemporaryViolated
                        fmt.Println("Constraint ", constraintName, "in temporary violated state for case ", caseId)
                        resultMap[traceId] = TemporaryViolated
                    }
                }
                for caseId := range violations {
                    if constraint.fsm.HasTransition(int(constraint.ConstraintState[traceId]), 2) {
                        constraint.ConstraintState[caseId] = Violated
                        fmt.Println("Constraint ", constraintName, "in violated state for case ", caseId)
                        resultMap[traceId] = Violated
                    }
                }
                for caseId := range satisfied {
                    if constraint.fsm.HasTransition(int(constraint.ConstraintState[traceId]), 3) {
                        constraint.ConstraintState[caseId] = Satisfied
                        fmt.Println("Constraint ", constraintName, "in satisfied state for case ", caseId)
                        resultMap[traceId] = Satisfied
                    }
                }
            }
        }(constraint,evInputLog)
	}
	wg.Wait()
	return resultMap
}


var fsmMap = map[string]*CustomFSM{

"iv_antibiotics_within_onehour": {
    Transitions: [][]int{
        {5},
        {},
        {},
        {},
        {2},
        {2, 4},
    },
},

"lactic_acid_within_onehour": {
    Transitions: [][]int{
        {5},
        {},
        {},
        {},
        {2},
        {2, 4},
    },
},

}
