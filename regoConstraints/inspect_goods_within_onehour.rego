package inspect_goods_within_onehour

#CONSTRAINT: Inspect goods must happen within one hour from the "Truck reached costumer activity"
import rego.v1

# Define a rule to check if "Inspect goods" happens within one hour after "Truck reached customer" for each trace
inspect_goods_within_one_hour[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	truck_reached := input.events[_];truck_reached.trace_concept_name == trace_id;truck_reached.concept_name == "Truck reached costumer"
	inspect_goods := input.events[_];inspect_goods.trace_concept_name == trace_id;inspect_goods.concept_name == "Inspect goods"
	inspect_goods.timestamp <= truck_reached.timestamp + 3600
}

# Define a rule to check if "Inspect goods" activity is present in the trace
inspect_goods_present[trace_id] {
	trace_id := input.events[_].trace_concept_name
	count({e | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Inspect goods"}) > 0
}

# Define a rule to get all trace IDs that do not satisfy the condition
violations[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	inspect_goods_present[trace_id]
	not inspect_goods_within_one_hour[trace_id]
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
#            "timestamp": 1726667738
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