// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	mmap "golang.org/x/exp/mmap"

	storage "github.com/milvus-io/milvus/internal/storage"

	time "time"
)

// ChunkManager is an autogenerated mock type for the ChunkManager type
type ChunkManager struct {
	mock.Mock
}

type ChunkManager_Expecter struct {
	mock *mock.Mock
}

func (_m *ChunkManager) EXPECT() *ChunkManager_Expecter {
	return &ChunkManager_Expecter{mock: &_m.Mock}
}

// Exist provides a mock function with given fields: filePath
func (_m *ChunkManager) Exist(filePath string) (bool, error) {
	ret := _m.Called(filePath)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(filePath)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(filePath)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ChunkManager_Exist_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Exist'
type ChunkManager_Exist_Call struct {
	*mock.Call
}

// Exist is a helper method to define mock.On call
//  - filePath string
func (_e *ChunkManager_Expecter) Exist(filePath interface{}) *ChunkManager_Exist_Call {
	return &ChunkManager_Exist_Call{Call: _e.mock.On("Exist", filePath)}
}

func (_c *ChunkManager_Exist_Call) Run(run func(filePath string)) *ChunkManager_Exist_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *ChunkManager_Exist_Call) Return(_a0 bool, _a1 error) *ChunkManager_Exist_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// ListWithPrefix provides a mock function with given fields: prefix, recursive
func (_m *ChunkManager) ListWithPrefix(prefix string, recursive bool) ([]string, []time.Time, error) {
	ret := _m.Called(prefix, recursive)

	var r0 []string
	if rf, ok := ret.Get(0).(func(string, bool) []string); ok {
		r0 = rf(prefix, recursive)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 []time.Time
	if rf, ok := ret.Get(1).(func(string, bool) []time.Time); ok {
		r1 = rf(prefix, recursive)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]time.Time)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(string, bool) error); ok {
		r2 = rf(prefix, recursive)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// ChunkManager_ListWithPrefix_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListWithPrefix'
type ChunkManager_ListWithPrefix_Call struct {
	*mock.Call
}

// ListWithPrefix is a helper method to define mock.On call
//  - prefix string
//  - recursive bool
func (_e *ChunkManager_Expecter) ListWithPrefix(prefix interface{}, recursive interface{}) *ChunkManager_ListWithPrefix_Call {
	return &ChunkManager_ListWithPrefix_Call{Call: _e.mock.On("ListWithPrefix", prefix, recursive)}
}

func (_c *ChunkManager_ListWithPrefix_Call) Run(run func(prefix string, recursive bool)) *ChunkManager_ListWithPrefix_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(bool))
	})
	return _c
}

func (_c *ChunkManager_ListWithPrefix_Call) Return(_a0 []string, _a1 []time.Time, _a2 error) *ChunkManager_ListWithPrefix_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

// Mmap provides a mock function with given fields: filePath
func (_m *ChunkManager) Mmap(filePath string) (*mmap.ReaderAt, error) {
	ret := _m.Called(filePath)

	var r0 *mmap.ReaderAt
	if rf, ok := ret.Get(0).(func(string) *mmap.ReaderAt); ok {
		r0 = rf(filePath)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*mmap.ReaderAt)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(filePath)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ChunkManager_Mmap_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Mmap'
type ChunkManager_Mmap_Call struct {
	*mock.Call
}

// Mmap is a helper method to define mock.On call
//  - filePath string
func (_e *ChunkManager_Expecter) Mmap(filePath interface{}) *ChunkManager_Mmap_Call {
	return &ChunkManager_Mmap_Call{Call: _e.mock.On("Mmap", filePath)}
}

func (_c *ChunkManager_Mmap_Call) Run(run func(filePath string)) *ChunkManager_Mmap_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *ChunkManager_Mmap_Call) Return(_a0 *mmap.ReaderAt, _a1 error) *ChunkManager_Mmap_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// MultiRead provides a mock function with given fields: filePaths
func (_m *ChunkManager) MultiRead(filePaths []string) ([][]byte, error) {
	ret := _m.Called(filePaths)

	var r0 [][]byte
	if rf, ok := ret.Get(0).(func([]string) [][]byte); ok {
		r0 = rf(filePaths)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([][]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]string) error); ok {
		r1 = rf(filePaths)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ChunkManager_MultiRead_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MultiRead'
type ChunkManager_MultiRead_Call struct {
	*mock.Call
}

// MultiRead is a helper method to define mock.On call
//  - filePaths []string
func (_e *ChunkManager_Expecter) MultiRead(filePaths interface{}) *ChunkManager_MultiRead_Call {
	return &ChunkManager_MultiRead_Call{Call: _e.mock.On("MultiRead", filePaths)}
}

func (_c *ChunkManager_MultiRead_Call) Run(run func(filePaths []string)) *ChunkManager_MultiRead_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]string))
	})
	return _c
}

