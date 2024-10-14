package database

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DATA-DOG/go-sqlmock"
)

// Comparison Operators

// =: Equal to
// <>: Not equal to
// <: Less than
// >: Greater than
// <=: Less than or equal to
// >=: Greater than or equal to

// Logical Operators

// AND: Combines two conditions, both must be true
// OR: Combines two conditions, at least one must be true
// NOT: Negates a condition

// Pattern Matching Operators

// LIKE: Matches patterns
// ILIKE: Case-insensitive pattern matching
// SIMILAR TO: More complex pattern matching using regular expressions

// Range Operators

// BETWEEN: Checks if a value is within a range
// IN: Checks if a value is in a list
// NOT IN: Checks if a value is not in a list

// Null Value Operators

// IS NULL: Checks for null values
// IS NOT NULL: Checks for non-null values

// Other Operators

// EXISTS: Checks if a subquery returns any rows
// NOT EXISTS: Checks if a subquery returns no rows

//  Aggregate functions

// COUNT(), SUM(), AVG()
// IN BETWEEN
// NOT
// LIKE

// integration test variables
const (
	host     = "localhost"
	database = "testdb"
	port     = "5432"
	user     = "cfurtado"
	pass     = "$Ilovecrowdstrike2021"
	ssl      = "disable"
)

type MockConnection struct {
	db   *sql.DB // db instance handle
	mock sqlmock.Sqlmock
}

func (m *MockConnection) Connect() error {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		return err
	}

	m.db = db
	m.mock = mock
	return nil
}

func (m *MockConnection) Close() error {
	m.db.Close()
	return nil
}

func (m *MockConnection) GetDB() *sql.DB {
	return m.db
}

func (m *MockConnection) GetMock() sqlmock.Sqlmock {
	return m.mock
}

// integration test for database connection
func TestConnection(t *testing.T) {

	// failure cases
	conn := NewConnection("", database, user, pass, port, ssl)
	err := conn.Connect()
	assert.NotNil(t, err, "host is empty")
	assert.Contains(t, err.Error(), "host not set")

	conn = NewConnection(host, "", user, pass, port, ssl)
	err = conn.Connect()
	assert.NotNil(t, err, "db is empty")
	assert.Contains(t, err.Error(), "database not set")

	conn = NewConnection(host, database, "", pass, port, ssl)
	err = conn.Connect()
	assert.NotNil(t, err, "user is empty")
	assert.Contains(t, err.Error(), "user not set")

	conn = NewConnection(host, database, user, "", port, ssl)
	err = conn.Connect()
	assert.NotNil(t, err, "pass is empty")
	assert.Contains(t, err.Error(), "creds not set")

	conn = NewConnection(host, database, user, pass, "", ssl)
	err = conn.Connect()
	assert.NotNil(t, err, "port is empty")
	assert.Contains(t, err.Error(), "port not set")

	conn = NewConnection(host, database, user, pass, port, "")
	err = conn.Connect()
	assert.NotNil(t, err, "ssl mode is empty")
	assert.Contains(t, err.Error(), "ssl mode not set")

	// success case
	conn = NewConnection(host, database, user, pass, port, ssl)

	err = conn.Connect()
	assert.Nil(t, err, "failed to connect to db instance")

	err = conn.Close()
	assert.Nil(t, err, "failed to close db connection")
}

