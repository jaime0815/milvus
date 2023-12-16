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
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/milvus-io/milvus/pkg/util/merr"

	"github.com/milvus-io/milvus-proto/go-api/v2/milvuspb"
	"github.com/milvus-io/milvus/internal/types"
	"github.com/milvus-io/milvus/internal/util/mhttp"
	"github.com/milvus-io/milvus/pkg/util/funcutil"
	"github.com/milvus-io/milvus/pkg/util/metricsinfo"
)

func getConfigs(configs map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		r, err := funcutil.MapToJSON(configs)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				mhttp.HTTPReturnCode:    http.StatusInternalServerError,
				mhttp.HTTPReturnMessage: err.Error(),
			})
			return
		}
		c.IndentedJSON(http.StatusOK, r)
	}
}

func getClusterInfo(node *Proxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		req, err := metricsinfo.ConstructRequestByMetricType(metricsinfo.SystemInfoMetrics)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				mhttp.HTTPReturnCode:    http.StatusInternalServerError,
				mhttp.HTTPReturnMessage: err.Error(),
			})
			return
		}

		resp, err := getSystemInfoMetrics(c, req, node)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				mhttp.HTTPReturnCode:    http.StatusInternalServerError,
				mhttp.HTTPReturnMessage: err.Error(),
			})
			return
		}

		if merr.Ok(resp.GetStatus()) {
			c.JSON(http.StatusOK, gin.H{
				mhttp.HTTPReturnCode:    resp.GetStatus().GetCode(),
				mhttp.HTTPReturnMessage: resp.Status.Reason,
			})
			return
		}

		c.IndentedJSON(http.StatusOK, resp.GetResponse())
	}

}

func getConnectedClients(c *gin.Context) {
	clients := GetConnectionManager().list()

	ret, err := json.Marshal(clients)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			mhttp.HTTPReturnCode:    http.StatusInternalServerError,
			mhttp.HTTPReturnMessage: err.Error(),
		})
		return
	}

	c.IndentedJSON(http.StatusOK, string(ret))
}

func getQueryNodeStats(qc types.QueryCoordClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := &milvuspb.GetMetricsRequest{}
		resp, err := qc.GetMetrics(c, req)

		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				mhttp.HTTPReturnCode:    http.StatusInternalServerError,
				mhttp.HTTPReturnMessage: err.Error(),
			})
			return
		}

		if merr.Ok(resp.GetStatus()) {
			c.JSON(http.StatusOK, gin.H{
				mhttp.HTTPReturnCode:    resp.GetStatus().GetCode(),
				mhttp.HTTPReturnMessage: resp.Status.Reason,
			})
			return
		}

		c.IndentedJSON(http.StatusOK, resp.Response)
	}
}
