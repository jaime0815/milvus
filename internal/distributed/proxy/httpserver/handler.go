package httpserver

import (
	"fmt"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/goccy/go-json"
	"github.com/golang/protobuf/proto"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"

	"github.com/milvus-io/milvus-proto/go-api/v2/commonpb"
	"github.com/milvus-io/milvus-proto/go-api/v2/milvuspb"
	"github.com/milvus-io/milvus-proto/go-api/v2/schemapb"

	"github.com/milvus-io/milvus/internal/common"
	"github.com/milvus-io/milvus/internal/log"
	"github.com/milvus-io/milvus/internal/proxy"
	"github.com/milvus-io/milvus/internal/types"
)

// Handlers handles http requests
type Handlers struct {
	proxy types.ProxyComponent
}

// NewHandlers creates a new Handlers
func NewHandlers(proxy types.ProxyComponent) *Handlers {
	return &Handlers{
		proxy: proxy,
	}
}

// RegisterRoutesTo  registers routes to given router
func (h *Handlers) RegisterRoutesTo(router gin.IRouter) {
	router.GET("/health", wrapHandler(h.handleGetHealth))
	router.POST("/dummy", wrapHandler(h.handleDummy))

	router.POST("/collection", wrapHandler(h.handleCreateCollection))
	router.DELETE("/collection", wrapHandler(h.handleDropCollection))
	router.GET("/collection/existence", wrapHandler(h.handleHasCollection))
	router.GET("/collection", wrapHandler(h.handleDescribeCollection))
	router.POST("/collection/load", wrapHandler(h.handleLoadCollection))
	router.DELETE("/collection/load", wrapHandler(h.handleReleaseCollection))
	router.GET("/collection/statistics", wrapHandler(h.handleGetCollectionStatistics))
	router.GET("/collections", wrapHandler(h.handleShowCollections))

	router.POST("/partition", wrapHandler(h.handleCreatePartition))
	router.DELETE("/partition", wrapHandler(h.handleDropPartition))
	router.GET("/partition/existence", wrapHandler(h.handleHasPartition))
	router.POST("/partitions/load", wrapHandler(h.handleLoadPartitions))
	router.DELETE("/partitions/load", wrapHandler(h.handleReleasePartitions))
	router.GET("/partition/statistics", wrapHandler(h.handleGetPartitionStatistics))
	router.GET("/partitions", wrapHandler(h.handleShowPartitions))

	router.POST("/alias", wrapHandler(h.handleCreateAlias))
	router.DELETE("/alias", wrapHandler(h.handleDropAlias))
	router.PATCH("/alias", wrapHandler(h.handleAlterAlias))

	router.POST("/index", wrapHandler(h.handleCreateIndex))
	router.GET("/index", wrapHandler(h.handleDescribeIndex))
	router.GET("/index/state", wrapHandler(h.handleGetIndexState))
	router.GET("/index/progress", wrapHandler(h.handleGetIndexBuildProgress))
	router.DELETE("/index", wrapHandler(h.handleDropIndex))

	router.POST("/entities", wrapHandler(h.handleInsert))
	router.DELETE("/entities", wrapHandler(h.handleDelete))
	router.POST("/search", wrapHandler(h.handleSearch))
	router.POST("/query", wrapHandler(h.handleQuery))

	router.POST("/persist", wrapHandler(h.handleFlush))
	router.GET("/distance", wrapHandler(h.handleCalcDistance))
	router.GET("/persist/state", wrapHandler(h.handleGetFlushState))
	router.GET("/persist/segment-info", wrapHandler(h.handleGetPersistentSegmentInfo))
	router.GET("/query-segment-info", wrapHandler(h.handleGetQuerySegmentInfo))
	router.GET("/replicas", wrapHandler(h.handleGetReplicas))

	router.GET("/metrics", wrapHandler(h.handleGetMetrics))
	router.POST("/load-balance", wrapHandler(h.handleLoadBalance))
	router.GET("/compaction/state", wrapHandler(h.handleGetCompactionState))
	router.GET("/compaction/plans", wrapHandler(h.handleGetCompactionStateWithPlans))
	router.POST("/compaction", wrapHandler(h.handleManualCompaction))

	router.POST("/import", wrapHandler(h.handleImport))
	router.GET("/import/state", wrapHandler(h.handleGetImportState))
	router.GET("/import/tasks", wrapHandler(h.handleListImportTasks))

	router.POST("/credential", wrapHandler(h.handleCreateCredential))
	router.PATCH("/credential", wrapHandler(h.handleUpdateCredential))
	router.DELETE("/credential", wrapHandler(h.handleDeleteCredential))
	router.GET("/credential/users", wrapHandler(h.handleListCredUsers))

}

