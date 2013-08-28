package mysql

import (
	"bitbucket.org/jaybill/sawsij/framework/model"
)

type Queries struct {
}

func GetQueries() (q model.Queries) {
	return
}

func (q *Queries) Fetch() string {
	return "foo"
}

func (q *Queries) Update() string {
	return "foo"
}
