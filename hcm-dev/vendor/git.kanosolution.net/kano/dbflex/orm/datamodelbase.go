package orm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"github.com/sebarcode/codekit"
)

type MetaField struct {
	Name   string
	DbName string
	Type   reflect.Type
}

// DataModelBase is a base struct for easier implementation of DataModel interface
type DataModelBase struct {
	self DataModel

	hasBeenCalc bool
	meta        map[string]MetaField
}

// FK return FKConfig for data model
func (dm *DataModelBase) FK() []*FKConfig {
	if UseRelationManager {
		return DefaultRelationManager().FKs(dm.This())
	}
	return []*FKConfig{}
}

// ReverseFK return FKConfig for data model
func (dm *DataModelBase) ReverseFK() []*ReverseFKConfig {
	if UseRelationManager {
		return DefaultRelationManager().ReverseFKs(dm.This())
	}
	return []*ReverseFKConfig{}
}

// NewDataModel abstraction to create new data model object
func NewDataModel(m DataModel) DataModel {
	m.SetThis(m)
	return m
}

func getFieldInfo(d *DataModelBase, fieldNameTag string) {
	if !d.hasBeenCalc {
		v := reflect.ValueOf(d.This())
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		t := v.Type()
		fieldNum := t.NumField()
		fields := make(map[string]MetaField, fieldNum)
		for i := 0; i < fieldNum; i++ {
			f := t.Field(i)
			tag := f.Tag
			dbname := strings.ToLower(f.Name)
			nameTag := tag.Get(fieldNameTag)
			if nameTag != "" {
				dbname = nameTag
			}
			fields[f.Name] = MetaField{
				Name:   f.Name,
				DbName: dbname,
				Type:   f.Type,
			}
		}
		d.meta = fields
		d.hasBeenCalc = true
	}
}

func getFieldDbName(d *DataModelBase, name, fieldNameTag string) string {
	getFieldInfo(d, fieldNameTag)
	mf, ok := d.meta[name]
	if !ok {
		return ""
	}
	return mf.DbName
}

func getFieldType(d *DataModelBase, name, fieldNameTag string) reflect.Type {
	getFieldInfo(d, fieldNameTag)
	mf, ok := d.meta[name]
	if !ok {
		return reflect.TypeOf("")
	}
	return mf.Type
}

// SetThis to create circle loop to refer to datamodel
func (d *DataModelBase) SetThis(m DataModel) {
	d.self = m
}

// This to get abstraction of datamodel object
func (d *DataModelBase) This() DataModel {
	if d.self == nil {
		return d
	}
	return d.self

}

// Get filter ID
func (d *DataModelBase) TableName() string {
	naturalName := strings.ToLower(reflect.TypeOf(d.This()).Elem().Name())
	switch naturalName[len(naturalName)-1] {
	case 'y':
		return naturalName[0:len(naturalName)-1] + "ies"

	case 's':
		return naturalName + "es"
	}
	return naturalName + "s"
}

// Get filter ID
func (d *DataModelBase) GetFilterID(conn dbflex.IConnection, ids ...interface{}) *dbflex.Filter {
	//panic("GetFilterID not implemented")
	fields, keys := d.This().GetID(conn)
	filters := []*dbflex.Filter{}
	for idx, field := range fields {
		if idx < len(keys) {
			filters = append(filters, dbflex.Eq(field, keys[idx]))
		}
	}
	return dbflex.And(filters...)
}

// Get where filter from any field that not blank
func (d *DataModelBase) GetWhereFilter(codekit.M) *dbflex.Filter {
	panic("GetWhereFilter not implemented")
}

// GetID to get ID
func (d *DataModelBase) GetID(conn dbflex.IConnection) ([]string, []interface{}) {
	fieldtag := ""
	keynametag := ""
	if conn != nil {
		fieldtag = conn.FieldNameTag()
		keynametag = conn.KeyNameTag()
	}

	if keynametag == "" {
		keynametag = "key"
	}

	v := reflect.ValueOf(d.This()).Elem()
	t := v.Type()

	ids := []string{}
	values := []interface{}{}
	fieldNum := t.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		tf := t.Field(idx)
		fn := strings.Split(tf.Tag.Get(fieldtag), ",")[0]
		if fn == "" {
			fn = tf.Name
		}
		iskey := tf.Tag.Get(keynametag)

		//fmt.Println(t.Name, "|", tf.Name, "|", fn, "|", iskey)
		if iskey != "" {
			ids = append(ids, fn)
			values = append(values, v.Field(idx).Interface())
		}
	}

	if len(ids) == 0 {
		panic("GetID can't be applied for " + t.Name() + ", please check your object definition.")
	}

	return ids, values
}

