package complianceCheckingLogic

import (
	"context"
	"fmt"
	"github.com/looplab/fsm"
	"github.com/open-policy-agent/opa/rego"
	"log"
	"sync"
)

// Generated process constraints code

var constraintNames = []string{
"iv_antibiotics_within_onehour", "lactic_acid_within_onehour"}

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
#Add here violation conditions: when the last events are received

# Satisfied condition
satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "IV Antibiotics"
    iv_antibiotics_within_one_hour[trace_id]
}

## Define a rule to check if "IV Antibiotics" happens within one hour after the latest "ER Sepsis Triage"
#iv_antibiotics_within_one_hour[trace_id] if {
#    trace_id := most_recent_event.trace_concept_name
#    #This is not needed as if we are in pending state, the triage is always present
#    triage_present[trace_id]
#    sepsisTriage := max([time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "ER Sepsis Triage"])
#    iv_antibiotics := most_recent_event
#    time.parse_rfc3339_ns(iv_antibiotics.timestamp) <= sepsisTriage + 3600000000000
#}
# Define a rule to check if "IV Antibiotics" happens within one hour after the latest "ER Sepsis Triage"
iv_antibiotics_within_one_hour[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    triage_present[trace_id]
    last_iv_antibiotics := max([time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "IV Antibiotics"])
    triage_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "ER Sepsis Triage"; time.parse_rfc3339_ns(e.timestamp) < last_iv_antibiotics]
    sepsisTriage := min(triage_events) # This will be 0 if triage_events is empty
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
#Add here violation conditions: when the last events are received


# Satisfied condition
satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "LacticAcid"
    lactic_acid_within_one_hour[trace_id]
}

## Define a rule to check if "LacticAcid" happens within one hour after the latest "ER Sepsis Triage"
#lactic_acid_within_one_hour[trace_id] if {
#    trace_id := most_recent_event.trace_concept_name
#    #This is not needed as if we are in pending state, the triage is always present
#    triage_present[trace_id]
#    sepsisTriage := max([time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "ER Sepsis Triage"])
#    lactic_acid := most_recent_event
#    time.parse_rfc3339_ns(lactic_acid.timestamp) <= sepsisTriage + 10800000000000
#}
## Define a rule to check if "LacticAcid" happens within one hour after the latest "ER Sepsis Triage"

lactic_acid_within_one_hour[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    triage_present[trace_id]
    # Get the last lactic acid event
    last_lactic_acid := max([time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "LacticAcid"])
    #Get the triage events before the last lactic acid event
    triage_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "ER Sepsis Triage"; time.parse_rfc3339_ns(e.timestamp) < last_lactic_acid]
    sepsisTriage := min(triage_events) # This will be 0 if triage_events is empty
    lactic_acid := most_recent_event
    time.parse_rfc3339_ns(lactic_acid.timestamp) <= sepsisTriage + 3600000000000
}`,

}


type Constraint struct {
    name              string
    preparedEvalQuery rego.PreparedEvalQuery
    ConstraintState   map[string]*fsm.FSM
}

type ComplianceCheckingLogic struct {
    preparedConstraints []Constraint
    ctx                 context.Context
}

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

func (ccl *ComplianceCheckingLogic) EvaluateEventLog(eventLog map[string]interface{}) map[string]interface{} {
    lastEvent := eventLog["events"].([]map[string]interface{})[len(eventLog["events"].([]map[string]interface{}))-1]
    traceId := lastEvent["trace_concept_name"].(string)
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
                violationMap[constraintName] = map[string]interface{}{
                    "violations":          violations,
                    "pending":             pending,
                    "satisfied":           satisfied,
                    "temporary_satisfied": temporarySatisfied,
                    "temporary_violated":  temporaryViolated,
                }
                for caseId := range pending {
                    if constraint.ConstraintState[caseId].Can("Pending") {
                        constraint.ConstraintState[caseId].Event(ccl.ctx, "Pending")
                    }
                }
                for caseId := range violations {
                    if constraint.ConstraintState[caseId].Can("Violated") {
                        constraint.ConstraintState[caseId].Event(ccl.ctx, "Violated")
                    }
                }
                for caseId := range satisfied {
                    if constraint.ConstraintState[caseId].Can("Satisfied") {
                        constraint.ConstraintState[caseId].Event(ccl.ctx, "Satisfied")
                    }
                }
                for caseId := range temporarySatisfied {
                    if constraint.ConstraintState[caseId].Can("TemporarySatisfied") {
                        constraint.ConstraintState[caseId].Event(ccl.ctx, "TemporarySatisfied")
                    }
                }
                for caseId := range temporaryViolated {
                    if constraint.ConstraintState[caseId].Can("TemporaryViolated") {
                        constraint.ConstraintState[caseId].Event(ccl.ctx, "TemporaryViolated")
                    }
                }
            }
        }(constraint)
    }
    wg.Wait()
    return violationMap
}


func getFSMfromConstraintName(constraintName string) *fsm.FSM {
    switch constraintName {
case "iv_antibiotics_within_onehour":
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

case "lactic_acid_within_onehour":
    finitestate := fsm.NewFSM(
        "Init",
        fsm.Events{
            {Name: "Satisfied", Src: []string{"Init"}, Dst: "Satisfied"},
            {Name: "Violated", Src: []string{"Init"}, Dst: "Violated"},
            {Name: "Init", Src: []string{"Satisfied"}, Dst: "Init"},
            {Name: "Violated", Src: []string{"Diocane"}, Dst: "Violated"},
        },
        fsm.Callbacks{},
    )
    return finitestate

    }
    return &fsm.FSM{}
}
