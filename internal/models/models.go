package models

import (
	"time"
)

// Users is the model for user
type User struct {
	ID          int
	FirstName   string
	LastName    string
	Email       string
	Password    string
	AccessLevel int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Rooms is the model for rooms
type Room struct {
	ID        int
	RoomName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Reservations is the model for reservations
type Reservation struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	Phone     string
	StartDate time.Time
	EndDate   time.Time
	RoomID    int
	CreatedAt time.Time
	UpdatedAt time.Time
	Processed int
	Room      Room
}

// restrictions is the model for restrictions
type Restriction struct {
	ID              int
	RestrictionName string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// RoomRestrictions is the model for room Restrictions
type RoomRestriction struct {
	ID            int
	StartDate     time.Time
	EndDate       time.Time
	RoomID        int
	ReservationID int
	RestrictionID int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Room          Room
	Reservation   Reservation
	Restriction   Restriction
}

// Holdes an email message
type MailData struct {
	To      string
	From    string
	Subject string
	Content string
}
