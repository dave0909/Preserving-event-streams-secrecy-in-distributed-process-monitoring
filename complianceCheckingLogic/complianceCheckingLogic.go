package complianceCheckingLogic


import (
    "context"
    "fmt"
    "github.com/open-policy-agent/opa/rego"
    "sync"
    "log"
)
type ConstraintState int
const (
	Init ConstraintState = iota
	Pending
	Violated
	Satisfied
	TemporarySatisfied
	TemporaryViolated
)

// Custom FSM struct using an integer matrix for transitions
type CustomFSM struct {
	Transitions [][]int
}

// Generated process constraints code

var constraintNames = []string{
"inspect_goods_within_onehour", "restock_goods", "separation_of_duty", "shipment_cost", "truck_policy"}

var constraints = []string{

`package inspect_goods_within_onehour
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

#temporary satisfied condition if the last event is Truck reached costumer (TRC)
InitToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Truck reached costumer (TRC)"
}
#temporary satisfied condition if the last event is "Inspect goods (IG)" and the difference with the older Truck reached costumer (TRC) is less than one hour
TemporaryViolatedToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Inspect goods (IG)"
    #Get the older Truck reached costumer (TRC) event
    reached_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Truck reached costumer (TRC)"]
    reached := min(reached_events) # This will be 0 if reached_events is empty
    #check if the fime difference is less than one hour
    time.parse_rfc3339_ns(most_recent_event.timestamp) - reached <= 3600000000000
}
#Violation condition if the last event is "Inspect goods (IG)" and the difference with the older Truck reached costumer (TRC) is more than one hour
TemporaryViolatedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Inspect goods (IG)"
    #Get the older Truck reached costumer (TRC) event
    reached_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Truck reached costumer (TRC)"]
    reached := min(reached_events) # This will be 0 if reached_events is empty
    #check if the fime difference is less than one hour
    time.parse_rfc3339_ns(most_recent_event.timestamp) - reached > 3600000000000
}
#Violation condition if the last event is "Inspect goods (IG)" and the difference with the older Truck reached costumer (TRC) is more than one hour
TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Inspect goods (IG)"
    #Get the older Truck reached costumer (TRC) event
    reached_events := [time.parse_rfc3339_ns(e.timestamp) | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Truck reached costumer (TRC)"]
    reached := min(reached_events) # This will be 0 if reached_events is empty
    #check if the fime difference is less than one hour
    time.parse_rfc3339_ns(most_recent_event.timestamp) - reached > 3600000000000
}

#Satisfied condition if the last event is "__END__"
TemporarySatisfiedToSatisfied[trace_id] if {
	trace_id := most_recent_event.trace_concept_name
	most_recent_event.concept_name == "__END__"
}

#Violated condition if the last event is "__END__"
TemporaryViolatedToViolated[trace_id] if {
	trace_id := most_recent_event.trace_concept_name
	most_recent_event.concept_name == "__END__"
}`,

`package restock_goods

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
}`,

`package separation_of_duty
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

# Pending condition 1
InitToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Fill in container (FC)"
}
# Pending condition 1
InitToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Fill in container (FC)"
}

# Violation condition 1
TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Fill in container (FC)"
    same_operator_exists(trace_id, "Fill in container (FC)", "Check container (CC)")
}

# Violation condition 2
TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Check container (CC)"
    same_operator_exists(trace_id, "Check container (CC)", "Fill in container (FC)")
}

# Satisfied condition, when the trace is over and the constraint is in pending state
TemporarySatisfiedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

# Define a rule to check if the same logistics operator exists for both activities in the same trace
same_operator_exists(trace_id, activity1, activity2) if {
    e1 := input.events[_]
    e1.trace_concept_name == trace_id
    e1.concept_name == activity1
    e2 := input.events[_]
    e2.trace_concept_name == trace_id
    e2.concept_name == activity2
    e1.logistics_operator == e2.logistics_operator
}`,

`package shipment_cost
import rego.v1

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

#Temporary satisfied condition if the last event is "Reserve shipment (RS)" and the cost condition is satisfied
InitToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Shipment reservation sent (SRS)"
	check_cost_condition(trace_id)
}
#Temporary satisfied condition if the last event is "Reserve shipment (RS)" and the cost condition is satisfied
InitToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Shipment reservation sent (SRS)"
	not check_cost_condition(trace_id)
}

TemporaryViolatedToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Drive to costumer (DC)"
    check_cost_condition(trace_id)
}

TemporaryViolatedToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Drive to manufacturer (DM)"
    check_cost_condition(trace_id)
}

TemporarySatisfiedToTemporaryViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Shipment reservation sent (SRS)"
    not check_cost_condition(trace_id)
}

# Define a function to check the cost condition
check_cost_condition(trace_id) if {
    reserve_cost := max([e.cost | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Shipment reservation sent (SRS)"])
    drive_distance_i := sum([e.km_distance | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Drive to costumer (DC)"])
    drive_distance_m := sum([e.km_distance | e := input.events[_]; e.trace_concept_name == trace_id; e.concept_name == "Drive to manufacturer (DM)"])
    reserve_cost <= (drive_distance_i + drive_distance_m) * 3
}
#Satisfied condition if the last event is "__END__"
TemporarySatisfiedToSatisfied[trace_id] if {
	trace_id := most_recent_event.trace_concept_name
	most_recent_event.concept_name == "__END__"
}
#Violated condition if the last event is "__END__"
TemporaryViolatedToViolated[trace_id] if {
	trace_id := most_recent_event.trace_concept_name
	most_recent_event.concept_name == "__END__"
}`,

`package truck_policy
import rego.v1

# Define the constant for five years in seconds
five_years_in_seconds := 5 * 365 * 24 * 60 * 60

# Get the most recent event
most_recent_event := input.events[count(input.events) - 1]

# Temporary satisfied condition if the last event is not "Select truck (ST)"
InitToTemporarySatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    1==1
}

# Violation condition
TemporarySatisfiedToViolated[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "Select truck (ST)"
    driver_experience_within_five_years[trace_id]
}

# Satisfied condition
TemporarySatisfiedToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}
InitToSatisfied[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    most_recent_event.concept_name == "__END__"
}

# Define a rule to check if the driver's experience is within five years
driver_experience_within_five_years[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    event := most_recent_event
    (time.parse_rfc3339_ns(event.timestamp) / 1000000000) - (time.parse_rfc3339_ns(event.license_first_issue) / 1000000000) <= five_years_in_seconds
}

## Define a rule to check if the driver's experience is within five years
driver_experience_within_five_years[trace_id] if {
    trace_id := most_recent_event.trace_concept_name
    event := most_recent_event
    (time.parse_rfc3339_ns(event.timestamp) / 1000000000) - (time.parse_rfc3339_ns(event.license_first_issue) / 1000000000) <= five_years_in_seconds
}`,

}


