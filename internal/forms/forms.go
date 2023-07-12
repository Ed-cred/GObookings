package forms

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
)

type Form struct {
	url.Values
	Errors errors
}

// New initializes the form struct
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

//Required checks for required fields in the form
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

//MinLength checks a field for a minimum length
func (f *Form) MinLength(field string, length int, r *http.Request) bool {
	x := r.Form.Get(field)
	if len(x) < length {
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long", length))
		return false
	}
	return true
}

// Has checks if form field is in post and not empty
func (f *Form) Has(field string, r *http.Request) bool {
	x := r.Form.Get(field)
	if x != "" {
		return true
	}
	return false
}

//Checks for vaild email adress
func (f *Form) IsEmail(field string) {
	if !govalidator.IsEmail(f.Get(field)) {
		f.Errors.Add(field, "Invalid email adress")
	}
}

// returns true if there are no errors, otherwise false
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}
