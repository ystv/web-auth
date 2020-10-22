package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/ystv/web-auth/views"
)

func allow(origin string) bool {
	return true
}

func main() {
	// Generate config
	conf := views.Config{
		DatabaseURL:    os.Getenv("WAUTH_DATABASE_URL"),
		DomainName:     os.Getenv("WAUTH_DOMAIN_NAME"),
		LogoutEndpoint: os.Getenv("WAUTH_LOGOUT_ENDPOINT"),
		Mail: views.SMTPConfig{
			Host:     os.Getenv("WAUTH_SMTP_HOST"),
			Username: os.Getenv("WAUTH_SMTP_USERNAME"),
			Password: os.Getenv("WAUTH_SMTP_PASSWORD"),
		},
		Security: views.SecurityConfig{
			EncryptionKey:     os.Getenv("WAUTH_ENCRYPTION_KEY"),
			AuthenticationKey: os.Getenv("WAUTH_AUTHENTICATION_KEY"),
			SigningKey:        os.Getenv("WAUTH_SIGNING_KEY"),
		},
	}
	v := views.New(conf)
	mux := mux.NewRouter()
	// Static
	fs := http.FileServer(http.Dir("./public/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Login logout
	mux.HandleFunc("/login", v.LoginFunc)
	mux.HandleFunc("/logout", v.LogoutFunc)
	mux.HandleFunc("/signup", v.SignUpFunc)
	mux.HandleFunc("/forgot", v.ForgotFunc)
	mux.HandleFunc("/reset", v.ResetFunc)

	// CORS

	c := cors.New(cors.Options{
		AllowOriginFunc:  allow,
		AllowCredentials: true,
	})

	// API
	// Sets a cookie with the JWT inside of it
	mux.Handle("/api/set_token", c.Handler(v.RequiresLogin(http.HandlerFunc(v.SetTokenHandler))))
	mux.HandleFunc("/api/test", v.TestAPI)

	// Login required
	mux.HandleFunc("/internal/users", v.RequiresLogin(http.HandlerFunc(v.UsersFunc)))
	mux.HandleFunc("/internal/user/{userid}", v.RequiresLogin(http.HandlerFunc(v.UserFunc)))
	mux.HandleFunc("/internal", v.RequiresLogin(http.HandlerFunc(v.InternalFunc)))

	// Public
	mux.HandleFunc("/", v.IndexFunc)

	// Serve
	log.Fatal(http.ListenAndServe(":8080", mux))
}
