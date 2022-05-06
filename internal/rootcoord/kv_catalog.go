package rootcoord

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/milvus-io/milvus/internal/kv"
	"github.com/milvus-io/milvus/internal/log"
	"github.com/milvus-io/milvus/internal/model"
	pb "github.com/milvus-io/milvus/internal/proto/etcdpb"
	"github.com/milvus-io/milvus/internal/util/typeutil"
	"go.uber.org/zap"
)

type KVCatalog struct {
	kv       kv.BaseKV
	snapshot kv.SnapShotKV
}

func toPB(coll *model.Collection) *pb.CollectionInfo {
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

func (kc *KVCatalog) CreateCollection(coll *model.Collection, ts typeutil.Timestamp) error {
	k1 := fmt.Sprintf("%s/%d", CollectionMetaPrefix, coll.CollectionID)
	collInfo := toPB(coll)
	v1, err := proto.Marshal(collInfo)
	if err != nil {
		log.Error("MetaTable AddCollection saveColl Marshal fail",
			zap.String("key", k1), zap.Error(err))
		return fmt.Errorf("metaTable AddCollection Marshal fail key:%s, err:%w", k1, err)
	}

	// save ddOpStr into etcd
	kvs := map[string]string{}
	if len(coll.Extra) > 0 {
		for k, v := range coll.Extra {
			kvs[k] = v
		}
		kvs[k1] = string(v1)
	} else {
		kvs = map[string]string{k1: string(v1)}
	}

	err = kc.snapshot.MultiSave(kvs, ts)
	if err != nil {
		log.Error("SnapShotKV MultiSave fail", zap.Error(err))
		panic("SnapShotKV MultiSave fail")
	}

	return nil
}
