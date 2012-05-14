// Copyright 2012 J. William McCarthy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sawsij

import (
	"encoding/base64"
	"log"
	"net/http"
)

// LoginHandler can be used by applications as a handler for authentication. It uses the GetUser() function you supply to
// in AppSetup and the TestPassword() function implemented in the User type. If the login credentials are valid, the handler
// will place the user data in the session, calling ClearPasswordHash() before it does so.
func LoginHandler(r *http.Request, a *AppScope, rs *RequestScope) (h HandlerResponse, err error) {
	h.Init()
    var dest string

	if rs.UrlParams["dest"] != "" {		
		if err != nil{
		    log.Print(err)
		} else {
            h.View["dest"] = rs.UrlParams["dest"]
		}
	} else {
	    log.Println("No destination specified, will redirect to /")
	}

	if r.Method == "POST" {

		username := r.FormValue("username")
		password := r.FormValue("password")
        dest64    := r.FormValue("dest")

        if dest64 != ""{
            bDest, err := base64.URLEncoding.DecodeString(dest64)
            if err != nil{
                log.Print(err)
            } else {
                 dest = string(bDest)
            }
            
        }

		if err != nil {
			log.Println(err)
		}
		log.Println("Checking username/password")
		user := a.Setup.GetUser(username, a)

		if user == nil {
			h.View["failed"] = true
		} else {
			if !user.TestPassword(password, a) {
				h.View["failed"] = true
			}
		}

		if h.View["failed"] == nil {
			user.ClearPasswordHash()
			rs.Session.Values["user"] = user
			log.Printf("Logging in userId: %+v", rs.Session.Values["user"])
			if dest != ""{
			    h.Redirect = dest
			} else {
			    h.Redirect = "/"
			}
			
		} else {
			h.View["username"] = username
		}

	}

	return
}

// A handler that you can use for the pattern "/denied", which is where requests will be sent when the user 
// attempts to go to a page they do not have the right role for.
func DeniedHandler(r *http.Request, a *AppScope, rs *RequestScope) (h HandlerResponse, err error) {    
    return
}

// A handler that can be used for clearing the session, logging the user out. No template is required.
func LogoutHandler(r *http.Request, a *AppScope, rs *RequestScope) (h HandlerResponse, err error) {
	h.Init()
	rs.Session.Values = nil
	h.Redirect = "/"
	return
}
