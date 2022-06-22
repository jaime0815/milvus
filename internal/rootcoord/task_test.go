package rootcoord

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/milvus-io/milvus/internal/util/typeutil"

	"github.com/milvus-io/milvus/internal/metastore/model"
	"github.com/milvus-io/milvus/internal/proto/commonpb"
	"github.com/milvus-io/milvus/internal/proto/milvuspb"
	"github.com/milvus-io/milvus/internal/proto/rootcoordpb"
	"github.com/milvus-io/milvus/internal/proto/schemapb"
)

func TestDescribeSegmentReqTask_Type(t *testing.T) {
	tsk := &DescribeSegmentsReqTask{
		Req: &rootcoordpb.DescribeSegmentsRequest{
			Base: &commonpb.MsgBase{
				MsgType: commonpb.MsgType_DescribeSegments,
			},
		},
	}
	assert.Equal(t, commonpb.MsgType_DescribeSegments, tsk.Type())
}

func TestDescribeSegmentsReqTask_Execute(t *testing.T) {
	collID := typeutil.UniqueID(1)
	partID := typeutil.UniqueID(2)
	segID := typeutil.UniqueID(100)
	fieldID := typeutil.UniqueID(3)
	buildID := typeutil.UniqueID(4)
	indexID := typeutil.UniqueID(1000)
	indexName := "test_describe_segments_index"

	c := &Core{}

	// failed to get flushed segments.
	c.CallGetFlushedSegmentsService = func(ctx context.Context, collID, partID typeutil.UniqueID) ([]typeutil.UniqueID, error) {
		return nil, errors.New("mock")
	}
	tsk := &DescribeSegmentsReqTask{
		baseReqTask: baseReqTask{
			core: c,
		},
		Req: &rootcoordpb.DescribeSegmentsRequest{
			Base: &commonpb.MsgBase{
				MsgType: commonpb.MsgType_DescribeSegments,
			},
			CollectionID: collID,
			SegmentIDs:   []typeutil.UniqueID{segID},
		},
		Rsp: &rootcoordpb.DescribeSegmentsResponse{},
	}
	assert.Error(t, tsk.Execute(context.Background()))

	// requested segment not found in flushed segments.
	c.CallGetFlushedSegmentsService = func(ctx context.Context, collID, partID typeutil.UniqueID) ([]typeutil.UniqueID, error) {
		return []typeutil.UniqueID{}, nil
	}
	assert.Error(t, tsk.Execute(context.Background()))

	// segment not found in meta.
	c.CallGetFlushedSegmentsService = func(ctx context.Context, collID, partID typeutil.UniqueID) ([]typeutil.UniqueID, error) {
		return []typeutil.UniqueID{segID}, nil
	}
	c.MetaTable = &MetaTable{
		segID2IndexID: make(map[typeutil.UniqueID]typeutil.UniqueID, 1),
	}
	assert.NoError(t, tsk.Execute(context.Background()))

	// index not found in meta. no return error
	c.MetaTable = &MetaTable{
		segID2IndexID: map[typeutil.UniqueID]typeutil.UniqueID{segID: indexID},
		indexID2Meta: map[typeutil.UniqueID]*model.Index{
			indexID: {
				CollectionID: collID,
				FieldID:      fieldID,
				IndexID:      indexID,
				SegmentIndexes: map[int64]model.SegmentIndex{
					segID + 1: {
						Segment: model.Segment{
							SegmentID:   segID,
							PartitionID: partID,
						},
						BuildID:     buildID,
						EnableIndex: true,
					},
				},
			},
		},
	}
	assert.NoError(t, tsk.Execute(context.Background()))

	// success.
	c.MetaTable = &MetaTable{
		segID2IndexID: map[typeutil.UniqueID]typeutil.UniqueID{segID: indexID},
		indexID2Meta: map[typeutil.UniqueID]*model.Index{
			indexID: {
				CollectionID: collID,
				FieldID:      fieldID,
				IndexID:      indexID,
				IndexName:    indexName,
				IndexParams:  nil,
				SegmentIndexes: map[int64]model.SegmentIndex{
					segID: {
						Segment: model.Segment{
							SegmentID:   segID,
							PartitionID: partID,
						},
						BuildID:     buildID,
						EnableIndex: true,
					},
				},
			},
		},
	}
	assert.NoError(t, tsk.Execute(context.Background()))
}

func Test_hasSystemFields(t *testing.T) {
	t.Run("no system fields", func(t *testing.T) {
		schema := &schemapb.CollectionSchema{Fields: []*schemapb.FieldSchema{{Name: "not_system_field"}}}
		assert.False(t, hasSystemFields(schema, []string{RowIDFieldName, TimeStampFieldName}))
	})

	t.Run("has row id field", func(t *testing.T) {
		schema := &schemapb.CollectionSchema{Fields: []*schemapb.FieldSchema{{Name: RowIDFieldName}}}
		assert.True(t, hasSystemFields(schema, []string{RowIDFieldName, TimeStampFieldName}))
	})

	t.Run("has timestamp field", func(t *testing.T) {
		schema := &schemapb.CollectionSchema{Fields: []*schemapb.FieldSchema{{Name: TimeStampFieldName}}}
		assert.True(t, hasSystemFields(schema, []string{RowIDFieldName, TimeStampFieldName}))
	})
}

func TestCreateCollectionReqTask_Execute_hasSystemFields(t *testing.T) {
	schema := &schemapb.CollectionSchema{Name: "test", Fields: []*schemapb.FieldSchema{{Name: TimeStampFieldName}}}
	marshaledSchema, err := proto.Marshal(schema)
	assert.NoError(t, err)
	task := &CreateCollectionReqTask{
		Req: &milvuspb.CreateCollectionRequest{
			Base:           &commonpb.MsgBase{MsgType: commonpb.MsgType_CreateCollection},
			CollectionName: "test",
			Schema:         marshaledSchema,
		},
	}
	err = task.Execute(context.Background())
	assert.Error(t, err)
}
