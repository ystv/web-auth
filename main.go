package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/ystv/web-auth/views"
	"log"
	"os"
	"strconv"
)

//go:generate ./node_modules/.bin/mjml -r ./templates/mjml/forgotEmail.mjml -o ./templates/forgotEmail.tmpl
//go:generate ./node_modules/.bin/mjml -r ./templates/mjml/resetEmail.mjml -o ./templates/resetEmail.tmpl

var Version = "unknown"

func main() {
	var local, global bool
	var err error
	err = godotenv.Load(".env") // Load .env
	if err != nil {
		global = false
	} else {
		global = true
	}

	err = godotenv.Overload(".env.local") // Load .env.local
	if err != nil {
		local = false
	} else {
		local = true
	}

	signingKey := os.Getenv("WAUTH_SIGNING_KEY")
	dbHost := os.Getenv("WAUTH_DB_HOST")

	if !local && !global && signingKey == "" && dbHost == "" {
		log.Fatal("unable to find env files and no env variables have been supplied")
	} else if !local && !global {
		log.Println("using env variables")
	} else if local && global {
		log.Println("using global and local env files")
	} else if !local {
		log.Println("using global env file")
	} else {
		log.Println("using local env file")
	}

	sessionCookieName := os.Getenv("WAUTH_SESSION_COOKIE_NAME")
	if sessionCookieName == "" {
		sessionCookieName = "session"
	}

	dbConnectionString := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		dbHost,
		os.Getenv("WAUTH_DB_PORT"),
		os.Getenv("WAUTH_DB_USER"),
		os.Getenv("WAUTH_DB_NAME"),
		os.Getenv("WAUTH_DB_SSLMODE"),
		os.Getenv("WAUTH_DB_PASS"),
	)

	log.Printf("web-auth version: %s\n", Version)

	mailPort, _ := strconv.Atoi(os.Getenv("WAUTH_MAIL_PORT"))

	debug, err := strconv.ParseBool(os.Getenv("WAUTH_DEBUG"))
	if err != nil {
		debug = false
	}

	if debug {
		fmt.Println()
		log.Println("running in debug mode, do not use in production")
		fmt.Println()
	}

	address := os.Getenv("WAUTH_ADDRESS")

	// Generate config
	conf := &views.Config{
		Version:           Version,
		Debug:             debug,
		Address:           address,
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

	v := views.New(conf, dbHost)

	router := NewRouter(&RouterConf{
		Config: conf,
		Views:  v,
	})

	err = router.Start()
	if err != nil {
		log.Fatalf("The web server couldn't be started!\n\n%s\n\nExiting!", err)
	}
}
