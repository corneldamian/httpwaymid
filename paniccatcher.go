package httpwaymid

import (
	"net/http"

	"fmt"
	"github.com/corneldamian/httpway"
	"github.com/julienschmidt/httprouter"
	"runtime"
)

//will catch a panic and if the logger is set will log it, if not will panic again
func PanicCatcher(w http.ResponseWriter, r *http.Request, pr httprouter.Params) {
	ctx := httpway.GetContext(r)

	defer func() {

		if rec := recover(); rec != nil {
			if ctx.StatusCode() == 0 {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Internal Server error")

				if _, ok := w.(http.Flusher); ok {
					println("HAHAHAHAHA FLUSHER")
				}
			}

			if !ctx.HasLog() {
				panic(rec)
			}

			file, line := getFileLine()
			ctx.Log().Error("Panic catched on %s:%d - %s", file, line, rec)

		}
	}()

	ctx.Next(w, r, pr)
}

func getFileLine() (file string, line int) {
	_, file, line, ok := runtime.Caller(4)
	if !ok {
		file = "???"
		line = 0
	}

	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	file = short
	return
}
