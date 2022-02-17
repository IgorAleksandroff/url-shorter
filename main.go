package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/caarlos0/env/v6"
	"github.com/google/uuid"
)

type config struct {
	Port int `env:"HTTP_PORT" envDefault:"8000"`
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
}

func main() {

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		panic(fmt.Sprintf("cannot parse configuration: %s", err.Error()))
	}

	h := handler{db: db{}}

	http.HandleFunc("/save", h.saveURL)
	http.HandleFunc("/go", h.getURL)

	log.Printf("start HTTP at a port %d", cfg.Port)
	panic(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil))

}

func (h handler) saveURL(w http.ResponseWriter, r *http.Request) {
	// /save?url=https://www.ozon.ru
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
	} else {
		msg = fmt.Sprintf("короткий URL: %q", key)
		w.Write([]byte(msg))
	}

	h.db[key] = parsedURL
	log.Println("saveURL", key, r.URL.String())

}

func (h handler) getURL(w http.ResponseWriter, r *http.Request) {
	// /go?to=uuid
	key := r.URL.Query().Get("to")
	requestedURL := h.db.Get(key)
	if requestedURL == nil {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, requestedURL.String(), http.StatusTemporaryRedirect)
	log.Println("getURL", key, requestedURL.String())
}