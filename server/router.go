package server

import (
	"bytes"
	"net/http"
	"time"

	"github.com/ptflp/go-light/email"
	"github.com/ptflp/go-light/request"

	"github.com/go-chi/cors"

	"github.com/ptflp/go-light/components"
	"github.com/ptflp/go-light/services"

	"github.com/ptflp/go-light/controllers"

	"github.com/ptflp/go-light/middlewares"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(services *services.Services, cmps components.Componenter) (*chi.Mux, error) {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	proxy := middlewares.NewReverseProxy()
	r.Use(proxy.ReverseProxy)
	// Basic CORS
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	r.Use(
		cors.Handler(
			cors.Options{
				// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
				AllowedOrigins: []string{"https://*", "http://*"},
				// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
				ExposedHeaders:   []string{"Link"},
				AllowCredentials: false,
				MaxAge:           300, // Maximum value not ignored by any of major browsers
			},
		),
	)

	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)

	authController := controllers.NewAuth(cmps.Responder(), services.AuthService, cmps.Logger())

	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {

		msg := email.NewMessage()
		var b bytes.Buffer

		b.Write([]byte("test"))

		msg.SetBody(b)
		msg.SetReceiver("globallinkliberty@gmail.com")
		msg.SetSubject("test")
		_ = msg.OpenFile(".gitignore")
		_ = msg.OpenFile(".env")
		err := cmps.Email().Send(msg)

		cmps.Responder().SendJSON(w, request.Response{
			Success: err == nil,
			Msg:     "test",
			Data:    err,
		})
	})

	token := middlewares.NewCheckToken(cmps.Responder(), cmps.JWTKeys())

	r.Get("/swagger", swaggerUI)
	r.Get("/static/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))).ServeHTTP(w, r)
	})
	r.Get("/public/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))).ServeHTTP(w, r)
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/email/registration", authController.EmailActivation())
		r.Post("/email/verification", authController.EmailVerification())
		r.Post("/email/login", authController.EmailLogin())
		r.Post("/checkemail", authController.CheckCode())

		r.Post("/token/refresh", authController.RefreshToken())

		r.Post("/code", authController.SendCode())
		r.Post("/checkcode", authController.CheckCode())
	})

	users := controllers.NewUsersController(cmps.Responder(), services.User, cmps.Logger())
	// ./docs/user.go
	r.Route("/people", func(r chi.Router) {
		r.Use(token.CheckStrict)
		r.Post("/autocomplete", users.Autocomplete())
		r.Route("/get", func(r chi.Router) {
			r.Post("/", users.Get())
			r.Post("/subscribes", users.Get())
			r.Post("/subscribers", users.Get())
		})
		r.Route("/list", func(r chi.Router) {
			r.Get("/", users.List())
			r.Post("/subscribers", users.TempList())
			r.Post("/recommends", users.Recommends())

		})
	})

	r.Route("/recover", func(r chi.Router) {
		r.Post("/password", users.RecoverPassword())
		r.Post("/check/phone", users.CheckPhoneCode())
		r.Post("/set/password", users.PasswordReset())
	})

	r.Route("/exist", func(r chi.Router) {
		r.Post("/email", users.EmailExist())
		r.Post("/nickname", users.NicknameExist())
	})
	r.Route("/system", func(r chi.Router) {
		r.Use(middleware.Timeout(200 * time.Millisecond))
		r.Use(token.CheckStrict)
		r.Get("/config", func(w http.ResponseWriter, r *http.Request) {
			cmps.Responder().SendJSON(w, cmps.Config())
		})
	})

	return r, nil
}
