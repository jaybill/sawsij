package sawsij

import (
	"flag"
	"testing"
	"time"

	"github.com/stathat/jconfig"

	"database/sql"
	_ "github.com/bmizerany/pq"
)

var config *jconfig.Config

type Post struct {
	Id        int64
	Title     string
	Body      string
	CreatedOn time.Time
}

func configure(t *testing.T) (db *sql.DB) {
	if config == nil {
		var configFile string

		flag.StringVar(&configFile, "c", "./config_test.json", "path to config file")
		flag.Parse()
		config = jconfig.LoadConfig(configFile)
	}
	db, err := sql.Open("postgres", config.GetString("dbConnect"))
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

	posts, err := model.FetchAll(&Post{}, "")
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

//TODO: Add tests for order by and limit queries

