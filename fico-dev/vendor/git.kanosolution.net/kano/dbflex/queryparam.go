package dbflex

import (
	"errors"

	"github.com/sebarcode/codekit"
)

// qpCommand external command not relevant to sql
type qpCommand struct {
	Command string
	Parm    interface{}
}

// QueryParam is query paramater like Where, Sort, Take, and Skip
type QueryParam struct {
	Where      *Filter
	Sort       []string
	Take, Skip int
	GroupBy    []string
	Select     []string
	Aggregates []*AggrItem
	Command    *qpCommand
	Param      codekit.M
}

// NewQueryParam create new QueryParam
func NewQueryParam() *QueryParam {
	return new(QueryParam)
}

// SetSelect setter for select
func (q *QueryParam) SetSelect(fields ...string) *QueryParam {
	q.Select = fields
	return q
}

// SetWhere setter for Where field
func (q *QueryParam) SetWhere(f *Filter) *QueryParam {
	q.Where = f
	return q
}

// SetWheres setter for Where field with none or multi filters
func (q *QueryParam) SetWheres(filters ...*Filter) *QueryParam {
	if len(filters) == 1 {
		q.SetWhere(filters[0])
	} else if len(filters) > 1 {
		q.SetWhere(And(filters...))
	}
	return q
}

// SetSort setter for Sort field
func (q *QueryParam) SetSort(sorts ...string) *QueryParam {
	q.Sort = sorts
	return q
}

// SetTake setter for Take field
func (q *QueryParam) SetTake(take int) *QueryParam {
	q.Take = take
	return q
}

// SetSkip setter for Skip field
func (q *QueryParam) SetSkip(skip int) *QueryParam {
	q.Skip = skip
	return q
}

// SetGroupBy setter for group by
func (q *QueryParam) SetGroupBy(gs ...string) *QueryParam {
	q.GroupBy = gs
	return q
}

// SetAggr setter for aggregates
func (q *QueryParam) SetAggr(aggrs ...*AggrItem) *QueryParam {
	q.Aggregates = aggrs
	return q
}

// MergeWhere merge current where with new filter
func (q *QueryParam) MergeWhere(isOr bool, fs ...*Filter) *QueryParam {
	if q.Where == nil {
		if len(fs) == 1 {
			q.Where = fs[0]
		} else if len(fs) > 1 && isOr {
			q.Where = Or(fs...)
		} else if len(fs) > 1 && !isOr {
			q.Where = And(fs...)
		}
	} else if isOr {
		fs = append(fs, q.Where)
		q.Where = Or(fs...)
	} else {
		fs = append(fs, q.Where)
		q.Where = And(fs...)
	}
	return q
}

// SetParam set parameter, it will be translated to Command.SetAttr()
func (q *QueryParam) SetParam(k string, v interface{}) *QueryParam {
	if q.Param == nil {
		q.Param = codekit.M{}
	}
	q.Param.Set(k, v)
	return q
}

// ToCommand quick method to trabslate to command
func (q *QueryParam) ToCommand(cmd ICommand) (ICommand, error) {
	if cmd == nil {
		return cmd, errors.New("command is nil")
	}

	if len(q.Select) > 0 {
		cmd.Select(q.Select...)
	}

	if q.Where != nil {
		cmd.Where(q.Where)
	}

	if len(q.Sort) > 0 {
		cmd.OrderBy(q.Sort...)
	}

	if len(q.GroupBy) > 0 {
		cmd.GroupBy(q.GroupBy...)
	}

	if len(q.Aggregates) > 0 {
		cmd.Aggr(q.Aggregates...)
	}

	if q.Skip > 0 {
		cmd.Skip(q.Skip)
	}

	if q.Take > 0 {
		cmd.Take(q.Take)
	}

	if q.Param != nil && len(q.Param) > 0 {
		for k, v := range q.Param {
			cmd.SetAttr(k, v)
		}
	}

	if q.Command != nil {
		cmd.Command(q.Command.Command, q.Command.Parm)
	}

	return cmd, nil
}
