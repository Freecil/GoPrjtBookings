package main

import (
	"net/http"

	"github.com/Freecil/GoPrjtBookings/internal/helpers"
	"github.com/justinas/nosurf"
)

// add CSRF protection to all post request
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
	return csrfHandler
}

// Loads and svae the  seesion on evert request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)

}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !helpers.IsAuthenticated(r) {
			session.Put(r.Context(), "Error", "Need to be logged in!")
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
