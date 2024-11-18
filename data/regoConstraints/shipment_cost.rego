package shipment_cost
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

# Pending condition
pending[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Reserve shipment (RS)"
}

# Violation condition
violations[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Drive to manufacturer (DM)"
    not check_cost_condition[trace_id]
}
# Violation condition
violations[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Drive to costumer (DC)"
    not check_cost_condition[trace_id]
}

# Satisfied condition
satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
	most_recent_event.concept_name == "Order reception confirmed (ORC)"
    check_cost_condition[trace_id]
}

# Define a rule to check the cost condition
check_cost_condition[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    reserve_cost := sum([e.cost | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Reserve shipment (RS)"])
    drive_distance_i := sum([e.km_distance | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Drive to costumer (DC)"])
    drive_distance_m := sum([e.km_distance | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Drive to manufacturer (DM)"])
    reserve_cost <= (drive_distance_i + drive_distance_m) * 3
}