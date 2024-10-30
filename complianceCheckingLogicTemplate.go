package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/open-policy-agent/opa/rego"
	"log"
	"main/complianceCheckingLogic"
)

var constraintNames2 = []string{
	"truck_policy",
}
var constraints2 = []string{
	"package truck_policy\nimport rego.v1\n\ndefault five_years_in_seconds := 157766400\ndefault driver_experience_violation :=false\n\ndriver_experience_violation if {\n\ttrace_name := input.events[_].trace_concept_name\n    violation[trace_name]\n}\n\nviolation[trace_id] if {\n    trace_id := input.events[_].trace_concept_name\n    some event in input.events\n    event.concept_name == \"Select truck\"\n    event.timestamp - event.license_first_issue < five_years_in_seconds\n    event.trace_concept_name == trace_id\n}",
}

type ComplianceCheckingLogic2 struct {
	preparedConstraints []rego.PreparedEvalQuery
	ctx                 context.Context
}

// Function that create a prepared constraint for each constraint
func InitComplianceCheckingLogic2() (ComplianceCheckingLogic2, error) {
	ctx := context.TODO()
	ccLogic := ComplianceCheckingLogic2{
		preparedConstraints: []rego.PreparedEvalQuery{},
		ctx:                 ctx,
	}
	for i, constraint := range complianceCheckingLogic.constraints {
		query, err := rego.New(
			rego.Query("data."+constraintNames2[i]),
			rego.Module(constraintNames2[i], constraint),
		).PrepareForEval(ctx)
		if err != nil {
			return ComplianceCheckingLogic2{}, err
		}
		ccLogic.preparedConstraints = append(ccLogic.preparedConstraints, query)
	}
	return ccLogic, nil
}

// Evalulate the event log with the prepared constraints
func (ccl *ComplianceCheckingLogic2) evaluateEventLog(eventLog map[string]interface{}) map[string]interface{} {
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
				log.Fatalf("Failed to convert result from policy inspection")
			}
			violations, ok := resultValueMap["violation"].(map[string]interface{})
			if !ok {
				log.Fatalf("Failed to convert violation from policy inspection")
			}
			violationMap[constraintName] = violations
		}
	}
	return violationMap
}

//func tryRego() {
//	module := `
//package truck_policy
//import rego.v1
//
//default five_years_in_seconds := 157766400
//default driver_experience_violation :=false
//
//driver_experience_violation if {
//	trace_name := input.events[_].trace_concept_name
//    violation[trace_name]
//}
//
//violation[trace_id] if {
//    trace_id := input.events[_].trace_concept_name
//    some event in input.events
//    event.concept_name == "Select truck"
//    event.timestamp - event.license_first_issue < five_years_in_seconds
//    event.trace_concept_name == trace_id
//}
//`
//	ctx := context.TODO()
//	query, err := rego.New(
//		rego.Query("data.truck_policy"),
//		rego.Module("example.rego", module),
//	).PrepareForEval(ctx)
//	if err != nil {
//		// Handle error.
//		fmt.Println(err)
//	}
//	input, _ := getInput()
//	res, err := query.Eval(ctx, rego.EvalInput(input))
//	if err != nil {
//		// Handle evaluation error.
//		fmt.Println(err)
//	}
//	fmt.Println(res)
//}

// Main function
func main() {
	ccLogic, err := InitComplianceCheckingLogic2()
	if err != nil {
		fmt.Println(err)
	}
	input, _ := getInput()
	violation := ccLogic.evaluateEventLog(input)
	fmt.Println(violation)
}

func getInput() (map[string]interface{}, string) {

	// Stringa JSON
	jsonData := `
{
    "events": [
        {
            "trace_concept_name": "0",
            "concept_name": "Select truck",
            "truck_id": "truck_1",
            "timestamp": 1726577947,
            "driving_license_code": "E4D2456",
            "license_first_issue": 14109
        },
        {
            "trace:concept:name": "0",
            "concept_name": "Drive to industry",
            "truck_id": "truck_2",
            "distance": 50,
            "timestamp": 123123124
        },
        {
            "trace_concept_name": "0",
            "concept_name": "Drive to manufacturer",
            "truck_id": "truck_2",
            "distance": 20,
            "timestamp": 123123125
        },
        {
            "trace_concept_name": "1",
            "concept_name": "Select truck",
            "truck_id": "truck_1",
            "timestamp": 1726577947,
            "driving_license_code": "E4D2456",
            "license_first_issue": 1726570747
        },
        {
            "trace_concept_name": "1",
            "concept_name": "Drive to industry",
            "truck_id": "truck_1",
            "distance": 90,
            "timestamp": 1726577498
        }
    ]
}`

	// Creare una variabile map
	var data map[string]interface{}

	// Decodifica del JSON in una mappa
	err := json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		log.Fatal(err)
	}

	return data, jsonData
}
