package main

import (
	"log"
	"net/http"

	"github.com/ystv/web-auth/views"
)

func main() {
	views.New()
	// Static
	fs := http.FileServer(http.Dir("./public/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Login logout
	http.HandleFunc("/login/", views.LoginFunc)
	http.HandleFunc("/logout/", views.LogoutFunc)
	http.HandleFunc("/signup/", views.SignUpFunc)
	http.HandleFunc("/forgot/", views.ForgotFunc)
	http.HandleFunc("/reset/", views.ResetFunc)

	// API
	// Sets a cookie with the JWT inside of it
	http.HandleFunc("/api/set_token", views.RequiresLogin(views.SetTokenHandler))
	http.HandleFunc("/api/test", views.TestAPI)

	// Login required
	http.HandleFunc("/internal/", views.RequiresLogin(views.InternalFunc))

	// Public
	http.HandleFunc("/", views.IndexFunc)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
