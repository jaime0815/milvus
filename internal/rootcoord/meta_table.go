// Licensed to the LF AI & Data foundation under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rootcoord

import (
	"context"
	"fmt"
	"sync"

	"github.com/milvus-io/milvus/internal/common"

	"github.com/milvus-io/milvus/internal/kv"
	"github.com/milvus-io/milvus/internal/log"
	"github.com/milvus-io/milvus/internal/metastore"
	kvmetestore "github.com/milvus-io/milvus/internal/metastore/kv"
	"github.com/milvus-io/milvus/internal/metastore/model"
	"github.com/milvus-io/milvus/internal/proto/commonpb"
	pb "github.com/milvus-io/milvus/internal/proto/etcdpb"
	"github.com/milvus-io/milvus/internal/proto/internalpb"
	"github.com/milvus-io/milvus/internal/proto/milvuspb"
	"github.com/milvus-io/milvus/internal/proto/schemapb"
	"github.com/milvus-io/milvus/internal/util/typeutil"
	"go.uber.org/zap"
)

const (
	// TimestampPrefix prefix for timestamp
	TimestampPrefix = kvmetestore.ComponentPrefix + "/timestamp"

	// DDOperationPrefix prefix for DD operation
	DDOperationPrefix = kvmetestore.ComponentPrefix + "/dd-operation"

	// DDMsgSendPrefix prefix to indicate whether DD msg has been send
	DDMsgSendPrefix = kvmetestore.ComponentPrefix + "/dd-msg-send"

	// CreateCollectionDDType name of DD type for create collection
	CreateCollectionDDType = "CreateCollection"

	// DropCollectionDDType name of DD type for drop collection
	DropCollectionDDType = "DropCollection"

	// CreatePartitionDDType name of DD type for create partition
	CreatePartitionDDType = "CreatePartition"

	// DropPartitionDDType name of DD type for drop partition
	DropPartitionDDType = "DropPartition"

	// DefaultIndexType name of default index type for scalar field
	DefaultIndexType = "STL_SORT"

	// DefaultStringIndexType name of default index type for varChar/string field
	DefaultStringIndexType = "Trie"
)

// MetaTable store all rootCoord meta info
type MetaTable struct {
	ctx      context.Context
	txn      kv.TxnKV      // client of a reliable txnkv service, i.e. etcd client
	snapshot kv.SnapShotKV // client of a reliable snapshotkv service, i.e. etcd client
	catalog  metastore.Catalog

	collID2Meta   map[typeutil.UniqueID]model.Collection           // collection id -> collection meta
	collName2ID   map[string]typeutil.UniqueID                     // collection name to collection id
	collAlias2ID  map[string]typeutil.UniqueID                     // collection alias to collection id
	partID2SegID  map[typeutil.UniqueID]map[typeutil.UniqueID]bool // partition id -> segment_id -> bool
	segID2IndexID map[typeutil.UniqueID]typeutil.UniqueID          // segment_id -> index_id
	indexID2Meta  map[typeutil.UniqueID]*model.Index               // collection id/index_id -> meta

	ddLock   sync.RWMutex
	credLock sync.RWMutex
}

// NewMetaTable creates meta table for rootcoord, which stores all in-memory information
// for collection, partition, segment, index etc.
func NewMetaTable(ctx context.Context, txn kv.TxnKV, snap kv.SnapShotKV) (*MetaTable, error) {
	mt := &MetaTable{
		ctx:       ctx,
		txn:       txn,
		snapshot:  snap,
		catalog:   &kvmetestore.Catalog{Txn: txn, Snapshot: snap},
		ddLock:    sync.RWMutex{},
		credLock:  sync.RWMutex{},
	}
	err := mt.reloadFromKV()
	if err != nil {
		return nil, err
	}
	return mt, nil
}

func (mt *MetaTable) reloadFromKV() error {
	mt.collID2Meta = make(map[typeutil.UniqueID]model.Collection)
	mt.collName2ID = make(map[string]typeutil.UniqueID)
	mt.collAlias2ID = make(map[string]typeutil.UniqueID)
	mt.partID2SegID = make(map[typeutil.UniqueID]map[typeutil.UniqueID]bool)
	mt.segID2IndexID = make(map[typeutil.UniqueID]typeutil.UniqueID)
	mt.indexID2Meta = make(map[typeutil.UniqueID]*model.Index)

	collMap, err := mt.catalog.ListCollections(mt.ctx, 0)
	if err != nil {
		return err
	}
	for _, coll := range collMap {
		mt.collID2Meta[coll.CollectionID] = *coll
		mt.collName2ID[coll.Name] = coll.CollectionID
	}

	indexes, err := mt.catalog.ListIndexes(mt.ctx)
	if err != nil {
		return err
	}
	for _, index := range indexes {
		for _, segIndexInfo := range index.SegmentIndexes {
			// update partID2SegID
			segIDMap, ok := mt.partID2SegID[segIndexInfo.Segment.PartitionID]
			if ok {
				segIDMap[segIndexInfo.Segment.SegmentID] = true
			} else {
				idMap := make(map[typeutil.UniqueID]bool)
				idMap[segIndexInfo.Segment.SegmentID] = true
				mt.partID2SegID[segIndexInfo.Segment.PartitionID] = idMap
			}

			mt.segID2IndexID[segIndexInfo.Segment.SegmentID] = index.IndexID
		}

		mt.indexID2Meta[index.IndexID] = index
	}

	log.Debug("reload meta table from KV successfully")
	return nil
}

// AddCollection add collection
func (mt *MetaTable) AddCollection(coll *model.Collection, ts typeutil.Timestamp, ddOpStr string) error {
	mt.ddLock.Lock()
	defer mt.ddLock.Unlock()

	if _, ok := mt.collName2ID[coll.Name]; ok {
		return fmt.Errorf("collection %s exist", coll.Name)
	}

	coll.CreateTime = ts
	for _, partition := range coll.Partitions {
		partition.PartitionCreatedTimestamp = ts
	}

	meta := map[string]string{}
	meta[DDMsgSendPrefix] = "false"
	meta[DDOperationPrefix] = ddOpStr
	coll.Extra = meta
	if err := mt.catalog.CreateCollection(mt.ctx, coll, ts); err != nil {
		return err
	}

	mt.collID2Meta[coll.CollectionID] = *coll
	mt.collName2ID[coll.Name] = coll.CollectionID
	return nil
}

