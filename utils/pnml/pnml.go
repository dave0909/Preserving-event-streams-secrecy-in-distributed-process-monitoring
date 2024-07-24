//package pnml
//
//import (
//	"encoding/xml"
//	"errors"
//	"fmt"
//	"strconv"
//	"strings"
//)
//
//type PNML struct {
//	XMLName xml.Name `xml:"pnml"`
//	Nets    []Net    `xml:"net"`
//}
//
//type Net struct {
//	XMLName xml.Name `xml:"net"`
//	ID      string   `xml:"id,attr"`
//	//Name        string       `xml:"name"`
//	Places      []Place      `xml:"place"`
//	Transitions []Transition `xml:"transition"`
//	Arcs        []Arc        `xml:"arc"`
//}
//
//type Place struct {
//	XMLName xml.Name `xml:"place"`
//	ID      string   `xml:"id,attr"`
//	Name    Name     `xml:"name"`
//}
//
//type Transition struct {
//	XMLName xml.Name `xml:"transition"`
//	ID      string   `xml:"id,attr"`
//	Name    Name     `xml:"name"`
//}
//
//type Arc struct {
//	XMLName xml.Name `xml:"arc"`
//	ID      string   `xml:"id,attr"`
//	Source  string   `xml:"source,attr"`
//	Target  string   `xml:"target,attr"`
//}
//
//type Name struct {
//	Text string `xml:"text"`
//}
//
//// Create an initial Petri net with some places, transitions, and arcs
//func CreateInitialPetriNet() PNML {
//	net := Net{
//		ID: "net1",
//		//Name:        "PetriNet1",
//		Places:      []Place{},
//		Transitions: []Transition{},
//		Arcs:        []Arc{},
//	}
//
//	pnml := PNML{
//		Nets: []Net{net},
//	}
//
//	return pnml
//}
//
//// Create a new place with the given ID and name
//func CreatePlace(id, name string) Place {
//	return Place{
//		ID:   id,
//		Name: Name{Text: name},
//	}
//}
//
//// Create a new transition with the given ID and name
//func CreateTransition(id, name string) Transition {
//	return Transition{
//		ID:   id,
//		Name: Name{Text: name},
//	}
//}
//
//// Create a new arc with the given ID, source, and target
//func CreateArc(id, source, target string) Arc {
//	return Arc{
//		ID:     id,
//		Source: source,
//		Target: target,
//	}
//}
//
//// Add a new transition to the Petri net
//func AddTransition(pnml *PNML, id, name string) {
//	net := &pnml.Nets[0] // Assuming only one net is present
//	transition := CreateTransition(id, name)
//	net.Transitions = append(net.Transitions, transition)
//}
//
//// Add a new element (place) to the Petri net
//func AddPlace(pnml *PNML, id, name string) {
//	net := &pnml.Nets[0] // Assuming only one net is present
//	place := CreatePlace(id, name)
//	net.Places = append(net.Places, place)
//}
//
//// Add a new arc to the Petri net
//func AddArc(pnml *PNML, id, source, target string) {
//	net := &pnml.Nets[0] // Assuming only one net is present
//	arc := CreateArc(id, source, target)
//	net.Arcs = append(net.Arcs, arc)
//}
//
//// Check if a transition with the given ID exists in the Petri net
//func TransitionExists(pnml PNML, id string) bool {
//	net := pnml.Nets[0] // Assuming only one net is present
//
//	for _, transition := range net.Transitions {
//		if transition.ID == id {
//			return true
//		}
//	}
//
//	return false
//}
//
//// Check if a place with the given ID exists in the Petri net
//func PlaceExists(pnml PNML, id string) bool {
//	net := pnml.Nets[0] // Assuming only one net is present
//
//	for _, place := range net.Places {
//		if place.ID == id {
//			return true
//		}
//	}
//
//	return false
//}
//
//// Check if an arc with the given ID exists in the Petri net
//func ArcExists(pnml PNML, id string) bool {
//	net := pnml.Nets[0] // Assuming only one net is present
//
//	for _, arc := range net.Arcs {
//		if arc.ID == id {
//			return true
//		}
//	}
//
//	return false
//}
//
//// Function that returns the next available ID for a place in the Petri net
//func NextPlaceID(pnml PNML, prefix string) string {
//	net := pnml.Nets[0] // Assuming only one net is present
//
//	highestNumber := 0
//	for _, place := range net.Places {
//		number := getNumericPart(place.ID, prefix)
//		if number > highestNumber {
//			highestNumber = number
//		}
//	}
//
//	nextNumber := highestNumber + 1
//	return fmt.Sprintf("%s%d", prefix, nextNumber)
//}
//
//// Function that returns the next available ID for an arc in the Petri net
//func NextArcID(pnml PNML, prefix string) string {
//	net := pnml.Nets[0] // Assuming only one net is present
//
//	highestNumber := 0
//	for _, arc := range net.Arcs {
//		number := getNumericPart(arc.ID, prefix)
//		if number > highestNumber {
//			highestNumber = number
//		}
//	}
//
//	nextNumber := highestNumber + 1
//	return fmt.Sprintf("%s%d", prefix, nextNumber)
//}
//
//// Helper function to extract the numeric part of an ID
//func getNumericPart(id, prefix string) int {
//	trimmedID := strings.TrimPrefix(id, prefix)
//	number, err := strconv.Atoi(trimmedID)
//	if err != nil {
//		return 0
//	}
//	return number
//}
//
//func (net Net) GetInputOutputMatrices() (map[string]int, [][]int, [][]int) {
//	placeIndex := make(map[string]int)
//	for i, place := range net.Places {
//		placeIndex[place.ID] = i
//	}
//
//	transitionIndex := make(map[string]int)
//	for i, transition := range net.Transitions {
//		transitionIndex[transition.Name.Text] = i
//	}
//
//	inputMatrix := make([][]int, len(net.Transitions))
//	outputMatrix := make([][]int, len(net.Transitions))
//	for i := range inputMatrix {
//		inputMatrix[i] = make([]int, len(net.Places))
//		outputMatrix[i] = make([]int, len(net.Places))
//	}
//
//	for _, arc := range net.Arcs {
//		if tIndex, isTransitionSource := transitionIndex[arc.Source]; isTransitionSource {
//			if pIndex, isPlaceTarget := placeIndex[arc.Target]; isPlaceTarget {
//				outputMatrix[tIndex][pIndex] = 1
//			}
//		} else if tIndex, isTransitionTarget := transitionIndex[arc.Target]; isTransitionTarget {
//			if pIndex, isPlaceSource := placeIndex[arc.Source]; isPlaceSource {
//				inputMatrix[tIndex][pIndex] = 1
//			}
//		}
//	}
//	return transitionIndex, inputMatrix, outputMatrix
//}
//
//// Function to merge multiple places into a single place
//func MergePlaces(pnml *PNML, placeIDs []string, mergedPlaceID string) error {
//	net := &pnml.Nets[0] // Assuming only one net is present
//
//	// Find the places and arcs to be merged
//	var places []Place
//	var arcs []Arc
//	for _, placeID := range placeIDs {
//		foundPlace := false
//		for _, place := range net.Places {
//			if place.ID == placeID {
//				places = append(places, place)
//				foundPlace = true
//				break
//			}
//		}
//		if !foundPlace {
//			return fmt.Errorf("place %s not found", placeID)
//		}
//
//		for _, arc := range net.Arcs {
//			if arc.Source == placeID || arc.Target == placeID {
//				arcs = append(arcs, arc)
//			}
//		}
//	}
//
//	// Check if the places and arcs to be merged were found
//	if len(places) != len(placeIDs) || len(arcs) == 0 {
//		return errors.New("places or arcs not found")
//	}
//
//	// Create the merged place
//	mergedPlace := Place{
//		ID:   mergedPlaceID,
//		Name: Name{Text: fmt.Sprintf("Merged Place (%s)", strings.Join(placeIDs, ", "))},
//	}
//	net.Places = append(net.Places, mergedPlace)
//
//	// Create arcs from merged place to target transitions
//	for _, arc := range arcs {
//		if arc.Source != mergedPlaceID {
//			arc.Source = mergedPlaceID
//		}
//		if arc.Target != mergedPlaceID {
//			arc.Target = mergedPlaceID
//		}
//	}
//
//	// Remove the original places and arcs
//	for _, place := range places {
//		removePlace(net, place.ID)
//	}
//	for _, arc := range arcs {
//		removeArc(net, arc.ID)
//	}
//
//	return nil
//}
//
//// Helper function to remove a place from the Petri net
//func removePlace(net *Net, placeID string) {
//	for i, place := range net.Places {
//		if place.ID == placeID {
//			net.Places = append(net.Places[:i], net.Places[i+1:]...)
//			break
//		}
//	}
//}
//
//// Helper function to remove an arc from the Petri net
//func removeArc(net *Net, arcID string) {
//	for i, arc := range net.Arcs {
//		if arc.ID == arcID {
//			net.Arcs = append(net.Arcs[:i], net.Arcs[i+1:]...)
//			break
//		}
//	}
//}
//
////
////// Function to check if it is possible to go from a start activity to the end activity without visiting a specific activity in a PNML
////func CanReachEndWithoutActivity(pnml *PNML, startActivityID string, excludedActivityID string) bool {
////	net := &pnml.Nets[0] // Assuming only one net is present
////
////	visited := make(map[string]bool) // Track visited activities
////	return dfs(net, startActivityID, excludedActivityID, visited)
////}
////
////// DFS helper function to recursively explore the PNML
////func dfs(net *Net, currentActivityID string, excludedActivityID string, visited map[string]bool) bool {
////	visited[currentActivityID] = true
////
////	// Base case: Check if the current activity is the end activity
////	if currentActivityID == "End" {
////		return true
////	}
////
////	// Iterate over arcs from the current activity
////	for _, arc := range net.Arcs {
////		if arc.Source == currentActivityID && arc.Target != excludedActivityID && !visited[arc.Target] {
////			// Recursively explore the target activity
////			fmt.Println("v -> ", visited)
////			if dfs(net, arc.Target, excludedActivityID, visited) {
////				return true
////			}
////		}
////	}
////
////	// The end activity was not reached without visiting the excluded activity
////	return false
////}
//
//// Function to find all possible paths from the first place to the last place in a PNML
//func FindAllPaths(pnml *PNML) [][]string {
//	net := &pnml.Nets[0] // Assuming only one net is present
//
//	startPlaceID := FindStartPlaceID(net)
//	endPlaceID := FindEndPlaceID(net)
//
//	// Create a map to track visited places
//	visited := make(map[string]bool)
//
//	// Create a slice to store the paths
//	paths := [][]string{}
//
//	// Perform depth-first search (DFS) starting from the first place
//	dfs(net, startPlaceID, endPlaceID, visited, []string{startPlaceID}, &paths)
//
//	return paths
//}
//
//func containsString(arr []string, target string) bool {
//	for _, s := range arr {
//		if s == target {
//			return true
//		}
//	}
//	return false
//}
//
//func FindAPathWithExclusion(paths [][]string, start string, exclusion string) bool {
//	for _, e := range paths {
//		if containsString(e, start) && !containsString(e, exclusion) {
//			//fmt.Println(e)
//			return true
//		}
//	}
//	return false
//}
//
//// Helper function to find the ID of the first place (start place)
//func FindStartPlaceID(net *Net) string {
//	for _, place := range net.Places {
//		// Check if the place has no incoming arcs
//		if !HasIncomingArcs(net.Arcs, place.ID) {
//			return place.ID
//		}
//	}
//	return ""
//}
//
//// Helper function to find the ID of the last place (end place)
//func FindEndPlaceID(net *Net) string {
//	for _, place := range net.Places {
//		// Check if the place has no outgoing arcs
//		if !HasOutgoingArcs(net.Arcs, place.ID) {
//			return place.ID
//		}
//	}
//	return ""
//}
//
//// Helper function to check if a place has incoming arcs
//func HasIncomingArcs(arcs []Arc, placeID string) bool {
//	for _, arc := range arcs {
//		if arc.Target == placeID {
//			return true
//		}
//	}
//	return false
//}
//
//// Helper function to check if a place has outgoing arcs
//func HasOutgoingArcs(arcs []Arc, placeID string) bool {
//	for _, arc := range arcs {
//		if arc.Source == placeID {
//			return true
//		}
//	}
//	return false
//}
//
//// DFS helper function to recursively explore the PNML and find paths
//func dfs(net *Net, currentPlaceID string, endPlaceID string, visited map[string]bool, currentPath []string, paths *[][]string) {
//	// Mark the current place as visited
//	visited[currentPlaceID] = true
//
//	// Base case: Check if the current place is the end place
//	if currentPlaceID == endPlaceID {
//		// Add the current path to the list of paths
//		*paths = append(*paths, append([]string{}, currentPath...))
//	}
//
//	// Iterate over arcs and explore neighboring places
//	for _, arc := range net.Arcs {
//		if arc.Source == currentPlaceID && !visited[arc.Target] {
//			// Recursively explore the neighboring place
//			dfs(net, arc.Target, endPlaceID, visited, append(currentPath, arc.Target), paths)
//		}
//	}
//	// Mark the current place as unvisited before backtracking
//	delete(visited, currentPlaceID)
//}

