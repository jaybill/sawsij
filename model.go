// Copyright 2012 J. William McCarthy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sawsij

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"
)

// The Model struct is intended to provide something analagous to a lightweight ORM, though not quite. 
// The general pattern of usage is that you create a struct that represents a row in your table, with the 
// fields mapping to column names. You then pass a pointer to that struct into Model's various methods 
// to perform database operations. At the moment, only postgres is supported.
//
// Struct field names are converted to database column names using sawsij.MakeDbName() and database column
// names are converted to struct field names using sawsij.MakeFieldName(). Generally, it works as follows:
//
// A struct field called "FirstName" will be mapped to a database column named "first_name".
// A struct field called "Type" will be mapped to a database column named "type".
//
// Table names are mapped the same way, wherin a struct of type "PersonAddress" would look for a table called "person_address".
//
// If you set the Schema property to a valid schema name, it will be used. If generally only use one schema in your app, you can just
// set the search path for the database user in postgres and omit the Schema property. (You can specifiy it if you want to use some other
// schema.) You can generally do this in postgres with the following query:
//
// ALTER USER [db_username] SET search_path to '[app_schema_name]'
// 
// As currently implemented, both your table and your struct must have an identity to do anything useful. 
// ("join" tables being the exception)
type Model struct {
	Db     *DbSetup
	Schema string
}

// DbVersion is a type representing the db_version table, which must exist in any schema you plan to use with 
// "sawsijcmd migrate"
type SawsijDbVersion struct {
	Id    int64
	RanOn time.Time
}

// Update expects a pointer to a struct that represents a row in your database. The "Id" field of the struct will be used in the where clause.
func (m *Model) Update(data interface{}) (err error) {

	rowInfo := m.getRowInfo(data, false)
	holders := make([]string, len(rowInfo.Keys))

	for i := 0; i < len(rowInfo.Keys); i++ {
		holders[i] = fmt.Sprintf("%v=$%v", rowInfo.Keys[i], i+1)
	}

	query := fmt.Sprintf("UPDATE %q SET %v WHERE id=%d", rowInfo.TableName, strings.Join(holders, ","), rowInfo.Id)
	log.Printf("Query: %q", query)

	_, err = m.Db.Db.Exec(query, rowInfo.Vals...)
	if err != nil {
		log.Print(err)
	}

	return
}

