package database

import (
	"fmt"
	"strconv"
	"strings"
)

// build the in clause
func (q *query) buildInClause(cond Condition) (string, int) {
	values := cond.Values[0].([]interface{})
	placeholders := make([]string, len(values))
	for i := range values {
		q.ArgCount++
		placeholders[i] = fmt.Sprintf("$%s", strconv.Itoa(q.ArgCount))
	}

	// check for not condition
	notStr := ""
	if cond.Not {
		notStr = "NOT"
	}

	return fmt.Sprintf("%s %s IN (%s)", cond.Field, notStr, strings.Join(placeholders, ", ")), q.ArgCount
}

// Recursive function to handle nested conditions
func (q *query) buildWhereClauses(conds []Condition, whereClauses []string) ([]string, int) {

	for i, cond := range conds {
		if cond.Nested != nil {
			// Open parentheses for nested conditions
			whereClauses = append(whereClauses, "(")

			wc, ac := q.buildWhereClauses(cond.Nested.Conditions, whereClauses)
			whereClauses = wc
			q.ArgCount = ac

			whereClauses = append(whereClauses, ")")
		} else {

			// check for the not condition
			notStr := ""
			if cond.Not {
				notStr = "NOT"
			}

			// Add the actual condition
			switch cond.Type {
			case ConditionIn: // handle in clause
				inClause, iargc := q.buildInClause(cond)
				q.ArgCount = iargc
				whereClauses = append(whereClauses, inClause)
				q.Args = append(q.Args, cond.Values[0].([]interface{})...)
			case ConditionBetween: // handle between clause
				q.ArgCount++
				arg1 := q.ArgCount
				q.ArgCount++
				arg2 := q.ArgCount
				betweenClause := fmt.Sprintf("%s %s BETWEEN $%s AND $%s", cond.Field, notStr, strconv.Itoa(arg1), strconv.Itoa(arg2))
				whereClauses = append(whereClauses, betweenClause)
				q.Args = append(q.Args, cond.Values[0], cond.Values[1])
			default:
				q.ArgCount++
				clause := fmt.Sprintf("%s %s %s $%s", notStr, cond.Field, cond.Operator, strconv.Itoa(q.ArgCount))
				whereClauses = append(whereClauses, clause)
				q.Args = append(q.Args, cond.Values[0])
			}
		}

		// Add logical operator between conditions, except for the last one
		if i < len(conds)-1 && cond.NextLogicalOp != "" {
			whereClauses = append(whereClauses, cond.NextLogicalOp)
		}
	}
	return whereClauses, q.ArgCount
}

func (q *query) addWhere() error {
	if len(q.whereConds) > 0 {
		q.QueryString += " WHERE "
		var whereClauses []string
		whereClauses, q.ArgCount = q.buildWhereClauses(q.whereConds, whereClauses)
		q.QueryString += strings.Join(whereClauses, " ")
	}
	return nil
}