func TestBasicQueries(t *testing.T) {
	conn := &MockConnection{}

	err := conn.Connect()
	assert.Nil(t, err, "failed to connect to mock db instance")

	mock := conn.GetMock()
	defer conn.Close()

	// SELECT * FROM people WHERE shift_tmtz_end < '10:35:23+05:30' AND fire_team IN ('bravo', 'charlie');
	queryString1 := `SELECT * FROM people WHERE shift_tmtz_end < $1 AND fire_team IN ($2, $3);`
	queryArgs1 := []driver.Value{`10:35:23+05:30`, `bravo`, `charlie`}
	query1, err := NewQuery(conn)
	assert.Nil(t, err)

	mock.ExpectQuery(queryString1).WithArgs(queryArgs1...).WillReturnRows(sqlmock.NewRows([]string{"_id"}).AddRow([]driver.Value{"1"}...))

	rows, err := query1.Select().For("people").Where([]Condition{
		{Field: "shift_tmtz_end", Operator: "<", Values: []any{`10:35:23+05:30`}, NextLogicalOp: "AND"},
		{Field: "fire_team", Type: ConditionIn, Values: []any{[]any{`bravo`, `charlie`}}},
	}).Find()

	assert.Nil(t, err)
	assert.NotNil(t, rows)

	// // SELECT * FROM people WHERE NOT shift_tmtz_end < '10:35:23+05:30' AND fire_team NOT IN ('bravo', 'charlie');
	queryString2 := `SELECT * FROM people WHERE NOT shift_tmtz_end < $1 AND fire_team NOT IN ($2, $3);`
	queryArgs2 := []driver.Value{`10:35:23+05:30`, `bravo`, `charlie`}
	query2, err := NewQuery(conn)
	assert.Nil(t, err)

	mock.ExpectQuery(queryString2).WithArgs(queryArgs2...).WillReturnRows(sqlmock.NewRows([]string{"_id"}).AddRow([]driver.Value{"1"}...))

	rows, err = query2.Select().For("people").Where([]Condition{
		{Not: true, Field: "shift_tmtz_end", Operator: "<", Values: []any{`10:35:23+05:30`}, NextLogicalOp: "AND"},
		{Not: true, Field: "fire_team", Type: ConditionIn, Values: []any{[]any{`bravo`, `charlie`}}},
	}).Find()

	assert.Nil(t, err)
	assert.NotNil(t, rows)

	// SELECT * FROM people WHERE NOT shift_tmtz_end < '10:35:23+05:30' AND fire_team NOT IN ('bravo', 'charlie') OR _id NOT BETWEEN '1' AND '500';
	queryString3 := `SELECT * FROM people WHERE NOT shift_tmtz_end < $1 AND fire_team NOT IN ($2, $3) OR _id NOT BETWEEN $4 AND $5;`
	queryArgs3 := []driver.Value{`10:35:23+05:30`, `bravo`, `charlie`, `1`, `500`}
	query3, err := NewQuery(conn)
	assert.Nil(t, err)

	mock.ExpectQuery(queryString3).WithArgs(queryArgs3...).WillReturnRows(sqlmock.NewRows([]string{"_id"}).AddRow([]driver.Value{"1"}...))

	rows, err = query3.Select().For("people").Where([]Condition{
		{Not: true, Field: "shift_tmtz_end", Operator: "<", Values: []any{`10:35:23+05:30`}, NextLogicalOp: "AND"},
		{Not: true, Field: "fire_team", Type: ConditionIn, Values: []any{[]any{"bravo", "charlie"}}, NextLogicalOp: "OR"},
		{Not: true, Field: "_id", Type: ConditionBetween, Values: []any{"1", "500"}},
	}).Find()

	assert.Nil(t, err)
	assert.NotNil(t, rows)

	// // SELECT * FROM people WHERE email LIKE 'csymespe@%';
	queryString4 := `SELECT * FROM people WHERE email LIKE $1;`
	queryArgs4 := []driver.Value{`csymespe@%`}
	query4, err := NewQuery(conn)
	assert.Nil(t, err)

	mock.ExpectQuery(queryString4).WithArgs(queryArgs4...).WillReturnRows(sqlmock.NewRows([]string{"_id"}).AddRow([]driver.Value{"1"}...))

	rows, err = query4.Select().For("people").Where([]Condition{
		{Field: "email", Operator: "LIKE", Values: []any{"csymespe@%"}},
	}).Find()

	assert.Nil(t, err)
	assert.NotNil(t, rows)

	// SELECT * FROM people WHERE email NOT LIKE 'csymespe@%';
	queryString5 := `SELECT * FROM people WHERE email NOT LIKE $1;`
	queryArgs5 := []driver.Value{`csymespe@%`}
	query5, err := NewQuery(conn)
	assert.Nil(t, err)

	mock.ExpectQuery(queryString5).WithArgs(queryArgs5...).WillReturnRows(sqlmock.NewRows([]string{"_id"}).AddRow([]driver.Value{"1"}...))

	rows, err = query5.Select().For("people").Where([]Condition{
		{Field: "email", Operator: "NOT LIKE", Values: []any{"csymespe@%"}},
	}).Find()

	assert.Nil(t, err)
	assert.NotNil(t, rows)

	// // Case insensitive search queries
	// // SELECT * FROM people WHERE email ILIKE 'csymespe@%';
	queryString6 := `SELECT * FROM people WHERE email ILIKE $1;`
	queryArgs6 := []driver.Value{`csymespe@%`}
	query6, err := NewQuery(conn)
	assert.Nil(t, err)

	mock.ExpectQuery(queryString6).WithArgs(queryArgs6...).WillReturnRows(sqlmock.NewRows([]string{"_id"}).AddRow([]driver.Value{"1"}...))

	rows, err = query6.Select().For("people").Where([]Condition{
		{Field: "email", Operator: "ILIKE", Values: []any{"csymespe@%"}},
	}).Find()

	assert.Nil(t, err)
	assert.NotNil(t, rows)

	// // SELECT _id, title FROM people WHERE shift_tmtz_start > '18:55:23+05:30' AND shift_tmtz_end < '10:35:23+05:30'
	// //  ORDER BY fire_team ASC
	// //  LIMIT 10 OFFSET 5;
	queryString7 := `SELECT _id, title FROM people WHERE  shift_tmtz_start > $1 AND  shift_tmtz_end < $2 ORDER BY fire_team ASC LIMIT 10 OFFSET 5;`
	queryArgs7 := []driver.Value{`18:55:23+05:30`, `10:35:23+05:30`}
	query7, err := NewQuery(conn)
	assert.Nil(t, err)

	mock.ExpectQuery(queryString7).WithArgs(queryArgs7...).WillReturnRows(sqlmock.NewRows([]string{"_id"}).AddRow([]driver.Value{"1"}...))

	rows, err = query7.Select("_id", "title").For("people").Where([]Condition{
		{Field: "shift_tmtz_start", Operator: ">", Values: []any{`18:55:23+05:30`}, NextLogicalOp: "AND"},
		{Field: "shift_tmtz_end", Operator: "<", Values: []any{"10:35:23+05:30"}},
	}).OrderBy([]OrderClause{
		{Field: "fire_team", Order: "ASC"},
	}).Limit(10).Offset(5).Find()

	assert.Nil(t, err)
	assert.NotNil(t, rows)

	// select _id,title, role, location, fire_team, email, slack_handle, shift_tmtz_start, shift_tmtz_end, shift_type from people
	//  where shift_tmtz_start >= '18:55:23+05:30'
	//  and shift_tmtz_end < '05:33:48+05:30'
	//  and shift_type = 'nocturnal';
	queryString8 := `SELECT _id, title, role, location, fire_team, email, slack_handle, shift_tmtz_start, shift_tmtz_end, shift_type FROM people WHERE  shift_tmtz_start >= $1 AND  shift_tmtz_end < $2 AND  shift_type = $3;`
	queryArgs8 := []driver.Value{`18:55:23+05:30`, `05:33:48+05:30`, `nocturnal`}
	query8, err := NewQuery(conn)
	assert.Nil(t, err)

	mock.ExpectQuery(queryString8).WithArgs(queryArgs8...).WillReturnRows(sqlmock.NewRows([]string{"_id"}).AddRow([]driver.Value{"1"}...))

	rows, err = query8.Select("_id", "title", "role", "location", "fire_team", "email", "slack_handle", "shift_tmtz_start", "shift_tmtz_end",
		"shift_type").For("people").Where([]Condition{
		{Field: "shift_tmtz_start", Operator: ">=", Values: []any{"18:55:23+05:30"}, NextLogicalOp: "AND"},
		{Field: "shift_tmtz_end", Operator: "<", Values: []any{"05:33:48+05:30"}, NextLogicalOp: "AND"},
		{Field: "shift_type", Operator: "=", Values: []any{"nocturnal"}},
	}).Find()

	assert.Nil(t, err)
	assert.NotNil(t, rows)

	// SELECT _id, email FROM people WHERE shift_tmtz_start > '18:55:23+05:30' AND shift_tmtz_end < '10:35:23+05:30'
	//  OR ( location = 'FR' AND fire_team = 'fortrot' )
	//  OR ( email = 'mdykasm1@naver.com' AND slack_handle = 'mdykasm1' )
	//  AND ( fire_team IN ('echo', 'foxtrot') AND _id BETWEEN '1' AND '500' )
	//  ORDER BY fire_team DESC LIMIT 10;
	queryString9 := `SELECT _id, email FROM people WHERE  shift_tmtz_start > $1 AND  shift_tmtz_end < $2 OR
	 (  location = $3 AND  fire_team = $4 ) OR (  email = $5 AND  slack_handle = $6 )
	    AND ( fire_team  IN ($7, $8) AND _id  BETWEEN $9 AND $10 ) ORDER BY fire_team DESC LIMIT 10;`
	queryArgs9 := []driver.Value{`18:55:23+05:30`, `10:35:23+05:30`, `FR`, `fortrot`, `mdykasm1@naver.com`, `mdykasm1`, `echo`, `foxtrot`, `1`, `500`}
	query9, err := NewQuery(conn)
	assert.Nil(t, err)

	mock.ExpectQuery(queryString9).WithArgs(queryArgs9...).WillReturnRows(sqlmock.NewRows([]string{"_id"}).AddRow([]driver.Value{"1"}...))

	rows, err = query9.Select("_id", "email").For("people").Where([]Condition{
		{Field: "shift_tmtz_start", Operator: ">", Values: []any{`18:55:23+05:30`}, NextLogicalOp: "AND"},
		{Field: "shift_tmtz_end", Operator: "<", Values: []any{"10:35:23+05:30"}, NextLogicalOp: "OR"},
		{
			NextLogicalOp: "OR",
			Nested: &WhereGroup{
				Conditions: []Condition{
					{Field: "location", Operator: "=", Values: []any{"FR"}, NextLogicalOp: "AND"},
					{Field: "fire_team", Operator: "=", Values: []any{"fortrot"}},
				},
			},
		},
		{
			NextLogicalOp: "AND",
			Nested: &WhereGroup{
				Conditions: []Condition{
					{Field: "email", Operator: "=", Values: []any{"mdykasm1@naver.com"}, NextLogicalOp: "AND"},
					{Field: "slack_handle", Operator: "=", Values: []any{"mdykasm1"}},
				},
			},
		},
		{
			Nested: &WhereGroup{
				Conditions: []Condition{
					{Field: "fire_team", Type: ConditionIn, Values: []any{[]any{"echo", "foxtrot"}}, NextLogicalOp: "AND"},
					{Field: "_id", Type: ConditionBetween, Values: []any{"1", "500"}},
				},
			},
		},
	}).OrderBy([]OrderClause{
		{Field: "fire_team", Order: "DESC"},
	}).Limit(10).Offset(0).Find()

	assert.Nil(t, err)
	assert.NotNil(t, rows)
}

