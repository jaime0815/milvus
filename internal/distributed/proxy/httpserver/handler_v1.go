package httpserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang/protobuf/proto"
	"github.com/milvus-io/milvus-proto/go-api/v2/commonpb"
	"github.com/milvus-io/milvus-proto/go-api/v2/milvuspb"
	"github.com/milvus-io/milvus-proto/go-api/v2/schemapb"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"net/http"

	"github.com/milvus-io/milvus/internal/common"
	"github.com/milvus-io/milvus/internal/log"
	"github.com/milvus-io/milvus/internal/proxy"
)

func (h *Handlers) describeCollection(c *gin.Context, collectionName string, needAuth bool) (*milvuspb.DescribeCollectionResponse, error) {
	req := milvuspb.DescribeCollectionRequest{
		Base:           nil,
		DbName:         "",
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
		c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 400, "message": "describe collection " + collectionName + " fail", "error": err.Error()})
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
		c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 400, "message": err})
		return false, err
	}
	return response.Value, nil
}

func printFields(fields []*schemapb.FieldSchema) []gin.H {
	res := []gin.H{}
	for _, field := range fields {
		res = append(res, gin.H{
			"name":        field.Name,
			"type":        field.DataType.String(),
			"primaryKey":  field.IsPrimaryKey,
			"autoId":      field.AutoID,
			"description": field.Description,
		})
	}
	return res
}

func getMetricType(pairs []*commonpb.KeyValuePair) string {
	metricType := "L2"
	for _, pair := range pairs {
		if pair.Key == "metric_type" {
			metricType = pair.Value
			break
		}
	}
	return metricType
}

func printIndexes(fields []*milvuspb.IndexDescription) []gin.H {
	res := []gin.H{}
	for _, field := range fields {
		res = append(res, gin.H{
			"indexName":  field.IndexName,
			"fieldName":  field.FieldName,
			"metricType": getMetricType(field.Params),
		})
	}
	return res
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
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "show collections fail", "error": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"code": 200, "data": response.CollectionNames})
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
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 400, "message": "check your parameters conform to the json format", "error": err.Error()})
			return
		}
		has, err := h.hasCollection(c, httpReq.CollectionName)
		if err != nil {
			return
		}
		if has {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 200, "message": "collection " + httpReq.CollectionName + " already exist."})
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
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 400, "message": "marshal collection schema to string", "error": err.Error()})
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
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 400, "message": "create collection " + httpReq.CollectionName + " fail", "error": err.Error()})
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
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 400, "message": "create index for collection " + httpReq.CollectionName + " fail, after the collection was created", "error": err.Error()})
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
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 400, "message": "load collection " + httpReq.CollectionName + " fail, after the index was created", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{}})
	})
}

