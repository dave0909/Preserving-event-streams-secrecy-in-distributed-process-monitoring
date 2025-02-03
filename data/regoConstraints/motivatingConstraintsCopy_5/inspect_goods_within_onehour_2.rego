package inspect_goods_within_onehour_2
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

#temporary satisfied condition if the last event is Truck reached costumer (TRC)
InitToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Truck reached costumer (TRC)"
}
#temporary satisfied condition if the last event is "Inspect goods (IG)" and the difference with the older Truck reached costumer (TRC) is less than one hour
TemporaryViolatedToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Inspect goods (IG)"
    #Get the older Truck reached costumer (TRC) event
    reached_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Truck reached costumer (TRC)"]
    reached := min(reached_events) # This will be 0 if reached_events is empty
    #check if the fime difference is less than one hour
    time.parse_rfc3339_ns(most_recent_event.timestamp) - reached <= 3600000000000
}
#Violation condition if the last event is "Inspect goods (IG)" and the difference with the older Truck reached costumer (TRC) is more than one hour
TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Inspect goods (IG)"
    #Get the older Truck reached costumer (TRC) event
    reached_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Truck reached costumer (TRC)"]
    reached := min(reached_events) # This will be 0 if reached_events is empty
    #check if the fime difference is less than one hour
    time.parse_rfc3339_ns(most_recent_event.timestamp) - reached > 3600000000000
}
#Violation condition if the last event is "Inspect goods (IG)" and the difference with the older Truck reached costumer (TRC) is more than one hour
TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Inspect goods (IG)"
    #Get the older Truck reached costumer (TRC) event
    reached_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Truck reached costumer (TRC)"]
    reached := min(reached_events) # This will be 0 if reached_events is empty
    #check if the fime difference is less than one hour
    time.parse_rfc3339_ns(most_recent_event.timestamp) - reached > 3600000000000
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