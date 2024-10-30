package main

import "main/utils/petrinet"

// Generated Petri Net Code

var places = []string{
	"Flow_0mv7v6p", "Flow_1cwq7bo", "ent_Activity_049nscu", "ent_Activity_0cpl2j1", "ent_Activity_0d8y5we", "ent_Activity_176fx6v", "ent_Activity_1j9rd9s", "ent_Event_1lbqhp4", "ent_Gateway_1auib84", "ent_Gateway_1ovdhbi", "ent_Gateway_1rjzzr9", "exi_Gateway_1auib84", "sink", "source",
}

var transitions = []string{
	"Activity A", "Activity B", "Activity C", "Activity D", "Activity E", "Activity G", "Activty F", "EX D-->E,F", "EX E,F-->G", "End event", "Parallel A->B,C", "Parallel B,C->D", "Start event",
}

var inputMatrix = [][]int{

	{0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0},
	{0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0},
	{1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
}

var outputMatrix = [][]int{

	{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0},
	{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0},
	{0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0},
	{0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0},
}

var initialMarking = []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

// Indices of transitions associated with gateways

var silentTransitionIndices = []int{

	11, 7, 10, 8,
}

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
func (wf *WorkflowLogic) GetTransitionIndexByName(name string) int {
	for i, t := range wf.Transitions {
		if t == name {
			return i
		}
	}
	return -1
}

// Fire by transition name
func (wf *WorkflowLogic) FireByTransitionName(name string) error {
	transitionIndex := wf.GetTransitionIndexByName(name)
	error := wf.Petrinet.Fire(transitionIndex)
	return error
}

func (wf *WorkflowLogic) FireTokenIdWithTransitionName(activityName string, caseId int) error {
	transitionIndex := wf.GetTransitionIndexByName(activityName)
	error := wf.Petrinet.FireWithTokenId(transitionIndex, caseId)
	if error == nil {
		//for each silent transition
		for _, t := range wf.SilentTransitions {
			//If the transition is enabled for the token
			enabledTransitions := wf.Petrinet.GetEnabledTransitionsForTokenId(caseId)
			//If the silent transition is enabled
			for _, et := range enabledTransitions {
				if et == t {
					//Fire the silent transition
					wf.Petrinet.FireWithTokenId(t, caseId)
				}
			}

		}

	}
	return error
}

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
