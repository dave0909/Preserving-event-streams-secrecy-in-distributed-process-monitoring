package two_create_send_notinsertfine
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

# Temporary satisfied if 1) Create Fine is in the past events 2) Send Fine is in the past events 3) There are only two events in the trace_id
temporary_satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    some e1; e1 = input.events[_]; e1.trace_concept_name == trace_id; e1.concept_name == "Create Fine"
    some e2; e2 = input.events[_]; e2.trace_concept_name == trace_id; e2.concept_name == "Send Fine"
    count([e | e := input.events; e.trace_concept_name == trace_id]) == 2
}

# Violated if the "Insert Fine" is the last event of the trace
violated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Insert Fine Notification"
}

# Satisfied if the last event is not "Insert Fine Notification"
satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name != "Insert Fine Notification"
}