// SetID is not used yet
func (d *DataModelBase) SetID(keys ...interface{}) {
	//-- do nothing
	panic("SetID is not yet implemented for this DataModel")
}

// SetObjectID is not used yet
func (d *DataModelBase) SetObjectID(keys ...interface{}) DataModel {
	d.This().SetID(keys...)
	return d.This()
}

// PostSave run after data is saved (insert, update, save)
func (d *DataModelBase) PostSave(conn dbflex.IConnection) error {
	//-- do nothing
	return nil
}

// PreSave run before data being saved (insert, update, save)
func (d *DataModelBase) PreSave(conn dbflex.IConnection) error {
	//-- do nothing
	return nil
}

// PreDelete run before data being deleted
func (d *DataModelBase) PreDelete(conn dbflex.IConnection) error {
	//-- do nothing
	return nil
}

// PostDelete run after data is deleted
func (d *DataModelBase) PostDelete(conn dbflex.IConnection) error {
	//-- do nothing
	return nil
}

// Indexes define index on database
func (d *DataModelBase) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{}
}

// EnsureDb to ensure table structure on database is per expected
func EnsureDb(conn dbflex.IConnection, o DataModel) error {
	o.SetThis(o)
	idFields, _ := o.GetID(conn)
	if e := conn.EnsureTable(o.TableName(), idFields, o); e != nil {
		return fmt.Errorf("fail ensure table struct. %s", e.Error())
	}
	for _, index := range o.Indexes() {
		if e := conn.EnsureIndex(o.TableName(), index.Name, index.IsUnique, index.Fields...); e != nil {
			return fmt.Errorf("fail ensure index. %s", e.Error())
		}
	}
	return nil
}

// Get a single data from given connection and model
// For single data you can make filter directly inside the model itself,
// the downside of this feature is currently you cannot use dbflex.And, Or, Range on the same field
func Get(conn dbflex.IConnection, model DataModel) error {
	model.SetThis(model)
	tablename := model.TableName()
	where := generateFilterFromDataModel(conn, model)
	cursor := conn.Cursor(dbflex.From(tablename).Select().Where(where), codekit.M{})
	defer cursor.Close()
	e := cursor.Fetch(model).Error()
	//fmt.Println(tablename, "where:", codekit.JsonString(where), "model:", codekit.JsonString(model))
	return e
	//return fmt.Errorf("asdada")
}

// GetFirst get first document, optionally based on filter and order
func GetFirst(conn dbflex.IConnection, model DataModel, where *dbflex.Filter, order string) error {
	model.SetThis(model)
	tablename := model.TableName()
	cmd := dbflex.From(tablename).Select().Take(1)
	if where != nil {
		cmd.Where(where)
	}
	if order != "" {
		cmd.OrderBy(order)
	}
	return conn.Cursor(cmd, nil).Fetch(model).Close()
}

// GetWhere get a single datamodel
func GetWhere(conn dbflex.IConnection, model DataModel, where *dbflex.Filter) error {
	return GetFirst(conn, model, where, "")
}

// GetQuery get single record using pre-defined query
func GetQuery(conn dbflex.IConnection, model DataModel, queryName string, queryParam codekit.M) error {
	query, hasQuery := model.Queries()[queryName]
	if !hasQuery {
		return errors.New("invalid query")
	}

	qp, err := makeQp(query, queryParam)
	if err != nil {
		return err
	}

	cmd := qpToCmd(model.TableName(), qp)
	return conn.Cursor(cmd, nil).Fetch(model).Close()
}

func makeQp(query Query, queryParam codekit.M) (*dbflex.QueryParam, error) {
	if len(query.Param.Select) == 0 && queryParam.Has("Select") {
		query.Param.Select = queryParam.Get("Select", []string{}).([]string)
	}

	if len(query.Param.Sort) == 0 && queryParam.Has("Sort") {
		query.Param.Sort = queryParam.Get("Sort", []string{}).([]string)
	}

	if query.Param.Take == 0 && queryParam.Has("Take") {
		query.Param.Take = queryParam.Get("Take", 0).(int)
	}

	if query.Param.Skip == 0 && queryParam.Has("Skip") {
		query.Param.Skip = queryParam.Get("Skip", 0).(int)
	}

	if query.Param.Where != nil {
		whereString := codekit.JsonString(query.Param.Where)
		hasParm := false
		for k, v := range queryParam {
			if k == "Sort" || k == "Take" || k == "Skip" {
				continue
			}
			hasParm = true
			vString := codekit.ToString(v)
			whereString = strings.Replace(whereString, fmt.Sprintf("$(%s)", k), vString, -1)
		}

		if hasParm {
			f := dbflex.Filter{}
			if e := codekit.UnjsonFromString(whereString, &f); e != nil {
				return nil, fmt.Errorf("fail to create where. %s", e.Error())
			}
			query.Param.Where = &f
		}
	}
	return query.Param, nil
}

