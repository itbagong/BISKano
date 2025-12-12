# kaos

Kano Open Service - library to build microservice solution on various protocol

## Components
### Internal
- Service, core element to setup service
- Context, component to embed data accross kaos ecosystem
- Deployer, expose end-point to public 
- Middleware, intercept request and playing around with context and data
- Module, to automate end-point creation

### External
- Entity model
- Midware and Module implementation
- Logger
- Datahub
- PubSub Event

## How to work with KAOS
### 1. Setup
```go
// prepare bolts and nuts
logger := appkit.LogWithPrefix('myapp')
hdata := datahub.NewHub(datahub.GeneralDbBuildConnector(dataConnectionString), usePool, 50)
hiam := datahub.NewHub(datahub.GeneralDbBuildConnector(iamConnectionString), usePool, 10)

// create the service and base point
s := kaos.NewService().SetBasePoint(serviceName + version)

// chain the support element into service
s.SetLogger(logger).
  RegisterDataHub(hdata,"default").
  RegisterDataHub(hiam,"iam").
  RegisterEventHub(ev, "default", eventSecret)
```

### 2. Create entity
```go
type Sales struct{
    orm.DataModelBase   `bson:"-" json:"-"`
    ID              string      `bson:"_id" json:"_id" key:"1"`
    SalesDate       time.Time
    ChannelID       string
    ChannelName     string
    CustomerID      string
    CustomerName    string
    SalesAmount     float64
    DiscountAmount  float64
    TaxAmount       float64
    TotalAmount     float64
    Created         time.Time
    LastUpdate      time.Time
}

func (s *Sales) TableName() string{
    return "Saleses"
}

func (s *Sales) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (s *Sales) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *PaymentChannel) FK() []*orm.FKConfig {
	return []*orm.FKConfig {
        {FieldId:"CustomerID", RefTableName:new(Customer).TableName(), RefTableField:"_id", Map:codekit.M{"CustomerName":"Name"}},
        {FieldId:"ChannelID", RefTableName:new(SalesChannel).TableName(), RefTableField:"_id", Map:codekit.M{"ChannelName":"Name"}},
    }
}

func (o *PaymentChannel) Indexes() []orm.DbIndex {
    return []dbflex.DbIndex {
        {Name:"CustomerIndex", IsUnique:false, Fields:[]string{"CustomerID"}},
        dbflex.NewIndex("DateIndex", false, "SalesDate", "CustomerID"),
    }
}
//-- dan seterusnya: GetID, SetID, Indexes, PreSave, PostSave, FK
```

### 3. Create business logic
Kaos has `auto-discovery` feature, system will scan all function within an object that match following contract to be auto-routed. Specifications of contract are:
- Has 2 inputs
- 1st input is `*kaos.Context`, 2nd input is anything
- Has 2 outputs
- 1st output is anything, 2ns is error

ie: `func (o *SalesLogic) CreateFromQuote(ctx *kaos.Context, quoteID string) (*Sales, error)`

Using auto-discovery, the needs to create `use case with interfaces` become ***not-mandatory***

Some standard on creating business logic:
- All logic should be created within one folder (Engine, Logic, etc) to simplify process
- All supporting object needed (datahub, event, etc) can be passed thru kaos.Context or using singleton
- Singleton is easier but scope will be only within a logic only

```
func (o *SalesLogic) CreateFromQuote(ctx *kaos.Context, quoteID string) (*Sales, error) {
    h := ctx.Hubs["default"]

    quote := new(SalesQuote)
    if e:=h.GetByID(quote, quoteID);e!=nil {
        return nil, errors.New("invalid quote")
    }

    quoteLines := populateQuoteLines(h, quote)

    sales := new(Sales)
    sales.CopyFromQuote(quote)
    salesLines := sales.CopyLinesFromQuoteLines(quoteLines)

    tx := h.BeginTx()
    e := wrapTx(tx, func(){
        if e:=h.Save(sales);e!=nil {
            panic(e.Error())
        }

        for _, line := range salesLines {
            if e:=h.Save(line);e!=nil {
                panic(e.Error())
            }
        }
    })
    return sales, e
}
```

### 4. Register logic
```go
s.Register(new(Sales), "sales")                                         // deployed on all protocol being usd
s.Register(new(SalesEvent), "sales").SetDeployer(knats.DeployerName)    // to be deployed as PubSub - NATS
s.Register(new(SalesRest), "sales").SetDeployer(hd.DeployerName)        // to be deployed as REST
```

### 5. Use mod and midware if needed
```go
dbMod := dbMod.New()
uiMod := suim.New()

s.RegisterModel(new(Customer),"sales").SetMod(uiMod, dbMod).SetDeployer(hd.DeployerName)
    .DisableRoute("Insert","Update","Save","Delete")

s.Group().RegisterMWS(kamis.NeedAccess("SalesAdmin")).Apply(
    s.RegisterModel(new(Customer),"sales").SetMod(dbMod).
        AllowOnlyRoute("Insert","Update","Save","Delete").
        SetDeployer(hd.DeployerName)
)
```

### 6. Activate the deployer
```go
// REST deployment
mux := http.NewServerMux()
if e := hd.NewDeployer(nil).Deploy(s, mux);e!=nil {
    // handle error
}
go http.ListenAndServer(host, mux)

// PubSub Event Deployment
e := knats.NewDeployer(ev).Deploy(s, nil)
```
