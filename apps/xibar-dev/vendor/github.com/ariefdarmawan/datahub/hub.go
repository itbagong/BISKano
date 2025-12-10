package datahub

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"

	"github.com/sebarcode/codekit"
	"github.com/sebarcode/kiva"
	"github.com/sebarcode/logger"
)

// Hub main datahub object. This object need to be initiated to work with datahub
type Hub struct {
	connFn   func() (dbflex.IConnection, error)
	usePool  bool
	pool     *dbflex.DbPooling
	poolSize int

	useCache      bool
	cacheProvider kiva.MemoryProvider

	poolItems []*dbflex.PoolItem
	mtx       *sync.Mutex
	_log      *logger.LogEngine

	txconn dbflex.IConnection
}

func GeneralDbConnBuilder(txt string) func() (dbflex.IConnection, error) {
	return GeneralDbConnBuilderWithTx(txt, false)
}

func GeneralDbConnBuilderWithTx(txt string, useTx bool) func() (dbflex.IConnection, error) {
	return func() (dbflex.IConnection, error) {
		conn, e := dbflex.NewConnectionFromURI(txt, nil)
		if e != nil {
			return nil, e
		}

		if !useTx {
			conn.DisableTx(!useTx)
		}

		if e = conn.Connect(); e != nil {
			return nil, fmt.Errorf("fail to connect to datahub. %s", e.Error())
		}

		return conn, nil
	}
}

type HubOptions struct {
	AutoClose   time.Duration
	AutoRelease time.Duration
	Timeout     time.Duration
	UsePool     bool
	PoolSize    int
}

// NewHub function to create new hub
func NewHub(fn func() (dbflex.IConnection, error), usePool bool, poolsize int) *Hub {
	h := new(Hub)
	h.connFn = fn
	h.usePool = usePool
	h.poolSize = poolsize

	if h.usePool {
		h.pool = dbflex.NewDbPooling(h.poolSize, h.connFn).SetLog(h.Log())
		h.pool.Timeout = 7 * time.Second
		h.pool.AutoClose = 5 * time.Second
		//h.pool.AutoRelease = 3 * time.Second
	}
	return h
}

// NewHub function to create new hub
func NewHubWithOpts(fn func() (dbflex.IConnection, error), opts *HubOptions) *Hub {
	if opts == nil {
		opts = new(HubOptions)
	}

	h := new(Hub)
	h.connFn = fn
	h.SetOptions(opts)
	return h
}

// Set Hub Opts
func (h *Hub) SetOptions(opts *HubOptions) *Hub {
	h.usePool = opts.UsePool
	h.poolSize = opts.PoolSize

	if h.usePool {
		h.pool = dbflex.NewDbPooling(h.poolSize, h.connFn).SetLog(h.Log())
		h.pool.Timeout = opts.Timeout
		h.pool.AutoClose = opts.AutoClose
		h.pool.AutoRelease = opts.AutoRelease
	}
	return h
}

// Log get logger object
func (h *Hub) Log() *logger.LogEngine {
	if h._log == nil {
		h._log = logger.NewLogEngine(true, false, "", "", "")
	}
	return h._log
}

// SetLog set logger
func (h *Hub) SetLog(l *logger.LogEngine) *Hub {
	h._log = l
	if h.pool != nil {
		h.pool.SetLog(l)
	}
	return h
}

// GetConnection to generate connection. It will return index, connection and error. Index and connection need
// to be retain for purpose of closing the connection. It is advised to use other DB operation related of Hub rather
// than build manual connection
func (h *Hub) GetConnection() (string, dbflex.IConnection, error) {
	return h.getConn()
}

// CloseConnection to close connection
func (h *Hub) CloseConnection(idx string, conn dbflex.IConnection) {
	h.closeConn(idx, conn)
}

// GetClassicConnection get connection without using pool. CleanUp operation need to be done manually
func (h *Hub) GetClassicConnection() (dbflex.IConnection, error) {
	return h.connFn()
}

