package inspect_goods_within_onehour
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

## Define a rule to check if the most recent event is "IV Antibiotics"
#reached_present[trace_id] if {
#    trace_id := most_recent_event.trace_concept_name
#    count({e | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Truck reached costumer (TRC)"}) > 0
#}
#
## Pending condition
#pending[trace_id] if {
#    trace_id := most_recent_event.trace_concept_name
#    most_recent_event.concept_name == "Truck reached costumer (TRC)"
#}
#
## Violation condition
#violations[trace_id] if {
#    trace_id := most_recent_event.trace_concept_name
#    most_recent_event.concept_name == "Inspect goods (IG)"
#    not inspect_goods_within_one_hour[trace_id]
#}
#
## Violation condition, when the trace is over and the constraint is in pending state
#violations[trace_id] if {
#    trace_id := most_recent_event.trace_concept_name
#    most_recent_event.concept_name == "Order reception confirmed (ORC)"
#}
#
## Satisfied condition, when I receive an inspect goods activity and the constraint is in pending state
#satisfied[trace_id] if {
#    trace_id := most_recent_event.trace_concept_name
#    most_recent_event.concept_name == "Inspect goods (IG)"
#    inspect_goods_within_one_hour[trace_id]
#}
#
## Define a rule to check if "Inspect goods (IG)" happens within one hour after the latest "Truck reached costumer (TRC)"
#inspect_goods_within_one_hour[trace_id] if {
#    trace_id := most_recent_event.trace_concept_name
#    reached_present[trace_id]
#    last_inspect := max([time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Inspect goods (IG)"])
#    reached_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Truck reached costumer (TRC)";time.parse_rfc3339_ns(e.timestamp) < last_inspect]
#    reached := min(reached_events) # This will be 0 if reached_events is empty
#    inspect := most_recent_event
#    time.parse_rfc3339_ns(inspect.timestamp) <= reached + 3600000000000
#}
#temporary satisfied condition if the last event is Truck reached costumer (TRC)
temporary_violated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Truck reached costumer (TRC)"
}
#temporary satisfied condition if the last event is "Inspect goods (IG)" and the difference with the older Truck reached costumer (TRC) is less than one hour
temporary_satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Inspect goods (IG)"
    #Get the older Truck reached costumer (TRC) event
    reached_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Truck reached costumer (TRC)"]
    reached := min(reached_events) # This will be 0 if reached_events is empty
    #check if the fime difference is less than one hour
    time.parse_rfc3339_ns(most_recent_event.timestamp) - reached <= 3600000000000
}
#Violation condition if the last event is "Inspect goods (IG)" and the difference with the older Truck reached costumer (TRC) is more than one hour
violations[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Inspect goods (IG)"
    #Get the older Truck reached costumer (TRC) event
    reached_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Truck reached costumer (TRC)"]
    reached := min(reached_events) # This will be 0 if reached_events is empty
    #check if the fime difference is less than one hour
    time.parse_rfc3339_ns(most_recent_event.timestamp) - reached > 3600000000000
}