func (_c *ChunkManager_MultiRead_Call) Return(_a0 [][]byte, _a1 error) *ChunkManager_MultiRead_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// MultiRemove provides a mock function with given fields: filePaths
func (_m *ChunkManager) MultiRemove(filePaths []string) error {
	ret := _m.Called(filePaths)

	var r0 error
	if rf, ok := ret.Get(0).(func([]string) error); ok {
		r0 = rf(filePaths)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ChunkManager_MultiRemove_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MultiRemove'
type ChunkManager_MultiRemove_Call struct {
	*mock.Call
}

// MultiRemove is a helper method to define mock.On call
//  - filePaths []string
func (_e *ChunkManager_Expecter) MultiRemove(filePaths interface{}) *ChunkManager_MultiRemove_Call {
	return &ChunkManager_MultiRemove_Call{Call: _e.mock.On("MultiRemove", filePaths)}
}

func (_c *ChunkManager_MultiRemove_Call) Run(run func(filePaths []string)) *ChunkManager_MultiRemove_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]string))
	})
	return _c
}

func (_c *ChunkManager_MultiRemove_Call) Return(_a0 error) *ChunkManager_MultiRemove_Call {
	_c.Call.Return(_a0)
	return _c
}

// MultiWrite provides a mock function with given fields: contents
func (_m *ChunkManager) MultiWrite(contents map[string][]byte) error {
	ret := _m.Called(contents)

	var r0 error
	if rf, ok := ret.Get(0).(func(map[string][]byte) error); ok {
		r0 = rf(contents)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ChunkManager_MultiWrite_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MultiWrite'
type ChunkManager_MultiWrite_Call struct {
	*mock.Call
}

// MultiWrite is a helper method to define mock.On call
//  - contents map[string][]byte
func (_e *ChunkManager_Expecter) MultiWrite(contents interface{}) *ChunkManager_MultiWrite_Call {
	return &ChunkManager_MultiWrite_Call{Call: _e.mock.On("MultiWrite", contents)}
}

func (_c *ChunkManager_MultiWrite_Call) Run(run func(contents map[string][]byte)) *ChunkManager_MultiWrite_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(map[string][]byte))
	})
	return _c
}

func (_c *ChunkManager_MultiWrite_Call) Return(_a0 error) *ChunkManager_MultiWrite_Call {
	_c.Call.Return(_a0)
	return _c
}

