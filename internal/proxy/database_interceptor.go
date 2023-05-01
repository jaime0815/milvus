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
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.DropCollectionRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.HasCollectionRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.LoadCollectionRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.ReleaseCollectionRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.DescribeCollectionRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetStatisticsRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetCollectionStatisticsRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.ShowCollectionsRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.AlterCollectionRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.CreatePartitionRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.DropPartitionRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.HasPartitionRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.LoadPartitionsRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.ReleasePartitionsRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetPartitionStatisticsRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.ShowPartitionsRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetLoadingProgressRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetLoadStateRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.CreateIndexRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.DescribeIndexRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.DropIndexRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetIndexBuildProgressRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetIndexStateRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.InsertRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.DeleteRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.SearchRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.FlushRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.QueryRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.CreateAliasRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.DropAliasRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.AlterAliasRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.CalcDistanceRequest:
		return ctx, r
	case *milvuspb.FlushAllRequest:
		return ctx, r
	case *milvuspb.GetPersistentSegmentInfoRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetQuerySegmentInfoRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.DummyRequest:
		return ctx, r
	case *milvuspb.GetMetricsRequest:
		return ctx, r
	case *milvuspb.LoadBalanceRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetReplicasRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetCompactionStateRequest:
		return ctx, r
	case *milvuspb.ManualCompactionRequest:
		return ctx, r
	case *milvuspb.GetCompactionPlansRequest:
		return ctx, r
	case *milvuspb.GetFlushStateRequest:
		return ctx, r
	case *milvuspb.GetFlushAllStateRequest:
		return ctx, r
	case *milvuspb.ImportRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.GetImportStateRequest:
		return ctx, r
	case *milvuspb.ListImportTasksRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.RenameCollectionRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	case *milvuspb.TransferReplicaRequest:
		r.DbName = GetCurDBNameFromContextOrDefault(ctx)
		return ctx, r
	default:
		return ctx, req
	}
}
