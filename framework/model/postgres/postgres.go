package postgres

import (
	"bitbucket.org/jaybill/sawsij/framework/model"
	"fmt"
	"strings"
)

type Queries struct {
}

func GetQueries() (q model.Queries) {
	q = &Queries{}
	return
}

func (q *Queries) TableCount() string {
	return "SELECT COUNT(*) AS TOTAL FROM PG_TABLES WHERE SCHEMANAME='%v';"
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

func (q *Queries) DeleteWhere() string {
	return "DELETE FROM %v WHERE %v"
}

func (q *Queries) LastInsertId(seqId string) string {
	return fmt.Sprintf("SELECT CURRVAL(%v)", seqId)
}

func (q *Queries) TableName(schema string, tablename string) string {
	return fmt.Sprintf("%q.%q", schema, tablename)
}

func (q *Queries) SequenceName(schema string, tablename string) string {
	return fmt.Sprintf("'%v.%v_id_seq'", schema, tablename)
}

func (q *Queries) DbVersion() string {
	return "SELECT version_id from %v.sawsij_db_version ORDER BY ran_on DESC LIMIT 1;"
}

func (q *Queries) DbEmpty(schema string, database string) string {
	return fmt.Sprintf("SELECT count(*) as tables FROM information_schema.tables WHERE table_schema = '%v';", schema)
}

func (q *Queries) DescribeTable(table string, schema string, database string) string {
	query := fmt.Sprintf("select column_name,data_type,is_nullable from information_schema.columns where table_name = '%v' and table_schema = '%v' order by ordinal_position;", table, schema)
	return query
}

func (q *Queries) ConnString(user string, password string, host string, dbname string, port string) string {

	return fmt.Sprintf("user=%v password=%v host=%v dbname=%v port=%v sslmode=disable", user, password, host, dbname, port)

}

func (q *Queries) P(ord int) string {
	return fmt.Sprintf("$%v", ord)
}

// Reads in a db connect string, like "user=hodor password=foobar dbname=hodor sslmode=disable" and returns a map.

func (q *Queries) ParseConnect(connectStr string) (connect map[string]string) {

	parts := strings.Split(connectStr, " ")
	connect = make(map[string]string, len(parts))

	for _, part := range parts {
		keyval := strings.TrimSpace(part)
		keyvalparts := strings.Split(keyval, "=")
		key := keyvalparts[0]

		connect[key] = keyvalparts[1]
	}

	return

}
