package main

//-----------------------------------------------------------------------------

// state represents a state activity
type state interface {
	Activate() (state, error)
}

//-----------------------------------------------------------------------------

// stateFunc is a function that satisfies the State interface
type stateFunc func() (state, error)

// Activate satisfies the State interface
func (stateFn stateFunc) Activate() (state, error) { return stateFn() }

//-----------------------------------------------------------------------------

// activate activates the state and it's consecutive states until the next state
// is nil or encounters an error
func activate(s state) error {
	next := s
	var err error
	for next != nil && err == nil {
		next, err = next.Activate()
	}
	return err
}

//-----------------------------------------------------------------------------