func (h *Hub) getConnFromPool() (string, dbflex.IConnection, error) {
	if h.txconn != nil {
		return "", h.txconn, nil
	}

	if h.poolSize == 0 {
		h.poolSize = 100
	}

	if h.mtx == nil {
		h.mtx = new(sync.Mutex)
	}

	if h.pool == nil {
		h.pool = dbflex.NewDbPooling(h.poolSize, h.connFn).SetLog(h.Log())
		h.pool.Timeout = 90 * time.Second
		h.pool.AutoClose = 5 * time.Second
		//h.pool.AutoRelease = 3 * time.Second
	}

	it, err := h.pool.Get()
	if err != nil {
		return "", nil, fmt.Errorf("unable get connection from pool. %s", err.Error())
	}

	conn := it.Connection()
	idx := ""
	h.mtx.Lock()
	defer h.mtx.Unlock()

	h.poolItems = append(h.poolItems, it)
	idx = it.ID
	return idx, conn, nil
}

// SetAutoCloseDuration set duration for a connection inside Hub Pool to be closed if it is not being used
func (h *Hub) SetAutoCloseDuration(d time.Duration) *Hub {
	if h.usePool {
		if h.pool == nil {
			h.pool = dbflex.NewDbPooling(h.poolSize, h.connFn)
		}
		h.pool.AutoClose = d
	}
	return h
}

// SetAutoReleaseDuration set duration for a connection in pool to be released for a process
func (h *Hub) SetAutoReleaseDuration(d time.Duration) *Hub {
	if h.usePool {
		if h.pool == nil {
			h.pool = dbflex.NewDbPooling(h.poolSize, h.connFn)
		}
		h.pool.Timeout = d + time.Duration(5*time.Second)
		h.pool.AutoRelease = d
	}
	return h
}

func (h *Hub) closeConn(idx string, conn dbflex.IConnection) {
	if h.txconn != nil {
		return
	}

	if !h.usePool {
		conn.Close()
	}

	if h.mtx == nil {
		h.mtx = new(sync.Mutex)
	}
	h.mtx.Lock()
	defer h.mtx.Unlock()

	for _, it := range h.poolItems {
		if it.ID == idx {
			it.Release()
			break
		}
	}
}

func (h *Hub) getConn() (string, dbflex.IConnection, error) {
	if h.txconn != nil {
		return "", h.txconn, nil
	}

	if h.connFn == nil {
		return "", nil, fmt.Errorf("connection fn is not yet defined")
	}

	if h.usePool {
		return h.getConnFromPool()
	}

	conn, err := h.connFn()
	if err != nil {
		return "", nil, fmt.Errorf("unable to open connection. %s", err.Error())
	}
	return "", conn, nil
}

// UsePool is a hub using pool
func (h *Hub) UsePool() bool {
	return h.usePool
}

// PoolSize returns size of the pool
func (h *Hub) PoolSize() int {
	return h.poolSize
}

// DeleteQuery delete object in database based on specific model and filter, will be DEPRECATED use DeleteByFilter instead
func (h *Hub) DeleteQuery(model orm.DataModel, where *dbflex.Filter) error {
	return h.DeleteByFilter(model, where)
}

// DeleteByFilter delete object in database based on specific model and filter
func (h *Hub) DeleteByFilter(model orm.DataModel, where *dbflex.Filter) error {
	idx, conn, err := h.getConn()
	if err != nil {
		return fmt.Errorf("connection error. %s", err.Error())
	}
	defer h.closeConn(idx, conn)

	cmd := dbflex.From(model.TableName()).Delete()
	if where != nil {
		cmd.Where(where)
	}
	_, err = conn.Execute(cmd, nil)
	return err
}

