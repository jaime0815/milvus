package model

import pb "github.com/milvus-io/milvus/internal/proto/etcdpb"

func ConvertCollectionPBToModel(coll *pb.CollectionInfo, extra map[string]string) *Collection {
	return &Collection{
		CollectionID:               coll.ID,
		Schema:                     coll.Schema,
		PartitionIDs:               coll.PartitionIDs,
		PartitionNames:             coll.PartitionNames,
		FieldIndexes:               coll.FieldIndexes,
		VirtualChannelNames:        coll.VirtualChannelNames,
		PhysicalChannelNames:       coll.PhysicalChannelNames,
		ShardsNum:                  coll.ShardsNum,
		PartitionCreatedTimestamps: coll.PartitionCreatedTimestamps,
		ConsistencyLevel:           coll.ConsistencyLevel,
		Extra:                      extra,
	}
}

func ConvertToCollectionPB(coll *Collection) *pb.CollectionInfo {
	return &pb.CollectionInfo{
		ID:                         coll.CollectionID,
		Schema:                     coll.Schema,
		PartitionIDs:               coll.PartitionIDs,
		PartitionNames:             coll.PartitionNames,
		FieldIndexes:               coll.FieldIndexes,
		VirtualChannelNames:        coll.VirtualChannelNames,
		PhysicalChannelNames:       coll.PhysicalChannelNames,
		ShardsNum:                  coll.ShardsNum,
		PartitionCreatedTimestamps: coll.PartitionCreatedTimestamps,
		ConsistencyLevel:           coll.ConsistencyLevel,
		StartPositions:             coll.StartPositions,
	}
}

func ConvertToSegmentIndexPB(segIndex *SegmentIndex) *pb.SegmentIndexInfo {
	return &pb.SegmentIndexInfo{
		CollectionID: segIndex.CollectionID,
		PartitionID:  segIndex.PartitionID,
		SegmentID:    segIndex.SegmentID,
		FieldID:      segIndex.FieldID,
		IndexID:      segIndex.IndexID,
		BuildID:      segIndex.BuildID,
		EnableIndex:  segIndex.EnableIndex,
	}
}
