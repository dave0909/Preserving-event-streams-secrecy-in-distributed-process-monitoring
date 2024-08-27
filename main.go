package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"github.com/veith/petrinet"
	"main/utils/pnml"
	netPackage "net"
	"os"
	"testing"
)

func makeExampleNet() petrinet.Net {
	f := petrinet.Net{
		//rows are transitions, columns are places. If 1, there is an input arc from place to transition
		InputMatrix: [][]int{
			// 0 -->
			{1, 0, 0},
			{0, 1, 0},
		},
		OutputMatrix: [][]int{
			{0, 1, 0},
			{0, 0, 1},
		},
		// ConditionMatrix: [][]string{{"true"}, {}, {}, {}, {}, {}, {}, {}},
		State:     []int{1, 0, 0},
		Variables: map[string]interface{}{"a": 4, "b": 8},
	}
	return f
}

func TestPetriNet_FireBadCondition(t *testing.T) {
	flow := makeExampleNet()
	flow.Variables = map[string]interface{}{"a": 2, "b": 8}
	flow.ConditionMatrix = [][]string{{"a > 11", "b == 8", "a != b", "true"}, {}, {}, {}, {}, {}, {}, {}}

	flow.Init()

	err := flow.Fire(0)

	if err == nil {
		t.Error("should not fire")
	}

	flow.UpdateVariable("a", 12)

	errr := flow.Fire(0)

	if errr != nil {
		t.Error("should have no error ")
	}
}

func TestPetriNet_WithoutConditionsMatrix(t *testing.T) {
	flow := makeExampleNet()

	flow.Init()

	err := flow.Fire(0)

	if err != nil {
		t.Error(err)
	}
	// example should have 2 possible transitions
	if len(flow.EnabledTransitions) != 2 {
		t.Error("Expected 2, got ", len(flow.EnabledTransitions))
	}

}

func BenchmarkNet_Fire(b *testing.B) {
	flow := makeExampleNet()
	flow.ConditionMatrix = [][]string{{"a > 11", "b == 8", "a != b", "true"}, {}, {}, {}, {}, {}, {}, {}}
	flow.OutputMatrix[0][0] = 1
	flow.Init()

	for i := 0; i < b.N; i++ {
		flow.Fire(0)
	}
}
func BenchmarkNet_FireWithoutConditions(b *testing.B) {
	flow := makeExampleNet()
	flow.OutputMatrix[0][0] = 1
	flow.Init()

	for i := 0; i < b.N; i++ {
		flow.Fire(0)
	}
}

func TestPetriNet_Fire(t *testing.T) {
	flow := makeExampleNet()
	flow.ConditionMatrix = [][]string{{"true"}, {}, {}, {}, {}, {}, {}, {}}
	flow.Init()
	err := flow.Fire(0)

	if err != nil {
		t.Error(err)
	}
	// example should have 2 possible transitions
	if len(flow.EnabledTransitions) != 2 {
		t.Error("Expected 2, got ", len(flow.EnabledTransitions))
	}

}

func TestPetriNet_TokenID(t *testing.T) {
	flow := makeExampleNet()
	flow.State = []int{3, 0, 0, 0, 0, 0, 0, 0, 0}
	flow.Init()
	if flow.TokenIds[0][0] != 1 {
		t.Error("token id counter wrong ", flow.TokenIds)
	}
	fmt.Println(flow.State)
	err := flow.Fire(0)
	err = flow.Fire(2)
	err = flow.Fire(1)
	err = flow.Fire(3)
	err = flow.Fire(4)
	flow.Fire(0)
	flow.Fire(0)
	err = flow.Fire(2)
	err = flow.Fire(1)
	err = flow.Fire(3)
	err = flow.Fire(4)
	err = flow.Fire(2)
	err = flow.Fire(1)
	err = flow.Fire(3)
	err = flow.Fire(4)

	if flow.TokenIds[6][0] != 9 {
		t.Error("token id counter wrong ", flow.TokenIds)
	}

	if len(flow.TokenIds[0]) != 0 {
		t.Error("token id counter wrong should be empty haves", len(flow.TokenIds[0]))
	}

	if err != nil {
		t.Error(err)
	}

	flow.FireWithTokenId(5, 17)

	if flow.TokenIds[6][0] != 9 {
		t.Error("token id counter wrong ", flow.TokenIds)
	}

	if flow.TokenIds[7][0] != 22 {
		t.Error("token id counter wrong ", flow.TokenIds)
	}

	err = flow.FireWithTokenId(5, 2)

	if err == nil {
		t.Error("should return error tokenid not found")
	}
	err = flow.FireWithTokenId(5, 21)

	if flow.TokenIds[6][0] != 9 {
		t.Error("token id counter wrong ", flow.TokenIds)
	}
	flow.Fire(6)
	flow.Fire(6)
	flow.Fire(5)
	flow.Fire(6)

	if len(flow.EnabledTransitions) != 0 {
		t.Error("should have no enabled transitions, got ", len(flow.EnabledTransitions))
	}

}

func main() {
	flow := makeExampleNet()
	flow.ConditionMatrix = [][]string{{"true"}, {}, {}, {}, {}, {}, {}, {}}
	flow.Init()
	//parse a pnmlObject file using the package PNML
	//read a file in a byte array
	file, err := os.ReadFile("sample_pnml.pnml")
	if err != nil {
		fmt.Println(err)
		return
	}
	//unmarshal the file into a PNML struct
	pnmlObject := pnml.PNML{}
	err = xml.Unmarshal(file, &pnmlObject)
	if err != nil {
		fmt.Println(err)
		return
	}
	pnml.RemovePages(pnmlObject)
	//print the PNML transitions
	for _, net := range pnmlObject.Nets {
		fmt.Println(net.Transitions)
		for _, transition := range net.Transitions {
			fmt.Println("--------------------", transition.Name.Text)
		}
	}
	//print the PNML places
	for _, net := range pnmlObject.Nets {
		for _, place := range net.Places {
			fmt.Println(place.Name.Text)
		}
	}
	//print the PNML arcs
	for _, net := range pnmlObject.Nets {
		for _, arc := range net.Arcs {
			fmt.Println(arc.Source, arc.Target)
		}

	}
	net := pnmlObject.Nets[0]
	transitionIndex, input, output := net.GetInputOutputMatrices()
	prova := petrinet.Net{
		//rows are transitions, columns are places. If 1, there is an input arc from place to transition
		InputMatrix: input,

		OutputMatrix: output,
		State:        []int{2, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		// ConditionMatrix: [][]string{{"true"}, {}, {}, {}, {}, {}, {}, {}},
		Variables: map[string]interface{}{"a": 4, "b": 8},
	}
	//map transitions to their index
	prova.Init()
	prova.Fire(transitionIndex["Dispatch order"])
	fmt.Println(prova.State)
	prova.Fire(transitionIndex["Dispatch order"])
	fmt.Println(prova.State)

	// Connect to the server on localhost at port 1234
	conn, err := netPackage.Dial("tcp", "localhost:8085")
	if err != nil {
		fmt.Println("Error connecting:", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("Connected to localhost:8085")

	for {
		// Receive a response from the server
		response, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from server:", err)
			return
		}

		// Print the response
		fmt.Print("RafTEEserver response: ", response)
	}

}

func handleConnection(conn netPackage.Conn) {
	defer conn.Close()
	// Create a buffer reader to read the incoming data
	reader := bufio.NewReader(conn)
	for {
		// Read data until newline
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading:", err)
			return
		}
		// Print the received message
		fmt.Print("Message received:", message)
	}
}
