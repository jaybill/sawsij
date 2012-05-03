package sawsij

import(    
    "net/http"
    "log"
    "fmt"
    "reflect"   
)

func LoginHandler(r *http.Request, a *AppScope, rs *RequestScope) (h HandlerResponse, err error) {
	h.Init()
    log.Printf("Setup: %+v",a.Setup.User)

	model := &Model{Db: a.Db}
	user := a.Setup.User
    
	if r.Method == "POST" {
        
		username := r.FormValue("username")
		password := r.FormValue("password")
		salt, err := a.Config.Get("encryption.salt")
		
		if err != nil {
			log.Println(err)
		}
		log.Println("Checking username/password")

		var q Query
		q.Where = fmt.Sprintf("username = $1")
		
		users, _ := model.FetchAll(user, q, username)
		log.Printf("Users: %+v",users)
		
		if len(users) != 1 {
			h.View["failed"] = bool(true)
			log.Printf("User %q does not exist", username)
		} else {		
			v := reflect.ValueOf(users[0])
			
			if !user.TestPassword(password, salt) {
				h.View["failed"] = bool(true)
				log.Printf("Password for %q is wrong", username)
			}
		}
		if h.View["failed"] == nil {
		    user.ClearPasswordHash()
			rs.Session.Values["user"] = user
			log.Printf("Logging in userId: %+v", rs.Session.Values["user"])
			h.Redirect = "/admin"
		} else {
			h.View["username"] = username
		}
		
	}

	return
}
