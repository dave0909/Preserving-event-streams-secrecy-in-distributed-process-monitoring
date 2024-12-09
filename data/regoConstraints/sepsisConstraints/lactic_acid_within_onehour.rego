package lactic_acid_within_onehour
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]


temporary_violated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "ER Sepsis Triage"
}
#temporary satisfied condition if the last event is "LacticAcid" and the difference with the older ER Sepsis Triage is less than one hour
temporary_satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "LacticAcid"
    #Get the older Truck reached costumer (TRC) event
    reached_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "ER Sepsis Triage"]
    reached := min(reached_events) # This will be 0 if reached_events is empty
    #check if the fime difference is less than one hour
    time.parse_rfc3339_ns(most_recent_event.timestamp) - reached <= 10800000000000
}
#Violation condition if the last event is "Inspect goods (IG)" and the difference with the older Truck reached costumer (TRC) is more than one hour
violations[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "LacticAcid"
    #Get the older Truck reached costumer (TRC) event
    reached_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "ER Sepsis Triage"]
    reached := min(reached_events) # This will be 0 if reached_events is empty
    #check if the fime difference is less than one hour
    time.parse_rfc3339_ns(most_recent_event.timestamp) - reached > 10800000000000
}
