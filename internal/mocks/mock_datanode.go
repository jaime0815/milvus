// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

	commonpb "github.com/milvus-io/milvus/api/commonpb"

	datapb "github.com/milvus-io/milvus/internal/proto/datapb"

	internalpb "github.com/milvus-io/milvus/internal/proto/internalpb"

	milvuspb "github.com/milvus-io/milvus/api/milvuspb"

	mock "github.com/stretchr/testify/mock"
)

// DataNode is an autogenerated mock type for the DataNode type
type DataNode struct {
	mock.Mock
}

type DataNode_Expecter struct {
	mock *mock.Mock
}

func (_m *DataNode) EXPECT() *DataNode_Expecter {
	return &DataNode_Expecter{mock: &_m.Mock}
}

// AddSegment provides a mock function with given fields: ctx, req
func (_m *DataNode) AddSegment(ctx context.Context, req *datapb.AddSegmentRequest) (*commonpb.Status, error) {
	ret := _m.Called(ctx, req)

	var r0 *commonpb.Status
	if rf, ok := ret.Get(0).(func(context.Context, *datapb.AddSegmentRequest) *commonpb.Status); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*commonpb.Status)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datapb.AddSegmentRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DataNode_AddSegment_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddSegment'
type DataNode_AddSegment_Call struct {
	*mock.Call
}

// AddSegment is a helper method to define mock.On call
//  - ctx context.Context
//  - req *datapb.AddSegmentRequest
func (_e *DataNode_Expecter) AddSegment(ctx interface{}, req interface{}) *DataNode_AddSegment_Call {
	return &DataNode_AddSegment_Call{Call: _e.mock.On("AddSegment", ctx, req)}
}

func (_c *DataNode_AddSegment_Call) Run(run func(ctx context.Context, req *datapb.AddSegmentRequest)) *DataNode_AddSegment_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*datapb.AddSegmentRequest))
	})
	return _c
}

func (_c *DataNode_AddSegment_Call) Return(_a0 *commonpb.Status, _a1 error) *DataNode_AddSegment_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Compaction provides a mock function with given fields: ctx, req
func (_m *DataNode) Compaction(ctx context.Context, req *datapb.CompactionPlan) (*commonpb.Status, error) {
	ret := _m.Called(ctx, req)

	var r0 *commonpb.Status
	if rf, ok := ret.Get(0).(func(context.Context, *datapb.CompactionPlan) *commonpb.Status); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*commonpb.Status)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datapb.CompactionPlan) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DataNode_Compaction_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Compaction'
type DataNode_Compaction_Call struct {
	*mock.Call
}

// Compaction is a helper method to define mock.On call
//  - ctx context.Context
//  - req *datapb.CompactionPlan
func (_e *DataNode_Expecter) Compaction(ctx interface{}, req interface{}) *DataNode_Compaction_Call {
	return &DataNode_Compaction_Call{Call: _e.mock.On("Compaction", ctx, req)}
}

func (_c *DataNode_Compaction_Call) Run(run func(ctx context.Context, req *datapb.CompactionPlan)) *DataNode_Compaction_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*datapb.CompactionPlan))
	})
	return _c
}

func (_c *DataNode_Compaction_Call) Return(_a0 *commonpb.Status, _a1 error) *DataNode_Compaction_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// FlushSegments provides a mock function with given fields: ctx, req
func (_m *DataNode) FlushSegments(ctx context.Context, req *datapb.FlushSegmentsRequest) (*commonpb.Status, error) {
	ret := _m.Called(ctx, req)

	var r0 *commonpb.Status
	if rf, ok := ret.Get(0).(func(context.Context, *datapb.FlushSegmentsRequest) *commonpb.Status); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*commonpb.Status)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datapb.FlushSegmentsRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DataNode_FlushSegments_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FlushSegments'
type DataNode_FlushSegments_Call struct {
	*mock.Call
}

