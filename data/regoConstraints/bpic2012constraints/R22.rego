package R22
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
}