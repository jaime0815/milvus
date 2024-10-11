package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/milvus-io/milvus-proto/go-api/v2/milvuspb"
	"github.com/milvus-io/milvus/pkg/util/merr"
	"github.com/milvus-io/milvus/pkg/util/metricsinfo"
)

func TestGetConfigs(t *testing.T) {
	configs := map[string]string{
		"config1": "value1",
		"config2": "value2",
	}

	router := gin.Default()
	router.GET("/configs", getConfigs(configs))

	req, _ := http.NewRequest("GET", "/configs", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "value1")
	assert.Contains(t, resp.Body.String(), "value2")
}

func TestGetClusterInfo(t *testing.T) {
	// Mock the Proxy node
	// mock the GetSystemInfoMetrics request and response
	node := &Proxy{metricsCacheManager: metricsinfo.NewMetricsCacheManager()}
	node.metricsCacheManager.UpdateSystemInfoMetrics(&milvuspb.GetMetricsResponse{
		Status:        merr.Success(),
		Response:      "cached response",
		ComponentName: "cached component",
	})
	router := gin.Default()
	router.GET("/clusterinfo", getClusterInfo(node))

	req, _ := http.NewRequest("GET", "/clusterinfo", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Assert the response code and body
	// This will depend on the mock implementation of the Proxy node
}

func TestGetConnectedClients(t *testing.T) {
	router := gin.Default()
	router.GET("/clients", getConnectedClients)

	req, _ := http.NewRequest("GET", "/clients", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Assert the response code and body
	// This will depend on the implementation of the GetManager().List() function
}

func TestGetDependencies(t *testing.T) {
	router := gin.Default()
	router.GET("/dependencies", getDependencies)

	req, _ := http.NewRequest("GET", "/dependencies", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Assert the response code and body
	// This will depend on the implementation of the getDependencies function
}

func TestBuildReqParams(t *testing.T) {
	router := gin.Default()
	router.GET("/params", func(c *gin.Context) {
		params := buildReqParams(c, "testType")
		c.JSON(http.StatusOK, params)
	})

	req, _ := http.NewRequest("GET", "/params?param1=value1&param2=value2", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "value1")
	assert.Contains(t, resp.Body.String(), "value2")
}

func TestGetQueryComponentMetrics(t *testing.T) {
	// Mock the Proxy node
	node := &Proxy{}

	router := gin.Default()
	router.GET("/metrics", getQueryComponentMetrics(node, "testType"))

	req, _ := http.NewRequest("GET", "/metrics", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Assert the response code and body
	// This will depend on the mock implementation of the Proxy node
}

// generate unit test for all method in http_req_impl.go file

// generate unit test for getDependencies in http_req_impl.go file
