package dismissal_notinsert

import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

#Define a set of strings that represent the events that are considered as "Dismissal"
dismissal_events := {"2", "3", "5", "A", "B", "E", "F", "I", "J", "K", "M", "N", "Q", "R", "T", "U", "V"}

#Temporary satisfied if the last event is the last event is "Create Fine" and the "dismissal" attribute is in the dismissal_events set
InitToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Create Fine"
    most_recent_event.dismissal in dismissal_events
}

#Violated if the "Insert Fine" activity exists in the trace
TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    some e1; e1 = input.events[_]; e1.trace_concept_name == trace_id; e1.concept_name == "Insert Fine Notification"
}

TemporarySatisfiedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name != "__END__"
}