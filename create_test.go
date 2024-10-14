package database

import (
	"database/sql/driver"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestFailureCreate(t *testing.T) {
	conn := &MockConnection{}

	err := conn.Connect()
	assert.Nil(t, err, "failed to connect to mock db instance")

	mock := conn.GetMock()

	defer conn.Close()

	errs := []error{fmt.Errorf("voilates unique constraint"), fmt.Errorf("columns / values length mismatch")}

	// INSERT into fire_teams (_id, description) VALUES ('golf', 'Team 7');
	queryString1 := `INSERT INTO fire_teams (_id, description) VALUES ($1, $2);`
	queryArgs1 := []driver.Value{`alpha`, `Team 1`}
	query1, err := NewQuery(conn)
	assert.Nil(t, err)

	// &Query{
	// 	Table:     "fire_teams",
	// 	Cols:      []string{`_id`, `description`},
	// 	ColValues: []any{`alpha`, `Team 1`},
	// }

	values1 := map[string]any{`_id`: `alpha`, `description`: `Team 1`}
	mock.ExpectPrepare(queryString1).WillBeClosed()
	mock.ExpectExec(queryString1).WithArgs(queryArgs1...).WillReturnError(errs[0])

	res, err := query1.Set(values1).For("fire_teams").Create()
	assert.Equal(t, errs[0].Error(), err.Error(), "should return a constraint error")
	assert.Equal(t, int64(0), res, "should not return result")
}

func TestSuccessCreate(t *testing.T) {
	conn := &MockConnection{}

	err := conn.Connect()
	assert.Nil(t, err, "failed to connect to mock db instance")

	defer conn.Close()

	successQueryStr := `INSERT INTO fire_teams (_id, description) VALUES ($1, $2);`
	successQueryArgs := []driver.Value{`india`, `Team 99`}
	successQuery, err := NewQuery(conn)
	assert.Nil(t, err)

	mock := conn.GetMock()
	mock.ExpectPrepare(successQueryStr).WillBeClosed()
	mock.ExpectExec(successQueryStr).WithArgs(successQueryArgs...).WillReturnResult(sqlmock.NewResult(1, 1))

	values := map[string]any{`_id`: `india`, `description`: `Team 99`}
	count, _ := successQuery.Set(values).For("fire_teams").Create()

	// assert.Nil(t, err)
	assert.Equal(t, int64(1), count)
}
