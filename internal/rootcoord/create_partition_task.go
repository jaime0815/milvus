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

	"go.uber.org/zap"

	"github.com/milvus-io/milvus-proto/go-api/commonpb"
	"github.com/milvus-io/milvus-proto/go-api/milvuspb"
	"github.com/milvus-io/milvus/internal/log"
	"github.com/milvus-io/milvus/internal/metastore/model"
	pb "github.com/milvus-io/milvus/internal/proto/etcdpb"
	"github.com/milvus-io/milvus/internal/util/typeutil"
)

type createPartitionTask struct {
	baseTask
	Req      *milvuspb.CreatePartitionRequest
	collMeta *model.Collection
}

func (t *createPartitionTask) Prepare(ctx context.Context) error {
	t.SetStep(typeutil.TaskStepPreExecute)
	if err := CheckMsgType(t.Req.GetBase().GetMsgType(), commonpb.MsgType_CreatePartition); err != nil {
		return err
	}
	collMeta, err := t.core.meta.GetCollectionByName(ctx, t.Req.GetDbName(), t.Req.GetCollectionName(), t.GetTs())
	if err != nil {
		return err
	}
	t.collMeta = collMeta
	return nil
}

func (t *createPartitionTask) Execute(ctx context.Context) error {
	t.SetStep(typeutil.TaskStepExecute)
	for _, partition := range t.collMeta.Partitions {
		if partition.PartitionName == t.Req.GetPartitionName() {
			log.Warn("add duplicate partition", zap.String("collection", t.Req.GetCollectionName()), zap.String("partition", t.Req.GetPartitionName()), zap.Uint64("ts", t.GetTs()))
			return nil
		}
	}

	cfgMaxPartitionNum := Params.RootCoordCfg.MaxPartitionNum
	if int64(len(t.collMeta.Partitions)) >= cfgMaxPartitionNum {
		return fmt.Errorf("partition number (%d) exceeds max configuration (%d), collection: %s",
			len(t.collMeta.Partitions), cfgMaxPartitionNum, t.collMeta.Name)
	}

	partID, err := t.core.idAllocator.AllocOne()
	if err != nil {
		return err
	}
	partition := &model.Partition{
		PartitionID:               partID,
		PartitionName:             t.Req.GetPartitionName(),
		PartitionCreatedTimestamp: t.GetTs(),
		Extra:                     nil,
		CollectionID:              t.collMeta.CollectionID,
		State:                     pb.PartitionState_PartitionCreated,
	}

	undoTask := newBaseUndoTask(t.core.stepExecutor)
	undoTask.AddStep(&expireCacheStep{
		baseStep:        baseStep{core: t.core},
		dbName:          t.Req.GetDbName(),
		collectionNames: []string{t.collMeta.Name},
		collectionID:    t.collMeta.CollectionID,
		ts:              t.GetTs(),
	}, &nullStep{})
	undoTask.AddStep(&addPartitionMetaStep{
		baseStep:  baseStep{core: t.core},
		partition: partition,
	}, &nullStep{}) // adding partition is atomic enough.

	return undoTask.Execute(ctx)
}
