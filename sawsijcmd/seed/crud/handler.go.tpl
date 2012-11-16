package {{.pName}}

import (
	"bitbucket.org/jaybill/sawsij/framework"
	"bitbucket.org/jaybill/sawsij/framework/model"	
	"log"
	"net/http"{{if .importTime}}
	"time"{{ end }}
)

type {{.typeName}} struct { {{ range $field := .struct }}
	{{$field.FName}} {{$field.FType}}{{ end }} 
}

func (o *{{.typeName}}) GetValidationErrors(a *framework.AppScope) (errors []string) {
	// Add validation here

	return
}

func Admin{{.typeName}}EditHandler(r *http.Request, a *framework.AppScope, rs *framework.RequestScope) (h framework.HandlerResponse, err error) {
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

		err = r.ParseForm()
		if err != nil {
			log.Print(err)
			h.Redirect = "/error"
			return
		}

		err = decoder.Decode({{.typeVar}}, r.Form)

		if err != nil {
			log.Printf("Can't map: %v", err)
		} else {
			errors := {{.typeVar}}.GetValidationErrors(a)
			if len(errors) == 0 {
				if {{.typeVar}}.Id == -1 {
					// This is an insert

					err = t.Insert({{.typeVar}})

					if err != nil {
						log.Print(err)
						h.Redirect = "/error"
						return
					} else {
						h.View["success"] = "Record created."
					}
				} else {
					// This is an update
					err = t.Update({{.typeVar}})
					if err != nil {
						log.Print(err)
						h.Redirect = "/error"
						return
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
	}
	if {{.typeVar}}.Id != -1 {
		h.View["update"] = true
	}
	return
}

func Admin{{.typeName}}ListHandler(r *http.Request, a *framework.AppScope, rs *framework.RequestScope) (h framework.HandlerResponse, err error) {
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

func Admin{{.typeName}}DeleteHandler(r *http.Request, a *framework.AppScope, rs *framework.RequestScope) (h framework.HandlerResponse, err error) {
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
