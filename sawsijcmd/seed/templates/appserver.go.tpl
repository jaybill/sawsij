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
)

// Returns a type that conforms to the framework.User interface.
func GetUser(username string, a *framework.AppScope) (user framework.User) {
	t := &model.Table{Db: a.Db}
	dbuser := &{{ .name }}.User{}
	q := model.Query{Where: "username = $1"}
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
	admin := []int{{{ .name }}.R_ADMIN}
	all := []int{{{ .name }}.R_ADMIN, framework.R_GUEST, {{ .name }}.R_MEMBER}

	// Create a new AppSetup  
	as := new(framework.AppSetup)

	// Register Callback functions and roles
	as.GetUser = GetUser
	as.Roles = &map[string]int{"admin": {{ .name }}.R_ADMIN, "guest": framework.R_GUEST, "member": {{ .name }}.R_MEMBER}

	// Configure the application
	framework.Configure(as, "")

	// Route patterns to handlers
	framework.Route(framework.RouteConfig{Pattern: "/", Handler: indexHandler, Roles: all})
	framework.Route(framework.RouteConfig{Pattern: "/admin", Handler: adminHandler, Roles: admin})
	framework.Route(framework.RouteConfig{Pattern: "/admin/users", Handler: {{ .name }}.UserAdminListHandler, Roles: admin})
	framework.Route(framework.RouteConfig{Pattern: "/admin/users/edit", Handler: {{ .name }}.UserAdminEditHandler, Roles: admin})
	framework.Route(framework.RouteConfig{Pattern: "/admin/users/delete", Handler: {{ .name }}.UserAdminDeleteHandler, Roles: admin})
	framework.Route(framework.RouteConfig{Pattern: "/login", Handler: framework.LoginHandler, Roles: all})
	framework.Route(framework.RouteConfig{Pattern: "/logout", Handler: framework.LogoutHandler, Roles: all})
	framework.Route(framework.RouteConfig{Pattern: "/denied", Handler: framework.DeniedHandler, Roles: all})
	framework.Route(framework.RouteConfig{Pattern: "/error", Handler: framework.ErrorHandler, Roles: all})

	// Start the server
	framework.Run()
}
