package main

import (
	"context"
	"fmt"
	"github.com/looplab/fsm"
	"github.com/open-policy-agent/opa/rego"
	"log"
	"sync"
)

var constraintNames = []string{
	"inspect_goods_within_onehour", "truck_policy",
}

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
}
`, `package truck_policy
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
}
`,
}

// Enum for the state of the constraint
type ConstraintState int

//const (
//	Init     ConstraintState = 0
//	Pending  ConstraintState = 1
//	Violated ConstraintState = 2
//)

type Constraint struct {
	name              string
	preparedEvalQuery rego.PreparedEvalQuery
	ConstraintState   map[string]*fsm.FSM
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
			ConstraintState:   make(map[string]*fsm.FSM),
		})
	}

	return ccLogic, constraintNames
}

// Evaluate the event log with the prepared constraints
func (ccl *ComplianceCheckingLogic) EvaluateEventLog(eventLog map[string]interface{}) map[string]interface{} {

	//Get the last event from the event log
	lastEvent := eventLog["events"].([]map[string]interface{})[len(eventLog["events"].([]map[string]interface{}))-1]
	//Get the trace_id of the last event
	traceId := lastEvent["trace_concept_name"].(string)
	//If the trace_id is not in the constraint state, add it
	for _, constraint := range ccl.preparedConstraints {
		if _, ok := constraint.ConstraintState[traceId]; !ok {
			constraint.ConstraintState[traceId] = getFSMfromConstraintName(constraint.name)
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
				violations, ok := resultValueMap["violations"].(map[string]interface{})
				if !ok {
					//log.Fatalf("Failed to convert violation from policy inspection")
				}
				pending, ok := resultValueMap["pending"].(map[string]interface{})
				if !ok {
					//log.Fatalf("Failed to convert pending from policy inspection")
				}
				satisfied, ok := resultValueMap["satisfied"].(map[string]interface{})
				if !ok {
					//log.Fatalf("Failed to convert satisfied from policy inspection")
				}
				temporarySatisfied, ok := resultValueMap["temporary_satisfied"].(map[string]interface{})
				if !ok {
					//log.Fatalf("Failed to convert satisfied from policy inspection")
				}
				temporaryViolated, ok := resultValueMap["temporary_violated"].(map[string]interface{})
				if !ok {
					//log.Fatalf("Failed to convert satisfied from policy inspection")
				}
				violationMap[constraintName] = map[string]interface{}{
					"violations":          violations,
					"pending":             pending,
					"satisfied":           satisfied,
					"temporary_satisfied": temporarySatisfied,
					"temporary_violated":  temporaryViolated,
				}
				//fmt.Println(violationMap)
				for caseId := range pending {
					//If the violation is in the next states
					if constraint.ConstraintState[caseId].Can("Pending") {
						//Change the state of the constraint to violated
						constraint.ConstraintState[caseId].Event(ccl.ctx, "Pending")
					}
				}
				for caseId := range violations {
					//If the violation is in the next states
					if constraint.ConstraintState[caseId].Can("Violated") {
						//Change the state of the constraint to violated
						constraint.ConstraintState[caseId].Event(ccl.ctx, "Violated")
					}
				}
				for caseId := range satisfied {
					if constraint.ConstraintState[caseId].Can("Satisfied") {
						//Change the state of the constraint to violated
						constraint.ConstraintState[caseId].Event(ccl.ctx, "Satisfied")
					}
				}
				for caseId := range temporarySatisfied {
					if constraint.ConstraintState[caseId].Can("TemporarySatisfied") {
						//Change the state of the constraint to violated
						constraint.ConstraintState[caseId].Event(ccl.ctx, "TemporarySatisfied")
					}
				}
				for caseId := range temporaryViolated {
					if constraint.ConstraintState[caseId].Can("TemporaryViolated") {
						//Change the state of the constraint to violated
						constraint.ConstraintState[caseId].Event(ccl.ctx, "TemporaryViolated")
					}
				}

			}
		}(constraint)
	}
	wg.Wait()
	return violationMap
}

// main for testing
func main() {
	ccLogic, _ := InitComplianceCheckingLogic()

	//Test the inspect_goods_within_onehour
	eventLog := map[string]interface{}{
		"events": []map[string]interface{}{
			{
				"trace_concept_name": "1",
				"concept_name":       "Truck reached costumer (TRC)",
				"timestamp":          "2030-06-01T00:00:00Z",
			},
		}}
	_ = ccLogic.EvaluateEventLog(eventLog)
	fmt.Println("Constraint state: ", ccLogic.preparedConstraints[0].ConstraintState["1"].Current())

	//// the state of the fill in container constraint
	//
	//Test different consecutive evaluations by adding events to the eventLog. I want to obtain a pending state and a successful state
	eventLog = map[string]interface{}{
		"events": []map[string]interface{}{
			{
				"trace_concept_name": "1",
				"concept_name":       "Truck reached costumer (TRC)",
				"timestamp":          "2030-06-01T00:00:00Z",
			},
			{
				"trace_concept_name": "1",
				"concept_name":       "Truck reached costumer (TRC)",
				"timestamp":          "2030-06-01T02:00:00Z",
			},
		}}
	_ = ccLogic.EvaluateEventLog(eventLog)
	fmt.Println("Constraint state: ", ccLogic.preparedConstraints[0].ConstraintState["1"].Current())

	//Test different consecutive evaluations by adding events to the eventLog. I want to obtain a pending state and a successful state
	eventLog = map[string]interface{}{
		"events": []map[string]interface{}{
			{
				"trace_concept_name": "1",
				"concept_name":       "Truck reached costumer (TRC)",
				"timestamp":          "2030-06-01T00:00:00Z",
			},
			{
				"trace_concept_name": "1",
				"concept_name":       "Truck reached costumer (TRC)",
				"timestamp":          "2030-06-01T00:10:00Z",
			},
			{
				"trace_concept_name": "1",
				"concept_name":       "Inspect goods (IG)",
				"timestamp":          "2030-06-01T03:00:00Z",
			},
		}}
	_ = ccLogic.EvaluateEventLog(eventLog)
	//print the state of the fill in container constraint
	fmt.Println("Constraint state: ", ccLogic.preparedConstraints[0].ConstraintState["1"].Current())

	//Test truck_policy
	eventLog = map[string]interface{}{
		"events": []map[string]interface{}{
			{
				"trace_concept_name":  "1",
				"concept_name":        "Select truck (ST)",
				"timestamp":           "2030-06-01T00:00:00Z",
				"license_first_issue": "2016-06-01T00:00:00Z",
			},
		}}
	_ = ccLogic.EvaluateEventLog(eventLog)
	fmt.Println("Constraint state: ", ccLogic.preparedConstraints[1].ConstraintState["1"].Current())
	////print the state of the select truck constraint

}

func getFSMfromConstraintName(constraintName string) *fsm.FSM {
	switch constraintName {
	case "inspect_goods_within_onehour":
		finitestate := fsm.NewFSM(
			"Init",
			fsm.Events{
				{Name: "Pending", Src: []string{"Init"}, Dst: "Pending"},
				{Name: "Violated", Src: []string{"Pending"}, Dst: "Violated"},
				{Name: "Satisfied", Src: []string{"Pending"}, Dst: "Init"},
			},
			fsm.Callbacks{},
		)
		return finitestate
	case "truck_policy":
		finitestate := fsm.NewFSM(
			"Init",
			fsm.Events{
				{Name: "Pending", Src: []string{"Init"}, Dst: "Pending"},
				{Name: "Violated", Src: []string{"Pending"}, Dst: "Violated"},
				{Name: "Satisfied", Src: []string{"Pending"}, Dst: "Init"},
			},
			fsm.Callbacks{},
		)
		return finitestate

	}
	return &fsm.FSM{}
}

//package main
//
//import (
//	"context"
//	"fmt"
//	"github.com/open-policy-agent/opa/rego"
//	"log"
//	"sync"
//)
//
//// Enum for the state of the constraint
//type ConstraintState int
//
//const (
//	Init     ConstraintState = 0
//	Pending  ConstraintState = 1
//	Violated ConstraintState = 2
//)
//
//type Constraint struct {
//	name              string
//	preparedEvalQuery rego.PreparedEvalQuery
//	ConstraintState   ConstraintState
//}
//
//// Generated process constraints code
//
//var constraintNames = []string{
//	"lactic_acid_within_onehour",
//}
//
//var constraints = []string{
//	`package lactic_acid_within_onehour
//import rego.v1
//
//# Get the most recent event
//most_recent_event := input.events[count(input.events) - 1]
//
//# Define a rule to check if the most recent event is "LacticAcid"
//triage_present[trace_id] if {
//    trace_id := most_recent_event.trace_concept_name
//    count({e | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "ER Sepsis Triage"}) > 0
//}
//
//# Pending condition
//pending[trace_id] if {
//    trace_id := most_recent_event.trace_concept_name
//    most_recent_event.concept_name == "ER Sepsis Triage"
//}
//
//# Violation condition
//violations[trace_id] if {
//    trace_id := most_recent_event.trace_concept_name
//    most_recent_event.concept_name == "LacticAcid"
//    not lactic_acid_within_one_hour[trace_id]
//}
//
//# Satisfied condition
//satisfied[trace_id] if {
//    trace_id := most_recent_event.trace_concept_name
//    most_recent_event.concept_name == "LacticAcid"
//    lactic_acid_within_one_hour[trace_id]
//}
//
//# Define a rule to check if "LacticAcid" happens within one hour after the latest "ER Sepsis Triage"
//lactic_acid_within_one_hour[trace_id] if {
//    trace_id := most_recent_event.trace_concept_name
//    triage_present[trace_id]
//    sepsisTriage := max([time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "ER Sepsis Triage"])
//    lactic_acid := most_recent_event
//    time.parse_rfc3339_ns(lactic_acid.timestamp) <= sepsisTriage + 10800000000000
//}
//`,
//}
//
//type ComplianceCheckingLogic struct {
//	preparedConstraints []rego.PreparedEvalQuery
//	ctx                 context.Context
//
//}
//
//// Function that creates a prepared constraint for each constraint
//func InitComplianceCheckingLogic() (ComplianceCheckingLogic, []string) {
//	ctx := context.TODO()
//	ccLogic := ComplianceCheckingLogic{
//		preparedConstraints: []rego.PreparedEvalQuery{},
//		ctx:                 ctx,
//	}
//	for i, constraint := range constraints {
//		query, err := rego.New(
//			rego.Query("data."+constraintNames[i]),
//			rego.Module(constraintNames[i], constraint),
//		).PrepareForEval(ctx)
//		if err != nil {
//			log.Fatal(err)
//		}
//		ccLogic.preparedConstraints = append(ccLogic.preparedConstraints, query)
//	}
//	return ccLogic, constraintNames
//}
//
//// Evaluate the event log with the prepared constraints
//func (ccl *ComplianceCheckingLogic) EvaluateEventLog(eventLog map[string]interface{}) map[string]interface{} {
//	violationMap := map[string]interface{}{}
//	var wg sync.WaitGroup
//	var mu sync.Mutex
//
//	for _, preparedConstraint := range ccl.preparedConstraints {
//		wg.Add(1)
//		go func(preparedConstraint rego.PreparedEvalQuery) {
//			defer wg.Done()
//			res, err := preparedConstraint.Eval(ccl.ctx, rego.EvalInput(eventLog))
//			if err != nil {
//				fmt.Println(err)
//				return
//			}
//			mu.Lock()
//			defer mu.Unlock()
//			for constraintName := range preparedConstraint.Modules() {
//				resultValue := res[0].Expressions[0].Value
//				resultValueMap, ok := resultValue.(map[string]interface{})
//				if !ok {
//					fmt.Println(res)
//					log.Fatalf("Failed to convert result from policy inspection")
//				}
//				violations, ok := resultValueMap["violations"].(map[string]interface{})
//				if !ok {
//					fmt.Println(res)
//					log.Fatalf("Failed to convert violation from policy inspection")
//				}
//				pending, ok := resultValueMap["pending"].(map[string]interface{})
//				if !ok {
//					fmt.Println(res)
//					log.Fatalf("Failed to convert pending from policy inspection")
//				}
//				satisfied, ok := resultValueMap["satisfied"].(map[string]interface{})
//				if !ok {
//					fmt.Println(res)
//					log.Fatalf("Failed to convert pending from policy inspection")
//				}
//				violationMap[constraintName] = map[string]interface{}{
//					"violations": violations,
//					"pending":    pending,
//					"satisfied":  satisfied,
//				}
//			}
//		}(preparedConstraint)
//	}
//	wg.Wait()
//	return violationMap
//}
//
//// main for testing
//func main() {
//	ccLogic, _ := InitComplianceCheckingLogic()
//	// Test lactic_acid_within_onehour
//	eventLog := map[string]interface{}{
//		"events": []map[string]interface{}{
//			{
//				"trace_concept_name": "1",
//				"concept_name":       "ER Sepsis Triage",
//				"timestamp":          "2021-06-01T00:00:00Z",
//			},
//			{
//				"trace_concept_name": "1",
//				"concept_name":       "LacticAcid",
//				"timestamp":          "2021-06-01T20:00:00Z",
//			},
//		}}
//	violations := ccLogic.EvaluateEventLog(eventLog)
//	//Test different consecutive evaluations by adding events to the eventLog. I want to obtein a pending state and a successful state
//	fmt.Println(violations)
//
//}