package pnml

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type PNML struct {
	XMLName xml.Name `xml:"pnml"`
	Nets    []Net    `xml:"net"`
}

type Net struct {
	XMLName     xml.Name     `xml:"net"`
	ID          string       `xml:"id,attr"`
	Places      []Place      `xml:"place"`
	Transitions []Transition `xml:"transition"`
	Arcs        []Arc        `xml:"arc"`
	Pages       []Page       `xml:"page"`
}

type Page struct {
	XMLName     xml.Name     `xml:"page"`
	Places      []Place      `xml:"place"`
	Transitions []Transition `xml:"transition"`
	Arcs        []Arc        `xml:"arc"`
}

type Place struct {
	XMLName xml.Name `xml:"place"`
	ID      string   `xml:"id,attr"`
	Name    Name     `xml:"name"`
}

type Transition struct {
	XMLName xml.Name `xml:"transition"`
	ID      string   `xml:"id,attr"`
	Name    Name     `xml:"name"`
}

type Arc struct {
	XMLName xml.Name `xml:"arc"`
	ID      string   `xml:"id,attr"`
	Source  string   `xml:"source,attr"`
	Target  string   `xml:"target,attr"`
}

type Name struct {
	Text string `xml:"text"`
}

// Create an initial Petri net with some places, transitions, and arcs
func CreateInitialPetriNet() PNML {
	net := Net{
		ID:          "net1",
		Places:      []Place{},
		Transitions: []Transition{},
		Arcs:        []Arc{},
	}

	pnml := PNML{
		Nets: []Net{net},
	}

	return pnml
}

