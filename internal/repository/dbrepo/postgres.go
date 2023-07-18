package dbrepo

import (
	"context"
	"log"
	"time"

	"github.com/Ed-cred/bookings/internal/models"
)

func (m *postgresDbRepo) AllUsers() bool {
	return true
}

func (m *postgresDbRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var newID int

	stmt := `insert into reservations (email, first_name, last_name, phone, start_date, end_date, room_id, created_at, updated_at) 
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			returning id`
	err := m.DB.QueryRowContext(ctx, stmt,
		res.Email,
		res.FirstName,
		res.LastName,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now(),
	).Scan(&newID)
	if err != nil {
		log.Printf("Error inserting reservation data into database: %v", err)
		return 0, err
	}
	return newID, nil
}

func (m *postgresDbRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	stmt := `insert into room_restrictions (start_date, end_date, room_id, reservation_id, restriction_id, created_at, updated_at)
			values ($1, $2, $3, $4, $5, $6, $7)`

	_, err := m.DB.ExecContext(ctx, stmt,
		r.StartDate,
		r.EndDate,
		r.RoomID,
		r.ReservationID,
		r.RestricitonID,
		time.Now(),
		time.Now(),	
	)
	if err != nil {
		return err
	}
	return nil
}
