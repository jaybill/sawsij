package sawsij

import(    
    "net/http"        
)

func LoginHandler(r *http.Request, c *Context, u map[string](string)) (view map[string](interface{}), err error) {
	view = make(map[string]interface{})

    if r.Method == "POST" {
        if r.FormValue("username") == "jaybill" && r.FormValue("password") == "password"{
        
        }      
                 
    }

    return	
}
