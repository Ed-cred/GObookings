package handlers

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ed-cred/bookings/internal/models"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name               string
	url                string
	method             string
	expectedStatusCode int
}{
	{"home", "/", "GET",  http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"gq", "/generals-quarters", "GET" , http.StatusOK},
	{"ms", "/majors-suite", "GET", http.StatusOK},
	{"sa", "/search_availability", "GET", http.StatusOK},
	{"contact", "/contact", "GET", http.StatusOK},

// 	{"post_search_avail", "/search_availability", "POST", []postData{
// 		{key:"start", value:"01-01-2020"},
// 		{key:"end", value:"09-02-2020"},
// 	}, http.StatusOK},
// 	{"post_search_avai_json", "/search_availability-json", "POST", []postData{
// 		{key:"start", value:"01-01-2020"},
// 		{key:"end", value:"09-02-2020"},
// 	}, http.StatusOK},
// 	{"make_reservation", "/make_reservation", "POST", []postData{
// 		{key:"first-name", value:"John"},
// 		{key:"last-name", value:"Doe"},
// 		{key:"email", value:"john@doe.com"},
// 		{key:"phone", value:"079292345"},
// 	}, http.StatusOK},
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

		}
	}
}



func TestRepoReservation(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID: 1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/make_reservation", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)
	handler := http.HandlerFunc(Repo.Reservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Reservation handler returned %v, expected %v", rr.Code, http.StatusOK)
	}
	//test case when reservation is not in session
	req, _ = http.NewRequest("GET", "/make_reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned %v, expected %v", rr.Code, http.StatusOK)
	}

	//test with non-existant room
	req, _ = http.NewRequest("GET", "/make_reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	reservation.RoomID = 100
	session.Put(ctx, "reservation", reservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned %v, expected %v", rr.Code, http.StatusOK)
	}
}


func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}