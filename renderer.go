package httpwaymid

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/corneldamian/httpway"
	"github.com/julienschmidt/httprouter"
	"path/filepath"
)

//a simple JSON renderer
//from your handler you need to set on context your data (and optional the http status code)
//you need to set here what are those keys where you are going to save them
//data from ctx key "ctxDataVar" must be and object that can be marshal by json.Marshal
//data from ctx key "ctxHttpStatusCodeVar" must be int
func JSONRenderer(ctxDataVar, ctxHttpStatusCodeVar string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, pr httprouter.Params) {
		ctx := httpway.GetContext(r)

		ctx.Next(w, r, pr)

		if ctx.Has(ctxDataVar) {
			if ctx.StatusCode() != 0 {
				ctx.Log().Error("I have json from context to write but the status already set")
				return
			}

			w.Header().Set("Content-Type", "application/json")

			data, err := json.Marshal(ctx.Get(ctxDataVar))
			if err != nil {
				ctx.Log().Error("I was unable to marshal to json: %s", err)
				w.Write([]byte("{\"Error\": \"Internal server error\"}"))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if ctx.Has(ctxHttpStatusCodeVar) {
				w.WriteHeader(ctx.Get(ctxHttpStatusCodeVar).(int))
			}

			n, err := w.Write(data)
			if err != nil {
				ctx.Log().Error("I was unable to write payload to the socket: %s", err)
				return
			}

			if n != len(data) {
				ctx.Log().Error("Not all data was written to the socket expected: %d sent %d (fix this with some retry's)", len(data), n)
			}
		}
	}
}

// a simple golang Template Renderer
// this will scan the templeteDir and load all the files that ends in tmpl
// in handler you must set "ctxTemplateNameVar" wich is the template file name and
// "ctxTempalteDataVar" the data to pass to the template
// you can set a diffrent (the 200) http status code with "ctxHttpStatusCodeVar"
func TemplateRenderer(templateDir, ctxTemplateNameVar, ctxTempalteDataVar, ctxHttpStatusCodeVar string) httprouter.Handle {
	templates := template.Must(template.ParseGlob(filepath.Join(templateDir, "*.tmpl")))
	println("l")
	return func(w http.ResponseWriter, r *http.Request, pr httprouter.Params) {
		ctx := httpway.GetContext(r)

		ctx.Next(w, r, pr)
		println("t")
		if ctx.Has(ctxTempalteDataVar) && ctx.Has(ctxTemplateNameVar) {
			println("t1")
			if ctx.Has(ctxHttpStatusCodeVar) {
				w.WriteHeader(ctx.Get(ctxHttpStatusCodeVar).(int))
			}

			err := templates.ExecuteTemplate(w, ctx.Get(ctxTemplateNameVar).(string), ctx.Get(ctxTempalteDataVar))
			if err != nil {
				ctx.Log().Error("I was unable to execute template %s with error: %s", ctx.Get(ctxTemplateNameVar), err)
			}
		}
	}
}
