import os
import subprocess
import sys
import threading

import yaml
import pm4py

import uuid
from enum import Enum

import json
from pm4py.objects.petri_net.utils import reduction
from pm4py.objects.petri_net.obj import PetriNet, Marking
from pm4py.objects.petri_net.utils.petri_utils import add_arc_from_to
from pm4py.util import exec_utils
from pm4py.objects.conversion.bpmn.variants.to_petri_net import build_digraph_from_petri_net
from pm4py.objects.petri_net.obj import PetriNet
from pm4py.objects.petri_net.importer.variants import pnml as pnml_importer

#TODO add silent transition names or unnamed acttivities' id to the set of names
#TODO then try to replace MessageFLow with SequenceFlow
class Parameters(Enum):
    USE_ID = "use_id"

def apply(bpmn_graph, parameters=None):
    if parameters is None:
        parameters = {}

    import networkx as nx
    from pm4py.objects.bpmn.obj import BPMN

    use_id = exec_utils.get_param_value(Parameters.USE_ID, parameters, True)


    net = PetriNet("")
    source_place = PetriNet.Place("source")
    net.places.add(source_place)
    sink_place = PetriNet.Place("sink")
    net.places.add(sink_place)
    im = Marking()
    fm = Marking()
    im[source_place] = 1
    fm[sink_place] = 1

    # keep this correspondence for adding invisible transitions for OR-gateways
    inclusive_gateway_exit = set()
    inclusive_gateway_entry = set()

    flow_place = {}
    source_count = {}
    target_count = {}

    for flow in bpmn_graph.get_flows():
        source = flow.get_source()
        target = flow.get_target()

        place = PetriNet.Place(str(flow.get_id()))
        net.places.add(place)
        flow_place[flow] = place
        if source not in source_count:
            source_count[source] = 0
        if target not in target_count:
            target_count[target] = 0
        source_count[source] = source_count[source] + 1
        target_count[target] = target_count[target] + 1

    for flow in bpmn_graph.get_flows():
        source = flow.get_source()
        target = flow.get_target()
        place = PetriNet.Place(str(flow.get_id()))
        if isinstance(source, BPMN.InclusiveGateway) and source_count[source] > 1:
            inclusive_gateway_exit.add(place.name)
        elif isinstance(target, BPMN.InclusiveGateway) and target_count[target] > 1:
            inclusive_gateway_entry.add(place.name)

    # remove possible places that are both in inclusive_gateway_exit and inclusive_gateway_entry,
    # because we do not need to add invisibles in this situation
    incl_gat_set_inters = inclusive_gateway_entry.intersection(inclusive_gateway_exit)
    inclusive_gateway_exit = inclusive_gateway_exit.difference(incl_gat_set_inters)
    inclusive_gateway_entry = inclusive_gateway_entry.difference(incl_gat_set_inters)

    nodes_entering = {}
    nodes_exiting = {}
    trans_names={}
    silent_transitions = set([])

    for node in bpmn_graph.get_nodes():
        if node.get_name()!="":
            trans_names[node.get_id()]=node.get_name()
        else:
            trans_names[node.get_id()]=node.get_id()
        entry_place = PetriNet.Place("ent_" + str(node.get_id()))
        net.places.add(entry_place)
        exiting_place = PetriNet.Place("exi_" + str(node.get_id()))
        net.places.add(exiting_place)
        if use_id:
            label = str(node.get_id())
        else:
            label = str(node.get_name()) if isinstance(node, BPMN.Task) else None
            if not label:
                label = None


        transition = PetriNet.Transition(name=str(node.get_id()), label=label)
        net.transitions.add(transition)
        add_arc_from_to(entry_place, transition, net)
        add_arc_from_to(transition, exiting_place, net)
        if isinstance(node, BPMN.Gateway):
            silent_transitions.add(node.get_id())

        if isinstance(node, BPMN.ParallelGateway) or isinstance(node, BPMN.InclusiveGateway):
            if source_count[node] > 1:
                exiting_object = PetriNet.Transition(str(uuid.uuid4()), None)
                net.transitions.add(exiting_object)
                silent_transitions.add(exiting_object.name)
                trans_names[exiting_object.name]=exiting_object.name
                add_arc_from_to(exiting_place, exiting_object, net)
            else:
                exiting_object = exiting_place

            if target_count[node] > 1:
                entering_object = PetriNet.Transition(str(uuid.uuid4()), None)
                net.transitions.add(entering_object)
                silent_transitions.add(entering_object.name)
                trans_names[entering_object.name]=entering_object.name
                add_arc_from_to(entering_object, entry_place, net)
            else:
                entering_object = entry_place
            nodes_entering[node] = entering_object
            nodes_exiting[node] = exiting_object
        else:
            nodes_entering[node] = entry_place
            nodes_exiting[node] = exiting_place

        if isinstance(node, BPMN.StartEvent):
            start_transition = PetriNet.Transition(str(uuid.uuid4()), None)
            net.transitions.add(start_transition)
            add_arc_from_to(source_place, start_transition, net)
            add_arc_from_to(start_transition, entry_place, net)
        elif isinstance(node, BPMN.EndEvent):
            end_transition = PetriNet.Transition(str(uuid.uuid4()), None)
            net.transitions.add(end_transition)
            add_arc_from_to(exiting_place, end_transition, net)
            add_arc_from_to(end_transition, sink_place, net)

    for flow in bpmn_graph.get_flows():
        source_object = nodes_exiting[flow.get_source()]
        target_object = nodes_entering[flow.get_target()]

        if isinstance(source_object, PetriNet.Place):
            inv1 = PetriNet.Transition(str(uuid.uuid4()), None)
            net.transitions.add(inv1)
            silent_transitions.add(inv1.name)
            trans_names[inv1.name]=inv1.name
            add_arc_from_to(source_object, inv1, net)
            source_object = inv1

        if isinstance(target_object, PetriNet.Place):
            inv2 = PetriNet.Transition(str(uuid.uuid4()), None)
            net.transitions.add(inv2)
            silent_transitions.add(inv2.name)
            trans_names[inv2.name]=inv2.name
            add_arc_from_to(inv2, target_object, net)
            target_object = inv2

        add_arc_from_to(source_object, flow_place[flow], net)
        add_arc_from_to(flow_place[flow], target_object, net)

    if inclusive_gateway_exit and inclusive_gateway_entry:
        # do the following steps if there are inclusive gateways:
        # - calculate the shortest paths
        # - add an invisible transition between couples of corresponding places
        # this ensures soundness and the correct translation of the BPMN
        inv_places = {x.name: x for x in net.places}
        digraph = build_digraph_from_petri_net(net)
        all_shortest_paths = dict(nx.all_pairs_dijkstra(digraph))
        keys = list(all_shortest_paths.keys())
        for pl1 in inclusive_gateway_exit:
            if pl1 in keys:
                output_places = sorted(
                    [(x, len(y)) for x, y in all_shortest_paths[pl1][1].items() if x in inclusive_gateway_entry],
                    key=lambda x: x[1])
                if output_places:
                    inv_trans = PetriNet.Transition(str(uuid.uuid4()), None)
                    net.transitions.add(inv_trans)
                    silent_transitions.add(inv_trans.name)
                    trans_names[inv_trans.name]=inv_trans.name
                    add_arc_from_to(inv_places[pl1], inv_trans, net)
                    add_arc_from_to(inv_trans, inv_places[output_places[0][0]], net)
    net=reduction.apply_simple_reduction(net)
    final_silent_transitions = set([])
    #Generate the set of all the transition names in net
    all_tr=set([x.name for x in net.transitions])
    for silent_transition in silent_transitions:
        if silent_transition in all_tr:
            final_silent_transitions.add(silent_transition)
    return net, im, fm,trans_names,final_silent_transitions




