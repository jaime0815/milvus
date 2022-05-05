package rootcoord

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/milvus-io/milvus/internal/kv"
	"github.com/milvus-io/milvus/internal/log"
	"github.com/milvus-io/milvus/internal/proto/modelpb"
	"github.com/milvus-io/milvus/internal/util/typeutil"
	"go.uber.org/zap"
)

type KVCatalog struct {
	kv       kv.BaseKV
	snapshot kv.SnapShotKV
}

func (kc *KVCatalog) CreateCollection(ctx context.Context, coll *modelpb.Collection, ts typeutil.Timestamp) error {
	k1 := fmt.Sprintf("%s/%d", CollectionMetaPrefix, coll.ID)
	v1, err := proto.Marshal(coll)
	if err != nil {
		log.Error("MetaTable AddCollection saveColl Marshal fail",
			zap.String("key", k1), zap.Error(err))
		return fmt.Errorf("metaTable AddCollection Marshal fail key:%s, err:%w", k1, err)
	}
	meta := map[string]string{k1: string(v1)}

	err = kc.snapshot.MultiSave(meta, ts)
	if err != nil {
		log.Error("SnapShotKV MultiSave fail", zap.Error(err))
		panic("SnapShotKV MultiSave fail")
	}

	return nil
}
