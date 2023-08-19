package dbrepo

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Ed-cred/bookings/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (m *postgresDbRepo) AllUsers() bool {
	return true
}

func (m *postgresDbRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var newID int

	stmt := `insert into reservations (first_name, last_name, email,  phone, start_date, end_date, room_id, created_at, updated_at) 
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			returning id`
	err := m.DB.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
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
		log.Println("Unable to insert data into room_restrictions table: ", err)
		return err
	}
	return nil
}

// Returns true if the date range is available for specified roomID,otherwise false
func (m *postgresDbRepo) SearchAvailabilityByRoomID(start, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var numRows int
	query := `select count(id) from room_restrictions 
			where room_id = $1 and $2 < end_date and $3 > start_date`
	row := m.DB.QueryRowContext(ctx, query, roomID, start, end)
	err := row.Scan(&numRows)
	if err != nil {
		return false, err
	}
	if numRows == 0 {
		return true, nil
	}
	return false, nil
}

func (m *postgresDbRepo) SearchAvailabilityAllRooms(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var rooms []models.Room
	query := `select r.id, r.room_name from rooms r 
			where r.id not in (select room_id from room_restrictions rr
			where $1 < rr.end_date and $2 > rr.start_date);`
	rows, err := m.DB.QueryContext(ctx, query, start, end)
	if err != nil {
		return rooms, err
	}
	for rows.Next() {
		var room models.Room
		err := rows.Scan(&room.ID, &room.RoomName)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, room)
	}
	if err := rows.Err(); err != nil {
		return rooms, err
	}
	return rooms, nil
}

func (m *postgresDbRepo) GetRoomById(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	var room models.Room
	query := `SELECT id, room_name, created_at, updated_at FROM rooms where id=$1`
	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(&room.ID, &room.RoomName, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return room, err
	}

	return room, nil
}

//Returns a models.User object containing the information from the database
func (m *postgresDbRepo) GetUserById (id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	query := `SELECT first_name, last_name, email, password, access_level, created_at, updated_at
	FROM users where id = $1`
	row := m.DB.QueryRowContext(ctx, query, id)
	var u models.User
	err := row.Scan(&u.FirstName, &u.LastName, &u.Email, &u.Password, &u.AccessLevel, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return u, err
	}
	return u, nil
}

func (m *postgresDbRepo) UpdateUser (u models.User) (error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	query := `UPDATE users SET first_name=$1, last_name=$2, email=$3, access_level=$4, updated_at=$5 
	WHERE id = $6`
	_, err := m.DB.ExecContext(ctx, query, 
		u.FirstName, 
		u.LastName, 
		u.Email, 
		u.AccessLevel, 
		time.Now(),
		u.ID,
	)
	if err != nil {
		return err
	}
	
	return nil 
}

// 
func (m *postgresDbRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	var id int 
	var hashPass string
	row := m.DB.QueryRowContext(ctx, "SELECT id, password FROM users WHERE email=$1", email)
	err := row.Scan(&id, &hashPass)
	if err != nil {
		return 0, "", err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashPass),[]byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword{
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}
	return id, hashPass, nil
}

func (m *postgresDbRepo) AllReservations () ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	var reservations []models.Reservation

	query := `SELECT r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date, r.room_id, r.created_at, r.updated_at,
	r.processed, rm.id, rm.room_name FROM reservations r
	LEFT JOIN rooms rm ON (r.room_id = rm.id) ORDER BY r.start_date DESC`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()
	for rows.Next() {
		var res models.Reservation
		err := rows.Scan(
			&res.ID,
			&res.FirstName,
			&res.LastName,
			&res.Email,
			&res.Phone,
			&res.StartDate,
			&res.EndDate,
			&res.RoomID,
			&res.CreatedAt,
			&res.UpdatedAt,
			&res.Processed,
			&res.Room.ID,
			&res.Room.RoomName,
		)
		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, res)
	}
	if err = rows.Err(); err != nil {
		return reservations, err 
	}

	return reservations, nil

}

func (m *postgresDbRepo) AllNewReservations () ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	var reservations []models.Reservation

	query := `SELECT r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date, r.room_id, r.created_at, r.updated_at,
	rm.id, rm.room_name FROM reservations r
	LEFT JOIN rooms rm ON (r.room_id = rm.id)
	WHERE processed = 0
	ORDER BY r.start_date DESC`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()
	for rows.Next() {
		var res models.Reservation
		err := rows.Scan(
			&res.ID,
			&res.FirstName,
			&res.LastName,
			&res.Email,
			&res.Phone,
			&res.StartDate,
			&res.EndDate,
			&res.RoomID,
			&res.CreatedAt,
			&res.UpdatedAt,
			&res.Room.ID,
			&res.Room.RoomName,
		)
		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, res)
	}
	if err = rows.Err(); err != nil {
		return reservations, err 
	}

	return reservations, nil

}