def generate_go_code_from_petri_net(places, transitions, input_matrix, output_matrix,trans_names,silent_transitions):
    transition_ids=[tr.name for tr in transitions]
    go_code = []
    # Define the net structure
    go_code.append("package workflowLogic\n")
    go_code.append("import \"main/utils/petrinet\"\n\n")
    go_code.append("import \"fmt\"\n\n")

    go_code.append("// Generated Petri Net Code\n")

    # Declare the places
    go_code.append("var places = []string{")
    go_code.append(", ".join([f"\"{place}\"" for place in places]) + ",")
    go_code.append("}\n")

    # Declare the transitions
    go_code.append("var transitions = []string{")
    go_code.append(", ".join([f"\"{trans_names[transition.name]}\"" for transition in transitions]) + ",")
    go_code.append("}\n")

    # Input matrix declaration
    go_code.append("var inputMatrix = [][]int{\n")
    for row in input_matrix:
        row_str = ", ".join(map(str, row))
        go_code.append(f"    {{{row_str}}},")
    go_code.append("}\n")

    # Output matrix declaration
    go_code.append("var outputMatrix = [][]int{\n")
    for row in output_matrix:
        row_str = ", ".join(map(str, row))
        go_code.append(f"    {{{row_str}}},")
    go_code.append("}\n")

    # Declare the initial marking (initially all zeros, can be modified as needed)
    initial_marking = [0] * len(places)
    go_code.append(f"var initialMarking = []int{{{', '.join(map(str, initial_marking))}}}\n")
    # Add the list of transitions associated with gates
    go_code.append("// Indices of transitions associated with gateways\n")
    if len(silent_transitions) > 0:
        go_code.append("var silentTransitionIndices = []int{\n")
        go_code.append(", ".join([str(transition_ids.index(t)) for t in silent_transitions])+",")
        go_code.append("\n}\n")
    else:
        go_code.append("var silentTransitionIndices = []int{}\n")
    # Define the WorkflowLogic structure and InitWorkflowLogic function
    go_code.append("""
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
""")
    # Get index of source and sink places
    source_index = [pl.name for pl in places].index("source")
    sink_index = [pl.name for pl in places].index("sink")
    # Generate Go code for source and sink index function
    go_code.append("func (wf *WorkflowLogic) GetSourceAndSinkIndices() (int, int) {")
    go_code.append(f"    return {source_index}, {sink_index} // source index, sink index")
    go_code.append("}\n\n")

    # Return the final Go code as a string
    return "\n".join(go_code)

