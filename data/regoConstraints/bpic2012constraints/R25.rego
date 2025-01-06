package R25

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



