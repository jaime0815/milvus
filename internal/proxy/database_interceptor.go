package proxy

import (
	"context"

	"google.golang.org/grpc"

	"github.com/milvus-io/milvus-proto/go-api/milvuspb"
)

// DatabaseInterceptor fill dbname into request based on kv pair <"dbname": "xx"> in header
func DatabaseInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		filledCtx, filledReq := fillDatabase(ctx, req)
		return handler(filledCtx, filledReq)
	}
}

func fillDatabase(ctx context.Context, req interface{}) (context.Context, interface{}) {
	switch r := req.(type) {
	case *milvuspb.CreateCollectionRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.DropCollectionRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.HasCollectionRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.LoadCollectionRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.ReleaseCollectionRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.DescribeCollectionRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.GetStatisticsRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.GetCollectionStatisticsRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.ShowCollectionsRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.AlterCollectionRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.CreatePartitionRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.DropPartitionRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.HasPartitionRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.LoadPartitionsRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.ReleasePartitionsRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.GetPartitionStatisticsRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.ShowPartitionsRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.GetLoadingProgressRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.GetLoadStateRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.CreateIndexRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.DescribeIndexRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.DropIndexRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.GetIndexBuildProgressRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.GetIndexStateRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.InsertRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.DeleteRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.SearchRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.FlushRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.QueryRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.CreateAliasRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.DropAliasRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.AlterAliasRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.CalcDistanceRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.FlushAllRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.GetPersistentSegmentInfoRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.GetQuerySegmentInfoRequest:
		r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.DummyRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.GetMetricsRequest:
		return ctx, r
	case *milvuspb.LoadBalanceRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.GetReplicasRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.GetCompactionStateRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.ManualCompactionRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.GetCompactionPlansRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.GetFlushStateRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.GetFlushAllStateRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.ImportRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.GetImportStateRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.ListImportTasksRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.RenameCollectionRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	case *milvuspb.TransferReplicaRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrEmpty(ctx)
		return ctx, r
	default:
		return ctx, req
	}
}
