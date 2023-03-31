package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/milvus-io/milvus/pkg/log"
	"github.com/milvus-io/milvus/pkg/util/paramtable"
	"github.com/milvus-io/milvus/pkg/util/typeutil"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

type DocContent struct {
	key          string
	defaultValue string
	sinceVersion string
	refreshable  string
	exportToUser bool
	comment      string
}

func collect() []DocContent {
	params := &paramtable.ComponentParam{}
	params.Init()

	val := reflect.ValueOf(params).Elem()
	data := make([]DocContent, 0)
	keySet := typeutil.NewSet[string]()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		collectRecursive(params, &data, &valueField)
	}
	result := make([]DocContent, 0)
	for _, d := range data {
		if keySet.Contain(d.key) {
			continue
		}
		keySet.Insert(d.key)
		result = append(result, d)
	}
	return result
}

func collectRecursive(params *paramtable.ComponentParam, data *[]DocContent, val *reflect.Value) {
	if val.Kind() != reflect.Struct {
		return
	}
	log.Debug("enter", zap.Any("variable", val.String()))
	for j := 0; j < val.NumField(); j++ {
		subVal := val.Field(j)
		tag := val.Type().Field(j).Tag
		t := val.Type().Field(j).Type.String()
		if t == "paramtable.ParamItem" {
			item := subVal.Interface().(paramtable.ParamItem)
			refreshable := tag.Get("refreshable")
			defaultValue := params.GetWithDefault(item.Key, item.DefaultValue)
			log.Debug("got key", zap.String("key", item.Key), zap.Any("value", defaultValue), zap.String("variable", val.Type().Field(j).Name))
			*data = append(*data, DocContent{item.Key, defaultValue, item.Version, refreshable, item.Export, item.Doc})
			for _, fk := range item.FallbackKeys {
				log.Debug("got fallback key", zap.String("key", fk), zap.Any("value", defaultValue), zap.String("variable", val.Type().Field(j).Name))
				*data = append(*data, DocContent{fk, defaultValue, item.Version, refreshable, item.Export, item.Doc})
			}
		} else if t == "paramtable.ParamGroup" {
			item := subVal.Interface().(paramtable.ParamGroup)
			log.Debug("got key", zap.String("key", item.KeyPrefix), zap.String("variable", val.Type().Field(j).Name))
			refreshable := tag.Get("refreshable")
			*data = append(*data, DocContent{item.KeyPrefix, "", item.Version, refreshable, item.Export, item.Doc})
		} else {
			collectRecursive(params, data, &subVal)
		}
	}
}

func WriteCsv() {
	f, err := os.Create("configs.csv")
	defer f.Close()
	if err != nil {
		log.Error("create file failed", zap.Error(err))
		os.Exit(-2)
	}
	w := csv.NewWriter(f)
	w.Write([]string{"key", "defaultValue", "sinceVersion", "refreshable", "exportToUser", "comment"})

	result := collect()
	w.WriteAll(lo.Map(result, func(d DocContent, _ int) []string {
		return []string{d.key, d.defaultValue, d.sinceVersion, d.refreshable, fmt.Sprintf("%t", d.exportToUser), d.comment}
	}))
	w.Flush()
}

type YamlGroup struct {
	name    string
	header  string
	disable bool
}

type YamlMarshaller struct {
	writer *os.File
	groups []YamlGroup
	data   []DocContent
}

func (m *YamlMarshaller) writeYamlRecursive(data []DocContent, level int) {
	var topLevels = typeutil.NewOrderedMap[string, []DocContent]()
	for _, d := range data {
		key := strings.Split(d.key, ".")[level]

		old, ok := topLevels.Get(key)
		if !ok {
			topLevels.Set(key, []DocContent{d})
		} else {
			topLevels.Set(key, append(old, d))
		}
	}

	var keys []string
	var extraHeaders map[string]string
	disabledGroups := lo.Map(
		lo.Filter(
			m.groups,
			func(g YamlGroup, _ int) bool { return g.disable }),
		func(g YamlGroup, _ int) string { return g.name })
	if level == 0 {
		keys = lo.Map(m.groups, func(g YamlGroup, _ int) string { return g.name })
		extraHeaders = lo.SliceToMap(m.groups, func(g YamlGroup) (string, string) { return g.name, g.header })
	} else {
		keys = topLevels.Keys()
	}
	for _, key := range keys {
		contents, ok := topLevels.Get(key)
		if !ok {
			log.Debug("didnot found config for " + key)
			continue
		}
		content := contents[0]
		isDisabled := slices.Contains(disabledGroups, strings.Split(content.key, ".")[0])
		if strings.Count(content.key, ".") == level {
			if isDisabled {
				m.writer.WriteString("# ")
			}
			m.writeContent(key, content.defaultValue, content.comment, level)
			continue
		}
		extra, ok := extraHeaders[key]
		if ok {
			m.writer.WriteString(extra + "\n")
		}
		if isDisabled {
			m.writer.WriteString("# ")
		}
		m.writer.WriteString(fmt.Sprintf("%s%s:\n", strings.Repeat(" ", level*2), key))
		m.writeYamlRecursive(contents, level+1)
	}
}

