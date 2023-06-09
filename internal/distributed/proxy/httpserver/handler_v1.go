package httpserver

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang/protobuf/proto"
	"github.com/milvus-io/milvus-proto/go-api/v2/commonpb"
	"github.com/milvus-io/milvus-proto/go-api/v2/milvuspb"
	"github.com/milvus-io/milvus-proto/go-api/v2/schemapb"
	"github.com/milvus-io/milvus/internal/common"
	"github.com/milvus-io/milvus/internal/log"
	"github.com/milvus-io/milvus/internal/proxy"
	"github.com/milvus-io/milvus/internal/util/funcutil"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

const (
	VectorCollectionsPath         = "/vector/collections"
	VectorCollectionsCreatePath   = "/vector/collections/create"
	VectorCollectionsDescribePath = "/vector/collections/describe"
	VectorCollectionsDropPath     = "/vector/collections/drop"
	VectorInsertPath              = "/vector/insert"
	VectorSearchPath              = "/vector/search"
	VectorGetPath                 = "/vector/get"
	VectorQueryPath               = "/vector/query"
	VectorDeletePath              = "/vector/delete"

	Int8    = "int8"
	Int16   = "int16"
	Int32   = "int32"
	Int64   = "int64"
	Varchar = "varchar"
	String  = "string"
	Float64 = "float64"
	Float32 = "float32"
	Bool    = "bool"

	InnerFloatVector  = "floatvector"
	InnerBinaryVector = "binaryvector"

	OuterFloatVector  = "floatVector"
	OuterBinaryVector = "binaryVector"

	MetricType              = "metric_type"
	VectorIndexDefaultName  = "vector_idx"
	VectorFieldDefaultName  = "vector"
	PrimaryFieldDefaultName = "id"

	CollectionNameRegex = "^[A-Za-z_]{1}[A-Za-z0-9_]{0,254}$"
	CollNameLengthMin   = 1
	CollNameLengthMax   = 255

	ShardNumMix     = 1
	ShardNumMax     = 32
	ShardNumDefault = 2

	EnableDynamic = true
	EnableAutoID  = true
	DisableAutoID = false

	DefaultListenPort = "9092"
	ListenPortEnvKey  = "RESTFUL_API_PORT"
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

func (h *Handlers) RegisterRoutesToV1(router gin.IRouter) {
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

func getPrimaryField(schema *schemapb.CollectionSchema) *schemapb.FieldSchema {
	primaryField := schema.Fields[0]
	for _, field := range schema.Fields {
		if field.IsPrimaryKey {
			primaryField = field
			break
		}
	}
	return primaryField
}

func getDim(field *schemapb.FieldSchema) (int64, error) {
	dimensionInSchema, err := funcutil.GetAttrByKeyFromRepeatedKV("dim", field.TypeParams)
	if err != nil {
		return 0, err
	}
	dim, err := strconv.Atoi(dimensionInSchema)
	if err != nil {
		return 0, err
	}
	return int64(dim), nil
}

func anyToColumns(rows []map[string]interface{}, sch *schemapb.CollectionSchema) ([]*schemapb.FieldData, error) {
	rowsLen := len(rows)
	if rowsLen == 0 {
		return []*schemapb.FieldData{}, errors.New("0 length column")
	}

	isDynamic := sch.EnableDynamicField
	var dim int64 = 0

	nameColumns := make(map[string]interface{})
	fieldData := make(map[string]*schemapb.FieldData)
	for _, field := range sch.Fields {
		// skip auto id pk field
		if field.IsPrimaryKey && field.AutoID {
			continue
		}
		var data interface{}
		switch field.DataType {
		case schemapb.DataType_Bool:
			data = make([]bool, 0, rowsLen)
		case schemapb.DataType_Int8:
			data = make([]int8, 0, rowsLen)
		case schemapb.DataType_Int16:
			data = make([]int16, 0, rowsLen)
		case schemapb.DataType_Int32:
			data = make([]int32, 0, rowsLen)
		case schemapb.DataType_Int64:
			data = make([]int64, 0, rowsLen)
		case schemapb.DataType_Float:
			data = make([]float32, 0, rowsLen)
		case schemapb.DataType_Double:
			data = make([]float64, 0, rowsLen)
		case schemapb.DataType_String:
			data = make([]string, 0, rowsLen)
		case schemapb.DataType_JSON:
			data = make([][]byte, 0, rowsLen)
		case schemapb.DataType_FloatVector:
			data = make([][]float32, 0, rowsLen)
			dim, _ = getDim(field)
		case schemapb.DataType_BinaryVector:
			data = make([][]byte, 0, rowsLen)
			dim, _ = getDim(field)
		}
		nameColumns[field.Name] = data
		fieldData[field.Name] = &schemapb.FieldData{
			Type:      field.DataType,
			FieldName: field.Name,
			FieldId:   field.FieldID,
			IsDynamic: field.IsDynamic,
		}
	}
	if dim == 0 {
		return nil, errors.New("cannot find dimension")
	}

	dynamicCol := make([][]byte, 0, rowsLen)

	for _, row := range rows {
		// collection schema name need not to be same, since receiver could has other names
		v := reflect.ValueOf(row)
		set, err := reflectValueCandi(v)
		if err != nil {
			return nil, err
		}

		for idx, field := range sch.Fields {
			// skip dynamic field if visible
			if isDynamic && field.IsDynamic {
				continue
			}
			// skip auto id pk field
			if field.IsPrimaryKey && field.AutoID {
				// remove pk field from candidates set, avoid adding it into dynamic column
				delete(set, field.Name)
				continue
			}

			candi, ok := set[field.Name]
			if !ok {
				return nil, fmt.Errorf("row %d does not has field %s", idx, field.Name)
			}
			switch field.DataType {
			case schemapb.DataType_Bool:
				nameColumns[field.Name] = append(nameColumns[field.Name].([]bool), candi.v.Interface().(bool))
			case schemapb.DataType_Int8:
				nameColumns[field.Name] = append(nameColumns[field.Name].([]int8), candi.v.Interface().(int8))
			case schemapb.DataType_Int16:
				nameColumns[field.Name] = append(nameColumns[field.Name].([]int16), candi.v.Interface().(int16))
			case schemapb.DataType_Int32:
				nameColumns[field.Name] = append(nameColumns[field.Name].([]int32), candi.v.Interface().(int32))
			case schemapb.DataType_Int64:
				nameColumns[field.Name] = append(nameColumns[field.Name].([]int64), candi.v.Interface().(int64))
			case schemapb.DataType_Float:
				nameColumns[field.Name] = append(nameColumns[field.Name].([]float32), candi.v.Interface().(float32))
			case schemapb.DataType_Double:
				nameColumns[field.Name] = append(nameColumns[field.Name].([]float64), candi.v.Interface().(float64))
			case schemapb.DataType_String:
				nameColumns[field.Name] = append(nameColumns[field.Name].([]string), candi.v.Interface().(string))
			case schemapb.DataType_JSON:
				nameColumns[field.Name] = append(nameColumns[field.Name].([][]byte), candi.v.Interface().([]byte))
			case schemapb.DataType_FloatVector:
				nameColumns[field.Name] = append(nameColumns[field.Name].([][]float32), candi.v.Interface().([]float32))
			case schemapb.DataType_BinaryVector:
				nameColumns[field.Name] = append(nameColumns[field.Name].([][]byte), candi.v.Interface().([]byte))
			}

			delete(set, field.Name)
		}

		if isDynamic {
			m := make(map[string]interface{})
			for name, candi := range set {
				m[name] = candi.v.Interface()
			}
			bs, err := json.Marshal(m)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal dynamic field %w", err)
			}
			dynamicCol = append(dynamicCol, bs)
			if err != nil {
				return nil, fmt.Errorf("failed to append value to dynamic field %w", err)
			}
		}
	}
	columns := make([]*schemapb.FieldData, 0, len(nameColumns))
	for name, column := range nameColumns {
		colData, _ := fieldData[name]
		switch colData.Type {
		case schemapb.DataType_Bool:
			colData.Field = &schemapb.FieldData_Scalars{
				Scalars: &schemapb.ScalarField{
					Data: &schemapb.ScalarField_BoolData{
						BoolData: &schemapb.BoolArray{
							Data: column.([]bool),
						},
					},
				},
			}
			break
		case schemapb.DataType_Int8:
			colData.Field = &schemapb.FieldData_Scalars{
				Scalars: &schemapb.ScalarField{
					Data: &schemapb.ScalarField_IntData{
						IntData: &schemapb.IntArray{
							Data: column.([]int32),
						},
					},
				},
			}
		case schemapb.DataType_Int16:
			colData.Field = &schemapb.FieldData_Scalars{
				Scalars: &schemapb.ScalarField{
					Data: &schemapb.ScalarField_IntData{
						IntData: &schemapb.IntArray{
							Data: column.([]int32),
						},
					},
				},
			}
		case schemapb.DataType_Int32:
			colData.Field = &schemapb.FieldData_Scalars{
				Scalars: &schemapb.ScalarField{
					Data: &schemapb.ScalarField_IntData{
						IntData: &schemapb.IntArray{
							Data: column.([]int32),
						},
					},
				},
			}
		case schemapb.DataType_Int64:
			colData.Field = &schemapb.FieldData_Scalars{
				Scalars: &schemapb.ScalarField{
					Data: &schemapb.ScalarField_LongData{
						LongData: &schemapb.LongArray{
							Data: column.([]int64),
						},
					},
				},
			}
		case schemapb.DataType_Float:
			colData.Field = &schemapb.FieldData_Scalars{
				Scalars: &schemapb.ScalarField{
					Data: &schemapb.ScalarField_FloatData{
						FloatData: &schemapb.FloatArray{
							Data: column.([]float32),
						},
					},
				},
			}
		case schemapb.DataType_Double:
			colData.Field = &schemapb.FieldData_Scalars{
				Scalars: &schemapb.ScalarField{
					Data: &schemapb.ScalarField_DoubleData{
						DoubleData: &schemapb.DoubleArray{
							Data: column.([]float64),
						},
					},
				},
			}
		case schemapb.DataType_String:
			colData.Field = &schemapb.FieldData_Scalars{
				Scalars: &schemapb.ScalarField{
					Data: &schemapb.ScalarField_StringData{
						StringData: &schemapb.StringArray{
							Data: column.([]string),
						},
					},
				},
			}
		case schemapb.DataType_JSON:
			colData.Field = &schemapb.FieldData_Scalars{
				Scalars: &schemapb.ScalarField{
					Data: &schemapb.ScalarField_BytesData{
						BytesData: &schemapb.BytesArray{
							Data: column.([][]byte),
						},
					},
				},
			}
		case schemapb.DataType_FloatVector:
			arr, err := convertFloatVectorToArray(column.([][]float32), dim)
			if err != nil {
				return nil, err
			}
			colData.Field = &schemapb.FieldData_Vectors{
				Vectors: &schemapb.VectorField{
					Dim: dim,
					Data: &schemapb.VectorField_FloatVector{
						FloatVector: &schemapb.FloatArray{
							Data: arr,
						},
					},
				},
			}
		case schemapb.DataType_BinaryVector:
			arr, err := convertBinaryVectorToArray(column.([][]byte), dim)
			if err != nil {
				return nil, err
			}
			colData.Field = &schemapb.FieldData_Vectors{
				Vectors: &schemapb.VectorField{
					Dim: dim,
					Data: &schemapb.VectorField_BinaryVector{
						BinaryVector: arr,
					},
				},
			}
		}
		columns = append(columns, colData)
	}
	if isDynamic {
		columns = append(columns, &schemapb.FieldData{
			Type:      0,
			FieldName: "",
			Field: &schemapb.FieldData_Scalars{
				Scalars: &schemapb.ScalarField{
					Data: &schemapb.ScalarField_JsonData{
						JsonData: &schemapb.JSONArray{
							Data: dynamicCol,
						},
					},
				},
			},
			FieldId:   0,
			IsDynamic: true,
		})
	}
	return columns, nil
}

func convertFloatVectorToArray(vector [][]float32, dim int64) ([]float32, error) {
	floatArray := make([]float32, 0)
	for _, arr := range vector {
		if int64(len(arr)) != dim {
			return nil, errors.New("wrong")
		}
		for i := int64(0); i < dim; i++ {
			floatArray = append(floatArray, arr[i])
		}
	}
	return floatArray, nil
}

func convertBinaryVectorToArray(vector [][]byte, dim int64) ([]byte, error) {
	binaryArray := make([]byte, 0)
	for _, arr := range vector {
		if int64(len(arr)) != dim {
			return nil, errors.New("wrong")
		}
		for i := int64(0); i < dim; i++ {
			binaryArray = append(binaryArray, arr[i])
		}
	}
	return binaryArray, nil
}

type fieldCandi struct {
	name    string
	v       reflect.Value
	options map[string]string
}

const (
	// MilvusTag struct tag const for milvus row based struct
	MilvusTag = `milvus`

	// MilvusSkipTagValue struct tag const for skip this field.
	MilvusSkipTagValue = `-`

	// MilvusTagSep struct tag const for attribute separator
	MilvusTagSep = `;`

	//MilvusTagName struct tag const for field name
	MilvusTagName = `NAME`
)

func reflectValueCandi(v reflect.Value) (map[string]fieldCandi, error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	result := make(map[string]fieldCandi)
	switch v.Kind() {
	case reflect.Map: // map[string]interface{}
		iter := v.MapRange()
		for iter.Next() {
			key := iter.Key().String()
			result[key] = fieldCandi{
				name: key,
				v:    iter.Value(),
			}
		}
		return result, nil
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			ft := v.Type().Field(i)
			name := ft.Name
			tag, ok := ft.Tag.Lookup(MilvusTag)

			settings := make(map[string]string)
			if ok {
				if tag == MilvusSkipTagValue {
					continue
				}
				settings = parseTagSetting(tag, MilvusTagSep)
				fn, has := settings[MilvusTagName]
				if has {
					// overwrite column to tag name
					name = fn
				}
			}
			_, ok = result[name]
			// duplicated
			if ok {
				return nil, fmt.Errorf("column has duplicated name: %s when parsing field: %s", name, ft.Name)
			}

			v := v.Field(i)
			if v.Kind() == reflect.Array {
				v = v.Slice(0, v.Len()-1)
			}

			result[name] = fieldCandi{
				name:    name,
				v:       v,
				options: settings,
			}
		}

		return result, nil
	default:
		return nil, fmt.Errorf("unsupport row type: %s", v.Kind().String())
	}
}