// FlushSegments is a helper method to define mock.On call
//  - ctx context.Context
//  - req *datapb.FlushSegmentsRequest
func (_e *DataNode_Expecter) FlushSegments(ctx interface{}, req interface{}) *DataNode_FlushSegments_Call {
	return &DataNode_FlushSegments_Call{Call: _e.mock.On("FlushSegments", ctx, req)}
}

func (_c *DataNode_FlushSegments_Call) Run(run func(ctx context.Context, req *datapb.FlushSegmentsRequest)) *DataNode_FlushSegments_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*datapb.FlushSegmentsRequest))
	})
	return _c
}

func (_c *DataNode_FlushSegments_Call) Return(_a0 *commonpb.Status, _a1 error) *DataNode_FlushSegments_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetCompactionState provides a mock function with given fields: ctx, req
func (_m *DataNode) GetCompactionState(ctx context.Context, req *datapb.CompactionStateRequest) (*datapb.CompactionStateResponse, error) {
	ret := _m.Called(ctx, req)

	var r0 *datapb.CompactionStateResponse
	if rf, ok := ret.Get(0).(func(context.Context, *datapb.CompactionStateRequest) *datapb.CompactionStateResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*datapb.CompactionStateResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datapb.CompactionStateRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DataNode_GetCompactionState_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetCompactionState'
type DataNode_GetCompactionState_Call struct {
	*mock.Call
}

// GetCompactionState is a helper method to define mock.On call
//  - ctx context.Context
//  - req *datapb.CompactionStateRequest
func (_e *DataNode_Expecter) GetCompactionState(ctx interface{}, req interface{}) *DataNode_GetCompactionState_Call {
	return &DataNode_GetCompactionState_Call{Call: _e.mock.On("GetCompactionState", ctx, req)}
}

func (_c *DataNode_GetCompactionState_Call) Run(run func(ctx context.Context, req *datapb.CompactionStateRequest)) *DataNode_GetCompactionState_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*datapb.CompactionStateRequest))
	})
	return _c
}

func (_c *DataNode_GetCompactionState_Call) Return(_a0 *datapb.CompactionStateResponse, _a1 error) *DataNode_GetCompactionState_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetComponentStates provides a mock function with given fields: ctx
func (_m *DataNode) GetComponentStates(ctx context.Context) (*internalpb.ComponentStates, error) {
	ret := _m.Called(ctx)

	var r0 *internalpb.ComponentStates
	if rf, ok := ret.Get(0).(func(context.Context) *internalpb.ComponentStates); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*internalpb.ComponentStates)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DataNode_GetComponentStates_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetComponentStates'
type DataNode_GetComponentStates_Call struct {
	*mock.Call
}

// GetComponentStates is a helper method to define mock.On call
//  - ctx context.Context
func (_e *DataNode_Expecter) GetComponentStates(ctx interface{}) *DataNode_GetComponentStates_Call {
	return &DataNode_GetComponentStates_Call{Call: _e.mock.On("GetComponentStates", ctx)}
}

func (_c *DataNode_GetComponentStates_Call) Run(run func(ctx context.Context)) *DataNode_GetComponentStates_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *DataNode_GetComponentStates_Call) Return(_a0 *internalpb.ComponentStates, _a1 error) *DataNode_GetComponentStates_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetMetrics provides a mock function with given fields: ctx, req
func (_m *DataNode) GetMetrics(ctx context.Context, req *milvuspb.GetMetricsRequest) (*milvuspb.GetMetricsResponse, error) {
	ret := _m.Called(ctx, req)

	var r0 *milvuspb.GetMetricsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *milvuspb.GetMetricsRequest) *milvuspb.GetMetricsResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*milvuspb.GetMetricsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *milvuspb.GetMetricsRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DataNode_GetMetrics_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetMetrics'
type DataNode_GetMetrics_Call struct {
	*mock.Call
}

