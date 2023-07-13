package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Freecil/GoPrjtBookings/internal/config"
	"github.com/Freecil/GoPrjtBookings/internal/driver"
	"github.com/Freecil/GoPrjtBookings/internal/forms"
	"github.com/Freecil/GoPrjtBookings/internal/helpers"
	"github.com/Freecil/GoPrjtBookings/internal/models"
	"github.com/Freecil/GoPrjtBookings/internal/render"
	"github.com/Freecil/GoPrjtBookings/internal/repository"
	"github.com/Freecil/GoPrjtBookings/internal/repository/dbrepo"
	"github.com/go-chi/chi/v5"
)

var Repo *Repository

// The Repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// creates a new Repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingRepo(a),
	}
}

// sets the Repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r

}

// The below are handlers
// HOME
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, "home.page.tmpl", &models.TemplateData{}, r)
}

// ABOUT
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {

	render.Template(w, "about.page.tmpl", &models.TemplateData{}, r)

}

// SEARCH AVAILABILITY the page displaying search
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, "search-availability.page.tmpl", &models.TemplateData{}, r)

}

// POST AVAILABILITY wokres to do the searching
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "Error", "cant parse form")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	start := r.Form.Get("start")
	end := r.Form.Get("end")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, start)
	if err != nil {
		m.App.Session.Put(r.Context(), "Error", "cant parse start date")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	endDate, err := time.Parse(layout, end)
	if err != nil {
		m.App.Session.Put(r.Context(), "Error", "cant parse end date ")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	rooms, err := m.DB.SearchAvailabilityByDates(startDate, endDate)
	if err != nil {
		m.App.Session.Put(r.Context(), "Error", "cant find dates")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return

	}

	if len(rooms) == 0 {
		//Then no rooms avlaibe
		m.App.Session.Put(r.Context(), "Error", "No Availability")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(r.Context(), "reservation", res)

	render.Template(w, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	}, r)
}

// AVAILABILITY JSON works with the page to do json
type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {

		resp := jsonResponse{
			OK:      false,
			Message: "server error",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")

	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	roomID, _ := strconv.Atoi(r.Form.Get("room_id"))

	available, err := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomID)
	if err != nil {

		resp := jsonResponse{
			OK:      false,
			Message: "error query database",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	resp := jsonResponse{
		OK:        available,
		Message:   "",
		RoomID:    strconv.Itoa(roomID),
		StartDate: sd,
		EndDate:   ed,
	}

	out, _ := json.MarshalIndent(resp, "", "     ")

	w.Header().Set("Content-Type", "application/json ")
	w.Write(out)
}

// MAJORS room
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, "majors.page.tmpl", &models.TemplateData{}, r)

}

// GENERALS room
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, "generals.page.tmpl", &models.TemplateData{}, r)
}

// CONTACT
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, "contact.page.tmpl", &models.TemplateData{}, r)

}

// RESERVATION  the  make reservation page
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "Error", "Cant get reservation from session")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "Error", "Cant find room")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	res.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", res)

	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	stringmap := make(map[string]string)
	stringmap["start_date"] = sd
	stringmap["end_date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, "make-reservation.page.tmpl", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringmap,
	}, r)

}

