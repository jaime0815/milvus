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

package proxy

import (
	"context"
	"strconv"
	"time"

	"github.com/milvus-io/milvus-proto/go-api/v2/commonpb"
	"github.com/milvus-io/milvus-proto/go-api/v2/milvuspb"
	"github.com/milvus-io/milvus-proto/go-api/v2/msgpb"
	"github.com/milvus-io/milvus-proto/go-api/v2/schemapb"
	"github.com/milvus-io/milvus/internal/allocator"
	"github.com/milvus-io/milvus/pkg/log"
	"github.com/milvus-io/milvus/pkg/metrics"
	"github.com/milvus-io/milvus/pkg/mq/msgstream"
	"github.com/milvus-io/milvus/pkg/util/commonpbutil"
	"github.com/milvus-io/milvus/pkg/util/paramtable"
	"github.com/milvus-io/milvus/pkg/util/retry"
	"github.com/milvus-io/milvus/pkg/util/typeutil"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
)

func buildInsertMsg(ctx context.Context, channelName string, partitionID UniqueID, partitionName string, segmentID UniqueID,
	insertMsg *msgstream.InsertMsg, rows int) *msgstream.InsertMsg {
	insertReq := &msgpb.InsertRequest{
		Base: commonpbutil.NewMsgBase(
			commonpbutil.WithMsgType(commonpb.MsgType_Insert),
			commonpbutil.WithTimeStamp(insertMsg.BeginTimestamp), // entity's timestamp was set to equal it.BeginTimestamp in preExecute()
			commonpbutil.WithSourceID(insertMsg.Base.SourceID),
		),
		CollectionID:   insertMsg.CollectionID,
		PartitionID:    partitionID,
		CollectionName: insertMsg.CollectionName,
		PartitionName:  partitionName,
		SegmentID:      segmentID,
		ShardName:      channelName,
		Version:        msgpb.InsertDataVersion_ColumnBased,
		FieldsData:     make([]*schemapb.FieldData, rows),
		Timestamps:     make([]uint64, 0, rows),
		RowIDs:         make([]int64, 0, rows),
	}

	return &msgstream.InsertMsg{
		BaseMsg: msgstream.BaseMsg{
			Ctx: ctx,
		},
		InsertRequest: insertReq,
	}
}

func protoSize(msg *msgstream.InsertMsg) (int, error) {
	curRowBytes, err := proto.Marshal(msg)
	if err != nil {
		log.Error("marshal insert msg fail", zap.Error(err))
		return 0, err
	}
	return len(curRowBytes), nil
}

func appendMsg(target, source *msgstream.InsertMsg, sourceOffset int) {
	typeutil.AppendFieldData(target.FieldsData, source.GetFieldsData(), int64(sourceOffset))
	target.HashValues = append(target.HashValues, source.HashValues[sourceOffset])
	target.Timestamps = append(target.Timestamps, source.Timestamps[sourceOffset])
	target.RowIDs = append(target.RowIDs, source.RowIDs[sourceOffset])
	target.NumRows++
}

// genInsertMsgsByPartition creates TsMsgs using insertMsg. If a tsMsg exceeds the size threshold, it will be split into multiple messages.
// Note That: Be aware that the encoded message size may still surpass the configured `pulsar.maxMessageSize` with the
// current splitting strategy. In such case, you can adjust `proxy.maxMessageSize` accordingly.
func genInsertMsgsByPartition(ctx context.Context,
	segmentID UniqueID,
	partitionID UniqueID,
	partitionName string,
	rowOffsets []int,
	channelName string,
	insertMsg *msgstream.InsertMsg,
) ([]msgstream.TsMsg, error) {
	repackedMsgs := make([]msgstream.TsMsg, 0)
	if len(rowOffsets) == 0 {
		return repackedMsgs, nil
	}

	msg := buildInsertMsg(ctx, channelName, partitionID, partitionName, segmentID, insertMsg, len(insertMsg.FieldsData))
	appendMsg(msg, insertMsg, rowOffsets[0])
	if len(rowOffsets) == 1 {
		repackedMsgs = append(repackedMsgs, msg)
		return repackedMsgs, nil
	}

	rowEncodedSize, err := protoSize(msg)
	if err != nil {
		return nil, err
	}

	threshold := Params.PulsarCfg.MaxMessageSize.GetAsInt()
	eachGroupRows := threshold / rowEncodedSize
	for idx := 1; idx < len(rowOffsets); idx++ {
		offset := rowOffsets[idx]
		if idx%eachGroupRows == 0 {
			repackedMsgs = append(repackedMsgs, msg)
			msg = buildInsertMsg(ctx, channelName, partitionID, partitionName, segmentID, insertMsg, eachGroupRows)
		}
		appendMsg(msg, insertMsg, offset)
	}

	repackedMsgs = append(repackedMsgs, msg)
	return repackedMsgs, nil
}