func parseTagSetting(str string, sep string) map[string]string {
	settings := map[string]string{}
	names := strings.Split(str, sep)

	for i := 0; i < len(names); i++ {
		j := i
		if len(names[j]) > 0 {
			for {
				if names[j][len(names[j])-1] == '\\' {
					i++
					names[j] = names[j][0:len(names[j])-1] + sep + names[i]
					names[i] = ""
				} else {
					break
				}
			}
		}

		values := strings.Split(names[j], ":")
		k := strings.TrimSpace(strings.ToUpper(values[0]))

		if len(values) >= 2 {
			settings[k] = strings.Join(values[1:], ":")
		} else if k != "" {
			settings[k] = k
		}
	}

	return settings
}

func vector2Placeholder(vectors []float32) *commonpb.PlaceholderValue {
	var placeHolderType commonpb.PlaceholderType
	ph := &commonpb.PlaceholderValue{
		Tag:    "$0",
		Values: make([][]byte, 0, len(vectors)),
	}
	if len(vectors) == 0 {
		return ph
	}
	placeHolderType = commonpb.PlaceholderType_FloatVector

	ph.Type = placeHolderType
	ph.Values = append(ph.Values, serialize(vectors))
	return ph
}

func serialize(fv []float32) []byte {
	data := make([]byte, 0, 4*len(fv)) // float32 occupies 4 bytes
	buf := make([]byte, 4)
	for _, f := range fv {
		binary.LittleEndian.PutUint32(buf, math.Float32bits(f))
		data = append(data, buf...)
	}
	return data
}