func MergePagesIntoNet(net *Net) {
	for _, page := range net.Pages {
		net.Places = append(net.Places, page.Places...)
		net.Transitions = append(net.Transitions, page.Transitions...)
		net.Arcs = append(net.Arcs, page.Arcs...)
	}
	net.Pages = nil // Clear the pages after merging
}

// Create a new place with the given ID and name
func CreatePlace(id, name string) Place {
	return Place{
		ID:   id,
		Name: Name{Text: name},
	}
}

// Create a new transition with the given ID and name
func CreateTransition(id, name string) Transition {
	return Transition{
		ID:   id,
		Name: Name{Text: name},
	}
}

// Create a new arc with the given ID, source, and target
func CreateArc(id, source, target string) Arc {
	return Arc{
		ID:     id,
		Source: source,
		Target: target,
	}
}

// Add a new transition to the Petri net
func AddTransition(pnml *PNML, id, name string) {
	net := &pnml.Nets[0] // Assuming only one net is present
	transition := CreateTransition(id, name)
	net.Transitions = append(net.Transitions, transition)
}

// Add a new element (place) to the Petri net
func AddPlace(pnml *PNML, id, name string) {
	net := &pnml.Nets[0] // Assuming only one net is present
	place := CreatePlace(id, name)
	net.Places = append(net.Places, place)
}

