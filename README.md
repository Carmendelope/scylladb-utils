# scylladb-utils

This repository contains Scylladb helpers for building providers in Go

## Getting Started

There are four types of functions

- Functions to manage the connection:
    - Connect
    - Disconnect
    - CheckConnection
    - CheckAndConnect 
    
- Functions to access tables whose pk is composed of a single field:
    - UnsafeGenericExist
    - UnsafeAdd
    - UnsafeUpdate
    - UnsafeGet
    - UnsafeRemove
    
- Functions to access tables whose pk is composed of more than one field:
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
func NewScyllaXXProvider(address string, port int, keyspace string) * ScyllaXXProvider{
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
    # one field in the primary key
    return sp.UnsafeUpdate(Table, TablePK, registry.id, AllColumnsNoPK, asset)
    # more than one value in the primary key
    return UnsafeCompositeUpdate(Table, PkMap, Columns, registry) 
}

func (sp *ScyllaXXProvider) Exists(id string) (bool, derrors.Error) {
    sp.Lock()
    defer sp.Unlock()
    # one field in the primary key
    return sp.UnsafeGenericExist(Table, TablePK, id)
    # more than one value in the primary key
    return UnsafeGenericCompositeExist(Table, PkMap)
}

func (sp *ScyllaXXProvider) Get(id string) (*entities.Registry, derrors.Error) {
    sp.Lock()
    defer sp.Unlock()
    var result interface{} = &entities.Registry{}
    # one field in the primary key
    err := sp.UnsafeGet(Table, TablePK, id, Columns, &result)
    # more than one value in the primary key
    err := sp.UnsafeCompositeGet(Table, pkMap, Columns, registry)
    if err != nil{
        return nil, err
    }
    return result.(*entities.Registry), nil
    
func (sp *ScyllaXXProvider) Remove(id string) derrors.Error {
    sp.Lock()
    defer sp.Unlock()
    # one field in the primary key
    return sp.UnsafeRemove(Table, TablePK, id)
    # more than one value in the primary key   
    return sp.UnsafeCompositeRemove(Table, pkMap)
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

### Build and compile

In order to build and compile this repository use the provided Makefile:

```
make all
```

This operation generates the binaries for this repo, download dependencies,
run existing tests and generate ready-to-deploy Kubernetes files.

### Run tests

Tests are executed using Ginkgo. To run all the available tests:

```
make test
```

### Integration tests

Some integration tests are included. To execute those, set up the following environment variables. 
​

​The following table contains the variables that activate the integration tests
 
 | Variable  | Example Value | Description |
 | ------------- | ------------- |------------- |
 | RUN_INTEGRATION_TEST  | true | Run integration tests |
 | IT_SCYLLA_HOST  | 127.0.0.1 | Scylla address |
 | IT_SCYLLA_PORT | 9042 | Scylla Port |
 | IT_NALEJ_KEYSPACE | testkeyspace | Test Schema |
 
In a scylla deploy, a database is required. Execute the following, you can create it executing: 
```
create KEYSPACE testkeyspace WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
create table testkeyspace.tableTest (id1 text, id2 text, id3 text, primary key (id1, id2));
create table testkeyspace.basicTableTest (id1 text, id2 text, id3 text, primary key (id1));
``` 

### Update dependencies

Dependencies are managed using Godep. For an automatic dependencies download use:

```
make dep
```

In order to have all dependencies up-to-date run:

```
dep ensure -update -v
```


## Contributing

Please read [contributing.md](contributing.md) for details on our code of conduct, and the process for submitting pull requests to us.


## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/your/project/tags). 

## Authors

See also the list of [contributors](https://github.com/nalej/grpc-utils/contributors) who participated in this project.

## License
This project is licensed under the Apache 2.0 License - see the [LICENSE-2.0.txt](LICENSE-2.0.txt) file for details.
# scylladb-utils

This repository contains Scylladb helpers for building providers in Go

    
