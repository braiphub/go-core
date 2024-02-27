package statemachine

import (
	"github.com/pkg/errors"
	"reflect"
)

type Switcher[T any, V any] struct {
	Src       V
	Dest      V
	Validator func(T) error
}

type FiniteStateMachine[T any, V any] struct {
	statusField string
	switchers   map[any]map[any]func(T) error
}

var (
	ErrBeforeAndAfterValuesAreTheSame = errors.New("stm: before and after values are the same")
	ErrSrcStatusChangeNotAllowed      = errors.New("stm: this state can't be changed")
	ErrSwitchNotAllowed               = errors.New("stm: this state switch isn't allowed")
)

func NewFiniteStateMachine[T any, V any](statusField string, switchers ...Switcher[T, V]) *FiniteStateMachine[T, V] {
	stm := &FiniteStateMachine[T, V]{
		statusField: statusField,
		switchers:   make(map[any]map[any]func(T) error),
	}

	stm.registerSwitches(switchers...)

	return stm
}

func (m *FiniteStateMachine[T, V]) registerSwitches(switchers ...Switcher[T, V]) {
	for _, switcher := range switchers {
		if m.switchers[switcher.Src] == nil {
			m.switchers[switcher.Src] = make(map[any]func(T) error)
		}

		m.switchers[switcher.Src][switcher.Dest] = switcher.Validator
	}
}

func (m *FiniteStateMachine[T, V]) ChangeState(t *T, newStatus V) error {
	if err := m.validate(t, newStatus); err != nil {
		return err
	}

	m.applyNewStatus(t, newStatus)

	return nil
}

func (m *FiniteStateMachine[T, V]) validate(t *T, newStatus V) error {
	currentVal := reflect.ValueOf(t).Elem().FieldByName(m.statusField)
	newVal := reflect.ValueOf(newStatus)

	if currentVal.Equal(newVal) {
		return ErrBeforeAndAfterValuesAreTheSame
	}

	srcSwitcher, ok := m.switchers[currentVal.Interface()]
	if !ok {
		return ErrSrcStatusChangeNotAllowed
	}

	dstSwitcher, ok := srcSwitcher[newStatus]
	if !ok {
		return ErrSwitchNotAllowed
	}

	if dstSwitcher != nil {
		return dstSwitcher(*t)
	}

	return nil
}

func (m *FiniteStateMachine[T, V]) applyNewStatus(t *T, newStatus V) {
	newVal := reflect.ValueOf(newStatus)
	reflect.Indirect(reflect.ValueOf(t)).FieldByName(m.statusField).Set(newVal)
}