def parse_bpmn_to_petri(bpmn_file_path):
    bpmn_model = pm4py.read_bpmn(bpmn_file_path)
    petrinet_conversion = apply(bpmn_model)
    return petrinet_conversion[0], petrinet_conversion[3],petrinet_conversion[4]
def generate_input_output_matrices(places, transitions, arcs):
    place_to_index = {place.name: idx for idx, place in enumerate(places)}
    transition_to_index = {transition.name: idx for idx, transition in enumerate(transitions)}
    input_matrix = [[0] * len(places) for _ in range(len(transitions))]
    output_matrix = [[0] * len(places) for _ in range(len(transitions))]

    for arc in arcs:
        source = arc.source
        target = arc.target
        if isinstance(source, PetriNet.Place) and isinstance(target, PetriNet.Transition):
            input_matrix[transition_to_index[target.name]][place_to_index[source.name]] = 1
        elif isinstance(source, PetriNet.Transition) and isinstance(target, PetriNet.Place):
            output_matrix[transition_to_index[source.name]][place_to_index[target.name]] = 1

    return input_matrix, output_matrix




#TODO: this is the right one
# def generate_control_flow_logic(bpmn_file_path):
#     petrinet, trans_names, silent_transition = parse_bpmn_to_petri(bpmn_file_path)
#     places = sorted(petrinet.places, key=lambda place: place.name)
#     transitions = sorted(petrinet.transitions, key=lambda transition: trans_names[transition.name])
#     silent_transition = sorted(silent_transition)
#     input_matrix, output_matrix = generate_input_output_matrices(places, transitions, petrinet.arcs)
#     go_code = generate_go_code_from_petri_net(places, transitions, input_matrix, output_matrix, trans_names,
#                                              silent_transition)
#     return go_code
def parse_pnml_to_petri(pnml_file_path):
    net, initial_marking, final_marking = pnml_importer.import_net(pnml_file_path)
    trans_names = {t.name: (t.label if t.label is not None else t.name) for t in net.transitions}
    silent_transitions = {t.name for t in net.transitions if t.label is None}
    return net, trans_names, silent_transitions