// DeleteCollection delete collection
func (mt *MetaTable) DeleteCollection(collID typeutil.UniqueID, ts typeutil.Timestamp, ddOpStr string) error {
	mt.ddLock.Lock()
	defer mt.ddLock.Unlock()

	col, ok := mt.collID2Meta[collID]
	if !ok {
		return fmt.Errorf("can't find collection. id = %d", collID)
	}

	var aliases []string
	// delete collection aliases
	for alias, cid := range mt.collAlias2ID {
		if cid == collID {
			aliases = append(aliases, alias)
		}
	}

	// save ddOpStr into etcd
	var meta = map[string]string{
		DDMsgSendPrefix:   "false",
		DDOperationPrefix: ddOpStr,
	}

	collection := &model.Collection{
		CollectionID: collID,
		Aliases:      aliases,
		Extra:        meta,
	}

	if err := mt.catalog.DropCollection(mt.ctx, collection, ts); err != nil {
		return err
	}

	// update segID2IndexID
	for _, partition := range col.Partitions {
		partID := partition.PartitionID
		if segIDMap, ok := mt.partID2SegID[partID]; ok {
			for segID := range segIDMap {
				delete(mt.segID2IndexID, segID)
			}
		}
		delete(mt.partID2SegID, partID)
	}

	for _, t := range col.FieldIDToIndexID {
		delete(mt.indexID2Meta, t.Value)
	}

	// delete collection aliases
	for _, alias := range aliases {
		delete(mt.collAlias2ID, alias)
	}

	delete(mt.collID2Meta, collID)
	delete(mt.collName2ID, col.Name)

	return nil
}

// HasCollection return collection existence
func (mt *MetaTable) HasCollection(collID typeutil.UniqueID, ts typeutil.Timestamp) bool {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()
	if ts == 0 {
		_, ok := mt.collID2Meta[collID]
		return ok
	}

	return mt.catalog.CollectionExists(mt.ctx, collID, ts)
}

// GetCollectionIDByName returns the collection ID according to its name.
// Returns an error if no matching ID is found.
func (mt *MetaTable) GetCollectionIDByName(cName string) (typeutil.UniqueID, error) {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()
	var cID UniqueID
	var ok bool
	if cID, ok = mt.collName2ID[cName]; !ok {
		return 0, fmt.Errorf("collection ID not found for collection name %s", cName)
	}
	return cID, nil
}

// GetCollectionNameByID returns the collection name according to its ID.
// Returns an error if no matching name is found.
func (mt *MetaTable) GetCollectionNameByID(collectionID typeutil.UniqueID) (string, error) {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()
	col, ok := mt.collID2Meta[collectionID]
	if !ok {
		return "", fmt.Errorf("can't find collection id : %d", collectionID)
	}
	return col.Name, nil
}

// GetCollectionByID return collection meta by collection id
func (mt *MetaTable) GetCollectionByID(collectionID typeutil.UniqueID, ts typeutil.Timestamp) (*model.Collection, error) {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()

	if ts == 0 {
		col, ok := mt.collID2Meta[collectionID]
		if !ok {
			return nil, fmt.Errorf("can't find collection id : %d", collectionID)
		}
		return model.CloneCollectionModel(col), nil
	}

	return mt.catalog.GetCollectionByID(mt.ctx, collectionID, ts)
}

// GetCollectionByName return collection meta by collection name
func (mt *MetaTable) GetCollectionByName(collectionName string, ts typeutil.Timestamp) (*model.Collection, error) {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()

	if ts == 0 {
		vid, ok := mt.collName2ID[collectionName]
		if !ok {
			if vid, ok = mt.collAlias2ID[collectionName]; !ok {
				return nil, fmt.Errorf("can't find collection: " + collectionName)
			}
		}
		col, ok := mt.collID2Meta[vid]
		if !ok {
			return nil, fmt.Errorf("can't find collection %s with id %d", collectionName, vid)
		}

		return model.CloneCollectionModel(col), nil
	}

	return mt.catalog.GetCollectionByName(mt.ctx, collectionName, ts)
}

// ListCollections list all collection names
func (mt *MetaTable) ListCollections(ts typeutil.Timestamp) (map[string]*model.Collection, error) {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()
	cols := make(map[string]*model.Collection)

	if ts == 0 {
		for collName, collID := range mt.collName2ID {
			col := mt.collID2Meta[collID]
			cols[collName] = model.CloneCollectionModel(col)
		}
		return cols, nil
	}

	return mt.catalog.ListCollections(mt.ctx, ts)
}

// ListAliases list all collection aliases
func (mt *MetaTable) ListAliases(collID typeutil.UniqueID) []string {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()
	var aliases []string
	for alias, cid := range mt.collAlias2ID {
		if cid == collID {
			aliases = append(aliases, alias)
		}
	}
	return aliases
}

// ListCollectionVirtualChannels list virtual channels of all collections
func (mt *MetaTable) ListCollectionVirtualChannels() map[typeutil.UniqueID][]string {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()
	chanMap := make(map[typeutil.UniqueID][]string)

	for id, collInfo := range mt.collID2Meta {
		chanMap[id] = collInfo.VirtualChannelNames
	}
	return chanMap
}

// ListCollectionPhysicalChannels list physical channels of all collections
func (mt *MetaTable) ListCollectionPhysicalChannels() map[typeutil.UniqueID][]string {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()
	chanMap := make(map[typeutil.UniqueID][]string)

	for id, collInfo := range mt.collID2Meta {
		chanMap[id] = collInfo.PhysicalChannelNames
	}
	return chanMap
}

