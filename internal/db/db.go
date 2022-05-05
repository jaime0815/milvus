package db

// DB contains base operations of rds
type DB interface {
	Query(key string) (string, error)
	Insert(key, value string) error
	//MultiInsert(kvs map[string]string) error
	//Delete(key string) error
	//MultiDelete(keys []string) error
	//Update(key, value string) error
	//MultiUpdate(kvs map[string]string) error

	Close()
}