def generate_control_flow_logic(file_path):
    if file_path.endswith(".bpmn"):
        petrinet, trans_names, silent_transition = parse_bpmn_to_petri(file_path)
    elif file_path.endswith(".pnml"):
        petrinet, trans_names, silent_transition = parse_pnml_to_petri(file_path)
        #find the index of the input place and the output place
        #the input place is a place that in petrinet has no input transition
        #the output place is a place that in petrinet has no output transition
        input_place_index=-1
        output_place_index=-1
        for i, place in enumerate(petrinet.places):
            if len(place.in_arcs)==0:
                input_place_index=i
                place.name="source"
            if len(place.out_arcs)==0:
                output_place_index=i
                place.name="sink"
        if input_place_index==-1 or output_place_index==-1:
            raise ValueError("The input and output place are not found")
    else:
        raise ValueError("Unsupported file type. Please use a .bpmn or .pnml file.")
    places = sorted(petrinet.places, key=lambda place: place.name)
    transitions = sorted(petrinet.transitions, key=lambda transition: trans_names[transition.name])
    silent_transition = sorted(silent_transition)
    input_matrix, output_matrix = generate_input_output_matrices(places, transitions, petrinet.arcs)
    go_code = generate_go_code_from_petri_net(places, transitions, input_matrix, output_matrix, trans_names, silent_transition)
    return go_code

def parse_yaml_file(file_path):
    with open(file_path, 'r') as file:
        return yaml.safe_load(file)

def generate_fsm_code(constraint_name, fsm_data):
    states = fsm_data['states']
    transitions = fsm_data['transitions']
    initial_state = fsm_data['initial_state']

    fsm_code = f"""
case "{constraint_name}":
    finitestate := fsm.NewFSM(
        "{initial_state}",
        fsm.Events{{"""
    for transition in transitions:
        fsm_code += f"""
            {{Name: "{transition['event']}", Src: []string{{"{transition['from']}"}}, Dst: "{transition['to']}"}},"""
    fsm_code += """
        },
        fsm.Callbacks{},
    )
    return finitestate
"""
    return fsm_code

def generate_default_fsm_code():
    return """
case "default":
    finitestate := fsm.NewFSM(
        "Init",
        fsm.Events{
            {Name: "Pending", Src: []string{"Init"}, Dst: "Pending"},
            {Name: "Violated", Src: []string{"Pending"}, Dst: "Violated"},
            {Name: "Satisfied", Src: []string{"Pending"}, Dst: "Init"},
        },
        fsm.Callbacks{},
    )
    return finitestate
"""

def generate_get_fsm_function(constraint_folder_path, constraint_names):
    fsm_code = """
func getFSMfromConstraintName(constraintName string) *fsm.FSM {
    switch constraintName {"""
    for constraint_name in constraint_names:
        yaml_file_path = os.path.join(constraint_folder_path, f"{constraint_name}.yaml")
        if os.path.exists(yaml_file_path):
            fsm_data = parse_yaml_file(yaml_file_path)
            fsm_code += generate_fsm_code(constraint_name, fsm_data)
        else:
            fsm_code += generate_default_fsm_code().replace("default", constraint_name)
    fsm_code += """
    }
    return &fsm.FSM{}
}
"""
    return fsm_code