// AddPartition add partition
func (mt *MetaTable) AddPartition(collID typeutil.UniqueID, partitionName string, partitionID typeutil.UniqueID, ts typeutil.Timestamp, ddOpStr string) error {
	mt.ddLock.Lock()
	defer mt.ddLock.Unlock()
	coll, ok := mt.collID2Meta[collID]
	if !ok {
		return fmt.Errorf("can't find collection. id = %d", collID)
	}

	// number of partition tags (except _default) should be limited to 4096 by default
	if int64(len(coll.Partitions)) >= Params.RootCoordCfg.MaxPartitionNum {
		return fmt.Errorf("maximum partition's number should be limit to %d", Params.RootCoordCfg.MaxPartitionNum)
	}

	for _, p := range coll.Partitions {
		if p.PartitionID == partitionID {
			return fmt.Errorf("partition id = %d already exists", partitionID)
		}
		if p.PartitionName == partitionName {
			return fmt.Errorf("partition name = %s already exists", partitionName)
		}
		// no necessary to check created timestamp
	}

	coll.Partitions = append(coll.Partitions,
		&model.Partition{
			PartitionID:               partitionID,
			PartitionName:             partitionName,
			PartitionCreatedTimestamp: ts,
		})
	mt.collID2Meta[collID] = coll

	metaTxn := map[string]string{}
	// save ddOpStr into etcd
	metaTxn[DDMsgSendPrefix] = "false"
	metaTxn[DDOperationPrefix] = ddOpStr
	coll.Extra = metaTxn
	return mt.catalog.CreatePartition(mt.ctx, &coll, ts)
}

// GetPartitionNameByID return partition name by partition id
func (mt *MetaTable) GetPartitionNameByID(collID, partitionID typeutil.UniqueID, ts typeutil.Timestamp) (string, error) {
	if ts == 0 {
		mt.ddLock.RLock()
		defer mt.ddLock.RUnlock()
		col, ok := mt.collID2Meta[collID]
		if !ok {
			return "", fmt.Errorf("can't find collection id = %d", collID)
		}
		for _, partition := range col.Partitions {
			if partition.PartitionID == partitionID {
				return partition.PartitionName, nil
			}
		}
		return "", fmt.Errorf("partition %d does not exist", partitionID)
	}

	col, err := mt.catalog.GetCollectionByID(mt.ctx, collID, ts)
	if err != nil {
		return "", err
	}
	for _, partition := range col.Partitions {
		if partition.PartitionID == partitionID {
			return partition.PartitionName, nil
		}
	}
	return "", fmt.Errorf("partition %d does not exist", partitionID)
}

func (mt *MetaTable) getPartitionByName(collID typeutil.UniqueID, partitionName string, ts typeutil.Timestamp) (typeutil.UniqueID, error) {
	if ts == 0 {
		col, ok := mt.collID2Meta[collID]
		if !ok {
			return 0, fmt.Errorf("can't find collection id = %d", collID)
		}
		for _, partition := range col.Partitions {
			if partition.PartitionName == partitionName {
				return partition.PartitionID, nil
			}
		}
		return 0, fmt.Errorf("partition %s does not exist", partitionName)
	}

	col, err := mt.catalog.GetCollectionByID(mt.ctx, collID, ts)
	if err != nil {
		return 0, err
	}
	for _, partition := range col.Partitions {
		if partition.PartitionName == partitionName {
			return partition.PartitionID, nil
		}
	}
	return 0, fmt.Errorf("partition %s does not exist", partitionName)
}

// GetPartitionByName return partition id by partition name
func (mt *MetaTable) GetPartitionByName(collID typeutil.UniqueID, partitionName string, ts typeutil.Timestamp) (typeutil.UniqueID, error) {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()
	return mt.getPartitionByName(collID, partitionName, ts)
}

// HasPartition check partition existence
func (mt *MetaTable) HasPartition(collID typeutil.UniqueID, partitionName string, ts typeutil.Timestamp) bool {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()
	_, err := mt.getPartitionByName(collID, partitionName, ts)
	return err == nil
}

// DeletePartition delete partition
func (mt *MetaTable) DeletePartition(collID typeutil.UniqueID, partitionName string, ts typeutil.Timestamp, ddOpStr string) (typeutil.UniqueID, error) {
	mt.ddLock.Lock()
	defer mt.ddLock.Unlock()

	if partitionName == Params.CommonCfg.DefaultPartitionName {
		return 0, fmt.Errorf("default partition cannot be deleted")
	}

	col, ok := mt.collID2Meta[collID]
	if !ok {
		return 0, fmt.Errorf("can't find collection id = %d", collID)
	}

	// check tag exists
	exist := false

	parts := make([]*model.Partition, 0, len(col.Partitions))

	var partID typeutil.UniqueID
	for _, partition := range col.Partitions {
		if partition.PartitionName == partitionName {
			partID = partition.PartitionID
			exist = true
		} else {
			parts = append(parts, partition)
		}
	}
	if !exist {
		return 0, fmt.Errorf("partition %s does not exist", partitionName)
	}
	col.Partitions = parts
	mt.collID2Meta[collID] = col

	// update segID2IndexID and partID2SegID
	if segIDMap, ok := mt.partID2SegID[partID]; ok {
		for segID := range segIDMap {
			indexID, ok := mt.segID2IndexID[segID]
			if !ok {
				continue
			}
			delete(mt.segID2IndexID, segID)

			indexMeta, ok := mt.indexID2Meta[indexID]
			if ok {
				delete(indexMeta.SegmentIndexes, segID)
			}
		}
	}
	delete(mt.partID2SegID, partID)

	metaTxn := make(map[string]string)
	// save ddOpStr into etcd
	metaTxn[DDMsgSendPrefix] = "false"
	metaTxn[DDOperationPrefix] = ddOpStr
	col.Extra = metaTxn

	err := mt.catalog.DropPartition(mt.ctx, &col, partID, ts)
	if err != nil {
		return 0, err
	}

	return partID, nil
}

func (mt *MetaTable) updateSegmentIndexMetaCache(oldIndex *model.Index, index *model.Index) error {
	for _, segIdxInfo := range index.SegmentIndexes {
		if _, ok := mt.partID2SegID[segIdxInfo.PartitionID]; !ok {
			segIDMap := map[typeutil.UniqueID]bool{segIdxInfo.SegmentID: true}
			mt.partID2SegID[segIdxInfo.PartitionID] = segIDMap
		} else {
			mt.partID2SegID[segIdxInfo.PartitionID][segIdxInfo.SegmentID] = true
		}

		mt.segID2IndexID[segIdxInfo.SegmentID] = index.IndexID
	}

	for segID, segmentIdx := range index.SegmentIndexes {
		oldIndex.SegmentIndexes[segID] = segmentIdx
	}

	return nil
}

