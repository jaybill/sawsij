/*
Make sure the user's default schema is set
ALTER USER jayblog SET search_path to 'jayblog'

*/

package sawsij

import (
	"database/sql"
	_ "github.com/bmizerany/pq"
	"log"
	"reflect"
	"strings"
)

type Model struct {
	Db *sql.DB
}

func (m *Model) Setup(db *sql.DB) {
	m.Db = db
}

func (m *Model) Insert(data interface{}) {

	s := reflect.ValueOf(data).Elem()
	typeOfT := s.Type()

	tableName := typeOfT.String()
	parts := strings.Split(tableName, ".")
	if len(parts) > 0 {
		tableName = parts[len(parts)-1]
	}
	tableName = strings.ToLower(tableName)
    
    
    
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		log.Printf("%d: %s %s = %v\n", i,
			typeOfT.Field(i).Name, f.Type(), f.Interface())
	}

	query := "INSERT INTO " + tableName + "(title, body, createdon) VALUES ('Five Post', 'Even better than the first', current_timestamp)"

	log.Printf("Query: %q", query)

	result, err := m.Db.Query(query)
	if err != nil {
		log.Print(err)
	} else {
		log.Printf("Result was %v", result)
	}

}
