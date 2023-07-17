package models

import "time"

// User model
type User struct {
	ID          int
	AccessLevel int
	FirstName   string
	LastName    string
	Email       string
	Password    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Room model
type Room struct {
	ID        int
	RoomName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Restriction struct {
	ID              int
	RestrictionName string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Reservation struct {
	ID        int
	RoomID    int
	FirstName string
	LastName  string
	Email     string
	Phone     string
	Room      Room
	StartDate time.Time
	EndDate   time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RoomRestriction struct {
	ID            int
	RoomID        int
	ReservationID int
	RestricitonID int
	Room          Room
	Reservation   Reservation
	Restriction   Restriction
	StartDate     time.Time
	EndDate       time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
