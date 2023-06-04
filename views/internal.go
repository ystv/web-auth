package views

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/ystv/web-auth/public/templates"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/ystv/web-auth/user"
)

type (
	// InternalTemplate represents the context for the internal template
	InternalTemplate struct {
		UserID              int
		Nickname            string
		LastLogin           string
		TotalUsers          int
		LoginsPast24Hours   int
		ActiveUsersPastYear int
		ActivePage          string
	}

	SettingsTemplate struct {
		User       user.User
		UserID     int
		LastLogin  string
		ActivePage string
		Gravatar   string
	}

	// UserStripped represents user information, an administrator can view
	UserStripped struct {
		UserID      int
		Username    string
		Nickname    string
		Name        string
		Email       string
		LastLogin   string
		Avatar      string
		UseGravatar bool
	}
)

// InternalFunc handles a request to the internal template
func (v *Views) InternalFunc(w http.ResponseWriter, r *http.Request) {
	session, _ := v.cookie.Get(r, v.conf.SessionCookieName)

	c := v.getData(session)
	lastLogin := time.Now()
	if c.User.LastLogin.Valid {
		lastLogin = c.User.LastLogin.Time
	}
	count, err := v.user.CountUsers(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hours24, err := v.user.CountUsers24Hours(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pastYear, err := v.user.CountUsersPastYear(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	ctx := InternalTemplate{
		UserID:              c.User.UserID,
		Nickname:            c.User.Nickname,
		LastLogin:           humanize.Time(lastLogin),
		TotalUsers:          count,
		LoginsPast24Hours:   hours24,
		ActiveUsersPastYear: pastYear,
		ActivePage:          "dashboard",
	}
	err = v.template.RenderTemplate(w, ctx, templates.InternalTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// SettingsFunc handles a request to the internal template
func (v *Views) SettingsFunc(w http.ResponseWriter, r *http.Request) {
	session, _ := v.cookie.Get(r, v.conf.SessionCookieName)

	c := v.getData(session)
	lastLogin := time.Now()
	if c.User.LastLogin.Valid {
		lastLogin = c.User.LastLogin.Time
	}

	var gravatar string

	if c.User.UseGravatar {
		hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace("liam.burnand@bswdi.co.uk"))))
		gravatar = fmt.Sprintf("https://www.gravatar.com/avatar/%s", hex.EncodeToString(hash[:]))
	}

	ctx := SettingsTemplate{
		User:       c.User,
		UserID:     c.User.UserID,
		LastLogin:  humanize.Time(lastLogin),
		ActivePage: "settings",
		Gravatar:   gravatar,
	}
	err := v.template.RenderTemplate(w, ctx, templates.SettingsTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
