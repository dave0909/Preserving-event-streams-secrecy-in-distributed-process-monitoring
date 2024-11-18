package lactic_acid_within_onehour
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

# Define a rule to check if the most recent event is "LacticAcid"
triage_present[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    count({e | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "ER Sepsis Triage"}) > 0
}

# Pending condition
pending[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "ER Sepsis Triage"
}

# Violation condition
violations[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "LacticAcid"
    not lactic_acid_within_one_hour[trace_id]
}

# Satisfied condition
satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "LacticAcid"
    lactic_acid_within_one_hour[trace_id]
}

# Define a rule to check if "LacticAcid" happens within one hour after the latest "ER Sepsis Triage"
lactic_acid_within_one_hour[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    triage_present[trace_id]
    sepsisTriage := max([time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "ER Sepsis Triage"])
    lactic_acid := most_recent_event
    time.parse_rfc3339_ns(lactic_acid.timestamp) <= sepsisTriage + 10800000000000
}