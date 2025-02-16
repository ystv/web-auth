package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"github.com/ystv/web-auth/utils"
	"github.com/ystv/web-auth/views"
)

//go:generate ./node_modules/.bin/mjml -r ./templates/mjml/forgotEmail.mjml -o ./templates/forgotEmail.tmpl
//go:generate ./node_modules/.bin/mjml -r ./templates/mjml/resetEmail.mjml -o ./templates/resetEmail.tmpl

var (
	Version = "unknown"
	Commit  = "unknown"
)

func main() {
	var local, global bool

	var err error
	err = godotenv.Load(".env") // Load .env
	global = err == nil

	err = godotenv.Overload(".env.local") // Load .env.local
	local = err == nil

	signingKey := os.Getenv("WAUTH_SIGNING_KEY")
	dbHost := os.Getenv("WAUTH_DB_HOST")

	if !local && !global && signingKey == "" && dbHost == "" {
		log.Fatal("unable to find env files and no env variables have been supplied")
	}
	//nolint:gocritic
	if !local && !global {
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

	mailPort, _ := strconv.Atoi(os.Getenv("WAUTH_MAIL_PORT"))

	debug, _ := strconv.ParseBool(os.Getenv("WAUTH_DEBUG"))

	if debug {
		log.Println("***running in debug mode, do not use in production***")
	}

	address := os.Getenv("WAUTH_ADDRESS")

	domainName := os.Getenv("WAUTH_DOMAIN_NAME")

	// CDN
	cdnConfig := utils.CDNConfig{
		Endpoint:        os.Getenv("WAUTH_CDN_ENDPOINT"),
		Region:          os.Getenv("WAUTH_CDN_REGION"),
		AccessKeyID:     os.Getenv("WAUTH_CDN_ACCESSKEYID"),
		SecretAccessKey: os.Getenv("WAUTH_CDN_SECRETACCESSKEY"),
	}
	cdn, err := utils.NewCDN(cdnConfig)
	if err != nil {
		log.Fatalf("Unable to connect to CDN: %v", err)
	}
	log.Printf("Connected to CDN: %s", cdnConfig.Endpoint)

	adPort, err := strconv.Atoi(os.Getenv("WAUTH_AD_PORT"))
	if err != nil {
		log.Fatalf("failed to get ad port env: %+v", err)
	}

	adSecurity, err := strconv.Atoi(os.Getenv("WAUTH_AD_SECURITY"))
	if err != nil {
		log.Fatalf("failed to get ad security env: %+v", err)
	}

	// Generate config
	conf := &views.Config{
		Version:           Version,
		Commit:            Commit,
		Debug:             debug,
		Address:           address,
		DatabaseURL:       dbConnectionString,
		BaseDomainName:    os.Getenv("WAUTH_BASE_DOMAIN_NAME"),
		DomainName:        domainName,
		LogoutEndpoint:    os.Getenv("WAUTH_LOGOUT_ENDPOINT"),
		SessionCookieName: sessionCookieName,
		CDNEndpoint:       os.Getenv("WAUTH_CDN_ENDPOINT"),
		Mail: views.SMTPConfig{
			Host:       os.Getenv("WAUTH_MAIL_HOST"),
			Username:   os.Getenv("WAUTH_MAIL_USER"),
			Password:   os.Getenv("WAUTH_MAIL_PASS"),
			Port:       mailPort,
			DomainName: domainName,
		},
		Security: views.SecurityConfig{
			EncryptionKey:     os.Getenv("WAUTH_ENCRYPTION_KEY"),
			AuthenticationKey: os.Getenv("WAUTH_AUTHENTICATION_KEY"),
			SigningKey:        signingKey,
		},
		AD: views.ADConfig{
			Server:   os.Getenv("WAUTH_AD_SERVER"),
			Port:     adPort,
			BaseDN:   os.Getenv("WAUTH_AD_BASE_DN"),
			Security: adSecurity,
			Bind: views.ADBind{
				Username: os.Getenv("WAUTH_AD_BIND_USERNAME"),
				Password: os.Getenv("WAUTH_AD_BIND_PASSWORD"),
			},
		},
	}

	v := views.New(conf, dbHost, cdn)

	router := NewRouter(&RouterConf{
		Config: conf,
		Views:  v,
	})

	//nolint:staticcheck
	err = router.Start()
	//nolint:staticcheck
	if err != nil {
		log.Fatalf("The web server couldn't be started!\n\n%s\n\nExiting!", err)
	}
}
