package database

import (
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestDelete(t *testing.T) {
	conn := &MockConnection{}

	err := conn.Connect()
	assert.Nil(t, err, "failed to connect to db instance")

	defer conn.Close()

	// DELETE FROM fire_teams WHERE (  _id = 'bravo' );
	delString := `DELETE FROM fire_teams WHERE (  _id = $1 );`
	delArgs := []driver.Value{`foxtrot`}
	query1, err := NewQuery(conn)
	assert.Nil(t, err)

	mock := conn.GetMock()

	mock.ExpectPrepare(delString).WillBeClosed()
	mock.ExpectExec(delString).WithArgs(delArgs...).WillReturnResult(sqlmock.NewResult(0, 1))

	count, err := query1.For("fire_teams").Where([]Condition{{
		Nested: &WhereGroup{
			Conditions: []Condition{
				{
					Field:    "_id",
					Operator: "=",
					Values:   []any{"foxtrot"},
				},
			},
		},
	}}).Delete()

	assert.Nil(t, err)
	assert.Equal(t, int64(1), count)
}
