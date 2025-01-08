package restock_goods

import rego.v1

#DESCRIPTION: When a Retrieve goods from the stock (RGFS) activity occours and product_units
#             is less than 1000, then the next activity must be a Restock goods (RG) activity.


# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

#InitToSatisfied if the last event is __END__
InitToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

InitToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Retrieve goods from the stock (RGFS)"
    most_recent_event["product_units"] < 1000
}

TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

TemporaryViolatedToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Restock goods (RG)"
}

TemporarySatifiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Retrieve goods from the stock (RGFS)"
    most_recent_event["product_units"] < 1000
}

TemporarySatisfiedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}