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

func (m *Model) Update(data interface{}) (err error) {
    
    rowInfo  := getRowInfo(data)
    holders     := make([]string,len(rowInfo.Keys))
    
    for i := 0; i < len(rowInfo.Keys); i++{
        holders[i] = fmt.Sprintf("%v=$%v",rowInfo.Keys[i],i + 1)
    }
    
    query := fmt.Sprintf("UPDATE %v SET %v WHERE id=%d",rowInfo.TableName,strings.Join(holders,","),rowInfo.Id)
    log.Printf("Query: %q", query)
    
    _, err = m.Db.Exec(query,rowInfo.Vals...)
	if err != nil {
		log.Print(err)
	}
    	
    return
}

func (m *Model) Insert(data interface{}) (err error) {

    rowInfo     := getRowInfo(data)
    holders     := make([]string,len(rowInfo.Keys))
    
    for i := 0; i < len(rowInfo.Keys); i++{
        holders[i] = fmt.Sprintf("$%v",i + 1)
    }
    
	query := fmt.Sprintf("INSERT INTO %v(%v) VALUES (%v)",rowInfo.TableName,strings.Join(rowInfo.Keys,","),strings.Join(holders,","))

	log.Printf("Query: %q", query)	
	_, err = m.Db.Exec(query,rowInfo.Vals...)
	if err != nil {
		log.Print(err)
	} else {
	    if rowInfo.IdIndex != -1 {
	        idq := fmt.Sprintf("select currval('%v_id_seq')",rowInfo.TableName)	    
	        row := m.Db.QueryRow(idq)
	        if err != nil {
		        log.Print(err)
		    } else {
		        var id int64
                err = row.Scan(&id)
                if err != nil{
                    log.Print(err)
                } else {    
                    s := reflect.ValueOf(data).Elem()
                    s.Field(rowInfo.IdIndex).SetInt(id)     
        		    log.Printf("Id was %v", id)
        		}
		    }
		}
	}
	return    
}

func (m *Model) Delete(data interface{}) (err error){
    rowInfo     := getRowInfo(data)
    if rowInfo.Id != -1{
        query := fmt.Sprintf("DELETE FROM %v WHERE id=%d",rowInfo.TableName,rowInfo.Id)
        log.Printf("Query: %q", query)
        
        _, err = m.Db.Exec(query)
	    if err != nil {
		    log.Print(err)
	    }
    }
    return
}

type forDb struct{
    Id          interface{}
    IdIndex     int
    Keys        []string
    Vals        []interface{}
    TableName   string    
}

func getRowInfo(data interface{}) (rowInfo forDb){
	s := reflect.ValueOf(data).Elem()
	typeOfT := s.Type()
    rowInfo.Vals = make([]interface{},0)
    
    rowInfo.IdIndex = -1
    
    rowInfo.TableName = typeOfT.String()
	parts := strings.Split(rowInfo.TableName, ".")
	if len(parts) > 0 {
		rowInfo.TableName = parts[len(parts)-1]
	}
	rowInfo.TableName = MakeDbName(rowInfo.TableName)
    
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		log.Printf("%d: %s %s = %v\n", i,typeOfT.Field(i).Name, f.Type(), f.Interface())
		if typeOfT.Field(i).Name != "Id" {
		    rowInfo.Keys = append(rowInfo.Keys,MakeDbName(typeOfT.Field(i).Name))
		    rowInfo.Vals = append(rowInfo.Vals,f.Interface())		    
		} else {
		    rowInfo.Id = f.Interface()
		    rowInfo.IdIndex = i
		}
	}
    return
}


