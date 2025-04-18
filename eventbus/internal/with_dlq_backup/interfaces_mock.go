// Code generated by MockGen. DO NOT EDIT.
// Source: orders/pkg/core/eventbus (interfaces: EventBusInterface,PubSubInterface)

// Package eventbus is a generated GoMock package.
package with_dlq_backup

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockEventBusInterface is a mock of EventBusInterface interface.
type MockEventBusInterface struct {
	ctrl     *gomock.Controller
	recorder *MockEventBusInterfaceMockRecorder
}

// MockEventBusInterfaceMockRecorder is the mock recorder for MockEventBusInterface.
type MockEventBusInterfaceMockRecorder struct {
	mock *MockEventBusInterface
}

// NewMockEventBusInterface creates a new mock instance.
func NewMockEventBusInterface(ctrl *gomock.Controller) *MockEventBusInterface {
	mock := &MockEventBusInterface{ctrl: ctrl}
	mock.recorder = &MockEventBusInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEventBusInterface) EXPECT() *MockEventBusInterfaceMockRecorder {
	return m.recorder
}

// Publish mocks base method.
func (m *MockEventBusInterface) Publish(arg0 ...EventInterface) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Publish", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Publish indicates an expected call of Publish.
func (mr *MockEventBusInterfaceMockRecorder) Publish(arg0 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Publish", reflect.TypeOf((*MockEventBusInterface)(nil).Publish), arg0...)
}

// Register mocks base method.
func (m *MockEventBusInterface) Register(arg0 EventInterface, arg1 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Register indicates an expected call of Register.
func (mr *MockEventBusInterfaceMockRecorder) Register(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockEventBusInterface)(nil).Register), arg0, arg1)
}

// RegisterList mocks base method.
func (m *MockEventBusInterface) RegisterList(arg0 map[EventInterface][]interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterList", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RegisterList indicates an expected call of RegisterList.
func (mr *MockEventBusInterfaceMockRecorder) RegisterList(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterList", reflect.TypeOf((*MockEventBusInterface)(nil).RegisterList), arg0)
}

// StartListen mocks base method.
func (m *MockEventBusInterface) StartListen(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartListen", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// StartListen indicates an expected call of StartListen.
func (mr *MockEventBusInterfaceMockRecorder) StartListen(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartListen", reflect.TypeOf((*MockEventBusInterface)(nil).StartListen), arg0)
}

// MockPubSubInterface is a mock of PubSubInterface interface.
type MockPubSubInterface struct {
	ctrl     *gomock.Controller
	recorder *MockPubSubInterfaceMockRecorder
}

// MockPubSubInterfaceMockRecorder is the mock recorder for MockPubSubInterface.
type MockPubSubInterfaceMockRecorder struct {
	mock *MockPubSubInterface
}

// NewMockPubSubInterface creates a new mock instance.
func NewMockPubSubInterface(ctrl *gomock.Controller) *MockPubSubInterface {
	mock := &MockPubSubInterface{ctrl: ctrl}
	mock.recorder = &MockPubSubInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPubSubInterface) EXPECT() *MockPubSubInterfaceMockRecorder {
	return m.recorder
}

// Configure mocks base method.
func (m *MockPubSubInterface) Configure(arg0 Config) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Configure", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Configure indicates an expected call of Configure.
func (mr *MockPubSubInterfaceMockRecorder) Configure(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Configure", reflect.TypeOf((*MockPubSubInterface)(nil).Configure), arg0)
}

// ListenToEvents mocks base method.
func (m *MockPubSubInterface) ListenToEvents(arg0 context.Context, arg1 SubscriberCallbackFunc) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListenToEvents", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// ListenToEvents indicates an expected call of ListenToEvents.
func (mr *MockPubSubInterfaceMockRecorder) ListenToEvents(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListenToEvents", reflect.TypeOf((*MockPubSubInterface)(nil).ListenToEvents), arg0, arg1)
}

// Publish mocks base method.
func (m *MockPubSubInterface) Publish(arg0 string, arg1 []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Publish", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Publish indicates an expected call of Publish.
func (mr *MockPubSubInterfaceMockRecorder) Publish(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Publish", reflect.TypeOf((*MockPubSubInterface)(nil).Publish), arg0, arg1)
}
