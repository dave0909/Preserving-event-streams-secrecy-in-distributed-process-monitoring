package separation_of_duty_4
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

# Pending condition 1
InitToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Fill in container (FC)"
}
# Pending condition 1
InitToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Fill in container (FC)"
}

# Violation condition 1
TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Fill in container (FC)"
    same_operator_exists(trace_id, "Fill in container (FC)", "Check container (CC)")
}

# Violation condition 2
TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Check container (CC)"
    same_operator_exists(trace_id, "Check container (CC)", "Fill in container (FC)")
}

# Satisfied condition, when the trace is over and the constraint is in pending state
TemporarySatisfiedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

# Define a rule to check if the same logistics operator exists for both activities in the same trace
same_operator_exists(trace_id, activity1, activity2) if {
    e1 := input.events[_]
    e1.trace_concept_name == trace_id
    e1.concept_name == activity1
    e2 := input.events[_]
    e2.trace_concept_name == trace_id
    e2.concept_name == activity2
    e1.logistics_operator == e2.logistics_operator
}