// Save will save data into database, if fields is activated it will save the mentioned fields only,
// fields will only work with update hence it will assume record already exist
func (h *Hub) Save(data orm.DataModel, fields ...string) error {
	objID := getID(data)
	data.SetThis(data)
	idx, conn, err := h.getConn()
	if err != nil {
		return fmt.Errorf("connection error. %s", err.Error())
	}
	defer h.closeConn(idx, conn)

	if len(fields) == 0 {
		if err = orm.Save(conn, data); err != nil {
			return err
		}
	} else {
		w := data.GetFilterID(conn)
		if err := orm.DataModel(data).PreSave(conn); err != nil {
			return err
		}
		cmd := dbflex.From(data.TableName()).Where(w).Update(fields...)
		if _, err := conn.Execute(cmd, codekit.M{}.Set("data", data)); err != nil {
			return err
		}
		orm.DataModel(data).PostSave(conn)
	}

	if h.useCache && h.cacheProvider != nil && getFieldCacheSetup(data.TableName(), data) != nil {
		h.cacheProvider.Set(data.TableName(), objID, data, nil)
	}

	return nil
}

// Insert will create data into database
func (h *Hub) Insert(data orm.DataModel) error {
	data.SetThis(data)
	idx, conn, err := h.getConn()
	if err != nil {
		return fmt.Errorf("connection error. %s", err.Error())
	}
	defer h.closeConn(idx, conn)

	if err = orm.Insert(conn, data); err != nil {
		return err
	}

	if h.useCache && h.cacheProvider != nil && getFieldCacheSetup(data.TableName(), data) != nil {
		h.cacheProvider.Set(data.TableName(), getID(data), data, nil)
	}

	return nil
}

// UpdateField update relevant fields in data based on specific filter
func (h *Hub) UpdateField(data orm.DataModel, where *dbflex.Filter, fields ...string) error {
	data.SetThis(data)
	idx, conn, err := h.getConn()
	if err != nil {
		return fmt.Errorf("connection error. %s", err.Error())
	}
	defer h.closeConn(idx, conn)

	updatedFields := fields
	if err := orm.DataModel(data).PreSave(conn); err != nil {
		return err
	}
	cmd := dbflex.From(data.TableName()).Update(updatedFields...).Where(where)
	conn.Execute(cmd, codekit.M{}.Set("data", data))
	orm.DataModel(data).PostSave(conn)

	return nil
}

// Update will update single data in database based on specific model
func (h *Hub) Update(data orm.DataModel, fields ...string) error {
	objID := getID(data)

	data.SetThis(data)
	idx, conn, err := h.getConn()
	if err != nil {
		return fmt.Errorf("connection error. %s", err.Error())
	}
	defer h.closeConn(idx, conn)

	if len(fields) == 0 {
		if err = orm.Update(conn, data); err != nil {
			return err
		}
		return nil
	}

	var f *dbflex.Filter
	idFields, idValues := data.GetID(conn)
	if len(idFields) == 0 {
		f = dbflex.Eq(idFields[0], idValues[0])
	} else {
		fs := make([]*dbflex.Filter, len(idFields))
		for idx, name := range idFields {
			fs[idx] = dbflex.Eq(name, idValues[idx])
		}
		f = dbflex.And(fs...)
	}
	err = h.UpdateField(data, f, fields...)
	if err != nil {
		return err
	}

	if h.useCache && h.cacheProvider != nil && getFieldCacheSetup(data.TableName(), data) != nil {
		// because no fields specifieds, means all update, we can set cache
		if len(fields) == 0 {
			h.cacheProvider.Set(data.TableName(), objID, data, nil)
			return nil
		}

		// fields update specified, we need to pull all object to pull all object and cache it
		obj, ok := reflect.New(reflect.TypeOf(data).Elem()).Interface().(orm.DataModel) // make new objevt with same type
		if !ok {
			return nil
		}
		obj.SetID(getID(data))
		if e := h.GetByID(obj); e == nil {
			h.cacheProvider.Set(data.TableName(), objID, obj, nil)
		}
	}

	return nil
}

