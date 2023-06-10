package httpserver

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang/protobuf/proto"
	"github.com/milvus-io/milvus-proto/go-api/v2/commonpb"
	"github.com/milvus-io/milvus-proto/go-api/v2/milvuspb"
	"github.com/milvus-io/milvus-proto/go-api/v2/schemapb"
	"github.com/milvus-io/milvus/internal/common"
	"github.com/milvus-io/milvus/internal/log"
	"github.com/milvus-io/milvus/internal/proxy"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

func (h *Handlers) describeCollection(c *gin.Context, collectionName string, needAuth bool) (*milvuspb.DescribeCollectionResponse, error) {
	req := milvuspb.DescribeCollectionRequest{
		CollectionName: collectionName,
	}
	if needAuth {
		_, authErr := proxy.PrivilegeInterceptor(c, &req)
		if authErr != nil {
			c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusUnauthorized, HTTPReturnMessage: authErr.Error()})
			return nil, authErr
		}
	}
	response, err := h.proxy.DescribeCollection(c, &req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "describe collection " + collectionName + " fail", HTTPReturnError: err.Error()})
		return nil, err
	} else if response.Status.ErrorCode != commonpb.ErrorCode_Success {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{HTTPReturnCode: response.Status.ErrorCode, HTTPReturnMessage: response.Status.Reason})
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
		CollectionName: collectionName,
	}
	response, err := h.proxy.HasCollection(c, &req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: err})
		return false, err
	}
	return response.Value, nil
}

func (h *Handlers) RegisterRoutesToV1(router gin.IRouter) {
	router.GET(VectorCollectionsPath, h.listCollections)
	router.POST(VectorCollectionsCreatePath, h.createCollection)
	router.GET(VectorCollectionsDescribePath, h.getCollectionDetails)
	router.POST(VectorCollectionsDropPath, h.dropCollection)
	router.POST(VectorQueryPath, h.query)
	router.POST(VectorGetPath, h.get)
	router.POST(VectorDeletePath, h.delete)
	router.POST(VectorInsertPath, h.insert)
	router.POST(VectorSearchPath, h.search)
}

func (h *Handlers) listCollections(c *gin.Context) {
	req := milvuspb.ShowCollectionsRequest{}
	_, authErr := proxy.PrivilegeInterceptor(c, &req)
	if authErr != nil {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusUnauthorized, HTTPReturnMessage: authErr.Error()})
		return
	}
	response, err := h.proxy.ShowCollections(c, &req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "show collections fail", HTTPReturnError: err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusOK, HTTPReturnData: response.CollectionNames})
	}
}

func (h *Handlers) createCollection(c *gin.Context) {
	httpReq := CreateCollectionReq{
		MetricType:   DefaultMetricType,
		PrimaryField: DefaultPrimaryFieldName,
		VectorField:  DefaultVectorFieldName,
	}
	if err := c.ShouldBindBodyWith(&httpReq, binding.JSON); err != nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "check your parameters conform to the json format", HTTPReturnError: err.Error()})
		return
	}
	if httpReq.CollectionName == "" || httpReq.Dimension == 0 {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "collectionName and dimension are both required."})
		return
	}
	has, err := h.hasCollection(c, httpReq.CollectionName)
	if err != nil {
		return
	}
	if has {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusOK, HTTPReturnMessage: "collection " + httpReq.CollectionName + " already exist."})
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
						Key:   Dim,
						Value: strconv.FormatInt(int64(httpReq.Dimension), 10),
					},
				},
				AutoID: DisableAutoID,
			},
		},
		EnableDynamicField: EnableDynamic,
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "marshal collection schema to string", HTTPReturnError: err.Error()})
		return
	}
	req := milvuspb.CreateCollectionRequest{
		CollectionName:   httpReq.CollectionName,
		Schema:           schema,
		ShardsNum:        ShardNumDefault,
		ConsistencyLevel: commonpb.ConsistencyLevel_Bounded,
	}
	_, authErr := proxy.PrivilegeInterceptor(c, &req)
	if authErr != nil {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusUnauthorized, HTTPReturnMessage: authErr.Error()})
		return
	}
	_, err = h.proxy.CreateCollection(c, &req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "create collection " + httpReq.CollectionName + " fail", HTTPReturnError: err.Error()})
		return
	}
	_, err = h.proxy.CreateIndex(c, &milvuspb.CreateIndexRequest{
		CollectionName: httpReq.CollectionName,
		FieldName:      httpReq.VectorField,
		ExtraParams:    []*commonpb.KeyValuePair{{Key: common.MetricTypeKey, Value: httpReq.MetricType}},
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "create index for collection " + httpReq.CollectionName + " fail, after the collection was created", HTTPReturnError: err.Error()})
		return
	}
	_, err = h.proxy.LoadCollection(c, &milvuspb.LoadCollectionRequest{
		CollectionName: httpReq.CollectionName,
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "load collection " + httpReq.CollectionName + " fail, after the index was created", HTTPReturnError: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusOK, HTTPReturnData: gin.H{}})
}