func repackInsertDataByPartition(ctx context.Context,
	partitionName string,
	rowOffsets []int,
	channelName string,
	insertMsg *msgstream.InsertMsg,
	segIDAssigner *segIDAssigner,
) ([]msgstream.TsMsg, error) {
	res := make([]msgstream.TsMsg, 0)

	maxTs := Timestamp(0)
	for _, offset := range rowOffsets {
		ts := insertMsg.Timestamps[offset]
		if maxTs < ts {
			maxTs = ts
		}
	}

	partitionID, err := globalMetaCache.GetPartitionID(ctx, insertMsg.GetDbName(), insertMsg.CollectionName, partitionName)
	if err != nil {
		return nil, err
	}
	beforeAssign := time.Now()
	assignedSegmentInfos, err := segIDAssigner.GetSegmentID(insertMsg.CollectionID, partitionID, channelName, uint32(len(rowOffsets)), maxTs)
	metrics.ProxyAssignSegmentIDLatency.WithLabelValues(strconv.FormatInt(paramtable.GetNodeID(), 10)).Observe(float64(time.Since(beforeAssign).Milliseconds()))
	if err != nil {
		log.Error("allocate segmentID for insert data failed",
			zap.String("collectionName", insertMsg.CollectionName),
			zap.String("channelName", channelName),
			zap.Int("allocate count", len(rowOffsets)),
			zap.Error(err))
		return nil, err
	}

	startPos := 0
	for segmentID, count := range assignedSegmentInfos {
		subRowOffsets := rowOffsets[startPos : startPos+int(count)]
		msgs, err := genInsertMsgsByPartition(ctx, segmentID, partitionID, partitionName, subRowOffsets, channelName, insertMsg)
		if err != nil {
			log.Warn("repack insert data to insert msgs failed",
				zap.String("collectionName", insertMsg.CollectionName),
				zap.Int64("partitionID", partitionID),
				zap.Error(err))
			return nil, err
		}
		res = append(res, msgs...)
		startPos += int(count)
	}

	return res, nil
}

func setMsgID(ctx context.Context,
	msgs []msgstream.TsMsg,
	idAllocator *allocator.IDAllocator,
) error {
	var idBegin int64
	var err error

	err = retry.Do(ctx, func() error {
		idBegin, _, err = idAllocator.Alloc(uint32(len(msgs)))
		return err
	})
	if err != nil {
		log.Error("failed to allocate msg id", zap.Error(err))
		return err
	}

	for i, msg := range msgs {
		msg.SetID(idBegin + UniqueID(i))
	}

	return nil
}

