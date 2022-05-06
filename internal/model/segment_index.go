package model

type SegmentIndex struct {
	CollectionID int64
	PartitionID  int64
	SegmentID    int64
	FieldID      int64
	IndexID      int64
	BuildID      int64
	EnableIndex  bool
	Extra        map[string]string
}
