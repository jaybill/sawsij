package mysql

import (
	"bitbucket.org/jaybill/sawsij/framework/model"
	"fmt"
	"regexp"
	"strings"
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

func (q *Queries) LastInsertId(seqId string) string {
	return "SELECT LAST_INSERT_ID();"
}

func (q *Queries) TableName(schema string, tablename string) string {
	return fmt.Sprintf("%v_%v", schema, tablename)
}

func (q *Queries) SequenceName(schema string, tablename string) string {
	return ""
}

func (q *Queries) DbVersion() string {
	return "SELECT version_id from %v_sawsij_db_version ORDER BY ran_on DESC LIMIT 1;"
}

func (q *Queries) DbEmpty(schema string, database string) string {
	return fmt.Sprintf("SELECT count(*) as tables FROM information_schema.tables WHERE table_schema = '%v';", database)
}

func (q *Queries) DescribeTable(table string, schema string, database string) string {
	query := fmt.Sprintf("select column_name,data_type,is_nullable from information_schema.columns where table_name = \"%v_%v\" and table_schema = %q order by ordinal_position;", schema, table, database)
	return query
}

func (q *Queries) ConnString(user string, password string, host string, dbname string, port string) string {

	return fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?autocommit=true&parseTime=true", user, password, host, port, dbname)

}

func (q *Queries) P(ord int) string {
	return "?"
}

func (q *Queries) ParseConnect(connectStr string) (connect map[string]string) {

	connect = make(map[string]string, 1)
	re := regexp.MustCompile("\\/(.+)\\?")
	connect["dbname"] = re.FindString(connectStr)
	connect["dbname"] = strings.TrimPrefix(connect["dbname"], "/")
	connect["dbname"] = strings.TrimSuffix(connect["dbname"], "?")

	return
}
