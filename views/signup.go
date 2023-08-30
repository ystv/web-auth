package views

import (
	"fmt"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
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
	switch c.Request().Method {
	case "POST":
		uSignup := UserSignup{}
		err := decoder.Decode(&uSignup, c.Request().PostForm)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to get form values for signup: %w", err))
		}
		uSignup.Email += "@york.ac.uk"
		err = v.validate.Struct(uSignup)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to parse form: %w", err))
		}

		uNormal := user.User{
			Email: uSignup.Email,
		}

		_, err = v.user.GetUser(c.Request().Context(), uNormal)
		if err == nil {
			return v.template.RenderTemplate(c.Response(), "Account already exists", templates.SignupTemplate, templates.NoNavType)
		}
		return c.Redirect(http.StatusFound, "/")

	case "GET":
		return v.template.RenderTemplate(c.Response(), "", templates.SignupTemplate, templates.NoNavType)
	}
	return fmt.Errorf("invalid mthod used")
}

//nolint:godox
// TODO: Implement signup holding page