// Delete delete the active record on database
func (h *Hub) Delete(data orm.DataModel) error {
	objID := getID(data)
	data.SetThis(data)
	idx, conn, err := h.getConn()
	if err != nil {
		return fmt.Errorf("connection error. %s", err.Error())
	}
	defer h.closeConn(idx, conn)

	if err = orm.Delete(conn, data); err != nil {
		return err
	}

	if h.useCache && h.cacheProvider != nil && getFieldCacheSetup(data.TableName(), data) != nil {
		h.cacheProvider.Delete(data.TableName(), objID)
	}

	return nil
}

// GetByID returns single data based on its ID. Data need to be comply with orm.DataModel
func (h *Hub) GetByID(data orm.DataModel, ids ...interface{}) error {
	data.SetThis(data)
	data.SetID(ids...)
	return h.Get(data)
}

// GetByAttr returns single data based on value on single field. Data need to be comply with orm.DataModel
func (h *Hub) GetByAttr(data orm.DataModel, attr string, value interface{}) error {
	w := dbflex.Eq(attr, value)
	qp := dbflex.NewQueryParam().SetWhere(w).SetTake(1)
	return h.GetByParm(data, qp)
}

/*
	GetByFilter returns single data based on filter enteered. Data need to be comply with orm.DataModel.

Because no sort is defined, it will only 1st row by any given sort
If sort is needed pls use by ByParm
*/
func (h *Hub) GetByFilter(data orm.DataModel, filter *dbflex.Filter) error {
	qp := dbflex.NewQueryParam().SetWhere(filter).SetTake(1)
	return h.GetByParm(data, qp)
}

// GetByParm return single data based on filter
func (h *Hub) GetByParm(data orm.DataModel, parm *dbflex.QueryParam) error {
	data.SetThis(data)
	if parm == nil {
		parm = dbflex.NewQueryParam()
	}

	idx, conn, err := h.getConn()
	if err != nil {
		return fmt.Errorf("connection error. %s", err.Error())
	}
	defer h.closeConn(idx, conn)

	cmd := dbflex.From(data.TableName())
	if len(parm.Select) == 0 {
		cmd.Select()
	} else {
		cmd.Select(parm.Select...)
	}
	if where := parm.Where; where != nil {
		cmd.Where(where)
	}
	if sort := parm.Sort; len(sort) > 0 {
		cmd.OrderBy(sort...)
	}
	if skip := parm.Skip; skip > 0 {
		cmd.Skip(skip)
	}
	if take := parm.Take; take > 0 {
		cmd.Take(take)
	}
	cursor := conn.Cursor(cmd, nil)
	if err := cursor.Error(); err != nil {
		return err
	}
	if err = cursor.Fetch(data).Close(); err != nil {
		return err
	}
	return nil
}

// Get return single data based on model. It will find record based on releant ID field
func (h *Hub) Get(data orm.DataModel) error {
	data.SetThis(data)

	if h.useCache && h.cacheProvider != nil && getFieldCacheSetup(data.TableName(), data) != nil {
		//if e := h.cacheKv.Get(fmt.Sprintf("%s:%s", data.TableName(), getId(data)), data); e != nil {
		if _, e := h.cacheProvider.Get(data.TableName(), getID(data), data); e != nil {
			return nil
		}
	}

	idx, conn, err := h.getConn()
	if err != nil {
		return fmt.Errorf("connection error. %s", err.Error())
	}
	defer h.closeConn(idx, conn)

	if err = orm.Get(conn, data); err != nil {
		return err
	}

	return nil
}

// Gets return all data based on model and filter
func (h *Hub) Gets(data orm.DataModel, parm *dbflex.QueryParam, dest interface{}) error {
	if parm == nil {
		parm = dbflex.NewQueryParam()
	}

	idx, conn, err := h.getConn()
	if err != nil {
		return fmt.Errorf("connection error. %s", err.Error())
	}
	defer h.closeConn(idx, conn)

	if err = orm.Gets(conn, data, dest, parm); err != nil {
		return err
	}

	return nil
}

