package main

import (
	"net/http"

	"github.com/Ed-cred/bookings/internal/config"
	"github.com/Ed-cred/bookings/internal/handlers"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)
	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/generals-quarters", handlers.Repo.Generals)
	mux.Get("/majors-suite", handlers.Repo.Majors)
	
	mux.Get("/search_availability", handlers.Repo.Availability)
	mux.Post("/search_availability", handlers.Repo.PostAvailability)
	mux.Post("/search_availability-json", handlers.Repo.AvailabilityJSON)
	mux.Get("/choose_room/{id}", handlers.Repo.ChooseRoom)
	mux.Get("/book_room", handlers.Repo.BookRoom)
	
	mux.Get("/contact", handlers.Repo.Contact)
	
	mux.Get("/make_reservation", handlers.Repo.Reservation)
	mux.Post("/make_reservation", handlers.Repo.PostReservation)
	
	mux.Get("/reservation_summary", handlers.Repo.Summary)

	mux.Get("/user/login", handlers.Repo.ShowLogin)
	mux.Post("/user/login", handlers.Repo.PostLogin)


	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
