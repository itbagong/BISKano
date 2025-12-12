package datahub

import (
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/sebarcode/codekit"
)

// Get return single data, generic mode
func Get[T orm.DataModel](h *Hub, model T) (T, error) {
	err := h.Get(model)
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

// GetByParmEl return first record by param element
func GetByParmEl[T orm.DataModel](h *Hub, model T, where *dbflex.Filter, order string, fields ...string) (T, error) {
	err := h.GetByParmEl(model, model, where, order, fields...)
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