// AlterIndex alter index
func (mt *MetaTable) AlterIndex(newIndex *model.Index) error {
	mt.ddLock.Lock()
	defer mt.ddLock.Unlock()

	_, ok := mt.collID2Meta[newIndex.CollectionID]
	if !ok {
		return fmt.Errorf("collection id = %d not found", newIndex.CollectionID)
	}

	if oldIndex, ok := mt.indexID2Meta[newIndex.IndexID]; !ok || newIndex.GetDeleted() {
		log.Error("index id not found or has been deleted", zap.Int64("indexID", newIndex.IndexID))
		return fmt.Errorf("index id = %d not found", newIndex.IndexID)
	}

	if err := mt.catalog.AlterIndex(mt.ctx, oldIndex, newIndex); err != nil {
		return err
	}

	err := mt.updateSegmentIndexMetaCache(oldIndex, newIndex)
	return err
}

func (mt *MetaTable) MarkIndexDeleted(collName, fieldName, indexName string) error {
	mt.ddLock.Lock()
	defer mt.ddLock.Unlock()

	collMeta, err := mt.getCollectionInfoInternal(collName)
	if err != nil {
		log.Error("get collection meta failed", zap.String("collName", collName), zap.Error(err))
		return fmt.Errorf("collection name  = %s not has meta", collName)
	}
	fieldSch, err := mt.getFieldSchemaInternal(collName, fieldName)
	if err != nil {
		return err
	}

	var clonedIndex *pb.IndexInfo
	var dropIdxID typeutil.UniqueID
	for _, info := range collMeta.FieldIndexes {
		if info.FiledID != fieldSch.FieldID {
			continue
		}
		idxMeta, ok := mt.indexID2Meta[info.IndexID]
		if !ok {
			errMsg := fmt.Errorf("index not has meta with ID = %d", info.IndexID)
			log.Error("index id not has meta", zap.Int64("index id", info.IndexID))
			return errMsg
		}
		if idxMeta.GetDeleted() {
			continue
		}
		if idxMeta.IndexName != indexName {
			continue
		}
		dropIdxID = info.IndexID
		clonedIndex = proto.Clone(&idxMeta).(*pb.IndexInfo)
		clonedIndex.Deleted = true
	}

	if dropIdxID == 0 {
		log.Warn("index not found", zap.String("collName", collName), zap.String("fieldName", fieldName),
			zap.String("indexName", indexName))
		return nil
	}
	log.Info("MarkIndexDeleted", zap.String("collName", collName), zap.String("fieldName", fieldName),
		zap.String("indexName", indexName), zap.Int64("dropIndexID", dropIdxID))

	k := fmt.Sprintf("%s/%d/%d", IndexMetaPrefix, collMeta.ID, dropIdxID)
	v, err := proto.Marshal(clonedIndex)
	if err != nil {
		log.Error("MetaTable MarkIndexDeleted Marshal idxInfo fail",
			zap.String("key", k), zap.Error(err))
		return err
	}

	err = mt.txn.Save(k, string(v))
	if err != nil {
		log.Error("MetaTable MarkIndexDeleted txn MultiSave failed", zap.Error(err))
		return err
	}

	// update meta cache
	mt.indexID2Meta[dropIdxID] = *clonedIndex
	return nil
}

// DropIndex drop index
// Deprecated, only ut are used.
func (mt *MetaTable) DropIndex(collName, fieldName, indexName string) (typeutil.UniqueID, bool, error) {
	mt.ddLock.Lock()
	defer mt.ddLock.Unlock()

	collID, ok := mt.collName2ID[collName]
	if !ok {
		collID, ok = mt.collAlias2ID[collName]
		if !ok {
			return 0, false, fmt.Errorf("collection name = %s not exist", collName)
		}
	}
	col, ok := mt.collID2Meta[collID]
	if !ok {
		return 0, false, fmt.Errorf("collection name  = %s not has meta", collName)
	}
	fieldSch, err := mt.getFieldSchemaInternal(collName, fieldName)
	if err != nil {
		return 0, false, err
	}

	fieldIDToIndexID := make([]common.Int64Tuple, 0, len(col.FieldIDToIndexID))
	var dropIdxID typeutil.UniqueID
	for i, t := range col.FieldIDToIndexID {
		fieldID := t.Key
		indexID := t.Value

		if fieldID != fieldSch.FieldID {
			fieldIDToIndexID = append(fieldIDToIndexID, t)
			continue
		}

		idxMeta, ok := mt.indexID2Meta[indexID]
		if !ok || idxMeta.GetDeleted() || idxMeta.IndexName != indexName {
			fieldIDToIndexID = append(fieldIDToIndexID, t)
			log.Warn("index id not has meta", zap.Int64("index id", indexID))
			continue
		}
		dropIdxID = info.IndexID
		fieldIDToIndexID = append(fieldIDToIndexID, collMeta.FieldIndexes[i+1:]...)
		break
	}

	if len(fieldIDToIndexID) == len(col.FieldIDToIndexID) {
		log.Warn("drop index,index not found", zap.String("collection name", collName), zap.String("filed name", fieldName), zap.String("index name", indexName))
		return 0, false, nil
	}

	// update cache
	col.FieldIDToIndexID = fieldIDToIndexID
	mt.collID2Meta[collID] = col

	delete(mt.indexID2Meta, dropIdxID)
	for _, part := range col.Partitions {
		if segIDMap, ok := mt.partID2SegID[part.PartitionID]; ok {
			for segID := range segIDMap {
				delete(mt.segID2IndexID, segID)
			}
		}
	}

	// update metastore
	err = mt.catalog.DropIndex(mt.ctx, &col, dropIdxID, 0)
	if err != nil {
		return 0, false, err
	}

	return dropIdxID, true, nil
}

