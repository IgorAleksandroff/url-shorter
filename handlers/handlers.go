package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rs/zerolog"
)

type Database interface {
	Set(ctx context.Context, key, value string) error
	Get(ctx context.Context, key string) (string, error)
}

type Shortner struct {
	db      Database
	newUUID func() string
}

func New(db Database, uuidFunc func() string) Shortner {
	return Shortner{
		db:      db,
		newUUID: uuidFunc,
	}
}

// Example: www.example.com/some-path?url=https://www.ozon.ru
func (s Shortner) SaveURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := zerolog.Ctx(ctx)

	v := r.URL.Query().Get("url")

	parsedURL, err := url.Parse(v)
	if err != nil {
		logger.Info().Err(err).Msgf("cannot parse URL: %q", v)
		msg := fmt.Sprintf("некорректный URL: %q", v)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
		return
	}

	key := s.newUUID()
	// Save URL into database
	if err := s.db.Set(ctx, key, parsedURL.String()); err != nil {
		logger.Error().Err(err).Msg("cannot insert into database")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	msg := fmt.Sprintf("короткий URL: %s", key)
	if _, err := w.Write([]byte(msg)); err != nil {
		logger.Error().Err(err).Msg("SaveURL: writing response")
	}
}

// Exmaple: www.example.com/go?to=uuid
func (s Shortner) GetURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := zerolog.Ctx(ctx)

	key := r.URL.Query().Get("to")

	// SQL DB
	strURL, err := s.db.Get(ctx, key)
	if err != nil {
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

	http.Redirect(w, r, strURL, http.StatusTemporaryRedirect)
}

func Help(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Для получения короткого URL, необходимо добавить к адресу: '/save?url=https://www.ozon.ru'")
}