func repackInsertData(ctx context.Context,
	channelNames []string,
	insertMsg *msgstream.InsertMsg,
	result *milvuspb.MutationResult,
	idAllocator *allocator.IDAllocator,
	segIDAssigner *segIDAssigner,
) (*msgstream.MsgPack, error) {
	msgPack := &msgstream.MsgPack{
		BeginTs: insertMsg.BeginTs(),
		EndTs:   insertMsg.EndTs(),
	}

	channel2RowOffsets := assignChannelsByPK(result.IDs, channelNames, insertMsg)
	for channel, rowOffsets := range channel2RowOffsets {
		partitionName := insertMsg.PartitionName
		msgs, err := repackInsertDataByPartition(ctx, partitionName, rowOffsets, channel, insertMsg, segIDAssigner)
		if err != nil {
			log.Warn("repack insert data to msg pack failed",
				zap.String("collectionName", insertMsg.CollectionName),
				zap.String("partition name", partitionName),
				zap.Error(err))
			return nil, err
		}

		msgPack.Msgs = append(msgPack.Msgs, msgs...)
	}

	err := setMsgID(ctx, msgPack.Msgs, idAllocator)
	if err != nil {
		log.Error("failed to set msgID when repack insert data",
			zap.String("collectionName", insertMsg.CollectionName),
			zap.String("partition name", insertMsg.PartitionName),
			zap.Error(err))
		return nil, err
	}

	return msgPack, nil
}

func repackInsertDataWithPartitionKey(ctx context.Context,
	channelNames []string,
	partitionKeys *schemapb.FieldData,
	insertMsg *msgstream.InsertMsg,
	result *milvuspb.MutationResult,
	idAllocator *allocator.IDAllocator,
	segIDAssigner *segIDAssigner,
) (*msgstream.MsgPack, error) {
	msgPack := &msgstream.MsgPack{
		BeginTs: insertMsg.BeginTs(),
		EndTs:   insertMsg.EndTs(),
	}

	channel2RowOffsets := assignChannelsByPK(result.IDs, channelNames, insertMsg)
	partitionNames, err := getDefaultPartitionsInPartitionKeyMode(ctx, insertMsg.GetDbName(), insertMsg.CollectionName)
	if err != nil {
		log.Warn("get default partition names failed in partition key mode",
			zap.String("collectionName", insertMsg.CollectionName),
			zap.Error(err))
		return nil, err
	}
	hashValues, err := typeutil.HashKey2Partitions(partitionKeys, partitionNames)
	if err != nil {
		log.Warn("has partition keys to partitions failed",
			zap.String("collectionName", insertMsg.CollectionName),
			zap.Error(err))
		return nil, err
	}

	for channel, rowOffsets := range channel2RowOffsets {
		partition2RowOffsets := make(map[string][]int)
		for _, idx := range rowOffsets {
			partitionName := partitionNames[hashValues[idx]]
			if _, ok := partition2RowOffsets[partitionName]; !ok {
				partition2RowOffsets[partitionName] = []int{}
			}
			partition2RowOffsets[partitionName] = append(partition2RowOffsets[partitionName], idx)
		}

		errGroup, _ := errgroup.WithContext(ctx)
		partition2Msgs := typeutil.NewConcurrentMap[string, []msgstream.TsMsg]()
		for partitionName, offsets := range partition2RowOffsets {
			partitionName := partitionName
			offsets := offsets
			errGroup.Go(func() error {
				msgs, err := repackInsertDataByPartition(ctx, partitionName, offsets, channel, insertMsg, segIDAssigner)
				if err != nil {
					return err
				}

				partition2Msgs.Insert(partitionName, msgs)
				return nil
			})
		}

		err = errGroup.Wait()
		if err != nil {
			log.Warn("repack insert data into insert msg pack failed",
				zap.String("collectionName", insertMsg.CollectionName),
				zap.String("channelName", channel),
				zap.Error(err))
			return nil, err
		}

		partition2Msgs.Range(func(name string, msgs []msgstream.TsMsg) bool {
			msgPack.Msgs = append(msgPack.Msgs, msgs...)
			return true
		})
	}

	err = setMsgID(ctx, msgPack.Msgs, idAllocator)
	if err != nil {
		log.Error("failed to set msgID when repack insert data",
			zap.String("collectionName", insertMsg.CollectionName),
			zap.Error(err))
		return nil, err
	}

	return msgPack, nil
}
