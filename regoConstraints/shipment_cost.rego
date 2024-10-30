package shipment_cost
import rego.v1

check_cost_condition[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	reserve_cost := sum([e.cost | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Reserve shipment"])
	drive_distance_i := sum([e.km_distance | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Drive to industry"])
	drive_distance_m := sum([e.km_distance | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Drive to manufacturer"])
	reserve_cost <= (drive_distance_i + drive_distance_m) * 3
}

# Define a rule to get all trace IDs that do not satisfy the cost condition
violations[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	not check_cost_condition[trace_id]
}

#INPUT
#{
#    "events": [
#        {
#            "trace_concept_name": "0",
#            "concept_name": "Reserve shipment",
#            "cost": 600
#        },
#        {
#            "trace_concept_name": "0",
#            "concept_name": "Fill in container",
#	    "Logistics operator":"A4475",
#            "cost": 600
#        },
#        {
#            "trace_concept_name": "0",
#            "concept_name": "Check container",
#	    "Logistics operator":"A4475",
#            "cost": 600
#        },
#        {
#            "trace_concept_name": "0",
#            "concept_name": "Drive to industry",
#            "truck_id": "truck_2",
#            "km_distance": 100,
#            "timestamp": 123123124
#        },
#        {
#            "trace_concept_name": "0",
#            "concept_name": "Drive to manufacturer",
#            "truck_id": "truck_2",
#            "km_distance": 100,
#            "timestamp": 123123125
#        },
#        {
#            "trace_concept_name": "1",
#            "concept_name": "Reserve shipment",
#            "cost": 150
#        },
#        {
#            "trace_concept_name": "1",
#            "concept_name": "Fill in container",
#	    "Logistics operator":"A4475",
#            "cost": 600
#        },
#        {
#            "trace_concept_name": "1",
#            "concept_name": "Check container",
#	    "Logistics operator":"A4475",
#            "cost": 600
#        },
#
#        {
#            "trace_concept_name": "1",
#            "concept_name": "Drive to industry",
#            "truck_id": "truck_2",
#            "km_distance": 200,
#            "timestamp": 123123124
#        },
#        {
#            "trace_concept_name": "1",
#            "concept_name": "Drive to manufacturer",
#            "truck_id": "truck_2",
#            "km_distance": 20,
#            "timestamp": 123123125
#        }
#    ]
#}

