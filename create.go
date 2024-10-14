package database

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

func (q *query) makePlaceholders(n int) ([]string, error) {
	placeholders := make([]string, n)
	for i := range placeholders {
		q.ArgCount++

		if q.cols[i] == "" {
			nerr := fmt.Errorf("column missing")
			q.errors = append(q.errors, nerr)

			return placeholders, nerr
		}

		placeholders[i] = "$" + strconv.Itoa(q.ArgCount)
	}

	q.Args = q.colValues
	return placeholders, nil
}

func (q *query) BuildInsertQuery() (*sql.Stmt, error) {

	err := q.checkPreBuildErrors()
	if err != nil {
		return nil, err
	}

	if len(q.cols) != len(q.colValues) {
		return nil, fmt.Errorf("columns / values length mismatch")
	}

	placeholders, err := q.makePlaceholders(len(q.cols))
	if err != nil {
		return nil, err
	}

	q.QueryString = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", q.table, strings.Join(q.cols, ", "), strings.Join(placeholders, ", "))

	q.QueryString += ";"

	fmt.Println("insert query => ", q.QueryString, q.Args, q.ArgCount)

	stmt, err := q.conn.GetDB().Prepare(q.QueryString)
	if err != nil {
		return nil, err
	}

	return stmt, nil
}

func (q *query) Create() (int64, error) {

	stmt, err := q.BuildInsertQuery()
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