func (h *Handlers) handleGetHealth(c *gin.Context) (interface{}, error) {
	req := milvuspb.CheckHealthRequest{}
	// use ShouldBind to supports binding JSON, XML, YAML, and protobuf.
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.CheckHealth(c, &req)
}

func (h *Handlers) handleDummy(c *gin.Context) (interface{}, error) {
	req := milvuspb.DummyRequest{}
	// use ShouldBind to supports binding JSON, XML, YAML, and protobuf.
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.Dummy(c, &req)
}

func (h *Handlers) handleCreateCollection(c *gin.Context) (interface{}, error) {
	wrappedReq := WrappedCreateCollectionRequest{}
	err := shouldBind(c, &wrappedReq)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	schemaProto, err := proto.Marshal(&wrappedReq.Schema)
	if err != nil {
		return nil, fmt.Errorf("%w: marshal schema failed: %v", errBadRequest, err)
	}
	req := &milvuspb.CreateCollectionRequest{
		Base:             wrappedReq.Base,
		DbName:           wrappedReq.DbName,
		CollectionName:   wrappedReq.CollectionName,
		Schema:           schemaProto,
		ShardsNum:        wrappedReq.ShardsNum,
		ConsistencyLevel: wrappedReq.ConsistencyLevel,
		Properties:       wrappedReq.Properties,
	}
	return h.proxy.CreateCollection(c, req)
}

func (h *Handlers) handleDropCollection(c *gin.Context) (interface{}, error) {
	req := milvuspb.DropCollectionRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.DropCollection(c, &req)
}

func (h *Handlers) handleHasCollection(c *gin.Context) (interface{}, error) {
	req := milvuspb.HasCollectionRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.HasCollection(c, &req)
}

func (h *Handlers) handleDescribeCollection(c *gin.Context) (interface{}, error) {
	req := milvuspb.DescribeCollectionRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.DescribeCollection(c, &req)
}

func (h *Handlers) handleLoadCollection(c *gin.Context) (interface{}, error) {
	req := milvuspb.LoadCollectionRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.LoadCollection(c, &req)
}

func (h *Handlers) handleReleaseCollection(c *gin.Context) (interface{}, error) {
	req := milvuspb.ReleaseCollectionRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.ReleaseCollection(c, &req)
}

func (h *Handlers) handleGetCollectionStatistics(c *gin.Context) (interface{}, error) {
	req := milvuspb.GetCollectionStatisticsRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.GetCollectionStatistics(c, &req)
}

func (h *Handlers) handleShowCollections(c *gin.Context) (interface{}, error) {
	req := milvuspb.ShowCollectionsRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.ShowCollections(c, &req)
}

func (h *Handlers) handleCreatePartition(c *gin.Context) (interface{}, error) {
	req := milvuspb.CreatePartitionRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.CreatePartition(c, &req)
}

func (h *Handlers) handleDropPartition(c *gin.Context) (interface{}, error) {
	req := milvuspb.DropPartitionRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.DropPartition(c, &req)
}

func (h *Handlers) handleHasPartition(c *gin.Context) (interface{}, error) {
	req := milvuspb.HasPartitionRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.HasPartition(c, &req)
}

func (h *Handlers) handleLoadPartitions(c *gin.Context) (interface{}, error) {
	req := milvuspb.LoadPartitionsRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.LoadPartitions(c, &req)
}

func (h *Handlers) handleReleasePartitions(c *gin.Context) (interface{}, error) {
	req := milvuspb.ReleasePartitionsRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.ReleasePartitions(c, &req)
}

func (h *Handlers) handleGetPartitionStatistics(c *gin.Context) (interface{}, error) {
	req := milvuspb.GetPartitionStatisticsRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.GetPartitionStatistics(c, &req)
}

func (h *Handlers) handleShowPartitions(c *gin.Context) (interface{}, error) {
	req := milvuspb.ShowPartitionsRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.ShowPartitions(c, &req)
}

func (h *Handlers) handleCreateAlias(c *gin.Context) (interface{}, error) {
	req := milvuspb.CreateAliasRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.CreateAlias(c, &req)
}

func (h *Handlers) handleDropAlias(c *gin.Context) (interface{}, error) {
	req := milvuspb.DropAliasRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.DropAlias(c, &req)
}

func (h *Handlers) handleAlterAlias(c *gin.Context) (interface{}, error) {
	req := milvuspb.AlterAliasRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.AlterAlias(c, &req)
}

func (h *Handlers) handleCreateIndex(c *gin.Context) (interface{}, error) {
	req := milvuspb.CreateIndexRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.CreateIndex(c, &req)
}

