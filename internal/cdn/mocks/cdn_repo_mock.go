// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/cdn/cdn_repo.go

// Package mock_cdn is a generated GoMock package.
package mock_cdn

import (
	dto "animakuro/cdn/internal/cdn/dto"
	entities "animakuro/cdn/internal/entities"
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockRepository is a mock of Repository interface.
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance.
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// DeleteFile mocks base method.
func (m *MockRepository) DeleteFile(ctx context.Context, bucket, uuid string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteFile", ctx, bucket, uuid)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteFile indicates an expected call of DeleteFile.
func (mr *MockRepositoryMockRecorder) DeleteFile(ctx, bucket, uuid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteFile", reflect.TypeOf((*MockRepository)(nil).DeleteFile), ctx, bucket, uuid)
}

// GetAllBuckets mocks base method.
func (m *MockRepository) GetAllBuckets(ctx context.Context) ([]*entities.Bucket, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllBuckets", ctx)
	ret0, _ := ret[0].([]*entities.Bucket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllBuckets indicates an expected call of GetAllBuckets.
func (mr *MockRepositoryMockRecorder) GetAllBuckets(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllBuckets", reflect.TypeOf((*MockRepository)(nil).GetAllBuckets), ctx)
}

// GetBucket mocks base method.
func (m *MockRepository) GetBucket(ctx context.Context, name string) (*entities.Bucket, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBucket", ctx, name)
	ret0, _ := ret[0].(*entities.Bucket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBucket indicates an expected call of GetBucket.
func (mr *MockRepositoryMockRecorder) GetBucket(ctx, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBucket", reflect.TypeOf((*MockRepository)(nil).GetBucket), ctx, name)
}

// GetFile mocks base method.
func (m *MockRepository) GetFile(ctx context.Context, bucket, uuid string) (*entities.File, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFile", ctx, bucket, uuid)
	ret0, _ := ret[0].(*entities.File)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFile indicates an expected call of GetFile.
func (mr *MockRepositoryMockRecorder) GetFile(ctx, bucket, uuid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFile", reflect.TypeOf((*MockRepository)(nil).GetFile), ctx, bucket, uuid)
}

// SaveBucket mocks base method.
func (m *MockRepository) SaveBucket(ctx context.Context, dto dto.CreateBucketDto) (*entities.Bucket, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveBucket", ctx, dto)
	ret0, _ := ret[0].(*entities.Bucket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SaveBucket indicates an expected call of SaveBucket.
func (mr *MockRepositoryMockRecorder) SaveBucket(ctx, dto interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveBucket", reflect.TypeOf((*MockRepository)(nil).SaveBucket), ctx, dto)
}

// SaveFile mocks base method.
func (m *MockRepository) SaveFile(ctx context.Context, dto dto.SaveFileDto) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveFile", ctx, dto)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SaveFile indicates an expected call of SaveFile.
func (mr *MockRepositoryMockRecorder) SaveFile(ctx, dto interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveFile", reflect.TypeOf((*MockRepository)(nil).SaveFile), ctx, dto)
}