package balance_notcollect

import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

#Satisfied if the sum af the paymentAmount attribute of all the "Payment" activities is greater than 10
satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    sum([e.paymentAmount | e := input.events; e.concept_name == "Payment"]) > 10
}

#Temporary satisfied if the sum af the paymentAmount attribute of all the "Payment" activities is less or equal than 10
temporary_satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    sum([e.paymentAmount | e := input.events; e.concept_name == "Payment"]) <= 10
}

#Violated if the "Send for Credit Collection" activity exists in the trace
temporary_violated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    some e1; e1 = input.events[_]; e1.trace_concept_name == trace_id; e1.concept_name == "Send for Credit Collection"
}