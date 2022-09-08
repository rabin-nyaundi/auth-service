package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"rabitech.auth.app/cmd/internal/data"
	"rabitech.auth.app/cmd/internal/data/mailer"

	_ "github.com/lib/pq"
)

// colors used in logging text in terminal
// const colorCyan = "\033[36m"
// const colorGreen = "\033[32m"
// const colorRed = "\033[31m"

type envelope map[string]interface{}

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}

	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedURLOrigins []*url.URL
	}
}
type application struct {
	config config
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

var (
	buildTime string
	version   string
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Env variables not loaded")
	}

	var cfg config

	flag.IntVar(&cfg.port, "port", 4002, "API Server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment(development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "database connection string")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL maximum open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL maximum idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "10m", "PostgreSQL maximum idle time")

	flag.StringVar(&cfg.smtp.host, "smtp-host", os.Getenv("SMTP_HOST"), "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 587, "SMPT port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", os.Getenv("SMTP_USERNAME"), "SMTP Username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", os.Getenv("SMTP_PASSWORD"), "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", os.Getenv("SMTP_SENDER"), "SMTP sender")

	// cors flags
	flag.Func("cors-trusted-origins", "list allowd origin urls", func(s string) error {
		for _, u := range strings.Fields(s) {
			parsedURL, err := url.Parse(u)
			if err != nil {
				return err
			}
			cfg.cors.trustedURLOrigins = append(cfg.cors.trustedURLOrigins, parsedURL)
		}
		return nil
	})
	// Version flag
	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	//  Print the build time and version of the application
	if *displayVersion {
		fmt.Printf("Version: \t%s\n", version)
		fmt.Printf("Build time: \t%s\n", buildTime)
		os.Exit(0)
	}

	db, err := openDB(cfg)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer db.Close()

	fmt.Println("Database connection successful")

	expvar.NewString("version").Set(version)

	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("database", expvar.Func(func() interface{} {
		return db.Stats()
	}))

	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))

	app := &application{
		config: cfg,
		models: data.NewModel(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	err = app.serve()

	if err != nil {
		fmt.Println(err)
		return
	}

}

// The function opens a postgres database connection takes postgres connection string.
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)

	if err != nil {
		fmt.Println("ERROR")
		return nil, err
	}
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)

	if err != nil {
		fmt.Println("ERROR")
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)

	if err != nil {
		fmt.Println("ERROR")
		return nil, err
	}

	return db, nil
}