// Method to check possible next states
func (fsm *CustomFSM) PossibleNextStates(currentState int) []int {
	return fsm.Transitions[currentState]
}

// Method to check if there is a transition from state s1 to state s2
func (fsm *CustomFSM) HasTransition(s1, s2 int) bool {
	for _, nextState := range fsm.PossibleNextStates(s1) {
		if nextState == s2 {
			return true
		}
	}
	return false
}

type Constraint struct {
	name              string
	preparedEvalQuery rego.PreparedEvalQuery
	fsm               *CustomFSM
	ConstraintState   map[string]ConstraintState
}

type ComplianceCheckingLogic struct {
	preparedConstraints []Constraint
	ctx                 context.Context
}

// Function that creates a prepared constraint for each constraint
func InitComplianceCheckingLogic() (ComplianceCheckingLogic, []string) {
	ctx := context.TODO()
	ccLogic := ComplianceCheckingLogic{
		preparedConstraints: []Constraint{},
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
		ccLogic.preparedConstraints = append(ccLogic.preparedConstraints, Constraint{
			name:              constraintNames[i],
			preparedEvalQuery: query,
			fsm:               fsmMap[constraintNames[i]],
			ConstraintState:   make(map[string]ConstraintState),
		})
	}

	return ccLogic, constraintNames
}