func (mt *MetaTable) GetInitBuildIDs(collName, indexName string) ([]UniqueID, error) {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()

	collMeta, err := mt.getCollectionInfoInternal(collName)
	if err != nil {
		return nil, err
	}

	var indexID typeutil.UniqueID
	for _, info := range collMeta.FieldIndexes {
		idxMeta, ok := mt.indexID2Meta[info.IndexID]
		if ok && idxMeta.IndexName == indexName {
			indexID = info.IndexID
			break
		}
	}

	if indexID == 0 {
		log.Warn("get init buildIDs, index not found", zap.String("collection name", collName),
			zap.String("index name", indexName))
		return nil, fmt.Errorf("index not found with name = %s in collection %s", indexName, collName)
	}

	initBuildIDs := make([]UniqueID, 0)
	for _, indexID2Info := range mt.segID2IndexMeta {
		segIndexInfo, ok := indexID2Info[indexID]
		if ok && segIndexInfo.EnableIndex && !segIndexInfo.ByAutoFlush {
			initBuildIDs = append(initBuildIDs, segIndexInfo.BuildID)
		}
	}
	return initBuildIDs, nil
}

func (mt *MetaTable) GetBuildIDsBySegIDs(segIDs []UniqueID) []UniqueID {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()

	buildIDs := make([]UniqueID, 0)
	for _, segID := range segIDs {
		if idxID2segIdx, ok := mt.segID2IndexMeta[segID]; ok {
			for _, segIndex := range idxID2segIdx {
				buildIDs = append(buildIDs, segIndex.BuildID)
			}
		}
	}
	return buildIDs
}

func (mt *MetaTable) AlignSegmentsMeta(collID, partID UniqueID, segIDs map[UniqueID]struct{}) ([]UniqueID, []UniqueID) {
	mt.ddLock.Lock()
	defer mt.ddLock.Unlock()

	allIndexID := make([]UniqueID, 0)
	if collMeta, ok := mt.collID2Meta[collID]; ok {
		for _, fieldIndex := range collMeta.FieldIndexes {
			allIndexID = append(allIndexID, fieldIndex.IndexID)
		}
	}

	recycledSegIDs := make([]UniqueID, 0)
	recycledBuildIDs := make([]UniqueID, 0)
	if segMap, ok := mt.partID2SegID[partID]; ok {
		for segID := range segMap {
			if _, ok := segIDs[segID]; !ok {
				recycledSegIDs = append(recycledSegIDs, segID)
			}
			if idxID2segIndex, ok := mt.segID2IndexMeta[segID]; ok {
				for _, segIndex := range idxID2segIndex {
					recycledBuildIDs = append(recycledBuildIDs, segIndex.BuildID)
				}
			}
		}
	}

	return recycledSegIDs, recycledBuildIDs
}

func (mt *MetaTable) RemoveSegments(collID, partID UniqueID, segIDs []UniqueID) error {
	mt.ddLock.Lock()
	defer mt.ddLock.Unlock()

	log.Info("RootCoord MetaTable remove segments", zap.Int64("collID", collID), zap.Int64("partID", partID),
		zap.Int64s("segIDs", segIDs))
	allIndexID := make([]UniqueID, 0)
	if collMeta, ok := mt.collID2Meta[collID]; ok {
		for _, fieldIndex := range collMeta.FieldIndexes {
			allIndexID = append(allIndexID, fieldIndex.IndexID)
		}
	}

	for _, segID := range segIDs {
		delMeta := make([]string, 0)
		for _, indexID := range allIndexID {
			delMeta = append(delMeta, fmt.Sprintf("%s/%d/%d/%d/%d", SegmentIndexMetaPrefix, collID, indexID, partID, segID))
		}
		if err := mt.txn.MultiRemove(delMeta); err != nil {
			log.Error("remove redundant segment failed, wait to retry", zap.Int64("collID", collID), zap.Int64("part", partID),
				zap.Int64("segID", segID), zap.Error(err))
			return err
		}
		delete(mt.partID2SegID[partID], segID)
		delete(mt.segID2IndexMeta, segID)
	}

	return nil
}

func (mt *MetaTable) GetDroppedIndex() map[UniqueID][]*pb.FieldIndexInfo {
	mt.ddLock.Lock()
	defer mt.ddLock.Unlock()

	droppedIndex := make(map[UniqueID][]*pb.FieldIndexInfo)
	for collID, meta := range mt.collID2Meta {
		for _, fieldIndex := range meta.FieldIndexes {
			if indexMeta, ok := mt.indexID2Meta[fieldIndex.IndexID]; ok && indexMeta.Deleted {
				droppedIndex[collID] = append(droppedIndex[collID], proto.Clone(fieldIndex).(*pb.FieldIndexInfo))
			}
		}
	}
	return droppedIndex
}

