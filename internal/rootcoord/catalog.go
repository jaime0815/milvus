package rootcoord

import (
	"context"

	"github.com/milvus-io/milvus/internal/metastore/model"
	"github.com/milvus-io/milvus/internal/proto/etcdpb"
	"github.com/milvus-io/milvus/internal/util/typeutil"
)

type Catalog interface {
	CreateCollection(ctx context.Context, collectionInfo *model.Collection, ts typeutil.Timestamp) error
	CreatePartition(ctx context.Context, coll *model.Collection, partitionInfo *model.Partition, ts typeutil.Timestamp) error
	CreateIndex(ctx context.Context, index *model.SegmentIndex) error
	CreateAlias(ctx context.Context, collAlias *model.CollectionAlias, ts typeutil.Timestamp) error
	CreateCredential(ctx context.Context, credential *model.Credential) error

	GetCollection(ctx context.Context, collectionID typeutil.UniqueID, ts typeutil.Timestamp) (*etcdpb.CollectionInfo, error)
	CollectionExists(ctx context.Context, collectionID typeutil.UniqueID, ts typeutil.Timestamp) bool

	AlterAlias(ctx context.Context, collAlias *model.CollectionAlias, ts typeutil.Timestamp) error

	DropCollection(ctx context.Context, collectionInfo *model.Collection, ts typeutil.Timestamp) error
	DropPartition(ctx context.Context, collectionInfo *model.Collection, partitionID typeutil.UniqueID, ts typeutil.Timestamp) error
	DropIndex(ctx context.Context, collectionInfo *model.Collection, dropIdxID typeutil.UniqueID, ts typeutil.Timestamp) error
	DropCredential(ctx context.Context, username string) error
	DropAlias(ctx context.Context, collectionID typeutil.UniqueID, alias string, ts typeutil.Timestamp) error
}
