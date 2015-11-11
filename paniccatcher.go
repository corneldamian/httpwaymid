package httpwaymid

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/corneldamian/httpway"
	"fmt"
)

func PanicCatcher(w http.ResponseWriter, r *http.Request, pr httprouter.Params) {
	ctx:=httpway.GetContext(r)

	defer func() {

		if rec := recover(); rec != nil {
			if ctx.StatusCode() == 0 {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Internal Server error")
			}

			if !ctx.HasLog() {
				panic(rec)
			}

			ctx.Log().Error("Panic catched: %s", rec)
		}
	}()

	ctx.Next(w, r, pr)
}