func (h *Handlers) handleDescribeIndex(c *gin.Context) (interface{}, error) {
	req := milvuspb.DescribeIndexRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.DescribeIndex(c, &req)
}

func (h *Handlers) handleGetIndexState(c *gin.Context) (interface{}, error) {
	req := milvuspb.GetIndexStateRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.GetIndexState(c, &req)
}

func (h *Handlers) handleGetIndexBuildProgress(c *gin.Context) (interface{}, error) {
	req := milvuspb.GetIndexBuildProgressRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.GetIndexBuildProgress(c, &req)
}

func (h *Handlers) handleDropIndex(c *gin.Context) (interface{}, error) {
	req := milvuspb.DropIndexRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.DropIndex(c, &req)
}

func (h *Handlers) handleInsert(c *gin.Context) (interface{}, error) {
	wrappedReq := WrappedInsertRequest{}
	err := shouldBind(c, &wrappedReq)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	req, err := wrappedReq.AsInsertRequest()
	if err != nil {
		return nil, fmt.Errorf("%w: convert body to pb failed: %v", errBadRequest, err)
	}
	return h.proxy.Insert(c, req)
}

func (h *Handlers) handleDelete(c *gin.Context) (interface{}, error) {
	req := milvuspb.DeleteRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.Delete(c, &req)
}

func (h *Handlers) handleSearch(c *gin.Context) (interface{}, error) {
	wrappedReq := SearchRequest{}
	err := shouldBind(c, &wrappedReq)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	req := milvuspb.SearchRequest{
		Base:               wrappedReq.Base,
		DbName:             wrappedReq.DbName,
		CollectionName:     wrappedReq.CollectionName,
		PartitionNames:     wrappedReq.PartitionNames,
		Dsl:                wrappedReq.Dsl,
		DslType:            wrappedReq.DslType,
		OutputFields:       wrappedReq.OutputFields,
		SearchParams:       wrappedReq.SearchParams,
		TravelTimestamp:    wrappedReq.TravelTimestamp,
		GuaranteeTimestamp: wrappedReq.GuaranteeTimestamp,
		Nq:                 wrappedReq.Nq,
	}
	if len(wrappedReq.BinaryVectors) > 0 {
		req.PlaceholderGroup = binaryVector2Bytes(wrappedReq.BinaryVectors)
	} else {
		req.PlaceholderGroup = vector2Bytes(wrappedReq.Vectors)
	}
	return h.proxy.Search(c, &req)
}

func (h *Handlers) handleQuery(c *gin.Context) (interface{}, error) {
	req := milvuspb.QueryRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.Query(c, &req)
}

func (h *Handlers) handleFlush(c *gin.Context) (interface{}, error) {
	req := milvuspb.FlushRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.Flush(c, &req)
}

func (h *Handlers) handleCalcDistance(c *gin.Context) (interface{}, error) {
	wrappedReq := WrappedCalcDistanceRequest{}
	err := shouldBind(c, &wrappedReq)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}

	req := milvuspb.CalcDistanceRequest{
		Base:    wrappedReq.Base,
		Params:  wrappedReq.Params,
		OpLeft:  wrappedReq.OpLeft.AsPbVectorArray(),
		OpRight: wrappedReq.OpRight.AsPbVectorArray(),
	}
	return h.proxy.CalcDistance(c, &req)
}

func (h *Handlers) handleGetFlushState(c *gin.Context) (interface{}, error) {
	req := milvuspb.GetFlushStateRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.GetFlushState(c, &req)
}

func (h *Handlers) handleGetPersistentSegmentInfo(c *gin.Context) (interface{}, error) {
	req := milvuspb.GetPersistentSegmentInfoRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.GetPersistentSegmentInfo(c, &req)
}

func (h *Handlers) handleGetQuerySegmentInfo(c *gin.Context) (interface{}, error) {
	req := milvuspb.GetQuerySegmentInfoRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.GetQuerySegmentInfo(c, &req)
}

func (h *Handlers) handleGetReplicas(c *gin.Context) (interface{}, error) {
	req := milvuspb.GetReplicasRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.GetReplicas(c, &req)
}

func (h *Handlers) handleGetMetrics(c *gin.Context) (interface{}, error) {
	req := milvuspb.GetMetricsRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.GetMetrics(c, &req)
}

func (h *Handlers) handleLoadBalance(c *gin.Context) (interface{}, error) {
	req := milvuspb.LoadBalanceRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.LoadBalance(c, &req)
}

func (h *Handlers) handleGetCompactionState(c *gin.Context) (interface{}, error) {
	req := milvuspb.GetCompactionStateRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.GetCompactionState(c, &req)
}

