package web

import (
	"bytes"
	"context"
	"encoding/json"
	"go-rest-api/api/model"
	"net/http"
	"net/url"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

var (
	timeLayout = time.RFC3339
	dateLayout = "2006-01-02"
)

type API struct {
	Repo   model.Repository
	AppURL *url.URL
}

// NewAPI API実装を初期化して生成します
func NewAPI(repo model.Repository, appURL *url.URL) (*API, error) {
	api := &API{
		Repo:   repo,
		AppURL: appURL,
	}
	return api, nil
}

func RenderJSON(w http.ResponseWriter, status int, v interface{}) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(buf.Bytes())
}

func RenderNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func VersionHeader(ver string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Version", ver)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func (api *API) RequireQueryKey(key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			query := r.URL.Query().Get(key)
			if query == "" {
				// panic(ErrRequiredQueryNotFound)
			}
			ctx = withQuery(ctx, key, query)

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func (api *API) OptionalQueryKey(key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			query := r.URL.Query().Get(key)
			ctx = withQuery(ctx, key, query)

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

type queryKey string

func withQuery(ctx context.Context, key, query string) context.Context {
	return context.WithValue(ctx, queryKey(key), query)
}

func queryFrom(ctx context.Context, key string) (string, bool) {
	query, ok := ctx.Value(queryKey(key)).(string)
	return query, ok
}
