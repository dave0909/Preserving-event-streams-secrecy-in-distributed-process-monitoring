package R21

import rego.v1

#Description: Lifecycle : W_Afhandelen leads - SCHEDULE --> W_Afhandelen leads - START --> W_Afhandelen leads - COMPLETE is always respected

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

InitToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Afhandelen leads"
    most_recent_event["lifecycle:transition"] == "SCHEDULE"
}

InitToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Afhandelen leads"
    most_recent_event["lifecycle:transition"] != "SCHEDULE"
}

InitToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

TemporaryViolatedToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Afhandelen leads"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Check if W_Afhandelen leads - START is present
    checkStart(trace_id)
}

TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Afhandelen leads"
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Check if W_Afhandelen leads - START is present
    not checkStart(trace_id)
}
TemporarySatisfiedToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "W_Afhandelen leads"
    most_recent_event["lifecycle:transition"] == "SCHEDULE"
}

TemporarySatisfiedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

#Function that checks if W_Afhandelen leads - SCHEDULE is present for the same trace
checkStart(trace_id) if {
    input.events[_].concept_name == "W_Afhandelen leads"
    input.events[_]["lifecycle:transition"] == "START"
    input.events[_].trace_concept_name == trace_id
}
