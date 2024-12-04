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
"inspect_goods_within_onehour", "separation_of_duty", "shipment_cost", "truck_policy"}

var constraints = []string{

`package inspect_goods_within_onehour
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

# Define a rule to check if the most recent event is "IV Antibiotics"
reached_present[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    count({e | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Truck reached costumer (TRC)"}) > 0
}

# Pending condition
pending[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Truck reached costumer (TRC)"
}

# Violation condition
violations[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Inspect goods (IG)"
    not inspect_goods_within_one_hour[trace_id]
}

# Violation condition, when the trace is over and the constraint is in pending state
violations[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Order reception confirmed (ORC)"
}

# Satisfied condition, when I receive an inspect goods activity and the constraint is in pending state
satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Inspect goods (IG)"
    inspect_goods_within_one_hour[trace_id]
}

# Define a rule to check if "Inspect goods (IG)" happens within one hour after the latest "Truck reached costumer (TRC)"
inspect_goods_within_one_hour[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    reached_present[trace_id]
    last_inspect := max([time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Inspect goods (IG)"])
    reached_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Truck reached costumer (TRC)";time.parse_rfc3339_ns(e.timestamp) < last_inspect]
    reached := min(reached_events) # This will be 0 if reached_events is empty
    inspect := most_recent_event
    time.parse_rfc3339_ns(inspect.timestamp) <= reached + 3600000000000
}`,

`package separation_of_duty
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

# Pending condition 1
pending[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Fill in container (FC)"
}

# Violation condition 1
violations[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Fill in container (FC)"
    same_operator_exists(trace_id, "Fill in container (FC)", "Check container (CC)")
}

# Violation condition 2
violations[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Check container (CC)"
    same_operator_exists(trace_id, "Check container (CC)", "Fill in container (FC)")
}

# Satisfied condition, when the trace is over and the constraint is in pending state
satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Order reception confirmed (ORC)"
    #This below is not needed
    #not same_operator_exists(trace_id, "Fill in container (FC)", "Check container (CC)")
}

# Define a rule to check if the same logistics operator exists for both activities in the same trace
same_operator_exists(trace_id, activity1, activity2) if {
    e1 := input.events[_]
    e1.trace_concept_name == trace_id
    e1.concept_name == activity1
    e2 := input.events[_]
    e2.trace_concept_name == trace_id
    e2.concept_name == activity2
    e1.logistics_operator == e2.logistics_operator
}`,

`package shipment_cost
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

# Pending condition
pending[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Reserve shipment (RS)"
}

# Violation condition
violations[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Drive to manufacturer (DM)"
    not check_cost_condition[trace_id]
}
# Violation condition
violations[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Drive to costumer (DC)"
    not check_cost_condition[trace_id]
}

# Satisfied condition
satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
	most_recent_event.concept_name == "Order reception confirmed (ORC)"
	#This below is not needed, if you are in pending state, the check cost condition is always true if you are in pending state
    check_cost_condition[trace_id]
}

# Define a rule to check the cost condition
check_cost_condition[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    reserve_cost := sum([e.cost | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Reserve shipment (RS)"])
    drive_distance_i := sum([e.km_distance | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Drive to costumer (DC)"])
    drive_distance_m := sum([e.km_distance | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Drive to manufacturer (DM)"])
    reserve_cost <= (drive_distance_i + drive_distance_m) * 3
}`,

`package truck_policy
import rego.v1

# Define the constant for five years in seconds
five_years_in_seconds := 5 * 365 * 24 * 60 * 60

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

# Pending condition
pending[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Select truck (ST)"
}

# Violation condition
violations[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    #Not needed
    #most_recent_event.concept_name == "Select truck (ST)"
    driver_experience_within_five_years[trace_id]
}

# Satisfied condition
satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    #Not needed
    #most_recent_event.concept_name == "Select truck (ST)"
    not driver_experience_within_five_years[trace_id]
}

# Define a rule to check if the driver's experience is within five years
driver_experience_within_five_years[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    event := most_recent_event
    (time.parse_rfc3339_ns(event.timestamp) / 1000000000) - (time.parse_rfc3339_ns(event.license_first_issue) / 1000000000) <= five_years_in_seconds
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

"inspect_goods_within_onehour": {
    Transitions: [][]int{
        {1},    // init -> pending
        {2, 3}, // pending -> violated, satisfied
        {},     // violated -> none
        {0},    // satisfied -> init
        {},     // temporary_satisfied -> none
        {},     // temporary_violated -> none
    },
},

"separation_of_duty": {
    Transitions: [][]int{
        {1},    // init -> pending
        {2, 3}, // pending -> violated, satisfied
        {},     // violated -> none
        {0},    // satisfied -> init
        {},     // temporary_satisfied -> none
        {},     // temporary_violated -> none
    },
},

"shipment_cost": {
    Transitions: [][]int{
        {1},    // init -> pending
        {2, 3}, // pending -> violated, satisfied
        {},     // violated -> none
        {0},    // satisfied -> init
        {},     // temporary_satisfied -> none
        {},     // temporary_violated -> none
    },
},

"truck_policy": {
    Transitions: [][]int{
        {1},    // init -> pending
        {2, 3}, // pending -> violated, satisfied
        {},     // violated -> none
        {0},    // satisfied -> init
        {},     // temporary_satisfied -> none
        {},     // temporary_violated -> none
    },
},

}
