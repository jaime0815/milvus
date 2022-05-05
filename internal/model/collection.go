package model

import (
	"github.com/milvus-io/milvus/internal/proto/commonpb"
	"github.com/milvus-io/milvus/internal/proto/etcdpb"
	"github.com/milvus-io/milvus/internal/proto/schemapb"
)

type Collection struct {
	TenantID                   string
	CollectionID               int64
	PartitionName              string
	PartitionID                int64
	Schema                     *schemapb.CollectionSchema
	FieldIndexes               []*etcdpb.FieldIndexInfo
	VirtualChannelNames        []string
	PhysicalChannelNames       []string
	PartitionCreatedTimestamps []uint64
	ShardsNum                  int32
	StartPositions             []*commonpb.KeyDataPair
	ConsistencyLevel           commonpb.ConsistencyLevel
}