// RecycleDroppedIndex remove the meta about index which is deleted.
func (mt *MetaTable) RecycleDroppedIndex() error {
	mt.ddLock.Lock()
	defer mt.ddLock.Unlock()

	for collID, collMeta := range mt.collID2Meta {
		meta := collMeta
		fieldIndexes := make([]*pb.FieldIndexInfo, 0)
		delMeta := make([]string, 0)
		deletedIndexIDs := make(map[UniqueID]struct{})
		for _, fieldIndex := range meta.FieldIndexes {
			// Prevent the number of transaction operations from exceeding the etcd limit (128), where a maximum of 100 are processed each time
			if len(delMeta) >= 100 {
				break
			}
			if idxInfo, ok := mt.indexID2Meta[fieldIndex.IndexID]; !ok || idxInfo.GetDeleted() {
				deletedIndexIDs[fieldIndex.IndexID] = struct{}{}
				delMeta = append(delMeta, fmt.Sprintf("%s/%d/%d", IndexMetaPrefix, collID, fieldIndex.IndexID))
				delMeta = append(delMeta, fmt.Sprintf("%s/%d/%d", SegmentIndexMetaPrefix, collID, fieldIndex.IndexID))
				continue
			}
			fieldIndexes = append(fieldIndexes, fieldIndex)
		}
		// node index is deleted
		if len(fieldIndexes) == len(meta.FieldIndexes) {
			continue
		}
		clonedCollMeta := proto.Clone(&meta).(*pb.CollectionInfo)
		clonedCollMeta.FieldIndexes = fieldIndexes

		saveMeta := make(map[string]string)
		k := path.Join(CollectionMetaPrefix, strconv.FormatInt(collID, 10))
		v, err := proto.Marshal(clonedCollMeta)
		if err != nil {
			log.Error("MetaTable RecycleDroppedIndex Marshal collMeta failed",
				zap.String("key", k), zap.Error(err))
			return err
		}
		saveMeta[k] = string(v)

		if err = mt.txn.MultiSaveAndRemoveWithPrefix(saveMeta, delMeta); err != nil {
			log.Error("MetaTable RecycleDroppedIndex MultiSaveAndRemoveWithPrefix failed", zap.Error(err))
			return err
		}
		mt.collID2Meta[collID] = *clonedCollMeta
		for indexID := range deletedIndexIDs {
			delete(mt.indexID2Meta, indexID)
		}
		// update segID2IndexMeta
		for _, partID := range meta.PartitionIDs {
			if segIDMap, ok := mt.partID2SegID[partID]; ok {
				for segID := range segIDMap {
					if segIndexInfos, ok := mt.segID2IndexMeta[segID]; ok {
						for indexID := range segIndexInfos {
							if _, ok := deletedIndexIDs[indexID]; ok {
								delete(mt.segID2IndexMeta[segID], indexID)
							}
						}
					}
				}
			}
		}
	}
	return nil
}

// GetSegmentIndexInfoByID return segment index info by segment id
func (mt *MetaTable) GetSegmentIndexInfoByID(segID typeutil.UniqueID, fieldID int64, idxName string) (model.Index, error) {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()
	idxMeta, err := mt.getIdxMetaBySegID(segID)
	if err != nil {
		return model.Index{}, err
	}

	// return default index
	if fieldID == -1 && idxName == "" && idxMeta.IndexName == Params.CommonCfg.DefaultIndexName {
		return idxMeta, nil
	}

	if idxMeta.IndexName == idxName && idxMeta.FieldID == fieldID {
		return idxMeta, nil
	}

	return model.Index{}, fmt.Errorf("can't find index name = %s on segment = %d, with filed id = %d", idxName, segID, fieldID)
	if fieldID == -1 && idxName == "" { // return default index
		for _, seg := range segIdxMap {
			info, ok := mt.indexID2Meta[seg.IndexID]
			if ok && !info.GetDeleted() && info.IndexName == Params.CommonCfg.DefaultIndexName {
				return seg, nil
			}
		}
	} else {
		for idxID, seg := range segIdxMap {
			idxMeta, ok := mt.indexID2Meta[idxID]
			if ok && !idxMeta.GetDeleted() && idxMeta.IndexName == idxName && seg.FieldID == fieldID {
				return seg, nil
			}
		}
	}
	return pb.SegmentIndexInfo{}, fmt.Errorf("can't find index name = %s on segment = %d, with filed id = %d", idxName, segID, fieldID)
}

func (mt *MetaTable) GetSegmentIndexInfos(segID typeutil.UniqueID) (model.Index, error) {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()
	return mt.getIdxMetaBySegID(segID)
}

// GetFieldSchema return field schema
func (mt *MetaTable) GetFieldSchema(collName string, fieldName string) (model.Field, error) {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()

	return mt.getFieldSchemaInternal(collName, fieldName)
}

func (mt *MetaTable) getFieldSchemaInternal(collName string, fieldName string) (model.Field, error) {
	collID, ok := mt.collName2ID[collName]
	if !ok {
		collID, ok = mt.collAlias2ID[collName]
		if !ok {
			return model.Field{}, fmt.Errorf("collection %s not found", collName)
		}
	}
	col, ok := mt.collID2Meta[collID]
	if !ok {
		return model.Field{}, fmt.Errorf("collection %s not found", collName)
	}

	for _, field := range col.Fields {
		if field.Name == fieldName {
			return *field, nil
		}
	}
	return model.Field{}, fmt.Errorf("collection %s doesn't have filed %s", collName, fieldName)
}

// IsSegmentIndexed check if segment has indexed
func (mt *MetaTable) IsSegmentIndexed(segID typeutil.UniqueID, fieldSchema *model.Field, indexParams []*commonpb.KeyValuePair) bool {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()
	return mt.isSegmentIndexedInternal(segID, fieldSchema, indexParams)
}

func (mt *MetaTable) isSegmentIndexedInternal(segID typeutil.UniqueID, fieldSchema *model.Field, indexParams []*commonpb.KeyValuePair) bool {
	index, err := mt.getIdxMetaBySegID(segID)
	if err != nil {
		return false
	}

	segIndex, ok := index.SegmentIndexes[segID]
	if ok && !idxMeta.GetDeleted() &&
		index.FieldID == fieldSchema.FieldID &&
		EqualKeyPairArray(indexParams, index.IndexParams) &&
		segIndex.EnableIndex {
		return true
	}

	return false
}

func (mt *MetaTable) getCollectionInfoInternal(collName string) (model.Collection, error) {
	collID, ok := mt.collName2ID[collName]
	if !ok {
		collID, ok = mt.collAlias2ID[collName]
		if !ok {
			return model.Collection{}, fmt.Errorf("collection not found: %s", collName)
		}
	}
	collMeta, ok := mt.collID2Meta[collID]
	if !ok {
		return model.Collection{}, fmt.Errorf("collection not found: %s", collName)
	}
	return collMeta, nil
}

