package sawsij

import (
	"fmt"
	"html"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request, c *Context) {

	//fmt.Fprintf(w, "Port is %s!", c.Config.GetString("port"))
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))

}