// Gets multiple data from given connection, model, buffer, and query param
func Gets(conn dbflex.IConnection, model DataModel, buffer interface{}, qp *dbflex.QueryParam) error {
	model.SetThis(model)
	tablename := model.TableName()

	if qp == nil {
		qp = dbflex.NewQueryParam()
	}

	cmd := qpToCmd(tablename, qp)
	cursor := conn.Cursor(cmd, nil)
	defer cursor.Close()
	return cursor.Fetchs(buffer, 0).Error()
}

// GetsWhere get multiple data from given connection, model, buffer, and filter
func GetsWhere(conn dbflex.IConnection, model DataModel, buffer interface{}, filter *dbflex.Filter, orderBy string) error {
	qp := dbflex.NewQueryParam().SetWhere(filter)
	if orderBy != "" {
		qp.SetSort(orderBy)
	}
	return Gets(conn, model, buffer, qp)
}

// GetsQuery get multi record using pre-defined query
func GetsQuery(conn dbflex.IConnection, model DataModel, queryName string, queryParam codekit.M, buffer interface{}) error {
	query, hasQuery := model.Queries()[queryName]
	if !hasQuery {
		return errors.New("invalid query")
	}

	qp, err := makeQp(query, queryParam)
	if err != nil {
		return err
	}

	cmd := qpToCmd(model.TableName(), qp)
	cursor := conn.Cursor(cmd, nil)
	defer cursor.Close()
	return cursor.Fetchs(buffer, 0).Error()
}

// CountQuery get number of record using pre-defined query
func CountQuery(conn dbflex.IConnection, model DataModel, queryName string, queryParam codekit.M) (int, error) {
	query, hasQuery := model.Queries()[queryName]
	if !hasQuery {
		return 0, errors.New("invalid query")
	}

	qp, err := makeQp(query, queryParam)
	if err != nil {
		return 0, err
	}
	qp.SetTake(0)
	qp.SetSkip(0)

	cmd := qpToCmd(model.TableName(), qp)
	cursor := conn.Cursor(cmd, nil)
	defer cursor.Close()
	return cursor.Count(), nil
}

func qpToCmd(tablename string, qp *dbflex.QueryParam) dbflex.ICommand {
	cmd := dbflex.From(tablename)
	if len(qp.Aggregates) > 0 {
		if len(qp.GroupBy) > 0 {
			cmd.GroupBy(qp.GroupBy...)
		}
		cmd.Aggr(qp.Aggregates...)
	} else if qp.Command != nil {
		cmd.Command(qp.Command.Command, qp.Command.Parm)
	} else {
		if len(qp.Select) == 0 {
			cmd = cmd.Select()
		} else {
			cmd = cmd.Select(qp.Select...)
		}
	}
	if qp != nil {
		if qp.Where != nil {
			cmd.Where(qp.Where)
		}

		if len(qp.Sort) > 0 {
			cmd.OrderBy(qp.Sort...)
		}

		if qp.Skip > 0 {
			cmd.Skip(qp.Skip)
		}

		if qp.Take > 0 {
			cmd.Take(qp.Take)
		}
	}
	return cmd
}

// Insert new data from given connection and data model
func Insert(conn dbflex.IConnection, dm DataModel) error {
	dm.SetThis(dm)
	tablename := dm.TableName()

	err := dm.PreSave(conn)
	if err != nil {
		return errors.New(fmt.Sprint("dbflex ", tablename+".PreSave ", err.Error()))
	}

	err = checkFK(conn, dm)
	if err != nil {
		return errors.New(fmt.Sprint("dbflex ", tablename+".FK ", err.Error()))
	}

	_, err = conn.Execute(
		dbflex.From(tablename).Insert(),
		codekit.M{}.Set("data", dm))

	if err == nil {
		err = dm.PostSave(conn)
		if err != nil {
			return errors.New(fmt.Sprint("dbflex ", tablename+".PostSave ", err.Error()))
		}
	}
	return err
}

