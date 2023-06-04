package main

import (
	"embed"
	json2 "encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/ystv/web-auth/permission/permissions"
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

	mux1.HandleFunc("/internal/permissions", v.RequiresLogin(v.RequiresMinimumPermission(http.HandlerFunc(v.PermissionsFunc), permissions.ManageMembersPermissions)))
	//mux1.HandleFunc("/internal/permission/add", v.RequiresLogin(v.RequiresMinimumPermission(http.HandlerFunc(v.PermissionAddFunc), permissions.ManageMembersPermissions)))
	//mux1.HandleFunc("/internal/permission/{permissionid}/edit", v.RequiresLogin(v.RequiresMinimumPermission(http.HandlerFunc(v.PermissionEditFunc), permissions.ManageMembersPermissions)))
	//mux1.HandleFunc("/internal/permission/{permissionid}/delete", v.RequiresLogin(v.RequiresMinimumPermission(http.HandlerFunc(v.PermissionDeleteFunc), permissions.ManageMembersPermissions)))
	//mux1.HandleFunc("/internal/permission/{permissionid}", v.RequiresLogin(v.RequiresMinimumPermission(http.HandlerFunc(v.PermissionFunc), permissions.ManageMembersPermissions)))

	mux1.HandleFunc("/internal/roles", v.RequiresLogin(v.RequiresMinimumPermission(http.HandlerFunc(v.RolesFunc), permissions.ManageMembersGroup)))
	//mux1.HandleFunc("/internal/role/add", v.RequiresLogin(v.RequiresMinimumPermission(http.HandlerFunc(v.RoleAddFunc), permissions.ManageMembersGroup)))
	//mux1.HandleFunc("/internal/role/{roleid}/edit", v.RequiresLogin(v.RequiresMinimumPermission(http.HandlerFunc(v.RoleEditFunc), permissions.ManageMembersGroup)))
	//mux1.HandleFunc("/internal/role/{roleid}/delete", v.RequiresLogin(v.RequiresMinimumPermission(http.HandlerFunc(v.RoleDeleteFunc), permissions.ManageMembersGroup)))
	//mux1.HandleFunc("/internal/role/{roleid}", v.RequiresLogin(v.RequiresMinimumPermission(http.HandlerFunc(v.RoleFunc), permissions.ManageMembersGroup)))

	mux1.HandleFunc("/internal/users", v.RequiresLogin(v.RequiresMinimumPermission(http.HandlerFunc(v.UsersFunc), permissions.ManageMembersMembersList)))
	//mux1.HandleFunc("/internal/user/add", v.RequiresLogin(v.RequiresMinimumPermission(http.HandlerFunc(v.UserAddFunc), permissions.ManageMembersMembersAdd)))
	//mux1.HandleFunc("/internal/user/{userid}/edit", v.RequiresLogin(v.RequiresMinimumPermission(http.HandlerFunc(v.UserEditFunc), permissions.ManageMembersMembersAdmin)))
	//mux1.HandleFunc("/internal/user/{userid}/delete", v.RequiresLogin(v.RequiresMinimumPermission(http.HandlerFunc(v.UserDeleteFunc), permissions.ManageMembersMembersAdmin)))
	mux1.HandleFunc("/internal/user/{userid}", v.RequiresLogin(v.RequiresMinimumPermission(http.HandlerFunc(v.UserFunc), permissions.ManageMembersMembersAdmin)))

	// Public
	mux1.HandleFunc("/", v.IndexFunc)

	mux1.NotFoundHandler = http.HandlerFunc(v.Error404)

	log.Println("web-auth started, listing on :8080")
	// Serve
	log.Fatal(http.ListenAndServe(":8080", mux1))
}
