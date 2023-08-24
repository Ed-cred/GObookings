package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Ed-cred/bookings/internal/driver"
	"github.com/Ed-cred/bookings/internal/models"
)

var theTests = []struct {
	name               string
	url                string
	method             string
	expectedStatusCode int
}{
	{"home", "/", "GET", http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"gq", "/generals-quarters", "GET", http.StatusOK},
	{"ms", "/majors-suite", "GET", http.StatusOK},
	{"sa", "/search_availability", "GET", http.StatusOK},
	{"contact", "/contact", "GET", http.StatusOK},
	{"non_existant", "/green/eggs/and/ham", "GET", http.StatusNotFound},
	// new routes
	{"login", "/user/login", "GET", http.StatusOK},
	{"logout", "/user/logout", "GET", http.StatusOK},
	{"dashboard", "/admin/dashboard", "GET", http.StatusOK},
	{"new_res", "/admin/reservations_new", "GET", http.StatusOK},
	{"all_res", "/admin/reservations_all", "GET", http.StatusOK},
	{"show_res", "/admin/reservations/new/10/show", "GET", http.StatusOK},
	{"show_res_cal", "/admin/reservations_calendar", "GET", http.StatusOK},
	{"show_res_cal_with_params", "/admin/reservations_calendar?y=2020&m=1", "GET", http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, e := range theTests {
		if e.method == "GET" {
			resp, err := ts.Client().Get(ts.URL + e.url)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}
			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("For %s, expected status code %d, got %d", e.name, e.expectedStatusCode, resp.StatusCode)
			}

		}
	}
}

func TestRepoReservation(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
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
	// test case when reservation is not in session
	req, _ = http.NewRequest("GET", "/make_reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned %v, expected %v", rr.Code, http.StatusOK)
	}

	// test with non-existant room
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

func TestNewRepo(t *testing.T) {
	var db driver.DB
	testRepo := NewRepository(&app, &db)

	if reflect.TypeOf(testRepo).String() != "*handlers.Repository" {
		t.Errorf("Did not get correct type from NewRepo: got %s, wanted *Repository", reflect.TypeOf(testRepo).String())
	}
}

