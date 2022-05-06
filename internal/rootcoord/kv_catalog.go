package rootcoord

import (
	"context"
	"fmt"
	"path"
	"strconv"

	"github.com/milvus-io/milvus/internal/metastore/model"

	"github.com/golang/protobuf/proto"
	"github.com/milvus-io/milvus/internal/kv"
	"github.com/milvus-io/milvus/internal/log"
	pb "github.com/milvus-io/milvus/internal/proto/etcdpb"
	"github.com/milvus-io/milvus/internal/proto/schemapb"
	"github.com/milvus-io/milvus/internal/util/typeutil"
	"go.uber.org/zap"
)

type KVCatalog struct {
	txn      kv.TxnKV
	snapshot kv.SnapShotKV
}

func (kc *KVCatalog) CreateCollection(ctx context.Context, coll *model.Collection, ts typeutil.Timestamp) error {
	k1 := fmt.Sprintf("%s/%d", CollectionMetaPrefix, coll.CollectionID)
	collInfo := model.ConvertToCollectionPB(coll)
	v1, err := proto.Marshal(collInfo)
	if err != nil {
		log.Error("marshal fail", zap.String("key", k1), zap.Error(err))
		return fmt.Errorf("marshal fail key:%s, err:%w", k1, err)
	}

	// save ddOpStr into etcd
	kvs := map[string]string{k1: string(v1)}
	for k, v := range coll.Extra {
		kvs[k] = v
	}

	err = kc.snapshot.MultiSave(kvs, ts)
	if err != nil {
		log.Error("SnapShotKV MultiSave fail", zap.Error(err))
		panic("SnapShotKV MultiSave fail")
	}

	return nil
}

func (kc *KVCatalog) CreatePartition(ctx context.Context, coll *model.Collection, partition *model.Partition, ts typeutil.Timestamp) error {
	kc.CreateCollection(ctx, coll, ts)

	err := kc.txn.MultiSave(partition.Extra)
	if err != nil {
		// will not panic, missing create msg
		log.Warn("TxnKV MultiSave fail", zap.Error(err))
	}

	return nil
}

func (kc *KVCatalog) CreateIndex(ctx context.Context, segIndex *model.SegmentIndex) error {
	k := fmt.Sprintf("%s/%d/%d/%d/%d", SegmentIndexMetaPrefix, segIndex.CollectionID, segIndex.IndexID, segIndex.PartitionID, segIndex.SegmentID)
	segIdxInfo := model.ConvertToSegmentIndexPB(segIndex)
	v, err := proto.Marshal(segIdxInfo)
	if err != nil {
		log.Error("marshal segIdxInfo fail", zap.String("key", k), zap.Error(err))
		return fmt.Errorf("marshal segIdxInfo fail key:%s, err:%w", k, err)
	}

	err = kc.txn.Save(k, string(v))
	if err != nil {
		log.Error("TxnKV Save fail", zap.Error(err))
		panic("TxnKV Save fail")
	}

	return nil
}

func (kc *KVCatalog) CreateAlias(ctx context.Context, collAlias *model.CollectionAlias, ts typeutil.Timestamp) error {
	k := fmt.Sprintf("%s/%s", CollectionAliasMetaPrefix, collAlias.Alias)
	v, err := proto.Marshal(&pb.CollectionInfo{ID: collAlias.CollectionID, Schema: &schemapb.CollectionSchema{Name: collAlias.Alias}})
	if err != nil {
		log.Error("marshal CollectionInfo fail", zap.String("key", k), zap.Error(err))
		return fmt.Errorf("marshal CollectionInfo fail key:%s, err:%w", k, err)
	}

	err = kc.snapshot.Save(k, string(v), ts)
	if err != nil {
		log.Error("SnapShotKV Save fail", zap.Error(err))
		panic("SnapShotKV Save fail")
	}

	return nil
}

func (kc *KVCatalog) DropCollection(ctx context.Context, collectionInfo *model.Collection, ts typeutil.Timestamp) error {
	delMetakeysSnap := []string{
		fmt.Sprintf("%s/%d", CollectionMetaPrefix, collectionInfo.CollectionID),
	}
	for _, alias := range collectionInfo.Aliases {
		delMetakeysSnap = append(delMetakeysSnap,
			fmt.Sprintf("%s/%s", CollectionAliasMetaPrefix, alias),
		)
	}

	err := kc.snapshot.MultiSaveAndRemoveWithPrefix(map[string]string{}, delMetakeysSnap, ts)
	if err != nil {
		log.Error("SnapshotKV save and remove failed", zap.Error(err))
		panic("save etcd failed")
	}

	// txn operation
	kvs := map[string]string{}
	for k, v := range collectionInfo.Extra {
		kvs[k] = v
	}

	delMetaKeysTxn := []string{
		fmt.Sprintf("%s/%d", SegmentIndexMetaPrefix, collectionInfo.CollectionID),
		fmt.Sprintf("%s/%d", IndexMetaPrefix, collectionInfo.CollectionID),
	}

	err = kc.txn.MultiSaveAndRemoveWithPrefix(kvs, delMetaKeysTxn)
	if err != nil {
		log.Warn("TxnKV save and remove failed", zap.Error(err))
	}

	return nil
}

