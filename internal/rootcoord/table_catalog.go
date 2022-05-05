package rootcoord

import (
	"context"

	"github.com/milvus-io/milvus/internal/db"
	"github.com/milvus-io/milvus/internal/proto/modelpb"
	"github.com/milvus-io/milvus/internal/util/typeutil"
)

type TableCatalog struct {
	db *db.DB
}

func (tc *TableCatalog) CreateCollection(ctx context.Context, collectionInfo *modelpb.Collection, ts typeutil.Timestamp) error {

	return nil
}
