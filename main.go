package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/ystv/web-auth/infrastructure/mail"
	"github.com/ystv/web-auth/routes"
	"github.com/ystv/web-auth/views"
	"html/template"
	"log"
	"os"
	"os/signal"
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
			Host:          os.Getenv("WAUTH_MAIL_HOST"),
			Username:      os.Getenv("WAUTH_MAIL_USER"),
			Password:      os.Getenv("WAUTH_MAIL_PASS"),
			Port:          mailPort,
			DefaultMailTo: os.Getenv("WAUTH_MAIL_DEFAULT_TO"),
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

	startingTemplate := template.New("Startup email")
	startingTemplate = template.Must(startingTemplate.Parse("<html><body>YSTV Web Auth starting{{if .Debug}} in debug mode!<br><b>Do not run in production! Authentication is disabled!</b>{{else}}!{{end}}<br><br><br><br>If you don't get another email then this has started correctly.<br><br>Version: {{.Version}}</body></html>"))

	subject := "YSTV Web Auth is starting"

	if conf.Debug {
		subject += " in debug mode"
		log.Println("Debug Mode - Disabled auth - do not run in production!")
	}

	subject += "!"

	starting := mail.Mail{
		Subject:     subject,
		UseDefaults: true,
		Tpl:         startingTemplate,
		TplData: struct {
			Debug   bool
			Version string
		}{
			Debug:   conf.Debug,
			Version: Version,
		},
	}

	err = v.Mailer.SendMail(starting)
	if err != nil {
		log.Printf("Unable to send email: %+v", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			exitingTemplate := template.New("Exiting Template")
			exitingTemplate = template.Must(exitingTemplate.Parse("<body>YSTV Web Auth has been stopped!<br><br>{{if .Debug}}Exit signal: {{.Sig}}<br><br>{{end}}<br><br>Version: {{.Version}}</body>"))

			stopped := mail.Mail{
				Subject:     "YSTV Web Auth has been stopped!",
				UseDefaults: true,
				Tpl:         exitingTemplate,
				TplData: struct {
					Debug   bool
					Sig     os.Signal
					Version string
				}{
					Debug:   conf.Debug,
					Sig:     sig,
					Version: Version,
				},
			}

			_ = v.Mailer.Close()
			v.Mailer, err = mail.NewMailer(mail.Config{
				Host:       conf.Mail.Host,
				Port:       conf.Mail.Port,
				Username:   conf.Mail.Username,
				Password:   conf.Mail.Password,
				DomainName: conf.DomainName,
			})
			if err != nil {
				log.Printf("Mailer failed: %+v", err)
			}

			v.Mailer.AddDefaults(mail.Defaults{
				DefaultTo:   conf.Mail.DefaultMailTo,
				DefaultFrom: "YSTV Web Auth <wauth@ystv.co.uk>",
			})

			err = v.Mailer.SendMail(stopped)
			if err != nil {
				log.Printf("send fatal email error: %+v", err)
			}
			_ = v.Mailer.Close()
			os.Exit(0)
		}
	}()

	err = router1.Start()
	if err != nil {
		_ = v.Mailer.Close()
		var err1 error
		v.Mailer, err1 = mail.NewMailer(mail.Config{
			Host:       conf.Mail.Host,
			Port:       conf.Mail.Port,
			Username:   conf.Mail.Username,
			Password:   conf.Mail.Password,
			DomainName: conf.DomainName,
		})
		if err1 != nil {
			log.Printf("Mailer failed: %+v", err1)
		}
		v.Mailer.AddDefaults(mail.Defaults{
			DefaultTo:   conf.Mail.DefaultMailTo,
			DefaultFrom: "YSTV Web Auth <wauth@ystv.co.uk>",
		})
		err1 = v.Mailer.SendErrorFatalMail(mail.Mail{
			UseDefaults: true,
			TplData: struct {
				Error   error
				Version string
			}{
				Error:   fmt.Errorf("the web server couldn't be started: %w... exiting", err),
				Version: Version,
			},
		})
		if err1 != nil {
			log.Printf("send fatal email error: %+v", err1)
		}
		_ = v.Mailer.Close()
		log.Fatalf("The web server couldn't be started!\n\n%s\n\nExiting!", err)
	}
}
