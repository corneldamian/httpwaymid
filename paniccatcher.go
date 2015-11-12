package httpwaymid

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/corneldamian/httpway"
	"fmt"
)

//will catch a panic and if the logger is set will log it, if not will panic again
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

			log := ctx.Log()

			var (
				l setCallDepth
				ok bool
			)

			if l, ok=log.(setCallDepth); ok {
				l.SetFileDepth(11)
			}

			log.Error("Panic catched: %s", rec)

			if ok {
				l.SetFileDepth(0) //reset
			}
		}
	}()

	ctx.Next(w, r, pr)
}

type setCallDepth interface {
	SetFileDepth(depth int)
}