package views

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
	"log"
	"net/http"
)

var decoder = schema.NewDecoder()

// UserSignup represents the HTML form
type UserSignup struct {
	Firstname       string `db:"first_name" schema:"firstname" validate:"required,gte=3"`
	Lastname        string `db:"last_name" schema:"lastname" validate:"required,gte=3"`
	Email           string `db:"email" schema:"email" validate:"required,email"`
	Password        string `db:"password" schema:"password" validate:"required,gte=8"`
	ConfirmPassword string `schema:"confirmpassword" validate:"required,eqfield=Password,gte=8"`
}

// SignUpFunc will enable new users to sign up to our service
func (v *Views) SignUpFunc(c echo.Context) error {
	//return nil
	switch c.Request().Method {
	case "POST":
		// Parsing form to struct
		err := c.Request().ParseForm()
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse form for signup: %w", err))
		}
		uSignup := UserSignup{}
		err = decoder.Decode(&uSignup, c.Request().PostForm)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to get form values for signup: %w", err))
		}
		uSignup.Email += "@york.ac.uk"
		err = v.validate.Struct(uSignup)
		if err != nil {
			var validationErrors *validator.ValidationErrors
			if errors.As(err, &validationErrors) {
				return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to validate: %w", err))
			}
			issues := ""
			for _, err := range err.(validator.ValidationErrors) {
				issues += " " + err.Error()
			}
			log.Println(issues)
			return v.template.RenderNoNavsTemplate(c.Response(), issues, templates.SignupTemplate)
		}

		uNormal := user.User{
			Email: uSignup.Email,
		}

		_, err = v.user.GetUser(c.Request().Context(), uNormal)
		if err == nil {
			return v.template.RenderNoNavsTemplate(c.Response(), "Account already exists", templates.SignupTemplate)
		}
		return c.Redirect(http.StatusFound, "/")

	case "GET":
		return v.template.RenderNoNavsTemplate(c.Response(), "", templates.SignupTemplate)
	}
	return fmt.Errorf("invalid mthod used")
}

//TODO: Implement signup holding page
