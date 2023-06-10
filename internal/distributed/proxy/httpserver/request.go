package httpserver

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
