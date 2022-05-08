package model

import "github.com/milvus-io/milvus/internal/proto/commonpb"

type Index struct {
	CollectionID   int64
	PartitionID    int64
	SegmentID      int64
	FieldID        int64
	IndexID        int64
	IndexName      string
	IndexParams    []*commonpb.KeyValuePair
	BuildID        int64
	IndexSize      uint64
	IndexFilePaths []string
	EnableIndex    bool
	Extra          map[string]string
}
