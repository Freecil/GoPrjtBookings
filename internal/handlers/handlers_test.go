package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Freecil/GoPrjtBookings/internal/driver"
	"github.com/Freecil/GoPrjtBookings/internal/models"
)

type postData struct {
	key   string
	value string
}

var theTest = []struct {
	name               string
	url                string
	method             string
	expectedStatusCode int
}{
	{"home", "/", "GET", http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"contact", "/contact", "GET", http.StatusOK},
	{"gq", "/generals-quarters", "GET", http.StatusOK},
	{"ms", "/majors-suite", "GET", http.StatusOK},
	{"sa", "/search-availability", "GET", http.StatusOK},
	{"mr", "/make-reservation", "GET", http.StatusOK},
	{"not-found", "/home/alone/by/myself", "GET", http.StatusNotFound},
	{"login", "/user/login", "GET", http.StatusOK},
	{"logout", "/user/logut", "GET", http.StatusOK},
	{"dashboard", "/admin/dashboard", "GET", http.StatusOK},
	{"new-res", "/admin/reservations-new", "GET", http.StatusOK},
	{"all-res", "/admin/reservations-all", "GET", http.StatusOK},
	{"single-res", "/admin/reservations/new/5/show", "GET", http.StatusOK},
	{"res-cal", "/admin/reservations-calender", "GET", http.StatusOK},
	{"res-cal + param", "/admin/reservations-calender?y=2020&m=1", "GET", http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, e := range theTest {
		resp, err := ts.Client().Get(ts.URL + e.url)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}
		if resp.StatusCode != e.expectedStatusCode {
			t.Errorf("for %s, expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
		}
	}
}

var reservationTest = []struct {
	name               string
	reservation        models.Reservation
	expectedStatusCode int
	expectedLocation   string
	expectedHTML       string
}{
	{
		name: "reservation-in-session",
		reservation: models.Reservation{
			RoomID: 1,
			Room: models.Room{
				ID:       1,
				RoomName: "General's Quarters",
			},
		},
		expectedStatusCode: http.StatusOK,
		expectedHTML:       `action="/make-reservation"`,
	},
	{
		name:               "no-reservation-in-session",
		reservation:        models.Reservation{},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
		expectedHTML:       "",
	},
	{
		name: "non-existing-room",
		reservation: models.Reservation{
			RoomID: 100,
			Room: models.Room{
				ID:       100,
				RoomName: "House Of Mine",
			},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
		expectedHTML:       "",
	},
}

// Testts the basinc make-reservation
func TestRepo_Reservation(t *testing.T) {
	for _, e := range reservationTest {
		req, _ := http.NewRequest("GET", "/make-reservation", nil)
		ctx := getCTX(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		if e.reservation.RoomID > 0 {
			session.Put(ctx, "reservation", e.reservation)
		}

		handler := http.HandlerFunc(Repo.Reservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("Get-Make-Reservation, %s returned wrong code: got %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}

		if e.expectedLocation != "" {
			// get the URL from test
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("Get-Make-Reservation, failed %s: expected location %s, but got location %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}

		if e.expectedHTML != "" {
			// read the response body into a string
			html := rr.Body.String()
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("Get-Make-Reservation, failed %s: expected to find %s but did not", e.name, e.expectedHTML)
			}
		}
	}
}

var postReservationTests = []struct {
	name                 string
	postedData           url.Values
	expectedResponseCode int
	expectedLocation     string
	expectedHTML         string
}{
	{
		name: "valid-data",
		postedData: url.Values{
			"start_date": {"2069-06-09"},
			"end_date":   {"2069-06-10"},
			"first_name": {"Helena"},
			"last_name":  {"Johnson"},
			"email":      {"Helena@Johnson.com"},
			"phone":      {"555"},
			"room_id":    {"1"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "/reservation-summary",
	},
	{
		name:                 "no-post-body",
		postedData:           nil,
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "/",
	},
	{
		name: "invalid-start-date",
		postedData: url.Values{
			"start_date": {"invalid"},
			"end_date":   {"2069-06-10"},
			"first_name": {"Cool"},
			"last_name":  {"Dude"},
			"email":      {"Cool@Dude.com"},
			"phone":      {"555"},
			"room_id":    {"1"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "/",
	},
	{
		name: "invalid-end-date",
		postedData: url.Values{
			"start_date": {"2069-06-09"},
			"end_date":   {"invalid"},
			"first_name": {"Fail"},
			"last_name":  {"Life"},
			"email":      {"Life@Heart.com"},
			"phone":      {"555"},
			"room_id":    {"1"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "/",
	},
	{
		name: "non-existing-room",
		postedData: url.Values{
			"start_date": {"2069-06-09"},
			"end_date":   {"2069-06-10"},
			"first_name": {"Country"},
			"last_name":  {"Road"},
			"email":      {"Take@Me.home"},
			"phone":      {"555"},
			"room_id":    {"69"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "/",
	},
	{
		name: "invalid-room-id",
		postedData: url.Values{
			"start_date": {"2069-06-09"},
			"end_date":   {"2069-06-10"},
			"first_name": {"Gnome"},
			"last_name":  {"Less"},
			"email":      {"Gnome@Buy.com"},
			"phone":      {"555"},
			"room_id":    {"invalid"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "/",
	},
	{
		name: "invalid-data",
		postedData: url.Values{
			"start_date": {"2069-06-09"},
			"end_date":   {"2069-06-10"},
			"first_name": {"L"},
			"last_name":  {"C04l"},
			"email":      {"Dud3"},
			"phone":      {"555"},
			"room_id":    {"1"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         `action="/make-reservation"`,
		expectedLocation:     "",
	},
	{
		name: "DB-Failure-Reservation",
		postedData: url.Values{
			"start_date": {"2069-06-09"},
			"end_date":   {"2069-06-10"},
			"first_name": {"Fail"},
			"last_name":  {"Here"},
			"email":      {"Here@Fail.com"},
			"phone":      {"555"},
			"room_id":    {"6969"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "/",
	},
	{
		name: "DB-Failure-Restriction",
		postedData: url.Values{
			"start_date": {"2069-06-09"},
			"end_date":   {"2069-06-10"},
			"first_name": {"Mistake"},
			"last_name":  {"Where"},
			"email":      {"Where@Mistake.com"},
			"phone":      {"555"},
			"room_id":    {"4444"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "/",
	},
}

func TestRepo_PostReservation(t *testing.T) {
	for _, e := range postReservationTests {
		var req *http.Request
		if e.postedData != nil {
			req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(e.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/make-reservation", nil)

		}
		ctx := getCTX(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.PostReservation)

		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedResponseCode {
			t.Errorf("POST-Make-Reservation, %s returned wrong code: got %d, wanted %d", e.name, rr.Code, e.expectedResponseCode)
		}

		if e.expectedLocation != "" {
			// get the URL from test
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("POST-Make-Reservation, failed %s: expected location %s, but got location %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}

		if e.expectedHTML != "" {
			// read the response body into a string
			html := rr.Body.String()
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("POST-Make-Reservation, failed %s: expected to find %s but did not", e.name, e.expectedHTML)
			}
		}

	}
}

var reservationSummaryTests = []struct {
	name               string
	reservation        models.Reservation
	url                string
	expectedStatusCode int
	expectedLocation   string
}{
	{
		name: "reservation-in-session",
		reservation: models.Reservation{
			RoomID: 1,
			Room: models.Room{
				ID:       1,
				RoomName: "General's Quarters",
			},
		},
		url:                "/reservation-summary",
		expectedStatusCode: http.StatusOK,
		expectedLocation:   "",
	},
	{
		name:               "reservation-not-in-session",
		reservation:        models.Reservation{},
		url:                "/reservation-summary",
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
}

func TestRepo_ReservationSummary(t *testing.T) {
	for _, e := range reservationSummaryTests {
		req, _ := http.NewRequest("GET", e.url, nil)
		ctx := getCTX(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		if e.reservation.RoomID > 0 {
			session.Put(ctx, "reservation", e.reservation)
		}

		handler := http.HandlerFunc(Repo.ReservationSummary)

		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("ReservationSummary, %s returned wrong response code: got %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}

		if e.expectedLocation != "" {
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("ReservationSummary, failed %s: expected location %s, but got location %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}
	}
}

var testAvailabilityJSONData = []struct {
	name            string
	postedData      url.Values
	expectedOK      bool
	expectedMessage string
}{
	{
		name: "room not available",
		postedData: url.Values{
			"start_date": {"2071-07-27"},
			"end_date":   {"2071-07-28"},
			"room_id":    {"1"},
		},
		expectedOK: false,
	}, {
		name: "rooms are available",
		postedData: url.Values{
			"start_date": {"2001-01-01"},
			"end_date":   {"2001-01-02"},
			"room_id":    {"1"},
		},
		expectedOK: true,
	},
	{
		name:            "empty post body",
		postedData:      nil,
		expectedOK:      false,
		expectedMessage: "server error",
	},
	{
		name: "database query fails",
		postedData: url.Values{
			"start_date": {"2004-04-04"},
			"end_date":   {"2004-04-06"},
			"room_id":    {"1"},
		},
		expectedOK:      false,
		expectedMessage: "error query database",
	},
}

func TestRepo_AvailabilityJSON(t *testing.T) {
	for _, e := range testAvailabilityJSONData {
		// create request, get the context with session, set header, create recorder
		var req *http.Request
		if e.postedData != nil {
			req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader(e.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/search-availability-json", nil)
		}
		ctx := getCTX(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// make our handler a http.HandlerFunc and call
		handler := http.HandlerFunc(Repo.AvailabilityJSON)
		handler.ServeHTTP(rr, req)

		var j jsonResponse
		err := json.Unmarshal([]byte(rr.Body.String()), &j)
		if err != nil {
			t.Error("failed to parse json!")
		}

		if j.OK != e.expectedOK {
			t.Errorf("Get-AvailabilityJSON, %s: expected %v but got %v", e.name, e.expectedOK, j.OK)
		}
	}

}

var testPostAvailabilityData = []struct {
	name               string
	postedData         url.Values
	expectedStatusCode int
	expectedLocation   string
}{
	{
		name: "rooms not available",
		postedData: url.Values{
			"start": {"2071-07-27"},
			"end":   {"2071-07-28"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name: "rooms are available",
		postedData: url.Values{
			"start":   {"2001-01-01"},
			"end":     {"2001-01-02"},
			"room_id": {"1"},
		},
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "empty post body",
		postedData:         url.Values{},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name: "start date wrong format",
		postedData: url.Values{
			"start":   {"invalid"},
			"end":     {"2001-01-02"},
			"room_id": {"1"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name: "end date wrong format",
		postedData: url.Values{
			"start": {"2001-01-01"},
			"end":   {"invalid"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name: "database query fails",
		postedData: url.Values{
			"start": {"2004-04-04"},
			"end":   {"2004-04-06"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
}

func TestRepo_PostAvailability(t *testing.T) {
	for _, e := range testPostAvailabilityData {
		req, _ := http.NewRequest("POST", "/search-availability", strings.NewReader(e.postedData.Encode()))

		// get the context with session
		ctx := getCTX(req)
		req = req.WithContext(ctx)

		// set the request header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// make our handler a http.HandlerFunc and call
		handler := http.HandlerFunc(Repo.PostAvailability)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("Post-Availability, %s gave wrong status code: got %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}
	}
}

func TestNewRepo(t *testing.T) {
	var db driver.DB
	testRepo := NewRepo(&app, &db)

	if reflect.TypeOf(testRepo).String() != "*handlers.Repository" {
		t.Errorf("Wrong type from newrepo : got %s, wanted *Repository", reflect.TypeOf(testRepo).String())
	}
}

var chooseRoomTests = []struct {
	name               string
	reservation        models.Reservation
	url                string
	expectedStatusCode int
	expectedLocation   string
}{
	{
		name: "reservation-in-session",
		reservation: models.Reservation{
			RoomID: 1,
			Room: models.Room{
				ID:       1,
				RoomName: "General's Quarters",
			},
		},
		url:                "/choose-room/1",
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/make-reservation",
	},
	{
		name:               "reservation-not-in-session",
		reservation:        models.Reservation{},
		url:                "/choose-room/1",
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
	{
		name:               "misformed-url",
		reservation:        models.Reservation{},
		url:                "/choose-room/cat-house",
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
}

func TestRepo_ChooseRoom(t *testing.T) {
	for _, e := range chooseRoomTests {
		req, _ := http.NewRequest("GET", e.url, nil)
		ctx := getCTX(req)
		req = req.WithContext(ctx)
		// set the RequestURI on the request so that we can grab the ID from the URL
		req.RequestURI = e.url

		rr := httptest.NewRecorder()
		if e.reservation.RoomID > 0 {
			session.Put(ctx, "reservation", e.reservation)
		}

		handler := http.HandlerFunc(Repo.ChooseRoom)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("ChooseRoom, %s returned wrong response code: got %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}

		if e.expectedLocation != "" {
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("ChooseRoom, failed %s: expected location %s, but got location %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}
	}
}

var bookRoomTests = []struct {
	name               string
	url                string
	expectedStatusCode int
}{
	{
		name:               "database-works",
		url:                "/book-room?s=2070-08-07&e=2070-08-07&id=1",
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name:               "database-fails",
		url:                "/book-room?s=2001-01-01&e=2001-01-02&id=69",
		expectedStatusCode: http.StatusSeeOther,
	},
}

func TestRepo_BookRoom(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	for _, e := range bookRoomTests {
		req, _ := http.NewRequest("GET", e.url, nil)
		ctx := getCTX(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		session.Put(ctx, "reservation", reservation)

		handler := http.HandlerFunc(Repo.BookRoom)

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Errorf("BookRoom, %s failed: returned wrong response code: got %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}
	}
}

var loginTests = []struct {
	name               string
	email              string
	expectedStatusCode int
	expectedHTML       string
	expectedLocation   string
}{
	{
		"valid-credentials",
		"valid@here.com",
		http.StatusSeeOther,
		"",
		"/",
	},
	{
		"invalid-credentials",
		"fail@this.one",
		http.StatusSeeOther,
		"",
		"/user/login",
	},
	{
		"invalid-data",
		"L",
		http.StatusOK,
		`action="/user/login"`,
		"",
	},
}

func TestRepo_Login(t *testing.T) {
	for _, e := range loginTests {
		postedData := url.Values{}
		postedData.Add("email", e.email)
		postedData.Add("password", "password")

		req, _ := http.NewRequest("POST", "/user/login", strings.NewReader(postedData.Encode()))
		ctx := getCTX(req)
		req = req.WithContext(ctx)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.PostShowLogin)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("Login, failed %s: expected code %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}

		if e.expectedLocation != "" {
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("Login, failed %s: expected location %s, but got location %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}

		if e.expectedHTML != "" {
			html := rr.Body.String()
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("Login, failed %s: expected to find %s but did not", e.name, e.expectedHTML)
			}
		}
	}
}

var adminPostSingleReservationTests = []struct {
	name                 string
	url                  string
	postedData           url.Values
	expectedResponseCode int
	expectedLocation     string
	expectedHTML         string
}{
	{
		name: "valid-data-from-new",
		url:  "/admin/reservations/new/1/show",
		postedData: url.Values{
			"first_name": {"PersonName"},
			"last_name":  {"LastPersonKilled"},
			"email":      {"Dead@Funeral.com"},
			"phone":      {"555"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "/admin/reservations-new",
		expectedHTML:         "",
	},
	{
		name: "valid-data-from-all",
		url:  "/admin/reservations/all/1/show",
		postedData: url.Values{
			"first_name": {"See"},
			"last_name":  {"Here"},
			"email":      {"Sing@With.me"},
			"phone":      {"555"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "/admin/reservations-all",
		expectedHTML:         "",
	},
	{
		name: "valid-data-from-cal",
		url:  "/admin/reservations/cal/1/show",
		postedData: url.Values{
			"first_name": {"First"},
			"last_name":  {"Second"},
			"email":      {"Third@Fourth.com"},
			"phone":      {"555"},
			"year":       {"2023"},
			"month":      {"04"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "/admin/reservations-calender?y=2023&m=04",
		expectedHTML:         "",
	},
}

func TestRepo_AdminPOSTSingleReservation(t *testing.T) {
	for _, e := range adminPostSingleReservationTests {
		var req *http.Request
		if e.postedData != nil {
			req, _ = http.NewRequest("POST", "/user/login", strings.NewReader(e.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/user/login", nil)
		}
		ctx := getCTX(req)
		req = req.WithContext(ctx)
		req.RequestURI = e.url

		// set the header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// call the handler
		handler := http.HandlerFunc(Repo.AdminPOSTSingleReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedResponseCode {
			t.Errorf("Post-AdminSingleReservation, failed %s: expected code %d, but got %d", e.name, e.expectedResponseCode, rr.Code)
		}

		if e.expectedLocation != "" {
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("Post-AdminSingleReservation, failed %s: expected location %s, but got location %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}

		if e.expectedHTML != "" {
			html := rr.Body.String()
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("Post-AdminSingleReservation, failed %s: expected to find %s but did not", e.name, e.expectedHTML)
			}
		}
	}
}

var adminPostReservationCalendarTests = []struct {
	name                 string
	postedData           url.Values
	expectedResponseCode int
	expectedLocation     string
	expectedHTML         string
	blocks               int
	reservations         int
}{
	{
		name: "cal",
		postedData: url.Values{
			"year":  {time.Now().Format("2006")},
			"month": {time.Now().Format("01")},
			fmt.Sprintf("add_block_1_%s", time.Now().AddDate(0, 0, 2).Format("2006-01-2")): {"1"},
		},
		expectedResponseCode: http.StatusSeeOther,
	},
	{
		name:                 "calender-blocks",
		postedData:           url.Values{},
		expectedResponseCode: http.StatusSeeOther,
		blocks:               1,
	},
	{
		name:                 "calender-res",
		postedData:           url.Values{},
		expectedResponseCode: http.StatusSeeOther,
		reservations:         1,
	},
}

func TestRepo_AdminPostReservationCalender(t *testing.T) {
	for _, e := range adminPostReservationCalendarTests {
		var req *http.Request
		if e.postedData != nil {
			req, _ = http.NewRequest("POST", "/admin/reservations-calender", strings.NewReader(e.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/admin/reservations-calender", nil)
		}
		ctx := getCTX(req)
		req = req.WithContext(ctx)

		now := time.Now()
		bm := make(map[string]int)
		rm := make(map[string]int)

		currentYear, currentMonth, _ := now.Date()
		currentLocation := now.Location()

		firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
		lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {
			rm[d.Format("2006-01-2")] = 0
			bm[d.Format("2006-01-2")] = 0
		}

		if e.blocks > 0 {
			bm[firstOfMonth.Format("2006-01-2")] = e.blocks
		}

		if e.reservations > 0 {
			rm[lastOfMonth.Format("2006-01-2")] = e.reservations
		}

		session.Put(ctx, "block_map_1", bm)
		session.Put(ctx, "reservation_map_1", rm)

		// set the header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// call the handler
		handler := http.HandlerFunc(Repo.AdminPostReservationsCalender)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedResponseCode {
			t.Errorf("Post-AdminReservationCalender, failed %s: expected code %d, but got %d", e.name, e.expectedResponseCode, rr.Code)
		}

	}
}

var adminProcessReservationTests = []struct {
	name                 string
	queryParams          string
	expectedResponseCode int
	expectedLocation     string
}{
	{
		name:                 "process-reservation",
		queryParams:          "",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
	{
		name:                 "process-reservation-back-to-calender",
		queryParams:          "?y=2023&m=07",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
}

func TestRepo_AdminProcessReservation(t *testing.T) {
	for _, e := range adminProcessReservationTests {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/admin/process-reservation/cal/1/do%s", e.queryParams), nil)
		ctx := getCTX(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AdminProcessReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Errorf("Process-Reservation, failed %s: expected code %d, but got %d", e.name, e.expectedResponseCode, rr.Code)
		}
	}
}

var adminDeleteReservationTests = []struct {
	name                 string
	queryParams          string
	expectedResponseCode int
	expectedLocation     string
}{
	{
		name:                 "delete-reservation",
		queryParams:          "",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
	{
		name:                 "delete-reservation-back-to-calender",
		queryParams:          "?y=2023&m=07",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
}

func TestRepo_AdminDeleteReservation(t *testing.T) {
	for _, e := range adminDeleteReservationTests {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/admin/process-reservation/cal/1/do%s", e.queryParams), nil)
		ctx := getCTX(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AdminDeleteReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedResponseCode, rr.Code)
		}
	}
}

func getCTX(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx

}
