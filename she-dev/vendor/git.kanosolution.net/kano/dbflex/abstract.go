package dbflex

import "github.com/sebarcode/codekit"

func Gets(conn IConnection, buffer interface{}, tableName string, qp *QueryParam) error {
	cmd, _ := qp.ToCommand(From(tableName))
	cursor := conn.Cursor(cmd, nil)
	err := cursor.Fetchs(buffer, 0).Close()
	return err
}

func GetsWhere(conn IConnection, buffer interface{}, tableName string, where *Filter, orderBy ...string) error {
	qp := NewQueryParam()
	if where != nil {
		qp.SetWhere(where)
	}
	if len(orderBy) == 0 {
		qp.SetSort(orderBy...)
	}
	return Gets(conn, buffer, tableName, qp)
}

func Get(conn IConnection, buffer interface{}, tableName string, qp *QueryParam) error {
	cmd, _ := qp.ToCommand(From(tableName))
	cursor := conn.Cursor(cmd, nil)
	err := cursor.Fetch(buffer).Close()
	return err
}

func GetWhere(conn IConnection, buffer interface{}, tableName string, where *Filter, orderBy ...string) error {
	qp := NewQueryParam()
	if where != nil {
		qp.SetWhere(where)
	}
	if len(orderBy) == 0 {
		qp.SetSort(orderBy...)
	}
	return Get(conn, buffer, tableName, qp)
}

func Insert(conn IConnection, buffer interface{}, tableName string) error {
	cmd := From(tableName).Insert()
	_, err := conn.Execute(cmd, codekit.M{}.Set("data", buffer))
	return err
}

func Update(conn IConnection, buffer interface{}, tableName string, where *Filter, fields ...string) error {
	cmd := From(tableName).Update(fields...)
	if where != nil {
		cmd.Where(where)
	}
	_, err := conn.Execute(cmd, codekit.M{}.Set("data", buffer))
	return err
}

func Delete(conn IConnection, buffer interface{}, tableName string, where *Filter) error {
	cmd := From(tableName).Delete()
	if where != nil {
		cmd.Where(where)
	}
	_, err := conn.Execute(cmd, nil)
	return err
}