def generate_compliance_checking_logic(constraint_folder_path):
    constraint_names = []
    constraints = []
    for file in sorted(os.listdir(constraint_folder_path)):
        if file.endswith(".rego"):
            constraint_name = file.split(".")[0]
            constraint_names.append(constraint_name)
            with open(os.path.join(constraint_folder_path, file)) as f:
                constraints.append(f.read())

    go_code = []
    go_code.append("package complianceCheckingLogic\n")
    go_code.append("""
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
""")
    go_code.append("// Generated process constraints code\n")
    go_code.append("var constraintNames = []string{")
    go_code.append(", ".join([f"\"{name}\"" for name in constraint_names]) + "}\n")
    go_code.append("var constraints = []string{\n")
    for constraint in constraints:
        go_code.append(f"`{constraint}`,\n")
    go_code.append("}\n")

    go_code.append("""
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
}

// Function that creates a prepared constraint for each constraint
func InitComplianceCheckingLogic() (ComplianceCheckingLogic, []string) {
	ctx := context.TODO()
	ccLogic := ComplianceCheckingLogic{
		preparedConstraints: []Constraint{},
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
			res, err := constraint.preparedEvalQuery.Eval(context.Background(), rego.EvalInput(eventLog))
			if err != nil {
				fmt.Println(err)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			for {
				transitionFound := false
				currentState := constraint.ConstraintState[traceId]
				for _, nextState := range constraint.fsm.PossibleNextStates(int(currentState)) {
					currentState := constraint.ConstraintState[traceId]
					ruleName := fmt.Sprintf("%sTo%s", stateName(currentState), stateName(ConstraintState(nextState)))
					if resultValue, ok := res[0].Expressions[0].Value.(map[string]interface{})[ruleName]; ok {
						if resultValueMap, ok := resultValue.(map[string]interface{}); ok {
							for caseId, isTrue := range resultValueMap {
								if isTrue.(bool) {
									constraint.ConstraintState[caseId] = ConstraintState(nextState)
									//fmt.Printf("Constraint %s transitioned from %s to %s for case %s", constraint.name, stateName(currentState), stateName(ConstraintState(nextState)), caseId)
									//fmt.Println()
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
""")
    go_code.append(generate_get_fsm_function(constraint_folder_path, constraint_names))
    return "\n".join(go_code)

statesMap = {"Init": 0, "Pending": 1, "Violated": 2, "Satisfied": 3, "TemporarySatisfied": 4, "TemporaryViolated": 5}
def generate_fsm_code(constraint_name, fsm_data):
    transitions = fsm_data['transitions']
    trList = [[],[],[],[],[],[]]
    for t in transitions:
        trList[statesMap[t['from']]].append(str(statesMap[t['to']]))
    fsm_code = f"""
"{constraint_name}": {{
    Transitions: [][]int{{"""
    for outputStates in trList:
        fsm_code += f"""
        {{{", ".join(outputStates)}}},"""
    fsm_code += """
    },
},
"""

    return fsm_code

def generate_get_fsm_function(constraint_folder_path, constraint_names):
    fsm_code = """
var fsmMap = map[string]*CustomFSM{
"""
    for constraint_name in constraint_names:
        yaml_file_path = os.path.join(constraint_folder_path, f"{constraint_name}.yaml")
        if os.path.exists(yaml_file_path):
            fsm_data = parse_yaml_file(yaml_file_path)
            fsm_code += generate_fsm_code(constraint_name, fsm_data)
        else:
            fsm_code += f"""
"{constraint_name}": {{
    Transitions: [][]int{{
        {{1}},    // init -> pending
        {{2, 3}}, // pending -> violated, satisfied
        {{}},     // violated -> none
        {{0}},    // satisfied -> init
        {{}},     // temporary_satisfied -> none
        {{}},     // temporary_violated -> none
    }},
}},
"""
    fsm_code += """
}
"""
    return fsm_code