func (kc *KVCatalog) DropPartition(ctx context.Context, collectionInfo *model.Collection, partitionID typeutil.UniqueID, ts typeutil.Timestamp) error {
	collMeta := model.ConvertToCollectionPB(collectionInfo)

	k := path.Join(CollectionMetaPrefix, strconv.FormatInt(collectionInfo.CollectionID, 10))
	v, err := proto.Marshal(collMeta)
	if err != nil {
		log.Error("MetaTable DeletePartition Marshal collectionMeta fail",
			zap.String("key", k), zap.Error(err))
		return fmt.Errorf("metaTable DeletePartition Marshal collectionMeta fail key:%s, err:%w", k, err)
	}

	err = kc.snapshot.Save(k, string(v), ts)
	if err != nil {
		log.Error("SnapShotKV MultiSaveAndRemoveWithPrefix fail", zap.Error(err))
		panic("SnapShotKV MultiSaveAndRemoveWithPrefix fail")
	}

	var delMetaKeys []string
	for _, idxInfo := range collMeta.FieldIndexes {
		k := fmt.Sprintf("%s/%d/%d/%d", SegmentIndexMetaPrefix, collMeta.ID, idxInfo.IndexID, partitionID)
		delMetaKeys = append(delMetaKeys, k)
	}

	// txn operation
	metaTxn := map[string]string{}
	for k, v := range collectionInfo.Extra {
		metaTxn[k] = v
	}
	err = kc.txn.MultiSaveAndRemoveWithPrefix(metaTxn, delMetaKeys)
	if err != nil {
		log.Warn("TxnKV MultiSaveAndRemoveWithPrefix fail", zap.Error(err))
		// will not panic, failed txn shall be treated by garbage related logic
	}

	return nil
}

func (kc *KVCatalog) DropIndex(ctx context.Context, collectionInfo *model.Collection, dropIdxID typeutil.UniqueID, ts typeutil.Timestamp) error {
	collMeta := model.ConvertToCollectionPB(collectionInfo)

	k := path.Join(CollectionMetaPrefix, strconv.FormatInt(collectionInfo.CollectionID, 10))
	v, err := proto.Marshal(collMeta)
	if err != nil {
		log.Error("MetaTable DropIndex Marshal collMeta fail",
			zap.String("key", k), zap.Error(err))
		return fmt.Errorf("metaTable DropIndex Marshal collMeta fail key:%s, err:%w", k, err)
	}
	saveMeta := map[string]string{k: string(v)}

	delMeta := []string{
		fmt.Sprintf("%s/%d/%d", SegmentIndexMetaPrefix, collectionInfo.CollectionID, dropIdxID),
		fmt.Sprintf("%s/%d/%d", IndexMetaPrefix, collectionInfo.CollectionID, dropIdxID),
	}

	err = kc.txn.MultiSaveAndRemoveWithPrefix(saveMeta, delMeta)
	if err != nil {
		log.Error("TxnKV MultiSaveAndRemoveWithPrefix fail", zap.Error(err))
		panic("TxnKV MultiSaveAndRemoveWithPrefix fail")
	}

	return nil
}

func (kc *KVCatalog) DropCredential(ctx context.Context, username string) error {
	k := fmt.Sprintf("%s/%s", CredentialPrefix, username)

	err := kc.txn.Remove(k)
	if err != nil {
		log.Error("MetaTable remove fail", zap.Error(err))
		return fmt.Errorf("remove credential fail key:%s, err:%w", username, err)
	}
	return nil
}

func (kc *KVCatalog) DropAlias(ctx context.Context, collectionID typeutil.UniqueID, alias string, ts typeutil.Timestamp) error {
	delMetakeys := []string{
		fmt.Sprintf("%s/%s", CollectionAliasMetaPrefix, alias),
	}

	meta := make(map[string]string)
	err := kc.snapshot.MultiSaveAndRemoveWithPrefix(meta, delMetakeys, ts)
	if err != nil {
		log.Error("SnapShotKV MultiSaveAndRemoveWithPrefix fail", zap.Error(err))
		panic("SnapShotKV MultiSaveAndRemoveWithPrefix fail")
	}

	return nil
}
