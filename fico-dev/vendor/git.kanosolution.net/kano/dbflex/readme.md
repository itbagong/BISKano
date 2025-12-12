# dbflex
dbflex is universal database connector. It wraps existing drivers using uniform contract, hence developer can access all databases using same method.

For better experience dealing with data, please check `orm` below and `datahub`

## Using DbFlex
To use dbflex, we need to import dbflex and its driver
```go
import (
    "git.kanosolution.net/kano/dbflex"
    _ "github.com/ariefdarmawan/flexmgo" //-- flexdb mongodb wrapper of mgo.v2
)

var (
    connTxt = "mongodb://localhost:27017"
)

func main() {
    conn, err := dbflex.NewConnectionFromURI(connTxt)
    if err!=nil {
        // handle error here
    }

    if err = conn.Connect();err!=nil {
        // handle connect error here
    }
    defer conn.Close()
}
```

## Command
We need to setup a command to work with DbFlex. Command is a marshallable object and meants to create abstraction of SQL command.

```go
cmdGet := dbflex.From("myTable").Select("_id","Name").Where(dbflex.Eq("_id","SomeID1"))
cmdInsert := dbflex.from("myTable").Insert()
```

Supported commands are:

- *From*, define source of table
- *Select*, prepare a select command, without parameter it will read all columns
- *Insert*, prepare an insert command
- *Update*, prepare an update command
- *Delete*, prepare a delete command
- *Save*, prepare update or insert, depend on existence of data
- *Where*, define filter for data, see *Filter* below
- *OrderBy*, define order of data to be read
- *GroupBy*, define grouping behavior of data to be read
- *Aggr*, define aggregation command
- *Take*, read only n data maxed
- *Skip*, skip n data for reading process
- *Command*, to execute a custom command *(depends on drivers/extension)* 

## Filter
Filter object is required as input of *Where* subcommand to identify the filter criteria

```go
f := dbflex.Or(dbflex.Eq("Grade",30), dbflex.Eqs("Role","admin","Scope","All")) 
// grade=30 or (role="admin" and scope="All")
```

Supported filter are:

- Eq
- Ne
- Eqs, multi Eq with and relation
- Gt
- Gte
- Lt
- Lte
- Range
- Contains
- StartWith
- EndWith
- In
- Nin
- And
- Or
  
## Reading Data
Reading data will utilizing `Cursor` method of connection command

### Get a data
```go
result := new(Result)
if e:= conn.Cursor(dbflex.From("mytable").Select().Where(dbflex.Eq("_id","MyID"))).Fetch(result).Close(); e! = nil {
    // handle error
} 
//-- close the cursor when it won't be needed
//-- do what is needed with result
```

### Get a multi row of data
```go
results := []Result{}
cursor := conn.Cursor(cmd)
for {
    // fetch every 5 record
    if e=cursor.Fetchs(&results, 5);e!=nil {
        break
    }
    // do something with results
}
```

### Get aggregation result
```go
cmd := From("tableName").GroupBy("Grade").
    Aggr(NewAggrItem("Count",AggrSum,1), NewAggrItem("SalaryAverage",AggrAvg,"Salary"))
results := []codekit.M{}
err := conn.Cursor(cmd).Fetchs(&results,0).Close()
```

## Write Operation
Write operation will use `Execute` command of `Connection` object

For write operation, normally it requires `codekit.M` object to be passed as 2nd parameter of `Execute` command. This is require to define the data to be write and for future enhancement requirement 

### Insert
```go
e := conn.Execute(From("table").Insert(), codekit.M{"data": myRecord})
```

### Update
```go
e := conn.Execute(From("table").Update("MinYearRequired").Where(Eq("Grade",20)), codekit.M{"data": codekit.M{"MinYearRequired":15}})
// update MinYearRequired=15 where Grade=20 
```

### Delete
```go
e := conn.Execute(From("table").Delete().Where(Eq("Grade",20)), nil)
// delete all records where Grade=20 
```

## Database Sync
dbflex connection object also provides 2 methods to sync database structure and indexes `EnsureTable` and `EnsureIndex`

```go
// ensure employeeTable has primary key "_id" and table structure should follow Employee object 
err := conn.EnsureTable("employeeTable", []string{"_id"}, new(Employee))

// ensure employeeTanle has index "EmailIndex" as not unique for indexing email field
err := conn.EnsureIndex("employeeTable", "emailIndex", false, "Email")
```

## Transaction
If database used support transaction and its driver + flex wrapper applying it. We can do trx also on dbflex

```go
    trxConn := conn.BeginTx()
    var e error
    func() {
        defer func() {
            if r:=recover();r!=nil {
               e = r.(string)
            } 
        }

        if e = orm.Insert(trxConn, aData); e!=nil {
            return
        }

        if e = orm.Update(trxConn, anotherData, "TotalAmount"); e!=nil {
            return 
        }
    }()

    if e==nil {
        trxConn.Commit()
    } else {
        trxConn.RollBack()
    }
```

## Pooling
Some database support connection pooling out of the box, but some others are not. dbflex provides universal pooling mechanism for all datastore supported by dbflex.

```go
// setup pooling
p := NewDbPooling(10, func()(IConnection, error) {
    conn := dbflex.NewConnectionFromURI(someConn)
    conn.Connect()
    return conn, nil
})
p.Timeout = 5 * time.Second()       //if idle for 5s, it will be removed
defer p.Close()


// get connection
pconn, err := p.Get()
defer pcon.Release()

// use connection
pconn.Connection().cursor(cmd)
```

## ORM
`orm` module provide better experience on dealing with data active record.


### orm struct
To work with orm, we need to build a struct implements `orm.DataModel`, to make it easy just inherits `orm.DataModelBase` and change only few methods (`TableName`, `SetID`, `GetID`)

`GetID` and `SetID` are not actually required if you tag key fields with `key:"1"`. But it is nice to have them since it will make performance better by avoiding to many cast

```go
type OrmRecord struct {
    orm.DataModelBase `bson:"-" json:"-"`  //ignore this on marshall process
    ID          string      `bson:"_id" json:"_id"`
    Name        string
    BirthDate   time.Time     
}

func (o *OrmRecord) TableName() string{
    return "myTable"
}

func (o *OrmRecord) SetID(key ...interface{}) {
    o.ID = key[0].(string)
}

func (o *OrmRecord) GetID() ([]string, []interface{}) {
    return []string{"_id"},     //-- name of fields for ID
        []interface{}{o.ID}     //-- field(s) value
}
```

Once you setup an Orm-struct, then we can play around with orm module
```go
    e := orm.Get(conn, &record)
    e := orm.GetWhere(conn, dbflex.Eq("_id", "myid"))
    e := orm.Gets(conn, &record, &records, &qp)
    e := orm.Insert(conn, data)
    e := orm.Update(conn, data)
    e := orm.Delete(conn, data)
    e := orm.Save(conn, data)
```