// POST RESERVATION redirects from form at make reservation to summary
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "Error", "Cant parse form")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	start := r.Form.Get("start_date")
	end := r.Form.Get("end_date")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, start)
	if err != nil {
		m.App.Session.Put(r.Context(), "Error", "Can't parse start date")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	endDate, err := time.Parse(layout, end)
	if err != nil {
		m.App.Session.Put(r.Context(), "Error", "Can't parse end date ")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		m.App.Session.Put(r.Context(), "Error", "invalid data!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "Error", "cannot find room")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    roomID,
		Room:      room,
	}

	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		w.WriteHeader(http.StatusSeeOther)
		render.Template(w, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		}, r)
		return
	}

	newReservationId, err := m.DB.InsertReservation(reservation)
	if err != nil {
		m.App.Session.Put(r.Context(), "Error", "Could not insert reservation to database!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: newReservationId,
		RestrictionID: 1,
	}

	err = m.DB.InsertRoomRestrictions(restriction)
	if err != nil {
		m.App.Session.Put(r.Context(), "Error", "Could not insert restrictions to database!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	//Send msg to customer
	htmlMessage := fmt.Sprintf(`
	<strong>Reservation Confirmation</strong><br>
	Dear %s: <br>
	This is a conformation of your reservation at <strong>Fort Smythe Bed and Breakfast</strong>, from the %s to the %s.
	`, reservation.FirstName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg := models.MailData{
		To:      reservation.Email,
		From:    "FortSmythe@BB.com",
		Subject: "Reservation Confirmation",
		Content: htmlMessage,
	}

	m.App.MailChan <- msg
	//Send msg to opproperty roperty owner
	htmlMessage = fmt.Sprintf(`
	<strong>Reservation Notification</strong><br>
	A reservation has been made by %s, %s from the %s to the %s in room %s.<br>

	`, reservation.FirstName, reservation.LastName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"), reservation.Room.RoomName)

	msg = models.MailData{
		To:      "FortSmythe@BB.com",
		From:    "FortSmythe@BB.com",
		Subject: "Reservation Notication",
		Content: htmlMessage,
	}

	m.App.MailChan <- msg

	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// RESERVATION SUMMARY display  the eresvation summary
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "Error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	render.Template(w, "reservation-summary.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	}, r)

}

// CHOOSE ROOM from search available
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	explosion := strings.Split(r.RequestURI, "/")
	roomID, err := strconv.Atoi(explosion[2])
	if err != nil {
		m.App.Session.Put(r.Context(), "Error", "missing url parts")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "Error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	res.RoomID = roomID

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)

}

// BOOK ROOM redirection to make reservation
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {

	roomID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		m.App.Session.Put(r.Context(), "Error", "invalid data")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	sd := r.URL.Query().Get("s")
	ed := r.URL.Query().Get("e")

	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	var res models.Reservation

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "Error", "cant get room from database")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	res.RoomID = roomID
	res.Room.RoomName = room.RoomName
	res.StartDate = startDate
	res.EndDate = endDate

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// Renders the login page
func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Template(w, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	}, r)
}

// Handels loggin the user in
func (m *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "Error", "Invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	var email string
	var password string
	email = r.Form.Get("email")
	password = r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")
	if !form.Valid() {
		render.Template(w, "login.page.tmpl", &models.TemplateData{
			Form: form,
		}, r)
		return
	}

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "Error", "Invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "Flash", "Logged in successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Loguts out the user
func (m *Repository) Logut(w http.ResponseWriter, r *http.Request) {

	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())
	m.App.Session.Put(r.Context(), "Flash", "Logged ut successfully!")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)

}

// admin dashboard
func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {

	render.Template(w, "admin-dashboard.page.tmpl", &models.TemplateData{}, r)
}

// new reservations page
func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.NewReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, "admin-new-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	}, r)
}

// all reservations  page
func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, "admin-all-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	}, r)
}

func (m *Repository) AdminSingleReservation(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	stringMap := make(map[string]string)
	stringMap["src"] = src

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	stringMap["month"] = month
	stringMap["year"] = year

	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = res

	render.Template(w, "admin-reservation-single.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
		Form:      forms.New(nil),
	}, r)
}

func (m *Repository) AdminPOSTSingleReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	stringMap := make(map[string]string)
	stringMap["src"] = src

	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		m.App.Session.Put(r.Context(), "Error", "Form not valid ")
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s/%d/show", src, id), http.StatusSeeOther)
		return
	}

	err = m.DB.UpdateReservation(res)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	month := r.Form.Get("month")
	year := r.Form.Get("year")

	m.App.Session.Put(r.Context(), "Flash", "Changes saved")
	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calender?y=%s&m=%s", year, month), http.StatusSeeOther)
	}

}

