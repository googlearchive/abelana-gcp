package wwwsite

import (
	"fmt"
	"net/http"

	//    "appengine"
)

func init() {
	http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, world!")
}