func vector2PlaceholderGroupBytes(floatArray []float32) []byte {
	phg := &commonpb.PlaceholderGroup{
		Placeholders: []*commonpb.PlaceholderValue{
			vector2Placeholder(floatArray),
		},
	}

	bs, _ := proto.Marshal(phg)
	return bs
}

func getDynamicOutputFields(schema *schemapb.CollectionSchema, outputFields []string) ([]string, []string) {
	fieldNames := []string{}
	for _, field := range schema.Fields {
		if field.IsPrimaryKey && !containsString(outputFields, field.Name) {
			outputFields = append(outputFields, field.Name)
		}
		fieldNames = append(fieldNames, field.Name)
	}
	if schema.EnableDynamicField {
		return outputFields, getNotExistData(outputFields, fieldNames)
	} else {
		return outputFields, []string{}
	}
}

func getNotExistData(arrayA []string, allArray []string) []string {
	// 使用一个 map 来记录数组 B 中的所有元素
	arrayBMap := make(map[string]bool)
	for _, value := range allArray {
		arrayBMap[value] = true
	}

	// 遍历数组 A，检查每个元素是否在该 map 中出现过
	resultArray := []string{}
	for _, value := range arrayA {
		if _, ok := arrayBMap[value]; !ok {
			resultArray = append(resultArray, value)
		}
	}

	return resultArray
}

