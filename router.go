package main

import (
	"go-rest-api/api/web"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
)

func NewRouter(api *web.API) *chi.Mux {

	// 	e := echo.New()

	// 	// Debug mode
	// 	e.Debug = true

	// 	// Middleware TODO:auth
	// 	e.Use(middleware.Logger())
	// 	e.Use(middleware.Recover())
	// 	// e.Use(web.WithSessionUser)

	// 	// Health check
	// 	e.GET("/health", healthCheckHandler)

	// 	// Routes
	// 	// TODO: versioning
	// 	e.GET("v1/users", api.GetAllUser)
	// 	e.POST("v1/users", api.CreateUser)
	// 	e.GET("v1/users/:id", api.GetUser)
	// 	e.PATCH("v1/users/:id", api.UpdateUser)
	// 	e.DELETE("v1/users/:id", api.DeleteUser)

	// 	return e
	// }

	// func healthCheckHandler(ctx echo.Context) error {
	// 	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)

	// 	return ctx.String(http.StatusOK, "{alive:true}")

	r := chi.NewRouter()

	middleware.DefaultLogger = middleware.RequestLogger(
		&middleware.DefaultLogFormatter{
			Logger: newLogger(),
		},
	)

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(web.PanicHandler)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	// Basic CORS
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// healthcheck
	r.Get("/", healthCheckHandler)

	r.Route("/v1", func(r chi.Router) {
		r.Use(web.VersionHeader("1"))

		r.Route("/users", func(r chi.Router) {
			r.Get("/", api.GetAllUser)
			r.Post("/", api.CreateUser)
			r.Route("/{id}", func(r chi.Router) {
				r.Use(api.RequireUserID)
				r.Patch("/", api.UpdateUser)
				r.Delete("/", api.DeleteUser)
			})
		})
	})
	return r
}

// paginate is a stub, but very possible to implement middleware logic
// to handle the request params for handling a paginated request.
func paginate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// just a stub.. some ideas are to look at URL query params for something like
		// the page number, or the limit, and send a query cursor down the chain
		next.ServeHTTP(w, r)
	})
}

func newLogger() *log.Logger {
	return log.New(os.Stdout, "chi-log: ", log.Lshortfile)
}

// healthCheckHandler ...
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// simple health check
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	io.WriteString(w, `{"alive": true}`)
}
