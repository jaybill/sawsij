// Copyright <year> <name>. All rights reserved.
// Use of this source code is governed by license 
// that can be found in the LICENSE file.

package main

import (
	"{{ .name }}"	
	"bitbucket.org/jaybill/sawsij/framework"
	"encoding/gob"
	"log"
	"net/http"
	"time"
)

func GetUser(username string,a *framework.AppScope)(user framework.User){
    model := &framework.Model{Db: a.Db}
	dbuser  := &{{ .name }}.User{}
	q := framework.Query{Where: "username = $1"}		
	users, _ := model.FetchAll(dbuser, q, username)
	if len(users) == 1{
	    user = users[0].(*{{ .name }}.User)
	}	
    return
}

func adminHandler(r *http.Request, a *framework.AppScope, rs *framework.RequestScope) (h framework.HandlerResponse, err error) {
	h.Init()
    h.View["time"] = time.Now()
	return
}

func indexHandler(r *http.Request, a *framework.AppScope, rs *framework.RequestScope) (h framework.HandlerResponse, err error) {
	h.Init()
	h.View["appname"] = "{{ .name }}"	
    h.View["time"] = time.Now()
	return
}

func main() {
    log.Print("Starting {{ .name }}...")

    // Required so that our User type can be used by the framework in a session 
	gob.Register(&{{ .name}}.User{})

    // define some role arrays
	admin := []int{ {{ .name }}.R_ADMIN}
	all := []int{ {{ .name }}.R_ADMIN, framework.R_GUEST,  {{ .name }}.R_MEMBER}

    // Create a new AppSetup  
    as := new(framework.AppSetup)

    // Register Callback functions
	as.GetUser = GetUser

    // Configure the application
	framework.Configure(as,"")

    // Route patterns to handlers
	framework.Route(framework.RouteConfig{Pattern: "/", Handler: indexHandler, Roles: all})
	framework.Route(framework.RouteConfig{Pattern: "/admin", Handler: adminHandler, Roles: admin})

    // Start the server
	framework.Run()
}