func convertRange(field *schemapb.FieldSchema, result gjson.Result) (string, error) {
	var resultStr string
	fieldType := field.DataType

	if fieldType == schemapb.DataType_Int64 {
		dataArray := []int64{}
		for _, data := range result.Array() {
			if data.Type == gjson.String {
				value, err := cast.ToInt64E(data.Str)
				if err != nil {
					return "", err
				}
				dataArray = append(dataArray, value)
			} else {
				value, err := cast.ToInt64E(data.Raw)
				if err != nil {
					return "", err
				}
				dataArray = append(dataArray, value)
			}
		}
		resultStr = JoinArray(dataArray)
	} else if fieldType == schemapb.DataType_VarChar {
		dataArray := []string{}
		for _, data := range result.Array() {
			value, err := cast.ToStringE(data.Str)
			if err != nil {
				return "", err
			}
			dataArray = append(dataArray, value)
		}
		resultStr = JoinArray(dataArray)
	}
	return resultStr, nil
}

func JoinArray(data interface{}) string {
	var buffer bytes.Buffer
	arr := reflect.ValueOf(data)

	for i := 0; i < arr.Len(); i++ {
		if i > 0 {
			buffer.WriteString(",")
		}

		buffer.WriteString(fmt.Sprintf("%v", arr.Index(i)))
	}

	return buffer.String()
}

