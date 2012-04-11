package sawsij

import(	
	"fmt"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request, c *Context) {    
	
	fmt.Fprintf(w, "Port is %s!", c.Config.GetString("port"))
	
}