"""

RUN IN NON SIMULATION MODE
python3 pv4.py ./data/BPMN/newmotivating.pnml ./workflowLogic/workflowLogic.go ./data/regoConstraints/motivatingConstraints ./complianceCheckingLogic/complianceCheckingLogic.go localhost:6969 data/input/extraction_manifest_motivating.json true true
RUN SEPSIS TEST IN NON SIMULATION
python3 pv4.py ./data/BPMN/sepsis.bpmn ./workflowLogic/workflowLogic.go ./data/regoConstraints/sepsisConstraints ./complianceCheckingLogic/complianceCheckingLogic.go localhost:6066 data/input/extraction_manifest_sepsis.json false true 15200 false
RUN MOTIVATING TEST IN NON SIMULATION
python3 pv4.py ./data/PNML/motivatingreduced.pnml ./workflowLogic/workflowLogic.go ./data/regoConstraints/motivatingConstraints ./complianceCheckingLogic/complianceCheckingLogic.go localhost:6066 data/input/extraction_manifest_motivating.json false true 40000 false
RUN TRAFFIC FINES TEST IN SIMULATION
python3 pv4.py ./data/PNML/trafficFinesrevised.pnml ./workflowLogic/workflowLogic.go ./data/regoConstraints/trafficFines ./complianceCheckingLogic/complianceCheckingLogic.go localhost:6066 data/input/extraction_manifest_traffic.json false true 561400 false 150
RUN BPIC2012 TEST IN SIMULATION
python3 pv4.py ./data/PNML/bpic2012top10alpha.pnml ./workflowLogic/workflowLogic.go ./data/regoConstraints/bpic2012constraints ./complianceCheckingLogic/complianceCheckingLogic.go localhost:6066 data/input/extraction_manifest_bpic2012.json true true 561400 false 150
"""

import subprocess
import sys
import os
import threading
import time
import csv
import psutil

def find_ego_host_pid(parent_pid):
    """Find the child process (ego-host) of the given parent PID."""
    for _ in range(20):  # Retry for ~2 seconds (20 * 100ms)
        try:
            parent = psutil.Process(parent_pid)
            children = parent.children(recursive=True)  # Get child processes
            for child in children:
                if "ego-host" in child.name():
                    return child.pid
        except psutil.NoSuchProcess:
            pass
        time.sleep(0.1)  # Wait 100ms before retrying
    return None  # If not found within 2 seconds

def monitor_process(pid, log_file):
    """Monitor the CPU and memory usage of a given process every 5ms."""
    with open(log_file, 'w', newline='') as csvfile:
        csv_writer = csv.writer(csvfile)
        csv_writer.writerow(["Timestamp_ms", "CPU_Usage", "Memory_MB"])  # Removed PID column
        try:
            process = psutil.Process(pid)
            while process.is_running():
                timestamp_ms = int(time.time() * 1000)  # Unix timestamp in milliseconds
                cpu_usage = process.cpu_percent(interval=0.005)  # 5ms
                memory_usage = process.memory_info().rss / (1024 * 1024)  # Convert to MB
                csv_writer.writerow([timestamp_ms, cpu_usage, round(memory_usage, 2)])
                csvfile.flush()
        except psutil.NoSuchProcess:
            print("Process ended, stopping monitoring.")

def run_process_with_output(command):
    """Run a process and print its output in real-time."""
    process = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)

    # Function to read stdout
    def read_stdout():
        for line in process.stdout:
            sys.stdout.write(line)
            sys.stdout.flush()

    # Function to read stderr
    def read_stderr():
        for line in process.stderr:
            sys.stderr.write(line)
            sys.stderr.flush()

    # Start threads to print stdout and stderr in real-time
    stdout_thread = threading.Thread(target=read_stdout)
    stderr_thread = threading.Thread(target=read_stderr)
    stdout_thread.start()
    stderr_thread.start()

    return process, stdout_thread, stderr_thread

