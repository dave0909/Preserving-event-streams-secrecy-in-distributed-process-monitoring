package inspect_goods_within_onehour

#CONSTRAINT: Inspect goods must happen within one hour from the "Truck reached costumer activity"
import rego.v1

# Define a rule to check if "Inspect goods" happens within one hour after "Truck reached customer" for each trace
inspect_goods_within_one_hour[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	truck_reached := input.events[_];truck_reached.trace_concept_name == trace_id;truck_reached.concept_name == "Truck reached costumer (TRC)"
	inspect_goods := input.events[_];inspect_goods.trace_concept_name == trace_id;inspect_goods.concept_name == "Inspect goods (IG)"
	time.parse_rfc3339_ns(inspect_goods.inspection_time) <= time.parse_rfc3339_ns(truck_reached.receipt_time) + 3600000000000
}

# Define a rule to check if "Inspect goods" activity is present in the trace
inspect_goods_present[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	count({e | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Inspect goods (IG)"}) > 0
}

# Define a rule to get all trace IDs that do not satisfy the condition
violations[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	inspect_goods_present[trace_id]
	not inspect_goods_within_one_hour[trace_id]
	}