// GetsByFilter like gets but require only filter
func (h *Hub) GetsByFilter(data orm.DataModel, filter *dbflex.Filter, dest interface{}) error {
	qp := dbflex.NewQueryParam()
	qp.SetWhere(filter)
	return h.Gets(data, qp, dest)
}

// Count returns number of data based on model and filter
func (h *Hub) Count(data orm.DataModel, qp *dbflex.QueryParam) (int, error) {
	return h.CountAny(data.TableName(), qp)
}

// CountByFilter returns number of data based on model and filter
func (h *Hub) CountByFilter(data orm.DataModel, f *dbflex.Filter) (int, error) {
	return h.CountAny(data.TableName(), dbflex.NewQueryParam().SetWhere(f))
}

// CountByAnyFilter returns number of data based on table name and filter
func (h *Hub) CountAnyByFilter(name string, f *dbflex.Filter) (int, error) {
	return h.CountAny(name, dbflex.NewQueryParam().SetWhere(f))
}

// CountAny returns number of data based on tablename and filter
func (h *Hub) CountAny(name string, qp *dbflex.QueryParam) (int, error) {
	if qp == nil {
		qp = dbflex.NewQueryParam()
	}

	idx, conn, err := h.getConn()
	if err != nil {
		return 0, fmt.Errorf("connection error. %s", err.Error())
	}
	defer h.closeConn(idx, conn)

	var cmd dbflex.ICommand
	if qp == nil || qp.Where == nil {
		cmd = dbflex.From(name).Select()
	} else {
		cmd = dbflex.From(name).Where(qp.Where).Select()
	}
	cur := conn.Cursor(cmd, nil)
	if err = cur.Error(); err != nil {
		return 0, fmt.Errorf("cursor error. %s", err.Error())
	}
	defer cur.Close()
	return cur.Count(), nil
}

/*
	GetAnyByFilter returns single data based on filter enteered. Data need to be comply with orm.DataModel.

Because no sort is defined, it will only return 1st row by any given resultset
If sort is needed pls use by ByParm
*/
func (h *Hub) GetAnyByFilter(tableName string, filter *dbflex.Filter, dest interface{}) error {
	qp := dbflex.NewQueryParam().SetWhere(filter).SetTake(1)
	return h.GetAnyByParm(tableName, qp, dest)
}

// GetAnyByParm return single data based on filter and table name
func (h *Hub) GetAnyByParm(tableName string, parm *dbflex.QueryParam, dest interface{}) error {
	if parm == nil {
		parm = dbflex.NewQueryParam()
	}

	idx, conn, err := h.getConn()
	if err != nil {
		return fmt.Errorf("connection error. %s", err.Error())
	}
	defer h.closeConn(idx, conn)

	cmd := dbflex.From(tableName)
	if len(parm.Select) == 0 {
		cmd.Select()
	} else {
		cmd.Select(parm.Select...)
	}
	if where := parm.Where; where != nil {
		cmd.Where(where)
	}
	if sort := parm.Sort; len(sort) > 0 {
		cmd.OrderBy(sort...)
	}
	if skip := parm.Skip; skip > 0 {
		cmd.Skip(skip)
	}
	if take := parm.Take; take > 0 {
		cmd.Take(take)
	}
	cursor := conn.Cursor(cmd, nil)
	if err := cursor.Error(); err != nil {
		return err
	}
	if err = cursor.Fetch(dest).Close(); err != nil {
		return err
	}
	return nil
}

