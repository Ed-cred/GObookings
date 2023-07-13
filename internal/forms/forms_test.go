package forms

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Valid(t *testing.T) {
	req := httptest.NewRequest("POST", "/wahtever", nil)
	form := New(req.PostForm)
	isValid := form.Valid()
	if !isValid {
		t.Error("Invalid form, expected valid")
	}
}

func TestForm_Required(t *testing.T) {
	req := httptest.NewRequest("POST", "/wahtever", nil)
	form := New(req.PostForm)

	form.Required("a", "b", "c")
	if form.Valid() {
		t.Error("Got valid form but required fields are missing")
	}
	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "a")
	postedData.Add("c", "a")
	req = httptest.NewRequest("POST", "/wahtever", nil)
	req.PostForm = postedData
	form = New(req.PostForm)
	form.Required("a", "b", "c")
	if !form.Valid() {
		t.Error("Expected valid form, got an error")
	}
}

func TestForm_MinLength(t *testing.T) {
	data := url.Values{}
	f := New(data)
	y := f.MinLength("a", 1)
	if y {
		t.Error("Expected return value False, got True")
	}
	s := f.Errors.Get("a")
	if s == "" {
		t.Error("Expected an error type and got empty string")
	}
	data = url.Values{}
	data.Add("rand", "and")
	data.Add("rand", "bnd")
	form := New(data)
	x := form.MinLength("rand", 2)
	if !x {
		t.Error("Expected return value True, got False")
	}
	s = form.Errors.Get("rand")
	if s != "" {
		t.Error("Expected empty string, got error type")
	}
}

func TestForm_Has(t *testing.T) {
	postedData := url.Values{}
	form := New(postedData)
	x := form.Has("rand")
	if x {
		t.Error("Expected False, got True")
	}
	postedData = url.Values{}
	postedData.Add("a", "abca")
	// r := httptest.NewRequest("POST", "/ever", nil)
	// r.PostForm = postedData
	form = New(postedData)
	x = form.Has("a")
	if !x {
		t.Error("Expected return value True, got False")
	}
}

func TestForm_IsEmail(t *testing.T) {
	postedData := url.Values{}
	postedData.Add("email", "c@c")
	form := New(postedData)
	form.IsEmail("email")
	if form.Valid() {
		t.Error("Invalid email adress read as valid")
	}

	Data := url.Values{}
	Data.Add("email_valid", "c@c.com")
	f := New(Data)

	f.IsEmail("email_valid")
	if !f.Valid() {
		t.Error("Valid email adress returned error")
	}
}
