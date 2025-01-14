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
// Source: task_ack_manager.go

// Package replication is a generated GoMock package.
package replication

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	types "github.com/uber/cadence/common/types"
)

// MockTaskAckManager is a mock of TaskAckManager interface.
type MockTaskAckManager struct {
	ctrl     *gomock.Controller
	recorder *MockTaskAckManagerMockRecorder
}

// MockTaskAckManagerMockRecorder is the mock recorder for MockTaskAckManager.
type MockTaskAckManagerMockRecorder struct {
	mock *MockTaskAckManager
}

// NewMockTaskAckManager creates a new mock instance.
func NewMockTaskAckManager(ctrl *gomock.Controller) *MockTaskAckManager {
	mock := &MockTaskAckManager{ctrl: ctrl}
	mock.recorder = &MockTaskAckManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTaskAckManager) EXPECT() *MockTaskAckManagerMockRecorder {
	return m.recorder
}

// GetTask mocks base method.
func (m *MockTaskAckManager) GetTask(ctx context.Context, taskInfo *types.ReplicationTaskInfo) (*types.ReplicationTask, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTask", ctx, taskInfo)
	ret0, _ := ret[0].(*types.ReplicationTask)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTask indicates an expected call of GetTask.
func (mr *MockTaskAckManagerMockRecorder) GetTask(ctx, taskInfo interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTask", reflect.TypeOf((*MockTaskAckManager)(nil).GetTask), ctx, taskInfo)
}

// GetTasks mocks base method.
func (m *MockTaskAckManager) GetTasks(ctx context.Context, pollingCluster string, lastReadTaskID int64) (*types.ReplicationMessages, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTasks", ctx, pollingCluster, lastReadTaskID)
	ret0, _ := ret[0].(*types.ReplicationMessages)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTasks indicates an expected call of GetTasks.
func (mr *MockTaskAckManagerMockRecorder) GetTasks(ctx, pollingCluster, lastReadTaskID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTasks", reflect.TypeOf((*MockTaskAckManager)(nil).GetTasks), ctx, pollingCluster, lastReadTaskID)
}
