package views

import (
	"context"
	"encoding/gob"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/patrickmn/go-cache"

	"github.com/ystv/web-auth/api"
	"github.com/ystv/web-auth/crowd"
	"github.com/ystv/web-auth/infrastructure/db"
	"github.com/ystv/web-auth/infrastructure/mail"
	"github.com/ystv/web-auth/officership"
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/role"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
)

type (
	// Config the global web-auth configuration
	Config struct {
		Version           string
		Commit            string
		Debug             bool
		Address           string
		DatabaseURL       string
		BaseDomainName    string
		DomainName        string
		LogoutEndpoint    string
		JWTCookieName     string
		SessionCookieName string
		CDNEndpoint       string
		Mail              SMTPConfig
		Security          SecurityConfig
	}

	// SMTPConfig stores the SMTP Mailer configuration
	SMTPConfig struct {
		Host       string
		Username   string
		Password   string
		Port       int
		DomainName string
	}

	// SecurityConfig stores the security configuration
	SecurityConfig struct {
		EncryptionKey     string
		AuthenticationKey string
		SigningKey        string
	}

	// Views encapsulates our view dependencies
	Views struct {
		api         api.Repo
		cache       *cache.Cache
		cdn         *s3.S3
		conf        *Config
		cookie      *sessions.CookieStore
		crowd       crowd.Repo
		Mailer      *mail.Mailer
		officership officership.Repo
		permission  permission.Repo
		role        role.Repo
		template    *templates.Templater
		user        user.Repo
		mailer      *mail.MailerInit
		validate    *validator.Validate
	}

	TemplateHelper struct {
		UserPermissions []permission.Permission
		ActivePage      string
		Assumed         bool
	}

	XMLError struct {
		XMLName xml.Name `xml:"error"`
		Reason  string   `xml:"reason"`
		Message string   `xml:"message"`
	}
)

// New initialises connections, templates, and cookies
func New(conf *Config, host string, cdn *s3.S3) *Views {
	v := &Views{}
	// Connecting to stores
	dbStore := db.NewStore(conf.DatabaseURL, host)
	v.officership = officership.NewOfficershipRepo(dbStore)
	v.permission = permission.NewPermissionRepo(dbStore)
	v.role = role.NewRoleRepo(dbStore)
	v.user = user.NewUserRepo(dbStore, conf.CDNEndpoint)
	v.api = api.NewAPIRepo(dbStore)
	v.crowd = crowd.NewCrowdRepo(dbStore)

	v.cdn = cdn

	v.template = templates.NewTemplate(v.permission, v.role, v.user)

	// Initialising cache
	v.cache = cache.New(1*time.Hour, 1*time.Hour)

	// Initialise mailer
	v.mailer = mail.NewMailer(mail.Config{
		Host:       conf.Mail.Host,
		Port:       conf.Mail.Port,
		Username:   conf.Mail.Username,
		Password:   conf.Mail.Password,
		DomainName: conf.Mail.DomainName,
	})

	// Initialising session cookie
	authKey, _ := hex.DecodeString(conf.Security.AuthenticationKey)
	if len(authKey) == 0 {
		authKey = securecookie.GenerateRandomKey(64)
	}

	encryptionKey, _ := hex.DecodeString(conf.Security.EncryptionKey)
	if len(encryptionKey) == 0 {
		encryptionKey = securecookie.GenerateRandomKey(32)
	}

	v.cookie = sessions.NewCookieStore(
		authKey,
		encryptionKey,
	)

	sixty := 60
	twentyFour := 24

	v.cookie.Options = &sessions.Options{
		MaxAge:   sixty * sixty * twentyFour,
		HttpOnly: true,
		Domain:   "." + conf.BaseDomainName,
		Path:     "/",
	}

	// So we can use our struct in the cookie
	gob.Register(user.User{})
	gob.Register(InternalContext{})

	v.conf = conf

	// Struct validator
	v.validate = validator.New()

	go func() {
		for {
			err := v.api.DeleteOldToken(context.Background())
			if err != nil {
				log.Printf("failed to delete old token func: %+v", err)
			}

			time.Sleep(30 * time.Second)
		}
	}()

	return v
}

func (v *Views) fileUpload(file *multipart.FileHeader) (string, []byte, error) {
	var fileName, fileType string
	switch file.Header.Get("content-type") {
	case "application/pdf":
		fileType = ".pdf"
		break
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		fileType = ".docx"
		break
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		fileType = ".pptx"
		break
	case "text/plain":
		fileType = ".txt"
		break
	case "image/apng":
		fileType = ".apng"
		break
	case "image/avif":
		fileType = ".avif"
		break
	case "image/gif":
		fileType = ".gif"
		break
	case "image/jpeg":
		fileType = ".jpg"
		break
	case "image/png":
		fileType = ".png"
		break
	case "image/svg+xml":
		fileType = ".svg"
		break
	case "image/webp":
		fileType = ".webp"
		break
	default:
		return "", []byte{}, fmt.Errorf("invalid file type: %s", file.Header.Get("content-type"))
	}

	fileName = uuid.NewString() + fileType

	src, err := file.Open()
	if err != nil {
		return "", []byte{}, fmt.Errorf("failed to open file for fileUpload: %w", err)
	}
	defer src.Close()

	bytes, err := io.ReadAll(src)
	if err != nil {
		return "", []byte{}, fmt.Errorf("failed to copy file to bytes for fileUpload: %w", err)
	}

	return fileName, bytes, nil
}