// Path provides a mock function with given fields: filePath
func (_m *ChunkManager) Path(filePath string) (string, error) {
	ret := _m.Called(filePath)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(filePath)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(filePath)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ChunkManager_Path_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Path'
type ChunkManager_Path_Call struct {
	*mock.Call
}

// Path is a helper method to define mock.On call
//  - filePath string
func (_e *ChunkManager_Expecter) Path(filePath interface{}) *ChunkManager_Path_Call {
	return &ChunkManager_Path_Call{Call: _e.mock.On("Path", filePath)}
}

func (_c *ChunkManager_Path_Call) Run(run func(filePath string)) *ChunkManager_Path_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *ChunkManager_Path_Call) Return(_a0 string, _a1 error) *ChunkManager_Path_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Read provides a mock function with given fields: filePath
func (_m *ChunkManager) Read(filePath string) ([]byte, error) {
	ret := _m.Called(filePath)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string) []byte); ok {
		r0 = rf(filePath)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(filePath)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ChunkManager_Read_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Read'
type ChunkManager_Read_Call struct {
	*mock.Call
}

// Read is a helper method to define mock.On call
//  - filePath string
func (_e *ChunkManager_Expecter) Read(filePath interface{}) *ChunkManager_Read_Call {
	return &ChunkManager_Read_Call{Call: _e.mock.On("Read", filePath)}
}

func (_c *ChunkManager_Read_Call) Run(run func(filePath string)) *ChunkManager_Read_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *ChunkManager_Read_Call) Return(_a0 []byte, _a1 error) *ChunkManager_Read_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// ReadAt provides a mock function with given fields: filePath, off, length
func (_m *ChunkManager) ReadAt(filePath string, off int64, length int64) ([]byte, error) {
	ret := _m.Called(filePath, off, length)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string, int64, int64) []byte); ok {
		r0 = rf(filePath, off, length)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, int64, int64) error); ok {
		r1 = rf(filePath, off, length)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ChunkManager_ReadAt_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ReadAt'
type ChunkManager_ReadAt_Call struct {
	*mock.Call
}

// ReadAt is a helper method to define mock.On call
//  - filePath string
//  - off int64
//  - length int64
func (_e *ChunkManager_Expecter) ReadAt(filePath interface{}, off interface{}, length interface{}) *ChunkManager_ReadAt_Call {
	return &ChunkManager_ReadAt_Call{Call: _e.mock.On("ReadAt", filePath, off, length)}
}

func (_c *ChunkManager_ReadAt_Call) Run(run func(filePath string, off int64, length int64)) *ChunkManager_ReadAt_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(int64), args[2].(int64))
	})
	return _c
}

func (_c *ChunkManager_ReadAt_Call) Return(p []byte, err error) *ChunkManager_ReadAt_Call {
	_c.Call.Return(p, err)
	return _c
}

