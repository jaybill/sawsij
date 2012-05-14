// Copyright 2012 J. William McCarthy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sawsij

import (
	"flag"
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
	"testing"
	"time"

	"database/sql"
	_ "github.com/bmizerany/pq"
)

var config *yaml.File
var configFile = flag.String("c", "./config_test.yaml", "path to config file")

type Post struct {
	Id        int64
	Title     string
	Body      string
	CreatedOn time.Time
}

func configure(t *testing.T) (db *sql.DB) {

	if config == nil {
		flag.Parse()
	}

	c, err := yaml.ReadFile(*configFile)
	if err != nil {
		t.Fatal(err)
	}

	driver, err := c.Get("database.driver")
	if err != nil {
		t.Fatal(err)
	}
	connect, err := c.Get("database.connect")
	if err != nil {
		t.Fatal(err)
	}

	db, err = sql.Open(driver, connect)
	if err != nil {
		t.Fatal(err)
	}

	return
}

func TestInsert(t *testing.T) {
	db := configure(t)
	defer db.Close()
	/* stand up test table */

	_, err := db.Exec("DROP TABLE IF EXISTS post")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE post( id bigserial NOT NULL,title text, body text,  created_on timestamp without time zone,  CONSTRAINT pk_posts PRIMARY KEY(id))")
	if err != nil {
		t.Fatal(err)
	}

	model := &Model{Db: db}
	post := &Post{}

	post.Title = "Test Post"
	post.Body = "Here is a test post which is \"awesome\". Can't believe how awesome it is. Geez."
	post.CreatedOn = time.Now()
	model.Insert(post)

	if err != nil {
		t.Fatal(err)
	}
	/* tear down */

	_, err = db.Exec("DROP TABLE IF EXISTS post")
	if err != nil {
		t.Fatal(err)
	}

}

func TestUpdate(t *testing.T) {
	db := configure(t)
	defer db.Close()
	/* stand up test table */

	_, err := db.Exec("DROP TABLE IF EXISTS post")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE post( id bigserial NOT NULL,title text, body text,  created_on timestamp without time zone,  CONSTRAINT pk_posts PRIMARY KEY(id))")
	if err != nil {
		t.Fatal(err)
	}

	model := &Model{Db: db}
	post := &Post{}

	post.Title = "Test Post"
	post.Body = "Here is a test post which is \"awesome\". Can't believe how awesome it is. Geez."
	post.CreatedOn = time.Now()
	model.Insert(post)
	if err != nil {
		t.Fatal(err)
	}

	post.Title = "Something Else"
	model.Update(post)
	if err != nil {
		t.Fatal(err)
	}

	secondPost := &Post{Id: post.Id}
	model.Fetch(secondPost)
	if secondPost.Title != post.Title {
		t.Fatal("Updated value is wrong")
	}

	/* tear down */

	_, err = db.Exec("DROP TABLE IF EXISTS post")
	if err != nil {
		t.Fatal(err)
	}

}

func TestDelete(t *testing.T) {

	db := configure(t)
	defer db.Close()

	/* stand up test table */

	_, err := db.Exec("DROP TABLE IF EXISTS post")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE post( id bigserial NOT NULL,title text, body text,  created_on timestamp without time zone,  CONSTRAINT pk_posts PRIMARY KEY(id))")
	if err != nil {
		t.Fatal(err)
	}

	model := &Model{Db: db}
	post := &Post{}

	post.Title = "Test Post"
	post.Body = "Here is a test post which is \"awesome\". Can't believe how awesome it is. Geez."
	post.CreatedOn = time.Now()
	model.Insert(post)
	if err != nil {
		t.Fatal(err)
	}

	id := post.Id

	err = model.Delete(post)
	if err != nil {
		t.Fatal(err)
	}
	secondPost := &Post{Id: id}
	err = model.Fetch(secondPost)
	if err.Error() != "sql: no rows in result set" {
		t.Fatal("Row still exists after attempted delete")
	}

}

