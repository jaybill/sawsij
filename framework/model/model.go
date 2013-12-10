// Copyright 2012 J. William McCarthy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/* Provides a simple database access layer and table/field mapping.

The Model package is intended to provide something analagous to a lightweight ORM, though not quite.
The general pattern of usage is that you create a struct that represents a row in your table, with the
fields mapping to column names. You then pass a pointer to that struct into Model's various methods
to perform database operations. At the moment, only postgres is supported.

Struct field names are converted to database column names using sawsij.MakeDbName() and database column
names are converted to struct field names using sawsij.MakeFieldName(). Generally, it works as follows:

A struct field called "FirstName" will be mapped to a database column named "first_name".
A struct field called "Type" will be mapped to a database column named "type".

Table names are mapped the same way, wherin a struct of type "PersonAddress" would look for a table called "person_address".

If you set the Schema property to a valid schema name, it will be used. If generally only use one schema in your app, you can just
set the search path for the database user in postgres and omit the Schema property. (You can specifiy it if you want to use some other
schema.) You can generally do this in postgres with the following query:

ALTER USER [db_username] SET search_path to '[app_schema_name]'

As currently implemented, both your table and your struct must have an identity to do anything useful.
("join" tables being the exception)
*/
package model

import (
	"database/sql"
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
	"io/ioutil"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// The Queries interface describes the functions needed for a specific database platform. All these functions will return strings
// with template database queries to be used by the model package.
type Queries interface {
	TableCount() string
	Update() string
	Fetch() string
	FetchAllSelect() string
	FetchAllWhere() string
	FetchAllOrder() string
	FetchAllLimit() string
	FetchAllOffset() string
	Insert() string
	LastInsertId(string) string
	Delete() string
	DeleteWhere() string
	TableName(string, string) string
	SequenceName(string, string) string
	DbVersion() string
	DbEmpty(string, string) string
	DescribeTable(string, string, string) string
	ConnString(string, string, string, string, string) string
	P(int) string
	ParseConnect(string) map[string]string
}

// Table is the primary means of interaction with the database. It represents the access to a table, not the table itself.
// The package figures out what table to use based on the type being passed to the various methods of Table.
// Using anything but a 'flat' struct as a type will have unpredictable results.
type Table struct {
	Db     *DbSetup
	Schema string
}

// A DbSetup is used to store a reference to the database connection and schema information.
type DbSetup struct {
	Db            *sql.DB
	DefaultSchema string
	Schemas       []Schema
	GetQueries    func() Queries
}

// A Schema is used to store schema information, like the schema name and what version it is.
type Schema struct {
	Name    string
	Version int64
}

// DbVersion is a type representing the db_version table, which must exist in any schema you plan to use with
// "[appserver] [directory] migrate"
type SawsijDbVersion struct {
	VersionId int64
	RanOn     time.Time
}

// Update expects a pointer to a struct that represents a row in your database. The "Id" field of the struct will be used in the where clause.
func (m *Table) Update(data interface{}) (err error) {

	rowInfo := m.getRowInfo(data, false)
	holders := make([]string, len(rowInfo.Keys))

	for i := 0; i < len(rowInfo.Keys); i++ {
		holders[i] = fmt.Sprintf("%v=%v", rowInfo.Keys[i], m.Db.GetQueries().P(i+1))
	}

	query := fmt.Sprintf(m.Db.GetQueries().Update(), rowInfo.TableName, strings.Join(holders, ","), rowInfo.Id)
	log.Printf("Query: %q", query)

	_, err = m.Db.Db.Exec(query, rowInfo.Vals...)
	if err != nil {
		log.Print(err)
	}

	return
}

// Insert expects a pointer to a struct that represents a row in your database. The "Id" field of the referenced struct will be populated with the
// identity value if the row is successfully inserted.
func (m *Table) Insert(data interface{}) (err error) {
	rowInfo := m.getRowInfo(data, false)

	holders := make([]string, len(rowInfo.Keys))

	for i := 0; i < len(rowInfo.Keys); i++ {
		holders[i] = fmt.Sprintf("%v", m.Db.GetQueries().P(i+1))
	}

	query := fmt.Sprintf(m.Db.GetQueries().Insert(), rowInfo.TableName, strings.Join(rowInfo.Keys, ","), strings.Join(holders, ","))

	log.Printf("Query: %q", query)
	log.Printf("Data: %+v", data)

	_, err = m.Db.Db.Exec(query, rowInfo.Vals...)
	if err != nil {
		log.Print(err)
	} else {
		if rowInfo.IdIndex != -1 {
			log.Printf("Table name is [%v]", rowInfo.TableName)
			idq := m.Db.GetQueries().LastInsertId(rowInfo.SequenceName)
			log.Printf("Sequence query: %v", idq)
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

// Insert expects a pointer to a struct that represents a row in your database. Assumes your Id field will be set in the struct.
func (m *Table) InsertWithoutAutoId(data interface{}) (err error) {
	rowInfo := m.getRowInfo(data, true)

	holders := make([]string, len(rowInfo.Keys))

	for i := 0; i < len(rowInfo.Keys); i++ {
		holders[i] = fmt.Sprintf("%v", m.Db.GetQueries().P(i+1))
	}

	query := fmt.Sprintf(m.Db.GetQueries().Insert(), rowInfo.TableName, strings.Join(rowInfo.Keys, ","), strings.Join(holders, ","))

	_, err = m.Db.Db.Exec(query, rowInfo.Vals...)
	if err != nil {
		log.Print(err)
	}

	return
}

// Delete takes a pointer to a struct and deletes the row where the id in the table is the Id of the struct.
// Note that you don't need to have acquired this struct from a row, passing in a pointer to something like {Id: 4} will totally work.
func (m *Table) Delete(data interface{}) (err error) {
	rowInfo := m.getRowInfo(data, false)
	if rowInfo.Id != -1 {
		query := fmt.Sprintf(m.Db.GetQueries().Delete(), rowInfo.TableName, rowInfo.Id)
		log.Printf("Query: %q", query)

		_, err = m.Db.Db.Exec(query)
		if err != nil {
			log.Print(err)
		}
	}
	return
}

// Delete takes a pointer to a struct and deletes the row where the id in the table is the Id of the struct.
// Note that you don't need to have acquired this struct from a row, passing in a pointer to something like {Id: 4} will totally work.
func (m *Table) DeleteWhere(data interface{}, whereClause string) (err error) {
	rowInfo := m.getRowInfo(data, false)

	query := fmt.Sprintf(m.Db.GetQueries().DeleteWhere(), rowInfo.TableName, whereClause)
	log.Printf("Query: %q", query)

	_, err = m.Db.Db.Exec(query)
	if err != nil {
		log.Print(err)
	}

	return
}

// Fetch returns a single row where the id in the table is the Id of the struct.
func (m *Table) Fetch(data interface{}) (err error) {

	rowInfo := m.getRowInfo(data, false)
	cols := make([]interface{}, 0)
	retRow := reflect.ValueOf(data).Elem()
	dataType := retRow.Type()
	if rowInfo.Id != -1 {
		query := fmt.Sprintf(m.Db.GetQueries().Fetch(), strings.Join(rowInfo.Keys, ","), rowInfo.TableName, rowInfo.Id)
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
func (m *Table) FetchAll(data interface{}, q Query, args ...interface{}) (ents []interface{}, err error) {
	ents = make([]interface{}, 0)
	rowInfo := m.getRowInfo(data, true)

	retRow := reflect.ValueOf(data).Elem()
	dataType := retRow.Type()
	t := reflect.TypeOf(data).Elem()

	query := fmt.Sprintf(m.Db.GetQueries().FetchAllSelect(), strings.Join(rowInfo.Keys, ","), rowInfo.TableName)
	if q.Where != "" {
		query = fmt.Sprintf(m.Db.GetQueries().FetchAllWhere(), query, q.Where)
	}

	if q.Order != "" {
		query = fmt.Sprintf(m.Db.GetQueries().FetchAllOrder(), query, q.Order)
	}

	if q.Limit != 0 {
		query = fmt.Sprintf(m.Db.GetQueries().FetchAllLimit(), query, q.Limit)
	}

	if q.Offset != 0 {
		query = fmt.Sprintf(m.Db.GetQueries().FetchAllOffset(), query, q.Offset)
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
			if err != nil {
				log.Print(err)
			}
			ents = append(ents, ent.Interface())

		}
	} else {
		log.Print(err)
	}
	return
}

type forDb struct {
	Id           interface{}
	IdIndex      int
	Keys         []string
	Vals         []interface{}
	TableName    string
	SequenceName string
}

func (m *Table) getTableNames(data interface{}) (tableName string, sequenceName string) {
	s := reflect.ValueOf(data).Elem()
	typeOfT := s.Type()

	tableName = typeOfT.String()
	parts := strings.Split(tableName, ".")
	if len(parts) > 0 {
		tableName = parts[len(parts)-1]
	}
	dbTableName := MakeDbName(tableName)
	if m.Schema != "" {
		tableName = m.Db.GetQueries().TableName(m.Schema, dbTableName)
		sequenceName = m.Db.GetQueries().SequenceName(m.Schema, dbTableName)
	} else {
		tableName = m.Db.GetQueries().TableName(m.Db.DefaultSchema, dbTableName)
		sequenceName = m.Db.GetQueries().SequenceName(m.Db.DefaultSchema, dbTableName)
	}

	return
}

func (m *Table) getRowInfo(data interface{}, includeId bool) (rowInfo forDb) {
	s := reflect.ValueOf(data).Elem()
	typeOfT := s.Type()

	rowInfo.Vals = make([]interface{}, 0)

	rowInfo.IdIndex = -1

	rowInfo.TableName = typeOfT.String()

	parts := strings.Split(rowInfo.TableName, ".")
	if len(parts) > 0 {
		rowInfo.TableName = parts[len(parts)-1]
	}

	rowInfo.TableName, rowInfo.SequenceName = m.getTableNames(data)

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

// RunScript takes a file path to a sql script, reads the file and runs it against the database/schema.
// Queries will be run one at a time in a transaction. If there's any kind of error, the whole transaction
// will be rolled back.
// Queries must be terminated by ";" and you must use C style comments, not --. (This limitation will
// probably be removed at some point.)
func RunScript(db *sql.DB, dbscript string) (err error) {

	bQuery, err := ioutil.ReadFile(dbscript)
	if err != nil {
		return
	} else {

		t, err := db.Begin()

		sQuery := string(bQuery)
		queries := strings.Split(sQuery, ";")
		for _, query := range queries {
			query = strings.TrimSpace(query)
			var isComment bool = false

			if strings.HasPrefix(query, "/*") && strings.HasSuffix(query, "*/") {
				isComment = true
			}

			if query != "" && !isComment {

				log.Printf("Query: %q", query)
				_, err = t.Exec(query)
				if err != nil {
					t.Rollback()
					return err
				}

			}
		}
		err = t.Commit()
		if err != nil {
			return err
		}
	}
	return

}

// InsertBatch expects an array of pointers to a structs that represent a rows in your database.
// The "Id" fields of the referenced struct will be populated with the identity values if the rows
// are successfully inserted. The inserts are done in a transaction and rolled back on the first error.
func (m *Table) InsertBatch(items []interface{}) (err error) {
	t, err := m.Db.Db.Begin()
	for _, data := range items {

		rowInfo := m.getRowInfo(data, false)
		holders := make([]string, len(rowInfo.Keys))

		for i := 0; i < len(rowInfo.Keys); i++ {
			holders[i] = fmt.Sprintf("$%v", i+1)
		}

		query := fmt.Sprintf(m.Db.GetQueries().Insert(), rowInfo.TableName, strings.Join(rowInfo.Keys, ","), strings.Join(holders, ","))

		log.Printf("Query: %q", query)
		log.Printf("Data: %+v", data)
		_, err = t.Exec(query, rowInfo.Vals...)
		if err != nil {
			log.Print(err)
			t.Rollback()
			return err
		} else {
			if rowInfo.IdIndex != -1 {
				log.Printf("Table name is [%v]", rowInfo.TableName)
				idq := m.Db.GetQueries().LastInsertId(rowInfo.SequenceName)
				log.Printf("Sequence query: %v", idq)
				row := t.QueryRow(idq)
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
	}
	err = t.Commit()
	if err != nil {
		return err
	}
	return
}

// MakeDbName converts a struct field name into a database field name.
// The string will be converted to lowercase, and any capital letters after the first one will be prepended with an underscore.
// "FirstName" will become "first_name" and so on.
func MakeDbName(fieldName string) string {
	runes := []rune(fieldName)
	copy := []rune{}
	usrunes := []rune("_")
	us := usrunes[0]
	for i := 0; i < len(runes); i++ {
		if i > 0 && unicode.IsUpper(runes[i]) {
			copy = append(copy, us)
		}
		runes[i] = unicode.ToLower(runes[i])
		copy = append(copy, runes[i])

	}
	return string(copy)
}

// MakeDbName converts a database column name into a struct field name.
// The first letter will be made capital, and underscores will be removed and the following letter made capital.
// "first_name" will become "FirstName", etc.
func MakeFieldName(dbName string) string {

	runes := []rune(dbName)
	copy := []rune{}
	usrunes := []rune("_")
	us := usrunes[0]
	for i := 0; i < len(runes); i++ {
		if runes[i] != us {
			if i == 0 {
				runes[i] = unicode.ToUpper(runes[i])
			}
			copy = append(copy, runes[i])
		} else {
			runes[i+1] = unicode.ToUpper(runes[i+1])
		}
	}
	return string(copy)
}

// Reads dbversions file specified by filename and returns schema information
func ParseDbVersionsFile(dBconfigFilename string) (defaultSchema string, allSchemas []Schema, err error) {

	dbvc, err := yaml.ReadFile(dBconfigFilename)
	if err != nil {
		err = &SawsijDbError{fmt.Sprintf("Can't read %v", dBconfigFilename)}
		return
	}

	defaultSchema, err = dbvc.Get("default_schema")
	if err != nil {
		err = &SawsijDbError{fmt.Sprintf("default_schema not defined in %v", dBconfigFilename)}
		return
	}

	schemasN, err := yaml.Child(dbvc.Root, ".schema_versions")
	if err != nil {
		err = &SawsijDbError{fmt.Sprintf("Error reading schema_versions in %v", dBconfigFilename)}
		return
	}

	if schemasN != nil {
		schemas := schemasN.(yaml.Map)
		for schema, version := range schemas {
			sV, _ := strconv.ParseInt(fmt.Sprintf("%v", version), 0, 0)
			allSchemas = append(allSchemas, Schema{Name: string(schema), Version: sV})
		}
	} else {
		err = &SawsijDbError{fmt.Sprintf("No schemas defined in %v", dBconfigFilename)}
		return
	}

	return
}

// A struct for returning model error messages
type SawsijDbError struct {
	What string
}

// Returns the error message defined in What as a string
func (e *SawsijDbError) Error() string {
	return e.What
}
