// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/cdn/cdn_service.go

// Package mock_cdn is a generated GoMock package.
package mock_cdn

import (
	dto "animakuro/cdn/internal/cdn/dto"
	entities "animakuro/cdn/internal/entities"
	formdata "animakuro/cdn/internal/formdata"
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockService is a mock of Service interface.
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
}

// MockServiceMockRecorder is the mock recorder for MockService.
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance.
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// DeleteAll mocks base method.
func (m *MockService) DeleteAll(path string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAll", path)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAll indicates an expected call of DeleteAll.
func (mr *MockServiceMockRecorder) DeleteAll(path interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAll", reflect.TypeOf((*MockService)(nil).DeleteAll), path)
}

// DeleteFileDB mocks base method.
func (m *MockService) DeleteFileDB(ctx context.Context, bucket, uuid string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteFileDB", ctx, bucket, uuid)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteFileDB indicates an expected call of DeleteFileDB.
func (mr *MockServiceMockRecorder) DeleteFileDB(ctx, bucket, uuid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteFileDB", reflect.TypeOf((*MockService)(nil).DeleteFileDB), ctx, bucket, uuid)
}

// GetAllBucketsDB mocks base method.
func (m *MockService) GetAllBucketsDB(ctx context.Context) ([]*entities.Bucket, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllBucketsDB", ctx)
	ret0, _ := ret[0].([]*entities.Bucket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllBucketsDB indicates an expected call of GetAllBucketsDB.
func (mr *MockServiceMockRecorder) GetAllBucketsDB(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllBucketsDB", reflect.TypeOf((*MockService)(nil).GetAllBucketsDB), ctx)
}

// GetBucketDB mocks base method.
func (m *MockService) GetBucketDB(ctx context.Context, bucketName string) (*entities.Bucket, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBucketDB", ctx, bucketName)
	ret0, _ := ret[0].(*entities.Bucket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBucketDB indicates an expected call of GetBucketDB.
func (mr *MockServiceMockRecorder) GetBucketDB(ctx, bucketName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBucketDB", reflect.TypeOf((*MockService)(nil).GetBucketDB), ctx, bucketName)
}

// GetFileDB mocks base method.
func (m *MockService) GetFileDB(ctx context.Context, bucket, uuid string) (*entities.File, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFileDB", ctx, bucket, uuid)
	ret0, _ := ret[0].(*entities.File)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFileDB indicates an expected call of GetFileDB.
func (mr *MockServiceMockRecorder) GetFileDB(ctx, bucket, uuid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFileDB", reflect.TypeOf((*MockService)(nil).GetFileDB), ctx, bucket, uuid)
}

// InitBuckets mocks base method.
func (m *MockService) InitBuckets(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InitBuckets", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// InitBuckets indicates an expected call of InitBuckets.
func (mr *MockServiceMockRecorder) InitBuckets(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitBuckets", reflect.TypeOf((*MockService)(nil).InitBuckets), ctx)
}

// MustSave mocks base method.
func (m *MockService) MustSave(buff []byte, path string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "MustSave", buff, path)
}

// MustSave indicates an expected call of MustSave.
func (mr *MockServiceMockRecorder) MustSave(buff, path interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MustSave", reflect.TypeOf((*MockService)(nil).MustSave), buff, path)
}

// ParseMime mocks base method.
func (m *MockService) ParseMime(buff []byte) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParseMime", buff)
	ret0, _ := ret[0].(string)
	return ret0
}

// ParseMime indicates an expected call of ParseMime.
func (mr *MockServiceMockRecorder) ParseMime(buff interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParseMime", reflect.TypeOf((*MockService)(nil).ParseMime), buff)
}

// ReadFile mocks base method.
func (m *MockService) ReadFile(isOrig bool, path string, hosts []string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadFile", isOrig, path, hosts)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadFile indicates an expected call of ReadFile.
func (mr *MockServiceMockRecorder) ReadFile(isOrig, path, hosts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadFile", reflect.TypeOf((*MockService)(nil).ReadFile), isOrig, path, hosts)
}

// SaveBucketDB mocks base method.
func (m *MockService) SaveBucketDB(ctx context.Context, dto dto.CreateBucketDto) (*entities.Bucket, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveBucketDB", ctx, dto)
	ret0, _ := ret[0].(*entities.Bucket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SaveBucketDB indicates an expected call of SaveBucketDB.
func (mr *MockServiceMockRecorder) SaveBucketDB(ctx, dto interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveBucketDB", reflect.TypeOf((*MockService)(nil).SaveBucketDB), ctx, dto)
}

// SaveFileDB mocks base method.
func (m *MockService) SaveFileDB(ctx context.Context, dto dto.SaveFileDto) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveFileDB", ctx, dto)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveFileDB indicates an expected call of SaveFileDB.
func (mr *MockServiceMockRecorder) SaveFileDB(ctx, dto interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveFileDB", reflect.TypeOf((*MockService)(nil).SaveFileDB), ctx, dto)
}

// TryDeleteLocally mocks base method.
func (m *MockService) TryDeleteLocally(dirPath string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "TryDeleteLocally", dirPath)
}

// TryDeleteLocally indicates an expected call of TryDeleteLocally.
func (mr *MockServiceMockRecorder) TryDeleteLocally(dirPath interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TryDeleteLocally", reflect.TypeOf((*MockService)(nil).TryDeleteLocally), dirPath)
}

// TryReadExisting mocks base method.
func (m *MockService) TryReadExisting(path string) ([]byte, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TryReadExisting", path)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// TryReadExisting indicates an expected call of TryReadExisting.
func (mr *MockServiceMockRecorder) TryReadExisting(path interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TryReadExisting", reflect.TypeOf((*MockService)(nil).TryReadExisting), path)
}

// UploadMany mocks base method.
func (m *MockService) UploadMany(ctx context.Context, bucket string, files []*formdata.UploadFile) ([]string, []string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UploadMany", ctx, bucket, files)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].([]string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// UploadMany indicates an expected call of UploadMany.
func (mr *MockServiceMockRecorder) UploadMany(ctx, bucket, files interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadMany", reflect.TypeOf((*MockService)(nil).UploadMany), ctx, bucket, files)
}
