// Code generated by mockery v2.16.0. DO NOT EDIT.

package mockrootcoord

import (
	context "context"

	etcdpb "github.com/milvus-io/milvus/internal/proto/etcdpb"
	internalpb "github.com/milvus-io/milvus/internal/proto/internalpb"

	milvuspb "github.com/milvus-io/milvus-proto/go-api/milvuspb"

	mock "github.com/stretchr/testify/mock"

	model "github.com/milvus-io/milvus/internal/metastore/model"
)

// IMetaTable is an autogenerated mock type for the IMetaTable type
type IMetaTable struct {
	mock.Mock
}

// AddCollection provides a mock function with given fields: ctx, coll
func (_m *IMetaTable) AddCollection(ctx context.Context, coll *model.Collection) error {
	ret := _m.Called(ctx, coll)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.Collection) error); ok {
		r0 = rf(ctx, coll)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AddCredential provides a mock function with given fields: credInfo
func (_m *IMetaTable) AddCredential(credInfo *internalpb.CredentialInfo) error {
	ret := _m.Called(credInfo)

	var r0 error
	if rf, ok := ret.Get(0).(func(*internalpb.CredentialInfo) error); ok {
		r0 = rf(credInfo)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AddPartition provides a mock function with given fields: ctx, partition
func (_m *IMetaTable) AddPartition(ctx context.Context, partition *model.Partition) error {
	ret := _m.Called(ctx, partition)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.Partition) error); ok {
		r0 = rf(ctx, partition)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AlterAlias provides a mock function with given fields: ctx, dbName, alias, collectionName, ts
func (_m *IMetaTable) AlterAlias(ctx context.Context, dbName string, alias string, collectionName string, ts uint64) error {
	ret := _m.Called(ctx, dbName, alias, collectionName, ts)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, uint64) error); ok {
		r0 = rf(ctx, dbName, alias, collectionName, ts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AlterCollection provides a mock function with given fields: ctx, oldColl, newColl, ts
func (_m *IMetaTable) AlterCollection(ctx context.Context, oldColl *model.Collection, newColl *model.Collection, ts uint64) error {
	ret := _m.Called(ctx, oldColl, newColl, ts)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.Collection, *model.Collection, uint64) error); ok {
		r0 = rf(ctx, oldColl, newColl, ts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AlterCredential provides a mock function with given fields: credInfo
func (_m *IMetaTable) AlterCredential(credInfo *internalpb.CredentialInfo) error {
	ret := _m.Called(credInfo)

	var r0 error
	if rf, ok := ret.Get(0).(func(*internalpb.CredentialInfo) error); ok {
		r0 = rf(credInfo)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ChangeCollectionState provides a mock function with given fields: ctx, collectionID, state, ts
func (_m *IMetaTable) ChangeCollectionState(ctx context.Context, collectionID int64, state etcdpb.CollectionState, ts uint64) error {
	ret := _m.Called(ctx, collectionID, state, ts)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, etcdpb.CollectionState, uint64) error); ok {
		r0 = rf(ctx, collectionID, state, ts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ChangePartitionState provides a mock function with given fields: ctx, collectionID, partitionID, state, ts
func (_m *IMetaTable) ChangePartitionState(ctx context.Context, collectionID int64, partitionID int64, state etcdpb.PartitionState, ts uint64) error {
	ret := _m.Called(ctx, collectionID, partitionID, state, ts)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, int64, etcdpb.PartitionState, uint64) error); ok {
		r0 = rf(ctx, collectionID, partitionID, state, ts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateAlias provides a mock function with given fields: ctx, dbName, alias, collectionName, ts
func (_m *IMetaTable) CreateAlias(ctx context.Context, dbName string, alias string, collectionName string, ts uint64) error {
	ret := _m.Called(ctx, dbName, alias, collectionName, ts)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, uint64) error); ok {
		r0 = rf(ctx, dbName, alias, collectionName, ts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateDatabase provides a mock function with given fields: ctx, dbName, ts
func (_m *IMetaTable) CreateDatabase(ctx context.Context, dbName string, ts uint64) error {
	ret := _m.Called(ctx, dbName, ts)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, uint64) error); ok {
		r0 = rf(ctx, dbName, ts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateRole provides a mock function with given fields: tenant, entity
func (_m *IMetaTable) CreateRole(tenant string, entity *milvuspb.RoleEntity) error {
	ret := _m.Called(tenant, entity)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *milvuspb.RoleEntity) error); ok {
		r0 = rf(tenant, entity)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteCredential provides a mock function with given fields: username
func (_m *IMetaTable) DeleteCredential(username string) error {
	ret := _m.Called(username)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(username)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DropAlias provides a mock function with given fields: ctx, dbName, alias, ts
func (_m *IMetaTable) DropAlias(ctx context.Context, dbName string, alias string, ts uint64) error {
	ret := _m.Called(ctx, dbName, alias, ts)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, uint64) error); ok {
		r0 = rf(ctx, dbName, alias, ts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DropDatabase provides a mock function with given fields: ctx, dbName, ts
func (_m *IMetaTable) DropDatabase(ctx context.Context, dbName string, ts uint64) error {
	ret := _m.Called(ctx, dbName, ts)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, uint64) error); ok {
		r0 = rf(ctx, dbName, ts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DropGrant provides a mock function with given fields: tenant, role
func (_m *IMetaTable) DropGrant(tenant string, role *milvuspb.RoleEntity) error {
	ret := _m.Called(tenant, role)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *milvuspb.RoleEntity) error); ok {
		r0 = rf(tenant, role)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DropRole provides a mock function with given fields: tenant, roleName
func (_m *IMetaTable) DropRole(tenant string, roleName string) error {
	ret := _m.Called(tenant, roleName)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(tenant, roleName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetCollectionByID provides a mock function with given fields: ctx, dbName, collectionID, ts, allowUnavailable
func (_m *IMetaTable) GetCollectionByID(ctx context.Context, dbName string, collectionID int64, ts uint64, allowUnavailable bool) (*model.Collection, error) {
	ret := _m.Called(ctx, dbName, collectionID, ts, allowUnavailable)

	var r0 *model.Collection
	if rf, ok := ret.Get(0).(func(context.Context, string, int64, uint64, bool) *model.Collection); ok {
		r0 = rf(ctx, dbName, collectionID, ts, allowUnavailable)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Collection)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, int64, uint64, bool) error); ok {
		r1 = rf(ctx, dbName, collectionID, ts, allowUnavailable)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCollectionByName provides a mock function with given fields: ctx, dbName, collectionName, ts
func (_m *IMetaTable) GetCollectionByName(ctx context.Context, dbName string, collectionName string, ts uint64) (*model.Collection, error) {
	ret := _m.Called(ctx, dbName, collectionName, ts)

	var r0 *model.Collection
	if rf, ok := ret.Get(0).(func(context.Context, string, string, uint64) *model.Collection); ok {
		r0 = rf(ctx, dbName, collectionName, ts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Collection)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, uint64) error); ok {
		r1 = rf(ctx, dbName, collectionName, ts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCollectionVirtualChannels provides a mock function with given fields: colID
func (_m *IMetaTable) GetCollectionVirtualChannels(colID int64) []string {
	ret := _m.Called(colID)

	var r0 []string
	if rf, ok := ret.Get(0).(func(int64) []string); ok {
		r0 = rf(colID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// GetCredential provides a mock function with given fields: username
func (_m *IMetaTable) GetCredential(username string) (*internalpb.CredentialInfo, error) {
	ret := _m.Called(username)

	var r0 *internalpb.CredentialInfo
	if rf, ok := ret.Get(0).(func(string) *internalpb.CredentialInfo); ok {
		r0 = rf(username)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*internalpb.CredentialInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(username)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPartitionByName provides a mock function with given fields: collID, partitionName, ts
func (_m *IMetaTable) GetPartitionByName(collID int64, partitionName string, ts uint64) (int64, error) {
	ret := _m.Called(collID, partitionName, ts)

	var r0 int64
	if rf, ok := ret.Get(0).(func(int64, string, uint64) int64); ok {
		r0 = rf(collID, partitionName, ts)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64, string, uint64) error); ok {
		r1 = rf(collID, partitionName, ts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPartitionNameByID provides a mock function with given fields: collID, partitionID, ts
func (_m *IMetaTable) GetPartitionNameByID(collID int64, partitionID int64, ts uint64) (string, error) {
	ret := _m.Called(collID, partitionID, ts)

	var r0 string
	if rf, ok := ret.Get(0).(func(int64, int64, uint64) string); ok {
		r0 = rf(collID, partitionID, ts)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64, int64, uint64) error); ok {
		r1 = rf(collID, partitionID, ts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsAlias provides a mock function with given fields: db, name
func (_m *IMetaTable) IsAlias(db string, name string) bool {
	ret := _m.Called(db, name)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string) bool); ok {
		r0 = rf(db, name)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// ListAbnormalCollections provides a mock function with given fields: ctx, ts
func (_m *IMetaTable) ListAbnormalCollections(ctx context.Context, ts uint64) ([]*model.Collection, error) {
	ret := _m.Called(ctx, ts)

	var r0 []*model.Collection
	if rf, ok := ret.Get(0).(func(context.Context, uint64) []*model.Collection); ok {
		r0 = rf(ctx, ts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Collection)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uint64) error); ok {
		r1 = rf(ctx, ts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAliasesByID provides a mock function with given fields: collID
func (_m *IMetaTable) ListAliasesByID(collID int64) []string {
	ret := _m.Called(collID)

	var r0 []string
	if rf, ok := ret.Get(0).(func(int64) []string); ok {
		r0 = rf(collID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// ListCollectionPhysicalChannels provides a mock function with given fields:
func (_m *IMetaTable) ListCollectionPhysicalChannels() map[int64][]string {
	ret := _m.Called()

	var r0 map[int64][]string
	if rf, ok := ret.Get(0).(func() map[int64][]string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[int64][]string)
		}
	}

	return r0
}

// ListCollections provides a mock function with given fields: ctx, dbName, ts
func (_m *IMetaTable) ListCollections(ctx context.Context, dbName string, ts uint64) ([]*model.Collection, error) {
	ret := _m.Called(ctx, dbName, ts)

	var r0 []*model.Collection
	if rf, ok := ret.Get(0).(func(context.Context, string, uint64) []*model.Collection); ok {
		r0 = rf(ctx, dbName, ts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Collection)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, uint64) error); ok {
		r1 = rf(ctx, dbName, ts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListCredentialUsernames provides a mock function with given fields:
func (_m *IMetaTable) ListCredentialUsernames() (*milvuspb.ListCredUsersResponse, error) {
	ret := _m.Called()

	var r0 *milvuspb.ListCredUsersResponse
	if rf, ok := ret.Get(0).(func() *milvuspb.ListCredUsersResponse); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*milvuspb.ListCredUsersResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListDatabases provides a mock function with given fields: ctx, ts
func (_m *IMetaTable) ListDatabases(ctx context.Context, ts uint64) ([]string, error) {
	ret := _m.Called(ctx, ts)

	var r0 []string
	if rf, ok := ret.Get(0).(func(context.Context, uint64) []string); ok {
		r0 = rf(ctx, ts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uint64) error); ok {
		r1 = rf(ctx, ts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListPolicy provides a mock function with given fields: tenant
func (_m *IMetaTable) ListPolicy(tenant string) ([]string, error) {
	ret := _m.Called(tenant)

	var r0 []string
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(tenant)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(tenant)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListUserRole provides a mock function with given fields: tenant
func (_m *IMetaTable) ListUserRole(tenant string) ([]string, error) {
	ret := _m.Called(tenant)

	var r0 []string
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(tenant)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(tenant)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OperatePrivilege provides a mock function with given fields: tenant, entity, operateType
func (_m *IMetaTable) OperatePrivilege(tenant string, entity *milvuspb.GrantEntity, operateType milvuspb.OperatePrivilegeType) error {
	ret := _m.Called(tenant, entity, operateType)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *milvuspb.GrantEntity, milvuspb.OperatePrivilegeType) error); ok {
		r0 = rf(tenant, entity, operateType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OperateUserRole provides a mock function with given fields: tenant, userEntity, roleEntity, operateType
func (_m *IMetaTable) OperateUserRole(tenant string, userEntity *milvuspb.UserEntity, roleEntity *milvuspb.RoleEntity, operateType milvuspb.OperateUserRoleType) error {
	ret := _m.Called(tenant, userEntity, roleEntity, operateType)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *milvuspb.UserEntity, *milvuspb.RoleEntity, milvuspb.OperateUserRoleType) error); ok {
		r0 = rf(tenant, userEntity, roleEntity, operateType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveCollection provides a mock function with given fields: ctx, collectionID, ts
func (_m *IMetaTable) RemoveCollection(ctx context.Context, collectionID int64, ts uint64) error {
	ret := _m.Called(ctx, collectionID, ts)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, uint64) error); ok {
		r0 = rf(ctx, collectionID, ts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemovePartition provides a mock function with given fields: ctx, collectionID, partitionID, ts
func (_m *IMetaTable) RemovePartition(ctx context.Context, collectionID int64, partitionID int64, ts uint64) error {
	ret := _m.Called(ctx, collectionID, partitionID, ts)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, int64, uint64) error); ok {
		r0 = rf(ctx, collectionID, partitionID, ts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RenameCollection provides a mock function with given fields: ctx, dbName, oldName, newName, ts
func (_m *IMetaTable) RenameCollection(ctx context.Context, dbName string, oldName string, newName string, ts uint64) error {
	ret := _m.Called(ctx, dbName, oldName, newName, ts)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, uint64) error); ok {
		r0 = rf(ctx, dbName, oldName, newName, ts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SelectGrant provides a mock function with given fields: tenant, entity
func (_m *IMetaTable) SelectGrant(tenant string, entity *milvuspb.GrantEntity) ([]*milvuspb.GrantEntity, error) {
	ret := _m.Called(tenant, entity)

	var r0 []*milvuspb.GrantEntity
	if rf, ok := ret.Get(0).(func(string, *milvuspb.GrantEntity) []*milvuspb.GrantEntity); ok {
		r0 = rf(tenant, entity)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*milvuspb.GrantEntity)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, *milvuspb.GrantEntity) error); ok {
		r1 = rf(tenant, entity)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SelectRole provides a mock function with given fields: tenant, entity, includeUserInfo
func (_m *IMetaTable) SelectRole(tenant string, entity *milvuspb.RoleEntity, includeUserInfo bool) ([]*milvuspb.RoleResult, error) {
	ret := _m.Called(tenant, entity, includeUserInfo)

	var r0 []*milvuspb.RoleResult
	if rf, ok := ret.Get(0).(func(string, *milvuspb.RoleEntity, bool) []*milvuspb.RoleResult); ok {
		r0 = rf(tenant, entity, includeUserInfo)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*milvuspb.RoleResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, *milvuspb.RoleEntity, bool) error); ok {
		r1 = rf(tenant, entity, includeUserInfo)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SelectUser provides a mock function with given fields: tenant, entity, includeRoleInfo
func (_m *IMetaTable) SelectUser(tenant string, entity *milvuspb.UserEntity, includeRoleInfo bool) ([]*milvuspb.UserResult, error) {
	ret := _m.Called(tenant, entity, includeRoleInfo)

	var r0 []*milvuspb.UserResult
	if rf, ok := ret.Get(0).(func(string, *milvuspb.UserEntity, bool) []*milvuspb.UserResult); ok {
		r0 = rf(tenant, entity, includeRoleInfo)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*milvuspb.UserResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, *milvuspb.UserEntity, bool) error); ok {
		r1 = rf(tenant, entity, includeRoleInfo)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewIMetaTable interface {
	mock.TestingT
	Cleanup(func())
}

// NewIMetaTable creates a new instance of IMetaTable. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewIMetaTable(t mockConstructorTestingTNewIMetaTable) *IMetaTable {
	mock := &IMetaTable{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
