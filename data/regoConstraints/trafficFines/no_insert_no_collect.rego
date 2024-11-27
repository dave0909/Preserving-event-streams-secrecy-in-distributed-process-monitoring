package no_insert_no_collect
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Insert Fine Notification"
}

temporary_satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name != "Insert Fine Notification"
    most_recent_event.concept_name != "Send for Credit Collection"
}
temporary_violated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name != "Send for Credit Collection"
}


