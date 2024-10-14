package database

import (
	"fmt"
	"strconv"
	"strings"
)

type QueryExecutor interface {
	Select(cols ...string) QueryExecutor
	For(table string) QueryExecutor
	Where(conditions []Condition) QueryExecutor

	Join([]JoinClause) QueryExecutor

	OrderBy(orderBy []OrderClause) QueryExecutor
	GroupBy(groupBy []string) QueryExecutor
	Having(having []Condition) QueryExecutor

	Limit(limit int) QueryExecutor
	Offset(offset int) QueryExecutor
	Set(values map[string]any) QueryExecutor

	Find() ([][]any, error) // select query to be executed
	FindOne(id string) ([]any, error)

	Create() (int64, error)
	Update() (int64, error)

	Delete() (int64, error)
	DeleteOne(d string) (int64, error)
}

type WhereGroup struct {
	Conditions []Condition
}

type ConditionType int

const (
	ConditionStandard ConditionType = iota
	ConditionIn
	ConditionBetween
)

// Struct for a WHERE condition
type Condition struct {
	Field         string
	Operator      string
	Values        []interface{}
	NextLogicalOp string
	Nested        *WhereGroup
	Not           bool
	Type          ConditionType
}

type JoinType string

const (
	BasicJoin JoinType = "JOIN"
	InnerJoin JoinType = "INNER JOIN"
	OuterJoin JoinType = "OUT JOIN"
	LeftJoin  JoinType = "LEFT JOIN"
	RightJoin JoinType = "RIGHT JOIN"
	FullJoin  JoinType = "FULL JOIN"
)

// Struct for a JOIN clause
type JoinClause struct {
	JoinType  JoinType
	Table     string
	Condition Condition
}

// Struct for an ORDER BY clause
type OrderClause struct {
	Field string
	Order string
}

type query struct {
	// handle to the db instance
	conn ConnectionExecutor

	table     string
	cols      []string
	colValues []interface{}

	whereConds  []Condition
	joins       []JoinClause
	groupBy     []string
	having      []Condition
	orderBy     []OrderClause
	limit       int
	offset      int
	Args        []any
	ArgCount    int
	QueryString string

	errors []error
}

func (q *query) Select(cols ...string) QueryExecutor {
	colLen := len(cols)
	if colLen != 0 {
		colsArr := make([]string, len(cols))
		copy(colsArr, cols)
		q.cols = cols
	} else {
		q.cols = []string{"*"}
	}
	return q
}

func (q *query) For(table string) QueryExecutor {
	if table == "" {
		q.errors = append(q.errors, fmt.Errorf("empty table not allowed"))
	}

	q.table = table
	return q
}

func (q *query) Where(conditions []Condition) QueryExecutor {
	if len(conditions) == 0 {
		q.errors = append(q.errors, fmt.Errorf("empty conditions not allowed"))
	}

	q.whereConds = conditions
	return q
}

func (q *query) Join(joins []JoinClause) QueryExecutor {
	if len(joins) == 0 {
		q.errors = append(q.errors, fmt.Errorf("empty joins not allowed"))
	}

	q.joins = joins
	return q
}

func (q *query) GroupBy(groupBy []string) QueryExecutor {
	if len(groupBy) == 0 {
		q.errors = append(q.errors, fmt.Errorf("empty group clauses not allowed"))
	}

	q.groupBy = groupBy
	return q
}

func (q *query) OrderBy(orderBy []OrderClause) QueryExecutor {
	if len(orderBy) == 0 {
		q.errors = append(q.errors, fmt.Errorf("empty orderby clauses not allowed"))
	}

	q.orderBy = orderBy
	return q
}

func (q *query) Having(having []Condition) QueryExecutor {
	if len(having) == 0 {
		q.errors = append(q.errors, fmt.Errorf("empty having clauses not allowed"))
	}

	q.having = having
	return q
}

func (q *query) Limit(limit int) QueryExecutor {
	q.limit = limit
	return q
}

func (q *query) Offset(offset int) QueryExecutor {
	q.offset = offset
	return q
}

func (q *query) addJoins() error {
	if len(q.whereConds) > 0 {
		for _, join := range q.joins {
			if join.JoinType == "" || join.Table == "" {
				nerr := fmt.Errorf("join missing type/table")
				q.errors = append(q.errors, nerr)
				return nerr
			}
			condition := join.Condition
			q.ArgCount++

			if condition.Field == "" || condition.Operator == "" || condition.Values[0] == nil {
				nerr := fmt.Errorf("join condition invalid")
				q.errors = append(q.errors, nerr)
				return nerr
			}

			q.QueryString += fmt.Sprintf(" %s %s ON %s %s $%s", join.JoinType, join.Table, condition.Field, condition.Operator, strconv.Itoa(q.ArgCount))
			q.Args = append(q.Args, condition.Values[0])
		}
	}
	return nil
}