func (h *Handlers) handleGetCompactionStateWithPlans(c *gin.Context) (interface{}, error) {
	req := milvuspb.GetCompactionPlansRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.GetCompactionStateWithPlans(c, &req)
}

func (h *Handlers) handleManualCompaction(c *gin.Context) (interface{}, error) {
	req := milvuspb.ManualCompactionRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.ManualCompaction(c, &req)
}

func (h *Handlers) handleImport(c *gin.Context) (interface{}, error) {
	req := milvuspb.ImportRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.Import(c, &req)
}

func (h *Handlers) handleGetImportState(c *gin.Context) (interface{}, error) {
	req := milvuspb.GetImportStateRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.GetImportState(c, &req)
}

func (h *Handlers) handleListImportTasks(c *gin.Context) (interface{}, error) {
	req := milvuspb.ListImportTasksRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.ListImportTasks(c, &req)
}

func (h *Handlers) handleCreateCredential(c *gin.Context) (interface{}, error) {
	req := milvuspb.CreateCredentialRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.CreateCredential(c, &req)
}

func (h *Handlers) handleUpdateCredential(c *gin.Context) (interface{}, error) {
	req := milvuspb.UpdateCredentialRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.UpdateCredential(c, &req)
}

func (h *Handlers) handleDeleteCredential(c *gin.Context) (interface{}, error) {
	req := milvuspb.DeleteCredentialRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.DeleteCredential(c, &req)
}

func (h *Handlers) handleListCredUsers(c *gin.Context) (interface{}, error) {
	req := milvuspb.ListCredUsersRequest{}
	err := shouldBind(c, &req)
	if err != nil {
		return nil, fmt.Errorf("%w: parse body failed: %v", errBadRequest, err)
	}
	return h.proxy.ListCredUsers(c, &req)
}

func (h *Handlers) describeCollection(c *gin.Context, dbName string, collectionName string, needAuth bool) (*milvuspb.DescribeCollectionResponse, error) {
	req := milvuspb.DescribeCollectionRequest{
		Base:           nil,
		DbName:         dbName,
		CollectionName: collectionName,
		CollectionID:   0,
		TimeStamp:      0,
	}
	if needAuth {
		_, authErr := proxy.PrivilegeInterceptor(c, &req)
		if authErr != nil {
			c.JSON(http.StatusOK, gin.H{"code": http.StatusUnauthorized, "message": authErr.Error()})
			return nil, authErr
		}
	}
	response, err := h.proxy.DescribeCollection(c, &req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "describe collection " + collectionName + " fail", "error": err.Error()})
		return nil, err
	} else if response.Status.ErrorCode != commonpb.ErrorCode_Success {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": response.Status.ErrorCode, "message": response.Status.Reason})
		return nil, errors.New(response.Status.Reason)
	}
	primaryField := getPrimaryField(response.Schema)
	if primaryField.AutoID && !response.Schema.AutoID {
		log.Warn("primary filed autoID VS schema autoID", zap.String("collectionName", collectionName), zap.Bool("primary Field", primaryField.AutoID), zap.Bool("schema", response.Schema.AutoID))
		response.Schema.AutoID = EnableAutoID
	}
	return response, nil
}

func (h *Handlers) hasCollection(c *gin.Context, collectionName string) (bool, error) {
	req := milvuspb.HasCollectionRequest{
		Base:           nil,
		DbName:         "",
		CollectionName: collectionName,
		TimeStamp:      0,
	}
	response, err := h.proxy.HasCollection(c, &req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": err})
		return false, err
	}
	return response.Value, nil
}

func (h *Handlers) listCollectionsRequest(router gin.IRouter) {
	router.GET(VectorCollectionsPath, func(c *gin.Context) {
		req := milvuspb.ShowCollectionsRequest{
			Base:      nil,
			DbName:    "",
			TimeStamp: 0,
			Type:      0,
		}
		_, authErr := proxy.PrivilegeInterceptor(c, &req)
		if authErr != nil {
			c.JSON(http.StatusOK, gin.H{"code": http.StatusUnauthorized, "message": authErr.Error()})
			return
		}
		response, err := h.proxy.ShowCollections(c, &req)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "show collections fail", "error": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "data": response.CollectionNames})
		}
	})
}