// Marks the reservation as
func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")
	err := m.DB.UpdatedProcessedForReservation(id, 1)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	m.App.Session.Put(r.Context(), "Flash", "Marked as processed")
	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calender?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

// Delete a resvation
func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")
	err := m.DB.DeleteReservation(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	m.App.Session.Put(r.Context(), "Warning", "Reservation has been Deleted!")
	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calender?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

// PAge for the clader over all resvation and just claneder so owner cna close rooms
func (m *Repository) AdminReservationsCalender(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	if r.URL.Query().Get("y") != "" {
		year, err := strconv.Atoi(r.URL.Query().Get("y"))
		if err != nil {
			helpers.ServerError(w, err)
			return
		}
		month, err := strconv.Atoi(r.URL.Query().Get("m"))
		if err != nil {
			helpers.ServerError(w, err)
			return
		}
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	data := make(map[string]interface{})
	data["now"] = now

	next := now.AddDate(0, 1, 0)
	last := now.AddDate(0, -1, 0)

	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")

	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	stringMap := make(map[string]string)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear

	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data["rooms"] = rooms

	for _, x := range rooms {
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format("2006-01-02")] = 0
			blockMap[d.Format("2006-01-02")] = 0

		}

		restrictions, err := m.DB.GetRestrictionsForRoomByDate(x.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		for _, y := range restrictions {
			if y.ReservationID > 0 {
				//reservation
				for d := y.StartDate; d.After(y.EndDate) == false; d = d.AddDate(0, 0, 1) {
					reservationMap[d.Format("2006-01-02")] = y.ReservationID
				}
			} else {
				//block
				blockMap[y.StartDate.Format("2006-01-02")] = y.ID
			}
		}
		data[fmt.Sprintf("reservation_map_%d", x.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", x.ID)] = blockMap

		m.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", x.ID), blockMap)
		m.App.Session.Put(r.Context(), fmt.Sprintf("reservation_map_%d", x.ID), reservationMap)

	}

	render.Template(w, "admin-reservations-calender.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		IntMap:    intMap,
	}, r)
}

// Handesl post of restrvation calender
func (m *Repository) AdminPostReservationsCalender(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		m.App.Session.Put(r.Context(), "Error", "Cant parse form")
		http.Redirect(w, r, "/admin/reservations-calender", http.StatusTemporaryRedirect)
		return
	}

	year, _ := strconv.Atoi(r.Form.Get("y"))
	month, _ := strconv.Atoi(r.Form.Get("m"))

	//The check and unchecks

	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	layout := "2006-01-02"
	form := forms.New(r.PostForm)

	for _, x := range rooms {
		curMap := m.App.Session.Get(r.Context(), fmt.Sprintf("block_map_%d", x.ID)).(map[string]int)
		for name, value := range curMap {
			if val, ok := curMap[name]; ok {
				if val > 0 {
					if !form.Has(fmt.Sprintf("remove_block_%d_%s", x.ID, name)) {
						err := m.DB.DeleteBlockedRoom(value)
						if err != nil {
							log.Println(err)
							return
						}
					}
				}
			}
		}
	}

	for name, _ := range r.PostForm {
		if strings.HasPrefix(name, "add_block") {
			explosion := strings.Split(name, "_")
			roomID, _ := strconv.Atoi(explosion[2])
			t, err := time.Parse(layout, explosion[3])
			if err != nil {
				helpers.ServerError(w, err)
				return
			}
			err = m.DB.InsertBlockOnRoom(roomID, t)
			if err != nil {
				helpers.ServerError(w, err)
				return
			}
		}
	}

	m.App.Session.Put(r.Context(), "Flash", "Changed has been saved")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calender?y=%d&m=%d", year, month), http.StatusSeeOther)

}
