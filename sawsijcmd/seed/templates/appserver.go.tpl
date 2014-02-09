// Copyright <year> <name>. All rights reserved.
// Use of this source code is governed by license 
// that can be found in the LICENSE file.

package main

import (
	"bitbucket.org/jaybill/sawsij/framework"
	"bitbucket.org/jaybill/sawsij/framework/model"
	"encoding/gob"
	"{{ .name }}"
	"log"
	"net/http"
	"time"
	"fmt"
)

// Returns a type that conforms to the framework.User interface.
func GetUser(username string, a *framework.AppScope) (user framework.User) {
	t := &model.Table{Db: a.Db}
	dbuser := &{{ .name }}.User{}
	q := model.Query{Where: fmt.Sprintf("username = %v",a.Db.GetQueries().P(1))}
	users, _ := t.FetchAll(dbuser, q, username)
	if len(users) == 1 {
		user = users[0].(*{{ .name }}.User)
	}
	return
}

// Handles the admin landing page.
func adminHandler(r *http.Request, a *framework.AppScope, rs *framework.RequestScope) (h framework.HandlerResponse, err error) {
	h.Init()
	h.View["time"] = time.Now()
	return
}

// Handles the main application landing page.
func indexHandler(r *http.Request, a *framework.AppScope, rs *framework.RequestScope) (h framework.HandlerResponse, err error) {
	h.Init()
	h.View["time"] = time.Now()
	return
}

func main() {
	log.Print("Starting {{ .name }}...")

	// Required so that our User type can be used by the framework in a session 
	gob.Register(&{{ .name }}.User{})

	// define some role arrays

	rg := map[string][]int{
		"admin": []int{ {{ .name }}.R_ADMIN},
		"all":   []int{ {{ .name }}.R_ADMIN, framework.R_GUEST, {{ .name }}.R_MEMBER},
	}

	// Create a new AppSetup  
	as := new(framework.AppSetup)

	// Register Callback functions and roles
	as.GetUser = GetUser
	as.Roles = &map[string]int{"admin": {{ .name }}.R_ADMIN, "guest": framework.R_GUEST, "member": {{ .name }}.R_MEMBER}

	// Configure the application
	framework.Configure(as, "")

	// Route patterns to handlers
	framework.Route(framework.RouteConfig{Pattern: "/", Handler: indexHandler, Roles: rg["all"]})
	framework.Route(framework.RouteConfig{Pattern: "/admin", Handler: adminHandler, Roles: rg["admin"]})
	framework.Route(framework.RouteConfig{Pattern: "/admin/users", Handler: {{ .name }}.UserAdminListHandler, Roles: rg["admin"]})
	framework.Route(framework.RouteConfig{Pattern: "/admin/users/edit", Handler: {{ .name }}.UserAdminEditHandler, Roles: rg["admin"]})
	framework.Route(framework.RouteConfig{Pattern: "/admin/users/delete", Handler: {{ .name }}.UserAdminDeleteHandler, Roles: rg["admin"]})
	framework.Route(framework.RouteConfig{Pattern: "/login", Handler: framework.LoginHandler, Roles: rg["all"]})
	framework.Route(framework.RouteConfig{Pattern: "/logout", Handler: framework.LogoutHandler, Roles: rg["all"]})
	framework.Route(framework.RouteConfig{Pattern: "/denied", Handler: framework.DeniedHandler, Roles: rg["all"]})
	framework.Route(framework.RouteConfig{Pattern: "/error", Handler: framework.ErrorHandler, Roles: rg["all"]})

	// Custom Routes

	// Start the server
	framework.Run()
}