// Add a new arc to the Petri net
func AddArc(pnml *PNML, id, source, target string) {
	net := &pnml.Nets[0] // Assuming only one net is present
	arc := CreateArc(id, source, target)
	net.Arcs = append(net.Arcs, arc)
}

// Check if a transition with the given ID exists in the Petri net
func TransitionExists(pnml PNML, id string) bool {
	net := pnml.Nets[0] // Assuming only one net is present

	for _, transition := range net.Transitions {
		if transition.ID == id {
			return true
		}
	}

	return false
}

func RemovePages(pnml PNML) {
	for i := range pnml.Nets {
		MergePagesIntoNet(&pnml.Nets[i])
	}
}

// Check if a place with the given ID exists in the Petri net
func PlaceExists(pnml PNML, id string) bool {
	net := pnml.Nets[0] // Assuming only one net is present

	for _, place := range net.Places {
		if place.ID == id {
			return true
		}
	}

	return false
}

// Check if an arc with the given ID exists in the Petri net
func ArcExists(pnml PNML, id string) bool {
	net := pnml.Nets[0] // Assuming only one net is present

	for _, arc := range net.Arcs {
		if arc.ID == id {
			return true
		}
	}

	return false
}

// Function that returns the next available ID for a place in the Petri net
func NextPlaceID(pnml PNML, prefix string) string {
	net := pnml.Nets[0] // Assuming only one net is present

	highestNumber := 0
	for _, place := range net.Places {
		number := getNumericPart(place.ID, prefix)
		if number > highestNumber {
			highestNumber = number
		}
	}

	nextNumber := highestNumber + 1
	return fmt.Sprintf("%s%d", prefix, nextNumber)
}