func (h *Handlers) createCollectionRequest(router gin.IRouter) {
	router.POST(VectorCollectionsCreatePath, func(c *gin.Context) {
		httpReq := CreateCollectionReq{
			CollectionName: "",
			Dimension:      128,
			Description:    "",
			MetricType:     "L2",
			PrimaryField:   "id",
			VectorField:    "vector",
		}
		if err := c.ShouldBindBodyWith(&httpReq, binding.JSON); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "check your parameters conform to the json format", "error": err.Error()})
			return
		}
		has, err := h.hasCollection(c, httpReq.CollectionName)
		if err != nil {
			return
		}
		if has {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "collection " + httpReq.CollectionName + " already exist."})
			return
		}
		schema, err := proto.Marshal(&schemapb.CollectionSchema{
			Name:        httpReq.CollectionName,
			Description: httpReq.Description,
			AutoID:      EnableAutoID,
			Fields: []*schemapb.FieldSchema{
				{
					FieldID:      common.StartOfUserFieldID,
					Name:         httpReq.PrimaryField,
					IsPrimaryKey: true,
					DataType:     schemapb.DataType_Int64,
					AutoID:       EnableAutoID,
				}, {
					FieldID:      common.StartOfUserFieldID + 1,
					Name:         httpReq.VectorField,
					IsPrimaryKey: false,
					DataType:     schemapb.DataType_FloatVector,
					TypeParams: []*commonpb.KeyValuePair{
						{
							Key:   "dim",
							Value: fmt.Sprintf("%d", httpReq.Dimension),
						},
					},
					AutoID: DisableAutoID,
				},
			},
			EnableDynamicField: EnableDynamic,
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "marshal collection schema to string", "error": err.Error()})
			return
		}
		req := milvuspb.CreateCollectionRequest{
			Base:             nil,
			DbName:           "",
			CollectionName:   httpReq.CollectionName,
			Schema:           schema,
			ShardsNum:        ShardNumDefault,
			ConsistencyLevel: commonpb.ConsistencyLevel_Bounded,
			Properties:       nil,
			NumPartitions:    0,
		}
		_, authErr := proxy.PrivilegeInterceptor(c, &req)
		if authErr != nil {
			c.JSON(http.StatusOK, gin.H{"code": http.StatusUnauthorized, "message": authErr.Error()})
			return
		}
		_, err = h.proxy.CreateCollection(c, &req)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "create collection " + httpReq.CollectionName + " fail", "error": err.Error()})
			return
		}
		_, err = h.proxy.CreateIndex(c, &milvuspb.CreateIndexRequest{
			Base:           nil,
			DbName:         "",
			CollectionName: httpReq.CollectionName,
			FieldName:      httpReq.VectorField,
			ExtraParams:    []*commonpb.KeyValuePair{{Key: MetricType, Value: httpReq.MetricType}},
			IndexName:      "",
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "create index for collection " + httpReq.CollectionName + " fail, after the collection was created", "error": err.Error()})
			return
		}
		_, err = h.proxy.LoadCollection(c, &milvuspb.LoadCollectionRequest{
			Base:           nil,
			DbName:         "",
			CollectionName: httpReq.CollectionName,
			ReplicaNumber:  0,
			ResourceGroups: nil,
			Refresh:        false,
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "load collection " + httpReq.CollectionName + " fail, after the index was created", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "data": gin.H{}})
	})
}

func (h *Handlers) describeCollectionRequest(router gin.IRouter) {
	router.GET(VectorCollectionsDescribePath, func(c *gin.Context) {
		dbName := c.Query("collectionName")
		collectionName := c.Query("collectionName")
		coll, err := h.describeCollection(c, dbName, collectionName, true)
		if err != nil {
			return
		}
		stateResp, stateErr := h.proxy.GetLoadState(c, &milvuspb.GetLoadStateRequest{
			DbName:         dbName,
			CollectionName: collectionName,
		})

		loadState := ""
		if stateErr != nil {
			log.Warn("get collection load state fail", zap.String("collection", collectionName), zap.String("err", stateErr.Error()))
		} else if stateResp.Status.ErrorCode != commonpb.ErrorCode_Success {
			log.Warn("get collection load state fail", zap.String("collection", collectionName), zap.String("err", stateResp.Status.Reason))
		} else {
			loadState = stateResp.GetState().String()
		}

		vectorField := ""
		for _, field := range coll.Schema.Fields {
			if field.DataType == schemapb.DataType_BinaryVector || field.DataType == schemapb.DataType_FloatVector {
				vectorField = field.Name
				break
			}
		}

		indexResp, indexErr := h.proxy.DescribeIndex(c, &milvuspb.DescribeIndexRequest{
			Base:           nil,
			DbName:         dbName,
			CollectionName: collectionName,
			FieldName:      vectorField,
			IndexName:      "",
		})

		indexDescriptions := make([]*milvuspb.IndexDescription, 0)
		if indexErr != nil {
			log.Warn("get indexes description fail", zap.String("collection", collectionName), zap.String("vectorField", vectorField), zap.String("err", indexErr.Error()))
		} else if indexResp.Status.ErrorCode != commonpb.ErrorCode_Success {
			log.Warn("get indexes description fail", zap.String("collection", collectionName), zap.String("vectorField", vectorField), zap.String("err", indexResp.Status.Reason))
		} else {
			indexDescriptions = indexResp.IndexDescriptions
		}

		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "data": gin.H{
			"collectionName": coll.CollectionName,
			"description":    coll.Schema.Description,
			"fields":         printFields(coll.Schema.Fields),
			"indexes":        printIndexes(indexDescriptions),
			"load":           loadState,
			"shardsNum":      coll.ShardsNum,
			"enableDynamic":  coll.Schema.EnableDynamicField,
		}})
	})
}

