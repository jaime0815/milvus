package rootcoord

import (
	"github.com/milvus-io/milvus/internal/model"
	"github.com/milvus-io/milvus/internal/util/typeutil"
)

type Catalog interface {
	CreateCollection(collectionInfo *model.Collection, ts typeutil.Timestamp, meta map[string]string) error
}
