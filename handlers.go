package httpwaymid

import (
	"net/http"
	"fmt"

	"github.com/corneldamian/httpway"
	"github.com/julienschmidt/httprouter"
)

func NotFound(router *httpway.Router) http.Handler {
	return &simpleServe{
		statusCode: http.StatusNotFound,
		status: "Page not found",
		router: router,
	}
}

func MethodNotAllowed(router *httpway.Router) http.Handler {
	return &simpleServe{
		statusCode: http.StatusMethodNotAllowed,
		status: "Method not allowed",
		router: router,
	}
}

type simpleServe struct {
	statusCode int
	status string
	router *httpway.Router
}

var one = 1
func (ss *simpleServe) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	httprouterHandler := func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("X-Content-Type-Options", "nosniff")

		if r.Header.Get("Content-Type") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(ss.statusCode)

			fmt.Fprintf(w, "{\"Error\":%q}", ss.status)
		} else {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(ss.statusCode)

			fmt.Fprintf(w, "%s", ss.status)
		}
	}

	h :=ss.router.GenerateChainHandler(httprouterHandler)
	h(w, r, make(httprouter.Params, 0))
}