def main():
    if len(sys.argv) != 12:
        print("Usage: python processvaultcompiler.py <bpmn_file_path> <output_go_file_path>")
        sys.exit(1)

    bpmn_file_path = sys.argv[1]
    output_go_file_path = sys.argv[2]
    constraint_folder_path = sys.argv[3]
    output_go_file_path_compliance = sys.argv[4]
    event_dispatcher_address = sys.argv[5]
    extraction_manifest_file_path = sys.argv[6]
    isInSimulation = sys.argv[7]
    isInTesting = sys.argv[8]
    nEvents = sys.argv[9]
    withExternalQueue = sys.argv[10]
    slidingWindowSize = sys.argv[11]

    if withExternalQueue == "true":
        def run_queue_server():
            os.system("go run ./queue/queue.go localhost:8387")
        queue_thread = threading.Thread(target=run_queue_server)
        queue_thread.start()
        print("External queue running at localhost:8387")
    control_flow_logic = generate_control_flow_logic(bpmn_file_path)
    compliance_checking_logic = generate_compliance_checking_logic(constraint_folder_path)

    # Write the Go code to the specified file
    with open(output_go_file_path_compliance, 'w') as go_file:
        go_file.write(compliance_checking_logic)

    # Set permissions to read and write for everyone (666)
    os.chmod(output_go_file_path_compliance, 0o666)


    # Write the Go code to the specified file
    with open(output_go_file_path, 'w') as go_file:
        go_file.write(control_flow_logic)

    print("Building the Process Vault...")
    build_command = "CGO_CFLAGS=-I/opt/ego/include CGO_LDFLAGS=-L/opt/ego/lib ego-go build -buildvcs=false main.go"
    sign_command = "ego sign main"
    try:
        subprocess.run(build_command, shell=True, check=True)
        subprocess.run(sign_command, shell=True, check=True)
    except subprocess.CalledProcessError as e:
        print("Build or sign command failed:", e)
        sys.exit(1)

    if isInSimulation == "true":
        print("Running the Process Vault in simulation mode")
        run_command = f"OE_SIMULATION=1 ego run main {event_dispatcher_address} {extraction_manifest_file_path} "
    else:
        print("Running the Process Vault in hardware mode")
        run_command = f"ego run main {event_dispatcher_address} {extraction_manifest_file_path} "

    run_command += "true" if isInSimulation == "true" else "false"
    run_command += f" {isInTesting} {nEvents} {withExternalQueue} {slidingWindowSize}"

    try:
        # Start the Process Vault and capture its output in real-time
        process, stdout_thread, stderr_thread = run_process_with_output(run_command)

        print(f"RUN_PID:{process.pid}", flush=True)  # Print PID for debugging

        # Wait for ego-host process
        ego_host_pid = find_ego_host_pid(process.pid)
        if not ego_host_pid:
            print("Error: ego-host process not found.")
            sys.exit(1)

        print(f"Found ego-host PID: {ego_host_pid}")

        # Start monitoring only if isInTesting == "true"
        if isInTesting.lower() == "true":
            print("Starting monitoring...")
            log_file = "./data/output/monitoring_result.csv"
            monitor_thread = threading.Thread(target=monitor_process, args=(ego_host_pid, log_file))
            monitor_thread.start()
        else:
            print("Skipping monitoring (isInTesting is false)")

        # Wait for the main process and ensure output is printed
        process.wait()
        stdout_thread.join()
        stderr_thread.join()

        # Ensure monitoring stops when process ends
        if isInTesting.lower() == "true":
            monitor_thread.join()

    except subprocess.CalledProcessError as e:
        print("Run command failed:", e)
        sys.exit(1)

if __name__ == "__main__":
    main()

