package truck_policy
import rego.v1

default five_years_in_seconds := 157766400
default driver_experience_violation :=false

driver_experience_violation if {
	trace_name := input.events[_].trace_concept_name
	violation[trace_name]
}

violations[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	some event in input.events
	event.concept_name == "Select truck"
	event.timestamp - event.license_first_issue < five_years_in_seconds
	event.trace_concept_name == trace_id
}



#{
#	"events": [
#    	{
#        	"trace_concept_name": "0",
#        	"concept_name": "Select truck",
#        	"truck_id": "truck_1",
#        	"timestamp": 1726,
#        	"driving_license_code": "E4D2456",
#        	"license_first_issue": 14109
#    	},
#    	{
#        	"trace:concept:name": "0",
#        	"concept_name": "Drive to industry",
#        	"truck_id": "truck_2",
#        	"distance": 50,
#        	"timestamp": 123123124
#    	},
#    	{
#        	"trace_concept_name": "0",
#        	"concept_name": "Drive to manufacturer",
#        	"truck_id": "truck_2",
#        	"distance": 20,
#        	"timestamp": 123123125
#    	},
#    	{
#        	"trace_concept_name": "1",
#        	"concept_name": "Select truck",
#        	"truck_id": "truck_1",
#        	"timestamp": 1726577947,
#        	"driving_license_code": "E4D2456",
#        	"license_first_issue": 172
#    	},
#    	{
#        	"trace_concept_name": "1",
#        	"concept_name": "Drive to industry",
#        	"truck_id": "truck_1",
#        	"distance": 90,
#        	"timestamp": 1726577498
#    	}
#	]
#}
