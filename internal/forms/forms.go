package forms

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
)

// Creates a custom form stuct
type Form struct {
	url.Values
	Errors errors
}

// Return true if there is no erros
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// Initalizes a form struct
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cant be blank")
		}
	}
}

// check if form field is in post and not empty
func (f *Form) Has(field string) bool {
	x := f.Get(field)
	if x == "" {
		return false
	}
	return true
}

// Check for string minimum length
func (f *Form) MinLength(field string, length int) bool {
	x := f.Get(field)
	if len(x) < length {
		f.Errors.Add(field, fmt.Sprintf("This field must be att least %d characters long", length))
		return false
	}
	return true
}

// Checks for valid email address
func (f *Form) IsEmail(field string) {
	if !govalidator.IsEmail(f.Get(field)) {
		f.Errors.Add(field, "Invalid email address")
	}
}