func (h *Handlers) dropCollectionRequest(router gin.IRouter) {
	router.POST(VectorCollectionsDropPath, func(c *gin.Context) {
		httpReq := DropCollectionReq{
			CollectionName: "",
		}
		if err := c.ShouldBindBodyWith(&httpReq, binding.JSON); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "check your parameters conform to the json format", "error": err.Error()})
			return
		}
		has, err := h.hasCollection(c, httpReq.CollectionName)
		if err != nil {
			return
		}
		if !has {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "can't find collection:  " + httpReq.CollectionName})
			return
		}
		req := milvuspb.DropCollectionRequest{
			Base:           nil,
			DbName:         "",
			CollectionName: httpReq.CollectionName,
		}
		_, authErr := proxy.PrivilegeInterceptor(c, &req)
		if authErr != nil {
			c.JSON(http.StatusOK, gin.H{"code": http.StatusUnauthorized, "message": authErr.Error()})
			return
		}
		response, err := h.proxy.DropCollection(c, &req)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "drop collection " + httpReq.CollectionName + " fail", "error": err.Error()})
		} else if response.ErrorCode != commonpb.ErrorCode_Success {
			c.JSON(http.StatusOK, gin.H{"code": response.ErrorCode, "message": response.Reason})
		} else {
			c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "data": gin.H{}})
		}
	})
}

func (h *Handlers) queryRequest(router gin.IRouter) {
	router.POST(VectorQueryPath, func(c *gin.Context) {
		httpReq := QueryReq{
			CollectionName: "",
			OutputFields:   nil,
			Filter:         "",
			Limit:          100,
			Offset:         0,
		}
		if err := c.ShouldBindBodyWith(&httpReq, binding.JSON); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "check your parameters conform to the json format", "error": err.Error()})
			return
		}
		coll, err := h.describeCollection(c, "", httpReq.CollectionName, false)
		if err != nil || coll == nil {
			return
		}
		dynamicFields := []string{}
		httpReq.OutputFields, dynamicFields = getDynamicOutputFields(coll.Schema, httpReq.OutputFields)
		req := milvuspb.QueryRequest{
			Base:                  nil,
			DbName:                "",
			CollectionName:        httpReq.CollectionName,
			Expr:                  httpReq.Filter,
			OutputFields:          httpReq.OutputFields,
			PartitionNames:        nil,
			TravelTimestamp:       0,
			GuaranteeTimestamp:    2, // BoundedTimestamp
			QueryParams:           nil,
			NotReturnAllMeta:      false,
			UseDefaultConsistency: true,
		}
		_, authErr := proxy.PrivilegeInterceptor(c, &req)
		if authErr != nil {
			c.JSON(http.StatusOK, gin.H{"code": http.StatusUnauthorized, "message": authErr.Error()})
			return
		}
		response, err := h.proxy.Query(c, &req)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "query fail", "error": err.Error()})
		} else if response.Status.ErrorCode != commonpb.ErrorCode_Success {
			c.JSON(http.StatusOK, gin.H{"code": response.Status.ErrorCode, "message": response.Status.Reason})
		} else {
			outputData, err := buildQueryResp(int64(0), dynamicFields, response.FieldsData, nil)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "show result by row wrong", "originData": response.FieldsData, "error": err.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "data": outputData})
			}
		}
	})
}

