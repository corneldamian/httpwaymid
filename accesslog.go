package httpwaymid

import (
	"net"
	"net/http"
	"strings"
	"time"

	"fmt"
	"github.com/corneldamian/httpway"
	"github.com/julienschmidt/httprouter"
	"io"
)

//this handler will write to logger function (the one form parameter) the w3c access log
//this are the fields:
//#Fields: c-ip x-c-user date time cs-method cs-uri-stem cs-uri-query cs(X-Forwarded-For) sc-bytes sc-status time-taken
//when you first init or change logging file, call AccessLogHeader to write the w3c fields
func AccessLog(logger func(v ...interface{})) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, pr httprouter.Params) {
		ctx := httpway.GetContext(r)

		starttime := time.Now()

		ctx.Next(w, r, pr)

		username := "-"
		if ctx.HasSession() && ctx.Session().Username() != "" {
			username = ctx.Session().Username()
		}
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		query := r.URL.RawQuery
		if query == "" {
			query = "-"
		}

		xforwarded := r.Header.Get("X-Forwarded-For")
		if xforwarded != "" {
			tmpsplit := strings.Split(xforwarded, ",")
			if len(tmpsplit) > 0 {
				xforwarded = tmpsplit[len(tmpsplit)-1]
			}

			xforwarded = strings.Trim(xforwarded, " ")
		} else {
			xforwarded = "-"
		}

		statuscode := ctx.StatusCode()
		if statuscode == 0 {
			statuscode = 200
		}

		logger("%s %s %s %s %s %s %s %d %d %d",
			ip,
			username,
			starttime.UTC().Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.EscapedPath(),
			query,
			xforwarded,
			ctx.TransferedBytes(),
			statuscode,
			time.Since(starttime).Nanoseconds()/1000,
		)
	}
}

//write w3c access log header
func AccessLogHeader(logger func(v ...interface{})) {
	logger("#Version: 1.0")
	logger("#Fields: c-ip x-c-user date time cs-method cs-uri-stem cs-uri-query cs(X-Forwarded-For) sc-bytes sc-status time-taken")
	logger("#Software: httpway accesslog")
	logger("#Start-Date: %s", time.Now().UTC().Format("2006-01-02 15:04:05"))
}

//write w3c access log header using io.Writer
func AccessLogHeaderWriter(w io.Writer) {
	fmt.Fprint(w, "#Version: 1.0\n")
	fmt.Fprint(w, "#Fields: c-ip x-c-user date time cs-method cs-uri-stem cs-uri-query cs(X-Forwarded-For) sc-bytes sc-status time-taken\n")
	fmt.Fprint(w, "#Software: httpway accesslog\n")
	fmt.Fprintf(w, "#Start-Date: %s\n", time.Now().UTC().Format("2006-01-02 15:04:05"))
}