func TestFetchAll(t *testing.T) {

	db := configure(t)
	defer db.Close()

	/* stand up test table */

	_, err := db.Exec("DROP TABLE IF EXISTS post")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE post( id bigserial NOT NULL,title text, body text,  created_on timestamp without time zone,  CONSTRAINT pk_posts PRIMARY KEY(id))")
	if err != nil {
		t.Fatal(err)
	}
	model := &Model{Db: db}
	firstPost := &Post{}
	firstPost.Title = "Test Post"
	firstPost.Body = "Here is a test post which is \"awesome\". Can't believe how awesome it is. Geez."
	firstPost.CreatedOn = time.Now()
	model.Insert(firstPost)
	if err != nil {
		t.Fatal(err)
	}
	secondPost := &Post{}
	secondPost.Title = "Second Post"
	secondPost.Body = "Here is a test post which is \"awesome\". Can't believe how awesome it is. Geez."
	secondPost.CreatedOn = time.Now()
	model.Insert(secondPost)
	if err != nil {
		t.Fatal(err)
	}
	thirdPost := &Post{}
	thirdPost.Title = "Third Post"
	thirdPost.Body = "Here is a test post which is \"awesome\". Can't believe how awesome it is. Geez."
	thirdPost.CreatedOn = time.Now()
	model.Insert(thirdPost)
	if err != nil {
		t.Fatal(err)
	}
	q := Query{}
	posts, err := model.FetchAll(&Post{}, q)
	if err != nil {
		t.Fatal(err)
	}

	if len(posts) != 3 {
		t.Fatal("Select did not get correct number of rows.")
	}

	_, err = db.Exec("DROP TABLE IF EXISTS post")
	if err != nil {
		t.Fatal(err)
	}

}

func TestLimit(t *testing.T) {

	db := configure(t)
	defer db.Close()

	/* stand up test table */

	_, err := db.Exec("DROP TABLE IF EXISTS post")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE post( id bigserial NOT NULL,title text, body text,  created_on timestamp without time zone,  CONSTRAINT pk_posts PRIMARY KEY(id))")
	if err != nil {
		t.Fatal(err)
	}
	model := &Model{Db: db}
	firstPost := &Post{}
	firstPost.Title = "Test Post"
	firstPost.Body = "Here is a test post which is \"awesome\". Can't believe how awesome it is. Geez."
	firstPost.CreatedOn = time.Now()
	model.Insert(firstPost)
	if err != nil {
		t.Fatal(err)
	}
	secondPost := &Post{}
	secondPost.Title = "Second Post"
	secondPost.Body = "Here is a test post which is \"awesome\". Can't believe how awesome it is. Geez."
	secondPost.CreatedOn = time.Now()
	model.Insert(secondPost)
	if err != nil {
		t.Fatal(err)
	}
	thirdPost := &Post{}
	thirdPost.Title = "Third Post"
	thirdPost.Body = "Here is a test post which is \"awesome\". Can't believe how awesome it is. Geez."
	thirdPost.CreatedOn = time.Now()
	model.Insert(thirdPost)
	if err != nil {
		t.Fatal(err)
	}
	q := Query{Limit: 2}
	posts, err := model.FetchAll(&Post{}, q)
	if err != nil {
		t.Fatal(err)
	}

	if len(posts) != 2 {
		t.Fatal("Select did not get correct number of rows.")
	}

	_, err = db.Exec("DROP TABLE IF EXISTS post")
	if err != nil {
		t.Fatal(err)
	}

}

func TestWhere(t *testing.T) {

	db := configure(t)
	defer db.Close()

	/* stand up test table */

	_, err := db.Exec("DROP TABLE IF EXISTS post")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE post( id bigserial NOT NULL,title text, body text,  created_on timestamp without time zone,  CONSTRAINT pk_posts PRIMARY KEY(id))")
	if err != nil {
		t.Fatal(err)
	}
	model := &Model{Db: db}
	firstPost := &Post{}
	firstPost.Title = "Test Post"
	firstPost.Body = "Here is a test post which is \"awesome\". Can't believe how awesome it is. Geez."
	firstPost.CreatedOn = time.Now()
	model.Insert(firstPost)
	if err != nil {
		t.Fatal(err)
	}
	secondPost := &Post{}
	secondPost.Title = "Second Post"
	secondPost.Body = "Here is a test post which is \"awesome\". Can't believe how awesome it is. Geez."
	secondPost.CreatedOn = time.Now()
	model.Insert(secondPost)
	if err != nil {
		t.Fatal(err)
	}
	thirdPost := &Post{}
	thirdPost.Title = "Third Post"
	thirdPost.Body = "Here is a test post which is \"awesome\". Can't believe how awesome it is. Geez."
	thirdPost.CreatedOn = time.Now()
	model.Insert(thirdPost)
	if err != nil {
		t.Fatal(err)
	}
	q := Query{Where: fmt.Sprintf("%v = 'Third Post'", MakeDbName("Title"))}
	posts, err := model.FetchAll(&Post{}, q)
	if err != nil {
		t.Fatal(err)
	}

	if posts[0].(*Post).Title != "Third Post" {
		t.Fatal("Select returned incorrect data based on where clause.")
	}

	_, err = db.Exec("DROP TABLE IF EXISTS post")
	if err != nil {
		t.Fatal(err)
	}
}

