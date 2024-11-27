package appeal_notcollect

import rego.v1


# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]


#Temporary satisfied if the last event is the last event is "Appeal to Judge"
temporary_satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Appeal to Judge"
}

#Temporary satisfied if the last event is or "Send Appeal to Prefecture"
temporary_satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Send Appeal to Prefecture"
}

#Violated if "Send for Credit Collection" activity exists in the trace
violated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    some e1; e1 = input.events[_]; e1.trace_concept_name == trace_id; e1.concept_name == "Send for Credit Collection"
}