func (q *query) addGroupBy() error {
	if len(q.groupBy) > 0 {
		q.QueryString += " GROUP BY " + strings.Join(q.groupBy, ", ")
	}
	return nil
}

func (q *query) addHaving() error {
	if len(q.having) > 0 {
		var havingClauses []string
		for _, cond := range q.having {
			q.ArgCount++

			if cond.Field == "" || cond.Operator == "" || cond.Values[0] == nil {
				nerr := fmt.Errorf("having condition invalid")
				q.errors = append(q.errors, nerr)
				return nerr
			}

			havingClauses = append(havingClauses, fmt.Sprintf("%s %s $%s", cond.Field, cond.Operator, strconv.Itoa(q.ArgCount)))
			q.Args = append(q.Args, cond.Values[0])
		}
		q.QueryString += " HAVING "
		q.QueryString += strings.Join(havingClauses, " AND ")
	}
	return nil
}

func (q *query) addOrderby() error {
	if len(q.orderBy) > 0 {
		var orderClauses []string
		for _, order := range q.orderBy {

			if order.Field == "" || order.Order == "" {
				nerr := fmt.Errorf("order condition invalid")
				q.errors = append(q.errors, nerr)
				return nerr
			}

			orderClauses = append(orderClauses, fmt.Sprintf("%s %s", order.Field, order.Order))
		}
		q.QueryString += " ORDER BY " + strings.Join(orderClauses, ", ")
	}
	return nil
}

func (q *query) checkPreBuildErrors() error {
	if len(q.errors) > 0 {
		errStrs := make([]string, len(q.errors))
		for i, err := range q.errors {
			errStrs[i] = err.Error()
		}

		errs := strings.Join(errStrs, ", ")
		return fmt.Errorf("query pre build errors [ %s ]", errs)
	}
	return nil
}

func (q *query) BuildSelectQuery() error {
	err := q.checkPreBuildErrors()
	if err != nil {
		return err
	}

	// Start building the query
	q.QueryString = fmt.Sprintf("SELECT %s FROM %s", strings.Join(q.cols, ", "), q.table)

	// Add JOIN clauses
	err = q.addJoins()
	if err != nil {
		return err
	}

	// Add WHERE conditions
	err = q.addWhere()
	if err != nil {
		return err
	}

	// Add GROUP BY
	err = q.addGroupBy()
	if err != nil {
		return err
	}

	// Add HAVING conditions
	err = q.addHaving()
	if err != nil {
		return err
	}

	// Add ORDER BY
	err = q.addOrderby()
	if err != nil {
		return err
	}

	// Add LIMIT and OFFSET
	if q.limit > 0 {
		q.QueryString += fmt.Sprintf(" LIMIT %d", q.limit)
	}
	if q.offset > 0 {
		q.QueryString += fmt.Sprintf(" OFFSET %d", q.offset)
	}

	q.QueryString += ";"

	return nil
}

func (q *query) FindOne(id string) ([]any, error) {
	q.whereConds = []Condition{
		{Field: "_id", Operator: "=", Values: []any{id}},
	}

	rows, err := q.Find()
	if err != nil {
		return nil, err
	}

	lr := len(rows)
	if lr < 1 {
		return nil, fmt.Errorf("could not find record")
	}

	return rows[0], nil
}

// implements the select functionality
func (q *query) Find() ([][]any, error) {
	err := q.BuildSelectQuery()
	if err != nil {
		return nil, err
	}

	fmt.Println("queryString ", q.QueryString, q.Args)
	rows, err := q.conn.GetDB().Query(q.QueryString, q.Args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rowData [][]any

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	for rows.Next() {

		vals := make([]any, len(cols))
		scanArgs := make([]any, len(cols))

		for i := range vals {
			scanArgs[i] = &vals[i]
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}

		rowData = append(rowData, vals)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rowData, err
}

// constructor for query which adds handle for the db connection
func (q *query) init(conn ConnectionExecutor) error {
	if conn == nil {
		return fmt.Errorf("could not find connection")
	}

	if conn.GetDB() == nil {
		return fmt.Errorf("could not find db instance")
	}

	return nil
}

// contructor for query
func NewQuery(conn ConnectionExecutor) (QueryExecutor, error) {
	q := &query{}

	err := q.init(conn)
	if err != nil {
		return nil, err
	}

	q.conn = conn

	q.errors = make([]error, 0)
	return q, nil
}
