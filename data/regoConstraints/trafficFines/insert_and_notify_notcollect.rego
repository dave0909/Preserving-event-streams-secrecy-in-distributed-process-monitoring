package insert_and_notify_notcollect
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

InitToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name != "Insert Date Appeal to Prefecture"
}

TemporarySatisfiedToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    #Send for Credit Collection extists in the trace
    some e1; e1 = input.events[_]; e1.trace_concept_name == trace_id; e1.concept_name == "Send for Credit Collection"
}

InitToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Notify Result Appeal to Offender"
}
TemporaryViolatedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Notify Result Appeal to Offender"
}
TemporarySatisfiedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Notify Result Appeal to Offender"
}