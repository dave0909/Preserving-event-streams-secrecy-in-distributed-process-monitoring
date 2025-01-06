package R24

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
}