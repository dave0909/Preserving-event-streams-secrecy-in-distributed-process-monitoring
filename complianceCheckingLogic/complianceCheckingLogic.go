package complianceCheckingLogic


    import (
	"context"
	"fmt"
	"github.com/open-policy-agent/opa/rego"
	"log")
    
// Generated process constraints code

var constraintNames = []string{
"inspect_goods_within_onehour", "separation_of_duty", "shipment_cost", "truck_policy",
}

var constraints = []string{

`package inspect_goods_within_onehour

#CONSTRAINT: Inspect goods must happen within one hour from the "Truck reached costumer activity"
import rego.v1

# Define a rule to check if "Inspect goods" happens within one hour after "Truck reached customer" for each trace
inspect_goods_within_one_hour[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	truck_reached := input.events[_];truck_reached.trace_concept_name == trace_id;truck_reached.concept_name == "Truck reached costumer (TRC)"
	inspect_goods := input.events[_];inspect_goods.trace_concept_name == trace_id;inspect_goods.concept_name == "Inspect goods (IG)"
	time.parse_rfc3339_ns(inspect_goods.inspection_time) <= time.parse_rfc3339_ns(truck_reached.receipt_time) + 3600000000000
}

# Define a rule to check if "Inspect goods" activity is present in the trace
inspect_goods_present[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	count({e | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Inspect goods (IG)"}) > 0
}

# Define a rule to get all trace IDs that do not satisfy the condition
violations[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	inspect_goods_present[trace_id]
	not inspect_goods_within_one_hour[trace_id]
	}


`,

`package separation_of_duty
import rego.v1

# Define a rule to check if the logistics operators for "Fill in container" and "Check container" activities are different for each trace
check_operators_condition[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	e1 := input.events[_]
	e1.trace_concept_name == trace_id
	e1.concept_name == "Fill in container (FC)"
	e2 := input.events[_]
	e2.trace_concept_name == trace_id
	e2.concept_name == "Check container (CC)"
	e1.logistics_operator != e2.logistics_operator
}

# Define a rule to get all trace IDs that do not satisfy the operators condition
violations[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	not check_operators_condition[trace_id]
	# Ensure that both activities are present
	count({e | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Fill in container (FC)"}) > 0
	count({e | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Check container (CC)"}) > 0
}

`,

`package shipment_cost
import rego.v1

check_cost_condition[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	reserve_cost := sum([e.cost | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Reserve shipment (RS)"])
	drive_distance_i := sum([e.km_distance | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Drive to costumer (DC)"])
	drive_distance_m := sum([e.km_distance | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Drive to manufacturer (DM)"])
	reserve_cost <= (drive_distance_i + drive_distance_m) * 3
}

# Check if the "Detach container" activity has been executed
detach_container_executed[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	some d
	input.events[d].trace_concept_name == trace_id
	input.events[d].concept_name == "Detach container (DCO)"
}

# Define a rule to get all trace IDs that do not satisfy the cost condition and have executed the "Detach container" activity
violations[trace_id] if {
	trace_id := input.events[_].trace_concept_name
	detach_container_executed[trace_id]
	not check_cost_condition[trace_id]
}`,

`package truck_policy

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
`,

}


    type ComplianceCheckingLogic struct {
	preparedConstraints []rego.PreparedEvalQuery
	ctx                 context.Context
}

// Function that creates a prepared constraint for each constraint
func InitComplianceCheckingLogic() (ComplianceCheckingLogic, []string) {
	ctx := context.TODO()
	ccLogic := ComplianceCheckingLogic{
		preparedConstraints: []rego.PreparedEvalQuery{},
		ctx:                 ctx,
	}
	for i, constraint := range constraints {
		query, err := rego.New(
			rego.Query("data."+constraintNames[i]),
			rego.Module(constraintNames[i], constraint),
		).PrepareForEval(ctx)
		if err != nil {
			log.Fatal(err)
		}
		ccLogic.preparedConstraints = append(ccLogic.preparedConstraints, query)
	}
	return ccLogic, constraintNames
}

// Evalulate the event log with the prepared constraints
func (ccl *ComplianceCheckingLogic) EvaluateEventLog(eventLog map[string]interface{}) map[string]interface{} {
	violationMap := map[string]interface{}{}
	for _, preparedConstraint := range ccl.preparedConstraints {
		res, err := preparedConstraint.Eval(ccl.ctx, rego.EvalInput(eventLog))
		//For each key value couple in results[0]
		if err != nil {
			// Handle evaluation error.
			fmt.Println(err)
		}
		//For
		for constraintName, _ := range preparedConstraint.Modules() {
			resultValue := res[0].Expressions[0].Value
			resultValueMap, ok := resultValue.(map[string]interface{})
			if !ok {
				fmt.Println(res)
				log.Fatalf("Failed to convert result from policy inspection")
			}
			violations, ok := resultValueMap["violations"].(map[string]interface{})
			if !ok {
				fmt.Println(res)
				log.Fatalf("Failed to convert violation from policy inspection")
			}
			violationMap[constraintName] = violations
		}
	}
	return violationMap
}
