package sawsij

import (
	"encoding/base64"
	"log"
	"net/http"
)

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

func DeniedHandler(r *http.Request, a *AppScope, rs *RequestScope) (h HandlerResponse, err error) {    
    return
}

func LogoutHandler(r *http.Request, a *AppScope, rs *RequestScope) (h HandlerResponse, err error) {
	h.Init()
	rs.Session.Values = nil
	h.Redirect = "/"
	return
}
