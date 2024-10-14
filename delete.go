package database

import (
	"database/sql"
	"fmt"
)

func (q *query) BuildDeleteQuery() (*sql.Stmt, error) {

	err := q.checkPreBuildErrors()
	if err != nil {
		return nil, err
	}

	if len(q.whereConds) == 0 {
		return nil, fmt.Errorf("where clause required for delete")
	}

	// Start building the query
	q.QueryString = fmt.Sprintf("DELETE FROM %s", q.table)

	// Add WHERE conditions
	q.addWhere()

	q.QueryString += ";"

	fmt.Println("delete query => ", q.QueryString, q.Args)

	stmt, err := q.conn.GetDB().Prepare(q.QueryString)
	if err != nil {
		return nil, err
	}

	return stmt, nil
}

func (q *query) DeleteOne(id string) (int64, error) {
	q.whereConds = []Condition{
		{Field: "_id", Operator: "=", Values: []any{id}},
	}

	count, err := q.Delete()
	if err != nil {
		return 0, err
	}

	if count < int64(1) {
		return 0, fmt.Errorf("could not find record")
	}

	return count, nil
}

func (q *query) Delete() (int64, error) {

	stmt, err := q.BuildDeleteQuery()
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(q.Args...)
	if err != nil {
		return 0, err
	}

	c, err := res.RowsAffected()
	return c, err
}
