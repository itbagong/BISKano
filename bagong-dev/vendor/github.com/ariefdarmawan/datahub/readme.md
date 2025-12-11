# Datahub
Datahub is wrapper of `dbflex`. Datahub automates some process to dealing with data

- Connection Pooling Management
- Open, Release and Close Connection

## Setup 
To work with datahub, we have to prepare a datahub object.
```go
h := datahub.New(func()(dbflex.IConnection, error){
    //-- code here

    return conn, nil
})
defer h.Close()
```

Once the `h` object is created, we can using it right away to handle our data processing within the solution.

## ORM-ish data
### Get
```
obj := new(MyDataRecord)
e := h.GetByID(obj, "myid")
```

### Get by Attribute
```
e := h.GetByAtrr(obj, "Name", "Paijo")
```

### Get by filter
```
e := h.GetByFilter(obj, dbflex.Eqs("Grade",10,"Location","Indonesia"))
```

### Gets
Reading multiple rows of data. We need to pass 3 parameters

- model, orm object that will be read
- parm, a dbflex QueryParam object
- target, target of data to be stored

```
items := []Item{}
qp := dbflex.NewQueryParam().SetWhere(f).SetTake(5) // filter by 5 and take 5 records only
e := h.Gets(new(item), parm, &items)
```

### Count
```
e := h.Count(new(item), parm)
```

### Write to database
```
e := h.Insert(model)
e := h.Update(model)
e := h.Save(model)
e := h.Delete(model)
```

## Non ORM
To access non-orm data, we can use method `GetAny*`, `Populate*` and `Execute`

### Get data
```
e := h.GetAnyByFilter("myTable", dbflex.Eq("_id","dataid"), target)
e := h.GetAnyByParm("myTable", dbflex.NewQueryParam().SetWhere("Grade",20).Aggr(dbflex.Avg("Salary")), target)
```

### Read multi records
```
e := h.Populate(dbflex.From("myTable").Select().Take(1), target)
e := h.PopulateByFilter("mytable", myFilter)
e := h.PopulateByParm("mytable", myParm)
e := h.PopulateSQL(sqlTxt)  // only for rdbms
```

### Write Data
```
e := h.InsertAny("mytable", record)
e := h.UpdateAny("mytable", dbflex.Eq("_id","myid"), record, "Status")
e := h.SaveAny("mytable", dbflex.Eq("_id", "myid"), record)
e := h.DeleteAny("mytable", nil)    // delete all record
e := h.DeleteAny("mytable", dbflex.Eq("Grade",0))
```

## Sync/Migrate Database
```
// ORM
e := h.EnsureDb(ormModel)

// Non ORM
e := h.EnsureTable("mytable", myTableRecord)
e := h.EnsureIndex("myTable", "Email_Index", true, "Email")
```
