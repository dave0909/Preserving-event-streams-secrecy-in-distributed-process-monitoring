package truck_policy
import rego.v1

# Define the constant for five years in seconds
five_years_in_seconds := 5 * 365 * 24 * 60 * 60

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

# Pending condition
pending[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Select truck (ST)"
}

# Violation condition
violations[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    #Not needed
    #most_recent_event.concept_name == "Select truck (ST)"
    driver_experience_within_five_years[trace_id]
}

# Satisfied condition
satisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    #Not needed
    #most_recent_event.concept_name == "Select truck (ST)"
    not driver_experience_within_five_years[trace_id]
}

# Define a rule to check if the driver's experience is within five years
driver_experience_within_five_years[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    event := most_recent_event
    (time.parse_rfc3339_ns(event.timestamp) / 1000000000) - (time.parse_rfc3339_ns(event.license_first_issue) / 1000000000) <= five_years_in_seconds
}