func (h *Handlers) getCollectionDetails(c *gin.Context) {
	collectionName := c.Query(HTTPCollectionName)
	coll, err := h.describeCollection(c, collectionName, true)
	if err != nil {
		return
	}
	stateResp, stateErr := h.proxy.GetLoadState(c, &milvuspb.GetLoadStateRequest{
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
		CollectionName: collectionName,
		FieldName:      vectorField,
	})
	if indexErr != nil {
		log.Warn("get indexes description fail", zap.String("collection", collectionName), zap.String("vectorField", vectorField), zap.String("err", indexErr.Error()))
	} else if indexResp.Status.ErrorCode != commonpb.ErrorCode_Success {
		log.Warn("get indexes description fail", zap.String("collection", collectionName), zap.String("vectorField", vectorField), zap.String("err", indexResp.Status.Reason))
	}
	c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusOK, HTTPReturnData: gin.H{
		HTTPCollectionName:    coll.CollectionName,
		HTTPReturnDescription: coll.Schema.Description,
		"fields":              printFields(coll.Schema.Fields),
		"indexes":             printIndexes(indexResp.IndexDescriptions),
		"load":                stateResp.State.String(),
		"shardsNum":           coll.ShardsNum,
		"enableDynamic":       coll.Schema.EnableDynamicField,
	}})
}

func (h *Handlers) dropCollection(c *gin.Context) {
	httpReq := DropCollectionReq{}
	if err := c.ShouldBindBodyWith(&httpReq, binding.JSON); err != nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "check your parameters conform to the json format", HTTPReturnError: err.Error()})
		return
	}
	has, err := h.hasCollection(c, httpReq.CollectionName)
	if err != nil {
		return
	}
	if !has {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusOK, HTTPReturnMessage: "can't find collection:  " + httpReq.CollectionName})
		return
	}
	req := milvuspb.DropCollectionRequest{
		CollectionName: httpReq.CollectionName,
	}
	_, authErr := proxy.PrivilegeInterceptor(c, &req)
	if authErr != nil {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusUnauthorized, HTTPReturnMessage: authErr.Error()})
		return
	}
	response, err := h.proxy.DropCollection(c, &req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "drop collection " + httpReq.CollectionName + " fail", HTTPReturnError: err.Error()})
	} else if response.ErrorCode != commonpb.ErrorCode_Success {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: response.ErrorCode, HTTPReturnMessage: response.Reason})
	} else {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusOK, HTTPReturnData: gin.H{}})
	}
}

