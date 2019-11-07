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

/*
 - ENVIRONMENT VARIABLES:
 RUN_INTEGRATION_TEST=true
 IT_SCYLLA_HOST=127.0.0.1
 IT_SCYLLA_PORT=9042
 IT_NALEJ_KEYSPACE=testkeyspace

 - commands to execute:
 docker run --name scylla -p 9042:9042 -d scylladb/scylla
 docker exec -it scylla cqlsh
 create KEYSPACE testkeyspace WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
 use testkeyspace;
 create table testkeyspace.tableTest (id1 text, id2 text, id3 text, primary key (id1, id2));
 create table testkeyspace.basicTableTest (id1 text, id2 text, id3 text, primary key (id1));
*/
package scylladb

import (
	"github.com/google/uuid"
	"github.com/nalej/scylladb-utils/pkg/utils"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
)

var _ = ginkgo.Describe("Scylla cluster provider", func() {

	if !utils.RunIntegrationTests() {
		log.Warn().Msg("Integration tests are skipped")
		return
	}

	var scyllaHost = os.Getenv("IT_SCYLLA_HOST")
	if scyllaHost == "" {
		ginkgo.Fail("missing environment variables")
	}
	var nalejKeySpace = os.Getenv("IT_NALEJ_KEYSPACE")
	if scyllaHost == "" {
		ginkgo.Fail("missing environment variables")
	}
	scyllaPort, _ := strconv.Atoi(os.Getenv("IT_SCYLLA_PORT"))
	if scyllaPort <= 0 {
		ginkgo.Fail("missing environment variables")
	}

	// create a provider and connect it
	sp := NewScyllaDBProvider(scyllaHost, scyllaPort, nalejKeySpace)

	ginkgo.BeforeSuite(func() {
		cError := sp.Connect()
		gomega.Expect(cError).To(gomega.Succeed())
	})

	ginkgo.AfterSuite(func() {
		sp.UnsafeClear([]string{Table, BasicTable})
		sp.Disconnect()
	})

	ginkgo.Context("Simple Test", func() {
		ginkgo.It("Should be able to add a register", func() {
			compo := GetCompositeStruct()
			pk, val := GetValues(*compo)

			err := sp.UnsafeAdd(BasicTable, pk, val, AllTableColumns, compo)
			gomega.Expect(err).To(gomega.Succeed())

		})
		ginkgo.It("Should not be able to add a register twice", func() {
			compo := GetCompositeStruct()

			pk, val := GetValues(*compo)
			err := sp.UnsafeAdd(BasicTable, pk, val, AllTableColumns, compo)
			gomega.Expect(err).To(gomega.Succeed())

			err = sp.UnsafeAdd(BasicTable, pk, val, AllTableColumns, compo)
			gomega.Expect(err).NotTo(gomega.Succeed())

		})
		ginkgo.It("Should be able to get a register", func() {
			compo := GetCompositeStruct()
			pk, val := GetValues(*compo)

			err := sp.UnsafeAdd(BasicTable, pk, val, AllTableColumns, compo)
			gomega.Expect(err).To(gomega.Succeed())

			var retrieved interface{} = &CompositeStruct{}

			err = sp.UnsafeGet(BasicTable, pk, val, AllTableColumns, &retrieved)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(retrieved).Should(gomega.Equal(compo))

		})
		ginkgo.It("should not be able to get a non exists register", func() {
			compo := GetCompositeStruct()
			pk, val := GetValues(*compo)

			var retrieved interface{} = &CompositeStruct{}
			err := sp.UnsafeGet(BasicTable, pk, val, AllTableColumns, &retrieved)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should be able to update a register", func() {

			compo := GetCompositeStruct()
			pk, val := GetValues(*compo)

			err := sp.UnsafeAdd(BasicTable, pk, val, AllTableColumns, compo)
			gomega.Expect(err).To(gomega.Succeed())

			compo.Id2 = uuid.New().String()
			compo.Id3 = uuid.New().String()
			err = sp.UnsafeUpdate(BasicTable, pk, val, AllTableColumnsNoPK, compo)
			gomega.Expect(err).To(gomega.Succeed())

			// check if the update works
			var retrieved interface{} = &CompositeStruct{}
			err = sp.UnsafeGet(BasicTable, pk, val, AllTableColumns, &retrieved)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(retrieved).Should(gomega.Equal(compo))
		})
		ginkgo.It("should not be able to update a non exists register", func() {
			compo := GetCompositeStruct()
			pk, val := GetValues(*compo)

			err := sp.UnsafeUpdate(BasicTable, pk, val, AllTableColumnsNoPK, compo)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should be able to delete a register", func() {
			compo := GetCompositeStruct()
			pk, val := GetValues(*compo)

			err := sp.UnsafeAdd(BasicTable, pk, val, AllTableColumns, compo)
			gomega.Expect(err).To(gomega.Succeed())

			err = sp.UnsafeRemove(BasicTable, pk, val)
			gomega.Expect(err).To(gomega.Succeed())
		})
		ginkgo.It("should not be able to delete a non exists register", func() {
			compo := GetCompositeStruct()
			pk, val := GetValues(*compo)

			err := sp.UnsafeRemove(BasicTable, pk, val)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})

	})

	ginkgo.Context("Composite tests", func() {
		ginkgo.It("Should be able to add a register in composite table", func() {
			compo := GetCompositeStruct()

			err := sp.UnsafeCompositeAdd(Table, GetCompositeValues(*compo), AllTableColumns, compo)
			gomega.Expect(err).To(gomega.Succeed())

		})
		ginkgo.It("Should not be able to add a register twice", func() {

			compo := GetCompositeStruct()

			err := sp.UnsafeCompositeAdd(Table, GetCompositeValues(*compo), AllTableColumns, compo)
			gomega.Expect(err).To(gomega.Succeed())

			err = sp.UnsafeCompositeAdd(Table, GetCompositeValues(*compo), AllTableColumns, compo)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should be able to update a register", func() {

			compo := GetCompositeStruct()

			err := sp.UnsafeCompositeAdd(Table, GetCompositeValues(*compo), AllTableColumns, compo)
			gomega.Expect(err).To(gomega.Succeed())

			compo.Id3 = uuid.New().String()
			err = sp.UnsafeCompositeUpdate(Table, GetCompositeValues(*compo), AllCompositeTableColumnsNoPK, compo)
			gomega.Expect(err).To(gomega.Succeed())

			// check if the update works
			var retrieved interface{} = &CompositeStruct{}
			err = sp.UnsafeCompositeGet(Table, GetCompositeValues(*compo), AllTableColumns, &retrieved)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(retrieved).Should(gomega.Equal(compo))
		})
		ginkgo.It("should not be able to update a non exists register", func() {
			compo := GetCompositeStruct()
			err := sp.UnsafeCompositeUpdate(Table, GetCompositeValues(*compo), AllCompositeTableColumnsNoPK, compo)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should be able to get a register", func() {
			compo := GetCompositeStruct()

			err := sp.UnsafeCompositeAdd(Table, GetCompositeValues(*compo), AllTableColumns, compo)
			gomega.Expect(err).To(gomega.Succeed())

			var retrieved interface{} = &CompositeStruct{}

			err = sp.UnsafeCompositeGet(Table, GetCompositeValues(*compo), AllTableColumns, &retrieved)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(retrieved).Should(gomega.Equal(compo))

		})
		ginkgo.It("should not be able to get a non exists register", func() {
			compo := GetCompositeStruct()
			var retrieved interface{} = &CompositeStruct{}
			err := sp.UnsafeCompositeGet(Table, GetCompositeValues(*compo), AllTableColumns, &retrieved)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should be able to delete a register", func() {
			compo := GetCompositeStruct()
			err := sp.UnsafeCompositeAdd(Table, GetCompositeValues(*compo), AllTableColumns, compo)
			gomega.Expect(err).To(gomega.Succeed())

			err = sp.UnsafeCompositeRemove(Table, GetCompositeValues(*compo))
			gomega.Expect(err).To(gomega.Succeed())
		})
		ginkgo.It("should not be able to delete a non exists register", func() {
			compo := GetCompositeStruct()
			err := sp.UnsafeCompositeRemove(Table, GetCompositeValues(*compo))
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
	})
})
