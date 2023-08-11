package main

import (
	"fmt"
	"github.com/ystv/web-auth/infrastructure/mail"

	"github.com/joho/godotenv"
	"github.com/ystv/web-auth/routes"
	"github.com/ystv/web-auth/views"
	"log"
	"os"
	"strconv"
)

var Version = "unknown"

func main() {
	var local, global bool
	var err error
	// Load environment
	err = godotenv.Load(".env")
	if err != nil {
		global = false
	} else {
		global = true
	}
	// Load .env file for production
	err = godotenv.Overload(".env.local") // Load .env.local for developing
	if err != nil {
		local = false
	} else {
		global = false
		local = true
	}

	signingKey := os.Getenv("WAUTH_SIGNING_KEY")
	dbHost := os.Getenv("WAUTH_DB_HOST")

	if !local && !global && signingKey == "" && dbHost == "" {
		log.Fatal("unable to find env files and no env variables have been supplied")
	} else if !local && !global {
		log.Println("using env variables")
	} else if !local {
		log.Println("using global env file")
	} else {
		log.Println("using local env file")
	}

	// Validate the required config is set
	if signingKey == "" {
		log.Fatalf("signing key not set")
	}

	if dbHost == "" {
		log.Fatalf("database host not set")
	}

	sessionCookieName := os.Getenv("WAUTH_SESSION_COOKIE_NAME")
	if sessionCookieName == "" {
		sessionCookieName = "session"
	}

	dbPort := os.Getenv("WAUTH_DB_PORT")

	dbConnectionString := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		dbHost,
		dbPort,
		os.Getenv("WAUTH_DB_USER"),
		os.Getenv("WAUTH_DB_NAME"),
		os.Getenv("WAUTH_DB_SSLMODE"),
		os.Getenv("WAUTH_DB_PASS"),
	)

	log.Printf("web-auth version: %s\n", Version)

	mailPort, err := strconv.Atoi(os.Getenv("WAUTH_MAIL_PORT"))
	if err != nil {
		log.Fatalf("failed to get port for mailer: %v", err)
	}

	debug, err := strconv.ParseBool(os.Getenv("WAUTH_DEBUG"))
	if err != nil {
		log.Printf("failed to get WAUTH_DEBUG, defaulting to false: %v", err)
		debug = false
	}

	if debug {
		log.Println("------running in debug mode, do not use in production------")
	}

	port := os.Getenv("WAUTH_PORT")

	// Generate config
	conf := &views.Config{
		Version:           Version,
		Debug:             debug,
		Port:              port,
		DatabaseURL:       dbConnectionString,
		BaseDomainName:    os.Getenv("WAUTH_BASE_DOMAIN_NAME"),
		DomainName:        os.Getenv("WAUTH_DOMAIN_NAME"),
		LogoutEndpoint:    os.Getenv("WAUTH_LOGOUT_ENDPOINT"),
		SessionCookieName: sessionCookieName,
		Mail: views.SMTPConfig{
			Host:     os.Getenv("WAUTH_MAIL_HOST"),
			Username: os.Getenv("WAUTH_MAIL_USER"),
			Password: os.Getenv("WAUTH_MAIL_PASS"),
			Port:     mailPort,
		},
		Security: views.SecurityConfig{
			EncryptionKey:     os.Getenv("WAUTH_ENCRYPTION_KEY"),
			AuthenticationKey: os.Getenv("WAUTH_AUTHENTICATION_KEY"),
			SigningKey:        signingKey,
		},
	}

	v := views.New(conf, dbHost, dbPort)

	router1 := router.New(&router.NewRouter{
		Config: conf,
		Views:  v,
	})

	err = router1.Start()
	if err != nil {
		err1 := v.Mailer.SendErrorFatalMail(mail.Mail{
			Error:       fmt.Errorf("the web server couldn't be started: %s... exiting", err),
			UseDefaults: true,
		})
		if err1 != nil {
			fmt.Println(err1)
		}
		log.Fatalf("The web server couldn't be started!\n\n%s\n\nExiting!", err)
	}
}
