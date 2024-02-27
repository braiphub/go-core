
### STATE-MACHINE IMPLEMENT EXAMPLE:

The below example shows how to handle a state change from a struct using the State Machine to protect unexpected state changes.

Aditional Validators can be used to help some specific changes to pass by validation before changing the state. 

```golang
package mypackage

import "github.com/braiphub/go-core/statemachine"

type SampleStruct struct {
	StatusField Status
}

type Status struct {
	s string
}

var (
	StatusDoing = Status{"doing"}
	StatusDone  = Status{"done"}
	StatusError = Status{"error"}
)

func validateIfCanItGoFromErrorToDone(SampleStruct) error {
	return nil
}

func implementFiniteStateMachine() {
	switchers := []statemachine.Switcher[SampleStruct, Status]{
		{Src: StatusDoing, Dest: StatusDone},
		{Src: StatusDoing, Dest: StatusError},
		{Src: StatusError, Dest: StatusDone, Validator: validateIfCanItGoFromErrorToDone},
	}
	stm := statemachine.NewFiniteStateMachine[SampleStruct, Status]("StatusField", switchers...)

	sampleStruct := &SampleStruct{
		StatusField: StatusDoing,
	}

	if err := stm.ChangeState(sampleStruct, StatusDone); err != nil {
		// handle error
	}

	// success
}
```