// Execute will execute command. Normally used with no-datamodel object
func (h *Hub) Execute(cmd dbflex.ICommand, object interface{}) (interface{}, error) {
	idx, conn, err := h.getConn()
	if err != nil {
		return nil, fmt.Errorf("connection error. %s", err.Error())
	}
	defer h.closeConn(idx, conn)

	parm := codekit.M{}
	return conn.Execute(cmd, parm.Set("data", object))
}

// Populate will return all data based on command. Normally used with no-datamodel object
func (h *Hub) Populate(cmd dbflex.ICommand, result interface{}) (int, error) {
	idx, conn, err := h.getConn()
	if err != nil {
		return 0, fmt.Errorf("connection error. %s", err.Error())
	}
	defer h.closeConn(idx, conn)

	c := conn.Cursor(cmd, nil)
	if err = c.Error(); err != nil {
		return 0, fmt.Errorf("unable to prepare cursor. %s", err.Error())
	}
	defer c.Close()
	if err = c.Fetchs(result, 0).Error(); err != nil {
		return 0, fmt.Errorf("unable to fetch data. %s", err.Error())
	}
	return c.Count(), nil
}

// PopulateByParm returns all data based on table name and QueryParm. Normally used with no-datamodel object
func (h *Hub) PopulateByParm(tableName string, parm *dbflex.QueryParam, dest interface{}) error {
	idx, conn, err := h.getConn()
	if err != nil {
		return fmt.Errorf("connection error. %s", err.Error())
	}
	defer h.closeConn(idx, conn)

	qry := dbflex.From(tableName)
	if w := parm.Select; w != nil {
		qry.Select(w...)
	} else {
		qry.Select()
	}
	if w := parm.Where; w != nil {
		qry.Where(w)
	}
	if o := parm.Sort; len(o) > 0 {
		qry.OrderBy(o...)
	}
	if o := parm.Skip; o > 0 {
		qry.Skip(o)
	}
	if o := parm.Take; o > 0 {
		qry.Take(o)
	}
	if o := parm.GroupBy; len(o) > 0 {
		qry.GroupBy(o...)
	}
	if o := parm.Aggregates; len(o) > 0 {
		qry.Aggr(o...)
	}

	cur := conn.Cursor(qry, nil)
	if err = cur.Error(); err != nil {
		return fmt.Errorf("error when running cursor for PopulateByParm. %s", err.Error())
	}

	err = cur.Fetchs(dest, 0).Close()
	return err
}

/*
PopulateByFilter returns all data based on filter and n count.
No sort is given, hence it will take n data for resp filter
If n-count is 0, it will load all
*/
func (h *Hub) PopulateByFilter(tableName string, filter *dbflex.Filter, n int, dest interface{}) error {
	qp := dbflex.NewQueryParam().SetWhere(filter)
	if n > 0 {
		qp = qp.SetTake(n)
	}
	return h.PopulateByParm(tableName, qp, dest)
}

// PopulateSQL returns data based on SQL Query
func (h *Hub) PopulateSQL(sql string, dest interface{}) error {
	idx, conn, err := h.getConn()
	if err != nil {
		return fmt.Errorf("connection error. %s", err.Error())
	}
	defer h.closeConn(idx, conn)

	qry := dbflex.SQL(sql)
	cur := conn.Cursor(qry, nil)
	if err = cur.Error(); err != nil {
		return fmt.Errorf("error when running cursor for PopulateSQL. %s", err.Error())
	}

	err = cur.Fetchs(dest, 0).Close()
	return err
}

func (h *Hub) Close() {
	if h.usePool {
		h.pool.Close()
	}
}

// InsertAny insert any object into database table
func (h *Hub) InsertAny(name string, object interface{}) error {
	idx, conn, err := h.getConn()
	if err != nil {
		return fmt.Errorf("connection error. %s", err.Error())
	}
	defer h.closeConn(idx, conn)

	cmd := dbflex.From(name).Insert()
	if _, err = conn.Execute(cmd, codekit.M{}.Set("data", object)); err != nil {
		return fmt.Errorf("unable to save. %s", err.Error())
	}
	return nil
}

