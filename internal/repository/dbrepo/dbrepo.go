package dbrepo

import (
	"database/sql"

	"github.com/Ed-cred/bookings/internal/config"
	"github.com/Ed-cred/bookings/internal/repository"
)

type postgresDbRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

type testDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

// NewPostgresDbRepo returns a struct containing the app config and database connections
func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) repository.DbRepo {
	return &postgresDbRepo{
		App: a,
		DB:  conn,
	}
}

func NewTestRepo(a *config.AppConfig) repository.DbRepo {
	return &testDBRepo{
		App: a,
	}	
}
