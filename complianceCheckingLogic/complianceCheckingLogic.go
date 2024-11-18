package complianceCheckingLogic


    import (
	"context"
	"fmt"
	"github.com/open-policy-agent/opa/rego"
	"sync"
	"log")
    
// Generated process constraints code

var constraintNames = []string{
"iv_antibiotics_within_onehour", "lactic_acid_within_onehour",
}

var constraints = []string{

`package iv_antibiotics_within_onehour
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

# Define a rule to check if the most recent event is "IV Antibiotics"
triage_present[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    count({e | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "ER Sepsis Triage"}) > 0
}

# Pending condition
pending[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "ER Sepsis Triage"
}

# Violation condition
violations[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "IV Antibiotics"
    not iv_antibiotics_within_one_hour[trace_id]
}

# Satisfied condition
satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "IV Antibiotics"
    iv_antibiotics_within_one_hour[trace_id]
}

# Define a rule to check if "IV Antibiotics" happens within one hour after the latest "ER Sepsis Triage"
iv_antibiotics_within_one_hour[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    triage_present[trace_id]
    sepsisTriage := max([time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "ER Sepsis Triage"])
    iv_antibiotics := most_recent_event
    time.parse_rfc3339_ns(iv_antibiotics.timestamp) <= sepsisTriage + 3600000000000
}`,

`package lactic_acid_within_onehour
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

# Define a rule to check if the most recent event is "LacticAcid"
triage_present[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    count({e | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "ER Sepsis Triage"}) > 0
}

# Pending condition
pending[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "ER Sepsis Triage"
}

# Violation condition
violations[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "LacticAcid"
    not lactic_acid_within_one_hour[trace_id]
}

# Satisfied condition
satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "LacticAcid"
    lactic_acid_within_one_hour[trace_id]
}

# Define a rule to check if "LacticAcid" happens within one hour after the latest "ER Sepsis Triage"
lactic_acid_within_one_hour[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    triage_present[trace_id]
    sepsisTriage := max([time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "ER Sepsis Triage"])
    lactic_acid := most_recent_event
    time.parse_rfc3339_ns(lactic_acid.timestamp) <= sepsisTriage + 10800000000000
}`,

}


// Enum for the state of the constraint
type ConstraintState int

const (
	Init     ConstraintState = 0
	Pending  ConstraintState = 1
	Violated ConstraintState = 2
)

type Constraint struct {
	name              string
	preparedEvalQuery rego.PreparedEvalQuery
	ConstraintState   map[string]ConstraintState // State for each case
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
			ConstraintState:   make(map[string]ConstraintState),
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
			constraint.ConstraintState[traceId] = Init
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
					fmt.Println(res)
					log.Fatalf("Failed to convert violation from policy inspection")
				}
				pending, ok := resultValueMap["pending"].(map[string]interface{})
				if !ok {
					fmt.Println(res)
					log.Fatalf("Failed to convert pending from policy inspection")
				}
				satisfied, ok := resultValueMap["satisfied"].(map[string]interface{})
				if !ok {
					fmt.Println(res)
					log.Fatalf("Failed to convert satisfied from policy inspection")
				}
				violationMap[constraintName] = map[string]interface{}{
					"violations": violations,
					"pending":    pending,
					"satisfied":  satisfied,
				}
				//serve una un set di caseId processati
				caseSet := map[string]bool{}
				for caseId := range pending {
					if constraint.ConstraintState[caseId] == Init {
						constraint.ConstraintState[caseId] = Pending
						fmt.Println("Constraint" + constraintName + " pending for case " + caseId)
					}
				}
				for caseId := range violations {
					if constraint.ConstraintState[caseId] == Pending {
						if !caseSet[caseId] {
							constraint.ConstraintState[caseId] = Violated
							caseSet[caseId] = true
							fmt.Println("Constraint" + constraintName + " violated for case " + caseId)
						}
					}
				}
				for caseId := range satisfied {
					if constraint.ConstraintState[caseId] == Pending {
						if !caseSet[caseId] {
							constraint.ConstraintState[caseId] = Init
							caseSet[caseId] = true
							fmt.Println("Constraint" + constraintName + " satisfied for case " + caseId)
						}
					}
				}
			}
		}(constraint)
	}
	wg.Wait()
	return violationMap
}

