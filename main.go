package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"url-shorter/db"
	"url-shorter/handlers"
	"url-shorter/middleware"
)

type config struct {
	Port int `env:"HTTP_PORT" envDefault:"8000"`
	DBConnStr string `env:"DB_CONN_STR" envDefault:"host=localhost user=db_user dbname=url sslmode=disable"`
}

func main() {
	// curl -v localhost:8000/save\?url=https://www.ozon.ru

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		panic(fmt.Sprintf("cannot parse configuration: %s", err.Error()))
	}

	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Logger()

	logger = logger.Output(zerolog.NewConsoleWriter())

	dbConn, err := sql.Open("postgres", cfg.DBConnStr)
	if err != nil {
		logger.Fatal().Err(err).Msg("cannot connect to databse")
	}

	if err := dbConn.Ping(); err != nil {
		logger.Panic().Err(err).Msg("cannot ping database")
	}

	h := handlers.New(db.NewPostgres(dbConn))
	// h := NewHandler(db.NewInMemory())

	r := chi.NewRouter()
	r.Use(hlog.NewHandler(logger))

	r.Use(middleware.RequestID("req_id", "Request-Id"))
	r.Use(middleware.RequestLogger)

	r.Get("/", handlers.Help)
	r.Get("/save", h.SaveURL) // TODO Post
	r.Get("/go", h.GetURL)

	logger.Info().Msgf("start HTTP at a port %d", cfg.Port)
	panic(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), r))
}