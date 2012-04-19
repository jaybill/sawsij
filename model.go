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

func (m *Model) Fetch(data interface{}) (err error){
    rowInfo     := getRowInfo(data)
    cols        := make([]interface{}, 0)    
    retRow      := reflect.ValueOf(data).Elem()
    dataType    := retRow.Type()
    if rowInfo.Id != -1{
        query := fmt.Sprintf("SELECT %v FROM %v WHERE id=%d",strings.Join(rowInfo.Keys,","),rowInfo.TableName,rowInfo.Id)
        log.Printf("Query: %q", query)
        row := m.Db.QueryRow(query)
        
        for i := 0; i < dataType.NumField(); i++ {            
            f := retRow.Field(i)
		    if dataType.Field(i).Name != "Id" {
		        cols = append(cols, f.Addr().Interface())            
		    }
        }   
        err = row.Scan(cols...) 
        if err != nil{
            log.Print(err)
        }    
    }
    return
}

func (m *Model) FetchAll(data interface{}, where string, args ...interface{}) (ents []interface{},err error){
    ents        = make([]interface{},0)
    rowInfo     := getRowInfo(data)
        
    retRow      := reflect.ValueOf(data).Elem()
    dataType    := retRow.Type()
    t           := reflect.TypeOf(data).Elem()
    
    //log.Println(empty)
    
    query := fmt.Sprintf("SELECT %v FROM %v",strings.Join(rowInfo.Keys,","),rowInfo.TableName)
    if where != ""{
        query = fmt.Sprintf("%v WHERE %v",query,where)
    }
    log.Printf("Query: %q", query)
    
    rows,err := m.Db.Query(query,args...)
    for rows.Next() {
        ent     := reflect.New(t)
        cols    := make([]interface{}, 0)
        for i := 0; i < dataType.NumField(); i++ {            
            f := ent.Elem().Field(i)
		    if dataType.Field(i).Name != "Id" {
		        cols = append(cols, f.Addr().Interface())            
		    }
        }   
        err = rows.Scan(cols...) 
        
        ents = append(ents, ent.Interface())
        if err != nil{
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

func getTableName(data interface{}) (tableName string){
    s := reflect.ValueOf(data).Elem()
	typeOfT := s.Type()

    tableName = typeOfT.String()
	parts := strings.Split(tableName, ".")
	if len(parts) > 0 {
		tableName = parts[len(parts)-1]
	}
	tableName = MakeDbName(tableName)

    return
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
	rowInfo.TableName = getTableName(data)
    
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		//log.Printf("%d: %s %s = %v\n", i,typeOfT.Field(i).Name, f.Type(), f.Interface())
		if typeOfT.Field(i).Name != "Id" {
		    dbName := MakeDbName(typeOfT.Field(i).Name)
		    rowInfo.Keys = append(rowInfo.Keys,dbName)
		    rowInfo.Vals = append(rowInfo.Vals,f.Interface())		    
		} else {
		    rowInfo.Id = f.Interface()
		    rowInfo.IdIndex = i
		}
	}
    return
}