func (mt *MetaTable) checkFieldCanBeIndexed(collMeta model.Collection, fieldSchema model.Field, idxInfo *model.Index) error {
	for _, tuple := range collMeta.FieldIDToIndexID {
		if tuple.Key == fieldSchema.FieldID {
			if info, ok := mt.indexID2Meta[tuple.Value]; ok {
				if info.GetDeleted() {
					continue
				}

				if idxInfo.IndexName != info.IndexName {
					return fmt.Errorf(
						"creating multiple indexes on same field is not supported, "+
							"collection: %s, field: %s, index name: %s, new index name: %s",
						collMeta.Name, fieldSchema.Name,
						info.IndexName, idxInfo.IndexName)
				}
			} else {
				// TODO: unexpected: what if index id not exist? Meta incomplete.
				log.Warn("index meta was incomplete, index id missing in indexID2Meta",
					zap.String("collection", collMeta.Name),
					zap.String("field", fieldSchema.Name),
					zap.Int64("collection id", collMeta.CollectionID),
					zap.Int64("field id", fieldSchema.FieldID),
					zap.Int64("index id", tuple.Value))
			}
		}
	}
	return nil
}

func (mt *MetaTable) checkFieldIndexDuplicate(collMeta model.Collection, fieldSchema model.Field, idxInfo *model.Index) (duplicate bool, err error) {
	for _, t := range collMeta.FieldIDToIndexID {
		if info, ok := mt.indexID2Meta[t.Value]; ok && !info.GetDeleted() {
			if info.IndexName == idxInfo.IndexName {
				// the index name must be different for different indexes
				if t.Key != fieldSchema.FieldID || !EqualKeyPairArray(info.IndexParams, idxInfo.IndexParams) {
					return false, fmt.Errorf("index already exists, collection: %s, field: %s, index: %s", collMeta.Name, fieldSchema.Name, idxInfo.IndexName)
				}

				// same index name, index params, and fieldId
				return true, nil
			}
		}
	}
	return false, nil
}

// GetNotIndexedSegments return segment ids which have no index
func (mt *MetaTable) GetNotIndexedSegments(collName string, fieldName string, idxInfo *model.Index, segIDs []typeutil.UniqueID) ([]typeutil.UniqueID, model.Field, error) {
	mt.ddLock.Lock()
	defer mt.ddLock.Unlock()

	fieldSchema, err := mt.getFieldSchemaInternal(collName, fieldName)
	if err != nil {
		return nil, fieldSchema, err
	}

	rstID := make([]typeutil.UniqueID, 0, 16)
	for _, segID := range segIDs {
		if ok := mt.isSegmentIndexedInternal(segID, &fieldSchema, idxInfo.IndexParams); !ok {
			rstID = append(rstID, segID)
		}
	}
	return rstID, fieldSchema, nil
}

// AddIndex add index
func (mt *MetaTable) AddIndex(colName string, fieldName string, idxInfo *model.Index, segIDs []typeutil.UniqueID) (bool, error) {
	mt.ddLock.Lock()
	defer mt.ddLock.Unlock()

	fieldSchema, err := mt.getFieldSchemaInternal(colName, fieldName)
	if err != nil {
		return false, err
	}

	collMeta, err := mt.getCollectionInfoInternal(colName)
	if err != nil {
		// error here if collection not found.
		return false, err
	}

	//TODO:: check index params for sclar field
	// set default index type for scalar index
	if !typeutil.IsVectorType(fieldSchema.DataType) {
		if fieldSchema.DataType == schemapb.DataType_VarChar {
			idxInfo.IndexParams = []*commonpb.KeyValuePair{{Key: "index_type", Value: DefaultStringIndexType}}
		} else {
			idxInfo.IndexParams = []*commonpb.KeyValuePair{{Key: "index_type", Value: DefaultIndexType}}
		}
	}

	if idxInfo.IndexParams == nil {
		return false, fmt.Errorf("index param is nil")
	}

	if err := mt.checkFieldCanBeIndexed(collMeta, fieldSchema, idxInfo); err != nil {
		return false, err
	}

	dupIdx, err := mt.checkFieldIndexDuplicate(collMeta, fieldSchema, idxInfo)
	if err != nil {
		// error here if index already exists.
		return dupIdx, err
	}

	if dupIdx {
		log.Warn("due to index already exists, skip add index to metastore", zap.Int64("collectionID", collMeta.CollectionID),
			zap.Int64("indexID", idxInfo.IndexID), zap.String("indexName", idxInfo.IndexName))
		// skip already exist index
		return dupIdx, nil
	}

	segmentIndexes := make(map[int64]model.SegmentIndex, len(segIDs))
	for _, segID := range segIDs {
		segmentIndex := model.SegmentIndex{
			Segment: model.Segment{
				SegmentID: segID,
			},
			EnableIndex: false,
		}
		segmentIndexes[segID] = segmentIndex
	}

	idxInfo.SegmentIndexes = segmentIndexes
	idxInfo.FieldID = fieldSchema.FieldID
	idxInfo.CollectionID = collMeta.CollectionID

	tuple := common.Int64Tuple{
		Key:   fieldSchema.FieldID,
		Value: idxInfo.IndexID,
	}
	collMeta.FieldIDToIndexID = append(collMeta.FieldIDToIndexID, tuple)
	if err := mt.catalog.CreateIndex(mt.ctx, &collMeta, idxInfo); err != nil {
		return false, nil
	}

	mt.collID2Meta[collMeta.CollectionID] = collMeta
	mt.indexID2Meta[idxInfo.IndexID] = idxInfo

	return false, nil
}

// GetIndexByName return index info by index name
func (mt *MetaTable) GetIndexByName(collName, indexName string) (model.Collection, []model.Index, error) {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()

	collID, ok := mt.collName2ID[collName]
	if !ok {
		collID, ok = mt.collAlias2ID[collName]
		if !ok {
			return model.Collection{}, nil, fmt.Errorf("collection %s not found", collName)
		}
	}
	col, ok := mt.collID2Meta[collID]
	if !ok {
		return model.Collection{}, nil, fmt.Errorf("collection %s not found", collName)
	}

	rstIndex := make([]model.Index, 0, len(col.FieldIDToIndexID))
	for _, t := range col.FieldIDToIndexID {
		indexID := t.Value
		idxInfo, ok := mt.indexID2Meta[indexID]
		if !ok {
			return model.Collection{}, nil, fmt.Errorf("index id = %d not found", indexID)
		}
		if idxInfo.GetDeleted() {
			continue
		}
		if indexName == "" || idxInfo.IndexName == indexName {
			rstIndex = append(rstIndex, *idxInfo)
		}
	}
	return col, rstIndex, nil
}