func checkGetPrimaryKey(coll *milvuspb.DescribeCollectionResponse, idResult gjson.Result) string {
	primaryField := getPrimaryField(coll.Schema)
	resultStr, _ := convertRange(primaryField, idResult)
	filter := primaryField.Name + " in [" + resultStr + "]"
	return filter
}

func containsString(arr []string, s string) bool {
	for _, str := range arr {
		if str == s {
			return true
		}
	}
	return false
}

func checkAndSetData(body string, collDescResp *milvuspb.DescribeCollectionResponse, req *InsertReq) error {
	reallyDataArray := []map[string]interface{}{}
	dataResult := gjson.Get(body, "data")
	dataResultArray := dataResult.Array()

	fieldNames := []string{}
	for _, field := range collDescResp.Schema.Fields {
		fieldNames = append(fieldNames, field.Name)
	}

	for _, data := range dataResultArray {
		reallyData := map[string]interface{}{}
		vectorArray := []float32{}
		if data.Type == gjson.JSON {
			for _, field := range collDescResp.Schema.Fields {
				fieldType := field.DataType
				fieldName := field.Name

				dataString := gjson.Get(data.Raw, fieldName).String()

				// 如果AutoId 则报错
				if field.IsPrimaryKey && collDescResp.Schema.AutoID {
					if dataString != "" {
						return errors.New(fmt.Sprintf("[checkAndSetData] fieldName %s AutoId already open, not support insert data %s", fieldName, dataString))
					} else {
						continue
					}
				}

				// 转换对应数据,同时检查数据合法性
				if fieldType == schemapb.DataType_FloatVector || fieldType == schemapb.DataType_BinaryVector {
					for _, vector := range gjson.Get(data.Raw, fieldName).Array() {
						vectorArray = append(vectorArray, cast.ToFloat32(vector.Num))
					}
					reallyData[fieldName] = vectorArray
				} else if fieldType == schemapb.DataType_Int8 {
					result, err := cast.ToInt8E(dataString)
					if err != nil {
						return errors.New(fmt.Sprintf("[checkAndSetData] dataString %s cast to int8 error: %s", dataString, err.Error()))
					}
					reallyData[fieldName] = result
				} else if fieldType == schemapb.DataType_Int16 {
					result, err := cast.ToInt16E(dataString)
					if err != nil {
						return errors.New(fmt.Sprintf("[checkAndSetData] dataString %s cast to int16 error: %s", dataString, err.Error()))
					}
					reallyData[fieldName] = result
				} else if fieldType == schemapb.DataType_Int32 {
					result, err := cast.ToInt32E(dataString)
					if err != nil {
						return errors.New(fmt.Sprintf("[checkAndSetData] dataString %s cast to int32 error: %s", dataString, err.Error()))
					}
					reallyData[fieldName] = result
				} else if fieldType == schemapb.DataType_Int64 {
					result, err := cast.ToInt64E(dataString)
					if err != nil {
						return errors.New(fmt.Sprintf("[checkAndSetData] dataString %s cast to int64 error: %s", dataString, err.Error()))
					}
					reallyData[fieldName] = result
				} else if fieldType == schemapb.DataType_VarChar || fieldType == schemapb.DataType_String {
					reallyData[fieldName] = dataString
				} else if fieldType == schemapb.DataType_Float {
					result, err := cast.ToFloat64E(dataString)
					if err != nil {
						return errors.New(fmt.Sprintf("[checkAndSetData] dataString %s cast to float64 error: %s", dataString, err.Error()))
					}
					reallyData[fieldName] = result
				} else if fieldType == schemapb.DataType_Bool {
					result, err := cast.ToBoolE(dataString)
					if err != nil {
						return errors.New(fmt.Sprintf("[checkAndSetData] dataString %s cast to bool error: %s", dataString, err.Error()))
					}
					reallyData[fieldName] = result
				} else {
					errMsg := fmt.Sprintf("[checkAndSetData] not support fieldName %s dataType %s", fieldName, fieldType)
					return errors.New(errMsg)
				}
			}

			// fill dynamic schema
			if collDescResp.Schema.EnableDynamicField {
				for mapKey, mapValue := range data.Map() {
					if !containsString(fieldNames, mapKey) {
						mapValueStr := mapValue.String()
						if mapValue.Type == gjson.True || mapValue.Type == gjson.False {
							reallyData[mapKey] = cast.ToBool(mapValueStr)
						} else if mapValue.Type == gjson.String {
							reallyData[mapKey] = mapValueStr
						} else if mapValue.Type == gjson.Number {
							if strings.Contains(mapValue.Raw, ".") {
								reallyData[mapKey] = cast.ToFloat64(mapValue.Raw)
							} else {
								reallyData[mapKey] = cast.ToInt64(mapValueStr)
							}
						} else if mapValue.Type == gjson.JSON {
							reallyData[mapKey] = mapValue.Value()
						} else {

						}
					}
				}
			}

			reallyDataArray = append(reallyDataArray, reallyData)
		} else {
			errMsg := fmt.Sprintf("[checkAndSetData] dataType %s not Json", data.Type)
			return errors.New(errMsg)
		}
	}
	req.Data = reallyDataArray
	return nil
}

