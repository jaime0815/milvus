package mck

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/milvus-io/milvus/internal/util/logutil"

	miniokv "github.com/milvus-io/milvus/internal/kv/minio"

	"github.com/milvus-io/milvus/internal/util/paramtable"

	pb "github.com/milvus-io/milvus/internal/proto/etcdpb"

	"github.com/milvus-io/milvus/internal/proto/commonpb"
	"github.com/milvus-io/milvus/internal/proto/querypb"

	"github.com/golang/protobuf/proto"
	etcdkv "github.com/milvus-io/milvus/internal/kv/etcd"
	"github.com/milvus-io/milvus/internal/log"
	"github.com/milvus-io/milvus/internal/proto/datapb"
	"github.com/milvus-io/milvus/internal/util/etcd"
	"go.uber.org/zap"
)

const (
	MckCmd            = "mck"
	segmentPrefix     = "datacoord-meta/s"
	collectionPrefix  = "snapshots/root-coord/collection"
	triggerTaskPrefix = "queryCoord-triggerTask"
	activeTaskPrefix  = "queryCoord-activeTask"
	taskInfoPrefix    = "queryCoord-taskInfo"
	MckTrash          = "mck-trash"
)

var (
	Params              paramtable.GrpcServerConfig
	taskKeyMap          = make(map[int64]string)
	taskNameMap         = make(map[int64]string)
	allTaskInfo         = make(map[string]string)
	partitionIdToTasks  = make(map[int64][]int64)
	segmentIdToTasks    = make(map[int64][]int64)
	taskIdToInvalidPath = make(map[int64][]string)
	segmentIds          = make(map[int64]*datapb.SegmentInfo)
	partitionIds        = make(map[int64]interface{})
	etcdKV              *etcdkv.EtcdKV
	minioKV             *miniokv.MinIOKV
)

