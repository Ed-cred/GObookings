package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Ed-cred/bookings/internal/config"
	"github.com/Ed-cred/bookings/internal/handlers"
	"github.com/Ed-cred/bookings/internal/helpers"
	"github.com/Ed-cred/bookings/internal/models"
	"github.com/Ed-cred/bookings/internal/render"
	"github.com/Ed-cred/bookings/internal/driver"
	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8080"

var (
	app      config.AppConfig
	session  *scs.SessionManager
	infoLog  *log.Logger
	errorLog *log.Logger
)

func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to the database")
	defer db.SQL.Close()
	fmt.Printf("Starting up app on port %v\n", portNumber)
	// _ = http.ListenAndServe(portNumber, nil)
	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func run() (*driver.DB, error) {
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})

	// change to true when in produciton
	app.InProd = false

	infoLog = log.New(os.Stdout, "Info\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "Error\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProd
	app.Session = session
	// Connect to database
	log.Println("Connecting to database...")
	db, err := driver.ConnectSql("host=localhost port=5432 dbname=bookings user=postgres password=rootuser")
	if err != nil {
		log.Fatal("Cannot connect to database: ", err)
		return nil, err
	}
	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Could not create template cache")
		return nil, err
	}
	app.TemplateCache = tc
	app.UseCache = false
	render.NewRenderer(&app)

	repo := handlers.NewRepository(&app, db)
	handlers.NewHandlers(repo)

	helpers.NewHelpers(&app)
	return db, nil
}