// GetMetrics is a helper method to define mock.On call
//  - ctx context.Context
//  - req *milvuspb.GetMetricsRequest
func (_e *DataNode_Expecter) GetMetrics(ctx interface{}, req interface{}) *DataNode_GetMetrics_Call {
	return &DataNode_GetMetrics_Call{Call: _e.mock.On("GetMetrics", ctx, req)}
}

func (_c *DataNode_GetMetrics_Call) Run(run func(ctx context.Context, req *milvuspb.GetMetricsRequest)) *DataNode_GetMetrics_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*milvuspb.GetMetricsRequest))
	})
	return _c
}

func (_c *DataNode_GetMetrics_Call) Return(_a0 *milvuspb.GetMetricsResponse, _a1 error) *DataNode_GetMetrics_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetStatisticsChannel provides a mock function with given fields: ctx
func (_m *DataNode) GetStatisticsChannel(ctx context.Context) (*milvuspb.StringResponse, error) {
	ret := _m.Called(ctx)

	var r0 *milvuspb.StringResponse
	if rf, ok := ret.Get(0).(func(context.Context) *milvuspb.StringResponse); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*milvuspb.StringResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DataNode_GetStatisticsChannel_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetStatisticsChannel'
type DataNode_GetStatisticsChannel_Call struct {
	*mock.Call
}

// GetStatisticsChannel is a helper method to define mock.On call
//  - ctx context.Context
func (_e *DataNode_Expecter) GetStatisticsChannel(ctx interface{}) *DataNode_GetStatisticsChannel_Call {
	return &DataNode_GetStatisticsChannel_Call{Call: _e.mock.On("GetStatisticsChannel", ctx)}
}

func (_c *DataNode_GetStatisticsChannel_Call) Run(run func(ctx context.Context)) *DataNode_GetStatisticsChannel_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *DataNode_GetStatisticsChannel_Call) Return(_a0 *milvuspb.StringResponse, _a1 error) *DataNode_GetStatisticsChannel_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Import provides a mock function with given fields: ctx, req
func (_m *DataNode) Import(ctx context.Context, req *datapb.ImportTaskRequest) (*commonpb.Status, error) {
	ret := _m.Called(ctx, req)

	var r0 *commonpb.Status
	if rf, ok := ret.Get(0).(func(context.Context, *datapb.ImportTaskRequest) *commonpb.Status); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*commonpb.Status)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datapb.ImportTaskRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DataNode_Import_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Import'
type DataNode_Import_Call struct {
	*mock.Call
}

// Import is a helper method to define mock.On call
//  - ctx context.Context
//  - req *datapb.ImportTaskRequest
func (_e *DataNode_Expecter) Import(ctx interface{}, req interface{}) *DataNode_Import_Call {
	return &DataNode_Import_Call{Call: _e.mock.On("Import", ctx, req)}
}

func (_c *DataNode_Import_Call) Run(run func(ctx context.Context, req *datapb.ImportTaskRequest)) *DataNode_Import_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*datapb.ImportTaskRequest))
	})
	return _c
}

func (_c *DataNode_Import_Call) Return(_a0 *commonpb.Status, _a1 error) *DataNode_Import_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Init provides a mock function with given fields:
func (_m *DataNode) Init() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DataNode_Init_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Init'
type DataNode_Init_Call struct {
	*mock.Call
}

// Init is a helper method to define mock.On call
func (_e *DataNode_Expecter) Init() *DataNode_Init_Call {
	return &DataNode_Init_Call{Call: _e.mock.On("Init")}
}

func (_c *DataNode_Init_Call) Run(run func()) *DataNode_Init_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *DataNode_Init_Call) Return(_a0 error) *DataNode_Init_Call {
	_c.Call.Return(_a0)
	return _c
}

// Register provides a mock function with given fields:
func (_m *DataNode) Register() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DataNode_Register_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Register'
type DataNode_Register_Call struct {
	*mock.Call
}

