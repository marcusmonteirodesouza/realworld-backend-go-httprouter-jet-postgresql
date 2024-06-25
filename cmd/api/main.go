package main

import (
	"context"
	"database/sql"
	"expvar"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"cloud.google.com/go/logging"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"

	"github.com/marcusmonteirodesouza/realworld-backend-go-jet-postgresql/internal/services"
)

type config struct {
	port int
}

type application struct {
	config       *config
	db           *sql.DB
	logger       *logging.Logger
	usersService *services.UsersService
	wg           sync.WaitGroup
}

func main() {
	googleProjectID := os.Getenv("GOOGLE_PROJECT_ID")
	if googleProjectID == "" {
		log.Fatal("Environment variable GOOGLE_PROJECT_ID is required")
	}

	googleCloudRunService := os.Getenv("K_SERVICE")
	if googleCloudRunService == "" {
		log.Fatal("Environment variable K_SERVICE is required")
	}

	ctx := context.Background()

	loggingClient, err := logging.NewClient(ctx, googleProjectID)
	if err != nil {
		log.Fatalf("Failed to create logging client: %v", err)
	}
	defer loggingClient.Close()

	logger := loggingClient.Logger(googleCloudRunService, logging.RedirectAsJSON(os.Stdout))

	jwtIss := os.Getenv("JWT_ISS")
	if jwtIss == "" {
		logger.StandardLogger(logging.Critical).Fatal("Environment variable JWT_ISS is required")
	}

	jwtKey := os.Getenv("JWT_KEY")
	if jwtKey == "" {
		logger.StandardLogger(logging.Critical).Fatal("Environment variable JWT_KEY is required")
	}

	jwtValidForSeconds, err := strconv.Atoi(os.Getenv("JWT_VALID_FOR_SECONDS"))
	if err != nil {
		logger.StandardLogger(logging.Critical).Fatal("Environment variable JWT_VALID_FOR_SECONDS is required and must be an integer")
	}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		logger.StandardLogger(logging.Critical).Fatal("Environment variable PORT is required and must be an integer")
	}

	postgresDB := os.Getenv("POSTGRES_DB")
	if postgresDB == "" {
		logger.StandardLogger(logging.Critical).Fatal("Environment variable POSTGRES_DB is required")
	}

	postgresHost := os.Getenv("POSTGRES_HOST")
	if postgresHost == "" {
		logger.StandardLogger(logging.Critical).Fatal("Environment variable POSTGRES_HOST is required")
	}

	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	if postgresPassword == "" {
		logger.StandardLogger(logging.Critical).Fatal("Environment variable POSTGRES_PASSWORD is required")
	}

	postgresPort, err := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
	if err != nil {
		logger.StandardLogger(logging.Critical).Fatal("Environment variable POSTGRES_PORT is required")
	}

	postgresUser := os.Getenv("POSTGRES_USER")
	if postgresUser == "" {
		logger.StandardLogger(logging.Critical).Fatal("Environment variable POSTGRES_USER is required")
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		postgresHost, postgresPort, postgresUser, postgresPassword, postgresDB)

	db, err := openDB(ctx, dsn)
	if err != nil {
		logger.StandardLogger(logging.Critical).Fatal(err.Error())
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.StandardLogger(logging.Critical).Fatal(err.Error())
		}
	}()

	expvar.Publish("database", expvar.Func(func() interface{} {
		return db.Stats()
	}))

	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))

	config := &config{
		port: port,
	}

	app := &application{
		db:           db,
		config:       config,
		logger:       logger,
		usersService: services.NewUsersService(db, *services.NewUsersServiceJWT(jwtIss, []byte(jwtKey), jwtValidForSeconds), logger),
	}

	err = app.serve(ctx)
	if err != nil {
		app.logger.StandardLogger(logging.Critical).Fatal(err.Error())
	}
}

func openDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
