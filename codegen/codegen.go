package codegen

var TplAppserverGo = 
`// Copyright <year> <name>. All rights reserved.
// Use of this source code is governed by license 
// that can be found in the LICENSE file.

package main

import (
	"{{ .name }}"
	"{{ .name }}/data"
	"bitbucket.org/jaybill/sawsij"
	"encoding/gob"
	"log"
	"net/http"
	"time"
)

func adminHandler(r *http.Request, a *sawsij.AppScope, rs *sawsij.RequestScope) (h sawsij.HandlerResponse, err error) {
	h.Init()
    h.View["time"] = time.Now()
	return
}

func indexHandler(r *http.Request, a *sawsij.AppScope, rs *sawsij.RequestScope) (h sawsij.HandlerResponse, err error) {
	h.Init()
	h.View["appname"] = "{{ .name }}"	
    h.View["time"] = time.Now()
	return
}

func main() {
    // Required so that our User type can be used by the framework in a session 
	gob.Register(&data.User{})

    // define some role arraysf
	admin := []int{constants.R_ADMIN}
	all := []int{constants.R_ADMIN, sawsij.R_GUEST, constants.R_MEMBER}

    // Create a new AppSetup type     
    as := new(sawsij.AppSetup)

    // Register Callback functions
	as.GetUser = GetUser

    // Configure the application
	sawsij.Configure(as,"")

    // Route patterns to handlers
	sawsij.Route(sawsij.RouteConfig{Pattern: "/", Handler: indexHandler, Roles: all})
	sawsij.Route(sawsij.RouteConfig{Pattern: "/admin", Handler: adminHandler, Roles: admin})

    // Start the server
	sawsij.Run()
}`

var TplUserGo = 
`// Copyright <year> <name>. All rights reserved.
// Use of this source code is governed by license 
// that can be found in the LICENSE file.

package data

import (
	"time"
	"crypto/md5"
	"io"
	"fmt"
	"bitbucket.org/jaybill/sawsij"
)

type User struct {
	Id           int64
	Username     string
	PasswordHash string
	FullName     string
	Email        string
	CreatedOn    time.Time
	Role         int64
}

func (u *User) SetPassword(password string,salt string){
    h := md5.New()
    io.WriteString(h,salt)
    io.WriteString(h,password)
    u.PasswordHash = fmt.Sprintf("%x",h.Sum(nil))
}

func (u *User) TestPassword(password string, a *sawsij.AppScope) (valid bool){
    valid = false
    salt, _ := a.Config.Get("encryption.salt")
    
    h := md5.New()
    if salt != ""{
        io.WriteString(h,salt)
    }   
    
    io.WriteString(h,password)
    tHash := fmt.Sprintf("%x",h.Sum(nil))
    
    if u.PasswordHash == tHash {
       valid = true
    } 
    return
}

func (u *User) GetRole() int64{
    return u.Role
}

func (u *User) ClearPasswordHash(){
    u.PasswordHash = ""
}`

var TplConstantsGo =
`// Copyright <year> <name>. All rights reserved.
// Use of this source code is governed by license 
// that can be found in the LICENSE file.

package constants

const(
    R_MEMBER = 2
    R_ADMIN = 3
)
`

var TplConfigYaml = 
`# Copyright <year> <name>. All rights reserved.
# Use of this source code is governed by license 
# that can be found in the LICENSE file.

server:
  port: {{ .port }}
  cacheTemplates: false

database:
  driver: {{ .driver }}
  connect: user={{ .dbuser }} password={{ .dbpass }} dbname={{ .dbname }} sslmode={{ .ssl }}
  schema: {{ .schema }}

encryption:
  salt: {{ .salt }}
  key: {{ .key }}`
  
var TplIndexHtml =
`
<% .welcome %>
`
var TplHeaderHtml =``
var TplFooterHtml =``

