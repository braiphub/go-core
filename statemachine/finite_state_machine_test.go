package statemachine

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MyTestStruct struct {
	Status  string
	NonText uint
}

func TestInfiniteStateMachine_validate(t *testing.T) {
	validateSuccess := func(MyTestStruct) error { return nil }
	validateError := func(MyTestStruct) error { return errors.New("unknown") }

	type args[T MyTestStruct, V string] struct {
		t         *T
		newStatus V
	}
	type testCase[T MyTestStruct, V string] struct {
		name    string
		m       FiniteStateMachine[T, V]
		args    args[T, V]
		wantErr assert.ErrorAssertionFunc
	}
	tests := []testCase[MyTestStruct, string]{
		{
			name: "error: current and new status are the same",
			m: FiniteStateMachine[MyTestStruct, string]{
				statusField: "Status",
			},
			args: args[MyTestStruct, string]{
				t: &MyTestStruct{
					Status: "test-status",
				},
				newStatus: "test-status",
			},
			wantErr: assert.Error,
		},
		{
			name: "error: source status can't be changed",
			m: FiniteStateMachine[MyTestStruct, string]{
				statusField: "Status",
			},
			args: args[MyTestStruct, string]{
				t: &MyTestStruct{
					Status: "test-status",
				},
				newStatus: "new-status",
			},
			wantErr: assert.Error,
		},
		{
			name: "error: destination status isn't allowed",
			m: func() FiniteStateMachine[MyTestStruct, string] {
				stm := NewFiniteStateMachine[MyTestStruct, string]("Status")
				stm.registerSwitches(
					Switcher[MyTestStruct, string]{
						Src:       "creating",
						Dest:      "created",
						Validator: nil,
					},
				)

				return *stm
			}(),
			args: args[MyTestStruct, string]{
				t: &MyTestStruct{
					Status: "creating",
				},
				newStatus: "done",
			},
			wantErr: assert.Error,
		},
		{
			name: "error: validator returns error",
			m: func() FiniteStateMachine[MyTestStruct, string] {
				stm := NewFiniteStateMachine[MyTestStruct, string]("Status")
				stm.registerSwitches(
					Switcher[MyTestStruct, string]{
						Src:       "creating",
						Dest:      "created",
						Validator: validateError,
					},
				)

				return *stm
			}(),
			args: args[MyTestStruct, string]{
				t: &MyTestStruct{
					Status: "creating",
				},
				newStatus: "created",
			},
			wantErr: assert.Error,
		},
		{
			name: "success: validator returns nil",
			m: func() FiniteStateMachine[MyTestStruct, string] {
				stm := NewFiniteStateMachine[MyTestStruct, string]("Status")
				stm.registerSwitches(
					Switcher[MyTestStruct, string]{
						Src:       "creating",
						Dest:      "created",
						Validator: validateSuccess,
					},
				)

				return *stm
			}(),
			args: args[MyTestStruct, string]{
				t: &MyTestStruct{
					Status: "creating",
				},
				newStatus: "created",
			},
			wantErr: assert.NoError,
		},
		{
			name: "success: doesn't have a valitador",
			m: func() FiniteStateMachine[MyTestStruct, string] {
				stm := NewFiniteStateMachine[MyTestStruct, string]("Status")
				stm.registerSwitches(
					Switcher[MyTestStruct, string]{
						Src:       "creating",
						Dest:      "created",
						Validator: nil,
					},
				)

				return *stm
			}(),
			args: args[MyTestStruct, string]{
				t: &MyTestStruct{
					Status: "creating",
				},
				newStatus: "created",
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, tt.m.validate(tt.args.t, tt.args.newStatus), fmt.Sprintf("validate(%v, %v)", tt.args.t, tt.args.newStatus))
		})
	}
}

func TestInfiniteStateMachine_applyNewStatus(t *testing.T) {
	myStruct := MyTestStruct{}

	m := FiniteStateMachine[MyTestStruct, string]{
		statusField: "Status",
	}
	m.applyNewStatus(&myStruct, "test")

	assert.Equal(t, "test", myStruct.Status)
}

func TestFiniteStateMachine_ChangeState(t *testing.T) {
	switchers := []Switcher[MyTestStruct, string]{
		{
			Src:       "creating",
			Dest:      "created",
			Validator: nil,
		},
	}

	myStruct := MyTestStruct{
		Status: "creating",
	}

	m := NewFiniteStateMachine[MyTestStruct, string]("Status", switchers...)

	// err: can't go done
	assert.Error(t, m.ChangeState(&myStruct, "done"))

	// err: success going to created
	assert.NoError(t, m.ChangeState(&myStruct, "created"))
}