// GetIndexByID return index info by index id
func (mt *MetaTable) GetIndexByID(indexID typeutil.UniqueID) (*model.Index, error) {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()

	indexInfo, ok := mt.indexID2Meta[indexID]
	if !ok || indexInfo.GetDeleted() {
		return nil, fmt.Errorf("cannot find index, id = %d", indexID)
	}
	return indexInfo, nil
}

func (mt *MetaTable) dupCollectionMeta() map[typeutil.UniqueID]pb.CollectionInfo {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()

	collID2Meta := map[typeutil.UniqueID]pb.CollectionInfo{}
	for k, v := range mt.collID2Meta {
		v := v
		collID2Meta[k] = *proto.Clone(&v).(*pb.CollectionInfo)
	}
	return collID2Meta
}

func (mt *MetaTable) dupMeta() (
	map[typeutil.UniqueID]model.Collection,
	map[typeutil.UniqueID]typeutil.UniqueID,
	map[typeutil.UniqueID]model.Index,
) {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()

	collID2Meta := make(map[typeutil.UniqueID]model.Collection, len(mt.collID2Meta))
	segID2IndexID := make(map[typeutil.UniqueID]typeutil.UniqueID, len(mt.segID2IndexID))
	indexID2Meta := make(map[typeutil.UniqueID]model.Index, len(mt.indexID2Meta))
	for k, v := range mt.collID2Meta {
		collID2Meta[k] = v
	}
	for k, v := range mt.segID2IndexID {
		segID2IndexID[k] = v
	}
	for k, v := range mt.indexID2Meta {
		indexID2Meta[k] = *v
	}
	return collID2Meta, segID2IndexID, indexID2Meta
}

// AddAlias add collection alias
func (mt *MetaTable) AddAlias(collectionAlias string, collectionName string, ts typeutil.Timestamp) error {
	mt.ddLock.Lock()
	defer mt.ddLock.Unlock()
	if _, ok := mt.collAlias2ID[collectionAlias]; ok {
		return fmt.Errorf("duplicate collection alias, alias = %s", collectionAlias)
	}

	if _, ok := mt.collName2ID[collectionAlias]; ok {
		return fmt.Errorf("collection alias collides with existing collection name. collection = %s, alias = %s", collectionAlias, collectionAlias)
	}

	id, ok := mt.collName2ID[collectionName]
	if !ok {
		return fmt.Errorf("aliased collection name does not exist, name = %s", collectionName)
	}
	mt.collAlias2ID[collectionAlias] = id

	coll := &model.Collection{
		CollectionID: id,
		Aliases:      []string{collectionAlias},
	}
	return mt.catalog.CreateAlias(mt.ctx, coll, ts)
}

// DropAlias drop collection alias
func (mt *MetaTable) DropAlias(collectionAlias string, ts typeutil.Timestamp) error {
	mt.ddLock.Lock()
	defer mt.ddLock.Unlock()
	collectionID, ok := mt.collAlias2ID[collectionAlias]
	if !ok {
		return fmt.Errorf("alias does not exist, alias = %s", collectionAlias)
	}
	delete(mt.collAlias2ID, collectionAlias)

	return mt.catalog.DropAlias(mt.ctx, collectionID, collectionAlias, ts)
}

// AlterAlias alter collection alias
func (mt *MetaTable) AlterAlias(collectionAlias string, collectionName string, ts typeutil.Timestamp) error {
	mt.ddLock.Lock()
	defer mt.ddLock.Unlock()
	if _, ok := mt.collAlias2ID[collectionAlias]; !ok {
		return fmt.Errorf("alias does not exist, alias = %s", collectionAlias)
	}

	id, ok := mt.collName2ID[collectionName]
	if !ok {
		return fmt.Errorf("aliased collection name does not exist, name = %s", collectionName)
	}
	mt.collAlias2ID[collectionAlias] = id

	coll := &model.Collection{
		CollectionID: id,
		Aliases:      []string{collectionAlias},
	}
	return mt.catalog.AlterAlias(mt.ctx, coll, ts)
}

// IsAlias returns true if specific `collectionAlias` is an alias of collection.
func (mt *MetaTable) IsAlias(collectionAlias string) bool {
	mt.ddLock.RLock()
	defer mt.ddLock.RUnlock()
	_, ok := mt.collAlias2ID[collectionAlias]
	return ok
}

// AddCredential add credential
func (mt *MetaTable) AddCredential(credInfo *internalpb.CredentialInfo) error {
	if credInfo.Username == "" {
		return fmt.Errorf("username is empty")
	}

	credential := &model.Credential{
		Username:          credInfo.Username,
		EncryptedPassword: credInfo.EncryptedPassword,
	}
	return mt.catalog.CreateCredential(mt.ctx, credential)
}

// GetCredential get credential by username
func (mt *MetaTable) getCredential(username string) (*internalpb.CredentialInfo, error) {
	credential, err := mt.catalog.GetCredential(mt.ctx, username)
	return model.ConvertToCredentialPB(credential), err
}

// DeleteCredential delete credential
func (mt *MetaTable) DeleteCredential(username string) error {
	return mt.catalog.DropCredential(mt.ctx, username)
}

// ListCredentialUsernames list credential usernames
func (mt *MetaTable) ListCredentialUsernames() (*milvuspb.ListCredUsersResponse, error) {
	usernames, err := mt.catalog.ListCredentials(mt.ctx)
	if err != nil {
		return nil, fmt.Errorf("list credential usernames err:%w", err)
	}
	return &milvuspb.ListCredUsersResponse{Usernames: usernames}, nil
}

func (mt *MetaTable) getIdxMetaBySegID(segID int64) (model.Index, error) {
	indexID, ok := mt.segID2IndexID[segID]
	if !ok {
		return model.Index{}, fmt.Errorf("segment not found in meta, segment: %d", segID)
	}

	idxMeta, ok := mt.indexID2Meta[indexID]
	if !ok {
		return model.Index{}, fmt.Errorf("segment id: %d not has any index, request index id: %d", segID, indexID)
	}

	return *idxMeta, nil
}