func RunMck(args []string) {
	//etcdCli, err := etcd.GetRemoteEtcdClient([]string{*etcdAddr})
	logutil.SetupLogger(&log.Config{
		Level: "info",
		File: log.FileLogConfig{
			Filename: fmt.Sprintf("mck-%s.log", time.Now().Format("20060102150405.99")),
		},
	})

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	etcdIP := flags.String("etcdIp", "", "Etcd endpoint to connect")
	ectdRootPath := flags.String("etcdRootPath", "", "Etcd root path")
	minioAddress := flags.String("minioAddress", "", "Minio endpoint to connect")
	minioUsername := flags.String("minioUsername", "", "Minio username")
	minioPassword := flags.String("minioPassword", "", "Minio password")
	minioUseSSL := flags.String("minioUseSSL", "", "Minio to use ssl")
	minioBucketName := flags.String("minioBucketName", "", "Minio bucket name")

	if err := flags.Parse(os.Args[2:]); err != nil {
		log.Fatal("failed to parse flags", zap.Error(err))
	}
	log.Info("args", zap.Strings("args", args))

	Params.Init()
	var etcdCli *clientv3.Client
	var err error
	if *etcdIP != "" {
		etcdCli, err = etcd.GetRemoteEtcdClient([]string{*etcdIP})
	} else {
		etcdCli, err = etcd.GetEtcdClient(&Params.EtcdCfg)
	}
	if err != nil {
		log.Fatal("failed to connect to etcd", zap.Error(err))
	}

	rootPath := getNotEmptyString(*ectdRootPath, Params.EtcdCfg.MetaRootPath)
	etcdKV = etcdkv.NewEtcdKV(etcdCli, rootPath)
	log.Info("Etcd root path", zap.String("root_path", rootPath))

	if len(args) >= 3 && args[2] == "clean" {
		cleanTrash()
		return
	}

	useSSL := Params.MinioCfg.UseSSL
	if *minioUseSSL == "true" || *minioUseSSL == "false" {
		useSSL, _ = strconv.ParseBool(*minioUseSSL)
	}
	option := &miniokv.Option{
		Address:           getNotEmptyString(*minioAddress, Params.MinioCfg.Address),
		AccessKeyID:       getNotEmptyString(*minioUsername, Params.MinioCfg.AccessKeyID),
		SecretAccessKeyID: getNotEmptyString(*minioPassword, Params.MinioCfg.SecretAccessKey),
		UseSSL:            useSSL,
		BucketName:        getNotEmptyString(*minioBucketName, Params.MinioCfg.BucketName),
		CreateBucket:      true,
	}

	minioKV, err = miniokv.NewMinIOKV(context.Background(), option)
	if err != nil {
		log.Fatal("failed to connect to etcd", zap.Error(err))
	}

	keys, values, err := etcdKV.LoadWithPrefix(segmentPrefix)
	if err != nil {
		log.Fatal("failed to list the segment info", zap.Error(err))
	}
	for _, value := range values {
		info := &datapb.SegmentInfo{}
		err = proto.Unmarshal([]byte(value), info)
		if err != nil {
			continue
		}
		segmentIds[info.ID] = info
	}

	keys, values, err = etcdKV.LoadWithPrefix(collectionPrefix)
	if err != nil {
		log.Fatal("failed to list the collection info", zap.Error(err))
	}
	for _, value := range values {
		collInfo := pb.CollectionInfo{}
		err = proto.Unmarshal([]byte(value), &collInfo)
		if err != nil {
			continue
		}
		for _, id := range collInfo.PartitionIDs {
			partitionIds[id] = nil
		}
	}

	// log Segment ids and partition ids
	var ids []int64
	for id := range segmentIds {
		ids = append(ids, id)
	}
	log.Info("segment ids", zap.Int64s("ids", ids))

	ids = []int64{}
	for id := range partitionIds {
		ids = append(ids, id)
	}
	log.Info("partition ids", zap.Int64s("ids", ids))

	keys, values, err = etcdKV.LoadWithPrefix(triggerTaskPrefix)
	if err != nil {
		log.Fatal("failed to list the trigger task info", zap.Error(err))
	}
	extractTask(triggerTaskPrefix, keys, values)

	keys, values, err = etcdKV.LoadWithPrefix(activeTaskPrefix)
	if err != nil {
		log.Fatal("failed to list the active task info", zap.Error(err))
	}
	extractTask(activeTaskPrefix, keys, values)

	// log all tasks
	if len(taskNameMap) > 0 {
		log.Info("all tasks")
		for taskId, taskName := range taskNameMap {
			log.Info("task info", zap.String("name", taskName), zap.Int64("id", taskId))
		}
	}

	// collect invalid tasks related with invalid partitions
	var invalidTasksOfPartition []int64
	//var invalidPartition []int64
	var buffer bytes.Buffer
	for id, tasks := range partitionIdToTasks {
		if _, ok := partitionIds[id]; !ok {
			invalidTasksOfPartition = append(invalidTasksOfPartition, tasks...)
			//invalidPartition = append(invalidPartition, id)
			buffer.WriteString(fmt.Sprintf("Partition ID: %d\n", id))
			buffer.WriteString(fmt.Sprintf("Tasks: %v\n", tasks))
		}
	}
	invalidTasksOfPartition = removeRepeatElement(invalidTasksOfPartition)
	if len(invalidTasksOfPartition) > 0 {
		line()
		fmt.Println("Invalid Partitions")
		fmt.Println(buffer.String())
	}

	// collect invalid tasks related with invalid segments
	var invalidTasksOfSegment []int64
	//var invalidSegment []int64
	buffer.Reset()
	if len(segmentIdToTasks) > 0 {
		for id, tasks := range segmentIdToTasks {
			if _, ok := segmentIds[id]; !ok {
				invalidTasksOfSegment = append(invalidTasksOfSegment, tasks...)
				//invalidSegment = append(invalidSegment, id)
				buffer.WriteString(fmt.Sprintf("Segment ID: %d\n", id))
				buffer.WriteString(fmt.Sprintf("Tasks: %v\n", tasks))
			}
		}
		invalidTasksOfSegment = removeRepeatElement(invalidTasksOfSegment)
	}
	if len(invalidTasksOfSegment) > 0 {
		line()
		fmt.Println("Invalid Segments")
		fmt.Println(buffer.String())
	}

	// collect invalid tasks related with incorrect file paths in minio
	var invalidTasksOfPath []int64
	if len(taskIdToInvalidPath) > 0 {
		line()
		fmt.Println("Invalid Paths")
		for id, paths := range taskIdToInvalidPath {
			fmt.Printf("Task ID: %d\n", id)
			for _, path := range paths {
				fmt.Printf("\t%s\n", path)
			}
			invalidTasksOfPath = append(invalidTasksOfPath, id)
		}
		invalidTasksOfPath = removeRepeatElement(invalidTasksOfPath)
	}

	// combine all invalid tasks
	var invalidTasks []int64
	invalidTasks = append(invalidTasks, invalidTasksOfPartition...)
	invalidTasks = append(invalidTasks, invalidTasksOfSegment...)
	invalidTasks = append(invalidTasks, invalidTasksOfPath...)
	invalidTasks = removeRepeatElement(invalidTasks)
	if len(invalidTasks) > 0 {
		line()
		fmt.Println("All invalid tasks")
		for _, invalidTask := range invalidTasks {
			line2()
			fmt.Printf("TaskID: %d\t%s\n", invalidTask, taskNameMap[invalidTask])
		}
	}

	if len(invalidTasks) > 0 {
		fmt.Print("Delete all invalid tasks, [Y/n]:")
		deleteAll := ""
		fmt.Scanln(&deleteAll)
		if deleteAll == "Y" {
			for _, invalidTask := range invalidTasks {
				if !removeTask(invalidTask) {
					continue
				}
				fmt.Println("Delete task id: ", invalidTask)
			}
		} else {
			deleteTask := ""
			idMap := int64Map(invalidTasks)
			fmt.Println("Enter 'Exit' to end")
			for len(idMap) > 0 {
				fmt.Print("Enter a invalid task id to delete, [Exit]:")
				fmt.Scanln(&deleteTask)
				if deleteTask == "Exit" {
					return
				}
				invalidTask, err := strconv.ParseInt(deleteTask, 10, 64)
				if err != nil {
					fmt.Println("Invalid task id.")
					continue
				}
				if _, ok := idMap[invalidTask]; !ok {
					fmt.Println("This is not a invalid task id.")
					continue
				}
				if !removeTask(invalidTask) {
					continue
				}
				delete(idMap, invalidTask)
				fmt.Println("Delete task id: ", invalidTask)
			}
		}
	}
}