func (h *Handlers) query(c *gin.Context) {
	httpReq := QueryReq{
		Limit: 100,
	}
	if err := c.ShouldBindBodyWith(&httpReq, binding.JSON); err != nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "check your parameters conform to the json format", HTTPReturnError: err.Error()})
		return
	}
	coll, err := h.describeCollection(c, httpReq.CollectionName, false)
	if err != nil || coll == nil {
		return
	}
	dynamicFields := []string{}
	httpReq.OutputFields, dynamicFields = getDynamicOutputFields(coll.Schema, httpReq.OutputFields)
	req := milvuspb.QueryRequest{
		CollectionName:     httpReq.CollectionName,
		Expr:               httpReq.Filter,
		OutputFields:       httpReq.OutputFields,
		GuaranteeTimestamp: BoundedTimestamp,
	}
	_, authErr := proxy.PrivilegeInterceptor(c, &req)
	if authErr != nil {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusUnauthorized, HTTPReturnMessage: authErr.Error()})
		return
	}
	response, err := h.proxy.Query(c, &req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "query fail", HTTPReturnError: err.Error()})
	} else if response.Status.ErrorCode != commonpb.ErrorCode_Success {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: response.Status.ErrorCode, HTTPReturnMessage: response.Status.Reason})
	} else {
		outputData, err := buildQueryResp(int64(0), dynamicFields, response.FieldsData, nil)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "show result by row wrong", "originData": response.FieldsData, HTTPReturnError: err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusOK, HTTPReturnData: outputData})
		}
	}
}

func (h *Handlers) get(c *gin.Context) {
	httpReq := GetReq{}
	if err := c.ShouldBindBodyWith(&httpReq, binding.JSON); err != nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "check your parameters conform to the json format", HTTPReturnError: err.Error()})
		return
	}
	coll, err := h.describeCollection(c, httpReq.CollectionName, false)
	if err != nil || coll == nil {
		return
	}
	dynamicFields := []string{}
	httpReq.OutputFields, dynamicFields = getDynamicOutputFields(coll.Schema, httpReq.OutputFields)
	body, _ := c.Get(gin.BodyBytesKey)
	filter := checkGetPrimaryKey(coll.Schema, gjson.Get(string(body.([]byte)), "id"))
	req := milvuspb.QueryRequest{
		CollectionName:     httpReq.CollectionName,
		Expr:               filter,
		OutputFields:       httpReq.OutputFields,
		GuaranteeTimestamp: BoundedTimestamp,
	}
	_, authErr := proxy.PrivilegeInterceptor(c, &req)
	if authErr != nil {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusUnauthorized, HTTPReturnMessage: authErr.Error()})
		return
	}
	response, err := h.proxy.Query(c, &req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "query fail", HTTPReturnError: err.Error()})
	} else if response.Status.ErrorCode != commonpb.ErrorCode_Success {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: response.Status.ErrorCode, HTTPReturnMessage: response.Status.Reason})
	} else {
		outputData, err := buildQueryResp(int64(0), dynamicFields, response.FieldsData, nil)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "show result by row wrong", "originData": response.FieldsData, HTTPReturnError: err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusOK, HTTPReturnData: outputData})
			log.Error("get resultIS: ", zap.Any("res", outputData))
		}
	}
}

func (h *Handlers) delete(c *gin.Context) {
	httpReq := DeleteReq{}
	if err := c.ShouldBindBodyWith(&httpReq, binding.JSON); err != nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "check your parameters conform to the json format", HTTPReturnError: err.Error()})
		return
	}
	coll, err := h.describeCollection(c, httpReq.CollectionName, false)
	if err != nil || coll == nil {
		return
	}
	body, _ := c.Get(gin.BodyBytesKey)
	expr := checkGetPrimaryKey(coll.Schema, gjson.Get(string(body.([]byte)), "id"))
	req := milvuspb.DeleteRequest{
		CollectionName: httpReq.CollectionName,
		Expr:           expr,
	}
	_, authErr := proxy.PrivilegeInterceptor(c, &req)
	if authErr != nil {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusUnauthorized, HTTPReturnMessage: authErr.Error()})
		return
	}
	response, err := h.proxy.Delete(c, &req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "delete fail", HTTPReturnError: err.Error()})
	} else if response.Status.ErrorCode != commonpb.ErrorCode_Success {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: response.Status.ErrorCode, HTTPReturnMessage: response.Status.Reason})
	} else {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusOK, HTTPReturnData: gin.H{}})
	}
}

