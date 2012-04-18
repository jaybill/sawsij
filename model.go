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

func (m *Model) Insert(data interface{}) (id int, err error) {

    tableName   := getTableName(data)
    keys, vals  := keysVals(data)
    holders     := make([]string,len(keys))
    
    for i := 0; i < len(keys); i++{
        holders[i] = fmt.Sprintf("$%v",i + 1)
    }
    
	query := fmt.Sprintf("INSERT INTO %v(%v) VALUES (%v)",tableName,strings.Join(keys,","),strings.Join(holders,","))

	log.Printf("Query: %q", query)	
	_, err = m.Db.Exec(query,vals...)
	if err != nil {
		log.Print(err)
	} else {
	    idq := fmt.Sprintf("select currval('%v_id_seq')",tableName)	    
	    row := m.Db.QueryRow(idq)
	    if err != nil {
		    log.Print(err)
		} else {
		    
            err = row.Scan(&id)	
            if err != nil{
                log.Print(err)
            } else {            
    		    log.Printf("Id was %v", id)
    		}
		}
	}
	return    
}

func getTableName(data interface{}) string {
	s := reflect.ValueOf(data).Elem()
	typeOfT := s.Type()

	tableName := typeOfT.String()
	parts := strings.Split(tableName, ".")
	if len(parts) > 0 {
		tableName = parts[len(parts)-1]
	}
	return MakeDbName(tableName)
}

func keysVals(data interface{}) (keys []string, vals []interface{}){
	s := reflect.ValueOf(data).Elem()
	typeOfT := s.Type()


    vals = make([]interface{},0)
    
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		log.Printf("%d: %s %s = %v\n", i,typeOfT.Field(i).Name, f.Type(), f.Interface())
		if typeOfT.Field(i).Name != "Id" {
		    keys = append(keys,MakeDbName(typeOfT.Field(i).Name))
		    vals = append(vals,f.Interface())
		    
		}
	}
    return
}