func TestOrder(t *testing.T) {

	db := configure(t)
	defer db.Close()

	/* stand up test table */

	_, err := db.Exec("DROP TABLE IF EXISTS post")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE post( id bigserial NOT NULL,title text, body text,  created_on timestamp without time zone,  CONSTRAINT pk_posts PRIMARY KEY(id))")
	if err != nil {
		t.Fatal(err)
	}
	model := &Model{Db: db}
	firstPost := &Post{}
	firstPost.Title = "Test Post"
	firstPost.Body = "Here is a test post which is \"awesome\". Can't believe how awesome it is. Geez."
	firstPost.CreatedOn = time.Now().Add(-(60000000000)*1)
	model.Insert(firstPost)
	if err != nil {
		t.Fatal(err)
	}
	secondPost := &Post{}
	secondPost.Title = "Second Post"
	secondPost.Body = "Here is a test post which is \"awesome\". Can't believe how awesome it is. Geez."
	secondPost.CreatedOn = time.Now().Add(-(60000000000)*2)
	model.Insert(secondPost)
	if err != nil {
		t.Fatal(err)
	}
	thirdPost := &Post{}
	thirdPost.Title = "Third Post"
	thirdPost.Body = "Here is a test post which is \"awesome\". Can't believe how awesome it is. Geez."
	thirdPost.CreatedOn = time.Now().Add(-(60000000000)*3)
	model.Insert(thirdPost)
	if err != nil {
		t.Fatal(err)
	}
	q := Query{Order: fmt.Sprintf("%v DESC", MakeDbName("CreatedOn"))}
	posts, err := model.FetchAll(&Post{}, q)
	if err != nil {
		t.Fatal(err)
	}

	if posts[0].(*Post).Title != "Test Post" {
		t.Fatal("Select returned incorrect data based on where clause.")
	}

	_, err = db.Exec("DROP TABLE IF EXISTS post")
	if err != nil {
		t.Fatal(err)
	}
}

func TestOffset(t *testing.T) {

	db := configure(t)
	defer db.Close()

	/* stand up test table */

	_, err := db.Exec("DROP TABLE IF EXISTS post")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE post( id bigserial NOT NULL,title text, body text,  created_on timestamp without time zone,  CONSTRAINT pk_posts PRIMARY KEY(id))")
	if err != nil {
		t.Fatal(err)
	}
	model := &Model{Db: db}
	firstPost := &Post{}
	firstPost.Title = "Test Post"
	firstPost.Body = "Here is a test post which is \"awesome\". Can't believe how awesome it is. Geez."
	firstPost.CreatedOn = time.Now().Add(-(60000000000)*1)
	model.Insert(firstPost)
	if err != nil {
		t.Fatal(err)
	}
	secondPost := &Post{}
	secondPost.Title = "Second Post"
	secondPost.Body = "Here is a test post which is \"awesome\". Can't believe how awesome it is. Geez."
	secondPost.CreatedOn = time.Now().Add(-(60000000000)*2)
	model.Insert(secondPost)
	if err != nil {
		t.Fatal(err)
	}
	thirdPost := &Post{}
	thirdPost.Title = "Third Post"
	thirdPost.Body = "Here is a test post which is \"awesome\". Can't believe how awesome it is. Geez."
	thirdPost.CreatedOn = time.Now().Add(-(60000000000)*3)
	model.Insert(thirdPost)
	if err != nil {
		t.Fatal(err)
	}
	
	fourthPost := &Post{}
	fourthPost.Title = "Fourth Post"
	fourthPost.Body = "Here is a test post which is \"awesome\". Can't believe how awesome it is. Geez."
	fourthPost.CreatedOn = time.Now().Add(-(60000000000)*4)
	model.Insert(fourthPost)
	if err != nil {
		t.Fatal(err)
	}
		
	q := Query{	    
	    Order: fmt.Sprintf("%v DESC", MakeDbName("CreatedOn")),
	    Offset: 2}
	posts, err := model.FetchAll(&Post{}, q)
	if err != nil {
		t.Fatal(err)
	}

	if posts[0].(*Post).Title != "Third Post" {
		t.Fatal("Select did not get correct number of rows.")
	}

	_, err = db.Exec("DROP TABLE IF EXISTS post")
	if err != nil {
		t.Fatal(err)
	}

}