func (h *Handlers) insert(c *gin.Context) {
	httpReq := InsertReq{}
	if err := c.ShouldBindBodyWith(&httpReq, binding.JSON); err != nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "check your parameters conform to the json format", HTTPReturnError: err.Error()})
		return
	}
	coll, err := h.describeCollection(c, httpReq.CollectionName, false)
	if err != nil || coll == nil {
		return
	}
	body, _ := c.Get(gin.BodyBytesKey)
	err = checkAndSetData(string(body.([]byte)), coll, &httpReq)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "checkout your params", "originRequest": httpReq, HTTPReturnError: err.Error()})
		return
	}
	req := milvuspb.InsertRequest{
		CollectionName: httpReq.CollectionName,
		PartitionName:  "_default",
		NumRows:        uint32(len(httpReq.Data)),
	}
	req.FieldsData, err = anyToColumns(httpReq.Data, coll.Schema)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "insert data by column wrong", "originData": httpReq.Data, "collectionSchema": coll.Schema, HTTPReturnError: err.Error()})
		return
	}
	_, authErr := proxy.PrivilegeInterceptor(c, &req)
	if authErr != nil {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusUnauthorized, HTTPReturnMessage: authErr.Error()})
		return
	}
	response, err := h.proxy.Insert(c, &req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "insert fail", HTTPReturnError: err.Error()})
	} else if response.Status.ErrorCode != commonpb.ErrorCode_Success {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: response.Status.ErrorCode, HTTPReturnMessage: response.Status.Reason, HTTPReturnData: req.FieldsData})
	} else {
		switch response.IDs.GetIdField().(type) {
		case *schemapb.IDs_IntId:
			c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusOK, HTTPReturnData: gin.H{"insertCount": response.InsertCnt, "insertIds": response.IDs.IdField.(*schemapb.IDs_IntId).IntId.Data}})
		case *schemapb.IDs_StrId:
			c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusOK, HTTPReturnData: gin.H{"insertCount": response.InsertCnt, "insertIds": response.IDs.IdField.(*schemapb.IDs_StrId).StrId.Data}})
		}
	}
}

func (h *Handlers) search(c *gin.Context) {
	httpReq := SearchReq{
		Limit: 100,
	}
	if err := c.ShouldBindBodyWith(&httpReq, binding.JSON); err != nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "check your parameters conform to the json format", HTTPReturnError: err.Error()})
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
		CollectionName: httpReq.CollectionName,
		FieldName:      vectorField,
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
		{Key: ParamAnnsField, Value: vectorField},
		{Key: common.TopKKey, Value: strconv.FormatInt(int64(httpReq.Limit), 10)},
		{Key: Params, Value: string(bs)},
		{Key: common.MetricTypeKey, Value: metricType},
		{Key: ParamRoundDecimal, Value: "-1"},
		{Key: ParamOffset, Value: strconv.FormatInt(int64(httpReq.Offset), 10)},
	}
	req := milvuspb.SearchRequest{
		CollectionName:     httpReq.CollectionName,
		Dsl:                httpReq.Filter,
		PlaceholderGroup:   vector2PlaceholderGroupBytes(httpReq.Vector),
		DslType:            commonpb.DslType_BoolExprV1,
		OutputFields:       httpReq.OutputFields,
		SearchParams:       searchParams,
		GuaranteeTimestamp: BoundedTimestamp,
		Nq:                 int64(1),
	}
	_, authErr := proxy.PrivilegeInterceptor(c, &req)
	if authErr != nil {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusUnauthorized, HTTPReturnMessage: authErr.Error()})
		return
	}
	response, err := h.proxy.Search(c, &req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "search fail", HTTPReturnError: err.Error()})
	} else if response.Status.ErrorCode != commonpb.ErrorCode_Success {
		c.JSON(http.StatusOK, gin.H{HTTPReturnCode: response.Status.ErrorCode, HTTPReturnMessage: response.Status.Reason})
	} else {
		outputData, err := buildQueryResp(response.Results.TopK, dynamicFields, response.Results.FieldsData, response.Results.Scores)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusBadRequest, HTTPReturnMessage: "show result by row wrong", "originResquest": req, "originResult": response.Results, HTTPReturnError: err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{HTTPReturnCode: http.StatusOK, HTTPReturnData: outputData})
		}
	}
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
