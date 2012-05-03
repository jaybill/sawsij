package sawsij

import(    
    "net/http"
    "log"    
)

func LoginHandler(r *http.Request, a *AppScope, rs *RequestScope) (h HandlerResponse, err error) {
	h.Init()
	
	if r.Method == "POST" {
        
		username := r.FormValue("username")
		password := r.FormValue("password")
				
		if err != nil {
			log.Println(err)
		}
		log.Println("Checking username/password")
        user := a.Setup.GetUser(username,a)
        
        if user == nil{
            h.View["failed"] = true
        } else {
            if !user.TestPassword(password,a){
                h.View["failed"] = true
            }
        }
		
		if h.View["failed"] == nil {
		    user.ClearPasswordHash()
			rs.Session.Values["user"] = user.(User)
			log.Printf("Logging in userId: %+v", rs.Session.Values["user"])
			h.Redirect = "/admin"
		} else {
			h.View["username"] = username
		}
		
	}

	return
}

func LogoutHandler(r *http.Request, a *AppScope, rs *RequestScope) (h HandlerResponse, err error) {
	h.Init()
	rs.Session.Values = nil
	h.Redirect = "/"
	return
}

