package repository

import "github.com/Ed-cred/bookings/internal/models"

type DbRepo interface {
	AllUsers() bool
	InsertReservation(res models.Reservation) error
}
