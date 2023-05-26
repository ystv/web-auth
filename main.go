package main

import (
	"embed"
	json2 "encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/views"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
)

//go:embed public/static/*
var content embed.FS

var Version = "unknown"

func allow(origin string) bool {
	return true
}

func main() {
	// Load environment
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("failed to load global env file")
	} // Load .env file for production
	err = godotenv.Overload(".env.local") // Load .env.local for developing
	if err != nil {
		log.Println("failed to load env file, using global env")
	}

	// Validate the required config is set
	if os.Getenv("WAUTH_SIGNING_KEY") == "" {
		log.Fatalf("signing key not set")
	}

	if os.Getenv("WAUTH_DB_HOST") == "" {
		log.Fatalf("database host not set")
	}

	sessionCookieName := os.Getenv("WAUTH_SESSION_COOKIE_NAME")
	if sessionCookieName == "" {
		sessionCookieName = "session"
	}

	dbConnectionString := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		os.Getenv("WAUTH_DB_HOST"),
		os.Getenv("WAUTH_DB_PORT"),
		os.Getenv("WAUTH_DB_USER"),
		os.Getenv("WAUTH_DB_NAME"),
		os.Getenv("WAUTH_DB_SSLMODE"),
		os.Getenv("WAUTH_DB_PASS"),
	)

	// Setup files
	static, err := fs.Sub(content, "public/static")
	if err != nil {
		log.Fatalf("static files failed: %+v", err)
	}

	log.Printf("web-auth version %s loaded", Version)

	port, err := strconv.Atoi(os.Getenv("WAUTH_MAIL_PORT"))
	if err != nil {
		log.Fatalf("failed to get port for mailer: %v", err)
	}

	// Generate config
	conf := views.Config{
		Version:           Version,
		DatabaseURL:       dbConnectionString,
		BaseDomainName:    os.Getenv("WAUTH_BASE_DOMAIN_NAME"),
		DomainName:        os.Getenv("WAUTH_DOMAIN_NAME"),
		LogoutEndpoint:    os.Getenv("WAUTH_LOGOUT_ENDPOINT"),
		SessionCookieName: sessionCookieName,
		Mail: views.SMTPConfig{
			Host:     os.Getenv("WAUTH_MAIL_HOST"),
			Username: os.Getenv("WAUTH_MAIL_USER"),
			Password: os.Getenv("WAUTH_MAIL_PASS"),
			Port:     port,
		},
		Security: views.SecurityConfig{
			EncryptionKey:     os.Getenv("WAUTH_ENCRYPTION_KEY"),
			AuthenticationKey: os.Getenv("WAUTH_AUTHENTICATION_KEY"),
			SigningKey:        os.Getenv("WAUTH_SIGNING_KEY"),
		},
	}

	v := views.New(conf)
	mux1 := mux.NewRouter()
	// Static
	fs1 := http.FileServer(http.FS(static))
	mux1.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs1))

	// Login logout
	mux1.HandleFunc("/login", v.LoginFunc)
	mux1.HandleFunc("/logout", v.LogoutFunc)
	mux1.HandleFunc("/signup", v.SignUpFunc)
	mux1.HandleFunc("/forgot", v.ForgotFunc)
	//mux1.HandleFunc("/reset", v.ResetFunc)
	mux1.HandleFunc("/reset/{url}", v.ResetURLFunc)

	// CORS

	c := cors.New(cors.Options{
		AllowOriginFunc:  allow,
		AllowCredentials: true,
	})

	// API
	// Sets a cookie with the JWT inside of it
	mux1.Handle("/api/set_token", c.Handler(v.RequiresLogin(http.HandlerFunc(v.SetTokenHandler))))
	mux1.HandleFunc("/api/test", v.TestAPI)
	mux1.HandleFunc("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		marshal, err := json2.Marshal(struct {
			Status int `json:"status"`
		}{
			Status: http.StatusOK,
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(marshal)
		if err != nil {
			fmt.Println(err)
			return
		}
	})

	// Login required
	mux1.HandleFunc("/internal", v.RequiresLogin(http.HandlerFunc(v.InternalFunc)))
	mux1.HandleFunc("/internal/settings", v.RequiresLogin(http.HandlerFunc(v.SettingsFunc)))
	mux1.HandleFunc("/internal/users", v.RequiresLogin(v.RequiresPermission(http.HandlerFunc(v.UsersFunc), permission.Permission{Name: permission.SuperUser})))
	mux1.HandleFunc("/internal/roles", v.RequiresLogin(v.RequiresPermission(http.HandlerFunc(v.RolesFunc), permission.Permission{Name: permission.SuperUser})))
	mux1.HandleFunc("/internal/permissions", v.RequiresLogin(v.RequiresPermission(http.HandlerFunc(v.PermissionsFunc), permission.Permission{Name: permission.SuperUser})))
	mux1.HandleFunc("/internal/user/{userid}", v.RequiresLogin(v.RequiresPermission(http.HandlerFunc(v.UserFunc), permission.Permission{Name: permission.SuperUser})))

	// Public
	mux1.HandleFunc("/", v.IndexFunc)

	mux1.NotFoundHandler = http.HandlerFunc(v.Error404)

	log.Println("web-auth started, listing on :8080")
	// Serve
	log.Fatal(http.ListenAndServe(":8080", mux1))
}
