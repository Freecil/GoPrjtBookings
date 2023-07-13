# Go Course project

This is a repository for the Go course project called bookings and reservations on Udemy

-built in Go version 1.20
-Uses [chi router](github.com/go-chi/chi/v5)
-Uses [alex edwards csc session managment](github.com/alexedwards/scs/v2)
-uses [nosurf](github.com/justinas/nosurf)
-Uses [Go Simple Mail v2.14.0](github.com/xhit/go-simple-mail/v2) 
-Uses [PGX v5 jackc](github.com/jackc/pgx) 

To run the program be in top folder and use 
go run ./cmd/web 
To run test go 
go test ./...

Uses postgres database
Uses MailHog for mail server
to set up databse use postgres and migrate to set up the tabels
Settings for Database setup in databse.yml