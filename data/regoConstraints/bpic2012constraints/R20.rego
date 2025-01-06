package R20
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
