package model

type Partition struct {
	PartitionID   int64
	PartitionName string
	Extra         map[string]string
}
