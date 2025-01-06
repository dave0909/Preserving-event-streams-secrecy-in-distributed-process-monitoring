package R17
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
