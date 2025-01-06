package main

import (
	"fmt"
	"main/complianceCheckingLogic"
)

// main function
func main() {
	ccLogic, _ := complianceCheckingLogic.InitComplianceCheckingLogic()
	//eventLog := map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name":  "1",
	//			"concept_name":        "Select truck (ST)",
	//			"timestamp":           "2021-10-01T00:00:00Z",
	//			"license_first_issue": "2017-10-01T00:00:00Z",
	//		},
	//	},
	//}
	//ccLogic.EvaluateEventLog(eventLog)
	//
	////Test the separation of duty constraint
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Fill in container (FC)",
	//			"logistics_operator": "operator1",
	//			"timestamp":          "2021-10-01T00:00:00Z",
	//		},
	//	},
	//}
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Fill in container (FC)",
	//			"logistics_operator": "operator1",
	//			"timestamp":          "2021-10-01T00:00:00Z",
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Check container (CC)",
	//			"logistics_operator": "operator2",
	//			"timestamp":          "2021-10-01T00:00:00Z",
	//		},
	//	},
	//}
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Fill in container (FC)",
	//			"logistics_operator": "operator1",
	//			"timestamp":          "2021-10-01T00:00:00Z",
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Check container (CC)",
	//			"logistics_operator": "operator2",
	//			"timestamp":          "2021-10-01T00:00:00Z",
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "__END__",
	//		},
	//	},
	//}
	//ccLogic.EvaluateEventLog(eventLog)
	//
	////// Test the inspect goods within one hour constraint
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Truck reached costumer (TRC)",
	//			"timestamp":          "2021-10-01T00:00:00Z",
	//		},
	//	},
	//}
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Truck reached costumer (TRC)",
	//			"timestamp":          "2021-10-01T00:00:00Z",
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Inspect goods (IG)",
	//			"timestamp":          "2021-10-01T01:00:00Z",
	//		},
	//	}}
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Truck reached costumer (TRC)",
	//			"timestamp":          "2021-10-01T00:00:00Z",
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Inspect goods (IG)",
	//			"timestamp":          "2021-10-01T01:00:00Z",
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Inspect goods (IG)",
	//			"timestamp":          "2021-10-01T01:00:00Z",
	//		},
	//	}}
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Truck reached costumer (TRC)",
	//			"timestamp":          "2021-10-01T00:00:00Z",
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Inspect goods (IG)",
	//			"timestamp":          "2021-10-01T01:00:00Z",
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Inspect goods (IG)",
	//			"timestamp":          "2021-10-01T01:00:00Z",
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "__END__",
	//		},
	//	}}
	//ccLogic.EvaluateEventLog(eventLog)
	////Test the shipment cost constraint
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Shipment reservation sent (SRS)",
	//			"cost":               10,
	//		},
	//	},
	//}
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Shipment reservation sent (SRS)",
	//			"cost":               10,
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Drive to costumer (DC)",
	//			"km_distance":        2,
	//		},
	//	},
	//}
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Shipment reservation sent (SRS)",
	//			"cost":               10,
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Drive to costumer (DC)",
	//			"km_distance":        2,
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Drive to costumer (DC)",
	//			"km_distance":        200,
	//		},
	//	},
	//}
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Shipment reservation sent (SRS)",
	//			"cost":               10,
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Drive to costumer (DC)",
	//			"km_distance":        2,
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Drive to costumer (DC)",
	//			"km_distance":        200,
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Shipment reservation sent (SRS)",
	//			"cost":               50000000,
	//		},
	//	},
	//}
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Shipment reservation sent (SRS)",
	//			"cost":               10,
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Drive to costumer (DC)",
	//			"km_distance":        2,
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Drive to costumer (DC)",
	//			"km_distance":        200,
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Shipment reservation sent (SRS)",
	//			"cost":               5000000000,
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "__END__",
	//		},
	//	},
	//}
	//ccLogic.EvaluateEventLog(eventLog)

	//Test iv_antibiotics_within_onehour constraint---------------------------------
	//eventLog := map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "ER Sepsis Triage",
	//			"timestamp":          "2021-10-01T00:00:00Z",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 1")
	//ccLogic.EvaluateEventLog(eventLog)
	//
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "ER Sepsis Triage",
	//			"timestamp":          "2021-10-01T00:00:00Z",
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "IV Antibiotics",
	//			"timestamp":          "2021-10-01T01:00:00Z",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 2")
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "ER Sepsis Triage",
	//			"timestamp":          "2021-10-01T00:00:00Z",
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "IV Antibiotics",
	//			"timestamp":          "2021-10-01T01:00:00Z",
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "LacticAcid",
	//			"timestamp":          "2021-10-01T00:30:00Z",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 3")
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "ER Sepsis Triage",
	//			"timestamp":          "2021-10-01T00:00:00Z",
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "IV Antibiotics",
	//			"timestamp":          "2021-10-01T01:00:00Z",
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "__END__",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 4")
	//ccLogic.EvaluateEventLog(eventLog)

	//Test the no_instert_no_collect  "if ¬InsertFine then ¬Collect"---------------------------------
	//eventLog := map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Insert Fine Notification",
	//			"timestamp":          "2021-10-01T00:00:00Z",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 1")
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Insert Fine Notification",
	//			"timestamp":          "2021-10-01T00:00:00Z",
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "Send for Credit Collection",
	//			"timestamp":          "2021-10-01T01:00:00Z",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 2")
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "__END__",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 3")
	//ccLogic.EvaluateEventLog(eventLog)

	////Test R17 constraint-----------------------------------------------------------------------------
	//eventLog := map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "A_PARTLYSUBMITTED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 1")
	////TODO:here it directly goes to Violated state (wrong), FIX recursive inspection of consraints
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "A_PARTLYSUBMITTED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "A_PARTLYSUBMITTED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 2")
	//ccLogic.EvaluateEventLog(eventLog)
	//Test R18 constraint-----------------------------------------------------------------------------
	//eventLog := map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "A_SUBMITTED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 1")
	////TODO:here it directly goes to Violated state (wrong), FIX recursive inspection of consraints
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "A_SUBMITTED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "A_SUBMITTED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 2")
	//ccLogic.EvaluateEventLog(eventLog)

	//Test R19 constraint-----------------------------------------------------------------------------
	//eventLog := map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "A_SUBMITTED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 1")
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "A_SUBMITTED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//		{"trace_concept_name": "1",
	//			"concept_name":         "A_PARTLYSUBMITTED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 2")
	//ccLogic.EvaluateEventLog(eventLog)

	//Test R20 constraint-----------------------------------------------------------------------------
	//eventLog := map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "A_SUBMITTED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 1")
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "A_SUBMITTED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//		{"trace_concept_name": "1",
	//			"concept_name":         "A_PARTLYSUBMITTED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 2")
	//ccLogic.EvaluateEventLog(eventLog)

	//Test R21 constraint-----------------------------------------------------------------------------
	//eventLog := map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "W_Afhandelen leads",
	//			"lifecycle:transition": "SCHEDULE",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 1")
	//ccLogic.EvaluateEventLog(eventLog)
	//
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "W_Afhandelen leads",
	//			"lifecycle:transition": "SCHEDULE",
	//		},
	//		{"trace_concept_name": "1",
	//			"concept_name":         "W_Afhandelen leads",
	//			"lifecycle:transition": "START",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 2")
	//ccLogic.EvaluateEventLog(eventLog)
	//
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "W_Afhandelen leads",
	//			"lifecycle:transition": "SCHEDULE",
	//		},
	//		{"trace_concept_name": "1",
	//			"concept_name":         "W_Afhandelen leads",
	//			"lifecycle:transition": "START",
	//		},
	//		{"trace_concept_name": "1",
	//			"concept_name":         "W_Afhandelen leads",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 3")
	//ccLogic.EvaluateEventLog(eventLog)
	//
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "W_Afhandelen leads",
	//			"lifecycle:transition": "SCHEDULE",
	//		},
	//		{"trace_concept_name": "1",
	//			"concept_name":         "W_Afhandelen leads",
	//			"lifecycle:transition": "START",
	//		},
	//		{"trace_concept_name": "1",
	//			"concept_name":         "W_Afhandelen leads",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//		{"trace_concept_name": "1",
	//			"concept_name": "__END__",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 4")
	//ccLogic.EvaluateEventLog(eventLog)

	//Test R22 constraint-----------------------------------------------------------------------------
	//eventLog := map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "W_Completeren aanvraag",
	//			"lifecycle:transition": "SCHEDULE",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 1")
	//ccLogic.EvaluateEventLog(eventLog)
	//
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "W_Completeren aanvraag",
	//			"lifecycle:transition": "SCHEDULE",
	//		},
	//		{"trace_concept_name": "1",
	//			"concept_name":         "W_Completeren aanvraag",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//	},
	//}
	//
	//fmt.Println("Evaluation 2")
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "W_Completeren aanvraag",
	//			"lifecycle:transition": "SCHEDULE",
	//		},
	//		{"trace_concept_name": "1",
	//			"concept_name":         "W_Completeren aanvraag",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "W_Completeren aanvraag",
	//			"lifecycle:transition": "SCHEDULE",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 3")
	//ccLogic.EvaluateEventLog(eventLog)
	//
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "W_Completeren aanvraag",
	//			"lifecycle:transition": "SCHEDULE",
	//		},
	//		{"trace_concept_name": "1",
	//			"concept_name":         "W_Completeren aanvraag",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "W_Completeren aanvraag",
	//			"lifecycle:transition": "SCHEDULE",
	//		},
	//		{
	//			"trace_concept_name": "1",
	//			"concept_name":       "__END__",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 4")
	//ccLogic.EvaluateEventLog(eventLog)
	//Test R24 constraint-----------------------------------------------------------------------------
	//eventLog := map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "O_SELECTED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 1")
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "O_SELECTED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "O_CREATED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 2")
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "O_SELECTED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "O_CREATED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "O_SENT",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 3")
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "O_SELECTED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "O_CREATED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "O_SENT",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "__END__",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 4")
	//ccLogic.EvaluateEventLog(eventLog)

	//Test R25 constraint-----------------------------------------------------------------------------
	//eventLog := map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "A_APPROVED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 1")
	//ccLogic.EvaluateEventLog(eventLog)
	//eventLog = map[string][]map[string]interface{}{
	//	"events": {
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "A_APPROVED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//		{
	//			"trace_concept_name":   "1",
	//			"concept_name":         "A_CANCELLED",
	//			"lifecycle:transition": "COMPLETE",
	//		},
	//	},
	//}
	//fmt.Println("Evaluation 2")
	//ccLogic.EvaluateEventLog(eventLog)
	//TODO: test R28 constraint
	eventLog := map[string][]map[string]interface{}{
		"events": {
			{
				"trace_concept_name":   "1",
				"concept_name":         "W_Completeren aanvraag",
				"lifecycle:transition": "SCHEDULE",
				"timestamp":            "2021-10-01T00:00:00Z",
			},
		},
	}
	fmt.Println("Evaluation 1")
	ccLogic.EvaluateEventLog(eventLog)
	eventLog = map[string][]map[string]interface{}{
		"events": {
			{
				"trace_concept_name":   "1",
				"concept_name":         "W_Completeren aanvraag",
				"lifecycle:transition": "SCHEDULE",
				"timestamp":            "2021-10-01T00:00:00Z",
			},
			{
				"trace_concept_name":   "1",
				"concept_name":         "W_Completeren aanvraag",
				"lifecycle:transition": "COMPLETE",
				"timestamp":            "2021-10-01T00:00:27Z",
			},
		},
	}
	fmt.Println("Evaluation 2")
	ccLogic.EvaluateEventLog(eventLog)
	eventLog = map[string][]map[string]interface{}{
		"events": {
			{
				"trace_concept_name":   "1",
				"concept_name":         "W_Completeren aanvraag",
				"lifecycle:transition": "SCHEDULE",
				"timestamp":            "2021-10-01T00:00:00Z",
			},
			{
				"trace_concept_name":   "1",
				"concept_name":         "W_Completeren aanvraag",
				"lifecycle:transition": "COMPLETE",
				"timestamp":            "2021-10-01T00:00:27Z",
			},
			{
				"trace_concept_name":   "1",
				"concept_name":         "__END__",
				"lifecycle:transition": "COMPLETE",
			},
		},
	}
	fmt.Println("Evaluation 3")
	ccLogic.EvaluateEventLog(eventLog)

}