func getNotEmptyString(a string, b string) string {
	if a != "" {
		return a
	} else {
		return b
	}
}

func cleanTrash() {
	keys, _, err := etcdKV.LoadWithPrefix(MckTrash)
	if err != nil {
		log.Error("failed to load backup info", zap.Error(err))
		return
	}
	if len(keys) == 0 {
		fmt.Println("Empty backup task info")
		return
	}
	fmt.Println(strings.Join(keys, "\n"))
	fmt.Print("Delete all backup infos, [Y/n]:")
	deleteAll := ""
	fmt.Scanln(&deleteAll)
	if deleteAll == "Y" {
		err = etcdKV.RemoveWithPrefix(MckTrash)
		if err != nil {
			log.Error("failed to remove backup infos", zap.Error(err))
			return
		}
	}
}

func redPrint(msg string) {
	fmt.Printf("\033[0;31;40m%s\033[0m", msg)
}

func line() {
	fmt.Println("================================================================================")
}

func line2() {
	fmt.Println("--------------------------------------------------------------------------------")
}

func getTrashKey(taskType, key string) string {
	return fmt.Sprintf("%s/%s/%s", MckTrash, taskType, key)
}

func extractTask(prefix string, keys []string, values []string) {

	for i := range keys {
		taskID, err := strconv.ParseInt(filepath.Base(keys[i]), 10, 64)
		if err != nil {
			log.Warn("failed to parse int", zap.String("tasks", filepath.Base(keys[i])))
			continue
		}

		taskName, pids, sids, err := unmarshalTask(taskID, values[i])
		if err != nil {
			log.Warn("failed to unmarshal task", zap.Int64("task_id", taskID))
			continue
		}
		for _, pid := range pids {
			partitionIdToTasks[pid] = append(partitionIdToTasks[pid], taskID)
		}
		for _, sid := range sids {
			segmentIdToTasks[sid] = append(segmentIdToTasks[sid], taskID)
		}

		taskKeyMap[taskID] = fmt.Sprintf("%s/%d", prefix, taskID)
		allTaskInfo[taskKeyMap[taskID]] = values[i]
		taskNameMap[taskID] = taskName
	}
}

