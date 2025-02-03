package shipment_cost_5
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

#Temporary satisfied condition if the last event is "Reserve shipment (RS)" and the cost condition is satisfied
InitToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Shipment reservation sent (SRS)"
	check_cost_condition(trace_id)
}
#Temporary satisfied condition if the last event is "Reserve shipment (RS)" and the cost condition is satisfied
InitToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Shipment reservation sent (SRS)"
	not check_cost_condition(trace_id)
}

TemporaryViolatedToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Drive to costumer (DC)"
    check_cost_condition(trace_id)
}

TemporaryViolatedToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Drive to manufacturer (DM)"
    check_cost_condition(trace_id)
}

TemporarySatisfiedToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Shipment reservation sent (SRS)"
    not check_cost_condition(trace_id)
}

# Define a function to check the cost condition
check_cost_condition(trace_id) if {
    reserve_cost := max([e.cost | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Shipment reservation sent (SRS)"])
    drive_distance_i := sum([e.km_distance | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Drive to costumer (DC)"])
    drive_distance_m := sum([e.km_distance | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Drive to manufacturer (DM)"])
    reserve_cost <= (drive_distance_i + drive_distance_m) * 3
}
#Satisfied condition if the last event is "__END__"
TemporarySatisfiedToSatisfied[trace_id] if {
	trace_id := most_recent_event.trace_concept_name
	most_recent_event.concept_name == "__END__"
}
#Violated condition if the last event is "__END__"
TemporaryViolatedToViolated[trace_id] if {
	trace_id := most_recent_event.trace_concept_name
	most_recent_event.concept_name == "__END__"
}