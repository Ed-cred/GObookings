package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Ed-cred/bookings/internal/config"
	"github.com/Ed-cred/bookings/internal/driver"
	"github.com/Ed-cred/bookings/internal/forms"
	"github.com/Ed-cred/bookings/internal/models"
	"github.com/Ed-cred/bookings/internal/render"
	"github.com/Ed-cred/bookings/internal/repository"
	"github.com/Ed-cred/bookings/internal/repository/dbrepo"
)

// Repository used by the handlers
var Repo *Repository

// Repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DbRepo
}

// Creates a new repository
func NewRepository(app *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: app,
		DB:  dbrepo.NewPostgresRepo(db.SQL, app),
	}
}

func NewTestRepository(app *config.AppConfig) *Repository {
	return &Repository{
		App: app,
		DB:  dbrepo.NewTestRepo(app),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

func (rep *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, "home.page.tmpl", r, &models.TemplateData{})
}

func (rep *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, "generals.page.tmpl", r, &models.TemplateData{})
}

func (rep *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, "majors.page.tmpl", r, &models.TemplateData{})
}

func (rep *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, "search_availability.page.tmpl", r, &models.TemplateData{})
}

// PostAvailability handler for post method
func (rep *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "can't parse form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	// 2020-01-01 -- 01/02 03:04:05PM '06 -0700

	layout := "2006-01-02"

	startDate, err := time.Parse(layout, sd)
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "can't parse start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	endDate, err := time.Parse(layout, ed)
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "can't get parse end date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	rooms, err := rep.DB.SearchAvailabilityAllRooms(startDate, endDate)
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "can't query database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if len(rooms) == 0 {
		rep.App.Session.Put(r.Context(), "error", "No availability")
		http.Redirect(w, r, "/search_availability", http.StatusSeeOther)
		return
	}
	data := make(map[string]interface{})
	data["rooms"] = rooms
	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}
	rep.App.Session.Put(r.Context(), "reservation", res)
	render.Template(w, "choose_room.page.tmpl", r, &models.TemplateData{
		Data: data,
	})
}

