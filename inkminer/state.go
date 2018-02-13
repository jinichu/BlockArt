package inkminer

import "../blockartlib"

// State represents the state of a block at a certain point.
type State struct {
	blockNum    int
	shapes      map[string]blockartlib.Shape // Map of shape hashes to their SVG string representation
	shapeOwners map[string]string            // Map of shape hashes to their owner (InkMiner PubKey)
	inkLevels   map[string]uint32            // Current ink levels of every InkMiner
	// commitedOperations is a set of currently committed operations
	commitedOperations map[string]struct{}
}

// NewState creates a new state.
func NewState() State {
	return State{
		shapes:             make(map[string]blockartlib.Shape),
		shapeOwners:        make(map[string]string),
		inkLevels:          make(map[string]uint32),
		commitedOperations: make(map[string]struct{}),
	}
}

// Copy returns a copy of the given state.
func (s State) Copy() State {
	s2 := NewState()

	s2.blockNum = s.blockNum

	for key, value := range s.shapes {
		s2.shapes[key] = value
	}

	for key, value := range s.shapeOwners {
		s2.shapeOwners[key] = value
	}

	for key, value := range s.inkLevels {
		s2.inkLevels[key] = value
	}

	for key, value := range s.commitedOperations {
		s2.commitedOperations[key] = value
	}

	return s2
}
