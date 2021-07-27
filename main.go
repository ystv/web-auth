package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/ystv/web-auth/user"
	"github.com/ystv/web-auth/views"
)

//go:embed public/*
var content embed.FS

func allow(origin string) bool {
	return true
}

func main() {
	// Load environment
	err := godotenv.Load()            // Load .env file for production
	err = godotenv.Load(".env.local") // Load .env.local for developing
	if err != nil {
		log.Print("failed to load env file, using global env")
	}

	// Setup files
	static, err := fs.Sub(content, "public/static")
	if err != nil {
		log.Fatalf("static files failed: %+v", err)
	}
	templates, err := fs.Sub(content, "public/templates")
	if err != nil {
		log.Fatalf("template files failed: %+v", err)
	}

	version := os.Getenv("WAUTH_VERSION")
	if version == "" {
		version = "unknown"
	}
	log.Println("web-auth version " + version + " loaded")

	// Generate config
	conf := views.Config{
		Version:        version,
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

	adminPerm := user.Permission{
		ID:   19,
		Name: "SuperUser",
	}

	v := views.New(conf, templates)
	mux := mux.NewRouter()
	// Static
	fs := http.FileServer(http.FS(static))
	mux.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

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
	mux.HandleFunc("/internal/users", v.RequiresPermission(v.RequiresLogin(http.HandlerFunc(v.UsersFunc)), adminPerm))
	mux.HandleFunc("/internal/user/{userid}", v.RequiresLogin(http.HandlerFunc(v.UserFunc)))
	mux.HandleFunc("/internal", v.RequiresLogin(http.HandlerFunc(v.InternalFunc)))

	// Public
	mux.HandleFunc("/", v.IndexFunc)

	log.Println("web-auth started, listing on :8080")
	// Serve
	log.Fatal(http.ListenAndServe(":8080", mux))
}