// Insert expects a pointer to a struct that represents a row in your database. The "Id" field of the referenced struct will be populated with the 
// identity value if the row is successfully inserted.
func (m *Model) Insert(data interface{}) (err error) {

	rowInfo := m.getRowInfo(data, false)
	holders := make([]string, len(rowInfo.Keys))

	for i := 0; i < len(rowInfo.Keys); i++ {
		holders[i] = fmt.Sprintf("$%v", i+1)
	}

	query := fmt.Sprintf("INSERT INTO %q(%v) VALUES (%v)", rowInfo.TableName, strings.Join(rowInfo.Keys, ","), strings.Join(holders, ","))

	log.Printf("Query: %q", query)
	log.Printf("Data: %+v", data)
	_, err = m.Db.Db.Exec(query, rowInfo.Vals...)
	if err != nil {
		log.Print(err)
	} else {
		if rowInfo.IdIndex != -1 {
			idq := fmt.Sprintf("select currval('%v_id_seq')", rowInfo.TableName)
			row := m.Db.Db.QueryRow(idq)
			if err != nil {
				log.Print(err)
			} else {
				var id int64
				err = row.Scan(&id)
				if err != nil {
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

// Delete takes a pointer to a struct and deletes the row where the id in the table is the Id of the struct.
// Note that you don't need to have acquired this struct from a row, passing in a pointer to something like {Id: 4} will totally work.
func (m *Model) Delete(data interface{}) (err error) {
	rowInfo := m.getRowInfo(data, false)
	if rowInfo.Id != -1 {
		query := fmt.Sprintf("DELETE FROM %q WHERE id=%d", rowInfo.TableName, rowInfo.Id)
		log.Printf("Query: %q", query)

		_, err = m.Db.Db.Exec(query)
		if err != nil {
			log.Print(err)
		}
	}
	return
}

// Fetch returns a single row where the id in the table is the Id of the struct.
func (m *Model) Fetch(data interface{}) (err error) {
	rowInfo := m.getRowInfo(data, false)
	cols := make([]interface{}, 0)
	retRow := reflect.ValueOf(data).Elem()
	dataType := retRow.Type()
	if rowInfo.Id != -1 {
		query := fmt.Sprintf("SELECT %v FROM %q WHERE id=%d", strings.Join(rowInfo.Keys, ","), rowInfo.TableName, rowInfo.Id)
		log.Printf("Query: %q", query)
		row := m.Db.Db.QueryRow(query)

		for i := 0; i < dataType.NumField(); i++ {
			f := retRow.Field(i)
			if dataType.Field(i).Name != "Id" {
				cols = append(cols, f.Addr().Interface())
			}
		}
		err = row.Scan(cols...)
		if err != nil {
			log.Print(err)
		}
	}
	return
}

// The Query type is used to construct a SQL query. You should always use MakeDbName() to get column names, as this will 
// ensure cross-RDBMS compatibility later on.
type Query struct {
	// A where clause, such as fmt.Sprintf("%v = 'Third Post'", MakeDbName("Title"))
	Where string
	// An order clause, such as fmt.Sprintf("%v DESC", MakeDbName("CreatedOn"))
	Order string
	// A number of records to limit the results to
	Limit int
	// The number of rows to offset the returned results by
	Offset int
}

// FetchAll accepts a reference to a struct (generally "blank", though it doesn't matter), a Query and a set of query arguments and returns a set of rows that match
// the query.
func (m *Model) FetchAll(data interface{}, q Query, args ...interface{}) (ents []interface{}, err error) {
	ents = make([]interface{}, 0)
	rowInfo := m.getRowInfo(data, true)

	retRow := reflect.ValueOf(data).Elem()
	dataType := retRow.Type()
	t := reflect.TypeOf(data).Elem()

	query := fmt.Sprintf("SELECT %v FROM %q", strings.Join(rowInfo.Keys, ","), rowInfo.TableName)
	if q.Where != "" {
		query = fmt.Sprintf("%v WHERE %v", query, q.Where)
	}

	if q.Order != "" {
		query = fmt.Sprintf("%v ORDER BY %v", query, q.Order)
	}

	if q.Limit != 0 {
		query = fmt.Sprintf("%v LIMIT %v", query, q.Limit)
	}

	if q.Offset != 0 {
		query = fmt.Sprintf("%v OFFSET %v", query, q.Offset)
	}

	log.Printf("Query: %q", query)

	rows, err := m.Db.Db.Query(query, args...)
	if err == nil {
		for rows.Next() {
			ent := reflect.New(t)
			cols := make([]interface{}, 0)
			for i := 0; i < dataType.NumField(); i++ {
				f := ent.Elem().Field(i)
				cols = append(cols, f.Addr().Interface())
			}
			err = rows.Scan(cols...)

			ents = append(ents, ent.Interface())
			if err != nil {
				log.Print(err)
			}
		}
	} else {
		log.Print(err)
	}
	return
}

type forDb struct {
	Id        interface{}
	IdIndex   int
	Keys      []string
	Vals      []interface{}
	TableName string
}

func (m *Model) getTableName(data interface{}) (tableName string) {
	s := reflect.ValueOf(data).Elem()
	typeOfT := s.Type()

	tableName = typeOfT.String()
	parts := strings.Split(tableName, ".")
	if len(parts) > 0 {
		tableName = parts[len(parts)-1]
	}
	tableName = MakeDbName(tableName)
	if m.Schema != "" {
		tableName += m.Schema + "." + tableName
	} else {
		tableName += m.Db.DefaultSchema + "." + tableName
	}
	return
}

func (m *Model) getRowInfo(data interface{}, includeId bool) (rowInfo forDb) {
	s := reflect.ValueOf(data).Elem()
	typeOfT := s.Type()
	rowInfo.Vals = make([]interface{}, 0)

	rowInfo.IdIndex = -1

	rowInfo.TableName = typeOfT.String()
	parts := strings.Split(rowInfo.TableName, ".")
	if len(parts) > 0 {
		rowInfo.TableName = parts[len(parts)-1]
	}
	rowInfo.TableName = m.getTableName(data)

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)

		if typeOfT.Field(i).Name != "Id" || includeId {
			dbName := MakeDbName(typeOfT.Field(i).Name)
			rowInfo.Keys = append(rowInfo.Keys, dbName)
			rowInfo.Vals = append(rowInfo.Vals, f.Interface())
		}

		if typeOfT.Field(i).Name == "Id" {
			rowInfo.Id = f.Interface()
			rowInfo.IdIndex = i
		}

	}
	return
}