// Register is a helper method to define mock.On call
func (_e *DataNode_Expecter) Register() *DataNode_Register_Call {
	return &DataNode_Register_Call{Call: _e.mock.On("Register")}
}

func (_c *DataNode_Register_Call) Run(run func()) *DataNode_Register_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *DataNode_Register_Call) Return(_a0 error) *DataNode_Register_Call {
	_c.Call.Return(_a0)
	return _c
}

// ResendSegmentStats provides a mock function with given fields: ctx, req
func (_m *DataNode) ResendSegmentStats(ctx context.Context, req *datapb.ResendSegmentStatsRequest) (*datapb.ResendSegmentStatsResponse, error) {
	ret := _m.Called(ctx, req)

	var r0 *datapb.ResendSegmentStatsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *datapb.ResendSegmentStatsRequest) *datapb.ResendSegmentStatsResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*datapb.ResendSegmentStatsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datapb.ResendSegmentStatsRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DataNode_ResendSegmentStats_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResendSegmentStats'
type DataNode_ResendSegmentStats_Call struct {
	*mock.Call
}

// ResendSegmentStats is a helper method to define mock.On call
//  - ctx context.Context
//  - req *datapb.ResendSegmentStatsRequest
func (_e *DataNode_Expecter) ResendSegmentStats(ctx interface{}, req interface{}) *DataNode_ResendSegmentStats_Call {
	return &DataNode_ResendSegmentStats_Call{Call: _e.mock.On("ResendSegmentStats", ctx, req)}
}

func (_c *DataNode_ResendSegmentStats_Call) Run(run func(ctx context.Context, req *datapb.ResendSegmentStatsRequest)) *DataNode_ResendSegmentStats_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*datapb.ResendSegmentStatsRequest))
	})
	return _c
}

func (_c *DataNode_ResendSegmentStats_Call) Return(_a0 *datapb.ResendSegmentStatsResponse, _a1 error) *DataNode_ResendSegmentStats_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// ShowConfigurations provides a mock function with given fields: ctx, req
func (_m *DataNode) ShowConfigurations(ctx context.Context, req *internalpb.ShowConfigurationsRequest) (*internalpb.ShowConfigurationsResponse, error) {
	ret := _m.Called(ctx, req)

	var r0 *internalpb.ShowConfigurationsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *internalpb.ShowConfigurationsRequest) *internalpb.ShowConfigurationsResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*internalpb.ShowConfigurationsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *internalpb.ShowConfigurationsRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DataNode_ShowConfigurations_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ShowConfigurations'
type DataNode_ShowConfigurations_Call struct {
	*mock.Call
}

// ShowConfigurations is a helper method to define mock.On call
//  - ctx context.Context
//  - req *internalpb.ShowConfigurationsRequest
func (_e *DataNode_Expecter) ShowConfigurations(ctx interface{}, req interface{}) *DataNode_ShowConfigurations_Call {
	return &DataNode_ShowConfigurations_Call{Call: _e.mock.On("ShowConfigurations", ctx, req)}
}

func (_c *DataNode_ShowConfigurations_Call) Run(run func(ctx context.Context, req *internalpb.ShowConfigurationsRequest)) *DataNode_ShowConfigurations_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*internalpb.ShowConfigurationsRequest))
	})
	return _c
}

func (_c *DataNode_ShowConfigurations_Call) Return(_a0 *internalpb.ShowConfigurationsResponse, _a1 error) *DataNode_ShowConfigurations_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Start provides a mock function with given fields:
func (_m *DataNode) Start() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DataNode_Start_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Start'
type DataNode_Start_Call struct {
	*mock.Call
}

// Start is a helper method to define mock.On call
func (_e *DataNode_Expecter) Start() *DataNode_Start_Call {
	return &DataNode_Start_Call{Call: _e.mock.On("Start")}
}

