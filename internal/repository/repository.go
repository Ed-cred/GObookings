package repository

import (
	"time"

	"github.com/Ed-cred/bookings/internal/models"
)

type DbRepo interface {
	AllUsers() bool
	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestriction (r models.RoomRestriction) error
	SearchAvailabilityByRoomID(start, end time.Time, roomID int) (bool, error)
	SearchAvailabilityAllRooms(start, end time.Time) ([]models.Room, error)
	GetRoomById (id int) (models.Room, error)
	GetUserById (id int) (models.User, error)
	UpdateUser (u models.User) (error)
	Authenticate(email, testPassword string) (int, string, error)
	AllReservations () ([]models.Reservation, error)
	AllNewReservations () ([]models.Reservation, error)
	FetchReservationById(id int) (models.Reservation, error)
	UpdateReservation (r models.Reservation) (error)
	DeleteReservation (id int) error
	UpdateProcessedReservation (id, processed int) error
}