func TestRepoPostReservation(t *testing.T) {
	var sd time.Time
	var ed time.Time
	sd = sd.AddDate(2060, 01, 01)
	ed = ed.AddDate(2060, 01, 02)
	reservation := models.Reservation{
		StartDate: sd,
		EndDate:   ed,
	}

	postedData := url.Values{}
	postedData.Add("start_date", "2050-01-01")
	postedData.Add("end_date", "2050-01-02")
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@smith.com")
	postedData.Add("phone", "076859432")
	postedData.Add("room_id", "1")

	req, _ := http.NewRequest("POST", "/make_reservation", strings.NewReader(postedData.Encode()))
	ctx := getCtx(req)
	session.Put(ctx, "reservation", reservation)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned %v for correct request, expected %v", rr.Code, http.StatusSeeOther)
	}

	// test for missing post body
	req, _ = http.NewRequest("POST", "/make_reservation", nil)
	ctx = getCtx(req)
	session.Put(ctx, "reservation", reservation)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned %v for missing post body, expected %v", rr.Code, http.StatusTemporaryRedirect)
	}

	// test for missing session data
	req, _ = http.NewRequest("POST", "/make_reservation", strings.NewReader(postedData.Encode()))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned %v for missing session data, expected %v", rr.Code, http.StatusTemporaryRedirect)
	}

	// test for wrong start date and end date format
	sd = sd.AddDate(010, 023, 11)
	ed = ed.AddDate(210, 123, 123)
	reservation = models.Reservation{
		StartDate: sd,
		EndDate:   ed,
	}

	req, _ = http.NewRequest("POST", "/make_reservation", nil)
	ctx = getCtx(req)
	session.Put(ctx, "reservation", reservation)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned %v for missing session data, expected %v", rr.Code, http.StatusTemporaryRedirect)
	}

	// test for invalid data
	sd = sd.AddDate(2060, 01, 01)
	ed = ed.AddDate(2060, 01, 02)
	reservation = models.Reservation{
		StartDate: sd,
		EndDate:   ed,
	}

	postedData = url.Values{}
	postedData.Add("start_date", "2050-01-01")
	postedData.Add("end_date", "2050-01-02")
	postedData.Add("first_name", "J")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@smith.com")
	postedData.Add("phone", "076859432")
	postedData.Add("room_id", "1")

	req, _ = http.NewRequest("POST", "/make_reservation", strings.NewReader(postedData.Encode()))
	ctx = getCtx(req)
	session.Put(ctx, "reservation", reservation)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("PostReservation handler returned %v for invalid form data, expected %v", rr.Code, http.StatusOK)
	}

	// test unable to insert reservation into db
	sdr := sd.AddDate(2070, 01, 01)
	edr := ed.AddDate(2070, 01, 02)
	res := models.Reservation{
		StartDate: sdr,
		EndDate:   edr,
		RoomID:    4,
	}

	postedData = url.Values{}
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@smith.com")
	postedData.Add("phone", "076859432")

	req, _ = http.NewRequest("POST", "/make_reservation", strings.NewReader(postedData.Encode()))
	ctx = getCtx(req)
	session.Put(ctx, "reservation", res)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned %v for failing to insert reservation, expected %v", rr.Code, http.StatusTemporaryRedirect)
	}

	// test unable to insert room restriction into db
	reservation.RoomID = 3

	postedData = url.Values{}
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@smith.com")
	postedData.Add("phone", "076859432")
	postedData.Add("room_id", "room_id=3")

	req, _ = http.NewRequest("POST", "/make_reservation", strings.NewReader(postedData.Encode()))
	ctx = getCtx(req)
	session.Put(ctx, "reservation", reservation)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned %v for failing to insert reservation, expected %v", rr.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepoAvailabilityJSON(t *testing.T) {
	// case: Rooms are not available
	var j jsonResponse
	reqBody := "start=2023-08-10"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2023-08-13")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")
	req, _ := http.NewRequest("POST", "/search_availablity-json", strings.NewReader(reqBody))
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.AvailabilityJSON)
	handler.ServeHTTP(rr, req)
	err := json.Unmarshal(rr.Body.Bytes(), &j)
	if err != nil {
		t.Error("failed to parse JSON")
	}
	if j.Ok != false {
		t.Errorf("AvailablityJSON handler returned %v for no availability, expected %v", j.Ok, false)
	}

	// case: Unable to parse form because it is missing
	req, _ = http.NewRequest("POST", "/search_availablity-json", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.AvailabilityJSON)
	handler.ServeHTTP(rr, req)
	err = json.Unmarshal(rr.Body.Bytes(), &j)
	if err != nil {
		t.Error("failed to parse JSON")
	}
	if j.Ok != false && j.Message != "Internal Server Error" {
		t.Errorf("AvailablityJSON handler returned %v and %v for non-existant form, expected %v and Internal Server Error", j.Ok, j.Message, false)
	}

	// case: Start date is in incorrect format
	reqBody = "start=10-08-2050"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2050-08-13")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")
	req, _ = http.NewRequest("POST", "/search_availablity-json", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.AvailabilityJSON)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("AvailabilityJSON handler returned %v for incorrect date format, expected %v", rr.Code, http.StatusTemporaryRedirect)
	}

	// case: End date is in incorrect format
	reqBody = "start=2050-08-10"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=13-08-2050")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")
	req, _ = http.NewRequest("POST", "/search_availablity-json", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.AvailabilityJSON)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("AvailabilityJSON handler returned %v for incorrect date format, expected %v", rr.Code, http.StatusTemporaryRedirect)
	}

	// case: Room ID cannot be converted to an int
	reqBody = "start=2050-08-10"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2050-08-13")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=invalid")
	req, _ = http.NewRequest("POST", "/search_availablity-json", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.AvailabilityJSON)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("AvailabilityJSON handler returned %v for incorrect date format, expected %v", rr.Code, http.StatusTemporaryRedirect)
	}

	// case: SearchAvailabilityByRoomID is not possible
	reqBody = "start=2050-08-10"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2050-08-13")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1000")
	req, _ = http.NewRequest("POST", "/search_availablity-json", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.AvailabilityJSON)
	handler.ServeHTTP(rr, req)
	err = json.Unmarshal(rr.Body.Bytes(), &j)
	if err != nil {
		t.Error("failed to parse JSON")
	}
	if j.Ok != false && j.Message != "Error querying database" {
		t.Errorf("AvailablityJSON handler returned %v and %v for bad db query, expected %v and Error querying database", j.Ok, j.Message, false)
	}
}

