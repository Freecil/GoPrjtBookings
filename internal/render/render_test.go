package render

import (
	"net/http"
	"testing"

	"github.com/Freecil/GoPrjtBookings/internal/models"
)

func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData
	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	session.Put(r.Context(), "Flash", "123")

	result := AddDefaultData(&td, r)
	if result.Flash != "123" {
		t.Error("Flash value of 123 not found")
	}
}

func TestRenderTemplate(t *testing.T) {
	pathToTemplates = "./../../templates"
	tc, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}

	app.TemplateCache = tc

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	var ww myWriter

	err = Template(&ww, "home.page.tmpl", &models.TemplateData{}, r)
	if err != nil {
		t.Error("Error writing  template to browser")
	}

	err = Template(&ww, "non.page.tmpl", &models.TemplateData{}, r)
	if err == nil {
		t.Error("Render non existing template")
	}
}

func getSession() (*http.Request, error) {
	r, err := http.NewRequest("Get", "/", nil)
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	ctx, _ = session.Load(ctx, r.Header.Get("X-Session"))

	r = r.WithContext(ctx)
	return r, nil
}

func TestNewTemplates(t *testing.T) {
	NewRenderer(app)
}

func TestCreateTemplateCache(t *testing.T) {
	pathToTemplates = "./../../templates"
	_, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}

}
