// petrinet is a simple petri net execution library
package petrinet

import (
	"errors"
	"fmt"
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"sync"
)

type Net struct {
	InputMatrix             [][]int                `json:"-"`                   // Input Matrix
	OutputMatrix            [][]int                `json:"-"`                   // Output Matrix
	ConditionMatrix         [][]string             `json:"-"`                   // Condition Matrix
	State                   []int                  `json:"-"`                   // State
	TokenIds                [][]int                `json:"token-identifier"`    // State
	Variables               map[string]interface{} `json:"variables"`           // variablen die mit dem Prozess mitlaufen
	EnabledTransitions      []int                  `json:"enabled_transitions"` // list of transitions which can be fired
	compiledConditionMatrix [][]*vm.Program        `json:"-"`                   // Precompiled Condition Matrix
	TokenId                 int
}

// token id counter

func (net *Net) Init() {
	// Initialize token ID counter (if needed)
	if net.TokenIds == nil {
		net.TokenIds = make([][]int, len(net.InputMatrix[0]))
	}

	// Assign token IDs only to tokens with no ID
	for i, tokens := range net.State {
		// Ensure that the TokenIds array for this place exists
		if net.TokenIds[i] == nil {
			net.TokenIds[i] = []int{}
		}

		// Assign new token IDs only for tokens with no ID
		for n := 0; n < tokens; n++ {
			// If the current place has fewer token IDs than the number of tokens, assign a new ID
			if len(net.TokenIds[i]) <= n {
				net.TokenIds[i] = append(net.TokenIds[i], net.NextTokenID())
			}
		}
	}

	//TODO: check on this. When init() is called, the variables may return to their initial state (precompiled)
	// Precompile conditions (unchanged logic for precompiling)
	if len(net.ConditionMatrix) > 0 {
		net.compiledConditionMatrix = make([][]*vm.Program, len(net.ConditionMatrix))
		for transitionIndex := range net.ConditionMatrix {
			net.compiledConditionMatrix[transitionIndex] = make([]*vm.Program, len(net.ConditionMatrix[transitionIndex]))
			for i, condition := range net.ConditionMatrix[transitionIndex] {
				prog, err := expr.Compile(condition, expr.Env(net.Variables))
				if err != nil {
					panic(err)
				}
				net.compiledConditionMatrix[transitionIndex][i] = prog
			}
		}
	}

	// Evaluate possible transitions after initialization
	net.EnabledTransitions = net.EvaluateNextPossibleTransitions()
}

func (net *Net) NextTokenID() int {
	net.TokenId++
	return net.TokenId
}

// Fire the transition only if all input places contain the same token ID
func (net *Net) FireWithTokenId(transition int, tokenId int) error {
	// Check if all input places have the specified token ID
	for place, step := range net.InputMatrix[transition] {
		if step > 0 {
			if len(net.TokenIds[place]) == 0 || !ContainsTokenId(net.TokenIds[place], tokenId) {
				return fmt.Errorf("place %d does not contain token ID %d", place, tokenId)
			}
		}
	}
	// If the check passes, consume tokens from input places
	for place, step := range net.InputMatrix[transition] {
		if step > 0 {
			// Remove the token ID from the place
			net.TokenIds[place] = removeTokenId(net.TokenIds[place], tokenId)
			net.State[place] -= step
		}
	}

	// Add tokens to output places with the same token ID
	for place, step := range net.OutputMatrix[transition] {
		net.State[place] += step
		for n := 0; n < step; n++ {
			net.TokenIds[place] = append(net.TokenIds[place], tokenId) // Keep the same token ID in output places
		}
	}

	// Update enabled transitions
	net.EnabledTransitions = net.EvaluateNextPossibleTransitions()
	return nil
}

// fires an enabled transition.

func (f *Net) Fire(transition int) error {
	var err error
	var mutex = &sync.Mutex{}
	if f.TransitionEnabled(transition) {
		mutex.Lock()
		f.EnabledTransitions = f.fastfire(transition)
		mutex.Unlock()
		return err
	} else {
		err = errors.New(fmt.Sprintf("Transition %v not enabled", transition))
		return err
	}
}

func (net *Net) TransitionEnabled(t int) bool {
	for _, b := range net.EnabledTransitions {
		if b == t {
			return true
		}
	}
	return false
}

func (net *Net) fastCheck(transition int) bool {
	for place, p := range net.InputMatrix[transition] {
		if p != 0 && net.State[place]-p < 0 {
			return false
		}
	}
	return true
}

func (net *Net) EvaluateNextPossibleTransitions() []int {
	var possibleTransitions []int

	for t := 0; t < len(net.InputMatrix); t++ {
		if net.fastCheck(t) {
			possibleTransitions = append(possibleTransitions, t)
		}
	}

	var lockedTransitions []int

	for t := 0; t < len(possibleTransitions); t++ {

		if !net.proveConditions(possibleTransitions[t]) {
			lockedTransitions = append(lockedTransitions, t)
			possibleTransitions = removeFromIntFromArray(possibleTransitions, t)
		}
	}
	return possibleTransitions
}

func removeFromIntFromArray(l []int, item int) []int {
	for i, other := range l {
		if other == item {
			return append(l[:i], l[i+1:]...)
		}
	}
	return l
}

func (net *Net) proveConditions(transitionIndex int) bool {

	if len(net.ConditionMatrix) > 0 {
		for _, prog := range net.compiledConditionMatrix[transitionIndex] {
			result, err := expr.Run(prog, net.Variables)
			if err != nil || !result.(bool) {
				return false
			}
		}
	}
	return true
}

func (net *Net) UpdateVariable(name string, value interface{}) {
	net.Variables[name] = value

	net.EnabledTransitions = net.EvaluateNextPossibleTransitions()
}

// fire ohne Check
func (net *Net) fastfire(transition int) []int {
	//For each place in the input matrix, subtract the step from the state
	for place, step := range net.InputMatrix[transition] {
		net.State[place] = net.State[place] - step
		//Remove the firtst n tokens from the place
		for n := 0; n < step; n++ {

			net.TokenIds[place] = net.TokenIds[place][1:]

		}
	}
	for place, step := range net.OutputMatrix[transition] {
		net.State[place] = net.State[place] + step

		for n := 0; n < step; n++ {

			nextID := net.NextTokenID()

			net.TokenIds[place] = append(net.TokenIds[place], nextID)

		}

	}

	return net.EvaluateNextPossibleTransitions()
}

// Firing transition without checking conditions but with token ID persistence
func (net *Net) fastfireWithTokenId(transition int, tokenId int) error {
	return net.FireWithTokenId(transition, tokenId)
}

// Check if the token ID exists in the token list of a place
func ContainsTokenId(tokens []int, tokenId int) bool {
	for _, id := range tokens {
		if id == tokenId {
			return true
		}
	}
	return false
}

// Remove the first occurrence of a specific token ID from a list of tokens
func removeTokenId(tokens []int, tokenId int) []int {
	for i, id := range tokens {
		if id == tokenId {
			return append(tokens[:i], tokens[i+1:]...) // Remove token by slicing
		}
	}
	return tokens
}

// Get enabled transitions for a token id
func (net *Net) GetEnabledTransitionsForTokenId(tokenId int) []int {

	enabledTransitions := []int{}
	for i, transition := range net.InputMatrix {
		enabled := true
		for j, step := range transition {
			if step > 0 {
				if !ContainsTokenId(net.TokenIds[j], tokenId) {
					enabled = false
					break
				}
			}
		}
		if enabled {
			enabledTransitions = append(enabledTransitions, i)
		}
	}
	return enabledTransitions
}
