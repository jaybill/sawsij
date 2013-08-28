package postgres

import (
	"bitbucket.org/jaybill/sawsij/framework/model"
	"fmt"
)

type Queries struct {
}

func GetQueries() (q model.Queries) {
	q = &Queries{}
	return
}

func (q *Queries) Fetch() string {
	return "SELECT %v FROM %v WHERE id=%d"
}

func (q *Queries) Update() string {
	return "UPDATE %v SET %v WHERE id=%d"
}

func (q *Queries) Insert() string {
	return "INSERT INTO %v (%v) VALUES (%v)"
}

func (q *Queries) FetchAllSelect() string {
	return "SELECT %v FROM %v"
}
func (q *Queries) FetchAllWhere() string {
	return "%v WHERE %v"
}
func (q *Queries) FetchAllOrder() string {
	return "%v ORDER BY %v"
}
func (q *Queries) FetchAllLimit() string {
	return "%v LIMIT %v"
}
func (q *Queries) FetchAllOffset() string {
	return "%v OFFSET %v"
}

func (q *Queries) Delete() string {
	return "DELETE FROM %v WHERE id=%d"
}

func (q *Queries) LastInsertId() string {
	return "SELECT CURRVAL(%v)"
}

func (q *Queries) TableName(schema string, tablename string) string {
	return fmt.Sprintf("%q.%q", schema, tablename)
}

func (q *Queries) SequenceName(schema string, tablename string) string {
	return fmt.Sprintf("'%v.%v_id_seq'", schema, tablename)
}
