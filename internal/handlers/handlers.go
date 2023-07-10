package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Ed-cred/bookings/internal/config"
	"github.com/Ed-cred/bookings/internal/forms"
	"github.com/Ed-cred/bookings/internal/models"
	"github.com/Ed-cred/bookings/internal/render"
)

//TemplateData holds data sent from handlers to templates

// Repository used by the handlers
var Repo *Repository

// Repository type
type Repository struct {
	App *config.AppConfig
}

// Creates a new repository
func NewRepository(app *config.AppConfig) *Repository {
	return &Repository{
		App: app,
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

func (rep *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIP := r.RemoteAddr
	rep.App.Session.Put(r.Context(), "remote_ip", remoteIP)

	render.RenderTemplate(w, "home.page.tmpl", r, &models.TemplateData{})

}

func (rep *Repository) Generals(w http.ResponseWriter, r *http.Request) {

	render.RenderTemplate(w, "generals.page.tmpl", r, &models.TemplateData{})

}
func (rep *Repository) Majors(w http.ResponseWriter, r *http.Request) {

	render.RenderTemplate(w, "majors.page.tmpl", r, &models.TemplateData{})

}
func (rep *Repository) Availability(w http.ResponseWriter, r *http.Request) {

	render.RenderTemplate(w, "search_availability.page.tmpl", r, &models.TemplateData{})

}

// PostAvailability handler for post method
func (rep *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start") //start being the name of the input in HTML form
	end := r.Form.Get("end")
	w.Write([]byte(fmt.Sprintf("Start date is %s and end date is %s", start, end)))

}

type jsonResponse struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

// AvailabilityJSON handler for on page post request and sends JSON response
func (rep *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	resp := jsonResponse{
		Ok:      true,
		Message: "Available",
	}

	out, err := json.MarshalIndent(resp, "", " 	")
	if err != nil {
		log.Fatal("Json failed to build response: ", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)

}

func (rep *Repository) Contact(w http.ResponseWriter, r *http.Request) {

	render.RenderTemplate(w, "contact.page.tmpl", r, &models.TemplateData{})

}
func (rep *Repository) Reservation(w http.ResponseWriter, r *http.Request) {

	render.RenderTemplate(w, "make_reservation.page.tmpl", r, &models.TemplateData{
		Form: forms.New(nil),
	})

}

// PostReservation handles the posting of a reservation form
func (rep *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}
	reservation := models.Reservation{
		FirstName: r.Form.Get("first-name"),
		LastName: r.Form.Get("last-name"),
		Email: r.Form.Get("email"),
		Phone: r.Form.Get("phone"),
	}
	form := forms.New(r.PostForm)
	form.Has("first-name", r)
	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		render.RenderTemplate(w, "make_reservation.page.tmpl", r, &models.TemplateData{
			Form: form,
			Data: data,
		})
	
	}

}

func (rep *Repository) About(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["test"] = "Hello again"
	remoteIP := rep.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP
	render.RenderTemplate(w, "about.page.tmpl", r, &models.TemplateData{
		StringMap: stringMap,
	})

}
