package workflowLogic

import "main/utils/petrinet"


import "fmt"


// Generated Petri Net Code

var places = []string{
"n1", "n10", "n11", "n12", "n2", "n3", "n4", "n5", "n6", "n7", "n8", "n9", "sink", "source",
}

var transitions = []string{
"A_CANCELLED", "A_DECLINED", "A_PARTLYSUBMITTED", "A_PREACCEPTED", "A_SUBMITTED", "W_Afhandelen leads", "W_Afhandelen leads", "W_Afhandelen leads", "W_Completeren aanvraag", "W_Completeren aanvraag", "W_Completeren aanvraag",
}

var inputMatrix = [][]int{

    {0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
    {0, 1, 1, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0},
    {0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0},
    {0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0},
    {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
    {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0},
    {0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0},
    {1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
    {0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0},
    {0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
    {0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0},
}

var outputMatrix = [][]int{

    {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0},
    {0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 0, 1, 0},
    {0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0},
    {0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0},
    {0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0},
    {1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
    {0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0},
    {0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
    {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0},
    {0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
    {0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0},
}

var initialMarking = []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

// Indices of transitions associated with gateways

var silentTransitionIndices = []int{}


type WorkflowLogic struct {
	Petrinet          petrinet.Net
	Places            []string
	Transitions       []string
	SilentTransitions []int
}

func InitWorkflowLogic() WorkflowLogic {
	wf := WorkflowLogic{
		Petrinet: petrinet.Net{
			InputMatrix:  inputMatrix,
			OutputMatrix: outputMatrix,
			State:        initialMarking,
		},
		Places:            places,
		Transitions:       transitions,
		SilentTransitions: silentTransitionIndices,
	}
	wf.Petrinet.Init()
	return wf
}

// Get index of a transition by its name
func (wf *WorkflowLogic) GetTransitionIndicesByName(name string) []int {
	indices := []int{}
	for i, t := range wf.Transitions {
		if t == name {
			indices = append(indices, i)
		}
	}
	return indices
}

//func (wf *WorkflowLogic) GetTransitionIndexByName(name string) int {
//	for i, t := range wf.Transitions {
//		if t == name {
//			return i
//		}
//	}
//	return -1
//}

//	func (wf *WorkflowLogic) FireTokenIdWithTransitionName(activityName string, caseId int) error {
//		transitionIndex := wf.GetTransitionIndexByName(activityName)
//		error := wf.Petrinet.FireWithTokenId(transitionIndex, caseId)
//		if error == nil {
//			// Loop to handle recursive firing of silent transitions
//			for {
//				enabledTransitions := wf.Petrinet.GetEnabledTransitionsForTokenId(caseId)
//				silentFired := false
//				for _, t := range wf.SilentTransitions {
//					for _, et := range enabledTransitions {
//						if et == t {
//							// Fire the silent transition
//							wf.Petrinet.FireWithTokenId(t, caseId)
//							silentFired = true
//						}
//					}
//				}
//				// If no silent transition was fired, break the loop
//				if !silentFired {
//					break
//				}
//			}
//		}
//		return error
//	}
func (wf *WorkflowLogic) FireTokenIdWithTransitionName(activityName string, caseId int) error {
	transitionIndices := wf.GetTransitionIndicesByName(activityName)
	allFailed := true
	for _, transitionIndex := range transitionIndices {
		err := wf.Petrinet.FireWithTokenId(transitionIndex, caseId)
		if err == nil {
			allFailed = false
			// Loop to handle recursive firing of silent transitions
			for {
				enabledTransitions := wf.Petrinet.GetEnabledTransitionsForTokenId(caseId)
				silentFired := false
				for _, t := range wf.SilentTransitions {
					for _, et := range enabledTransitions {
						if et == t {
							// Fire the silent transition
							wf.Petrinet.FireWithTokenId(t, caseId)
							silentFired = true
						}
					}
				}
				// If no silent transition was fired, break the loop
				if !silentFired {
					break
				}
			}
		}
	}
	if allFailed {
		return fmt.Errorf("Cannot fire any of the transition index %v", transitionIndices)
	} else {
		return nil
	}}

// Get next activities by their names
func (wf *WorkflowLogic) GetNextActivities() []string {
	nextActivities := []string{}
	for _, t := range wf.Petrinet.EnabledTransitions {
		nextActivities = append(nextActivities, wf.Transitions[t])
	}
	return nextActivities
}

// Given a token id, give me all the places that have that token id
func (wf *WorkflowLogic) GetPlacesWithTokenId(tokenId int) []string {
	places := []string{}
	for i, tokens := range wf.Petrinet.TokenIds {
		if petrinet.ContainsTokenId(tokens, tokenId) {
			places = append(places, wf.Places[i])
		}
	}
	return places
}

// Get the set of next transitions from a place
//func (wf *WorkflowLogic) GetNextTransitionsFromPlace(place string) []string {
//	placeIndex := -1
//	for i, p := range wf.Places {
//		if p == place {
//			placeIndex = i
//			break
//		}
//	}
//	if placeIndex == -1 {
//		return []string{}
//	}
//	transitions := []string{}
//	for i, step := range wf.Petrinet.InputMatrix {
//		if step[placeIndex] > 0 {
//			transitions = append(transitions, wf.Transitions[i])
//		}
//	}
//	return transitions
//}

// Get enabled tranision names for a token id
func (wf *WorkflowLogic) GetEnabledTransitionsForTokenId(tokenId int) []string {
	enabledTransitions := []string{}
	for _, t := range wf.Petrinet.GetEnabledTransitionsForTokenId(tokenId) {
		enabledTransitions = append(enabledTransitions, wf.Transitions[t])
	}
	return enabledTransitions
}

func (wf *WorkflowLogic) GetSourceAndSinkIndices() (int, int) {
    return 13, 12 // source index, sink index
}