func TestRepoPostAvailability(t *testing.T) {
	// case: No rooms are available
	reqBody := "start=2060-08-10"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2060-08-13")
	req, _ := http.NewRequest("POST", "/search_availablity", strings.NewReader(reqBody))
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.PostAvailability)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostAvailablity handler returned %v for no availability, expected %v", rr.Code, http.StatusSeeOther)
	}

	// case: Cannot parse form
	req, _ = http.NewRequest("POST", "/search_availablity", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostAvailability)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostAvailablity handler returned %v for no availability, expected %v", rr.Code, http.StatusTemporaryRedirect)
	}
	// case: Start date is in incorrect format
	reqBody = "start=10-08-2050"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2050-08-13")
	req, _ = http.NewRequest("POST", "/search_availablity-json", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostAvailability)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostAvailability handler returned %v for incorrect date format, expected %v", rr.Code, http.StatusTemporaryRedirect)
	}

	// case: End date is in incorrect format
	reqBody = "start=2050-08-10"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=13-08-2050")
	req, _ = http.NewRequest("POST", "/search_availablity-json", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostAvailability)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostAvailability handler returned %v for incorrect date format, expected %v", rr.Code, http.StatusTemporaryRedirect)
	}

	// case: failed to get availability form database

	reqBody = "start=2012-08-10"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2050-08-10")
	req, _ = http.NewRequest("POST", "/search_availablity-json", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostAvailability)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostAvailability handler returned %v for unavailable db connection, expected %v", rr.Code, http.StatusTemporaryRedirect)
	}

	// case: Room is available

	reqBody = "start=2050-08-10"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2020-01-01")
	req, _ = http.NewRequest("POST", "/search_availablity-json", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostAvailability)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("PostAvailability handler returned %v for unavailable db connection, expected %v", rr.Code, http.StatusOK)
	}
}

func TestRepoSummary(t *testing.T) {
	// case: complete reservation data is in session
	var reservation models.Reservation
	var sd time.Time
	var ed time.Time
	sd = sd.AddDate(2060, 0o1, 0o1)
	ed = ed.AddDate(2060, 0o1, 0o2)
	reservation = models.Reservation{
		ID:        10,
		RoomID:    1,
		FirstName: "John",
		LastName:  "Sue",
		Email:     "john@sue.com",
		Phone:     "076654387",
		StartDate: sd,
		EndDate:   ed,
	}
	req, _ := http.NewRequest("GET", "/reservation_summary", nil)
	ctx := getCtx(req)
	session.Put(ctx, "reservation", reservation)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.Summary)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("ReservationSummary handler returned %v for correct call, expected %v", rr.Code, http.StatusOK)
	}

	// case: no reservation data in session
	req, _ = http.NewRequest("GET", "/reservation_summary", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.Summary)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ReservationSummary handler returned %v for missing reservation data, expected %v", rr.Code, http.StatusSeeOther)
	}
}

func TestRepoChooseRoom(t *testing.T) {
	// case: room id provided in incorrect format
	req, _ := http.NewRequest("GET", "/choose_room", nil)
	req.RequestURI = "/choose_room/invalid"
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom handler returned %v for invalid room ID, expected %v", rr.Code, http.StatusSeeOther)
	}

	// case: no reservation object in session

	req, _ = http.NewRequest("GET", "/choose_room", nil)
	req.RequestURI = "/choose_room/1"
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom handler returned %v for invalid room ID, expected %v", rr.Code, http.StatusSeeOther)
	}

	// case: correct function call
	var res models.Reservation
	req, _ = http.NewRequest("GET", "/choose_room", nil)
	req.RequestURI = "/choose_room/1"
	ctx = getCtx(req)
	session.Put(ctx, "reservation", res)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom handler returned %v for invalid room ID, expected %v", rr.Code, http.StatusSeeOther)
	}
}

