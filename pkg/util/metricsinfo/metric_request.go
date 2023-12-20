// Copyright (C) 2019-2020 Zilliz. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under the License.

package metricsinfo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/milvus-io/milvus-proto/go-api/v2/commonpb"
	"github.com/milvus-io/milvus-proto/go-api/v2/milvuspb"
	"github.com/milvus-io/milvus/pkg/common"
	"github.com/milvus-io/milvus/pkg/log"
	"github.com/milvus-io/milvus/pkg/util/commonpbutil"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

const (
	// MetricTypeKey are the key of metric type in GetMetrics request.
	MetricTypeKey = common.MetricTypeKey

	// SystemInfoMetrics means users request for system information metrics.
	SystemInfoMetrics = "system_info"

	// CollectionStorageMetrics means users request for collection storage metrics.
	CollectionStorageMetrics = "collection_storage"

	// MetricRequestTypeKey is a key for identify request type.
	MetricRequestTypeKey = "req_type"

	// QuerySegmentDist request for segment distribution on the query node
	QuerySegmentDist = "query_segment_dist"

	// QueryChannelDist request for channel distribution on the query node
	QueryChannelDist = "query_channel_dist"

	// MetricRequestParamKey is a key as a request parameter
	MetricRequestParamKey = "req_params"

	// MetricRequestParamVerboseKey as a request parameter decide to whether return verbose value
	MetricRequestParamVerboseKey = "req_params.verbose"
)

type MetricsRequestAction func(ctx context.Context, req *milvuspb.GetMetricsRequest, jsonReq gjson.Result) (string, error)

var metricsReqType2Action = make(map[string]MetricsRequestAction)

func RegisterMetricsRequest(reqType string, action MetricsRequestAction) {
	action, ok := metricsReqType2Action[reqType]
	if ok {
		log.Info("metrics request type already exists", zap.String("reqType", reqType))
	}

	metricsReqType2Action[reqType] = action
}

func ExecuteMetricsRequest(ctx context.Context, req *milvuspb.GetMetricsRequest) (string, error) {
	jsonReq := gjson.Parse(req.Request)
	reqType, err := ParseMetricRequestType(jsonReq)
	if err != nil {
		msg := "failed to parse metric type"
		log.Warn(msg, zap.Error(err))
		return "", err
	}

	action, ok := metricsReqType2Action[reqType]
	if !ok {
		log.Warn("unimplemented metric request type", zap.String("req_type", reqType))
		return "", errors.New(MsgUnimplementedMetric)
	}

	actionRet, err := action(ctx, req, jsonReq)
	if err != nil {
		msg := fmt.Sprintf("failed to execute %s", reqType)
		log.Warn(msg, zap.Error(err))
		return "", err
	}
	return actionRet, nil
}

func ParseMetricRequestParams(jsonReq gjson.Result) (gjson.Result, error) {
	v := jsonReq.Get(MetricRequestParamKey)
	if v.Exists() {
		return v, nil
	}
	return gjson.Result{}, fmt.Errorf("%s not found in request", MetricRequestParamKey)
}

func RequestWithVerbose(jsonReq gjson.Result) bool {
	v := jsonReq.Get(MetricRequestParamVerboseKey)
	return v.Exists()
}

// ParseMetricRequestType returns the metric type of req
func ParseMetricRequestType(jsonRet gjson.Result) (string, error) {
	v := jsonRet.Get(MetricRequestTypeKey)
	if v.Exists() {
		return v.String(), nil
	}

	v = jsonRet.Get(MetricTypeKey)
	if v.Exists() {
		return v.String(), nil
	}

	return "", fmt.Errorf("%s or %s not found in request", MetricTypeKey, MetricRequestTypeKey)
}

// ConstructRequestByMetricType constructs a request according to the metric type
func ConstructRequestByMetricType(metricType string) (*milvuspb.GetMetricsRequest, error) {
	m := make(map[string]interface{})
	m[MetricTypeKey] = metricType
	binary, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to construct request by metric type %s: %s", metricType, err.Error())
	}
	// TODO:: switch metricType to different msgType and return err when metricType is not supported
	return &milvuspb.GetMetricsRequest{
		Base: commonpbutil.NewMsgBase(
			commonpbutil.WithMsgType(commonpb.MsgType_SystemInfo),
		),
		Request: string(binary),
	}, nil
}
