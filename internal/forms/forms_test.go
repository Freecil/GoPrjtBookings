package forms

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Valid(t *testing.T) {
	r := httptest.NewRequest("POST", "/Whatever", nil)
	form := New(r.PostForm)

	isValid := form.Valid()
	if !isValid {
		t.Error("got invalid when should be valid")
	}

}

func TestForm_Required(t *testing.T) {
	r := httptest.NewRequest("POST", "/Whatever", nil)
	form := New(r.PostForm)

	form.Required("a", "b", "c")
	if form.Valid() {
		t.Error("got invalid when should be valid")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "a")
	postedData.Add("c", "a")

	r, _ = http.NewRequest("POST", "/Whatever", nil)

	r.PostForm = postedData
	form = New(r.PostForm)
	form.Required("a", "b", "c")
	if !form.Valid() {
		t.Error(" does not have the required fields when it does")
	}
}

func TestForm_Has(t *testing.T) {
	postedData := url.Values{}
	form := New(postedData)

	has := form.Has("whaterver")
	if has {
		t.Error("form shows fiedl has field when it does not")
	}

	postedData = url.Values{}
	postedData.Add("a", "a")

	form = New(postedData)
	has = form.Has("a")
	if !has {
		t.Error("form shows does not have fiedl when it shoudl")
	}

}

func TestForms_MinLength(t *testing.T) {
	postedData := url.Values{}
	form := New(postedData)

	form.MinLength("x", 10)
	if form.Valid() {
		t.Error("the field shows lenght when it shoudl not")
	}

	isError := form.Errors.Get("x")
	if isError == "" {
		t.Error("error shoudle exist but is not")
	}

	postedData = url.Values{}
	postedData.Add("some_field", "some value")
	form = New(postedData)

	form.MinLength("some_field", 100)
	if form.Valid() {
		t.Error("minlenght shows valuve has met required 100 when it does not ")
	}

	postedData = url.Values{}
	postedData.Add("another_field", "home")
	form = New(postedData)

	form.MinLength("another_field", 2)
	if !form.Valid() {
		t.Error("minlenght shows valuve has not met required when it does ")
	}

	isError = form.Errors.Get("another_field")
	if isError != "" {
		t.Error("error shoudle not exist but does ")
	}
}

func TestForms_IsEmail(t *testing.T) {
	postedData := url.Values{}
	form := New(postedData)

	form.IsEmail("x")
	if form.Valid() {
		t.Error("form shows valid email when non exists")
	}

	postedData = url.Values{}
	postedData.Add("email", "c@c.com")
	form = New(postedData)

	form.IsEmail("email")
	if !form.Valid() {
		t.Error("form shows not valid when email is ")
	}

	postedData = url.Values{}
	postedData.Add("email", "x")
	form = New(postedData)

	form.IsEmail("email")
	if form.Valid() {
		t.Error("form shows  valid when email is not ")
	}
}
