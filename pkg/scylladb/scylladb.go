/*
 * Copyright (C)  2019 Nalej - All Rights Reserved
 */

package scylladb

import (
	"fmt"
	"github.com/gocql/gocql"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/rs/zerolog/log"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
)

// NOTE: UnsafeList does not exist because we need to know the struct where the data will be loaded
//       to unmarshall the result.

// RowNotFoundMsg corresponds to the error message returned by ScyllaDB if the row is not found.
const RowNotFoundMsg = "not found"

// General purpose structure to be reused to build ScyllaDB providers on top sharing common functionality.
type ScyllaDB struct{
	Address  string
	Port     int
	Keyspace string
	Session  *gocql.Session
}

// Connect to the ScyllaDB .
func (s *ScyllaDB) Connect() derrors.Error{
	// connect to the cluster
	conf := gocql.NewCluster(s.Address)
	conf.Keyspace = s.Keyspace
	conf.Port = s.Port

	session, err := conf.CreateSession()
	if err != nil {
		log.Error().Str("trace", conversions.ToDerror(err).DebugReport()).Msg("unable to connect")
		return derrors.AsError(err, "cannot connect")
	}
	s.Session = session
	return nil
}

// Disconnect from the database
func (s *ScyllaDB) Disconnect() {
	if s.Session != nil {
		s.Session.Close()
		s.Session = nil
	}
}

// CheckConnection checks that the session is created
func (s *ScyllaDB) CheckConnection() derrors.Error {
	if s.Session == nil {
		return derrors.NewGenericError("Session not created")
	}
	return nil
}

// CheckAndConnect checks if the connection is set and tries to reconnect otherwise.
func (s *ScyllaDB) CheckAndConnect() derrors.Error {
	err := s.CheckConnection()
	if err != nil {
		log.Info().Msg("session no created, trying to reconnect...")
		// try to reconnect
		err = s.Connect()
		if err != nil {
			return err
		}
	}
	return nil
}

// ----------------------------------------------------------------
// functions for when the PK is composite of one field
// ----------------------------------------------------------------

// UnsafeGenericExist checks if an element identified by a single primary key exists.
func (s *ScyllaDB) UnsafeGenericExist(table string, pkColumn string, pkValue string) (bool, derrors.Error){
	// check connection
	if err := s.CheckAndConnect(); err != nil {
		return false, err
	}
	var count int

	stmt, names := qb.Select(table).CountAll().Where(qb.Eq(pkColumn)).ToCql()
	q := gocqlx.Query(s.Session.Query(stmt), names).BindMap(qb.M{pkColumn: pkValue})

	err := q.GetRelease(&count)
	if err != nil {
		if err.Error() == RowNotFoundMsg {
			return false, nil
		} else {
			return false, derrors.AsError(err, "cannot determinate if elements exists")
		}
	}

	return count > 0, nil
}

// UnsafeAdd adds a new element to a table identified by a single primary key.
func (s *ScyllaDB) UnsafeAdd(table string, pkColumn string, pkValue string, tableColumnNames []string, toAdd interface{}) derrors.Error{
	// check connection
	if err := s.CheckAndConnect(); err != nil {
		return err
	}
	exists, err := s.UnsafeGenericExist(table, pkColumn, pkValue)
	if err != nil {
		return err
	}
	if exists {
		return derrors.NewAlreadyExistsError(pkValue)
	}

	// insert the instance
	stmt, names := qb.Insert(table).Columns(tableColumnNames...).ToCql()
	q := gocqlx.Query(s.Session.Query(stmt), names).BindStruct(toAdd)
	cqlErr := q.ExecRelease()

	if cqlErr != nil {
		log.Warn().Str("err", cqlErr.Error()).Msg("error adding the element")
		return derrors.AsError(cqlErr, "cannot add new element")
	}

	return nil
}