type jsonResponse struct {
	Ok        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomId    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// AvailabilityJSON handler for on page post request and sends JSON response
func (rep *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		resp := jsonResponse{
			Ok:      false,
			Message: "Internal server error",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, sd)
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "can't parse start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	endDate, err := time.Parse(layout, ed)
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "can't parse end date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "type conversion failed for string to integer")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	available, err := rep.DB.SearchAvailabilityByRoomID(startDate, endDate, roomID)
	if err != nil {

		resp := jsonResponse{
			Ok:      false,
			Message: "",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}
	resp := jsonResponse{
		Ok:        available,
		Message:   "",
		StartDate: sd,
		EndDate:   ed,
		RoomId:    strconv.Itoa(roomID),
	}
	out, _ := json.MarshalIndent(resp, "", "     ")

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (rep *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, "contact.page.tmpl", r, &models.TemplateData{})
}

func (rep *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	res, ok := rep.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		rep.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	room, err := rep.DB.GetRoomById(res.RoomID)
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "can't find room")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	res.Room.RoomName = room.RoomName
	rep.App.Session.Put(r.Context(), "reservation", res)
	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed
	data := make(map[string]interface{})
	data["reservation"] = res
	render.Template(w, "make_reservation.page.tmpl", r, &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

// PostReservation handles the posting of a reservation form
func (rep *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	reservation, ok := rep.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		rep.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	err := r.ParseForm()
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "can't parse form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Phone = r.Form.Get("phone")
	reservation.Email = r.Form.Get("email")

	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")
	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		http.Error(w, "Invalid form data", http.StatusSeeOther)
		render.Template(w, "make_reservation.page.tmpl", r, &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}
	newReservationID, err := rep.DB.InsertReservation(reservation)
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "can't insert reservation into database!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: newReservationID,
		RestricitonID: 1,
	}
	err = rep.DB.InsertRoomRestriction(restriction)
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "can't insert room restriction!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	// send notification

	htmlMessage := fmt.Sprintf(`
		<strong>Reservation confirmation</strong><br>
		Dear %s, <br>
		This is a confirmation for your reservation from %s to %s for the %s room.
	`, reservation.FirstName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"), reservation.Room.RoomName)

	msg := models.MailData{
		To:       reservation.Email,
		From:     "me@here.com",
		Subject:  "Reservation confirmation",
		Content:  htmlMessage,
		Template: "basic.html",
	}
	rep.App.MailChan <- msg

	htmlMessage = fmt.Sprintf(`
		<strong>New Reservation</strong><br>
		A reservation has been made from %s to %s for the %s room.
	`, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"), reservation.Room.RoomName)

	msg = models.MailData{
		To:      "property@owner.com",
		From:    "me@here.com",
		Subject: "New Reservation",
		Content: htmlMessage,
	}
	rep.App.MailChan <- msg
	rep.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/reservation_summary", http.StatusSeeOther)
}

func (rep *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.Template(w, "about.page.tmpl", r, &models.TemplateData{})
}

func (rep *Repository) Summary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := rep.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		rep.App.Session.Put(r.Context(), "error", "Unable to get reservation from session")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	rep.App.Session.Remove(r.Context(), "reservation")
	layout := "2006-01-02"
	sd := reservation.StartDate.Format(layout)
	ed := reservation.EndDate.Format(layout)
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed
	data := make(map[string]interface{})
	data["reservation"] = reservation
	render.Template(w, "reservation_summary.page.tmpl", r, &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

func (rep *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")
	roomId, err := strconv.Atoi(exploded[2])
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "Unable to get room ID from URL")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	res, ok := rep.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		rep.App.Session.Put(r.Context(), "error", "Unable to get reservation from session")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	res.RoomID = roomId
	rep.App.Session.Put(r.Context(), "reservation", res)
	http.Redirect(w, r, "/make_reservation", http.StatusSeeOther)
}

// Takes URL params, builds session var and redirects to make_resevation page
func (rep *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	roomId, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "Unable to get parameter from URL")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	sd := r.URL.Query().Get("s")
	ed := r.URL.Query().Get("e")
	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)
	room, err := rep.DB.GetRoomById(roomId)
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "Unable to get room from database")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	var res models.Reservation
	res.RoomID = roomId
	res.StartDate = startDate
	res.EndDate = endDate
	res.Room.RoomName = room.RoomName
	rep.App.Session.Put(r.Context(), "reservation", res)
	http.Redirect(w, r, "make_reservation", http.StatusSeeOther)
}

func (rep *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Template(w, "login.page.tmpl", r, &models.TemplateData{
		Form: forms.New(nil),
	})
}

func (rep *Repository) PostLogin(w http.ResponseWriter, r *http.Request) {
	err := rep.App.Session.RenewToken(r.Context())
	if err != nil {
		panic(err)
	}
	err = r.ParseForm()
	if err != nil {
		log.Println("parsing login form error", err)
	}
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")
	if !form.Valid() {
		render.Template(w, "login.page.tmpl", r, &models.TemplateData{
			Form: form, 
		})
		return 
	}
	id, _, err := rep.DB.Authenticate(email, password)
	if err != nil {
		log.Println(err)
		rep.App.Session.Put(r.Context(), "error", "invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}
	rep.App.Session.Put(r.Context(), "user_id", id)
	rep.App.Session.Put(r.Context(), "flash", "Successfully logged in")
	http.Redirect(w, r, "/", http.StatusSeeOther)

	log.Println("works!")
}


func (rep *Repository) UserLogout(w http.ResponseWriter, r *http.Request) {
	rep.App.Session.Destroy(r.Context())
	rep.App.Session.RenewToken(r.Context())
	http.Redirect(w, r, "/", http.StatusSeeOther)
}


func (rep *Repository) AdminDashboard (w http.ResponseWriter, r *http.Request) {
	render.Template(w, "admin_dashboard.page.tmpl", r, &models.TemplateData{})
}