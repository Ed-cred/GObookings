package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name               string
	url                string
	method             string
	params             []postData
	expectedStatusCode int
}{
	{"home", "/", "GET", []postData{}, http.StatusOK},
	{"about", "/about", "GET", []postData{}, http.StatusOK},
	{"gq", "/generals-quarters", "GET", []postData{}, http.StatusOK},
	{"ms", "/majors-suite", "GET", []postData{}, http.StatusOK},
	{"sa", "/search_availability", "GET", []postData{}, http.StatusOK},
	{"contact", "/contact", "GET", []postData{}, http.StatusOK},
	{"mr", "/make_reservation", "GET", []postData{}, http.StatusOK},
	{"mr", "/make_reservation", "GET", []postData{}, http.StatusOK},
	{"post_search_avail", "/search_availability", "POST", []postData{
		{key:"start", value:"01-01-2020"},
		{key:"end", value:"09-02-2020"},
	}, http.StatusOK},
	{"post_search_avai_json", "/search_availability-json", "POST", []postData{
		{key:"start", value:"01-01-2020"},
		{key:"end", value:"09-02-2020"},
	}, http.StatusOK},
	{"make_reservation", "/make_reservation", "POST", []postData{
		{key:"first-name", value:"John"},
		{key:"last-name", value:"Doe"},
		{key:"email", value:"john@doe.com"},
		{key:"phone", value:"079292345"},
	}, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, e := range theTests {
		if e.method == "GET" {
		resp, err :=  ts.Client().Get(ts.URL + e.url)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}
		if resp.StatusCode != e.expectedStatusCode {
			t.Errorf("For %s, expected status code %d, got %d",e.name, e.expectedStatusCode, resp.StatusCode)
		}

		} else{
			values := url.Values{}
			for _, x := range e.params {
				values.Add(x.key, x.value)
			}
			resp, err :=ts.Client().PostForm(ts.URL+e.url, values)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}
			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("For %s, expected status code %d, got %d",e.name, e.expectedStatusCode, resp.StatusCode)
			}

		}
	}
}
