package httpserver

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

	ShardNumDefault = 2

	EnableDynamic = true
	EnableAutoID  = true
	DisableAutoID = false

	HTTPCollectionName = "collectionName"

	HTTPReturnCode    = "code"
	HTTPReturnMessage = "message"
	HTTPReturnError   = "error"
	HTTPReturnData    = "data"

	HTTPReturnFieldName       = "name"
	HTTPReturnFieldType       = "type"
	HTTPReturnFieldPrimaryKey = "primaryKey"
	HTTPReturnFieldAutoID     = "autoId"
	HTTPReturnDescription     = "description"

	HTTPReturnIndexName        = "indexName"
	HTTPReturnIndexField       = "fieldName"
	HTTPReturnIndexMetricsType = "metricType"

	DefaultMetricType       = "L2"
	DefaultPrimaryFieldName = "id"
	DefaultVectorFieldName  = "vector"

	Dim = "dim"
)

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

const (
	ParamAnnsField    = "anns_field"
	Params            = "params"
	ParamRoundDecimal = "round_decimal"
	ParamOffset       = "offset"
	BoundedTimestamp  = 2
)
