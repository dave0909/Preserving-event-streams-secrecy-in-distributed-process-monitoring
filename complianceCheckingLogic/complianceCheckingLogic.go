package complianceCheckingLogic


import (
    "context"
    "fmt"
    "github.com/open-policy-agent/opa/rego"
    "sync"
    "log"
    "github.com/looplab/fsm"
)

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
    most_recent_event.concept_name != "Send for Credit Collection"
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
                        fmt.Println("Constraint ",constraintName,"in pending state for case ",caseId)

                    }
                }
                for caseId := range temporarySatisfied {
                    if constraint.ConstraintState[caseId].Can("TemporarySatisfied") {
                        constraint.ConstraintState[caseId].Event(ccl.ctx, "TemporarySatisfied")
                        fmt.Println("Constraint ",constraintName,"in temporary satisfied state for case ",caseId)

                    }
                }
                for caseId := range temporaryViolated {
                    if constraint.ConstraintState[caseId].Can("TemporaryViolated") {
                        constraint.ConstraintState[caseId].Event(ccl.ctx, "TemporaryViolated")
                        fmt.Println("Constraint ",constraintName,"in temporary violated state for case ",caseId)

                    }
                }
                for caseId := range violations {
                    if constraint.ConstraintState[caseId].Can("Violated") {
                        constraint.ConstraintState[caseId].Event(ccl.ctx, "Violated")
                        fmt.Println("Constraint ",constraintName,"in violated state for case ",caseId)

                    }
                }
                for caseId := range satisfied {
                    if constraint.ConstraintState[caseId].Can("Satisfied") {
                        constraint.ConstraintState[caseId].Event(ccl.ctx, "Satisfied")
                        fmt.Println("Constraint ",constraintName,"in satisfied state for case ",caseId)

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
case "appeal_notcollect":
    finitestate := fsm.NewFSM(
        "Init",
        fsm.Events{
            {Name: "TemporarySatisfied", Src: []string{"Init"}, Dst: "TemporarySatisfied"},
            {Name: "Violated", Src: []string{"TemporarySatisfied"}, Dst: "Violated"},
        },
        fsm.Callbacks{},
    )
    return finitestate

case "balance_notcollect":
    finitestate := fsm.NewFSM(
        "Init",
        fsm.Events{
            {Name: "TemporarySatisfied", Src: []string{"Init"}, Dst: "TemporarySatisfied"},
            {Name: "Satisfied", Src: []string{"Init"}, Dst: "Satisfied"},
            {Name: "Satisfied", Src: []string{"TemporaryViolated"}, Dst: "Satisfied"},
            {Name: "TemporaryViolated", Src: []string{"TemporarySatisfied"}, Dst: "TemporaryViolated"},
            {Name: "Satisfied", Src: []string{"TemporarySatisfied"}, Dst: "Satisfied"},
        },
        fsm.Callbacks{},
    )
    return finitestate

case "case_start":
    finitestate := fsm.NewFSM(
        "Init",
        fsm.Events{
            {Name: "TemporarySatisfied", Src: []string{"Init"}, Dst: "TemporarySatisfied"},
            {Name: "Violated", Src: []string{"TemporarySatisfied"}, Dst: "Violated"},
            {Name: "Satisfied", Src: []string{"Init"}, Dst: "Satisfied"},
        },
        fsm.Callbacks{},
    )
    return finitestate

case "delay_no_collect":
    finitestate := fsm.NewFSM(
        "Init",
        fsm.Events{
            {Name: "TemporaryViolated", Src: []string{"Init"}, Dst: "TemporaryViolated"},
            {Name: "TemporarySatisfied", Src: []string{"Init"}, Dst: "TemporarySatisfied"},
            {Name: "Satisfied", Src: []string{"Init"}, Dst: "Satisfied"},
            {Name: "TemporarySatisfied", Src: []string{"TemporaryViolated"}, Dst: "TemporarySatisfied"},
            {Name: "Satisfied", Src: []string{"TemporaryViolated"}, Dst: "Satisfied"},
            {Name: "Violated", Src: []string{"TemporarySatisfied"}, Dst: "Violated"},
        },
        fsm.Callbacks{},
    )
    return finitestate

case "dismissal_notinsert":
    finitestate := fsm.NewFSM(
        "Init",
        fsm.Events{
            {Name: "TemporarySatisfied", Src: []string{"Init"}, Dst: "TemporarySatisfied"},
            {Name: "Violated", Src: []string{"TemporarySatisfied"}, Dst: "Violated"},
        },
        fsm.Callbacks{},
    )
    return finitestate

case "insert_and_notify_notcollect":
    finitestate := fsm.NewFSM(
        "Init",
        fsm.Events{
            {Name: "TemporarySatisfied", Src: []string{"Init"}, Dst: "TemporarySatisfied"},
            {Name: "Satisfied", Src: []string{"Init"}, Dst: "Satisfied"},
            {Name: "Satisfied", Src: []string{"TemporarySatisfied"}, Dst: "Satisfied"},
            {Name: "TemporaryViolated", Src: []string{"TemporarySatisfied"}, Dst: "TemporaryViolated"},
            {Name: "Satisfied", Src: []string{"TemporaryViolated"}, Dst: "Satisfied"},
        },
        fsm.Callbacks{},
    )
    return finitestate

case "no_insert_no_collect":
    finitestate := fsm.NewFSM(
        "Init",
        fsm.Events{
            {Name: "Satisfied", Src: []string{"Init"}, Dst: "Satisfied"},
            {Name: "TemporaryViolated", Src: []string{"Init"}, Dst: "TemporaryViolated"},
            {Name: "TemporarySatisfied", Src: []string{"Init"}, Dst: "TemporarySatisfied"},
            {Name: "Satisfied", Src: []string{"TemporarySatisfied"}, Dst: "Satisfied"},
            {Name: "TemporaryViolated", Src: []string{"TemporarySatisfied"}, Dst: "TemporaryViolated"},
            {Name: "Satisfied", Src: []string{"TemporaryViolated"}, Dst: "Satisfied"},
        },
        fsm.Callbacks{},
    )
    return finitestate

case "two_create_send_notinsertfine":
    finitestate := fsm.NewFSM(
        "Init",
        fsm.Events{
            {Name: "TemporarySatisfied", Src: []string{"Init"}, Dst: "TemporarySatisfied"},
            {Name: "Violated", Src: []string{"TemporarySatisfied"}, Dst: "Violated"},
            {Name: "Satisfied", Src: []string{"TemporarySatisfied"}, Dst: "Satisfied"},
        },
        fsm.Callbacks{},
    )
    return finitestate

    }
    return &fsm.FSM{}
}