func (h *Handlers) getRequest(router gin.IRouter) {
	router.POST(VectorGetPath, func(c *gin.Context) {
		httpReq := GetReq{
			CollectionName: "",
			OutputFields:   nil,
			Id:             nil,
		}
		if err := c.ShouldBindBodyWith(&httpReq, binding.JSON); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "check your parameters conform to the json format", "error": err.Error()})
			return
		}
		coll, err := h.describeCollection(c, "", httpReq.CollectionName, false)
		if err != nil || coll == nil {
			return
		}
		dynamicFields := []string{}
		httpReq.OutputFields, dynamicFields = getDynamicOutputFields(coll.Schema, httpReq.OutputFields)
		body, _ := c.Get(gin.BodyBytesKey)
		filter := checkGetPrimaryKey(coll.Schema, gjson.Get(string(body.([]byte)), "id"))
		req := milvuspb.QueryRequest{
			Base:                  nil,
			DbName:                "",
			CollectionName:        httpReq.CollectionName,
			Expr:                  filter,
			OutputFields:          httpReq.OutputFields,
			PartitionNames:        []string{},
			GuaranteeTimestamp:    2, // BoundedTimestamp
			QueryParams:           nil,
			NotReturnAllMeta:      false,
			UseDefaultConsistency: true,
		}
		_, authErr := proxy.PrivilegeInterceptor(c, &req)
		if authErr != nil {
			c.JSON(http.StatusOK, gin.H{"code": http.StatusUnauthorized, "message": authErr.Error()})
			return
		}
		response, err := h.proxy.Query(c, &req)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "query fail", "error": err.Error()})
		} else if response.Status.ErrorCode != commonpb.ErrorCode_Success {
			c.JSON(http.StatusOK, gin.H{"code": response.Status.ErrorCode, "message": response.Status.Reason})
		} else {
			outputData, err := buildQueryResp(int64(0), dynamicFields, response.FieldsData, nil)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "show result by row wrong", "originData": response.FieldsData, "error": err.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "data": outputData})
				log.Error("get resultIS: ", zap.Any("res", outputData))
			}
		}
	})
}

func (h *Handlers) deleteRequest(router gin.IRouter) {
	router.POST(VectorDeletePath, func(c *gin.Context) {
		httpReq := DeleteReq{
			CollectionName: "",
			Id:             nil,
		}
		if err := c.ShouldBindBodyWith(&httpReq, binding.JSON); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "check your parameters conform to the json format", "error": err.Error()})
			return
		}
		coll, err := h.describeCollection(c, "", httpReq.CollectionName, false)
		if err != nil || coll == nil {
			return
		}
		body, _ := c.Get(gin.BodyBytesKey)
		expr := checkGetPrimaryKey(coll.Schema, gjson.Get(string(body.([]byte)), "id"))
		req := milvuspb.DeleteRequest{
			Base:           nil,
			DbName:         "",
			CollectionName: httpReq.CollectionName,
			Expr:           expr,
		}
		_, authErr := proxy.PrivilegeInterceptor(c, &req)
		if authErr != nil {
			c.JSON(http.StatusOK, gin.H{"code": http.StatusUnauthorized, "message": authErr.Error()})
			return
		}
		response, err := h.proxy.Delete(c, &req)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "delete fail", "error": err.Error()})
		} else if response.Status.ErrorCode != commonpb.ErrorCode_Success {
			c.JSON(http.StatusOK, gin.H{"code": response.Status.ErrorCode, "message": response.Status.Reason})
		} else {
			c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "data": gin.H{}})
		}
	})
}

func (h *Handlers) insertRequest(router gin.IRouter) {
	router.POST(VectorInsertPath, func(c *gin.Context) {
		httpReq := InsertReq{
			CollectionName: "",
			Data:           nil,
		}
		if err := c.ShouldBindBodyWith(&httpReq, binding.JSON); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "check your parameters conform to the json format", "error": err.Error()})
			return
		}
		coll, err := h.describeCollection(c, "", httpReq.CollectionName, false)
		if err != nil || coll == nil {
			return
		}
		body, _ := c.Get(gin.BodyBytesKey)
		err = checkAndSetData(string(body.([]byte)), coll, &httpReq)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "checkout your params", "originRequest": httpReq, "error": err.Error()})
			return
		}
		req := milvuspb.InsertRequest{
			DbName:         "", // reserved
			CollectionName: httpReq.CollectionName,
			PartitionName:  "_default",
			NumRows:        uint32(len(httpReq.Data)),
		}
		req.FieldsData, err = anyToColumns(httpReq.Data, coll.Schema)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "insert data by column wrong", "originData": httpReq.Data, "collectionSchema": coll.Schema, "error": err.Error()})
			return
		}
		_, authErr := proxy.PrivilegeInterceptor(c, &req)
		if authErr != nil {
			c.JSON(http.StatusOK, gin.H{"code": http.StatusUnauthorized, "message": authErr.Error()})
			return
		}
		response, err := h.proxy.Insert(c, &req)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "insert fail", "error": err.Error()})
		} else if response.Status.ErrorCode != commonpb.ErrorCode_Success {
			c.JSON(http.StatusOK, gin.H{"code": response.Status.ErrorCode, "message": response.Status.Reason, "data": req.FieldsData})
		} else {
			switch response.IDs.GetIdField().(type) {
			case *schemapb.IDs_IntId:
				c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "data": gin.H{"insertCount": response.InsertCnt, "insertIds": response.IDs.IdField.(*schemapb.IDs_IntId).IntId.Data}})
			case *schemapb.IDs_StrId:
				c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "data": gin.H{"insertCount": response.InsertCnt, "insertIds": response.IDs.IdField.(*schemapb.IDs_StrId).StrId.Data}})
			}
		}
	})
}

