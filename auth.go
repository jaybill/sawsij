package sawsij

import(    
    "net/http" 
    "log"       
)

// LoginHandler is the built in method for handling auth. It can be used as is or called by a function you write that does
// additional things, like SSO.
func LoginHandler(r *http.Request, c *Context, u map[string](string)) (h HandlerResponse, err error) {
	h.Init()

    if r.Method == "POST" {
        log.Println("Checking username/password")
        if r.FormValue("username") == "jaybill" && r.FormValue("password") == "password"{
            h.Redirect = "/admin"
        }    
                 
    }

    return	
}