func (_c *DataNode_Start_Call) Run(run func()) *DataNode_Start_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *DataNode_Start_Call) Return(_a0 error) *DataNode_Start_Call {
	_c.Call.Return(_a0)
	return _c
}

// Stop provides a mock function with given fields:
func (_m *DataNode) Stop() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DataNode_Stop_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Stop'
type DataNode_Stop_Call struct {
	*mock.Call
}

// Stop is a helper method to define mock.On call
func (_e *DataNode_Expecter) Stop() *DataNode_Stop_Call {
	return &DataNode_Stop_Call{Call: _e.mock.On("Stop")}
}

func (_c *DataNode_Stop_Call) Run(run func()) *DataNode_Stop_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *DataNode_Stop_Call) Return(_a0 error) *DataNode_Stop_Call {
	_c.Call.Return(_a0)
	return _c
}

// SyncSegments provides a mock function with given fields: ctx, req
func (_m *DataNode) SyncSegments(ctx context.Context, req *datapb.SyncSegmentsRequest) (*commonpb.Status, error) {
	ret := _m.Called(ctx, req)

	var r0 *commonpb.Status
	if rf, ok := ret.Get(0).(func(context.Context, *datapb.SyncSegmentsRequest) *commonpb.Status); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*commonpb.Status)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datapb.SyncSegmentsRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DataNode_SyncSegments_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SyncSegments'
type DataNode_SyncSegments_Call struct {
	*mock.Call
}

// SyncSegments is a helper method to define mock.On call
//  - ctx context.Context
//  - req *datapb.SyncSegmentsRequest
func (_e *DataNode_Expecter) SyncSegments(ctx interface{}, req interface{}) *DataNode_SyncSegments_Call {
	return &DataNode_SyncSegments_Call{Call: _e.mock.On("SyncSegments", ctx, req)}
}

func (_c *DataNode_SyncSegments_Call) Run(run func(ctx context.Context, req *datapb.SyncSegmentsRequest)) *DataNode_SyncSegments_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*datapb.SyncSegmentsRequest))
	})
	return _c
}

func (_c *DataNode_SyncSegments_Call) Return(_a0 *commonpb.Status, _a1 error) *DataNode_SyncSegments_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// WatchDmChannels provides a mock function with given fields: ctx, req
func (_m *DataNode) WatchDmChannels(ctx context.Context, req *datapb.WatchDmChannelsRequest) (*commonpb.Status, error) {
	ret := _m.Called(ctx, req)

	var r0 *commonpb.Status
	if rf, ok := ret.Get(0).(func(context.Context, *datapb.WatchDmChannelsRequest) *commonpb.Status); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*commonpb.Status)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datapb.WatchDmChannelsRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DataNode_WatchDmChannels_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WatchDmChannels'
type DataNode_WatchDmChannels_Call struct {
	*mock.Call
}

// WatchDmChannels is a helper method to define mock.On call
//  - ctx context.Context
//  - req *datapb.WatchDmChannelsRequest
func (_e *DataNode_Expecter) WatchDmChannels(ctx interface{}, req interface{}) *DataNode_WatchDmChannels_Call {
	return &DataNode_WatchDmChannels_Call{Call: _e.mock.On("WatchDmChannels", ctx, req)}
}

func (_c *DataNode_WatchDmChannels_Call) Run(run func(ctx context.Context, req *datapb.WatchDmChannelsRequest)) *DataNode_WatchDmChannels_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*datapb.WatchDmChannelsRequest))
	})
	return _c
}

func (_c *DataNode_WatchDmChannels_Call) Return(_a0 *commonpb.Status, _a1 error) *DataNode_WatchDmChannels_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

type mockConstructorTestingTNewDataNode interface {
	mock.TestingT
	Cleanup(func())
}

// NewDataNode creates a new instance of DataNode. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewDataNode(t mockConstructorTestingTNewDataNode) *DataNode {
	mock := &DataNode{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
