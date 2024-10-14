package database

import (
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestBasicUpdate(t *testing.T) {
	conn := &MockConnection{}

	err := conn.Connect()
	assert.Nil(t, err, "failed to connect to db instance")

	defer conn.Close()

	// UPDATE fire_teams SET description = 'Team 666' WHERE (  _id = `foxtrot` );
	queryString1 := `UPDATE fire_teams SET description = $1 WHERE (  _id = $2 );`
	queryArgs1 := []driver.Value{`Team 666`, `foxtrot`}
	query1, err := NewQuery(conn)
	assert.Nil(t, err)

	values := map[string]any{`description`: `Team 666`}

	mock := conn.GetMock()

	mock.ExpectPrepare(queryString1).WillBeClosed()
	mock.ExpectExec(queryString1).WithArgs(queryArgs1...).WillReturnResult(sqlmock.NewResult(0, 1))

	count, err := query1.Set(values).For("fire_teams").Where([]Condition{{
		Nested: &WhereGroup{
			Conditions: []Condition{
				{
					Field:    "_id",
					Operator: "=",
					Values:   []any{"foxtrot"},
				},
			},
		},
	}}).Update()

	assert.Nil(t, err)
	assert.Equal(t, int64(1), count)
}
