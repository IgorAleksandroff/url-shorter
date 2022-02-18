package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

type config struct {
	Port int `env:"HTTP_PORT" envDefault:"8000"`
	DBConnStr string `env:"DB_CONN_STR" envDefault:"host=localhost user=db_user dbname=url sslmode=disable"`
}

type db map[string]*url.URL

func (d db) Set(key string, value *url.URL) {
	d[key] = value
}

func (d db) Get(key string) *url.URL {
	return d[key]
}

type handler struct {
	db db
	sqlDB *sql.DB
}

func main() {
	// curl -v localhost:8000/save\?url=https://www.ozon.ru

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		panic(fmt.Sprintf("cannot parse configuration: %s", err.Error()))
	}

	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("foo", "bar").
		Logger()

		logger = logger.Output(zerolog.NewConsoleWriter())

		dbConn, err := sql.Open("postgres", cfg.DBConnStr)
		if err != nil {
			logger.Fatal().Err(err).Msg("cannot connect to databse")
		}
	
		if err := dbConn.Ping(); err != nil {
			logger.Fatal().Err(err).Msg("cannot ping database")
		}
	
		h := handler{
			db:    db{},
			sqlDB: dbConn,
		}

	r := chi.NewRouter()
	r.Use(hlog.NewHandler(logger))

	r.Use(RequestIDHandler("req_id", "Request-Id"))
	r.Use(hlog.RemoteAddrHandler("ip"))
	r.Use(middleware)

	r.Get("/save", h.saveURL)
	r.Get("/go", h.getURL)
	r.Get("/echo", echo)

	logger.Info().Msgf("start HTTP at a port %d", cfg.Port)
	panic(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), r))
}

func (h handler) saveURL(w http.ResponseWriter, r *http.Request) {
	// /save?url=https://www.ozon.ru
	ctx := r.Context()
	logger := zerolog.Ctx(ctx)

	v := r.URL.Query().Get("url")
	key := uuid.New().String()

	var msg string
	parsedURL, err := url.Parse(v)

	if err != nil {
		log.Printf("ERR: %s", err.Error())

		msg = fmt.Sprintf("некорректный URL: %q", v)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
		
		return
	}

	//h.db[key] = parsedURL

	// Save URL in our sql
	_, err = h.sqlDB.ExecContext(ctx, "INSERT INTO urls (token, url) VALUES ($1, $2);", key, parsedURL.String())
	if err != nil {
		logger.Error().Err(err).Msg("cannot insert into sql database")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	msg = fmt.Sprintf("короткий URL: %q", key)
	w.Write([]byte(msg))
	logger.Info().Msgf("saveURL %s %s", key, parsedURL)

}

func (h handler) getURL(w http.ResponseWriter, r *http.Request) {
	// /go?to=uuid
	ctx := r.Context()
	logger := zerolog.Ctx(ctx)

	key := r.URL.Query().Get("to")
	//requestedURL := h.db.Get(key)
	var requestedURL string
	row := h.sqlDB.QueryRowContext(ctx, "SELECT url FROM urls WHERE token=$1", key)

	if err := row.Scan(&requestedURL); err != nil {
		if err == sql.ErrNoRows {
			logger.Debug().Msgf("cannot find URL for token: '%s'", key)
			http.NotFound(w, r)
			return
		}

		logger.Error().Err(err).Msg("cannot read data from database")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	if len(requestedURL) == 0 {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, requestedURL, http.StatusTemporaryRedirect)
	logger.Info().Msgf("getURL %s %s", key, requestedURL)
}

func middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		logger := zerolog.Ctx(r.Context())
		logger.Info().Msgf("Request started, URL: %s, mehtod: %s", r.URL.String(), r.Method)
		next.ServeHTTP(w, r)
		logger.Info().Msg("Request finished")
	}
	return http.HandlerFunc(fn)
}

func RequestID(header string, r *http.Request) string {
	id := r.Header.Get(header)

	if id == "" {
		id = xid.New().String()
	}

	return id
}

func RequestIDHandler(fieldKey, headerName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := zerolog.Ctx(ctx)

			id := RequestID(headerName, r)

			ctx = context.WithValue(ctx, "requestID", id)
			r = r.WithContext(ctx)

			logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str(fieldKey, id)
			})

			w.Header().Set(headerName, id)

			next.ServeHTTP(w, r)
		})
	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	data, err := httputil.DumpRequest(r, true)
	if err != nil {
		zerolog.Ctx(r.Context()).Error().Err(err).Msg("cannot dump request")
	}

	w.Write(data)
}