// Save will check if given DataModel filter is exist in the table or not
// If exist then update the data if not insert new data
func Save(conn dbflex.IConnection, dm DataModel, fields ...string) error {
	dm.SetThis(dm)
	tablename := dm.TableName()
	filter := generateFilterFromDataModel(conn, dm)

	dmexist := codekit.M{}
	cursor := conn.Cursor(dbflex.From(tablename).Select().Where(filter), nil)
	errexist := cursor.Fetch(&dmexist).Close()

	err := dm.PreSave(conn)
	if err != nil {
		return errors.New(fmt.Sprint("dbflex ", tablename+".PreSave ", err.Error()))
	}

	err = checkFK(conn, dm)
	if err != nil {
		return errors.New(fmt.Sprint("dbflex ", tablename+".FK ", err.Error()))
	}

	if errexist == nil {
		_, err = conn.Execute(dbflex.From(tablename).Where(filter).Update(fields...),
			codekit.M{}.Set("data", dm))
	} else {
		_, err = conn.Execute(dbflex.From(tablename).Insert(),
			codekit.M{}.Set("data", dm))
	}

	if err == nil {
		err = dm.PostSave(conn)
		if err != nil {
			return errors.New(fmt.Sprint("dbflex ", tablename+".PostSave ", err.Error()))
		}
	}

	errFK := updateReverseFK(conn, dm)
	if errFK != nil {
		return errors.New(fmt.Sprint("dbflex ", tablename+".UpdReverseFK ", errFK.Error(), " ", err.Error()))
	}

	return err
}

// Update  data from given connection and DataModel,
// filter is generated from given DataModel
func Update(conn dbflex.IConnection, dm DataModel, fields ...string) error {
	dm.SetThis(dm)
	tablename := dm.TableName()
	filter := generateFilterFromDataModel(conn, dm)

	err := dm.PreSave(conn)
	if err != nil {
		return errors.New(fmt.Sprint("dbflex ", tablename+".PreSave ", err.Error()))
	}

	err = checkFK(conn, dm)
	if err != nil {
		return errors.New(fmt.Sprint("dbflex ", tablename+".FK ", err.Error()))
	}

	_, err = conn.Execute(
		dbflex.From(tablename).Where(filter).Update(fields...),
		codekit.M{}.Set("data", dm))

	if err == nil {
		err = dm.PostSave(conn)
		if err != nil {
			return errors.New(fmt.Sprint("dbflex ", tablename+".PostSave ", err.Error()))
		}
	}

	errFK := updateReverseFK(conn, dm)
	if errFK != nil {
		return errors.New(fmt.Sprint("dbflex ", tablename+".UpdReverseFK ", errFK.Error(), " ", err.Error()))
	}

	return err
}

// Delete data from given connection and DataModel,
// filter is generated from given DataModel
func Delete(conn dbflex.IConnection, dm DataModel) error {
	dm.SetThis(dm)
	tablename := dm.TableName()
	filter := generateFilterFromDataModel(conn, dm)

	err := dm.PreDelete(conn)
	if err != nil {
		return errors.New(fmt.Sprint("dbflex ", tablename+".PreDelete ", err.Error()))
	}

	err = checkEmptyFK(conn, dm)
	if err != nil {
		return errors.New(fmt.Sprint("dbflex ", tablename+".PreDelete ", err.Error()))
	}

	_, err = conn.Execute(dbflex.From(tablename).Where(filter).Delete(), nil)

	if err == nil {
		err = dm.PostDelete(conn)
		if err != nil {
			return errors.New(fmt.Sprint("dbflex ", tablename+".PostDelete ", err.Error()))
		}
	}

	return err
}

func (o *DataModelBase) Queries() map[string]Query {
	return map[string]Query{}
}

func generateFilterFromDataModel(conn dbflex.IConnection, dm DataModel) *dbflex.Filter {
	fields, values := dm.GetID(conn)
	if len(fields) == 0 {
		return new(dbflex.Filter)
	}

	fieldNameTag := conn.FieldNameTag()
	useTag := fieldNameTag != ""

	if useTag {
		vt := reflect.Indirect(reflect.ValueOf(dm)).Type()

		eqs := []*dbflex.Filter{}
		for idx, field := range fields {
			if f, ok := vt.FieldByName(field); ok {
				if tag, ok := f.Tag.Lookup(fieldNameTag); ok {
					eqs = append(eqs, dbflex.Eq(strings.Split(tag, ",")[0], values[idx]))
				} else {
					eqs = append(eqs, dbflex.Eq(field, values[idx]))
				}
			} else {
				eqs = append(eqs, dbflex.Eq(field, values[idx]))
			}
		}

		if len(eqs) == 1 {
			return eqs[0]
		}

		return dbflex.And(eqs...)
	}

	if len(fields) == 1 {
		return dbflex.Eq(fields[0], values[0])
	}

	eqs := []*dbflex.Filter{}
	for idx, field := range fields {
		eqs = append(eqs, dbflex.Eq(field, values[idx]))
	}
	return dbflex.And(eqs...)
}

func GetFieldName(obj interface{}, name, tag string, conn dbflex.IConnection) string {
	if tag == "" {
		tag = conn.FieldNameTag()
	}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	if f, ok := t.FieldByName(name); !ok {
		return ""
	} else {
		if tag == "" {
			return f.Name
		}
		if tagName := f.Tag.Get(tag); tagName == "" {
			return f.Name
		} else {
			return tagName
		}
	}
}