// Function that returns the next available ID for an arc in the Petri net
func NextArcID(pnml PNML, prefix string) string {
	net := pnml.Nets[0] // Assuming only one net is present

	highestNumber := 0
	for _, arc := range net.Arcs {
		number := getNumericPart(arc.ID, prefix)
		if number > highestNumber {
			highestNumber = number
		}
	}

	nextNumber := highestNumber + 1
	return fmt.Sprintf("%s%d", prefix, nextNumber)
}

// Helper function to extract the numeric part of an ID
func getNumericPart(id, prefix string) int {
	trimmedID := strings.TrimPrefix(id, prefix)
	number, err := strconv.Atoi(trimmedID)
	if err != nil {
		return 0
	}
	return number
}

func (net Net) GetInputOutputMatrices() (map[string]int, [][]int, [][]int) {
	placeIndex := make(map[string]int)
	for i, place := range net.Places {
		placeIndex[place.ID] = i
	}

	transitionIndex := make(map[string]int)
	for i, transition := range net.Transitions {
		transitionIndex[transition.Name.Text] = i
	}

	inputMatrix := make([][]int, len(net.Transitions))
	outputMatrix := make([][]int, len(net.Transitions))
	for i := range inputMatrix {
		inputMatrix[i] = make([]int, len(net.Places))
		outputMatrix[i] = make([]int, len(net.Places))
	}

	for _, arc := range net.Arcs {
		if tIndex, isTransitionSource := transitionIndex[arc.Source]; isTransitionSource {
			if pIndex, isPlaceTarget := placeIndex[arc.Target]; isPlaceTarget {
				outputMatrix[tIndex][pIndex] = 1
			}
		} else if tIndex, isTransitionTarget := transitionIndex[arc.Target]; isTransitionTarget {
			if pIndex, isPlaceSource := placeIndex[arc.Source]; isPlaceSource {
				inputMatrix[tIndex][pIndex] = 1
			}
		}
	}
	return transitionIndex, inputMatrix, outputMatrix
}

// Function to merge multiple places into a single place
func MergePlaces(pnml *PNML, placeIDs []string, mergedPlaceID string) error {
	net := &pnml.Nets[0] // Assuming only one net is present

	// Find the places and arcs to be merged
	var places []Place
	var arcs []Arc
	for _, placeID := range placeIDs {
		foundPlace := false
		for _, place := range net.Places {
			if place.ID == placeID {
				places = append(places, place)
				foundPlace = true
				break
			}
		}
		if !foundPlace {
			return fmt.Errorf("place %s not found", placeID)
		}

		for _, arc := range net.Arcs {
			if arc.Source == placeID || arc.Target == placeID {
				arcs = append(arcs, arc)
			}
		}
	}

	// Check if the places and arcs to be merged were found
	if len(places) != len(placeIDs) || len(arcs) == 0 {
		return errors.New("places or arcs not found")
	}

	// Create the merged place
	mergedPlace := Place{
		ID:   mergedPlaceID,
		Name: Name{Text: fmt.Sprintf("Merged Place (%s)", strings.Join(placeIDs, ", "))},
	}
	net.Places = append(net.Places, mergedPlace)

	// Create arcs from merged place to target transitions
	for _, arc := range arcs {
		if arc.Source != mergedPlaceID {
			arc.Source = mergedPlaceID
		}
		if arc.Target != mergedPlaceID {
			arc.Target = mergedPlaceID
		}
	}

	// Remove the original places and arcs
	for _, place := range places {
		removePlace(net, place.ID)
	}
	for _, arc := range arcs {
		removeArc(net, arc.ID)
	}

	return nil
}

