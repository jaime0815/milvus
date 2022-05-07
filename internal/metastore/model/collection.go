package model

import (
	"github.com/milvus-io/milvus/internal/proto/commonpb"
	"github.com/milvus-io/milvus/internal/proto/etcdpb"
	"github.com/milvus-io/milvus/internal/proto/schemapb"
)

type Collection struct {
	TenantID             string
	CollectionID         int64
	Partitions           []Partition
	Name                 string
	Description          string
	AutoID               bool
	Fields               []*schemapb.FieldSchema
	FieldIndexes         []*etcdpb.FieldIndexInfo
	VirtualChannelNames  []string
	PhysicalChannelNames []string
	ShardsNum            int32
	StartPositions       []*commonpb.KeyDataPair
	CreateTime           uint64
	ConsistencyLevel     commonpb.ConsistencyLevel
	Aliases              []string
	Extra                map[string]string // extra kvs
}

type Partition struct {
	PartitionID               int64
	PartitionName             string
	PartitionCreatedTimestamp uint64
	Extra                     map[string]string
}
