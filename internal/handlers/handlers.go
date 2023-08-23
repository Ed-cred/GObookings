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
	"github.com/Ed-cred/bookings/internal/helpers"
	"github.com/Ed-cred/bookings/internal/models"
	"github.com/Ed-cred/bookings/internal/render"
	"github.com/Ed-cred/bookings/internal/repository"
	"github.com/Ed-cred/bookings/internal/repository/dbrepo"
	"github.com/go-chi/chi"
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
		RestrictionID: 1,
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

func (rep *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, "admin_dashboard.page.tmpl", r, &models.TemplateData{})
}

func (rep *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	newReservations, err := rep.DB.AllNewReservations()
	if err != nil {
		helpers.ServerError(w, err)
		rep.App.Session.Put(r.Context(), "error", "could not fetch reservations from database")
		// http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}
	data := make(map[string]interface{})
	data["reservations"] = newReservations

	render.Template(w, "admin_new_reservations.page.tmpl", r, &models.TemplateData{
		Data: data,
	})
}

func (rep *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := rep.DB.AllReservations()
	if err != nil {
		helpers.ServerError(w, err)
		rep.App.Session.Put(r.Context(), "error", "could not fetch reservations from database")
		// http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}
	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, "admin_all_reservations.page.tmpl", r, &models.TemplateData{
		Data: data,
	})
}

func (rep *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	//assume there is no month or year specified

	now := time.Now()
	if r.URL.Query().Get("y") != "" {
		year, err := strconv.Atoi(r.URL.Query().Get("y"))
		if err != nil {
			helpers.ServerError(w, err)
			rep.App.Session.Put(r.Context(), "error", "could not convert type string to int")
			return
		}
		month, err := strconv.Atoi(r.URL.Query().Get("m"))
		if err != nil {
			helpers.ServerError(w, err)
			rep.App.Session.Put(r.Context(), "error", "could not convert type string to int")
			return
		}
		now = time.Date(year, time.Month(month),1, 0, 0, 0, 0, time.UTC)
	}

	data := make(map[string]interface{})
	data["now"] = now


	next := now.AddDate(0, 1, 0)
	last := now.AddDate(0, -1, 0)
	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")
	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	stringMap := make(map[string]string)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear
	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	rooms, err := rep.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		rep.App.Session.Put(r.Context(), "error", "could not fetch room data from database")
		return
	}
	data["rooms"] = rooms
	for _, x := range rooms {
		// create maps to store reservation data
		resMap := make(map[string]int)
		blockMap := make(map[string]int)
		for d := firstOfMonth; !d.After(lastOfMonth); d = d.AddDate(0, 0, 1) {
			resMap[d.Format("2006-01-02")] = 0
			blockMap[d.Format("2006-01-02")] = 0
		} 
		// get all restrictions for the current room
		restrictions, err := rep.DB.FetchRestrictionsForRoomByDay(x.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			helpers.ServerError(w, err)
			rep.App.Session.Put(r.Context(), "error", "could not fetch room restrictions from database")
			return
		}
		for _, y := range restrictions {
			if y.ReservationID != 0 {
				//it's a reservation yay!!
				for d := y.StartDate; !d.After(y.EndDate); d = d.AddDate(0, 0, 1) {
					resMap[d.Format("2006-01-02")] = y.ReservationID
				}
			} else {
				//admin block
				blockMap[y.StartDate.Format("2006-01-02")] = y.ID
			}
		}
		data[fmt.Sprintf("reservation_map_%d", x.ID)] = resMap
		data[fmt.Sprintf("block_map_%d", x.ID)] = blockMap

		rep.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", x.ID), blockMap )

	}
	render.Template(w, "admin_reservations_calendar.page.tmpl", r, &models.TemplateData{
		StringMap: stringMap,
		Data: data,
		IntMap: intMap,
	})
}


func (rep *Repository) AdminShowReservation (w http.ResponseWriter, r *http.Request) {
	exp := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exp[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	log.Println(id)
	src := exp[3]
	stringMap := make(map[string]string)
	stringMap["src"] = src
	//get reservation from the database
	res, err := rep.DB.FetchReservationById(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	data := make(map[string]interface{})
	data["reservation"] = res
	render.Template(w, "admin_reservations_show.page.tmpl", r, &models.TemplateData{
		Data: data,
		StringMap: stringMap,
		Form: forms.New(nil),
	})

}


func (rep *Repository) AdminPostReservation (w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	exp := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exp[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	log.Println(id)
	src := exp[3]
	stringMap := make(map[string]string)
	stringMap["src"] = src
	res, err := rep.DB.FetchReservationById(id)	
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	err = rep.DB.UpdateReservation(res)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	rep.App.Session.Put(r.Context(), "flash", "Changes saved!")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations_%s", src), http.StatusSeeOther)
}



func (rep *Repository) AdminProcessReservation (w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")
	err := rep.DB.UpdateProcessedReservation(id, 1)	
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	rep.App.Session.Put(r.Context(), "flash", "Processed reservation!")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations_%s", src), http.StatusSeeOther)
}

func (rep *Repository) AdminDeleteReservation (w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")
	err := rep.DB.DeleteReservation(id)	
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	rep.App.Session.Put(r.Context(), "flash", "Reservation has been deleted!")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations_%s", src), http.StatusSeeOther)
}

func (rep *Repository) AdminPostReservationsCalendar (w http.ResponseWriter, r *http.Request) {
	err  := r.ParseForm()
	if err != nil {
			helpers.ServerError(w, err)
			return	
	}
	year, _ := strconv.Atoi(r.Form.Get("y"))
	month, _ := strconv.Atoi(r.Form.Get("m"))

	//process calendar blocks
	rooms, err := rep.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	form := forms.New(r.PostForm)
	for _, x := range rooms {
		//If we have an entry in the map that does not exist in the posted data
		//And restricton_id > 0, then we know it is a block we need to remove
		currMap := rep.App.Session.Get(r.Context(), fmt.Sprintf("block_map_%d", x.ID)).(map[string]int)
		for name, value := range currMap {
			if val, ok := currMap[name]; ok {
				// if ok is indeed true, we only need to look at values > 0
				// that are not in the form post data
				if val > 0{
					if !form.Has(fmt.Sprintf("remove_block_%d_%s", x.ID, name)) {
					log.Println("Identified unchecked block that would be removed: ", value)
					}
				}

			}
		}
	}
	//Handle newly checked blocks
	for name := range r.PostForm {
		if strings.HasPrefix(name,"add_block"){
			exp := strings.Split(name, "_")
			roomId, err := strconv.Atoi(exp[2])
			if err != nil {
				helpers.ServerError(w, err)
				return
			}
			log.Println("would insert block for room id:", roomId, "for date", exp[3])
		}
		log.Println("form has name: ", name)
	}


	rep.App.Session.Put(r.Context(), "flash", "Changes saved!")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations_calendar?y=%d&m=%d", year, month), http.StatusSeeOther)
}