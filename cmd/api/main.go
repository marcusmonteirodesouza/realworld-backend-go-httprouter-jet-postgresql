package main

import (
	"context"
	"database/sql"
	"expvar"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"

	"github.com/marcusmonteirodesouza/realworld-backend-go-jet-postgresql/internal/services"
)

type config struct {
	port int
}

type application struct {
	articlesService *services.ArticlesService
	config          *config
	db              *sql.DB
	logger          *slog.Logger
	profilesService *services.ProfilesService
	usersService    *services.UsersService
	wg              sync.WaitGroup
}

func main() {

	jwtIss := os.Getenv("JWT_ISS")
	if jwtIss == "" {
		log.Fatal("Environment variable JWT_ISS is required")
	}

	jwtKey := os.Getenv("JWT_KEY")
	if jwtKey == "" {
		log.Fatal("Environment variable JWT_KEY is required")
	}

	jwtValidForSeconds, err := strconv.Atoi(os.Getenv("JWT_VALID_FOR_SECONDS"))
	if err != nil {
		log.Fatal("Environment variable JWT_VALID_FOR_SECONDS is required and must be an integer")
	}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal("Environment variable PORT is required and must be an integer")
	}

	postgresDB := os.Getenv("POSTGRES_DB")
	if postgresDB == "" {
		log.Fatal("Environment variable POSTGRES_DB is required")
	}

	postgresHost := os.Getenv("POSTGRES_HOST")
	if postgresHost == "" {
		log.Fatal("Environment variable POSTGRES_HOST is required")
	}

	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	if postgresPassword == "" {
		log.Fatal("Environment variable POSTGRES_PASSWORD is required")
	}

	postgresPort, err := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
	if err != nil {
		log.Fatal("Environment variable POSTGRES_PORT is required")
	}

	postgresUser := os.Getenv("POSTGRES_USER")
	if postgresUser == "" {
		log.Fatal("Environment variable POSTGRES_USER is required")
	}

	ctx := context.Background()

	dsn := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		postgresHost, postgresPort, postgresUser, postgresPassword, postgresDB)

	db, err := openDB(ctx, dsn)
	if err != nil {
		log.Fatal(err.Error())
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal(err.Error())
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

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	usersServiceJWT := services.NewUsersServiceJWT(jwtIss, []byte(jwtKey), jwtValidForSeconds)

	usersService := services.NewUsersService(db, &usersServiceJWT, logger)

	profilesService := services.NewProfilesService(db, logger, &usersService)

	articlesService := services.NewArticlesService(db, logger, &usersService)

	app := &application{
		articlesService: &articlesService,
		db:              db,
		config:          config,
		logger:          logger,
		profilesService: &profilesService,
		usersService:    &usersService,
	}

	if err = app.serve(ctx); err != nil {
		log.Fatal(err.Error())
	}
}

func openDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
