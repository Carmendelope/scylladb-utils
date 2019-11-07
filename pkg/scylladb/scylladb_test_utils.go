/*
 * Copyright 2019 Nalej
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package scylladb

import (
	"github.com/google/uuid"
)

type CompositeStruct struct {
	Id1 string `json:"id1,omitempty" cql:"id1"`
	Id2 string `json:"id2,omitempty" cql:"id2"`
	Id3 string `json:"id3,omitempty" cql:"id3"`
}

func NewScyllaDBProvider(address string, port int, keyspace string) *ScyllaDB {
	provider := ScyllaDB{Address: address, Port: port, Keyspace: keyspace, Session: nil}
	provider.Connect()
	return &provider
}

var AllTableColumns = []string{"id1", "id2", "id3"}
var AllCompositeTableColumnsNoPK = []string{"id3"}
var AllTableColumnsNoPK = []string{"id2", "id3"}

const Table = "tabletest"
const BasicTable = "basictabletest"

func GetCompositeValues(composite CompositeStruct) map[string]interface{} {
	return map[string]interface{}{"id1": composite.Id1, "id2": composite.Id2}
}
func GetValues(composite CompositeStruct) (string, string) {
	return "id1", composite.Id1
}

func GetCompositeStruct() *CompositeStruct {

	return &CompositeStruct{
		Id1: uuid.New().String(),
		Id2: uuid.New().String(),
		Id3: uuid.New().String(),
	}
}
