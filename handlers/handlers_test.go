package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"url-shorter/pkg"
)

type mockShortener struct {
	SetFn func(ctx context.Context, key, value string) error
	GetFn func(ctx context.Context, key string) (string, error)
}

func (m mockShortener) Set(ctx context.Context, key, value string) error {
	return m.SetFn(ctx, key, value)
}

func (m mockShortener) Get(ctx context.Context, key string) (string, error) {
	return m.GetFn(ctx, key)
}

func TestShortenerGet(t *testing.T) {
	expectedKey := "foo"

	mock := mockShortener{
		GetFn: func(ctx context.Context, key string) (string, error) {
			if key != expectedKey {
				t.Errorf("expecting key to be %q, got %q instead", expectedKey, key)
			}
			return "https://tech.ozon.ru", nil
		},
	}

	s := New(mock, pkg.GeneratorShortURL)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/"+expectedKey, nil)

	s.GetURL(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("expecting status code %d, got %d",
			http.StatusTemporaryRedirect,
			resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	expectedBody := "<a href=\"https://tech.ozon.ru\">Temporary Redirect</a>.\n\n"
	strBody := string(body)
	if strBody != expectedBody {
		t.Errorf("expecting body to be:\n%q, got\n%q", expectedBody, strBody)
	}
}

func TestShortenerGetDatabaseError(t *testing.T) {
	mock := mockShortener{
		GetFn: func(ctx context.Context, key string) (string, error) {
			return "", errors.New("something went wrong")
		},
	}

	s := New(mock, pkg.GeneratorShortURL)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/foo", nil)

	s.GetURL(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expecting status code %d, got %d",
			http.StatusInternalServerError,
			resp.StatusCode)
	}
}

func TestShortenerSaveURL(t *testing.T) {
	expecteURL := "https://tech.ozon.ru"
	expectedKey := "bar"

	mock := mockShortener{
		SetFn: func(ctx context.Context, key, value string) error {
			if got, expect := value, expecteURL; got != expect {
				t.Errorf("expecting value to be %q, got %q", expect, got)
			}

			if got, expect := key, expectedKey; got != expect {
				t.Errorf("expecting key to be %q, got %q", expect, got)
			}

			return nil
		},
	}

	s := New(mock, func() string { return expectedKey })

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/save?url="+expecteURL, nil)

	s.SaveURL(w, r)

	resp := w.Result()

	if got, expect := resp.StatusCode, http.StatusOK; got != expect {
		t.Errorf("expecting status code %d, got %d", expect, got)
	}

	body, _ := io.ReadAll(resp.Body)

	expectedBody := "короткий URL: " + expectedKey
	strBody := string(body)
	if strBody != expectedBody {
		t.Errorf("expecting body to be:\n%q, got\n%q", expectedBody, strBody)
	}
}
