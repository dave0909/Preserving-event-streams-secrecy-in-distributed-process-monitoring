package separation_of_duty
import rego.v1

# Define a rule to check if the logistics operators for "Fill in container" and "Check container" activities are different for each trace
check_operators_condition[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	e1 := input.events[_]
	e1.trace_concept_name == trace_id
	e1.concept_name == "Fill in container (FC)"
	e2 := input.events[_]
	e2.trace_concept_name == trace_id
	e2.concept_name == "Check container (CC)"
	e1.logistics_operator != e2.logistics_operator
}

# Define a rule to get all trace IDs that do not satisfy the operators condition
violations[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	not check_operators_condition[trace_id]
	# Ensure that both activities are present
	count({e | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Fill in container (FC)"}) > 0
	count({e | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Check container (CC)"}) > 0
}

