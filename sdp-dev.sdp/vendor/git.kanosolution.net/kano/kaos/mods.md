# kaos-mods
Dibawah ini adalah daftar mods yang sudah dikembangkan

## dbmod (Data Gateway)
Otomatis akan membuat routing untuk CRUD API

### Penggunaan
```
dbm := dbmod.New()
s.RegisterModel(new(model.Customer),"customer").SetMod(dbm).SetDeployer(hd.DeployerName)
```
Secara otomatis akan membuat endpoint berikut:
- /customer/get => `input: keys []string` | output-ok: `customer model` 
- /customer/gets => `input: queryparam` | output-ok: `{data:[]customer, count:int}`
- /customer/find => `input: queryparam` | output-ok: `[]customer`
- /customer/insert 
- /customer/update
- /customer/save
- /customer/delete
- /customer/deteleMany

### Queries
Apabila kita mendefiniskan queries di data model
```
func (c *Customer) Queries() map[string]dbflex.Query{
    retur map[string]dbflex.Query{
        "Group": {Param: dbflex.NewQueryParam.SetWhere(dbflex.Eq("Group","$(Group)"))}, 
        "Country": {Param: dbflex.NewQueryParam.SetWhere(dbflex.Eq("Group","$(Country)"))}, 
    }
}
```

Maka secara otomaris akan membuat end-point `GetByGroup`,`GetsByGroup`,`FindByGroup` dan yang sama untuk `Country`.
Kita bisa memanggil end-point tersebut di js misal:

```
axios.post("/customer/getsbygroup",{"Group":"Local","Take":10,"OrderBy":["Name","-Created"]}).
    then(r => {
        data.records = r.data
        data.count = r.count
    })
```

## suim (UI Presenter)
### Penggunaan
```
s.RegisterModel(new(model.Customer), "customer").SetMode(uimod).SetDeployer(hd.DeployerName)
```

Otomatis akan menghasilkan:
- /customer/formconfig
- /customer/gridconfig

Dimana bisa kita gunakan untuk membuat grid, form atau data list di UI sebagai berikut:

```
<template>
    <s-list 
        form-config="/customer/formconfig"
        list-config="/customer/gridconfig"
        form-insert="/customer/insert"
        form-update="/customer/update"
        form-read="/customer/get" 
        list-read="/customer/gets"
        list-delete="/customer/delete" />
</template>
```
