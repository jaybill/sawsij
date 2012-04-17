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
	"fmt"
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
	tableName = MakeDbName(tableName)
    fieldNames  := []string{}
    fieldValues := make([]interface{},0)
    marks       := []string{}
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		log.Printf("%d: %s %s = %v\n", i,typeOfT.Field(i).Name, f.Type(), f.Interface())
		if typeOfT.Field(i).Name != "Id" {
		    fieldNames = append(fieldNames,MakeDbName(typeOfT.Field(i).Name))
		    fieldValues = append(fieldValues,f.Interface())
		    marks = append(marks,"?")
		}
	}
    
	query := fmt.Sprintf("INSERT INTO %v(%v) VALUES (%v)",tableName,strings.Join(fieldNames,","),strings.Join(marks,","))

	log.Printf("Query: %q", query)
    
	result, err := m.Db.Exec(query,fieldValues)
	if err != nil {
		log.Print(err)
	} else {
		log.Printf("Result was %v", result)
	}
    
}
