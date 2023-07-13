package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Freecil/GoPrjtBookings/internal/config"
	"github.com/Freecil/GoPrjtBookings/internal/models"
	"github.com/justinas/nosurf"
)

var functions = template.FuncMap{
	"humanDate":  HumanDate,
	"formatDate": FormatDate,
	"iterate":    Iterate,
	"add":        Add,
}

var app *config.AppConfig
var pathToTemplates = "./Templates"

func Add(a, b int) int {
	return a + b
}

// retruns a slice of intes start at 1 gpoing to count
func Iterate(count int) []int {
	var i int
	var items []int
	for i = 0; i < count; i++ {
		items = append(items, i)
	}
	return items
}

func NewRenderer(a *config.AppConfig) {
	app = a
}

// changes the date to a nicer looking format
func HumanDate(t time.Time) string {
	return t.Format("2006-01-02")

}

func FormatDate(t time.Time, f string) string {
	return t.Format(f)
}

func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "Flash")
	td.Error = app.Session.PopString(r.Context(), "Error")
	td.Warning = app.Session.PopString(r.Context(), "Warning")
	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = 1
	}

	td.CSRFToken = nosurf.Token(r)
	return td
}

// Renders template, does if for you no need to do it manually
func Template(w http.ResponseWriter, tmpl string, td *models.TemplateData, r *http.Request) error {
	var tc map[string]*template.Template

	if app.UseCache {
		// get the template cache from the app config
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	//get requested template from cache
	t, ok := tc[tmpl]
	if !ok {
		return errors.New("cant get templae from cahce")
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td, r)

	_ = t.Execute(buf, td)

	_, err := buf.WriteTo(w)
	if !ok {
		log.Println(err)
		return err
	}
	return nil
}

func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	//get all files enidng in .page.tmpl from /GoTemplates
	pages, err := filepath.Glob(fmt.Sprintf("%s/*page.tmpl", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	//range trough all the files
	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s/*layout.tmpl", pathToTemplates))
		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*layout.tmpl", pathToTemplates))
		}
		if err != nil {
			return myCache, err
		}

		myCache[name] = ts
	}

	return myCache, nil

}
