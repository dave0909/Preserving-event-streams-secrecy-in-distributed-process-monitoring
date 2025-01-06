package R18

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
}