// Helper function to remove a place from the Petri net
func removePlace(net *Net, placeID string) {
	for i, place := range net.Places {
		if place.ID == placeID {
			net.Places = append(net.Places[:i], net.Places[i+1:]...)
			break
		}
	}
}

// Helper function to remove an arc from the Petri net
func removeArc(net *Net, arcID string) {
	for i, arc := range net.Arcs {
		if arc.ID == arcID {
			net.Arcs = append(net.Arcs[:i], net.Arcs[i+1:]...)
			break
		}
	}
}

// Function to find all possible paths from the first place to the last place in a PNML
func FindAllPaths(pnml *PNML) [][]string {
	net := &pnml.Nets[0] // Assuming only one net is present

	startPlaceID := FindStartPlaceID(net)
	endPlaceID := FindEndPlaceID(net)

	// Create a map to track visited places
	visited := make(map[string]bool)

	// Create a slice to store the paths
	paths := [][]string{}

	// Perform depth-first search (DFS) starting from the first place
	dfs(net, startPlaceID, endPlaceID, visited, []string{startPlaceID}, &paths)

	return paths
}

func containsString(arr []string, target string) bool {
	for _, s := range arr {
		if s == target {
			return true
		}
	}
	return false
}

func FindAPathWithExclusion(paths [][]string, start string, exclusion string) bool {
	for _, e := range paths {
		if containsString(e, start) && !containsString(e, exclusion) {
			//fmt.Println(e)
			return true
		}
	}
	return false
}

// Helper function to find the ID of the first place (start place)
func FindStartPlaceID(net *Net) string {
	for _, place := range net.Places {
		// Check if the place has no incoming arcs
		if !HasIncomingArcs(net.Arcs, place.ID) {
			return place.ID
		}
	}
	return ""
}

// Helper function to find the ID of the last place (end place)
func FindEndPlaceID(net *Net) string {
	for _, place := range net.Places {
		// Check if the place has no outgoing arcs
		if !HasOutgoingArcs(net.Arcs, place.ID) {
			return place.ID
		}
	}
	return ""
}

// Helper function to check if a place has incoming arcs
func HasIncomingArcs(arcs []Arc, placeID string) bool {
	for _, arc := range arcs {
		if arc.Target == placeID {
			return true
		}
	}
	return false
}

// Helper function to check if a place has outgoing arcs
func HasOutgoingArcs(arcs []Arc, placeID string) bool {
	for _, arc := range arcs {
		if arc.Source == placeID {
			return true
		}
	}
	return false
}

// DFS helper function to recursively explore the PNML and find paths
func dfs(net *Net, currentPlaceID string, endPlaceID string, visited map[string]bool, currentPath []string, paths *[][]string) {
	// Mark the current place as visited
	visited[currentPlaceID] = true

	// Base case: Check if the current place is the end place
	if currentPlaceID == endPlaceID {
		// Add the current path to the list of paths
		*paths = append(*paths, append([]string{}, currentPath...))
	}

	// Iterate over arcs and explore neighboring places
	for _, arc := range net.Arcs {
		if arc.Source == currentPlaceID && !visited[arc.Target] {
			// Recursively explore the neighboring place
			dfs(net, arc.Target, endPlaceID, visited, append(currentPath, arc.Target), paths)
		}
	}
	// Mark the current place as unvisited before backtracking
	delete(visited, currentPlaceID)
}

// main function to test the package
func main() {
	//read a file in a byte array
	file, err := os.ReadFile("sample_pnml.pnml")
	if err != nil {
		fmt.Println(err)
		return
	}
	//unmarshal the file into a PNML struct
	pnml := PNML{}
	err = xml.Unmarshal(file, &pnml)
	if err != nil {
		fmt.Println(err)
		return
	}
	//print the PNML transitions
	RemovePages(pnml)
}