func TestRepoBookRoom(t *testing.T) {
	// case : correct url call
	req, _ := http.NewRequest("GET", "/book_room?id=1&s=2050-01-01&e=2050-01-02", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.BookRoom)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("BookRoom handler returned %v for correct query, expected %v", rr.Code, http.StatusSeeOther)
	}
	// case: missing id from url
	req, _ = http.NewRequest("GET", "/book_room?s=2050-01-01&e=2050-01-02", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.BookRoom)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("BookRoom handler returned %v for correct query, expected %v", rr.Code, http.StatusSeeOther)
	}
	// case: bad database call
	req, _ = http.NewRequest("GET", "/book_room?id=10&s=2050-01-01&e=2050-01-02", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.BookRoom)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("BookRoom handler returned %v for correct query, expected %v", rr.Code, http.StatusSeeOther)
	}
}

var loginTests = []struct {
	name          string
	email         string
	expStatusCode int
	expHTML       string
	expLocation   string
}{
	{name: "valid_cred", email: "me@sosmart.com", expStatusCode: http.StatusSeeOther, expHTML: "", expLocation: "/"},
	{name: "invalid_cred", email: "me@nosmart.ro", expStatusCode: http.StatusSeeOther, expHTML: "", expLocation: "/user/login"},
	{name: "invalid_data", email: "nosmart", expStatusCode: http.StatusOK, expHTML: "action='/user/login'", expLocation: ""},
}

func TestRepoLogin(t *testing.T) {
	for _, e := range loginTests {
		postedData := url.Values{}
		postedData.Add("email", e.email)
		postedData.Add("password", "password")

		req, _ := http.NewRequest("POST", "/user/login", strings.NewReader(postedData.Encode()))
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.PostLogin)
		handler.ServeHTTP(rr, req)
		if rr.Code != e.expStatusCode {
			t.Errorf("Failed %v: returned %v, expected %v", e.name, rr.Code, e.expStatusCode)
		}
		if e.expLocation != "" {
			location, _ := rr.Result().Location()
			if location.String() != e.expLocation {
				t.Errorf("Failed %s: got location %s, expected location %s", e.name, location.String(), e.expLocation)
			}
		}
		if e.expHTML != "" {
			html := rr.Body.String()
			if !strings.Contains(html, e.expHTML) {
				t.Errorf("Failed %s: got  %s, expected %s", e.name, html, e.expHTML)
			}
		}
	}
}

func TestRepoAdminProcessReservation(t *testing.T) {
	// case: coming from a page that does not need time data aka new_res or all_res
	src := "new"
	id := 10
	req, _ := http.NewRequest("GET", fmt.Sprintf("/admin/process_reservation/%s/%d/do", src, id), nil)
	expLocation := fmt.Sprintf("/admin/reservations_%s", src)
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.AdminProcessReservation)
	handler.ServeHTTP(rr, req)
	location, _ := rr.Result().Location()
	if rr.Code != http.StatusSeeOther && expLocation != location.String() {
		t.Errorf("Failed ProcessReservation: returned %v and %s, expected %v and %s", rr.Code, location.String(), http.StatusSeeOther, expLocation)
	}

	// case: coming from calendar page
	src = "cal"
	id = 13
	year := "2023"
	month := "08"

	req, _ = http.NewRequest("GET", fmt.Sprintf("/admin/process_reservation/%s/%d/do?y=%s&m=%s", src, id, year, month), nil)
	expLocation = fmt.Sprintf("/admin/reservations_calendar?y=%s&m=%s", year, month)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.AdminProcessReservation)
	handler.ServeHTTP(rr, req)
	location, _ = rr.Result().Location()
	if rr.Code != http.StatusSeeOther && expLocation != location.String() {
		t.Errorf("Failed ProcessReservation: returned %v and %s, expected %v and %s", rr.Code, location.String(), http.StatusSeeOther, expLocation)
	}
}

func TestRepoAdminDeleteReservation(t *testing.T) {
	// case: coming from a page that does not need time data aka new_res or all_res
	src := "all"
	id := 10
	req, _ := http.NewRequest("GET", fmt.Sprintf("/admin/delete_reservation/%s/%d/do", src, id), nil)
	expLocation := fmt.Sprintf("/admin/reservations_%s", src)
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.AdminDeleteReservation)
	handler.ServeHTTP(rr, req)
	location, _ := rr.Result().Location()
	if rr.Code != http.StatusSeeOther && expLocation != location.String() {
		t.Errorf("Failed ProcessReservation: returned %v and %s, expected %v and %s", rr.Code, location.String(), http.StatusSeeOther, expLocation)
	}

	// case: coming from calendar page
	src = "cal"
	id = 13
	year := "2023"
	month := "08"

	req, _ = http.NewRequest("GET", fmt.Sprintf("/admin/delete_reservation/%s/%d/do?y=%s&m=%s", src, id, year, month), nil)
	expLocation = fmt.Sprintf("/admin/reservations_calendar?y=%s&m=%s", year, month)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.AdminDeleteReservation)
	handler.ServeHTTP(rr, req)
	location, _ = rr.Result().Location()
	if rr.Code != http.StatusSeeOther && expLocation != location.String() {
		t.Errorf("Failed ProcessReservation: returned %v and %s, expected %v and %s", rr.Code, location.String(), http.StatusSeeOther, expLocation)
	}
}

