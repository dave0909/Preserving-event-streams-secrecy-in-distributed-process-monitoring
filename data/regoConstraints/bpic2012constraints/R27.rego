package R27
import rego.v1

#Description: “A_PARTLYSUBMITTED-complete” occurs at most 22 s
              #after “A_SUBMITTED-complete”;

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

#InitToSatisfied if the last event is __END__
InitToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

#InitToTemporarySatisfied if the last event is A_SUBMITTED-complete
InitToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_SUBMITTED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
}

#TemporarySatisfiedToViolated if the last event is A_PARTLYSUBMITTED-complete and the difference with the older A_SUBMITTED-complete is more than 22 seconds
TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_PARTLYSUBMITTED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Get the older A_SUBMITTED-complete event
    submitted_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "A_SUBMITTED";e["lifecycle:transition"] == "COMPLETE"]
    submitted := min(submitted_events) # This will be 0 if submitted_events is empty
    #check if the time difference is more than 22 seconds
    time.parse_rfc3339_ns(most_recent_event.timestamp) - submitted > 22000000000
}

TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
    #Check if exists an A_PARTLYSUBMITTED-complete event for the same trace
    checkPartlySubmitted(trace_id)
}

#TemporarySatisfiedToSatisfied  if the last event is __END__
TemporarySatisfiedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
    #Check if exists an A_PARTLYSUBMITTED-complete event for the same trace
    checkPartlySubmitted(trace_id)
}

#TemporaryViolatedToTemporarySatisfied if the last event is A_PARTLYSUBMITTED-complete and the difference with the older A_SUBMITTED-complete is less than 22 seconds
TemporaryViolatedToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_PARTLYSUBMITTED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Get the older A_SUBMITTED-complete event
    submitted_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "A_SUBMITTED";e["lifecycle:transition"] == "COMPLETE"]
    submitted := min(submitted_events) # This will be 0 if submitted_events is empty
    #check if the time difference is less than 22 seconds
    time.parse_rfc3339_ns(most_recent_event.timestamp) - submitted <= 22000000000
    }
TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "A_PARTLYSUBMITTED"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Get the older A_SUBMITTED-complete event
    submitted_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "A_SUBMITTED";e["lifecycle:transition"] == "COMPLETE"]
    submitted := min(submitted_events) # This will be 0 if submitted_events is empty
    #check if the time difference is more than 22 seconds
    time.parse_rfc3339_ns(most_recent_event.timestamp) - submitted > 22000000000
}

checkPartlySubmitted(trace_id) if {
    input.events[_].concept_name == "A_PARTLYSUBMITTED"
    input.events[_]["lifecycle:transition"] == "COMPLETE"
    input.events[_].trace_concept_name == trace_id
}