// UnsafeUpdate updates an element in a table identified by a single primary key.
func (s *ScyllaDB) UnsafeUpdate(table string, pkColumn string, pkValue string, tableColumnNames []string, toUpdate interface{}) derrors.Error{
	// check connection
	if err := s.CheckAndConnect(); err != nil {
		return err
	}
	exists, err := s.UnsafeGenericExist(table, pkColumn, pkValue)
	if err != nil {
		return err
	}
	if !exists {
		return derrors.NewNotFoundError(pkValue)
	}

	// update the instance
	stmt, names := qb.Update(table).Set(tableColumnNames...).Where(qb.Eq(pkColumn)).ToCql()
	q := gocqlx.Query(s.Session.Query(stmt), names).BindStruct(toUpdate)
	cqlErr := q.ExecRelease()

	if cqlErr != nil {
		return derrors.AsError(cqlErr, "cannot update element")
	}

	return nil
}

// UnsafeGet retrieves an element from a table identified by a single primary key.
func (s *ScyllaDB) UnsafeGet(table string, pkColumn string, pkValue string, tableColumnNames []string, result * interface{}) derrors.Error{
	// check connection
	if err := s.CheckAndConnect(); err != nil {
		return err
	}

	stmt, names := qb.Select(table).Columns(tableColumnNames...).Where(qb.Eq(pkColumn)).ToCql()
	q := gocqlx.Query(s.Session.Query(stmt), names).BindMap(qb.M{pkColumn: pkValue})

	err := q.GetRelease(*result)
	if err != nil {
		if err.Error() == RowNotFoundMsg {
			return derrors.NewNotFoundError(table).WithParams(pkValue)
		} else {
			return derrors.AsError(err, "cannot get element")
		}
	}

	return nil
}

// UnsafeRemove removes an element from a table identified by a single primary key.
func (s* ScyllaDB) UnsafeRemove(table string, pkColumn string, pkValue string) derrors.Error{
	if err := s.CheckAndConnect(); err != nil {
		return err
	}

	// check if the asset exists
	exists, err := s.UnsafeGenericExist(table, pkColumn, pkValue)
	if err != nil {
		return err
	}
	if ! exists {
		return derrors.NewNotFoundError(pkValue)
	}

	// delete instance
	stmt, _ := qb.Delete(table).Where(qb.Eq(pkColumn)).ToCql()
	cqlErr := s.Session.Query(stmt, pkValue).Exec()

	if cqlErr != nil {
		return derrors.AsError(cqlErr, "cannot remove element")
	}
	return nil
}


// ----------------------------------------------------------------
// Others
// ----------------------------------------------------------------

// UnsafeClear truncates a set of tables.
func (s* ScyllaDB) UnsafeClear(tableNames []string) derrors.Error{
	// check connection
	if err := s.CheckAndConnect(); err != nil {
		return err
	}

	for _, targetTable := range tableNames{
		query := fmt.Sprintf("TRUNCATE TABLE %s", targetTable)
		// delete table
		err := s.Session.Query(query).Exec()
		if err != nil {
			log.Error().Str("trace", conversions.ToDerror(err).DebugReport()).Str("table", targetTable).Msg("failed to truncate table")
			return derrors.AsError(err, "cannot truncate table")
		}
	}
	return nil
}

// ----------------------------------------------------------------
// functions for when the PK is composite of more than one field
// ----------------------------------------------------------------

// UnsafeGenericExist checks if an element identified by a composite primary key exists.
func (s * ScyllaDB) UnsafeGenericCompositeExist(table string, pkColumn map[string]interface{}) (bool, derrors.Error){
	// check connection
	if err := s.CheckAndConnect(); err != nil {
		return false, err
	}

	var count int

	sb:= qb.Select(table).CountAll()
	for p := range pkColumn {
		sb = sb.Where(qb.Eq(p))
	}

	stmt, names := sb.ToCql()
	q := gocqlx.Query(s.Session.Query(stmt), names).BindMap(pkColumn)

	err := q.GetRelease(&count)
	if err != nil {
		if err.Error() == RowNotFoundMsg {
			return false, nil
		} else {
			return false, derrors.AsError(err, "cannot determinate if elements exists")
		}
	}

	return count > 0, nil
}

