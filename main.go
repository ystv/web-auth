package main

import (
	"log"
	"net/http"

	"github.com/rs/cors"
	"github.com/ystv/web-auth/views"
)

func allow(origin string) bool {
	return true
}

func main() {
	views.New()
	mux := http.NewServeMux()
	// Static
	fs := http.FileServer(http.Dir("./public/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Login logout
	mux.HandleFunc("/login", views.LoginFunc)
	mux.HandleFunc("/logout", views.LogoutFunc)
	mux.HandleFunc("/signup", views.SignUpFunc)
	mux.HandleFunc("/forgot", views.ForgotFunc)
	mux.HandleFunc("/reset", views.ResetFunc)

	// CORS

	c := cors.New(cors.Options{
		AllowOriginFunc:  allow,
		AllowCredentials: true,
	})

	// API
	// Sets a cookie with the JWT inside of it
	mux.Handle("/api/set_token", c.Handler(views.RequiresLogin(http.HandlerFunc(views.SetTokenHandler))))
	mux.HandleFunc("/api/test", views.TestAPI)

	// Login required
	mux.HandleFunc("/internal/users", views.RequiresLogin(http.HandlerFunc(views.UsersFunc)))
	mux.HandleFunc("/internal/user", views.RequiresLogin(http.HandlerFunc(views.UserFunc)))
	mux.HandleFunc("/internal", views.RequiresLogin(http.HandlerFunc(views.InternalFunc)))

	// Public
	mux.HandleFunc("/", views.IndexFunc)

	// Serve
	log.Fatal(http.ListenAndServe(":8080", mux))
}
