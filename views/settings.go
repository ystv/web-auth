package views

import (
	"bytes"
	// #nosec
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dustin/go-humanize"
	"github.com/labstack/echo/v4"

	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
)

type (
	// SettingsTemplate is for the settings front end
	SettingsTemplate struct {
		User      user.User
		LastLogin string
		Gravatar  string
		TemplateHelper
	}
)

// SettingsFunc handles a request to the internal template
func (v *Views) SettingsFunc(c echo.Context) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

	c1 := v.getSessionData(c)

	if c.Request().Method == http.MethodPost {
		firstName := c.Request().FormValue("firstname")
		nickname := c.Request().FormValue("nickname")
		lastName := c.Request().FormValue("lastname")
		// avatar type can't be changed yet but the infrastructure is in
		avatar := c.Request().FormValue("avatar")
		_ = avatar

		if firstName != c1.User.Firstname && len(firstName) > 0 {
			c1.User.Firstname = firstName
		}

		if nickname != c1.User.Nickname && len(nickname) > 0 {
			c1.User.Nickname = nickname
		}

		if lastName != c1.User.Lastname && len(lastName) > 0 {
			c1.User.Lastname = lastName
		}

		err := v.user.EditUser(c.Request().Context(), c1.User, c1.User.UserID)
		if err != nil {
			return fmt.Errorf("failed to edit user for settings: %w", err)
		}

		session.Values["user"] = c1.User

		err = session.Save(c.Request(), c.Response())
		if err != nil {
			return fmt.Errorf("failed to save user session in settings: %w", err)
		}

		return c.Redirect(http.StatusFound, "/internal/settings")
	}

	lastLogin := time.Now()
	if c1.User.LastLogin.Valid {
		lastLogin = c1.User.LastLogin.Time
	}

	var gravatar string

	if c1.User.UseGravatar {
		// #nosec
		hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(c1.User.Email))))
		gravatar = "https://www.gravatar.com/avatar/" + hex.EncodeToString(hash[:])
	}

	p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
	if err != nil {
		return fmt.Errorf("failed to get user permissions for settings: %w", err)
	}

	ctx := SettingsTemplate{
		User:      c1.User,
		LastLogin: humanize.Time(lastLogin),
		Gravatar:  gravatar,
		TemplateHelper: TemplateHelper{
			UserPermissions: p1,
			ActivePage:      "settings",
			Assumed:         c1.Assumed,
		},
	}

	return v.template.RenderTemplate(c.Response(), ctx, templates.SettingsTemplate, templates.RegularType)
}

func (v *Views) UploadAvatarFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
		c1 := v.getSessionData(c)

		data := struct {
			Error string `json:"error"`
		}{}

		useGravatarTemp := c.FormValue("useGravatar")
		useGravatar := false

		if useGravatarTemp == "on" {
			useGravatar = true
		}

		if !useGravatar {
			file, err := c.FormFile("upload")
			if err != nil {
				log.Printf("failed to get file for uploadAvatar, user id: %d, error: %+v", c1.User.UserID, err)
				data.Error = fmt.Sprintf("failed to get file for uploadAvatar: %+v", err)
				return c.JSON(http.StatusOK, data)
			}

			fileName, fileBytes, err := v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for uploadAvatar, user id: %d, error: %+v", c1.User.UserID, err)
				data.Error = fmt.Sprintf("failed to upload file for uploadAvatar: %+v", err)
				return c.JSON(http.StatusOK, data)
			}

			buf := bytes.NewReader(fileBytes)

			// This uploads the contents of the buffer to S3
			_, err = v.cdn.PutObjectWithContext(c.Request().Context(), &s3.PutObjectInput{
				Bucket: aws.String("avatars"),
				Key:    aws.String(fileName),
				Body:   buf,
			})
			if err != nil {
				log.Printf("failed to upload file to cdn for uploadAvatar, user id: %d, error: %+v", c1.User.UserID, err)
				data.Error = fmt.Sprintf("failed to upload file to cdn for uploadAvatar: %+v", err)
				return c.JSON(http.StatusOK, data)
			}

			c1.User.Avatar = fmt.Sprintf("%s/avatars/%s", v.conf.CDNEndpoint, fileName)
		}

		c1.User.UseGravatar = useGravatar

		err := v.user.EditUserAvatar(c.Request().Context(), c1.User)
		if err != nil {
			log.Printf("failed to edit user for uploadAvatar, user id: %d, error: %+v", c1.User.UserID, err)
			data.Error = fmt.Sprintf("failed to edit user for uploadAvatar: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = "successfully uploaded avatar"
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for uploadAvatar, user id: %d, error: %+v", c1.User.UserID, err)
		}

		session.Values["user"] = c1.User

		err = session.Save(c.Request(), c.Response())
		if err != nil {
			return fmt.Errorf("failed to save user session in settings: %w", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return v.invalidMethodUsed(c)
}

func (v *Views) RemoveAvatarFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
		c1 := v.getSessionData(c)

		data := struct {
			Error string `json:"error"`
		}{}

		if len(c1.User.Avatar) > 0 && strings.Contains(c1.User.Avatar, v.conf.CDNEndpoint) {
			split := strings.Split(c1.User.Avatar, "/")
			key := split[len(split)-1]
			_, err := v.cdn.DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String("avatars"),
				Key:    aws.String(key),
			})
			if err != nil {
				log.Printf("failed to delete file from cdn for removeAvatar, user id: %d, error: %+v", c1.User.UserID, err)
				data.Error = fmt.Sprintf("failed to delete file from cdn for removeAvatar: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}

		c1.User.Avatar = ""

		err := v.user.EditUserAvatar(c.Request().Context(), c1.User)
		if err != nil {
			log.Printf("failed to edit user for removeAvatar, user id: %d, error: %+v", c1.User.UserID, err)
			data.Error = fmt.Sprintf("failed to edit user for removeAvatar: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = "successfully removed image"
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for removedAvatar, user id: %d, error: %+v", c1.User.UserID, err)
		}

		session.Values["user"] = c1.User

		err = session.Save(c.Request(), c.Response())
		if err != nil {
			return fmt.Errorf("failed to save user session in settings: %w", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return v.invalidMethodUsed(c)
}
