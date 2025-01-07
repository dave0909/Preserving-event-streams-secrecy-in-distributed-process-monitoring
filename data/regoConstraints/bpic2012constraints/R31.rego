package R31

import rego.v1

#DESCRIPTION: A_SUBMITTED-complete” and “A_PARTLYSUBMITTED-
              #complete” are always performed by the same actor

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

target_events := {"A_SUBMITTED", "A_FINALIZED"}


InitToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

InitToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == target_events[_]
    most_recent_event["lifecycle:transition"] == "COMPLETE"
}

#TemporaryViolatedToViolated if last event is in target_events and the actor is different from the actor of the previous events in target_events
TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == target_events[_]
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Check if the resource attribute of the current event is equal to all resources of all the events in target_events}
    checkActorViolation(trace_id)
}

TemporaryViolatedToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == target_events[_]
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Check if the resource attribute of the current event is equal to all resources of all the events in target_events}
    not checkActorViolation(trace_id)
}

TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == target_events[_]
    most_recent_event["lifecycle:transition"] == "COMPLETE"
    #Check if the resource attribute of the current event is equal to all resources of all the events in target_events}
    checkActorViolation(trace_id)
}

TemporarySatisfiedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

#This is true if the resource attribute of the current event is not equal to all resources of all the events in target_events
checkActorViolation(trace_id) if {
    # Get the actor attribute of the current event
    actor := most_recent_event["org:resource"]
    # Get the actor attribute of all the events in target_events, excluding the current event
    actors := {e["org:resource"] | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == target_events[_];e.concept_name != most_recent_event.concept_name}
    # Check if the actor attribute of the current event is in actors attribute of all the events in target_events
    actor == actors[_]
}