// ReadWithPrefix provides a mock function with given fields: prefix
func (_m *ChunkManager) ReadWithPrefix(prefix string) ([]string, [][]byte, error) {
	ret := _m.Called(prefix)

	var r0 []string
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(prefix)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 [][]byte
	if rf, ok := ret.Get(1).(func(string) [][]byte); ok {
		r1 = rf(prefix)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([][]byte)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(string) error); ok {
		r2 = rf(prefix)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// ChunkManager_ReadWithPrefix_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ReadWithPrefix'
type ChunkManager_ReadWithPrefix_Call struct {
	*mock.Call
}

// ReadWithPrefix is a helper method to define mock.On call
//  - prefix string
func (_e *ChunkManager_Expecter) ReadWithPrefix(prefix interface{}) *ChunkManager_ReadWithPrefix_Call {
	return &ChunkManager_ReadWithPrefix_Call{Call: _e.mock.On("ReadWithPrefix", prefix)}
}

func (_c *ChunkManager_ReadWithPrefix_Call) Run(run func(prefix string)) *ChunkManager_ReadWithPrefix_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *ChunkManager_ReadWithPrefix_Call) Return(_a0 []string, _a1 [][]byte, _a2 error) *ChunkManager_ReadWithPrefix_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

// Reader provides a mock function with given fields: filePath
func (_m *ChunkManager) Reader(filePath string) (storage.FileReader, error) {
	ret := _m.Called(filePath)

	var r0 storage.FileReader
	if rf, ok := ret.Get(0).(func(string) storage.FileReader); ok {
		r0 = rf(filePath)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(storage.FileReader)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(filePath)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ChunkManager_Reader_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Reader'
type ChunkManager_Reader_Call struct {
	*mock.Call
}

// Reader is a helper method to define mock.On call
//  - filePath string
func (_e *ChunkManager_Expecter) Reader(filePath interface{}) *ChunkManager_Reader_Call {
	return &ChunkManager_Reader_Call{Call: _e.mock.On("Reader", filePath)}
}

func (_c *ChunkManager_Reader_Call) Run(run func(filePath string)) *ChunkManager_Reader_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *ChunkManager_Reader_Call) Return(_a0 storage.FileReader, _a1 error) *ChunkManager_Reader_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Remove provides a mock function with given fields: filePath
func (_m *ChunkManager) Remove(filePath string) error {
	ret := _m.Called(filePath)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(filePath)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ChunkManager_Remove_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Remove'
type ChunkManager_Remove_Call struct {
	*mock.Call
}

// Remove is a helper method to define mock.On call
//  - filePath string
func (_e *ChunkManager_Expecter) Remove(filePath interface{}) *ChunkManager_Remove_Call {
	return &ChunkManager_Remove_Call{Call: _e.mock.On("Remove", filePath)}
}

func (_c *ChunkManager_Remove_Call) Run(run func(filePath string)) *ChunkManager_Remove_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *ChunkManager_Remove_Call) Return(_a0 error) *ChunkManager_Remove_Call {
	_c.Call.Return(_a0)
	return _c
}

// RemoveWithPrefix provides a mock function with given fields: prefix
func (_m *ChunkManager) RemoveWithPrefix(prefix string) error {
	ret := _m.Called(prefix)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(prefix)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ChunkManager_RemoveWithPrefix_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RemoveWithPrefix'
type ChunkManager_RemoveWithPrefix_Call struct {
	*mock.Call
}

// RemoveWithPrefix is a helper method to define mock.On call
//  - prefix string
func (_e *ChunkManager_Expecter) RemoveWithPrefix(prefix interface{}) *ChunkManager_RemoveWithPrefix_Call {
	return &ChunkManager_RemoveWithPrefix_Call{Call: _e.mock.On("RemoveWithPrefix", prefix)}
}

func (_c *ChunkManager_RemoveWithPrefix_Call) Run(run func(prefix string)) *ChunkManager_RemoveWithPrefix_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *ChunkManager_RemoveWithPrefix_Call) Return(_a0 error) *ChunkManager_RemoveWithPrefix_Call {
	_c.Call.Return(_a0)
	return _c
}

// RootPath provides a mock function with given fields:
func (_m *ChunkManager) RootPath() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ChunkManager_RootPath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RootPath'
type ChunkManager_RootPath_Call struct {
	*mock.Call
}

// RootPath is a helper method to define mock.On call
func (_e *ChunkManager_Expecter) RootPath() *ChunkManager_RootPath_Call {
	return &ChunkManager_RootPath_Call{Call: _e.mock.On("RootPath")}
}

func (_c *ChunkManager_RootPath_Call) Run(run func()) *ChunkManager_RootPath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ChunkManager_RootPath_Call) Return(_a0 string) *ChunkManager_RootPath_Call {
	_c.Call.Return(_a0)
	return _c
}

// Size provides a mock function with given fields: filePath
func (_m *ChunkManager) Size(filePath string) (int64, error) {
	ret := _m.Called(filePath)

	var r0 int64
	if rf, ok := ret.Get(0).(func(string) int64); ok {
		r0 = rf(filePath)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(filePath)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ChunkManager_Size_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Size'
type ChunkManager_Size_Call struct {
	*mock.Call
}

// Size is a helper method to define mock.On call
//  - filePath string
func (_e *ChunkManager_Expecter) Size(filePath interface{}) *ChunkManager_Size_Call {
	return &ChunkManager_Size_Call{Call: _e.mock.On("Size", filePath)}
}

func (_c *ChunkManager_Size_Call) Run(run func(filePath string)) *ChunkManager_Size_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *ChunkManager_Size_Call) Return(_a0 int64, _a1 error) *ChunkManager_Size_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Write provides a mock function with given fields: filePath, content
func (_m *ChunkManager) Write(filePath string, content []byte) error {
	ret := _m.Called(filePath, content)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, []byte) error); ok {
		r0 = rf(filePath, content)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ChunkManager_Write_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Write'
type ChunkManager_Write_Call struct {
	*mock.Call
}

// Write is a helper method to define mock.On call
//  - filePath string
//  - content []byte
func (_e *ChunkManager_Expecter) Write(filePath interface{}, content interface{}) *ChunkManager_Write_Call {
	return &ChunkManager_Write_Call{Call: _e.mock.On("Write", filePath, content)}
}

func (_c *ChunkManager_Write_Call) Run(run func(filePath string, content []byte)) *ChunkManager_Write_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].([]byte))
	})
	return _c
}

func (_c *ChunkManager_Write_Call) Return(_a0 error) *ChunkManager_Write_Call {
	_c.Call.Return(_a0)
	return _c
}

type mockConstructorTestingTNewChunkManager interface {
	mock.TestingT
	Cleanup(func())
}

// NewChunkManager creates a new instance of ChunkManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewChunkManager(t mockConstructorTestingTNewChunkManager) *ChunkManager {
	mock := &ChunkManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
