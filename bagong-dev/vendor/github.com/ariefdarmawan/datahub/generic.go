package datahub

import (
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/sebarcode/codekit"
)

// Get return single data, generic mode
func Get[T orm.DataModel](h *Hub, model T) (T, error) {
	err := h.Get(model)
	return model, err
}

// GetWithoutCache return single data ignoring cache, generic mode
func GetWithoutCache[T orm.DataModel](h *Hub, model T) (T, error) {
	err := h.GetWithoutCache(model)
	return model, err
}

// GetByFilter return first record by filter
func GetByFilter[T orm.DataModel](h *Hub, model T, filter *dbflex.Filter) (T, error) {
	err := h.GetByFilter(model, filter)
	return model, err
}

// GetByParm return first record by param
func GetByParm[T orm.DataModel](h *Hub, model T, qp *dbflex.QueryParam) (T, error) {
	err := h.GetByParm(model, qp)
	return model, err
}

// GetByQuery return first record by query
func GetByQuery[T orm.DataModel](h *Hub, model T, queryName string, payload codekit.M) (T, error) {
	err := h.GetByQuery(model, queryName, payload)
	return model, err
}

// GetByID return first record by theirs ID
func GetByID[T orm.DataModel](h *Hub, model T, ids ...interface{}) (T, error) {
	err := h.GetByID(model, ids...)
	return model, err
}

// GetByIDWithoutCache return first record by theirs ID, ignoring cache
func GetByIDWithoutCache[T orm.DataModel](h *Hub, model T, ids ...interface{}) (T, error) {
	model.SetID(ids...)
	err := h.GetWithoutCache(model)
	return model, err
}

// Find return all data based on model and filter, generic mode
func Find[T orm.DataModel](h *Hub, model T, param *dbflex.QueryParam) ([]T, error) {
	dest := []T{}
	err := h.Gets(model, param, &dest)
	return dest, err
}

// FindByFilter like find but require only filter, generic mode
func FindByFilter[T orm.DataModel](h *Hub, model T, filter *dbflex.Filter) ([]T, error) {
	dest := []T{}
	err := h.GetsByFilter(model, filter, &dest)
	return dest, err
}

// FindByCommand returns all data using dbflex command, generic mode
func FindByCommand[T orm.DataModel](h *Hub, data T, cmd dbflex.ICommand) ([]T, error) {
	dest := []T{}
	_, err := h.Populate(cmd, &dest)
	return dest, err
}

// Count returns number of data based on model and filter, generic mode
func Count[T orm.DataModel](h *Hub, data T, qp *dbflex.QueryParam) (int, error) {
	return h.Count(data, qp)
}

// FindAnyByCommand returns all data using dbflex command
func FindAnyByCommand[T any](h *Hub, model T, cmd dbflex.ICommand) ([]T, error) {
	dest := []T{}
	_, err := h.Populate(cmd, &dest)
	return dest, err
}

func FindAnyByParm[T any](h *Hub, model T, tablename string, qp *dbflex.QueryParam) ([]T, error) {
	dest := []T{}
	err := h.PopulateByParm(tablename, qp, &dest)
	return dest, err
}

func FindAnyByFilter[T any](h *Hub, model T, tablename string, filter *dbflex.Filter, n int) ([]T, error) {
	dest := []T{}
	err := h.PopulateByFilter(tablename, filter, n, &dest)
	return dest, err
}

func FindAnyBySQL[T any](h *Hub, model T, sql string) ([]T, error) {
	dest := []T{}
	err := h.PopulateSQL(sql, &dest)
	return dest, err
}

// Save
func Save(db *Hub, records ...orm.DataModel) error {
	for _, r := range records {
		if e := db.Insert(r); e != nil {
			_, ids := r.GetID(nil)
			return fmt.Errorf("save %s: %v: %s", r.TableName(), ids, e.Error())
		}
	}

	return nil
}

// Insert
func Insert(db *Hub, records ...orm.DataModel) error {
	for _, r := range records {
		if e := db.Insert(r); e != nil {
			_, ids := r.GetID(nil)
			return fmt.Errorf("insert %s: %v: %s", r.TableName(), ids, e.Error())
		}
	}

	return nil
}

// Update
func Update(db *Hub, records ...orm.DataModel) error {
	for _, r := range records {
		if e := db.Update(r); e != nil {
			_, ids := r.GetID(nil)
			return fmt.Errorf("update %s: %v: %s", r.TableName(), ids, e.Error())
		}
	}

	return nil
}

// UpdateFields
func UpdateFields(db *Hub, fields []string, records ...orm.DataModel) error {
	for _, r := range records {
		if e := db.Update(r, fields...); e != nil {
			_, ids := r.GetID(nil)
			return fmt.Errorf("update %s: %v: %s", r.TableName(), ids, e.Error())
		}
	}
	return nil
}

// TxDone to commit or rollback based on error. It will raise error if the db object is not tx support
func TxDone(db *Hub, err error, ignoreTxSupport bool) error {
	if !db.IsTx() {
		if !ignoreTxSupport {
			return fmt.Errorf("IsNotTx")
		}
		return nil
	}

	if err != nil {
		if err := db.Rollback(); err != nil {
			return fmt.Errorf("rollback fail: %s", err.Error())
		}
	}

	if err := db.Commit(); err != nil {
		return fmt.Errorf("commit fail: %s", err.Error())
	}

	return nil
}
