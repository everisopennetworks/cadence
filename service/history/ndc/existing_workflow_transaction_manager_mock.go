// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Code generated by MockGen. DO NOT EDIT.
// Source: existing_workflow_transaction_manager.go

// Package ndc is a generated GoMock package.
package ndc

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
	execution "github.com/uber/cadence/service/history/execution"
)

// MocktransactionManagerForExistingWorkflow is a mock of transactionManagerForExistingWorkflow interface.
type MocktransactionManagerForExistingWorkflow struct {
	ctrl     *gomock.Controller
	recorder *MocktransactionManagerForExistingWorkflowMockRecorder
}

// MocktransactionManagerForExistingWorkflowMockRecorder is the mock recorder for MocktransactionManagerForExistingWorkflow.
type MocktransactionManagerForExistingWorkflowMockRecorder struct {
	mock *MocktransactionManagerForExistingWorkflow
}

// NewMocktransactionManagerForExistingWorkflow creates a new mock instance.
func NewMocktransactionManagerForExistingWorkflow(ctrl *gomock.Controller) *MocktransactionManagerForExistingWorkflow {
	mock := &MocktransactionManagerForExistingWorkflow{ctrl: ctrl}
	mock.recorder = &MocktransactionManagerForExistingWorkflowMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MocktransactionManagerForExistingWorkflow) EXPECT() *MocktransactionManagerForExistingWorkflowMockRecorder {
	return m.recorder
}

// dispatchForExistingWorkflow mocks base method.
func (m *MocktransactionManagerForExistingWorkflow) dispatchForExistingWorkflow(ctx context.Context, now time.Time, isWorkflowRebuilt bool, targetWorkflow, newWorkflow execution.Workflow) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "dispatchForExistingWorkflow", ctx, now, isWorkflowRebuilt, targetWorkflow, newWorkflow)
	ret0, _ := ret[0].(error)
	return ret0
}

// dispatchForExistingWorkflow indicates an expected call of dispatchForExistingWorkflow.
func (mr *MocktransactionManagerForExistingWorkflowMockRecorder) dispatchForExistingWorkflow(ctx, now, isWorkflowRebuilt, targetWorkflow, newWorkflow interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "dispatchForExistingWorkflow", reflect.TypeOf((*MocktransactionManagerForExistingWorkflow)(nil).dispatchForExistingWorkflow), ctx, now, isWorkflowRebuilt, targetWorkflow, newWorkflow)
}
