package rootcoord

import (
	"context"

	"github.com/milvus-io/milvus/internal/proto/modelpb"
	"github.com/milvus-io/milvus/internal/util/typeutil"
)

type Catalog interface {
	CreateCollection(ctx context.Context, collectionInfo *modelpb.Collection, ts typeutil.Timestamp) error
}