func removeTask(invalidTask int64) bool {
	taskType := taskNameMap[invalidTask]
	key := taskKeyMap[invalidTask]
	err := etcdKV.Save(getTrashKey(taskType, key), allTaskInfo[key])
	if err != nil {
		log.Warn("failed to backup task", zap.Int64("task_id", invalidTask), zap.Error(err))
		return false
	}
	fmt.Printf("Back up task successfully, back path: %s\n", getTrashKey(taskType, key))
	err = etcdKV.Remove(key)
	if err != nil {
		log.Warn("failed to remove task", zap.Int64("task_id", invalidTask), zap.Error(err))
		return false
	}

	key = fmt.Sprintf("%s/%d", taskInfoPrefix, invalidTask)
	taskInfo, err := etcdKV.Load(key)
	if err != nil {
		log.Warn("failed to load task info", zap.Int64("task_id", invalidTask), zap.Error(err))
		return false
	}
	err = etcdKV.Save(getTrashKey(taskType, key), taskInfo)
	if err != nil {
		log.Warn("failed to backup task info", zap.Int64("task_id", invalidTask), zap.Error(err))
		return false
	}
	fmt.Printf("Back up task info successfully, back path: %s\n", getTrashKey(taskType, key))
	err = etcdKV.Remove(key)
	if err != nil {
		log.Warn("failed to remove task info", zap.Int64("task_id", invalidTask), zap.Error(err))
	}
	return true
}

func emptyInt64() []int64 {
	return []int64{}
}

func errReturn(taskId int64, pbName string, err error) (string, []int64, []int64, error) {
	return "", emptyInt64(), emptyInt64(), fmt.Errorf("task id: %d, failed to unmarshal %s, err %s ", taskId, pbName, err.Error())
}

func int64Map(ids []int64) map[int64]interface{} {
	idMap := make(map[int64]interface{})
	for _, id := range ids {
		idMap[id] = nil
	}
	return idMap
}

func removeRepeatElement(ids []int64) []int64 {
	idMap := int64Map(ids)
	elements := make([]int64, 0, len(idMap))
	for k := range idMap {
		elements = append(elements, k)
	}
	return elements
}

func extractDataSegmentInfos(taskId int64, infos []*datapb.SegmentInfo) ([]int64, []int64) {
	var partitionIds []int64
	var segmentIds []int64
	for _, info := range infos {
		partitionIds = append(partitionIds, info.PartitionID)
		segmentIds = append(segmentIds, info.ID)
		extractFieldBinlog(taskId, info.Binlogs)
		extractFieldBinlog(taskId, info.Statslogs)
		extractFieldBinlog(taskId, info.Deltalogs)
	}

	return partitionIds, segmentIds
}

func extractQuerySegmentInfos(taskId int64, infos []*querypb.SegmentInfo) ([]int64, []int64) {
	var partitionIds []int64
	var segmentIds []int64
	for _, info := range infos {
		partitionIds = append(partitionIds, info.PartitionID)
		segmentIds = append(segmentIds, info.SegmentID)
		extractVecFieldIndexInfo(taskId, info.IndexInfos)
	}

	return partitionIds, segmentIds
}

func extractVchannelInfo(taskId int64, infos []*datapb.VchannelInfo) ([]int64, []int64) {
	var partitionIds []int64
	var segmentIds []int64
	for _, info := range infos {
		pids, sids := extractDataSegmentInfos(taskId, info.DroppedSegments)
		partitionIds = append(partitionIds, pids...)
		segmentIds = append(segmentIds, sids...)

		pids, sids = extractDataSegmentInfos(taskId, info.FlushedSegments)
		partitionIds = append(partitionIds, pids...)
		segmentIds = append(segmentIds, sids...)

		pids, sids = extractDataSegmentInfos(taskId, info.UnflushedSegments)
		partitionIds = append(partitionIds, pids...)
		segmentIds = append(segmentIds, sids...)
	}
	return partitionIds, segmentIds
}

