package {{.pName}}

import (
	"bitbucket.org/jaybill/sawsij/framework"
	"bitbucket.org/jaybill/sawsij/framework/model"	
	"log"{{if .importStrconv}}
	"strconv"{{ end }}
	"net/http"{{if .importTime}}
	"time"{{ end }}{{if .importFmt}}	
	"fmt"{{ end }}
)

type {{.typeName}} struct { {{ range $field := .struct }}
	{{$field.FName}} {{ if $field.CanBeNull}}*{{ end }}{{$field.FType}}{{ end }} 
}

func (o *{{.typeName}}) GetValidationErrors(a *framework.AppScope) (errors []string) {
	// Add validation here

	return
}

func {{.typeName}}AdminEditHandler(r *http.Request, a *framework.AppScope, rs *framework.RequestScope) (h framework.HandlerResponse, err error) {
	h.Init()
	
	t := &model.Table{Db: a.Db}
	{{.typeVar}} := &{{.typeName}}{}

	{{.typeVar}}.Id = framework.GetIntId(rs.UrlParamMap["id"])
	if {{.typeVar}}.Id != -1 {
		err = t.Fetch({{.typeVar}})
		if err != nil {
			log.Print(err)
			h.Redirect = "/error"
			return
		} else {
			h.View["{{.typeVar}}"] = {{.typeVar}}
		}
	}

	if r.Method == "POST" {
		{{ range $field := .struct }}{{ if $field.IsPk }}{{ else }}{{if eq $field.DisplayType "timestamp"}}// Timestamp field set to current time. This might not be what you want.
		t{{$field.FName}} := time.Now()
		{{$.typeVar}}.{{$field.FName}} = {{ if $field.CanBeNull }}&{{ end}}t{{$field.FName}}{{end}}
		{{if eq $field.FType "string"}}t{{$field.FName}} := r.FormValue("{{$field.FName}}")
		{{$.typeVar}}.{{$field.FName}} = {{ if $field.CanBeNull }}&{{ end}}t{{$field.FName}}{{end}}
		{{if eq $field.FType "int64"}}t{{$field.FName}},_ := strconv.ParseInt(r.FormValue("{{$field.FName}}"),10,0)
		{{$.typeVar}}.{{$field.FName}} = {{ if $field.CanBeNull }}&{{ end}}t{{$field.FName}}
		{{end}}
		{{ if eq $field.DisplayType "date"}}		
		ts{{$field.FName}} := r.FormValue("{{$field.FName}}")
		if ts{{$field.FName}} != ""{
			t{{$field.FName}}, _ := time.Parse("01/02/2006",ts{{$field.FName}})			
			{{$.typeVar}}.{{$field.FName}} = {{ if $field.CanBeNull }}&{{ end}}t{{$field.FName}}							
		}{{ end }}{{ end }}{{ end }}
		errors := {{.typeVar}}.GetValidationErrors(a)
		if len(errors) == 0 {
			if {{.typeVar}}.Id == -1 {
				// This is an insert

				err = t.Insert({{.typeVar}})

				if err != nil {
					log.Print(err)
					h.Redirect = "/error"
					return h,err
				} else {
					h.View["success"] = "Record created."
				}
			} else {
				// This is an update
				err = t.Update({{.typeVar}})
				if err != nil {
					log.Print(err)
					h.Redirect = "/error"
					return h,err
				} else {
					h.View["success"] = "Record updated."
				}

			}

		} else {
			h.View["errors"] = errors
		}

		// Pass back marshaled struct, even if it isn't valid, to allow correction of mistakes.
		h.View["{{.typeVar}}"] = {{.typeVar}}
		
	}
	if {{.typeVar}}.Id != -1 {
		h.View["update"] = true
	}
	return
}

func {{.typeName}}AdminListHandler(r *http.Request, a *framework.AppScope, rs *framework.RequestScope) (h framework.HandlerResponse, err error) {
	h.Init()

	t := &model.Table{Db: a.Db}
	{{.typeVar}} := &{{.typeName}}{}
	q := model.Query{}
	q.Order = model.MakeDbName("Id")
	{{.typeVar}}s, err := t.FetchAll({{.typeVar}}, q)
	if err == nil {
		h.View["{{.typeVar}}s"] = {{.typeVar}}s
	} else {
		h.Redirect = "/error"
	}

	return
}

func {{.typeName}}AdminDeleteHandler(r *http.Request, a *framework.AppScope, rs *framework.RequestScope) (h framework.HandlerResponse, err error) {
	h.Init()

	t := &model.Table{Db: a.Db}
	{{.typeVar}} := &{{.typeName}}{}

	{{.typeVar}}.Id = framework.GetIntId(rs.UrlParamMap["id"])
	if {{.typeVar}}.Id != -1 {
		err = t.Fetch({{.typeVar}})
		if err != nil {
			log.Print(err)
			h.Redirect = "/error"
			return
		} else {
			h.View["{{.typeVar}}"] = {{.typeVar}}
		}
	} else {
		log.Print("Delete {{.typeVar}} called without {{.typeVar}} id.")
		h.Redirect = "/error"
		return
	}

	h.View["{{.typeVar}}"] = {{.typeVar}}

	if r.Method == "POST" {
		t.Delete({{.typeVar}})
		h.Redirect = "/admin/{{.typeVar}}"
	}

	return
}
