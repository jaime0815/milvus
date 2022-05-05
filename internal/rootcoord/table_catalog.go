package rootcoord

import (
	"github.com/milvus-io/milvus/internal/db"
	"github.com/milvus-io/milvus/internal/model"
	"github.com/milvus-io/milvus/internal/util/typeutil"
)

type TableCatalog struct {
	db *db.DB
}

func (tc *TableCatalog) CreateCollection(collectionInfo *model.Collection, ts typeutil.Timestamp, meta map[string]string) error {

	return nil
}