func extractFieldBinlog(taskId int64, fieldBinlogList []*datapb.FieldBinlog) {
	for _, fieldBinlog := range fieldBinlogList {
		for _, binlog := range fieldBinlog.Binlogs {
			_, err := minioKV.Load(binlog.LogPath)
			if err != nil {
				taskIdToInvalidPath[taskId] = append(taskIdToInvalidPath[taskId], binlog.LogPath)
			}
		}
	}
}

func extractVecFieldIndexInfo(taskId int64, infos []*querypb.VecFieldIndexInfo) {
	for _, info := range infos {
		for _, indexPath := range info.IndexFilePaths {
			_, err := minioKV.Load(indexPath)
			if err != nil {
				taskIdToInvalidPath[taskId] = append(taskIdToInvalidPath[taskId], indexPath)
			}
		}
	}
}

// return partitionIds,segmentIds,error
func unmarshalTask(taskId int64, t string) (string, []int64, []int64, error) {
	header := commonpb.MsgHeader{}
	err := proto.Unmarshal([]byte(t), &header)

	if err != nil {
		return errReturn(taskId, "MsgHeader", err)
	}

	switch header.Base.MsgType {
	case commonpb.MsgType_LoadCollection:
		loadReq := querypb.LoadCollectionRequest{}
		err = proto.Unmarshal([]byte(t), &loadReq)
		if err != nil {
			return errReturn(taskId, "LoadCollectionRequest", err)
		}
		log.Info("LoadCollection", zap.String("detail", fmt.Sprintf("+%v", loadReq)))
		return "LoadCollection", emptyInt64(), emptyInt64(), nil
	case commonpb.MsgType_LoadPartitions:
		loadReq := querypb.LoadPartitionsRequest{}
		err = proto.Unmarshal([]byte(t), &loadReq)
		if err != nil {
			return errReturn(taskId, "LoadPartitionsRequest", err)
		}
		log.Info("LoadPartitions", zap.String("detail", fmt.Sprintf("+%v", loadReq)))
		return "LoadPartitions", loadReq.PartitionIDs, emptyInt64(), nil
	case commonpb.MsgType_ReleaseCollection:
		loadReq := querypb.ReleaseCollectionRequest{}
		err = proto.Unmarshal([]byte(t), &loadReq)
		if err != nil {
			return errReturn(taskId, "ReleaseCollectionRequest", err)
		}
		log.Info("ReleaseCollection", zap.String("detail", fmt.Sprintf("+%v", loadReq)))
		return "ReleaseCollection", emptyInt64(), emptyInt64(), nil
	case commonpb.MsgType_ReleasePartitions:
		loadReq := querypb.ReleasePartitionsRequest{}
		err = proto.Unmarshal([]byte(t), &loadReq)
		if err != nil {
			return errReturn(taskId, "ReleasePartitionsRequest", err)
		}
		log.Info("ReleasePartitions", zap.String("detail", fmt.Sprintf("+%v", loadReq)))
		return "ReleasePartitions", loadReq.PartitionIDs, emptyInt64(), nil
	case commonpb.MsgType_LoadSegments:
		loadReq := querypb.LoadSegmentsRequest{}
		err = proto.Unmarshal([]byte(t), &loadReq)
		if err != nil {
			return errReturn(taskId, "LoadSegmentsRequest", err)
		}

		fmt.Printf("LoadSegments, task-id: %d, %+v\n", taskId, loadReq)

		var partitionIds []int64
		var segmentIds []int64
		if loadReq.LoadMeta != nil {
			partitionIds = append(partitionIds, loadReq.LoadMeta.PartitionIDs...)
		}
		for _, info := range loadReq.Infos {
			partitionIds = append(partitionIds, info.PartitionID)
			segmentIds = append(segmentIds, info.SegmentID)
			extractFieldBinlog(taskId, info.BinlogPaths)
			extractFieldBinlog(taskId, info.Statslogs)
			extractFieldBinlog(taskId, info.Deltalogs)
			extractVecFieldIndexInfo(taskId, info.IndexInfos)
		}
		log.Info("LoadSegments", zap.String("detail", fmt.Sprintf("+%v", loadReq)))
		return "LoadSegments", removeRepeatElement(partitionIds), removeRepeatElement(segmentIds), nil
	case commonpb.MsgType_ReleaseSegments:
		loadReq := querypb.ReleaseSegmentsRequest{}
		err = proto.Unmarshal([]byte(t), &loadReq)
		if err != nil {
			return errReturn(taskId, "ReleaseSegmentsRequest", err)
		}
		log.Info("ReleaseSegments", zap.String("detail", fmt.Sprintf("+%v", loadReq)))
		return "ReleaseSegments", loadReq.PartitionIDs, loadReq.SegmentIDs, nil
	case commonpb.MsgType_WatchDmChannels:
		loadReq := querypb.WatchDmChannelsRequest{}
		err = proto.Unmarshal([]byte(t), &loadReq)
		if err != nil {
			return errReturn(taskId, "WatchDmChannelsRequest", err)
		}

		var partitionIds []int64
		var segmentIds []int64
		if loadReq.LoadMeta != nil {
			partitionIds = append(partitionIds, loadReq.LoadMeta.PartitionIDs...)
		}

		pids, sids := extractVchannelInfo(taskId, loadReq.Infos)
		partitionIds = append(partitionIds, pids...)
		segmentIds = append(segmentIds, sids...)
		pids, sids = extractDataSegmentInfos(taskId, loadReq.ExcludeInfos)
		partitionIds = append(partitionIds, pids...)
		segmentIds = append(segmentIds, sids...)
		log.Info("WatchDmChannels", zap.String("detail", fmt.Sprintf("+%v", loadReq)))
		return "WatchDmChannels", removeRepeatElement(partitionIds), removeRepeatElement(segmentIds), nil
	case commonpb.MsgType_WatchDeltaChannels:
		loadReq := querypb.WatchDeltaChannelsRequest{}
		err = proto.Unmarshal([]byte(t), &loadReq)
		if err != nil {
			return errReturn(taskId, "WatchDeltaChannelsRequest", err)
		}
		var partitionIds []int64
		var segmentIds []int64
		if loadReq.LoadMeta != nil {
			partitionIds = append(partitionIds, loadReq.LoadMeta.PartitionIDs...)
		}
		pids, sids := extractVchannelInfo(taskId, loadReq.Infos)
		partitionIds = append(partitionIds, pids...)
		segmentIds = append(segmentIds, sids...)
		log.Info("WatchDeltaChannels", zap.String("detail", fmt.Sprintf("+%v", loadReq)))
		return "WatchDeltaChannels", removeRepeatElement(partitionIds), removeRepeatElement(segmentIds), nil
	case commonpb.MsgType_WatchQueryChannels:
		loadReq := querypb.AddQueryChannelRequest{}
		err = proto.Unmarshal([]byte(t), &loadReq)
		if err != nil {
			return errReturn(taskId, "WatchDeltaChannelsRequest", err)
		}
		pids, sids := extractQuerySegmentInfos(taskId, loadReq.GlobalSealedSegments)
		log.Info("WatchQueryChannels", zap.String("detail", fmt.Sprintf("+%v", loadReq)))
		return "WatchQueryChannels", pids, sids, nil
	case commonpb.MsgType_LoadBalanceSegments:
		loadReq := querypb.LoadBalanceRequest{}
		err = proto.Unmarshal([]byte(t), &loadReq)
		if err != nil {
			return errReturn(taskId, "LoadBalanceRequest", err)
		}
		log.Info("LoadBalanceSegments", zap.String("detail", fmt.Sprintf("+%v", loadReq)))
		return "LoadBalanceSegments", emptyInt64(), loadReq.SealedSegmentIDs, nil
	case commonpb.MsgType_HandoffSegments:
		handoffReq := querypb.HandoffSegmentsRequest{}
		err = proto.Unmarshal([]byte(t), &handoffReq)
		if err != nil {
			return errReturn(taskId, "HandoffSegmentsRequest", err)
		}
		pids, sids := extractQuerySegmentInfos(taskId, handoffReq.SegmentInfos)
		log.Info("HandoffSegments", zap.String("detail", fmt.Sprintf("+%v", handoffReq)))
		return "HandoffSegments", pids, sids, nil
	default:
		err = errors.New("inValid msg type when unMarshal task")
		log.Error("invalid message task", zap.Int("type", int(header.Base.MsgType)), zap.Error(err))
		return "", emptyInt64(), emptyInt64(), err
	}
}
