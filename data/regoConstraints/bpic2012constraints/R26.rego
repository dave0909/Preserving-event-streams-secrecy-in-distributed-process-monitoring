package R26
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