func (h *Handlers) describeCollectionRequest(router gin.IRouter) {
	router.GET(VectorCollectionsDescribePath, func(c *gin.Context) {
		collectionName := c.Query("collectionName")
		coll, err := h.describeCollection(c, collectionName, true)
		if err != nil {
			return
		}
		stateResp, stateErr := h.proxy.GetLoadState(c, &milvuspb.GetLoadStateRequest{
			DbName:         "",
			CollectionName: collectionName,
		})
		if stateErr != nil {
			log.Warn("get collection load state fail", zap.String("collection", collectionName), zap.String("err", stateErr.Error()))
		} else if stateResp.Status.ErrorCode != commonpb.ErrorCode_Success {
			log.Warn("get collection load state fail", zap.String("collection", collectionName), zap.String("err", stateResp.Status.Reason))
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
			DbName:         "",
			CollectionName: collectionName,
			FieldName:      vectorField,
			IndexName:      "",
		})
		if indexErr != nil {
			log.Warn("get indexes description fail", zap.String("collection", collectionName), zap.String("vectorField", vectorField), zap.String("err", indexErr.Error()))
		} else if indexResp.Status.ErrorCode != commonpb.ErrorCode_Success {
			log.Warn("get indexes description fail", zap.String("collection", collectionName), zap.String("vectorField", vectorField), zap.String("err", indexResp.Status.Reason))
		}
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{
			"collectionName": coll.CollectionName,
			"description":    coll.Schema.Description,
			"fields":         printFields(coll.Schema.Fields),
			"indexes":        printIndexes(indexResp.IndexDescriptions),
			"load":           stateResp.State.String(),
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
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 400, "message": "check your parameters conform to the json format", "error": err.Error()})
			return
		}
		has, err := h.hasCollection(c, httpReq.CollectionName)
		if err != nil {
			return
		}
		if !has {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 200, "message": "can't find collection:  " + httpReq.CollectionName})
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
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "drop collection " + httpReq.CollectionName + " fail", "error": err.Error()})
		} else if response.ErrorCode != commonpb.ErrorCode_Success {
			c.JSON(http.StatusOK, gin.H{"code": response.ErrorCode, "message": response.Reason})
		} else {
			c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{}})
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
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 400, "message": "check your parameters conform to the json format", "error": err.Error()})
			return
		}
		coll, err := h.describeCollection(c, httpReq.CollectionName, false)
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
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "query fail", "error": err.Error()})
		} else if response.Status.ErrorCode != commonpb.ErrorCode_Success {
			c.JSON(http.StatusOK, gin.H{"code": response.Status.ErrorCode, "message": response.Status.Reason})
		} else {
			outputData, err := buildQueryResp(int64(0), dynamicFields, response.FieldsData, nil)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"code": 400, "message": "show result by row wrong", "originData": response.FieldsData, "error": err.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{"code": 200, "data": outputData})
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
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 400, "message": "check your parameters conform to the json format", "error": err.Error()})
			return
		}
		coll, err := h.describeCollection(c, httpReq.CollectionName, false)
		if err != nil || coll == nil {
			return
		}
		dynamicFields := []string{}
		httpReq.OutputFields, dynamicFields = getDynamicOutputFields(coll.Schema, httpReq.OutputFields)
		body, _ := c.Get(gin.BodyBytesKey)
		filter := checkGetPrimaryKey(coll, gjson.Get(string(body.([]byte)), "id"))
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
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "query fail", "error": err.Error()})
		} else if response.Status.ErrorCode != commonpb.ErrorCode_Success {
			c.JSON(http.StatusOK, gin.H{"code": response.Status.ErrorCode, "message": response.Status.Reason})
		} else {
			outputData, err := buildQueryResp(int64(0), dynamicFields, response.FieldsData, nil)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"code": 400, "message": "show result by row wrong", "originData": response.FieldsData, "error": err.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{"code": 200, "data": outputData})
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
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 400, "message": "check your parameters conform to the json format", "error": err.Error()})
			return
		}
		coll, err := h.describeCollection(c, httpReq.CollectionName, false)
		if err != nil || coll == nil {
			return
		}
		body, _ := c.Get(gin.BodyBytesKey)
		expr := checkGetPrimaryKey(coll, gjson.Get(string(body.([]byte)), "id"))
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
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "delete fail", "error": err.Error()})
		} else if response.Status.ErrorCode != commonpb.ErrorCode_Success {
			c.JSON(http.StatusOK, gin.H{"code": response.Status.ErrorCode, "message": response.Status.Reason})
		} else {
			c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{}})
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
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 400, "message": "check your parameters conform to the json format", "error": err.Error()})
			return
		}
		coll, err := h.describeCollection(c, httpReq.CollectionName, false)
		if err != nil || coll == nil {
			return
		}
		body, _ := c.Get(gin.BodyBytesKey)
		err = checkAndSetData(string(body.([]byte)), coll, &httpReq)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 400, "message": "checkout your params", "originRequest": httpReq, "error": err.Error()})
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
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 400, "message": "insert data by column wrong", "originData": httpReq.Data, "collectionSchema": coll.Schema, "error": err.Error()})
			return
		}
		_, authErr := proxy.PrivilegeInterceptor(c, &req)
		if authErr != nil {
			c.JSON(http.StatusOK, gin.H{"code": http.StatusUnauthorized, "message": authErr.Error()})
			return
		}
		response, err := h.proxy.Insert(c, &req)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "insert fail", "error": err.Error()})
		} else if response.Status.ErrorCode != commonpb.ErrorCode_Success {
			c.JSON(http.StatusOK, gin.H{"code": response.Status.ErrorCode, "message": response.Status.Reason, "data": req.FieldsData})
		} else {
			switch response.IDs.GetIdField().(type) {
			case *schemapb.IDs_IntId:
				c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"insertCount": response.InsertCnt, "insertIds": response.IDs.IdField.(*schemapb.IDs_IntId).IntId.Data}})
			case *schemapb.IDs_StrId:
				c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"insertCount": response.InsertCnt, "insertIds": response.IDs.IdField.(*schemapb.IDs_StrId).StrId.Data}})
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
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 400, "message": "check your parameters conform to the json format", "error": err.Error()})
			return
		}
		coll, err := h.describeCollection(c, httpReq.CollectionName, false)
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
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "search fail", "error": err.Error()})
		} else if response.Status.ErrorCode != commonpb.ErrorCode_Success {
			c.JSON(http.StatusOK, gin.H{"code": response.Status.ErrorCode, "message": response.Status.Reason})
		} else {
			outputData, err := buildQueryResp(response.Results.TopK, dynamicFields, response.Results.FieldsData, response.Results.Scores)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"code": 400, "message": "show result by row wrong", "originResquest": req, "originResult": response.Results, "error": err.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{"code": 200, "data": outputData})
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

type CreateCollectionReq struct {
	CollectionName string `json:"collectionName" validate:"required"`
	Dimension      int32  `json:"dimension" validate:"required"`
	Description    string `json:"description"`
	MetricType     string `json:"metricType"`
	PrimaryField   string `json:"primaryField"`
	VectorField    string `json:"vectorField"`
}

type DropCollectionReq struct {
	CollectionName string `json:"collectionName" validate:"required"`
}

type QueryReq struct {
	CollectionName string   `json:"collectionName" validate:"required"`
	OutputFields   []string `json:"outputFields"`
	Filter         string   `json:"filter" validate:"required"`
	Limit          int32    `json:"limit"`
	Offset         int32    `json:"offset"`
}

type GetReq struct {
	CollectionName string   `json:"collectionName" validate:"required"`
	OutputFields   []string `json:"outputFields"`
	// TODO 先统一按照 id处理，后续再调整
	Id interface{} `json:"id" validate:"required"`
}

type DeleteReq struct {
	CollectionName string `json:"collectionName" validate:"required"`
	// TODO 先统一按照 id处理，后续再调整
	Id interface{} `json:"id" validate:"required"`
}

type InsertReq struct {
	CollectionName string                   `json:"collectionName" validate:"required"`
	Data           []map[string]interface{} `json:"data" validate:"required"`
}

type SearchReq struct {
	CollectionName string    `json:"collectionName" validate:"required"`
	Filter         string    `json:"filter"`
	Limit          int32     `json:"limit"`
	Offset         int32     `json:"offset"`
	OutputFields   []string  `json:"outputFields"`
	Vector         []float32 `json:"vector"`
}
