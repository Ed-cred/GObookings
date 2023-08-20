package dbrepo

import (
	"errors"
	"time"

	"github.com/Ed-cred/bookings/internal/models"
)

func (m *testDBRepo) AllUsers() bool {
	return true
}

func (m *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	if res.RoomID == 4 {
		return 0, errors.New("failed to insert reservation into database")
	}
	return 1, nil
}

func (m *testDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	if r.RoomID == 3 {
		return errors.New("failed to insert room restriction into database")
	}
	return nil
}

// Returns true if the date range is available for specified roomID,otherwise false
func (m *testDBRepo) SearchAvailabilityByRoomID(start, end time.Time, roomID int) (bool, error) {
	if roomID == 1000{
		return false, errors.New("failed to search availability")
	}
	return false, nil
}

func (m *testDBRepo) SearchAvailabilityAllRooms(start, end time.Time) ([]models.Room, error) {
	var rooms []models.Room
	today := time.Now()
	avail := time.Date(2020, time.January, 01, 0, 0, 0, 0, time.UTC)
	if today.After(start)  {
		return rooms, errors.New("failed to search availability")
	}
	if end  == avail {
		rooms = []models.Room {
			{
				RoomName: "General's Quarters",
			},
		}
		return rooms, nil

	}
	
	return rooms, nil
}

func (m *testDBRepo) GetRoomById(id int) (models.Room, error) {
	var room models.Room
	if id > 2 {
		return room, errors.New("some error")
	}
	return room, nil
}

func (m *testDBRepo) GetUserById (id int) (models.User, error) {
	var u models.User
	return u, nil
}

func (m *testDBRepo) UpdateUser (u models.User) (error) {
	return nil
}

func (m *testDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	return 0, "", nil
}

func (m *testDBRepo) AllReservations () ([]models.Reservation, error) {
	return []models.Reservation{}, nil
}

func (m *testDBRepo) AllNewReservations () ([]models.Reservation, error) {
	return []models.Reservation{}, nil
}

func (m *testDBRepo) FetchReservationById(id int) (models.Reservation, error) {
	return models.Reservation{}, nil
}