// Evaluate the event log with the prepared constraints
func (ccl *ComplianceCheckingLogic) EvaluateEventLog(eventLog map[string][]map[string]interface{}) map[string]interface{} {
	lastEvent := eventLog["events"][len(eventLog["events"])-1]
	traceId := lastEvent["trace_concept_name"].(string)
	for _, constraint := range ccl.preparedConstraints {
		if _, ok := constraint.ConstraintState[traceId]; !ok {
			constraint.ConstraintState[traceId] = Init // Init state
		}
	}
	resultMap := map[string]interface{}{}
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, constraint := range ccl.preparedConstraints {
		wg.Add(1)
		go func(constraint Constraint) {
			defer wg.Done()
			res, err := constraint.preparedEvalQuery.Eval(ccl.ctx, rego.EvalInput(eventLog))
			if err != nil {
				fmt.Println(err)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			//currentState := constraint.ConstraintState[traceId]
			//if constraint.name == "truck_policy" {
			//	fmt.Println(res)
			//}
			for {
				transitionFound := false
				currentState := constraint.ConstraintState[traceId]
				for _, nextState := range constraint.fsm.PossibleNextStates(int(currentState)) {
					currentState := constraint.ConstraintState[traceId]
					ruleName := fmt.Sprintf("%sTo%s", stateName(currentState), stateName(ConstraintState(nextState)))
					if resultValue, ok := res[0].Expressions[0].Value.(map[string]interface{})[ruleName]; ok {
						//if constraint.name == "shipment_cost" {
						//	fmt.Println("Constraint name: ", constraint.name, "in state ", constraint.ConstraintState, "next state: ", stateName(ConstraintState(nextState)), "rulename: ", ruleName, "resultValue: ", resultValue)
						//	fmt.Println("Result: ", res)
						//}
						if resultValueMap, ok := resultValue.(map[string]interface{}); ok {
							for caseId, isTrue := range resultValueMap {
								if isTrue.(bool) {
									constraint.ConstraintState[caseId] = ConstraintState(nextState)
									fmt.Printf("Constraint %s transitioned from %s to %s for case %s", constraint.name, stateName(currentState), stateName(ConstraintState(nextState)), caseId)
									fmt.Println()
									resultMap[traceId] = nextState
									transitionFound = false
									//TODO: set the above variable to true to enable the recursive inspection of the constraints
								}
							}
						}
					}
				}
				if !transitionFound {
					break
				}
			}

		}(constraint)
	}
	wg.Wait()
	return resultMap
}

func stateName(state ConstraintState) string {
	switch state {
	case Init:
		return "Init"
	case Pending:
		return "Pending"
	case Violated:
		return "Violated"
	case Satisfied:
		return "Satisfied"
	case TemporarySatisfied:
		return "TemporarySatisfied"
	case TemporaryViolated:
		return "TemporaryViolated"
	default:
		return "Unknown"
	}
}


var fsmMap = map[string]*CustomFSM{

"inspect_goods_within_onehour": {
    Transitions: [][]int{
        {5},
        {},
        {},
        {},
        {2, 3},
        {2, 4, 2},
    },
},

"restock_goods": {
    Transitions: [][]int{
        {5, 3},
        {},
        {},
        {},
        {3, 5},
        {2, 4},
    },
},

"separation_of_duty": {
    Transitions: [][]int{
        {4},
        {},
        {},
        {},
        {2, 3},
        {},
    },
},

"shipment_cost": {
    Transitions: [][]int{
        {4, 5},
        {},
        {},
        {},
        {5, 3},
        {4, 2},
    },
},

"truck_policy": {
    Transitions: [][]int{
        {4, 2, 3},
        {},
        {},
        {},
        {2, 3},
        {},
    },
},

}
