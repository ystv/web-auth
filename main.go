package main

import (
	"log"
	"net/http"

	"github.com/rmil/web-auth/views"
)

func main() {
	views.PopulateTemplates()

	// Static
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Login logout
	http.HandleFunc("/login/", views.LoginFunc)
	http.HandleFunc("/logout/", views.LogoutFunc)
	http.HandleFunc("/signup/", views.SignUpFunc)

	// API
	// Returns the JWT directly
	http.HandleFunc("/api/get_token", views.GetTokenHandler)
	// Sets a cookie with the JWT inside of it
	http.HandleFunc("/api/set_token", views.SetTokenHandler)
	http.HandleFunc("/api/refresh_token", views.RefreshHandler)
	http.HandleFunc("/api/test", views.TestAPI)

	// Login required
	http.HandleFunc("/internal/", views.RequiresLogin(views.InternalFunc))

	// Public
	http.HandleFunc("/", views.WelcomeFunc)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