func buildQueryResp(rowsNum int64, dynamicOutputFields []string, fieldDataList []*schemapb.FieldData, scores []float32) ([]map[string]interface{}, error) {
	if len(fieldDataList) == 0 {
		return nil, errors.New("no columns")
	}

	querysResp := []map[string]interface{}{}

	columnNum := len(fieldDataList)
	if rowsNum == int64(0) {
		dataType := fieldDataList[0].Type
		if dataType == schemapb.DataType_Bool {
			rowsNum = int64(len(fieldDataList[0].GetScalars().GetBoolData().Data))
		} else if dataType == schemapb.DataType_Int8 {
			rowsNum = int64(len(fieldDataList[0].GetScalars().GetIntData().Data))
		} else if dataType == schemapb.DataType_Int16 {
			rowsNum = int64(len(fieldDataList[0].GetScalars().GetIntData().Data))
		} else if dataType == schemapb.DataType_Int32 {
			rowsNum = int64(len(fieldDataList[0].GetScalars().GetIntData().Data))
		} else if dataType == schemapb.DataType_Int64 {
			rowsNum = int64(len(fieldDataList[0].GetScalars().GetLongData().Data))
		} else if dataType == schemapb.DataType_Float {
			rowsNum = int64(len(fieldDataList[0].GetScalars().GetFloatData().Data))
		} else if dataType == schemapb.DataType_Double {
			rowsNum = int64(len(fieldDataList[0].GetScalars().GetDoubleData().Data))
		} else if dataType == schemapb.DataType_VarChar {
			rowsNum = int64(len(fieldDataList[0].GetScalars().GetStringData().Data))
		} else if dataType == schemapb.DataType_BinaryVector {
			rowsNum = int64(len(fieldDataList[0].GetVectors().GetBinaryVector())*8) / fieldDataList[0].GetVectors().GetDim()
		} else if dataType == schemapb.DataType_FloatVector {
			rowsNum = int64(len(fieldDataList[0].GetVectors().GetFloatVector().Data)) / fieldDataList[0].GetVectors().GetDim()
		} else if dataType == schemapb.DataType_JSON {
			rowsNum = int64(len(fieldDataList[0].GetScalars().GetJsonData().Data))
		}
	}
	if scores != nil && rowsNum != int64(len(scores)) {
		return nil, errors.New("search result invalid")
	}
	for i := int64(0); i < rowsNum; i++ {
		row := map[string]interface{}{}
		for j := 0; j < columnNum; j++ {
			dataType := fieldDataList[j].Type
			if dataType == schemapb.DataType_Bool {
				row[fieldDataList[j].FieldName] = fieldDataList[j].GetScalars().GetBoolData().Data[i]
			} else if dataType == schemapb.DataType_Int8 {
				row[fieldDataList[j].FieldName] = fieldDataList[j].GetScalars().GetIntData().Data[i]
			} else if dataType == schemapb.DataType_Int16 {
				row[fieldDataList[j].FieldName] = fieldDataList[j].GetScalars().GetIntData().Data[i]
			} else if dataType == schemapb.DataType_Int32 {
				row[fieldDataList[j].FieldName] = fieldDataList[j].GetScalars().GetIntData().Data[i]
			} else if dataType == schemapb.DataType_Int64 {
				row[fieldDataList[j].FieldName] = fieldDataList[j].GetScalars().GetLongData().Data[i]
			} else if dataType == schemapb.DataType_Float {
				row[fieldDataList[j].FieldName] = fieldDataList[j].GetScalars().GetFloatData().Data[i]
			} else if dataType == schemapb.DataType_Double {
				row[fieldDataList[j].FieldName] = fieldDataList[j].GetScalars().GetDoubleData().Data[i]
			} else if dataType == schemapb.DataType_VarChar {
				row[fieldDataList[j].FieldName] = fieldDataList[j].GetScalars().GetStringData().Data[i]
			} else if dataType == schemapb.DataType_BinaryVector {
				row[fieldDataList[j].FieldName] = fieldDataList[j].GetVectors().GetBinaryVector()[i*(fieldDataList[j].GetVectors().GetDim()/8) : (i+1)*(fieldDataList[j].GetVectors().GetDim()/8)]
			} else if dataType == schemapb.DataType_FloatVector {
				row[fieldDataList[j].FieldName] = fieldDataList[j].GetVectors().GetFloatVector().Data[i*fieldDataList[j].GetVectors().GetDim() : (i+1)*fieldDataList[j].GetVectors().GetDim()]
			} else if dataType == schemapb.DataType_JSON {
				data, ok := fieldDataList[j].GetScalars().Data.(*schemapb.ScalarField_JsonData)
				if ok && !fieldDataList[j].IsDynamic {
					row[fieldDataList[j].FieldName] = string(data.JsonData.Data[i])
				} else {
					var dataMap map[string]interface{}

					err := json.Unmarshal(fieldDataList[j].GetScalars().GetJsonData().Data[i], &dataMap)
					if err != nil {
						msg := fmt.Sprintf("[BuildQueryResp] Unmarshal error %s", err.Error())
						log.Error(msg)
						return nil, err
					}

					if containsString(dynamicOutputFields, "*") || containsString(dynamicOutputFields, "$meta") {
						for key, value := range dataMap {
							row[key] = value
						}
					} else {
						for _, dynamicField := range dynamicOutputFields {
							if _, ok := dataMap[dynamicField]; ok {
								row[dynamicField] = dataMap[dynamicField]
							}
						}
					}

				}

			}
		}
		if scores != nil {
			row["distance"] = scores[i]
		}
		querysResp = append(querysResp, row)
	}

	return querysResp, nil
}