var adminPostReservationTests = []struct {
	name          string
	pageSrc       string
	resID         int
	postData      url.Values
	expStatusCode int
	expLocation   string
}{
	{
		name:    "post_from_all_res",
		pageSrc: "all",
		resID:   13,
		postData: url.Values{
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@example.com"},
			"phone":      {"555-555-5555"},
		},
		expStatusCode: http.StatusSeeOther,
		expLocation:   fmt.Sprintf("/admin/reservations_%s", "all"),
	},
	{
		name:    "post_from_cal",
		pageSrc: "cal",
		resID:   13,
		postData: url.Values{
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@example.com"},
			"phone":      {"555-555-5555"},
			"year":       {"2023"},
			"month":      {"08"},
		},
		expStatusCode: http.StatusSeeOther,
		expLocation:   fmt.Sprintf("/admin/reservations_calendar?y=%s&m=%s", "2023", "08"),
	},
}

func TestRepoAdminPostReservation(t *testing.T) {
	for _, e := range adminPostReservationTests {
		req, _ := http.NewRequest("POST", fmt.Sprintf("/admin/reservations/%s/%d", e.pageSrc, e.resID), strings.NewReader(e.postData.Encode()))
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		location, _ := rr.Result().Location()
		handler := http.HandlerFunc(Repo.AdminPostReservation)
		handler.ServeHTTP(rr, req)
		if rr.Code != e.expStatusCode && e.expLocation != location.String() {
			t.Errorf("Failed %s: returned %v and %s, expected %v and %s", e.name, rr.Code, location.String(), e.expStatusCode, e.expLocation)
		}
	}
}

var adminPostCalendarTests = []struct {
	name          string
	postedData    url.Values
	expStatusCode int
	blocks        int
	reservations  int
}{
	{
		name: "cal",
		postedData: url.Values{
			"year":  {time.Now().Format("2006")},
			"month": {time.Now().Format("01")},
			fmt.Sprintf("add_block_1_%s", time.Now().AddDate(0, 0, 1).Format("2006-01-2")): {"1"},
		},
		expStatusCode: http.StatusSeeOther,
	},
	{
		name:          "cal_block",
		postedData:    url.Values{},
		expStatusCode: http.StatusSeeOther,
		blocks:        1,
	},
	{
		name:          "cal_res",
		postedData:    url.Values{},
		expStatusCode: http.StatusSeeOther,
		reservations:  1,
	},
}

func TestRepoAdminPostCalendar(t *testing.T) {
	for _, e := range adminPostCalendarTests {
		var req *http.Request
		if e.postedData != nil {
			req, _ = http.NewRequest("POST", "/admin/reservations_calendar", strings.NewReader(e.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/admin/reservations_calendar", nil)
		}
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		now := time.Now()
		bm := make(map[string]int)
		rm := make(map[string]int)
		currentYear, currentMonth, _ := now.Date()
		currentLocation := now.Location()

		firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
		lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {
			rm[d.Format("2006-01-2")] = 0
			bm[d.Format("2006-01-2")] = 0
		}

		if e.blocks > 0 {
			bm[firstOfMonth.Format("2006-01-2")] = e.blocks
		}

		if e.reservations > 0 {
			rm[lastOfMonth.Format("2006-01-2")] = e.reservations
		}

		session.Put(ctx, "block_map_1", bm)
		session.Put(ctx, "reservation_map_1", rm)

		// set the header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// call the handler
		handler := http.HandlerFunc(Repo.AdminPostReservationsCalendar)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expStatusCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expStatusCode, rr.Code)
		}

	}
}

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}
