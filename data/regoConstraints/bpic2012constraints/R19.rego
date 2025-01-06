package R19
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
