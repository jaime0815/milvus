// Licensed to the LF AI & Data foundation under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/milvus-io/milvus/internal/log"

	"go.uber.org/zap"
)

const (
	// RequestTimeout is default timeout for mysql operation
	RequestTimeout = 10 * time.Second
)

// Mysql implements DB interface
type Mysql struct {
	pool *sql.DB
}

// NewDBPool creates connections to a database
func NewDBPool(db *sql.DB) *Mysql {
	mysql := &Mysql{
		pool: db,
	}
	return mysql
}

// Close closes the connection to mysql
func (kv *Mysql) Close() {
	if kv.pool != nil {
		kv.pool.Close()
	}
}

// Query the database for the information requested and prints the results.
// If the query fails exit the program with an error.
func (kv *Mysql) Query(ctx context.Context, id int64) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var name string
	err := kv.pool.QueryRowContext(ctx, "select p.name from people as p where p.id = :id;", sql.Named("id", id)).Scan(&name)
	if err != nil {
		log.Fatal("unable to execute search query", zap.Error(err))
	}
	log.Info("query result", zap.String("name", name))
}

// Insert a db record
func (kv *Mysql) Insert(key, value string) {
	projects := []struct {
		mascot  string
		release int
	}{
		{"tux", 1991},
		{"duke", 1996},
		{"gopher", 2009},
		{"moby dock", 2013},
	}

	stmt, err := kv.pool.Prepare("INSERT INTO projects(id, mascot, release, category) VALUES( ?, ?, ?, ? )")
	if err != nil {
		log.Fatal("unable to create a prepared statement", zap.Error(err))
	}
	defer stmt.Close() // Prepared statements take up server resources and should be closed after use.

	for id, project := range projects {
		if _, err := stmt.Exec(id+1, project.mascot, project.release, "open source"); err != nil {
			log.Fatal("unable to execute insert statements", zap.Error(err))
		}
	}
}
