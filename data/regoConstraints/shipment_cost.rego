package shipment_cost
import rego.v1

check_cost_condition[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	reserve_cost := sum([e.cost | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Reserve shipment (RS)"])
	drive_distance_i := sum([e.km_distance | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Drive to costumer (DC)"])
	drive_distance_m := sum([e.km_distance | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Drive to manufacturer (DM)"])
	reserve_cost <= (drive_distance_i + drive_distance_m) * 3
}

# Check if the "Detach container" activity has been executed
detach_container_executed[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	some d
	input.events[d].trace_concept_name == trace_id
	input.events[d].concept_name == "Detach container (DCO)"
}

# Define a rule to get all trace IDs that do not satisfy the cost condition and have executed the "Detach container" activity
violations[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	detach_container_executed[trace_id]
	not check_cost_condition[trace_id]
}