package httpserver

import (
	"github.com/gin-gonic/gin"
	"github.com/milvus-io/milvus-proto/go-api/v2/commonpb"
	"github.com/milvus-io/milvus-proto/go-api/v2/milvuspb"
	"github.com/milvus-io/milvus-proto/go-api/v2/schemapb"
	"github.com/milvus-io/milvus/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"testing"
)

func generatePrimaryField(datatype schemapb.DataType) *schemapb.FieldSchema {
	return &schemapb.FieldSchema{
		FieldID:      common.StartOfUserFieldID,
		Name:         "book_id",
		IsPrimaryKey: true,
		Description:  "",
		DataType:     datatype,
		AutoID:       false,
	}
}

func generateCollectionSchema() *schemapb.CollectionSchema {
	return &schemapb.CollectionSchema{
		Name:        "book",
		Description: "",
		AutoID:      false,
		Fields: []*schemapb.FieldSchema{
			{
				FieldID:      common.StartOfUserFieldID,
				Name:         "book_id",
				IsPrimaryKey: true,
				Description:  "",
				DataType:     5,
				AutoID:       false,
			}, {
				FieldID:      common.StartOfUserFieldID + 1,
				Name:         "word_count",
				IsPrimaryKey: false,
				Description:  "",
				DataType:     5,
				AutoID:       false,
			}, {
				FieldID:      common.StartOfUserFieldID + 2,
				Name:         "book_intro",
				IsPrimaryKey: false,
				Description:  "",
				DataType:     101,
				AutoID:       false,
			},
		},
		EnableDynamicField: true,
	}
}

func generateIndexes() []*milvuspb.IndexDescription {
	return []*milvuspb.IndexDescription{
		{
			IndexName: "_default_idx_102",
			IndexID:   442051985533243300,
			Params: []*commonpb.KeyValuePair{
				{
					Key:   common.MetricTypeKey,
					Value: DefaultMetricType,
				},
				{
					Key:   "index_type",
					Value: "IVF_FLAT",
				}, {
					Key:   Params,
					Value: "{\"nlist\":1024}",
				},
			},
			State:     3,
			FieldName: "book_intro",
		},
	}
}

func TestPrintCollectionDetails(t *testing.T) {
	coll := generateCollectionSchema()
	indexes := generateIndexes()
	assert.Equal(t, printFields(coll.Fields), []gin.H{
		{
			HTTPReturnFieldName:       "book_id",
			HTTPReturnFieldType:       "Int64",
			HTTPReturnFieldPrimaryKey: true,
			HTTPReturnFieldAutoID:     false,
			HTTPReturnDescription:     ""},
		{
			HTTPReturnFieldName:       "word_count",
			HTTPReturnFieldType:       "Int64",
			HTTPReturnFieldPrimaryKey: false,
			HTTPReturnFieldAutoID:     false,
			HTTPReturnDescription:     ""},
		{
			HTTPReturnFieldName:       "book_intro",
			HTTPReturnFieldType:       "FloatVector",
			HTTPReturnFieldPrimaryKey: false,
			HTTPReturnFieldAutoID:     false,
			HTTPReturnDescription:     ""},
	})
	assert.Equal(t, printIndexes(indexes), []gin.H{
		{
			HTTPReturnIndexName:        "_default_idx_102",
			HTTPReturnIndexField:       "book_intro",
			HTTPReturnIndexMetricsType: DefaultMetricType},
	})
	assert.Equal(t, getMetricType(indexes[0].Params), DefaultMetricType)
}

func TestPrimaryField(t *testing.T) {
	coll := generateCollectionSchema()
	primaryField := generatePrimaryField(schemapb.DataType_Int64)
	assert.Equal(t, getPrimaryField(coll), primaryField)

	assert.Equal(t, joinArray([]int64{1, 2, 3}), "1,2,3")
	assert.Equal(t, joinArray([]string{"1", "2", "3"}), "\"1\",\"2\",\"3\"")

	jsonStr := "{\"id\": [1, 2, 3]}"
	idStr := gjson.Get(jsonStr, "id")
	rangeStr, err := convertRange(primaryField, idStr)
	assert.Equal(t, err, nil)
	assert.Equal(t, rangeStr, "1,2,3")
	assert.Equal(t, checkGetPrimaryKey(coll, idStr), "book_id in [1,2,3]")

	primaryField = generatePrimaryField(schemapb.DataType_VarChar)
	jsonStr = "{\"id\": [\"1\", \"2\", \"3\"]}"
	idStr = gjson.Get(jsonStr, "id")
	rangeStr, err = convertRange(primaryField, idStr)
	assert.Equal(t, err, nil)
	assert.Equal(t, rangeStr, "\"1\",\"2\",\"3\"")
	assert.Equal(t, checkGetPrimaryKey(coll, idStr), "book_id in [\"1\",\"2\",\"3\"]")
}

func TestDynamicFields(t *testing.T) {
	outputFields := []string{"book_id", "word_count", "author", "date"}
	allFields := []string{"book_id", "word_count", "book_intro"}
	assert.Equal(t, getNotExistData(outputFields, allFields), []string{"author", "date"})
	fields, dynamicFields := getDynamicOutputFields(generateCollectionSchema(), outputFields)
	assert.Equal(t, fields, allFields)
	assert.Equal(t, dynamicFields, []string{"author", "date"})
}

func TestSerialize(t *testing.T) {
	parameters := []float32{0.11111, 0.22222}
	assert.Equal(t, string(serialize(parameters)), "\ufffd\ufffd\ufffd=\ufffd\ufffdc\u003e")
	assert.Equal(t, string(vector2PlaceholderGroupBytes(parameters)), "vector2PlaceholderGroupBytes") // todo
}
