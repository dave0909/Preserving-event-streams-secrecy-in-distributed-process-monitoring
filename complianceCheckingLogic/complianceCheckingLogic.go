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
"R17", "R18", "R19", "R20", "R21", "R22", "R23", "R24", "R25", "R26", "R27", "R28"}

var constraints = []string{

`package R17
import rego.v1
#Description: "A_PARTLYSUBMITTED-complete" occouurs exactly once in the trace

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

# InitToTemporaryViolated if the last event is "A_PARTLYSUBMITTED"
InitToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name != "A_PARTLYSUBMITTED"
}

#InitToTemporarySatisfied if the last event is "A_PARTLYSUBMITTED"
InitToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_PARTLYSUBMITTED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
}

TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_PARTLYSUBMITTED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
}

TemporarySatisfiedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

TemporaryViolatedToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_PARTLYSUBMITTED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
}
`,

`package R18

import rego.v1

#Description: "A_SUBMITTED"-COMPLETE occurs exactly once in the traces

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

# InitToTemporaryViolated if the last event is "A_SUBMITTED"
InitToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name != "A_SUBMITTED"
}

#InitToTemporarySatisfied if the last event is "A_SUBMITTED"
InitToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_SUBMITTED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
}

TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_SUBMITTED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
}

TemporarySatisfiedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

TemporaryViolatedToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_SUBMITTED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
}`,

`package R19
import rego.v1
#Description: "A_SUBMITTED"-COMPLETE is always immediately followed by "A_PARTLYSUBMITTED" and vice versa

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

# InitToTemporaryViolated if the last event is "A_SUBMITTED"
InitToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_SUBMITTED"
    #Life cycle transition is "COMPLETE"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
}
# InitToViolated if the last event is "A_PARTLYSUBMITTED"
InitToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_PARTLYSUBMITTED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
}

#InitToTemporarySatisfied if the last event is anything other than "A_SUBMITTED" or "A_PARTLYSUBMITTED"
InitToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name != "A_SUBMITTED"
    most_recent_event.concept_name != "A_PARTLYSUBMITTED"
}

#TemporaryViolatedToViolated if the last event is any event other than "A_PARTLYSUBMITTED"
TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name != "A_PARTLYSUBMITTED"
}

#TemporaryViolatedToViolated is the last event is __END__
TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}
#TemporaryViolatedToTemporarySatisfied if the last event is "A_PARTLYSUBMITTED" and lifecycle transition is "COMPLETE"
TemporaryViolatedToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_PARTLYSUBMITTED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
}

#TemporarySatisfiedToViolated if the last event is __END__
TemporarySatifiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}
#TemporarySatisfiedToTemporaryViolated if the last event is "A_SUBMITTED" and lifecycle transition is "COMPLETE"
TemporarySatifiedToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_SUBMITTED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"}
`,

`package R20
import rego.v1
#Description: "A_SUBMITTED"-COMPLETE and "A_PARTLYSUBMITTED"-COMPLETE are always the firts two events of the trace

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

#InitToViolated if the last event is not "A_SUBMITTED"
InitToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name != "A_SUBMITTED"
}

InitToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    #Life cycle transition is not "COMPLETE"
    most_recent_event["lifecycle:transition"] != "COMPLETE"
}

#InitToTemporaryViolated if the last event is "A_SUBMITTED"
InitToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_SUBMITTED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    }

#TemporaryViolatedToViolated if the last event is not "A_PARTLYSUBMITTED"
TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name != "A_PARTLYSUBMITTED"
}

TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event["lifecycle:transition"] != "COMPLETE"
}

TemporaryViolatedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_PARTLYSUBMITTED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
}
`,

`package R21

import rego.v1

#Description: Lifecycle : W_Afhandelen leads - SCHEDULE --> W_Afhandelen leads - START --> W_Afhandelen leads - COMPLETE is always respected

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

InitToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Afhandelen leads"
    most_recent_event["lifecycle:transition"] == "SCHEDULE"
}

InitToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Afhandelen leads"
    most_recent_event["lifecycle:transition"] != "SCHEDULE"
}

InitToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

TemporaryViolatedToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Afhandelen leads"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Check if W_Afhandelen leads - START is present
    checkStart(trace_id)
}

TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Afhandelen leads"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Check if W_Afhandelen leads - START is present
    not checkStart(trace_id)
}
TemporarySatisfiedToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Afhandelen leads"
    most_recent_event["lifecycle:transition"] == "SCHEDULE"
}

TemporarySatisfiedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

#Function that checks if W_Afhandelen leads - SCHEDULE is present for the same trace
checkStart(trace_id) if {
    input.events[_].concept_name == "W_Afhandelen leads"
    input.events[_]["lifecycle:transition"] == "START"
    input.events[_].trace_concept_name == trace_id
}
`,

`package R22
import rego.v1


#Description: W_Completeren aanvraag - SCHEDULE is always followed eventually by W_Completeren aanvraag - COMPLETE and vice versa

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

InitToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Completeren aanvraag"
    most_recent_event["lifecycle:transition"] == "SCHEDULE"
}

InitToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Completeren aanvraag"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
}

InitToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

TemporaryViolatedToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Completeren aanvraag"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
}

TemporarySatisfiedToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Completeren aanvraag"
    most_recent_event["lifecycle:transition"] == "SCHEDULE"
}

TemporarySatisfiedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}`,

`package R23

import rego.v1

#Description: The lifecycle “W_Beoordelen fraude-schedule”,
              #“W_Beoordelen fraude-start”, “W_Beoordelen
              #fraude-complete” is always respected;

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

InitToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Beoordelen fraude"
    most_recent_event["lifecycle:transition"] == "SCHEDULE"
}

InitToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Beoordelen fraude"
    most_recent_event["lifecycle:transition"] != "SCHEDULE"
}

InitToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

TemporaryViolatedToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Beoordelen fraude"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Check if W_Beoordelen fraude - START is present
    checkStart(trace_id)
}

TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Beoordelen fraude"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Check if W_Beoordelen fraude - START is present
    not checkStart(trace_id)
}
TemporarySatisfiedToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Beoordelen fraude"
    most_recent_event["lifecycle:transition"] == "SCHEDULE"
}

TemporarySatisfiedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

#Function that checks if W_Beoordelen fraude - SCHEDULE is present for the same trace
checkStart(trace_id) if {
    input.events[_].concept_name == "W_Beoordelen fraude"
    input.events[_]["lifecycle:transition"] == "START"
    input.events[_].trace_concept_name == trace_id
}`,

`package R24

import rego.v1

#Description: “O_SELECTED-complete” is always followed eventually
              #by “O_CREATED-complete”, “O_CREATED-complete” is
              #always followed eventually by “O_SENT-complete”
              #and, vice versa, “O_SENT-complete” is always preceded
              #by “O_CREATED-complete” and “O_CREATED-com-plete”
              #is always preceded by “O_SELECTED-complete”

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

#InitToSatisfied if the last event is __END__
InitToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

InitToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "O_SELECTED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
}

#TemporaryViolatedToViolated if the last event is __END__
TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

#TemporaryViolatedToTemporarySatisfied if O_SENT-complete is the last event and O_CREATED-complete is present for the given trace
TemporaryViolatedToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "O_SENT"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    checkCreated(trace_id)
}

TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "O_SENT"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    not checkCreated(trace_id)
}

checkCreated(trace_id) if {
    input.events[_].concept_name == "O_CREATED"
    input.events[_]["lifecycle:transition"] == "COMPLETE"
    input.events[_].trace_concept_name == trace_id
}

TemporarySatisfiedToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "O_SELECTED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
}

TemporarySatisfiedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}`,

`package R25

import rego.v1

#Description: “A_CANCELLED-complete” does not coexist neither
              #with “A_ACTIVATED-complete” nor with “A_REGIS-
              #TERED-complete” nor with “A_APPROVED-complete”


# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

#Define the set of the events that cannot coexist with A_CANCELLED-complete
cannot_coexist := {"A_ACTIVATED", "A_REGISTERED", "A_APPROVED"}

InitToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

InitToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_CANCELLED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
}

InitToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_CANCELLED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
}

InitToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    #Check if the event name is in the set of the events that cannot coexist with A_CANCELLED-complete
    most_recent_event.concept_name == cannot_coexist[_]
    most_recent_event["lifecycle:transition"] == "COMPLETE"
}

TemporarySatisfiedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_CANCELLED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Check if one of the events that cannot coexist with A_CANCELLED-complete is present for the same trace
    checkCannotCoexist(trace_id)
}

TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == cannot_coexist[_]
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Check if A_CANCELLED-complete is present for the same trace
    checkCancelled(trace_id)
}
checkCannotCoexist(trace_id) if {
    input.events[_].concept_name == cannot_coexist[_]
    input.events[_]["lifecycle:transition"] == "COMPLETE"
    input.events[_].trace_concept_name == trace_id
}
checkCancelled(trace_id) if {
    input.events[_].concept_name == "A_CANCELLED"
    input.events[_]["lifecycle:transition"] == "COMPLETE"
    input.events[_].trace_concept_name == trace_id
}



`,

`package R26
import rego.v1

#Description: W_Beoordelen fraude-schedule” does not coexist
              #with “W_Wijzigen contractgegevens-schedule”

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

InitToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

InitToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Beoordelen fraude"
    most_recent_event["lifecycle:transition"] == "SCHEDULE"
    }
InitToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Wijzigen contractgegevens"
    most_recent_event["lifecycle:transition"] == "SCHEDULE"
}
TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Beoordelen fraude"
    most_recent_event["lifecycle:transition"] == "SCHEDULE"
    checkWijzigen(trace_id)
    }
TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Wijzigen contractgegevens"
    most_recent_event["lifecycle:transition"] == "SCHEDULE"
    checkFraude(trace_id)

}
TemporarySatisfiedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}
checkWijzigen(trace_id) if {
    input.events[_].concept_name == "W_Wijzigen contractgegevens"
    input.events[_]["lifecycle:transition"] == "SCHEDULE"
    input.events[_].trace_concept_name == trace_id
}
checkFraude(trace_id) if {
    input.events[_].concept_name == "W_Beoordelen fraude"
    input.events[_]["lifecycle:transition"] == "SCHEDULE"
    input.events[_].trace_concept_name == trace_id
}
`,

`package R27
import rego.v1

#Description: “A_PARTLYSUBMITTED-complete” occurs at most 22 s
              #after “A_SUBMITTED-complete”;

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

#InitToSatisfied if the last event is __END__
InitToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

#InitToTemporarySatisfied if the last event is A_SUBMITTED-complete
InitToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_SUBMITTED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
}

#TemporarySatisfiedToViolated if the last event is A_PARTLYSUBMITTED-complete and the difference with the older A_SUBMITTED-complete is more than 22 seconds
TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_PARTLYSUBMITTED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Get the older A_SUBMITTED-complete event
    submitted_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "A_SUBMITTED";e["lifecycle:transition"] == "COMPLETE"]
    submitted := min(submitted_events) # This will be 0 if submitted_events is empty
    #check if the time difference is more than 22 seconds
    time.parse_rfc3339_ns(most_recent_event.timestamp) - submitted > 22000000000
}

TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
    #Check if exists an A_PARTLYSUBMITTED-complete event for the same trace
    checkPartlySubmitted(trace_id)
}

#TemporarySatisfiedToSatisfied  if the last event is __END__
TemporarySatisfiedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
    #Check if exists an A_PARTLYSUBMITTED-complete event for the same trace
    checkPartlySubmitted(trace_id)
}

#TemporaryViolatedToTemporarySatisfied if the last event is A_PARTLYSUBMITTED-complete and the difference with the older A_SUBMITTED-complete is less than 22 seconds
TemporaryViolatedToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_PARTLYSUBMITTED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Get the older A_SUBMITTED-complete event
    submitted_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "A_SUBMITTED";e["lifecycle:transition"] == "COMPLETE"]
    submitted := min(submitted_events) # This will be 0 if submitted_events is empty
    #check if the time difference is less than 22 seconds
    time.parse_rfc3339_ns(most_recent_event.timestamp) - submitted <= 22000000000
    }
TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_PARTLYSUBMITTED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Get the older A_SUBMITTED-complete event
    submitted_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "A_SUBMITTED";e["lifecycle:transition"] == "COMPLETE"]
    submitted := min(submitted_events) # This will be 0 if submitted_events is empty
    #check if the time difference is more than 22 seconds
    time.parse_rfc3339_ns(most_recent_event.timestamp) - submitted > 22000000000
}

checkPartlySubmitted(trace_id) if {
    input.events[_].concept_name == "A_PARTLYSUBMITTED"
    input.events[_]["lifecycle:transition"] == "COMPLETE"
    input.events[_].trace_concept_name == trace_id
}
`,

`package R28
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]
#Description: “W_Completeren aanvraag-complete” occurs at least
              #22 s and at most 2 days, 18 h, 29 min, and 28 s after
              #“W_Completeren aanvraag-schedule”

#InitToSatisfied if the last event is __END__
InitToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

#InitToTemporaryViolated if the last event is W_Completeren aanvraag-schedule
InitToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Completeren aanvraag"
    most_recent_event["lifecycle:transition"] == "SCHEDULE"
}

#TemporaryViolatedToViolated if the last event is W_Completeren aanvraag-complete and the difference with the older W_Completeren aanvraag-schedule is more than 2 days, 18 h, 29 min, and 28 s
TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Completeren aanvraag"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Get the older W_Completeren aanvraag-schedule event
    schedule_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "W_Completeren aanvraag";e["lifecycle:transition"] == "SCHEDULE"]
    schedule := min(schedule_events) # This will be 0 if schedule_events is empty
    #check if the time difference is more than 2 days, 18 h, 29 min, and 28 s
    time.parse_rfc3339_ns(most_recent_event.timestamp) - schedule > 239368000000000
    #check if the time difference is less than 22 s
    #time.parse_rfc3339_ns(most_recent_event.timestamp) - schedule < 22000000000
}
#TemporaryViolatedToViolated if the last event is W_Completeren aanvraag-complete and the difference with the older W_Completeren aanvraag-schedule is more than 2 days, 18 h, 29 min, and 28 s
TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Completeren aanvraag"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Get the older W_Completeren aanvraag-schedule event
    schedule_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "W_Completeren aanvraag";e["lifecycle:transition"] == "SCHEDULE"]
    schedule := min(schedule_events) # This will be 0 if schedule_events is empty
    #check if the time difference is more than 2 days, 18 h, 29 min, and 28 s
    #time.parse_rfc3339_ns(most_recent_event.timestamp) - schedule > 239368000000000
    #check if the time difference is less than 22 s
    time.parse_rfc3339_ns(most_recent_event.timestamp) - schedule < 22000000000
}
#TemporaryViolatedToViolated if the last event is W_Completeren aanvraag-complete and the difference with the older W_Completeren aanvraag-schedule is more than 2 days, 18 h, 29 min, and 28 s
TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

#TemporaryViolatedToSatisfied if the last event is W_Completeren aanvraag-complete and the difference with the older W_Completeren aanvraag-schedule is less than 2 days, 18 h, 29 min, and 28 s and more than 22 s
TemporaryViolatedToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Completeren aanvraag"
    #Get the older W_Completeren aanvraag-schedule event
    schedule_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "W_Completeren aanvraag";e["lifecycle:transition"] == "SCHEDULE"]
    schedule := min(schedule_events) # This will be 0 if schedule_events is empty
    #check if the time difference is more than 22 s
    time.parse_rfc3339_ns(most_recent_event.timestamp) - schedule >= 22000000000
    #check if the time difference is less than 2 days, 18 h, 29 min, and 28 s
    time.parse_rfc3339_ns(most_recent_event.timestamp) - schedule <= 239368000000000
}

TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Completeren aanvraag"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Get the older W_Completeren aanvraag-schedule event
    schedule_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "W_Completeren aanvraag";e["lifecycle:transition"] == "SCHEDULE"]
    schedule := min(schedule_events) # This will be 0 if schedule_events is empty
    #check if the time difference is more than 2 days, 18 h, 29 min, and 28 s
    time.parse_rfc3339_ns(most_recent_event.timestamp) - schedule > 239368000000000
    #check if the time difference is less than 2 days, 18 h, 29 min, and 28 s
    #time.parse_rfc3339_ns(most_recent_event.timestamp) - schedule < 22000000000
}
TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Completeren aanvraag"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Get the older W_Completeren aanvraag-schedule event
    schedule_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "W_Completeren aanvraag";e["lifecycle:transition"] == "SCHEDULE"]
    schedule := min(schedule_events) # This will be 0 if schedule_events is empty
    #check if the time difference is more than 2 days, 18 h, 29 min, and 28 s
    #time.parse_rfc3339_ns(most_recent_event.timestamp) - schedule > 239368000000000
    #check if the time difference is less than 2 days, 18 h, 29 min, and 28 s
    time.parse_rfc3339_ns(most_recent_event.timestamp) - schedule < 22000000000
}

#TemporarySatisfiedToSatisfied condition if the last event is __END__
TemporarySatisfiedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
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
			constraint.ConstraintState[traceId] = Init // Init state
		}
	}
	resultMap := map[string]interface{}{}
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
			//currentState := constraint.ConstraintState[traceId]
			//if constraint.name == "truck_policy" {
			//	fmt.Println(res)
			//}
			for {
				transitionFound := false
				currentState := constraint.ConstraintState[traceId]
				for _, nextState := range constraint.fsm.PossibleNextStates(int(currentState)) {
					currentState := constraint.ConstraintState[traceId]
					ruleName := fmt.Sprintf("%sTo%s", stateName(currentState), stateName(ConstraintState(nextState)))
					if resultValue, ok := res[0].Expressions[0].Value.(map[string]interface{})[ruleName]; ok {
						//if constraint.name == "shipment_cost" {
						//	fmt.Println("Constraint name: ", constraint.name, "in state ", constraint.ConstraintState, "next state: ", stateName(ConstraintState(nextState)), "rulename: ", ruleName, "resultValue: ", resultValue)
						//	fmt.Println("Result: ", res)
						//}
						if resultValueMap, ok := resultValue.(map[string]interface{}); ok {
							for caseId, isTrue := range resultValueMap {
								if isTrue.(bool) {
									constraint.ConstraintState[caseId] = ConstraintState(nextState)
									fmt.Printf("Constraint %s transitioned from %s to %s for case %s", constraint.name, stateName(currentState), stateName(ConstraintState(nextState)), caseId)
									fmt.Println()
									resultMap[traceId] = nextState
									transitionFound = false
									//TODO: set the above variable to true to enable the recursive inspection of the constraints
								}
							}
						}
					}
				}
				if !transitionFound {
					break
				}
			}

		}(constraint)
	}
	wg.Wait()
	return resultMap
}

func stateName(state ConstraintState) string {
	switch state {
	case Init:
		return "Init"
	case Pending:
		return "Pending"
	case Violated:
		return "Violated"
	case Satisfied:
		return "Satisfied"
	case TemporarySatisfied:
		return "TemporarySatisfied"
	case TemporaryViolated:
		return "TemporaryViolated"
	default:
		return "Unknown"
	}
}


var fsmMap = map[string]*CustomFSM{

"R17": {
    Transitions: [][]int{
        {4, 5},
        {},
        {},
        {},
        {3, 2},
        {4, 2},
    },
},

"R18": {
    Transitions: [][]int{
        {4, 5},
        {},
        {},
        {},
        {3, 2},
        {4, 2},
    },
},

"R19": {
    Transitions: [][]int{
        {2, 5, 4},
        {},
        {},
        {},
        {3, 5},
        {4, 2},
    },
},

"R20": {
    Transitions: [][]int{
        {5},
        {},
        {},
        {},
        {},
        {2, 3},
    },
},

"R21": {
    Transitions: [][]int{
        {5, 3},
        {},
        {},
        {},
        {5, 3},
        {2, 4, 2},
    },
},

"R22": {
    Transitions: [][]int{
        {5, 3, 2},
        {},
        {},
        {},
        {5, 3},
        {2, 4},
    },
},

"R23": {
    Transitions: [][]int{
        {5, 3, 2},
        {},
        {},
        {},
        {5, 3},
        {2, 4},
    },
},

"R24": {
    Transitions: [][]int{
        {5, 3},
        {},
        {},
        {},
        {5, 3},
        {2, 4, 2},
    },
},

"R25": {
    Transitions: [][]int{
        {4, 3},
        {},
        {},
        {},
        {3, 2},
        {},
    },
},

"R26": {
    Transitions: [][]int{
        {4, 3},
        {},
        {},
        {},
        {3, 2},
        {},
    },
},

"R27": {
    Transitions: [][]int{
        {5, 3},
        {},
        {},
        {},
        {3, 2},
        {2, 4},
    },
},

"R28": {
    Transitions: [][]int{
        {5, 3},
        {},
        {},
        {},
        {3, 2},
        {2, 4},
    },
},

}
