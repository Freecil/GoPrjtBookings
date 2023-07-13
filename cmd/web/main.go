package main

//Usage of web application , to see it go to localhost:and port and slash
import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Freecil/GoPrjtBookings/internal/config"
	"github.com/Freecil/GoPrjtBookings/internal/driver"
	"github.com/Freecil/GoPrjtBookings/internal/handlers"
	"github.com/Freecil/GoPrjtBookings/internal/helpers"
	"github.com/Freecil/GoPrjtBookings/internal/models"
	"github.com/Freecil/GoPrjtBookings/internal/render"

	"github.com/alexedwards/scs/v2"
)

const portNr = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorlog *log.Logger

func main() {

	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()

	defer close(app.MailChan)
	fmt.Println("Starting email listener...")
	listenForMail()

	fmt.Println(fmt.Sprintf("starting application on port %s", portNr))

	srv := &http.Server{
		Addr:    portNr,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(models.RoomRestriction{})
	gob.Register(map[string]int{})

	inProduction := flag.Bool("production", true, "app is in production")
	useChache := flag.Bool("cache", true, "use template cache")
	dbName := flag.String("dbname", "bookings", "databasename")
	dbHost := flag.String("dbhost", "localhost", "database host")
	dbUser := flag.String("dbuser", "postgres", "database User")
	dbPassword := flag.String("dbpass", "password", "database password")
	dbPort := flag.String("dbport", "5432", "database port")
	dbSSL := flag.String("dbssl", "disable", "database ssl")

	flag.Parse()

	if *dbName == "" || *dbUser == "" {
		log.Println("Missing required flags")
		os.Exit(1)
	}

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	//change this to true when in pordutiocn
	app.InProduction = *inProduction
	app.UseCache = *useChache

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorlog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorlog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	app.Session = session

	//connect to database
	log.Println("Connecting to database...")
	connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", *dbHost, *dbPort, *dbName, *dbUser, *dbPassword, *dbSSL)
	db, err := driver.ConnectSQL(connectionString)
	if err != nil {
		log.Fatal("Cannot connect to dabase")
		return nil, err
	}
	log.Println("Connected to database")

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache", err)
		return nil, err
	}

	app.TemplateCache = tc

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, nil
}
