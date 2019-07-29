# scylladb-utils

This repository contains Scylladb helpers for building providers in Go

## Guidelines

There are three types of functions

- Functions to manage the connection:
    - Connect
    - Disconnect
    - CheckConnection
    - CheckAndConnect 
    
- Functions to access tables whose pk is composed of a single field
    - UnsafeGenericExist
    - UnsafeAdd
    - UnsafeUpdate
    - UnsafeGet
    - UnsafeRemove
    
- Functions to access tables whose pk is composed of more than one field
    - UnsafeGenericCompositeExist
    - UnsafeCompositeAdd
    - UnsafeCompositeUpdate
    - UnsafeCompositeGet
    - UnsafeCompositeRemove
    
- and one more function to truncate the tables:
    - UnsafeClear
    
## Basic Example
To use this library in nalej providers, to have to declare a `scylladb.ScyllaDB` be able to call the functions above
and a `sync.Mutex` to avoid concurrent access to Session

```
# type declaration
type ScyllaXXProvider struct {
	scylladb.ScyllaDB
	sync.Mutex
}

# funcion to build a provider (and connect it)
func NewScyllaXXProvider(address string, port int, keyspace string) * ScyllaAssetProvider{
	provider := ScyllaXXProvider{
		ScyllaDB : scylladb.ScyllaDB{
			Address: address,
			Port : port,
			Keyspace: keyspace,
		},
	}
	provider.Connect()
	return &provider
}

func (sp *ScyllaXXProvider) Disconnect() {
	sp.Lock()
	defer sp.Unlock()
	sp.ScyllaDB.Disconnect()
}


func (sp *ScyllaXXProvider) Add(registry entities.Registry) derrors.Error {
	sp.Lock()
	defer sp.Unlock()
	log.Debug().Interface("registry", registry).Msg("provider add registry")
	# one field in the primary key
	return sp.UnsafeAdd(Table, TablePK, pkValue, columns, registry)
	# more than one value in the primary key
	return UnsafeCompositeAdd(Table, PkMap, Columns, registry) 
	
}
func (sp *ScyllaXXProvider) Update(registry entities.Registry) derrors.Error {
	sp.Lock()
	defer sp.Unlock()
	return sp.UnsafeUpdate(Table, TablePK, registry.id, AllColumnsNoPK, asset)
}

func (sp *ScyllaXXProvider) Exists(id string) (bool, derrors.Error) {
	sp.Lock()
	defer sp.Unlock()
	return sp.UnsafeGenericExist(Table, TablePK, id)
}

func (sp *ScyllaXXProvider) Get(id string) (*entities.Registry, derrors.Error) {
	sp.Lock()
	defer sp.Unlock()
	var result interface{} = &entities.Registry{}
	err := sp.UnsafeGet(Table, TablePK, id, Columns, &result)
	if err != nil{
		return nil, err
	}
	return result.(*entities.Registry), nil
	
func (sp *ScyllaXXProvider) Remove(id string) derrors.Error {
	sp.Lock()
	defer sp.Unlock()
	return sp.UnsafeRemove(Table, TablePK, id)
}
}

```
where:
 - `Table` is the name of the table
 - `TablePK` is the name of the column of the Primary Key
 - `Columns` array with the names of all the columns in the table
 - `AllColumnsNoPK` array with the names of the columns that do not belong to the PK
 - `PkMap` is a `map[string]interface{}` Primary key values indexed by the column name
 - `Registry` is the record to be stored
 
 Note: `List method` must be implemented by the user