// UnsafeAdd adds a new element to a table identified by a composite primary key.
func (s * ScyllaDB) UnsafeCompositeAdd(table string, pkColumn map[string]interface{}, tableColumnNames []string, toAdd interface{}) derrors.Error{
	// check connection
	if err := s.CheckAndConnect(); err != nil {
		return err
	}
	exists, err := s.UnsafeGenericCompositeExist(table, pkColumn)
	if err != nil {
		return err
	}
	if exists {
		return derrors.NewAlreadyExistsError(table)
	}

	// insert the instance
	stmt, names := qb.Insert(table).Columns(tableColumnNames...).ToCql()
	q := gocqlx.Query(s.Session.Query(stmt), names).BindStruct(toAdd)
	cqlErr := q.ExecRelease()

	if cqlErr != nil {
		return derrors.AsError(cqlErr, "cannot add new element")
	}

	return nil
}

// UnsafeUpdate updates an element in a table identified by a single primary key.
func (s * ScyllaDB) UnsafeCompositeUpdate(table string, pkColumn map[string]interface{}, tableColumnNames []string, toUpdate interface{}) derrors.Error{
	// check connection
	if err := s.CheckAndConnect(); err != nil {
		return err
	}
	exists, err := s.UnsafeGenericCompositeExist(table, pkColumn)
	if err != nil {
		return err
	}
	if ! exists {
		return derrors.NewNotFoundError(table).WithParams(getParams(pkColumn))
	}

	sb:= qb.Update(table).Set(tableColumnNames...)
	for p := range pkColumn {
		sb = sb.Where(qb.Eq(p))
	}

	stmt, names := sb.ToCql()
	q := gocqlx.Query(s.Session.Query(stmt), names).BindStruct(toUpdate)
	cqlErr := q.ExecRelease()

	if cqlErr != nil {
		return derrors.AsError(cqlErr, "cannot update element")
	}

	return nil
}

// UnsafeGet retrieves an element from a table identified by a composite primary key.
func (s *ScyllaDB) UnsafeCompositeGet(table string, pkColumn map[string]interface{}, tableColumnNames []string, result * interface{}) derrors.Error{
	// check connection
	if err := s.CheckAndConnect(); err != nil {
		return err
	}

	sb:= qb.Select(table)
	for p := range pkColumn {
		sb = sb.Where(qb.Eq(p))
	}
	stmt, names := sb.ToCql()
	q := gocqlx.Query(s.Session.Query(stmt), names).BindMap(pkColumn)

	err := q.GetRelease(*result)
	if err != nil {
		if err.Error() == RowNotFoundMsg {
			return derrors.NewNotFoundError(table).WithParams(getParams(pkColumn))
		} else {
			return derrors.AsError(err, "cannot get element")
		}
	}

	return nil
}

// UnsafeRemove removes an element from a table identified by a single primary key.
func (s*ScyllaDB) UnsafeCompositeRemove(table string, pkColumn map[string]interface{}) derrors.Error{
	if err := s.CheckAndConnect(); err != nil {
		return err
	}

	// check if the asset exists
	exists, err := s.UnsafeGenericCompositeExist(table, pkColumn)
	if err != nil {
		return err
	}
	if ! exists {
		return derrors.NewNotFoundError(table).WithParams(getParams(pkColumn))
	}

	// delete
	sb:= qb.Delete(table)
	for p := range pkColumn {
		sb = sb.Where(qb.Eq(p))
	}

	stmt, names := sb.ToCql()
	q := gocqlx.Query(s.Session.Query(stmt), names).BindMap(pkColumn)

	cqlErr := q.Exec()

	if cqlErr != nil {
		return derrors.AsError(cqlErr, "cannot remove element")
	}
	return nil
}

func getParams(pkColumn map[string]interface{})([]interface{}){
	params := make([]interface{}, 0)
	for _, p := range pkColumn {
		params = append(params, p)
	}
	return params
}