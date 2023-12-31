package handlers

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Ed-cred/bookings/internal/config"
	"github.com/Ed-cred/bookings/internal/models"
	"github.com/Ed-cred/bookings/internal/render"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/justinas/nosurf"
)

var (
	app            config.AppConfig
	session        *scs.SessionManager
	pathToTemplate = "./../../templates"
)


var functions =template.FuncMap{
	"humanDate": render.HumanDate,
	"formatDate": render.FormatDate,
	"iterate": render.IterateDays,
}

func TestMain(m *testing.M) {
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})

	// change to true when in produciton
	app.InProd = false
	infoLog := log.New(os.Stdout, "Info\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog := log.New(os.Stdout, "Error\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProd
	app.Session = session

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan
	defer close(mailChan)
	listenForMail()


	tc, err := CreateTestTemplateCache()
	if err != nil {
		log.Fatal("Could not create template cache")
	}

	app.TemplateCache = tc
	app.UseCache = true
	repo := NewTestRepository(&app)
	NewHandlers(repo) 
	render.NewRenderer(&app)
	os.Exit(m.Run())	
}

func getRoutes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	//	mux.Use(NoSurf)
	mux.Use(SessionLoad)
	mux.Get("/", Repo.Home)
	mux.Get("/about", Repo.About)
	mux.Get("/generals-quarters", Repo.Generals)
	mux.Get("/majors-suite", Repo.Majors)

	mux.Get("/search_availability", Repo.Availability)
	mux.Post("/search_availability", Repo.PostAvailability)
	mux.Post("/search_availability-json", Repo.AvailabilityJSON)

	mux.Get("/contact", Repo.Contact)

	mux.Get("/make_reservation", Repo.Reservation)
	mux.Post("/make_reservation", Repo.PostReservation)

	mux.Get("/reservation_summary", Repo.Summary)

	
	mux.Get("/user/login", Repo.ShowLogin)
	mux.Post("/user/login", Repo.PostLogin)
	mux.Get("/user/logout", Repo.UserLogout)

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	mux.Route("/admin", func(mux chi.Router) {
		mux.Get("/dashboard", Repo.AdminDashboard)
		mux.Get("/reservations_new", Repo.AdminNewReservations)
		mux.Get("/reservations_all", Repo.AdminAllReservations)
		
		mux.Get("/reservations_calendar", Repo.AdminReservationsCalendar)
		mux.Post("/reservations_calendar", Repo.AdminPostReservationsCalendar)

		mux.Get("/process_reservation/{src}/{id}/do", Repo.AdminProcessReservation)
		mux.Get("/delete_reservation/{src}/{id}/do", Repo.AdminDeleteReservation)

		mux.Get("/reservations/{src}/{id}/show", Repo.AdminShowReservation)
		mux.Post("/reservations/{src}/{id}", Repo.AdminPostReservation)
	})

	return mux
}

func listenForMail () {
	go func () {
		for {
			_ = <- app.MailChan
		}
	}()
}

// Nosurf adds CSRF protection to all POST requests
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProd,
		SameSite: http.SameSiteLaxMode,
	})
	return csrfHandler
}

// SessionLoad saves and loads the session on request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

func CreateTestTemplateCache() (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)
	// get files named .page.tmpl

	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplate))
	if err != nil {
		return cache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return cache, err
		}
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplate))
		if err != nil {
			return cache, err
		}
		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplate))
			if err != nil {
				return cache, err
			}
		}
		cache[name] = ts

	}
	return cache, nil
}
