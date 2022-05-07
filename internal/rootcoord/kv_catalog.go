package rootcoord

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/milvus-io/milvus/internal/kv"
	"github.com/milvus-io/milvus/internal/log"
	"github.com/milvus-io/milvus/internal/metastore/model"
	"github.com/milvus-io/milvus/internal/proto/etcdpb"
	pb "github.com/milvus-io/milvus/internal/proto/etcdpb"
	"github.com/milvus-io/milvus/internal/proto/internalpb"
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

func (kc *KVCatalog) CreateCredential(ctx context.Context, credential *model.Credential) error {
	k := fmt.Sprintf("%s/%s", CredentialPrefix, credential.Username)
	v, err := json.Marshal(&internalpb.CredentialInfo{EncryptedPassword: credential.EncryptedPassword})
	if err != nil {
		log.Error("marshal credential info fail", zap.String("key", k), zap.Error(err))
		return fmt.Errorf("marshal credential info fail key:%s, err:%w", k, err)
	}
	err = kc.txn.Save(k, string(v))
	if err != nil {
		log.Error("TxnKV save fail", zap.Error(err))
		return fmt.Errorf("TxnKV save fail key:%s, err:%w", credential.Username, err)
	}

	return nil
}

func (kc *KVCatalog) GetCollection(ctx context.Context, collectionID typeutil.UniqueID, ts typeutil.Timestamp) (*pb.CollectionInfo, error) {
	collKey := fmt.Sprintf("%s/%d", CollectionMetaPrefix, collectionID)
	collVal, err := kc.snapshot.Load(collKey, ts)
	if err != nil {
		log.Error("SnapShotKV Load fail", zap.Error(err))
		return nil, err
	}
	collMeta := &pb.CollectionInfo{}
	err = proto.Unmarshal([]byte(collVal), collMeta)
	if err != nil {
		log.Error("unmarshal collection info fail", zap.String("key", collKey), zap.Error(err))
		return nil, err
	}
	return collMeta, nil
}

func (kc *KVCatalog) CollectionExists(ctx context.Context, collectionID typeutil.UniqueID, ts typeutil.Timestamp) bool {
	_, err := kc.GetCollection(ctx, collectionID, ts)
	return err == nil
}

func (kc *KVCatalog) GetCredential(ctx context.Context, username string) (*model.Credential, error) {
	k := fmt.Sprintf("%s/%s", CredentialPrefix, username)
	v, err := kc.txn.Load(k)
	if err != nil {
		log.Warn("TxnKV load fail", zap.String("key", k), zap.Error(err))
		return nil, err
	}

	credentialInfo := internalpb.CredentialInfo{}
	err = json.Unmarshal([]byte(v), &credentialInfo)
	if err != nil {
		return nil, fmt.Errorf("unmarshal credential info err:%w", err)
	}
	return &model.Credential{Username: username, EncryptedPassword: credentialInfo.EncryptedPassword}, nil
}

func (kc *KVCatalog) AlterAlias(ctx context.Context, collAlias *model.CollectionAlias, ts typeutil.Timestamp) error {
	return kc.CreateAlias(ctx, collAlias, ts)
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

func (kc *KVCatalog) ListCollections(ctx context.Context, ts typeutil.Timestamp) (map[string]*etcdpb.CollectionInfo, error) {
	_, vals, err := kc.snapshot.LoadWithPrefix(CollectionMetaPrefix, ts)
	if err != nil {
		log.Error("load with prefix error", zap.Uint64("timestamp", ts), zap.Error(err))
		return nil, err
	}
	colls := make(map[string]*pb.CollectionInfo)
	for _, val := range vals {
		collMeta := pb.CollectionInfo{}
		err := proto.Unmarshal([]byte(val), &collMeta)
		if err != nil {
			log.Warn("unmarshal collection info failed", zap.Error(err))
			continue
		}
		colls[collMeta.Schema.Name] = &collMeta
	}
	return colls, nil
}

func (kc *KVCatalog) ListCredentials(ctx context.Context) ([]string, error) {
	keys, _, err := kc.txn.LoadWithPrefix(CredentialPrefix)
	if err != nil {
		log.Error("MetaTable list all credential usernames fail", zap.Error(err))
		return nil, err
	}

	var usernames []string
	for _, path := range keys {
		username := typeutil.After(path, UserSubPrefix+"/")
		if len(username) == 0 {
			log.Warn("no username extract from path:", zap.String("path", path))
			continue
		}
		usernames = append(usernames, username)
	}
	return usernames, nil
}
