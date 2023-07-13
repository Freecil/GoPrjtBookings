package repository

import (
	"time"

	"github.com/Freecil/GoPrjtBookings/internal/models"
)

type DatabaseRepo interface {
	AllUsers() bool

	InsertRoomRestrictions(r models.RoomRestriction) error
	GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error)
	InsertBlockOnRoom(roomID int, start time.Time) error
	DeleteBlockedRoom(restrictionID int) error

	SearchAvailabilityByDatesByRoomID(start, end time.Time, roomId int) (bool, error)
	SearchAvailabilityByDates(start, end time.Time) ([]models.Room, error)

	GetRoomByID(id int) (models.Room, error)
	AllRooms() ([]models.Room, error)

	GetUserByID(id int) (models.User, error)
	UpdateUser(u models.User) error
	Authenticate(email, testPassword string) (int, string, error)

	InsertReservation(res models.Reservation) (int, error)
	AllReservations() ([]models.Reservation, error)
	NewReservations() ([]models.Reservation, error)
	GetReservationByID(id int) (models.Reservation, error)
	UpdateReservation(r models.Reservation) error
	DeleteReservation(id int) error
	UpdatedProcessedForReservation(id, processed int) error
}
