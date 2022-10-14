package sam

import (
	"reflect"
	"testing"
)

const nilStep = State("")

func TestNewStateMachine(t *testing.T) {
	sm := NewStateMachine()
	EqualStates(t, nilStep, sm.current)
	NotNil(t, sm.states)
	NotNil(t, sm.hooks)
}

func TestStateMachine_State(t *testing.T) {
	sm := NewStateMachine()

	EqualStates(t, nilStep, sm.current)
	NotNil(t, sm.states)
	NotNil(t, sm.hooks)

	sm.current = "test"
	EqualStates(t, sm.current, sm.State())
}

func TestStateMachine_getState(t *testing.T) {
	sm := NewStateMachine()
	name := State("test")

	state := sm.getState(name)
	EqualStates(t, name, state.name)
	EqualStates(t, nilStep, state.prev)
	NotNil(t, state.from)
	NotNil(t, state.to)
	False(t, state.exist)
}

func TestStateMachine_SetState(t *testing.T) {
	sm := NewStateMachine()
	name := State("test")

	_ = sm.SetState(name)
	EqualStates(t, name, sm.State())

	state := sm.getState(name)
	EqualStates(t, name, state.name)
	NotNil(t, state.from)
	NotNil(t, state.to)
	True(t, state.exist)
}

func TestStateMachine_AddTransition(t *testing.T) {
	sm := NewStateMachine()
	from := State("from")
	to := State("to")

	err := sm.
		AddTransition(from, from).
		AddTransition(from, to).Error()
	NotNil(t, err)
	EqualStr(t, invalidTransition(from, from).Error(), err.Error())

	state := sm.getState(from)
	EqualStates(t, from, state.name)
	False(t, state.exist)
	NotNil(t, state.from)
	NotNil(t, state.to)
	_, ok := state.to[to]
	False(t, ok)

	state = sm.getState(to)
	EqualStates(t, to, state.name)
	False(t, state.exist)
	NotNil(t, state.from)
	NotNil(t, state.to)
	_, ok = state.from[from]
	False(t, ok)
}

func TestStateMachine_AddTransitions(t *testing.T) {
	sm := NewStateMachine()
	from := State("from")
	tos := []State{"to_1", "to_2", "to_3"}
	err := sm.AddTransitions(from, tos...).Error()
	NoError(t, err)

	state := sm.getState(from)
	EqualStates(t, from, state.name)
	True(t, state.exist)
	NotNil(t, state.from)
	NotNil(t, state.to)

	for _, to := range tos {
		_, ok := state.to[to]
		True(t, ok)

		stateTo := sm.getState(to)
		EqualStates(t, to, stateTo.name)
		True(t, stateTo.exist)
		NotNil(t, stateTo.from)
		NotNil(t, stateTo.to)
		_, ok = stateTo.from[from]
		True(t, ok)
	}

}

func TestStateMachine_DoTransition(t *testing.T) {
	sm := NewStateMachine()
	from := State("from")
	tos := []State{"to_1", "to_2", "to_3"}
	err := sm.AddTransitions(from, tos...).SetState(from)
	NoError(t, err)

	err = sm.GoTo("universe")
	NotNil(t, err)
	EqualStr(t, stateNotFound("universe").Error(), err.Error())
	EqualStates(t, from, sm.State())

	err = sm.GoTo(from)
	NoError(t, err)
	EqualStates(t, from, sm.State())

	clone := sm.Clone()
	{
		err = clone.GoTo(tos[1])
		NoError(t, err)
		EqualStates(t, tos[1], clone.State())
	}

	err = sm.
		AddTransitions(tos[0], tos[2]).
		AddTransitions(tos[1], tos[0], tos[2]).
		SetState(from)

	NoError(t, err)

	{
		clone = sm.Clone()
		err = clone.GoTo(tos[0])
		NoError(t, err)
		EqualStates(t, tos[0], clone.State())

		err = clone.GoTo(tos[2])
		NoError(t, err)
		EqualStates(t, tos[2], clone.State())
	}

	err = sm.GoTo(tos[1])
	NoError(t, err)
	EqualStates(t, tos[1], sm.State())

	clone = sm.Clone()
	err = clone.GoTo(tos[0])
	NoError(t, err)
	EqualStates(t, tos[0], clone.State())

	clone = sm.Clone()
	err = clone.GoTo(tos[2])
	NoError(t, err)
	EqualStates(t, tos[2], clone.State())
}

func TestStateMachine_GoBack(t *testing.T) {
	sm := NewStateMachine()
	from := State("from")
	tos := []State{"to_1", "to_2", "to_3"}

	err := sm.
		AddTransitions(from, tos[0]).
		AddTransitions(tos[0], tos[1]).
		SetState(from)
	NoError(t, err)

	err = sm.GoBack()
	NotNil(t, err)
	EqualStr(t, invalidTransition(from, "").Error(), err.Error())

	err = sm.GoTo(tos[0])
	NoError(t, err)
	EqualStates(t, tos[0], sm.State())

	err = sm.GoTo(tos[1])
	NoError(t, err)
	EqualStates(t, tos[1], sm.State())

	err = sm.GoBack()
	NoError(t, err)
	EqualStates(t, tos[0], sm.State())

	err = sm.GoBack()
	NoError(t, err)
	EqualStates(t, from, sm.State())
}

func NoError(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}

func EqualStates(t *testing.T, got, expected State) {
	if got != expected {
		t.Errorf("expected(%s) != got(%s)", expected, got)
		t.FailNow()
	}
}

func EqualStr(t *testing.T, got, expected string) {
	if got != expected {
		t.Errorf("expected(%s) != got(%s)", expected, got)
		t.FailNow()
	}
}

func NotNil(t *testing.T, v interface{}) {
	value := reflect.ValueOf(v)
	if v == nil || value.IsNil() {
		t.Error("expected to be not nil")
		t.FailNow()
	}
}

func True(t *testing.T, v bool) {
	if !v {
		t.Error("expected to be true")
		t.FailNow()
	}
}

func False(t *testing.T, v bool) {
	if v {
		t.Error("expected to be false")
		t.FailNow()
	}
}