func (h *Handlers) searchRequest(router gin.IRouter) {
	router.POST(VectorSearchPath, func(c *gin.Context) {
		httpReq := SearchReq{
			CollectionName: "",
			Filter:         "",
			Limit:          100,
			Offset:         0,
			OutputFields:   nil,
			Vector:         nil,
		}
		if err := c.ShouldBindBodyWith(&httpReq, binding.JSON); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "check your parameters conform to the json format", "error": err.Error()})
			return
		}
		coll, err := h.describeCollection(c, "", httpReq.CollectionName, false)
		if err != nil {
			return
		}
		dynamicFields := []string{}
		httpReq.OutputFields, dynamicFields = getDynamicOutputFields(coll.Schema, httpReq.OutputFields)
		vectorField := ""
		for _, field := range coll.Schema.Fields {
			if field.DataType == schemapb.DataType_BinaryVector || field.DataType == schemapb.DataType_FloatVector {
				vectorField = field.Name
				break
			}
		}
		indexResp, indexErr := h.proxy.DescribeIndex(c, &milvuspb.DescribeIndexRequest{
			Base:           nil,
			DbName:         "",
			CollectionName: httpReq.CollectionName,
			FieldName:      vectorField,
			IndexName:      "",
		})
		if indexErr != nil {
			log.Warn("get indexes description fail", zap.String("collection", httpReq.CollectionName), zap.String("vectorField", vectorField), zap.String("err", indexErr.Error()))
		} else if indexResp.Status.ErrorCode != commonpb.ErrorCode_Success {
			log.Warn("get indexes description fail", zap.String("collection", httpReq.CollectionName), zap.String("vectorField", vectorField), zap.String("err", indexResp.Status.Reason))
		}

		metricType := getMetricType(indexResp.IndexDescriptions[0].Params)
		params := map[string]interface{}{ //auto generated mapping
			"level": int(commonpb.ConsistencyLevel_Bounded),
		}
		bs, err := json.Marshal(params)
		if err != nil {
			log.Warn("get indexes description fail", zap.String("collection", httpReq.CollectionName))
		}

		searchParams := []*commonpb.KeyValuePair{
			{Key: "anns_field", Value: vectorField},
			{Key: "topk", Value: fmt.Sprintf("%d", httpReq.Limit)},
			{Key: "params", Value: string(bs)},
			{Key: "metric_type", Value: metricType},
			{Key: "round_decimal", Value: "-1"},
			{Key: "offset", Value: fmt.Sprintf("%d", httpReq.Offset)},
		}
		req := milvuspb.SearchRequest{
			DbName:             "",
			CollectionName:     httpReq.CollectionName,
			PartitionNames:     []string{},
			Dsl:                httpReq.Filter,
			PlaceholderGroup:   vector2PlaceholderGroupBytes(httpReq.Vector),
			DslType:            commonpb.DslType_BoolExprV1,
			OutputFields:       httpReq.OutputFields,
			SearchParams:       searchParams,
			GuaranteeTimestamp: 2, // BoundedTimestamp
			Nq:                 int64(1),
		}
		_, authErr := proxy.PrivilegeInterceptor(c, &req)
		if authErr != nil {
			c.JSON(http.StatusOK, gin.H{"code": http.StatusUnauthorized, "message": authErr.Error()})
			return
		}
		response, err := h.proxy.Search(c, &req)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "search fail", "error": err.Error()})
		} else if response.Status.ErrorCode != commonpb.ErrorCode_Success {
			c.JSON(http.StatusOK, gin.H{"code": response.Status.ErrorCode, "message": response.Status.Reason})
		} else {
			outputData, err := buildQueryResp(response.Results.TopK, dynamicFields, response.Results.FieldsData, response.Results.Scores)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "message": "show result by row wrong", "originResquest": req, "originResult": response.Results, "error": err.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "data": outputData})
			}
		}
	})
}

func (h *Handlers) RegisterRoutesToV1(router gin.IRouter) {
	h.listCollectionsRequest(router)
	h.createCollectionRequest(router)
	h.describeCollectionRequest(router)
	h.dropCollectionRequest(router)
	h.queryRequest(router)
	h.deleteRequest(router)
	h.insertRequest(router)
	h.searchRequest(router)
}
