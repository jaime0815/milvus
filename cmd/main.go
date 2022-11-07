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

package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/milvus-io/milvus/internal/util/timerecord"

	"golang.org/x/sync/errgroup"

	"github.com/milvus-io/milvus/internal/log"
	"go.uber.org/zap"

	"github.com/milvus-io/milvus-proto/go-api/commonpb"
	"github.com/milvus-io/milvus/internal/mq/msgstream"
	"github.com/milvus-io/milvus/internal/mq/msgstream/mqwrapper"
	"github.com/milvus-io/milvus/internal/proto/internalpb"
	"github.com/milvus-io/milvus/internal/util/commonpbutil"
	"github.com/milvus-io/milvus/internal/util/dependency"
	"github.com/milvus-io/milvus/internal/util/paramtable"
	"github.com/milvus-io/milvus/internal/util/typeutil"
)

func ConsumePerfTest(ctx context.Context, msgStream msgstream.MsgStream, chanName string, msgCount int) {
	subName := fmt.Sprintf("subName-%d", rand.Int())
	msgStream.AsConsumer([]string{chanName}, subName, mqwrapper.SubscriptionPositionEarliest)

	tr := timerecord.NewTimeRecorder("produce")
	ticker := time.Tick(time.Second * 60)
	count := 0
	count2 := 0
	for count2 != msgCount {
		select {
		case <-ctx.Done():
			return
		case <-ticker:
			log.Info("consume time ticker",
				zap.Any("msg count/minutes", count),
				zap.Any("consume count", count2))
			count = 0
		case _, ok := <-msgStream.Chan():
			if ok {
				count++
			}
			count2++
		}
	}
	log.Info("consumer perf finished", zap.Any("time taken", tr.ElapseSpan()))
}

func startSendTT(ctx context.Context, msgStream msgstream.MsgStream, chanName string, msgCount int) {
	msgStream.AsProducer([]string{chanName})
	ts := time.Now().UnixMilli()

	var count count64 = 0
	ticker := time.Tick(time.Second * 60)
	g, gctx := errgroup.WithContext(ctx)

	go func() {
		for {
			select {
			case <-gctx.Done():
				log.Info("startSendTT over")
				return
			case <-ticker:
				log.Info("produce tt count stats", zap.Any("count", count.get()))
			}
		}
	}()

	tr := timerecord.NewTimeRecorder("produce")
	for i := 0; i < 100; i++ {
		g.Go(func() error {
			for i := 0; i < msgCount/100; i++ {
				sendTimeTickToChannel(msgStream, uint64(ts))
				count.inc()
			}
			return nil
		})
	}

	log.Info("waiting for produce finished")
	g.Wait()
	log.Info("produce message finished", zap.Int("msgCount", msgCount), zap.Any("time taken", tr.ElapseSpan()))
}

// SendTimeTickToChannel send each channel's min timetick to msg stream
func sendTimeTickToChannel(msgStream msgstream.MsgStream, ts typeutil.Timestamp) error {
	msgPack := msgstream.MsgPack{}
	baseMsg := msgstream.BaseMsg{
		BeginTimestamp: ts,
		EndTimestamp:   ts,
		HashValues:     []uint32{0},
	}
	timeTickResult := internalpb.TimeTickMsg{
		Base: commonpbutil.NewMsgBase(
			commonpbutil.WithMsgType(commonpb.MsgType_TimeTick),
			commonpbutil.WithMsgID(0),
			commonpbutil.WithTimeStamp(ts),
			commonpbutil.WithSourceID(0),
		),
	}
	timeTickMsg := &msgstream.TimeTickMsg{
		BaseMsg:     baseMsg,
		TimeTickMsg: timeTickResult,
	}
	msgPack.Msgs = append(msgPack.Msgs, timeTickMsg)

	if err := msgStream.Produce(&msgPack); err != nil {
		return err
	}

	return nil
}

type count64 int64

func (c *count64) inc() int64 {
	return atomic.AddInt64((*int64)(c), 1)
}

func (c *count64) reset() {
	atomic.StoreInt64((*int64)(c), 0)
}

func (c *count64) get() int64 {
	return atomic.LoadInt64((*int64)(c))
}

func newStream(ctx context.Context) msgstream.MsgStream {
	paramtable.Init()
	factory := dependency.NewFactory(true)
	Params := paramtable.Get()
	factory.Init(Params)

	msgStream, err := factory.NewMsgStream(ctx)
	if err != nil {
		panic(err)
	}

	return msgStream
}

func runConsumePerfTest(msgCount int) {
	ctx0 := context.Background()
	ctx, cancel := context.WithCancel(ctx0)
	defer cancel()

	msgStream := newStream(ctx)
	defer msgStream.Close()

	log.Info("===== start consume tt perf test...")

	channelName := fmt.Sprintf("tt_chanel-%d", rand.Int())
	log.Info("start produce tt message")
	startSendTT(ctx, msgStream, channelName, msgCount)

	ConsumePerfTest(ctx, msgStream, channelName, msgCount)
}

func runProducePerfTest(msgCount int) {
	log.Info("===== start produce tt perf test...")
	ctx0 := context.Background()
	ctx, cancel := context.WithCancel(ctx0)
	defer cancel()

	msgStream := newStream(ctx)
	defer msgStream.Close()
	channelName := fmt.Sprintf("tt_chanel-%d", rand.Int())

	msgStream.AsProducer([]string{channelName})

	ticker := time.Tick(time.Second * 60)
	var count count64 = 0
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info("runProducePerfTest ticker over")
				return
			case <-ticker:
				log.Info("produce time ticker", zap.Any("msg count/minutes", count))
				count.reset()
			}
		}
	}()

	log.Info("start produce tt message")
	tr := timerecord.NewTimeRecorder("produce")
	ts := time.Now().UnixMilli()
	for i := 0; i < msgCount; i++ {
		sendTimeTickToChannel(msgStream, uint64(ts))
		count.inc()
	}
	log.Info("produce perf finished", zap.Int("msgCount", msgCount), zap.Any("time taken", tr.ElapseSpan()))
}

func main() {
	runConsumePerfTest(500000)
	runProducePerfTest(100000)
}