func (m *YamlMarshaller) writeContent(key, value, comment string, level int) {
	if strings.Contains(comment, "\n") {
		multilines := strings.Split(comment, "\n")
		for _, line := range multilines {
			m.writer.WriteString(fmt.Sprintf("%s# %s\n", strings.Repeat(" ", level*2), line))
		}
		m.writer.WriteString(fmt.Sprintf("%s%s: %s\n", strings.Repeat(" ", level*2), key, value))
	} else if comment != "" {
		m.writer.WriteString(fmt.Sprintf("%s%s: %s # %s\n", strings.Repeat(" ", level*2), key, value, comment))
	} else {
		m.writer.WriteString(fmt.Sprintf("%s%s: %s\n", strings.Repeat(" ", level*2), key, value))
	}
}

func WriteYaml() {
	f, err := os.Create("milvus.yaml")
	defer f.Close()
	if err != nil {
		log.Error("create file failed", zap.Error(err))
		os.Exit(-2)
	}

	result := collect()

	f.WriteString(`# Licensed to the LF AI & Data foundation under one
# or more contributor license agreements. See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership. The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License. You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
`)
	groups := []YamlGroup{
		{
			name:   "etcd",
			header: "\n# Related configuration of etcd, used to store Milvus metadata & service discovery.",
		},
		{
			name: "metastore",
		},
		{
			name:   "mysql",
			header: "\n# Related configuration of mysql, used to store Milvus metadata.",
		},
		{
			name: "localStorage",
		},
		{
			name: "minio",
			header: `
# Related configuration of MinIO/S3/GCS or any other service supports S3 API, which is responsible for data persistence for Milvus.
# We refer to the storage service as MinIO/S3 in the following description for simplicity.`,
		},
		{
			name: "pulsar",
			header: `
# Milvus supports three MQ: rocksmq(based on RockDB), Pulsar and Kafka, which should be reserved in config what you use.
# There is a note about enabling priority if we config multiple mq in this file
# 1. standalone(local) mode: rocksmq(default) > Pulsar > Kafka
# 2. cluster mode:  Pulsar(default) > Kafka (rocksmq is unsupported)

# Related configuration of pulsar, used to manage Milvus logs of recent mutation operations, output streaming log, and provide log publish-subscribe services.`,
		},
		{
			name:    "kafka",
			header:  "\n# If you want to enable kafka, needs to comment the pulsar configs",
			disable: true,
		},
		{
			name: "rocksmq",
		},
		{
			name:   "rootCoord",
			header: "\n# Related configuration of rootCoord, used to handle data definition language (DDL) and data control language (DCL) requests",
		},
		{
			name:   "proxy",
			header: "\n# Related configuration of proxy, used to validate client requests and reduce the returned results.",
		},
		{
			name:   "queryCoord",
			header: "\n# Related configuration of queryCoord, used to manage topology and load balancing for the query nodes, and handoff from growing segments to sealed segments.",
		},
		{
			name:   "queryNode",
			header: "\n# Related configuration of queryNode, used to run hybrid search between vector and scalar data.",
		},
		{
			name: "indexCoord",
		},
		{
			name: "indexNode",
		},
		{
			name: "dataCoord",
		},
		{
			name: "dataNode",
		},
		{
			name:   "log",
			header: "\n# Configures the system log output.",
		},
		{
			name: "grpc",
		},
		{
			name:   "tls",
			header: "\n# Configure the proxy tls enable.",
		},
		{
			name: "common",
		},
		{
			name: "quotaAndLimits",
			header: `
# QuotaConfig, configurations of Milvus quota and limits.
# By default, we enable:
#   1. TT protection;
#   2. Memory protection.
#   3. Disk quota protection.
# You can enable:
#   1. DML throughput limitation;
#   2. DDL, DQL qps/rps limitation;
#   3. DQL Queue length/latency protection;
#   4. DQL result rate protection;
# If necessary, you can also manually force to deny RW requests.`,
		},
		{
			name: "trace",
		},
	}
	marshller := YamlMarshaller{f, groups, result}
	marshller.writeYamlRecursive(lo.Filter(result, func(d DocContent, _ int) bool {
		return d.exportToUser
	}), 0)
}
