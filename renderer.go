package httpwaymid

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/corneldamian/httpway"
)

//a simple JSON renderer
//from your handler you need to set on context your data (and optional the http status code)
//you need to set here what are those keys where you are going to save them
//data from ctx key "ctxDataVar" must be and object that can be marshal by json.Marshal
//data from ctx key "ctxHttpStatusCodeVar" must be int
func JSONRenderer(ctxDataVar, ctxHttpStatusCodeVar string) httpway.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := httpway.GetContext(r)

		ctx.Next(w, r)

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
func TemplateRenderer(templateDir, ctxTemplateNameVar, ctxTempalteDataVar, ctxHttpStatusCodeVar string) httpway.Handler {

	templates := template.Must(parseFiles(templateDir, getAllTemplatesFiles(templateDir)...))

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := httpway.GetContext(r)

		ctx.Next(w, r)

		if ctx.Has(ctxTempalteDataVar) && ctx.Has(ctxTemplateNameVar) {
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

func getAllTemplatesFiles(templateDirName string) []string {
	templatePages := make([]string, 0)

	filepath.Walk(templateDirName, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			templatePages = append(templatePages, path)
		}
		return nil
	})

	return templatePages
}

func parseFiles(templateDir string, filenames ...string) (*template.Template, error) {
	if len(filenames) == 0 {
		return nil, fmt.Errorf("template: no files named in call to ParseFiles")
	}

	if templateDir[len(templateDir)-1] != '/' {
		templateDir = filepath.Clean(templateDir) + "/"
	}

	var t *template.Template

	for _, filename := range filenames {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		s := string(b)
		name := filename[len(templateDir):]

		var tmpl *template.Template
		if t == nil {
			t = template.New(name)
		}
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}

		_, err = tmpl.Parse(s)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}
