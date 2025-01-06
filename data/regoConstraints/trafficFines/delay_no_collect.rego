package delay_no_collect
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

InitToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    some e1; e1 = input.events[_]; e1.trace_concept_name == trace_id; e1.concept_name == "Add penalty"
    some e2; e2 = input.events[_]; e2.trace_concept_name == trace_id; e2.concept_name == "Payment"
    (time.parse_rfc3339_ns(e2.timestamp) / 1000000000) - (time.parse_rfc3339_ns(e1.timestamp) / 1000000000) > 3 * 24 * 60 * 60
}
TemporaryViolatedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    some e1; e1 = input.events[_]; e1.trace_concept_name == trace_id; e1.concept_name == "Add penalty"
    some e2; e2 = input.events[_]; e2.trace_concept_name == trace_id; e2.concept_name == "Payment"
    (time.parse_rfc3339_ns(e2.timestamp) / 1000000000) - (time.parse_rfc3339_ns(e1.timestamp) / 1000000000) > 3 * 24 * 60 * 60
}

# Violated if the last event is "Send for Credit Collection"
TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Send for Credit Collection"
}

InitToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    some e3; e3 = input.events[_]; e3.trace_concept_name == trace_id; e3.concept_name == "Add penalty"
    some e4; e4 = input.events[_]; e4.trace_concept_name == trace_id; e4.concept_name == "Payment"
    (time.parse_rfc3339_ns(e4.timestamp) / 1000000000) - (time.parse_rfc3339_ns(e3.timestamp) / 1000000000) <= 3 * 24 * 60 * 60
}

TemporaryViolatedToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    some e3; e3 = input.events[_]; e3.trace_concept_name == trace_id; e3.concept_name == "Add penalty"
    some e4; e4 = input.events[_]; e4.trace_concept_name == trace_id; e4.concept_name == "Payment"
    (time.parse_rfc3339_ns(e4.timestamp) / 1000000000) - (time.parse_rfc3339_ns(e3.timestamp) / 1000000000) <= 3 * 24 * 60 * 60
}

InitToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name != "Send for Credit Collection"
}