package qibo

import (
	"regexp"
	"strings"
)

// Query struct binder for default query param
type Query struct {
	Sort   string `json:"sort"`
	Filter map[string]interface{}
}

// Operator string translation
var Operator = map[string]string{
	"gt":   ">",
	"lt":   "<",
	"eq":   "=",
	"ne":   "!=",
	"gte":  ">=",
	"lte":  "<=",
	"like": "LIKE",
	"in":   "IN",
	"or":   "OR",
}

// SetFilter to replace filters
func (q *Query) SetFilter(filter map[string]interface{}) {
	q.Filter = filter
}

// GetFilter to get list of existing filters
func (q *Query) GetFilter() map[string]interface{} {
	return q.Filter
}

// Where generate sql WHERE statement ,with format
//		key :"{columnName}{$operator}"
//		value : interface
// with default operator value "$eq"
// for example :
//     "amount$gte": 19200.00
// 	   "status": 1
// will be translated into sql format :
// 		WHERE amount >= 19200.00
//		AND status = 1
func (q *Query) Where() (string, []interface{}) {
	var wheres []string
	var args []interface{}

	for k, v := range q.Filter {
		var validDate = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)
		fields := strings.Split(k, "$")
		columnName := fields[0]
		isRequire := func(s string) bool {
			return s[len(s)-1:] == "!"
		}(fields[1])
		opr := translateOperator(strings.TrimSuffix(fields[1], "!"))
		if isRequire || !IsArgNil(v) {
			switch opr {
			case Operator["like"]:
				wheres = append(wheres, columnName+` `+opr+` ?`)
				tmpArgs, _ := v.(string)
				args = append(args, "%"+tmpArgs+"%")
			case Operator["in"]:
				wheres = append(wheres, columnName+` `+opr+` (?)`)
				args = append(args, v)
			case Operator["lte"]:
				wheres = append(wheres, columnName+` `+opr+` ?`)
				tmpArgs, _ := v.(string)
				if validDate.MatchString(tmpArgs) {
					tmpArgs += " 23:59:59"
				}
				args = append(args, tmpArgs)
			case Operator["gte"]:
				wheres = append(wheres, columnName+` `+opr+` ?`)
				tmpArgs, _ := v.(string)
				if validDate.MatchString(tmpArgs) {
					tmpArgs += " 00:00:00"
				}
				args = append(args, tmpArgs)
			case Operator["or"]:
				wheres = append(wheres, columnName+` `+opr+` (?)`)
				args = append(args, v)
			default:
				wheres = append(wheres, columnName+` `+opr+` ?`)
				args = append(args, v)
			}
		} else {
			wheres = append(wheres, ` 1 = 1 `)
		}

	}
	return strings.Join(wheres, " AND "), args
}

// Order generate string ordering query statement
func (q *Query) Order() string {
	if len(q.Sort) > 0 {
		field := strings.Split(q.Sort, ",")
		var sort string
		for _, v := range field {
			sortType := func(str string) string {
				if strings.HasPrefix(str, "-") {
					return `DESC`
				}
				return `ASC`
			}
			sort += strings.TrimPrefix(v, "-") + ` ` + sortType(v) + `,`
		}
		return sort[:len(sort)-1]
	}
	return ``
}

func translateOperator(s string) string {
	operator := Operator[strings.ToLower(s)]
	if operator == "" {
		return Operator["eq"]
	}
	return operator
}
