import os
import subprocess
import sys
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
    #TODO dovremmo aggiungere l'end event come una silent
    return net, im, fm,trans_names,final_silent_transitions




def generate_go_code_from_petri_net(places, transitions, input_matrix, output_matrix,trans_names,silent_transitions):
    transition_ids=[tr.name for tr in transitions]
    go_code = []
    # Define the net structure
    go_code.append("package workflowLogic\n")
    go_code.append("import \"main/utils/petrinet\"\n\n")


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
    go_code.append("var silentTransitionIndices = []int{\n")
    go_code.append(", ".join([str(transition_ids.index(t)) for t in silent_transitions])+",")
    go_code.append("\n}\n")
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

//func (wf *WorkflowLogic) FireTokenIdWithTransitionName(activityName string, caseId int) error {
//	transitionIndex := wf.GetTransitionIndexByName(activityName)
//	error := wf.Petrinet.FireWithTokenId(transitionIndex, caseId)
//	if error == nil {
//		//for each silent transition
//		for _, t := range wf.SilentTransitions {
//			//If the transition is enabled for the token
//			enabledTransitions := wf.Petrinet.GetEnabledTransitionsForTokenId(caseId)
//			//If the silent transition is enabled
//			for _, et := range enabledTransitions {
//				if et == t {
//					//Fire the silent transition
//					wf.Petrinet.FireWithTokenId(t, caseId)
//				}
//			}
//
//		}
//
//	}
//	return error
//}
func (wf *WorkflowLogic) FireTokenIdWithTransitionName(activityName string, caseId int) error {
	transitionIndex := wf.GetTransitionIndexByName(activityName)
	error := wf.Petrinet.FireWithTokenId(transitionIndex, caseId)
	if error == nil {
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






def main():
    if len(sys.argv) != 10:
        print("Usage: python processvaultcompiler.py <bpmn_file_path> <output_go_file_path>")
        sys.exit(1)

    bpmn_file_path = sys.argv[1]
    output_go_file_path = sys.argv[2]
    constraint_folder_path=sys.argv[3]
    output_go_file_path_compliance = sys.argv[4]
    event_dispatcher_address= sys.argv[5]
    extraction_manifest_file_path = sys.argv[6]
    isInSimulation = sys.argv[7]
    isInTesting = sys.argv[8]
    nEvents = sys.argv[9]

    control_flow_logic = generate_control_flow_logic(bpmn_file_path)
    compliance_checking_logic = generate_compliance_checking_logic(constraint_folder_path)


    # Write the Go code to the specified file
    with open(output_go_file_path, 'w') as go_file:
        go_file.write(control_flow_logic)

    # Set permissions to read and write for everyone (666)
    os.chmod(output_go_file_path, 0o666)


    if isInSimulation=="true":
        command="CGO_CFLAGS=-I/opt/ego/include CGO_LDFLAGS=-L/opt/ego/lib ego-go build -buildvcs=false main.go && ego sign main && OE_SIMULATION=1 ego run main "+event_dispatcher_address+" "+extraction_manifest_file_path +" true"+ " "+isInTesting+" "+nEvents
    else:
        command="CGO_CFLAGS=-I/opt/ego/include CGO_LDFLAGS=-L/opt/ego/lib ego-go build -buildvcs=false main.go && ego sign main && ego run main "+event_dispatcher_address + " "+extraction_manifest_file_path+" false"+ " "+isInTesting+" "+nEvents
    # Execute the build and run commands
    try:
        print("Building and running the Process Vault...")
        subprocess.run(
            command,
            shell=True,
            check=True
        )
        print("Process Vault successfully built and run")
    except subprocess.CalledProcessError as e:
        print(e)

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
    net, initial_marking, final_marking = pnml_importer.apply(pnml_file_path)
    trans_names = {t.name: t.label for t in net.transitions}
    silent_transitions = {t.name for t in net.transitions if t.label is None}
    return net, trans_names, silent_transitions

def generate_control_flow_logic(file_path):
    if file_path.endswith(".bpmn"):
        petrinet, trans_names, silent_transition = parse_bpmn_to_petri(file_path)
    elif file_path.endswith(".pnml"):
        petrinet, trans_names, silent_transition = parse_pnml_to_petri(file_path)
    else:
        raise ValueError("Unsupported file type. Please use a .bpmn or .pnml file.")

    places = sorted(petrinet.places, key=lambda place: place.name)
    transitions = sorted(petrinet.transitions, key=lambda transition: trans_names[transition.name])
    silent_transition = sorted(silent_transition)
    input_matrix, output_matrix = generate_input_output_matrices(places, transitions, petrinet.arcs)
    go_code = generate_go_code_from_petri_net(places, transitions, input_matrix, output_matrix, trans_names, silent_transition)
    return go_code
# def generate_compliance_checking_logic(contraints_file_path):
#     #For each file in the folder of the constraints_file_path
#     constaint_names=[]
#     constraints=[]
#     for file in sorted(os.listdir(contraints_file_path)):
#         if file.endswith(".rego"):
#             constaint_names.append(file.split(".")[0])
#             #Read the file
#             with open(contraints_file_path+"/"+file) as f:
#                 data_list = f.readlines()
#                 data_string = ''.join(data_list)
#                 constraints.append(data_string)
#
#     return generate_gocode_compliance(constaint_names, constraints)


# def generate_gocode_compliance(constaint_names, constraints):
#     go_code = []
#     # Define the net structure
#     go_code.append("package complianceCheckingLogic\n")
#     go_code.append("""
#     import (
# 	"context"
# 	"fmt"
# 	"github.com/open-policy-agent/opa/rego"
# 	"sync"
# 	"log")
#     """)
#     go_code.append("// Generated process constraints code\n")
#     # Declare the places
#     go_code.append("var constraintNames = []string{")
#     go_code.append(", ".join([f"\"{constr_name}\"" for constr_name in constaint_names]) + ",")
#     go_code.append("}\n")
#     go_code.append("var constraints = []string{\n")
#     for constr in constraints:
#         # Wrap each constraint in backticks for Go raw string
#         go_code.append(f"`{constr}`,\n")
#     go_code.append("}\n")
#     # Define the ComplianceCheckingLogic structure and InitComplianceCheckingLogic function
#     go_code.append("""
# // Enum for the state of the constraint
# type ConstraintState int
#
# const (
# 	Init     ConstraintState = 0
# 	Pending  ConstraintState = 1
# 	Violated ConstraintState = 2
# )
#
# type Constraint struct {
# 	name              string
# 	preparedEvalQuery rego.PreparedEvalQuery
# 	ConstraintState   map[string]ConstraintState // State for each case
# }
#
# type ComplianceCheckingLogic struct {
# 	preparedConstraints []Constraint
# 	ctx                 context.Context
# }
#
# // Function that creates a prepared constraint for each constraint
# func InitComplianceCheckingLogic() (ComplianceCheckingLogic, []string) {
# 	ctx := context.TODO()
# 	ccLogic := ComplianceCheckingLogic{
# 		preparedConstraints: []Constraint{},
# 		ctx:                 ctx,
# 	}
# 	for i, constraint := range constraints {
# 		query, err := rego.New(
# 			rego.Query("data."+constraintNames[i]),
# 			rego.Module(constraintNames[i], constraint),
# 		).PrepareForEval(ctx)
# 		if err != nil {
# 			log.Fatal(err)
# 		}
# 		ccLogic.preparedConstraints = append(ccLogic.preparedConstraints, Constraint{
# 			name:              constraintNames[i],
# 			preparedEvalQuery: query,
# 			ConstraintState:   make(map[string]ConstraintState),
# 		})
# 	}
# 	return ccLogic, constraintNames
# }
#
# // Evaluate the event log with the prepared constraints
# func (ccl *ComplianceCheckingLogic) EvaluateEventLog(eventLog map[string]interface{}) map[string]interface{} {
# 	//Get the last event from the event log
# 	lastEvent := eventLog["events"].([]map[string]interface{})[len(eventLog["events"].([]map[string]interface{}))-1]
# 	//Get the trace_id of the last event
# 	traceId := lastEvent["trace_concept_name"].(string)
# 	//If the trace_id is not in the constraint state, add it
# 	for _, constraint := range ccl.preparedConstraints {
# 		if _, ok := constraint.ConstraintState[traceId]; !ok {
# 			constraint.ConstraintState[traceId] = Init
# 		}
# 	}
# 	violationMap := map[string]interface{}{}
# 	var wg sync.WaitGroup
# 	var mu sync.Mutex
# 	for _, constraint := range ccl.preparedConstraints {
# 		wg.Add(1)
# 		go func(constraint Constraint) {
# 			defer wg.Done()
# 			res, err := constraint.preparedEvalQuery.Eval(ccl.ctx, rego.EvalInput(eventLog))
# 			if err != nil {
# 				fmt.Println(err)
# 				return
# 			}
# 			mu.Lock()
# 			defer mu.Unlock()
# 			for constraintName := range constraint.preparedEvalQuery.Modules() {
# 				resultValue := res[0].Expressions[0].Value
# 				resultValueMap, ok := resultValue.(map[string]interface{})
# 				if !ok {
# 					fmt.Println(res)
# 					log.Fatalf("Failed to convert result from policy inspection")
# 				}
# 				violations, ok := resultValueMap["violations"].(map[string]interface{})
# 				if !ok {
# 					fmt.Println(res)
# 					log.Fatalf("Failed to convert violation from policy inspection")
# 				}
# 				pending, ok := resultValueMap["pending"].(map[string]interface{})
# 				if !ok {
# 					fmt.Println(res)
# 					log.Fatalf("Failed to convert pending from policy inspection")
# 				}
# 				satisfied, ok := resultValueMap["satisfied"].(map[string]interface{})
# 				if !ok {
# 					fmt.Println(res)
# 					log.Fatalf("Failed to convert satisfied from policy inspection")
# 				}
# 				violationMap[constraintName] = map[string]interface{}{
# 					"violations": violations,
# 					"pending":    pending,
# 					"satisfied":  satisfied,
# 				}
# 				//serve una un set di caseId processati
# 				caseSet := map[string]bool{}
# 				for caseId := range pending {
# 					if constraint.ConstraintState[caseId] == Init {
# 						constraint.ConstraintState[caseId] = Pending
# 						fmt.Println("Constraint" + constraintName + " pending for case " + caseId)
# 					}
# 				}
# 				for caseId := range violations {
# 					if constraint.ConstraintState[caseId] == Pending {
# 						if !caseSet[caseId] {
# 							constraint.ConstraintState[caseId] = Violated
# 							caseSet[caseId] = true
# 							fmt.Println("Constraint" + constraintName + " violated for case " + caseId)
# 						}
# 					}
# 				}
# 				for caseId := range satisfied {
# 					if constraint.ConstraintState[caseId] == Pending {
# 						if !caseSet[caseId] {
# 							constraint.ConstraintState[caseId] = Init
# 							caseSet[caseId] = true
# 							fmt.Println("Constraint" + constraintName + " satisfied for case " + caseId)
# 						}
# 					}
# 				}
# 			}
# 		}(constraint)
# 	}
# 	wg.Wait()
# 	return violationMap
# }
#
# """)
#     return "\n".join(go_code)
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
        print(yaml_file_path)
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
func (ccl *ComplianceCheckingLogic) EvaluateEventLog(eventLog map[string]interface{}) map[string]interface{} {
	lastEvent := eventLog["events"].([]map[string]interface{})[len(eventLog["events"].([]map[string]interface{}))-1]
	traceId := lastEvent["trace_concept_name"].(string)
	for _, constraint := range ccl.preparedConstraints {
		if _, ok := constraint.ConstraintState[traceId]; !ok {
			constraint.ConstraintState[traceId] = Init // Init state
		}
	}
	violationMap := map[string]interface{}{}
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
			for constraintName := range constraint.preparedEvalQuery.Modules() {
				resultValue := res[0].Expressions[0].Value
				resultValueMap, ok := resultValue.(map[string]interface{})
				if !ok {
					fmt.Println(res)
					log.Fatalf("Failed to convert result from policy inspection")
				}
				//currentState := constraint.ConstraintState[traceId]
				//nextStates := constraint.fsm.PossibleNextStates(int(currentState))
				violations, ok := resultValueMap["violations"].(map[string]interface{})
				if !ok {
				}
				pending, ok := resultValueMap["pending"].(map[string]interface{})
				if !ok {
				}
				satisfied, ok := resultValueMap["satisfied"].(map[string]interface{})
				if !ok {
				}
				temporarySatisfied, ok := resultValueMap["temporary_satisfied"].(map[string]interface{})
				if !ok {
				}
				temporaryViolated, ok := resultValueMap["temporary_violated"].(map[string]interface{})
				if !ok {
				}
				for caseId := range pending {
					if constraint.fsm.HasTransition(int(constraint.ConstraintState[traceId]), 1) {
						constraint.ConstraintState[caseId] = Pending
						fmt.Println("Constraint ", constraintName, "in pending state for case ", caseId)

					}
				}
				for caseId := range temporarySatisfied {
					if constraint.fsm.HasTransition(int(constraint.ConstraintState[traceId]), 4) {
						constraint.ConstraintState[caseId] = TemporarySatisfied
						fmt.Println("Constraint ", constraintName, "in temporary satisfied state for case ", caseId)

					}
				}
				for caseId := range temporaryViolated {
					if constraint.fsm.HasTransition(int(constraint.ConstraintState[traceId]), 5) {
						constraint.ConstraintState[caseId] = TemporaryViolated
						fmt.Println("Constraint ", constraintName, "in temporary violated state for case ", caseId)

					}
				}
				for caseId := range violations {
					if constraint.fsm.HasTransition(int(constraint.ConstraintState[traceId]), 2) {
						constraint.ConstraintState[caseId] = Violated
						fmt.Println("Constraint ", constraintName, "in violated state for case ", caseId)

					}
				}
				for caseId := range satisfied {
					if constraint.fsm.HasTransition(int(constraint.ConstraintState[traceId]), 3) {
						constraint.ConstraintState[caseId] = Satisfied
						fmt.Println("Constraint ", constraintName, "in satisfied state for case ", caseId)
					}
				}
				violationMap[constraintName] = map[string]interface{}{
					"violations":          resultValueMap["violations"],
					"pending":             resultValueMap["pending"],
					"satisfied":           resultValueMap["satisfied"],
					"temporary_satisfied": resultValueMap["temporary_satisfied"],
					"temporary_violated":  resultValueMap["temporary_violated"],
				}
			}
		}(constraint)
	}
	wg.Wait()
	return violationMap
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

python3 pv3.py ./data/BPMN/motivating.bpmn ./workflowLogic/workflowLogic.go ./data/regoConstraints ./complianceCheckingLogic/complianceCheckingLogic.go localhost:6969 data/input/extraction_manifest_motivating.json true true 
python3 pv3.py ./data/BPMN/sepsis.bpmn ./workflowLogic/workflowLogic.go ./data/regoConstraints/sepsisConstraints ./complianceCheckingLogic/complianceCheckingLogic.go localhost:6969 data/input/extraction_manifest_sepsis.json true true
RUN IN NON SIMULATION MODE
python3 processvaultcompiler.py ./data/BPMN/motivating.bpmn ./workflowLogic/workflowLogic.go ./data/regoConstraints ./complianceCheckingLogic/complianceCheckingLogic.go localhost:6066 data/input/extraction_manifest_motivating.json false true 10

RUN SEPSIS TEST IN NON SIMULATION
python3 pv3.py ./data/BPMN/sepsis.bpmn ./workflowLogic/workflowLogic.go ./data/regoConstraints/sepsisConstraints ./complianceCheckingLogic/complianceCheckingLogic.go localhost:6066 data/input/extraction_manifest_sepsis.json false true 15200
UN MOTIVATING TEST IN NON SIMULATION
python3 pv3.py ./data/BPMN/motivating.bpmn ./workflowLogic/workflowLogic.go ./data/regoConstraints/motivatingConstraints ./complianceCheckingLogic/complianceCheckingLogic.go localhost:6066 data/input/extraction_manifest_motivating.json false true 35999
RUN TRAFFIC FINES TEST IN SIMULATION
python3 pv3.py ./data/BPMN/trafficFines.bpmn ./workflowLogic/workflowLogic.go ./data/regoConstraints/trafficFines ./complianceCheckingLogic/complianceCheckingLogic.go localhost:6066 data/input/extraction_manifest_traffic.json true false 10000
"""

def main():
    if len(sys.argv) != 10:
        print("Usage: python processvaultcompiler.py <bpmn_file_path> <output_go_file_path>")
        sys.exit(1)

    bpmn_file_path = sys.argv[1]
    output_go_file_path = sys.argv[2]
    constraint_folder_path=sys.argv[3]
    output_go_file_path_compliance = sys.argv[4]
    event_dispatcher_address= sys.argv[5]
    extraction_manifest_file_path = sys.argv[6]
    isInSimulation = sys.argv[7]
    isInTesting = sys.argv[8]
    nEvents = sys.argv[9]

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


    #TODO: we should add the extraciton manifest file differently to the process vault
    #Set permissions to read and write for everyone (666)
    os.chmod(output_go_file_path, 0o666)
    if isInSimulation=="true":
        command="CGO_CFLAGS=-I/opt/ego/include CGO_LDFLAGS=-L/opt/ego/lib ego-go build -buildvcs=false main.go && ego sign main && OE_SIMULATION=1 ego run main "+event_dispatcher_address+" "+extraction_manifest_file_path +" true"+ " "+isInTesting+" "+nEvents
    else:
        command="CGO_CFLAGS=-I/opt/ego/include CGO_LDFLAGS=-L/opt/ego/lib ego-go build -buildvcs=false main.go && ego sign main && ego run main "+event_dispatcher_address + " "+extraction_manifest_file_path+" false"+ " "+isInTesting+" "+nEvents
        # Execute the build and run commands
    try:
        print("Building and running the Process Vault...")
        subprocess.run(
            command,
            shell=True,
            check=True
        )
        print("Process Vault successfully built and run")
    except subprocess.CalledProcessError as e:
        print(e)
if __name__ == "__main__":
    main()
