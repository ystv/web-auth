package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	_ "golang.org/x/crypto/x509roots/fallback" // CA bundle for FROM Scratch

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

	//nolint:reassign
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	logger := utils.NewLogger(zlog.With().
		Str("service", "web-auth").
		Str("version", Version).
		Str("commit", Commit).
		Logger(), utils.DefaultSkipper)

	var err error
	err = godotenv.Load(".env") // Load .env
	global = err == nil

	err = godotenv.Overload(".env.local") // Load .env.local
	local = err == nil

	signingKey := os.Getenv("WAUTH_SIGNING_KEY")
	dbHost := os.Getenv("WAUTH_DB_HOST")

	if !local && !global && signingKey == "" && dbHost == "" {
		logger.Fatal(nil, errors.New("unable to find env files and no env variables have been supplied"))
	}
	//nolint:gocritic
	if !local && !global {
		logger.Debug(nil, "using env variables")
	} else if local && global {
		logger.Debug(nil, "using global and local env files")
	} else if !local {
		logger.Debug(nil, "using global env file")
	} else {
		logger.Debug(nil, "using local env file")
	}

	sessionCookieName := os.Getenv("WAUTH_SESSION_COOKIE_NAME")
	if sessionCookieName == "" {
		sessionCookieName = "session"
	}

	jwtCookieName := os.Getenv("WAUTH_JWT_COOKIE_NAME")
	if jwtCookieName == "" {
		jwtCookieName = "session_jwt"
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

	logger.Info(nil, "web-auth version: %s", Version)

	mailPort, _ := strconv.Atoi(os.Getenv("WAUTH_MAIL_PORT"))

	debug, _ := strconv.ParseBool(os.Getenv("WAUTH_DEBUG"))

	if debug {
		kv := []utils.KVAny{
			{Key: "dbHost", Any: dbHost},
			{Key: "dbPort", Any: dbPort},
			{Key: "sessionCookieName", Any: sessionCookieName},
			{Key: "jwtCookieName", Any: jwtCookieName},
			{Key: "mailPort", Any: mailPort},
			{Key: "debug", Any: debug},
		}
		logger.Warn(&kv, "***running in debug mode, do not use in production***")
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
		logger.Fatal(nil, fmt.Errorf("unable to connect to cdn: %w", err))
	}
	logger.Debug(nil, "connected to cdn: %s", cdnConfig.Endpoint)

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
		JWTCookieName:     jwtCookieName,
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
		Logger: logger,
	}

	v := views.New(conf, dbHost, cdn)

	router := NewRouter(&RouterConf{
		Config: conf,
		Views:  v,
	})

	//nolint:staticcheck
	err = router.Start()
	if err != nil {
		logger.Fatal(nil, err)
	}
}
