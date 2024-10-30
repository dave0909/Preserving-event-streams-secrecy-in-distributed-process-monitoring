package truck_policy

import rego.v1

default five_years_in_seconds := 157766400

default driver_experience_violation := false

driver_experience_violation if {
	trace_name := input.events[_].trace_concept_name
	violations[trace_name]
}

violations[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	some event in input.events
	event.concept_name == "Select truck (ST)"
	(time.parse_rfc3339_ns(event.timestamp) / 1000000000) - (time.parse_rfc3339_ns(event.license_first_issue) / 1000000000) < five_years_in_seconds
	event.trace_concept_name == trace_id
}
