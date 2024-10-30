package separation_of_duty
import rego.v1

# Define a rule to check if the logistics operators for "Fill in container" and "Check container" activities are different for each trace
check_operators_condition[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	e1 := input.events[_]
	e1.trace_concept_name == trace_id
	e1.concept_name == "Fill in container"
	e2 := input.events[_]
	e2.trace_concept_name == trace_id
	e2.concept_name == "Check container"
	e1.logistics_operator != e2.logistics_operator
}

# Define a rule to get all trace IDs that do not satisfy the operators condition
violations[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	not check_operators_condition[trace_id]
	# Ensure that both activities are present
	count({e | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Fill in container"}) > 0
	count({e | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Check container"}) > 0
}



#//INPUT
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
#            "logistics_operator": "A4472",
#            "cost": 600
#        },
#        {
#            "trace_concept_name": "0",
#            "concept_name": "Check container",
#            "logistics_operator": "A4472",
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
#            "logistics_operator": "A4471",
#            "cost": 600
#        },
#        {
#            "trace_concept_name": "1",
#            "concept_name": "Check container",
#            "logistics_operator": "A4475",
#            "cost": 600
#        },
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
#
#
#//INPUT
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
#            "logistics_operator": "A4472",
#            "cost": 600
#        },
#        {
#            "trace_concept_name": "0",
#            "concept_name": "Check container",
#            "logistics_operator": "A4472",
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
#            "trace_concept_name": "0",
#            "concept_name": "Truck reached costumer",
#            "truck_id": "truck_2",
#            "km_distance": 100,
#            "timestamp": 1726664138
#        },
#        {
#            "trace_concept_name": "0",
#            "concept_name": "Inspect goods",
#            "truck_id": "truck_2",
#            "km_distance": 100,
#            "timestamp": 1726664138
#        },
#        {
#            "trace_concept_name": "1",
#            "concept_name": "Reserve shipment",
#            "cost": 150
#        },
#        {
#            "trace_concept_name": "1",
#            "concept_name": "Fill in container",
#            "logistics_operator": "A4471",
#            "cost": 600
#        },
#        {
#            "trace_concept_name": "1",
#            "concept_name": "Check container",
#            "logistics_operator": "A4475",
#            "cost": 600
#        },
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
