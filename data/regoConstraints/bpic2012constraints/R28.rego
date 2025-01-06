package R28
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]
#Description: “W_Completeren aanvraag-complete” occurs at least
              #22 s and at most 2 days, 18 h, 29 min, and 28 s after
              #“W_Completeren aanvraag-schedule”

#InitToSatisfied if the last event is __END__
InitToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

#InitToTemporaryViolated if the last event is W_Completeren aanvraag-schedule
InitToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Completeren aanvraag"
    most_recent_event["lifecycle:transition"] == "SCHEDULE"
}

#TemporaryViolatedToViolated if the last event is W_Completeren aanvraag-complete and the difference with the older W_Completeren aanvraag-schedule is more than 2 days, 18 h, 29 min, and 28 s
TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Completeren aanvraag"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Get the older W_Completeren aanvraag-schedule event
    schedule_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "W_Completeren aanvraag";e["lifecycle:transition"] == "SCHEDULE"]
    schedule := min(schedule_events) # This will be 0 if schedule_events is empty
    #check if the time difference is more than 2 days, 18 h, 29 min, and 28 s
    time.parse_rfc3339_ns(most_recent_event.timestamp) - schedule > 239368000000000
    #check if the time difference is less than 22 s
    #time.parse_rfc3339_ns(most_recent_event.timestamp) - schedule < 22000000000
}
#TemporaryViolatedToViolated if the last event is W_Completeren aanvraag-complete and the difference with the older W_Completeren aanvraag-schedule is more than 2 days, 18 h, 29 min, and 28 s
TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Completeren aanvraag"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Get the older W_Completeren aanvraag-schedule event
    schedule_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "W_Completeren aanvraag";e["lifecycle:transition"] == "SCHEDULE"]
    schedule := min(schedule_events) # This will be 0 if schedule_events is empty
    #check if the time difference is more than 2 days, 18 h, 29 min, and 28 s
    #time.parse_rfc3339_ns(most_recent_event.timestamp) - schedule > 239368000000000
    #check if the time difference is less than 22 s
    time.parse_rfc3339_ns(most_recent_event.timestamp) - schedule < 22000000000
}
#TemporaryViolatedToViolated if the last event is W_Completeren aanvraag-complete and the difference with the older W_Completeren aanvraag-schedule is more than 2 days, 18 h, 29 min, and 28 s
TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

#TemporaryViolatedToSatisfied if the last event is W_Completeren aanvraag-complete and the difference with the older W_Completeren aanvraag-schedule is less than 2 days, 18 h, 29 min, and 28 s and more than 22 s
TemporaryViolatedToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Completeren aanvraag"
    #Get the older W_Completeren aanvraag-schedule event
    schedule_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "W_Completeren aanvraag";e["lifecycle:transition"] == "SCHEDULE"]
    schedule := min(schedule_events) # This will be 0 if schedule_events is empty
    #check if the time difference is more than 22 s
    time.parse_rfc3339_ns(most_recent_event.timestamp) - schedule >= 22000000000
    #check if the time difference is less than 2 days, 18 h, 29 min, and 28 s
    time.parse_rfc3339_ns(most_recent_event.timestamp) - schedule <= 239368000000000
}

TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Completeren aanvraag"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Get the older W_Completeren aanvraag-schedule event
    schedule_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "W_Completeren aanvraag";e["lifecycle:transition"] == "SCHEDULE"]
    schedule := min(schedule_events) # This will be 0 if schedule_events is empty
    #check if the time difference is more than 2 days, 18 h, 29 min, and 28 s
    time.parse_rfc3339_ns(most_recent_event.timestamp) - schedule > 239368000000000
    #check if the time difference is less than 2 days, 18 h, 29 min, and 28 s
    #time.parse_rfc3339_ns(most_recent_event.timestamp) - schedule < 22000000000
}
TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Completeren aanvraag"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Get the older W_Completeren aanvraag-schedule event
    schedule_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "W_Completeren aanvraag";e["lifecycle:transition"] == "SCHEDULE"]
    schedule := min(schedule_events) # This will be 0 if schedule_events is empty
    #check if the time difference is more than 2 days, 18 h, 29 min, and 28 s
    #time.parse_rfc3339_ns(most_recent_event.timestamp) - schedule > 239368000000000
    #check if the time difference is less than 2 days, 18 h, 29 min, and 28 s
    time.parse_rfc3339_ns(most_recent_event.timestamp) - schedule < 22000000000
}

#TemporarySatisfiedToSatisfied condition if the last event is __END__
TemporarySatisfiedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}