// should be able to select specific columns
func TestAdvanceQueries(t *testing.T) {
	conn := &MockConnection{}

	err := conn.Connect()
	assert.Nil(t, err, "failed to connect to db instance")

	mock := conn.GetMock()

	defer conn.Close()

	// Inner Join with Full join
	// SELECT * FROM people INNER JOIN people_skills ON people._id = people_skills.skills_id FULL JOIN skills ON skills._id = people._id WHERE shift_tmtz_end < $1; [10:35:23+05:30]
	queryString1 := `SELECT * FROM people INNER JOIN people_skills ON people._id = $1 FULL JOIN skills ON skills._id = $2 WHERE shift_tmtz_end < $3;`
	queryArgs1 := []driver.Value{`people_skills.skills_id`, `people._id`, `10:35:23+05:30`}
	query1, err := NewQuery(conn)
	assert.Nil(t, err)

	columns := []string{"_id", "title", "role", "location", "fire_team", "email", "slack_handle", "shift_tmtz_start", "shift_tmtz_end", "shift_type"}
	mockRow1 := []driver.Value{`15`, `Jamie Pomfrett`, `fireteammanager`, `CN`, `charlie`, `jpomfrette@mail.ru`, `jpomfrette`, `0000-01-01 23:18:50 +0530 +0530`, `0000-01-01 07:12:08 +0530 +0530`, `diurnal`}
	mock.ExpectQuery(queryString1).WithArgs(queryArgs1...).WillReturnRows(sqlmock.NewRows(columns).AddRow(mockRow1...))

	rows, err := query1.Select().For("people").Join([]JoinClause{
		{JoinType: InnerJoin, Table: "people_skills", Condition: Condition{Field: "people._id", Operator: "=", Values: []any{"people_skills.skills_id"}}},
		{JoinType: FullJoin, Table: "skills", Condition: Condition{Field: "skills._id", Operator: "=", Values: []any{"people._id"}}},
	}).Where([]Condition{
		{Field: "shift_tmtz_end", Operator: "<", Values: []any{"10:35:23+05:30"}},
	}).Find()

	assert.Nil(t, err)
	assert.NotNil(t, rows)

	// Groupd by Clause
	// SELECT _id, email FROM people WHERE shift_tmtz_start > '18:55:23+05:30' AND shift_tmtz_end < '10:35:23+05:30'
	//  OR ( location = 'FR' AND fire_team = 'fortrot' )
	//  OR ( email = 'mdykasm1@naver.com' AND slack_handle = 'mdykasm1' )
	//  GROUP BY _id, location, fire_team
	//  ORDER BY shift_tmtz_start ASC
	//  LIMIT 10;
	queryString2 := `SELECT _id, email FROM people WHERE  shift_tmtz_start > $1 AND  shift_tmtz_end < $2 OR (  location = $3 AND  fire_team = $4 ) OR (  email = $5 AND  slack_handle = $6 ) GROUP BY _id, location, fire_team ORDER BY shift_tmtz_start ASC LIMIT 10;`
	queryArgs2 := []driver.Value{`18:55:23+05:30`, `10:35:23+05:30`, `FR`, `fortrot`, `mdykasm1@naver.com`, `mdykasm1`}
	query2, err := NewQuery(conn)
	assert.Nil(t, err)

	mock.ExpectQuery(queryString2).WithArgs(queryArgs2...).WillReturnRows(sqlmock.NewRows(columns).AddRow(mockRow1...))

	rows, err = query2.Select("_id", "email").For("people").Where([]Condition{
		{Field: "shift_tmtz_start", Operator: ">", Values: []any{"18:55:23+05:30"}, NextLogicalOp: "AND"},
		{Field: "shift_tmtz_end", Operator: "<", Values: []any{"10:35:23+05:30"}, NextLogicalOp: "OR"},
		{
			NextLogicalOp: "OR",
			Nested: &WhereGroup{
				[]Condition{
					{Field: "location", Operator: "=", Values: []any{"FR"}, NextLogicalOp: "AND"},
					{Field: "fire_team", Operator: "=", Values: []any{"fortrot"}},
				},
			},
		},
		{
			Nested: &WhereGroup{
				[]Condition{
					{Field: "email", Operator: "=", Values: []any{"mdykasm1@naver.com"}, NextLogicalOp: "AND"},
					{Field: "slack_handle", Operator: "=", Values: []any{"mdykasm1"}},
				},
			},
		},
	}).GroupBy([]string{"_id", "location", "fire_team"}).OrderBy([]OrderClause{
		{Field: "shift_tmtz_start", Order: "ASC"},
	}).Limit(10).Offset(0).Find()

	assert.Nil(t, err)
	assert.NotNil(t, rows)

	// Group By with Having Clauses
	// SELECT * FROM people WHERE shift_tmtz_start > '18:55:23+05:30' AND shift_tmtz_end < '10:35:23+05:30'
	//  OR ( location = 'FR' AND fire_team = 'fortrot' ) OR
	//  ( email = 'mdykasm1@naver.com' AND slack_handle = 'mdykasm1' )
	//  GROUP BY _id, location, fire_team
	//  HAVING location = 'FR' AND shift_type = 'nocturnal'
	//  ORDER BY shift_tmtz_start ASC
	//  LIMIT 10;
	queryString3 := `SELECT * FROM people WHERE  shift_tmtz_start > $1 AND  shift_tmtz_end < $2 OR (  location = $3 AND  fire_team = $4 ) OR (  email = $5 AND  slack_handle = $6 ) GROUP BY _id, location, fire_team HAVING location = $7 AND shift_type = $8 ORDER BY shift_tmtz_start ASC LIMIT 10;`
	queryArgs3 := []driver.Value{`18:55:23+05:30`, `10:35:23+05:30`, `FR`, `fortrot`, `mdykasm1@naver.com`, `mdykasm1`, `FR`, `nocturnal`}
	query3, err := NewQuery(conn)
	assert.Nil(t, err)

	mock.ExpectQuery(queryString3).WithArgs(queryArgs3...).WillReturnRows(sqlmock.NewRows(columns).AddRow(mockRow1...))

	rows, err = query3.Select().For("people").Where([]Condition{
		{Field: "shift_tmtz_start", Operator: ">", Values: []any{"18:55:23+05:30"}, NextLogicalOp: "AND"},
		{Field: "shift_tmtz_end", Operator: "<", Values: []any{"10:35:23+05:30"}, NextLogicalOp: "OR"},
		{
			NextLogicalOp: "OR",
			Nested: &WhereGroup{
				[]Condition{
					{Field: "location", Operator: "=", Values: []any{"FR"}, NextLogicalOp: "AND"},
					{Field: "fire_team", Operator: "=", Values: []any{"fortrot"}},
				},
			},
		},
		{
			Nested: &WhereGroup{
				[]Condition{
					{Field: "email", Operator: "=", Values: []any{"mdykasm1@naver.com"}, NextLogicalOp: "AND"},
					{Field: "slack_handle", Operator: "=", Values: []any{"mdykasm1"}},
				},
			},
		},
	}).GroupBy([]string{"_id", "location", "fire_team"}).Having([]Condition{
		{Field: "location", Operator: "=", Values: []any{"FR"}},
		{Field: "shift_type", Operator: "=", Values: []any{"nocturnal"}},
	}).OrderBy([]OrderClause{
		{Field: "shift_tmtz_start", Order: "ASC"},
	}).Limit(10).Offset(0).Find()

	assert.Nil(t, err)
	assert.NotNil(t, rows)
}
