package example

import "github.com/braiphub/go-core/statemachine"

type sampleStruct struct {
	StatusField status
}

type status struct {
	s string
}

var (
	statusDoing = status{"doing"}
	statusDone  = status{"done"}
	statusError = status{"error"}
)

func validateIfCanItGoFromErrorToDone(sampleStruct) error {
	return nil
}

func main() {
	switchers := []statemachine.Switcher[sampleStruct, status]{
		{Src: statusDoing, Dest: statusDone},
		{Src: statusDoing, Dest: statusError},
		{Src: statusError, Dest: statusDone, Validator: validateIfCanItGoFromErrorToDone},
	}
	stm := statemachine.NewFiniteStateMachine[sampleStruct, status]("StatusField", switchers...)

	sampleStruct := &sampleStruct{
		StatusField: statusDoing,
	}

	if err := stm.ChangeState(sampleStruct, statusDone); err != nil {
		// handle error
	}

	// success
}
