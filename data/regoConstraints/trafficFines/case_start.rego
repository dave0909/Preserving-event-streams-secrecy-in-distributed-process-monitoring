package case_start

import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]
#Init a Timestamp variable of the first event in the log is 2000-01-01 00:00:00+00:00
first_event_timestamp := time.parse_rfc3339_ns("2000-01-01T00:00:00Z")

temporary_satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name != "Send for Credit Collection"
    #The timestamp of the event is after 4481 days from the first event
    (time.parse_rfc3339_ns(most_recent_event.timestamp) / 1000000000) - (first_event_timestamp / 1000000000) >= 4481 * 24 * 60 * 60
}

#Check among the pasts events if there is a "Send for Credit Collection" event
violated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    some e1; e1 = input.events[_]; e1.trace_concept_name == trace_id; e1.concept_name == "Send for Credit Collection"
}

satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    #The timestamp of the event is after 4481 days from the first event
    (time.parse_rfc3339_ns(most_recent_event.timestamp) / 1000000000) - (first_event_timestamp / 1000000000) < 4481 * 24 * 60 * 60
}
