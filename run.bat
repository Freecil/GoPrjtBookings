go build -o bookings.exe cmd/web/*.go
./bookings.exe -dbname=bookings -dbuser=postgres -cache=false -production=false -dbpassword=password -dbport=5432 -dbhost=localhost