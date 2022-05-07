package model

import (
	pb "github.com/milvus-io/milvus/internal/proto/etcdpb"
	"github.com/milvus-io/milvus/internal/proto/internalpb"
	"github.com/milvus-io/milvus/internal/proto/schemapb"
)

func ConvertCollectionPBToModel(coll *pb.CollectionInfo, extra map[string]string) *Collection {
	partitions := make([]Partition, len(coll.PartitionIDs))
	for idx := range coll.PartitionIDs {
		partitions[idx] = Partition{
			PartitionID:               coll.PartitionIDs[idx],
			PartitionName:             coll.PartitionNames[idx],
			PartitionCreatedTimestamp: coll.PartitionCreatedTimestamps[idx],
		}
	}
	return &Collection{
		CollectionID:         coll.ID,
		Name:                 coll.Schema.Name,
		Description:          coll.Schema.Description,
		AutoID:               coll.Schema.AutoID,
		Fields:               coll.Schema.Fields,
		Partitions:           partitions,
		FieldIndexes:         coll.FieldIndexes,
		VirtualChannelNames:  coll.VirtualChannelNames,
		PhysicalChannelNames: coll.PhysicalChannelNames,
		ShardsNum:            coll.ShardsNum,
		ConsistencyLevel:     coll.ConsistencyLevel,
		CreateTime:           coll.CreateTime,
		Extra:                extra,
	}
}

func CloneCollectionModel(coll Collection) *Collection {
	return &Collection{
		CollectionID:         coll.CollectionID,
		Name:                 coll.Name,
		Description:          coll.Description,
		AutoID:               coll.AutoID,
		Fields:               coll.Fields,
		Partitions:           coll.Partitions,
		FieldIndexes:         coll.FieldIndexes,
		VirtualChannelNames:  coll.VirtualChannelNames,
		PhysicalChannelNames: coll.PhysicalChannelNames,
		ShardsNum:            coll.ShardsNum,
		ConsistencyLevel:     coll.ConsistencyLevel,
		CreateTime:           coll.CreateTime,
		Extra:                coll.Extra,
	}
}

func ConvertToCollectionPB(coll *Collection) *pb.CollectionInfo {
	collSchema := &schemapb.CollectionSchema{
		Name:        coll.Name,
		Description: coll.Description,
		AutoID:      coll.AutoID,
		Fields:      coll.Fields,
	}
	partitionIDs := make([]int64, len(coll.Partitions))
	partitionNames := make([]string, len(coll.Partitions))
	partitionCreatedTimestamps := make([]uint64, len(coll.Partitions))

	for idx, partition := range coll.Partitions {
		partitionIDs[idx] = partition.PartitionID
		partitionNames[idx] = partition.PartitionName
		partitionCreatedTimestamps[idx] = partition.PartitionCreatedTimestamp
	}
	return &pb.CollectionInfo{
		ID:                         coll.CollectionID,
		Schema:                     collSchema,
		PartitionIDs:               partitionIDs,
		PartitionNames:             partitionNames,
		FieldIndexes:               coll.FieldIndexes,
		VirtualChannelNames:        coll.VirtualChannelNames,
		PhysicalChannelNames:       coll.PhysicalChannelNames,
		ShardsNum:                  coll.ShardsNum,
		PartitionCreatedTimestamps: partitionCreatedTimestamps,
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

func ConvertToCredentialPB(cred *Credential) *internalpb.CredentialInfo {
	if cred == nil {
		return nil
	}
	return &internalpb.CredentialInfo{
		Username:          cred.Username,
		EncryptedPassword: cred.EncryptedPassword,
	}
}
