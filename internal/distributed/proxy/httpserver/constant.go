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

	// MilvusTag struct tag const for milvus row based struct
	MilvusTag = `milvus`

	// MilvusSkipTagValue struct tag const for skip this field.
	MilvusSkipTagValue = `-`

	// MilvusTagSep struct tag const for attribute separator
	MilvusTagSep = `;`

	//MilvusTagName struct tag const for field name
	MilvusTagName = `NAME`
)
