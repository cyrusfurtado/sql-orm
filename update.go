package database

import (
	"database/sql"
	"fmt"
	"strings"

	"strconv"
)

func (q *query) Set(values map[string]any) QueryExecutor {
	l := len(values)

	if l > 0 {
		q.cols = make([]string, l)
		q.colValues = make([]any, l)

		i := 0
		for k, v := range values {
			if k == "" {
				q.errors = append(q.errors, fmt.Errorf("Set found empty key at position %v", i))
			}
			if v == "" {
				q.errors = append(q.errors, fmt.Errorf("Set found empty value at position %v", i))
			}
			q.cols[i] = k
			q.colValues[i] = v
			i++
		}
	}

	return q
}

func (q *query) setColumns() error {
	setCols := make([]string, len(q.cols))
	for i, v := range q.cols {
		q.ArgCount++

		if v == "" {
			nerr := fmt.Errorf("set columns found empty key at position %v", i)
			q.errors = append(q.errors, nerr)
			return nerr
		}

		setCols[i] = fmt.Sprintf("%s = $%s", v, strconv.Itoa(q.ArgCount))
	}

	q.QueryString += fmt.Sprintf(" SET %s", strings.Join(setCols, ", "))
	q.Args = q.colValues

	return nil
}

func (q *query) BuildUpdateQuery() (*sql.Stmt, error) {
	err := q.checkPreBuildErrors()
	if err != nil {
		return nil, err
	}

	if len(q.cols) != len(q.colValues) {
		return nil, fmt.Errorf("columns / values length mismatch")
	}

	// Start building the query
	q.QueryString = fmt.Sprintf("UPDATE %s", q.table)

	// set the columns to be updated
	err = q.setColumns()
	if err != nil {
		return nil, err
	}

	// Add WHERE conditions
	err = q.addWhere()
	if err != nil {
		return nil, err
	}

	q.QueryString += ";"

	fmt.Println("update query => ", q.QueryString, q.Args)

	stmt, err := q.conn.GetDB().Prepare(q.QueryString)
	if err != nil {
		return nil, err
	}

	return stmt, nil
}

func (q *query) Update() (int64, error) {

	stmt, err := q.BuildUpdateQuery()
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
