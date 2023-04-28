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
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.DropCollectionRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.HasCollectionRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.LoadCollectionRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.ReleaseCollectionRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.DescribeCollectionRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetStatisticsRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetCollectionStatisticsRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.ShowCollectionsRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.AlterCollectionRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.CreatePartitionRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.DropPartitionRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.HasPartitionRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.LoadPartitionsRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.ReleasePartitionsRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetPartitionStatisticsRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.ShowPartitionsRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetLoadingProgressRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetLoadStateRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.CreateIndexRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.DescribeIndexRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.DropIndexRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetIndexBuildProgressRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetIndexStateRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.InsertRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.DeleteRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.SearchRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.FlushRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.QueryRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.CreateAliasRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.DropAliasRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.AlterAliasRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.CalcDistanceRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.FlushAllRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetPersistentSegmentInfoRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetQuerySegmentInfoRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.DummyRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetMetricsRequest:
		return ctx, r
	case *milvuspb.LoadBalanceRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetReplicasRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetCompactionStateRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.ManualCompactionRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetCompactionPlansRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetFlushStateRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetFlushAllStateRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.ImportRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetImportStateRequest:
		// TODO
		// r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.ListImportTasksRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.RenameCollectionRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.TransferReplicaRequest:
		r.DbName = GetCurDatabaseFromContextOrDefault(ctx)
		return ctx, r
	default:
		return ctx, req
	}
}