// SaveAny save any object into database table. Normally used with no-datamodel object
func (h *Hub) SaveAny(name string, filter *dbflex.Filter, object interface{}) error {
	idx, conn, err := h.getConn()
	if err != nil {
		return fmt.Errorf("connection error. %s", err.Error())
	}
	defer h.closeConn(idx, conn)

	if filter == nil {
		return errors.New("SaveAny should have valid filter")
	}

	cmd := dbflex.From(name).Where(filter).Save()
	if _, err = conn.Execute(cmd, codekit.M{}.Set("data", object)); err != nil {
		return fmt.Errorf("unable to save. %s", err.Error())
	}
	return nil
}

// UpdateAny update specific fields on database table. Normally used with no-datamodel object
func (h *Hub) UpdateAny(name string, filter *dbflex.Filter, object interface{}, fields ...string) error {
	idx, conn, err := h.getConn()
	if err != nil {
		return fmt.Errorf("connection error. %s", err.Error())
	}
	defer h.closeConn(idx, conn)

	if filter == nil {
		return errors.New("UpdateAny should have filter")
	}

	cmd := dbflex.From(name).Where(filter).Update(fields...)
	if _, err = conn.Execute(cmd, codekit.M{}.Set("data", object)); err != nil {
		return fmt.Errorf("unable to save. %s", err.Error())
	}
	return nil
}

// DeleteAny delete record on database table. Normally used with no-datamodel object
func (h *Hub) DeleteAny(name string, filter *dbflex.Filter) error {
	idx, conn, err := h.getConn()
	if err != nil {
		return fmt.Errorf("connection error. %s", err.Error())
	}
	defer h.closeConn(idx, conn)

	cmd := dbflex.From(name)
	if filter != nil {
		cmd.Where(filter)
	}
	cmd.Delete()

	if _, err = conn.Execute(cmd, nil); err != nil {
		return fmt.Errorf("unable to save. %s", err.Error())
	}
	return nil
}

// EnsureTable will ensure existense of table according to given object
func (h *Hub) EnsureTable(name string, keys []string, object interface{}) error {
	idx, conn, e := h.GetConnection()
	if e != nil {
		return e
	}
	defer h.CloseConnection(idx, conn)
	return conn.EnsureTable(name, keys, object)
}

// EnsureIndex ensure index existence
func (h *Hub) EnsureIndex(tableName, indexName string, unique bool, fields ...string) error {
	idx, conn, e := h.GetConnection()
	if e != nil {
		return e
	}
	defer h.CloseConnection(idx, conn)
	return conn.EnsureIndex(tableName, indexName, unique, fields...)
}

// EnsureDb
func (h *Hub) EnsureDb(obj orm.DataModel) error {
	idx, conn, e := h.GetConnection()
	if e != nil {
		return e
	}
	defer h.CloseConnection(idx, conn)
	return orm.EnsureDb(conn, obj)
}

// Validate validate if a connection can be established
func (h *Hub) Validate() error {
	idx, conn, e := h.GetConnection()
	if e != nil {
		return e
	}
	defer h.CloseConnection(idx, conn)
	return nil
}

func (h *Hub) GetByQuery(obj orm.DataModel, queryName string, param codekit.M) error {
	if param == nil {
		param = codekit.M{}
	}
	idx, conn, e := h.GetConnection()
	if e != nil {
		return e
	}
	defer h.CloseConnection(idx, conn)
	return orm.GetQuery(conn, obj, queryName, param)
}

func (h *Hub) GetsByQuery(obj orm.DataModel, queryName string, param codekit.M, target interface{}) error {
	if param == nil {
		param = codekit.M{}
	}
	idx, conn, e := h.GetConnection()
	if e != nil {
		return e
	}
	defer h.CloseConnection(idx, conn)
	return orm.GetsQuery(conn, obj, queryName, param, target)
}
