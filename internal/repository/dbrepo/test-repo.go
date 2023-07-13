package dbrepo

import (
	"errors"
	"log"
	"time"

	"github.com/Freecil/GoPrjtBookings/internal/models"
)

func (m *testDBRepo) AllUsers() bool {
	return true
}

// Inserts a reservaation into the database
func (m *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	if res.RoomID == 6969 {
		return 0, errors.New("error")
	}
	return 1, nil
}

// inserts room restrictions into database
func (m *testDBRepo) InsertRoomRestrictions(r models.RoomRestriction) error {
	if r.RoomID == 4444 {
		return errors.New("error")
	}
	return nil
}

// seraches avlaily using date for spesfik room, returns yes or no
func (m *testDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomId int) (bool, error) {
	layout := "2006-01-02"
	str := "2070-07-27"
	t, err := time.Parse(layout, str)
	if err != nil {
		log.Println(err)
	}

	// this is test to fail
	testDateToFail, err := time.Parse(layout, "2004-04-04")
	if err != nil {
		log.Println(err)
	}

	if start == testDateToFail {
		return false, errors.New("error")
	}

	if start.After(t) {
		return false, nil
	}
	return true, nil

}

// searches all rooms for avaliby and return the rooms if any avalile
func (m *testDBRepo) SearchAvailabilityByDates(start, end time.Time) ([]models.Room, error) {
	var rooms []models.Room

	layout := "2006-01-02"
	str := "2070-07-27"
	t, err := time.Parse(layout, str)
	if err != nil {
		log.Println(err)
	}

	testDateToFail, err := time.Parse(layout, "2004-04-04")
	if err != nil {
		log.Println(err)
	}

	if start == testDateToFail {
		return rooms, errors.New("error")
	}

	if start.After(t) {
		return rooms, nil
	}

	room := models.Room{
		ID: 1,
	}
	rooms = append(rooms, room)

	return rooms, nil
}

// Get the room by id
func (m *testDBRepo) GetRoomByID(id int) (models.Room, error) {
	var room models.Room
	if id > 2 {
		return room, errors.New("error")
	}
	return room, nil
}

// get user by id
func (m *testDBRepo) GetUserByID(id int) (models.User, error) {
	var u models.User

	return u, nil
}

// update user
func (m *testDBRepo) UpdateUser(u models.User) error {
	return nil
}

// authenicat user
func (m *testDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	if email != "fail@this.one" {
		return 1, "", nil
	}
	return 0, "", errors.New("error")

}

func (m *testDBRepo) AllReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation

	return reservations, nil
}

func (m *testDBRepo) NewReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation

	return reservations, nil
}
func (m *testDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	var res models.Reservation

	return res, nil
}

func (m *testDBRepo) UpdateReservation(r models.Reservation) error {
	return nil
}

func (m *testDBRepo) DeleteReservation(id int) error {
	return nil
}

func (m *testDBRepo) UpdatedProcessedForReservation(id, processed int) error {
	return nil
}

func (m *testDBRepo) AllRooms() ([]models.Room, error) {
	var rooms []models.Room
	rooms = append(rooms, models.Room{ID: 1})
	return rooms, nil
}

func (m *testDBRepo) GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {
	var restrictions []models.RoomRestriction

	restrictions = append(restrictions, models.RoomRestriction{
		ID:            1,
		StartDate:     time.Now(),
		EndDate:       time.Now().AddDate(0, 0, 1),
		RoomID:        1,
		ReservationID: 0,
		RestrictionID: 2,
	})

	restrictions = append(restrictions, models.RoomRestriction{
		ID:            2,
		StartDate:     time.Now().AddDate(0, 0, 2),
		EndDate:       time.Now().AddDate(0, 0, 3),
		RoomID:        1,
		ReservationID: 1,
		RestrictionID: 1,
	})

	return restrictions, nil
}

func (m *testDBRepo) DeleteBlockedRoom(restrictionID int) error {

	return nil
}

// inserts a block for by using date
func (m *testDBRepo) InsertBlockOnRoom(roomID int, start time.Time) error {

	return nil

}
