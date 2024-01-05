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
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/milvus-io/milvus/pkg/log"
	"go.uber.org/zap"

	"github.com/milvus-io/milvus/internal/proxy/connection"
	"github.com/milvus-io/milvus/internal/util/mhttp"
	"github.com/milvus-io/milvus/pkg/util/merr"
	"github.com/milvus-io/milvus/pkg/util/metricsinfo"
)

func getConfigs(configs map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		bs, err := json.Marshal(configs)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				mhttp.HTTPReturnCode:    http.StatusInternalServerError,
				mhttp.HTTPReturnMessage: err.Error(),
			})
			return
		}
		c.IndentedJSON(http.StatusOK, string(bs))
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

		resp, err := node.metricsCacheManager.GetSystemInfoMetrics()

		// fetch metrics from remote and update local cache if getting metrics failed from local cache
		if err != nil {
			var err1 error
			resp, err1 = getSystemInfoMetrics(c, req, node)
			if err1 != nil {
				c.JSON(http.StatusOK, gin.H{
					mhttp.HTTPReturnCode:    http.StatusInternalServerError,
					mhttp.HTTPReturnMessage: err1.Error(),
				})
				return
			}
			node.metricsCacheManager.UpdateSystemInfoMetrics(resp)
		}

		if !merr.Ok(resp.GetStatus()) {
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
	clients := connection.GetManager().List()
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

func getQueryComponentMetrics(node *Proxy, condition string) gin.HandlerFunc {
	return func(c *gin.Context) {
		req, err := metricsinfo.ConstructRequestByMetricType(condition)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				mhttp.HTTPReturnCode:    http.StatusInternalServerError,
				mhttp.HTTPReturnMessage: err.Error(),
			})
			return
		}

		resp, err := node.queryCoord.GetMetrics(c, req)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				mhttp.HTTPReturnCode:    http.StatusInternalServerError,
				mhttp.HTTPReturnMessage: err.Error(),
			})
			return
		}

		if !merr.Ok(resp.GetStatus()) {
			c.JSON(http.StatusOK, gin.H{
				mhttp.HTTPReturnCode:    resp.GetStatus().GetCode(),
				mhttp.HTTPReturnMessage: resp.Status.Reason,
			})
			return
		}

		ret, err := strconv.Unquote(resp.Response)
		if err != nil {
			log.Error("====getQueryComponentMetrics=======", zap.Any("resp", resp.Response), zap.Any("ret", ret), zap.Error(err))
			//c.JSON(http.StatusOK, gin.H{
			//	mhttp.HTTPReturnCode:    http.StatusInternalServerError,
			//	mhttp.HTTPReturnMessage: err.Error(),
			//})
			//return
		}

		c.IndentedJSON(http.StatusOK, resp.Response)
	}
}
