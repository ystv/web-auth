package views

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"

	"github.com/ystv/web-auth/templates"
)

func (v *Views) CustomHTTPErrorHandler(err error, c echo.Context) {
	log.Print(err)

	var he *echo.HTTPError

	var status int
	if errors.As(err, &he) {
		status = he.Code
	} else {
		status = 500
	}

	var message interface{}
	message = err

	if he != nil {
		message = he.Message
	}

	c.Response().WriteHeader(status)

	data := struct {
		Code  int
		Error any
	}{
		Code:  status,
		Error: message,
	}

	err1 := v.template.RenderTemplate(c.Response().Writer, data, templates.ErrorTemplate, templates.NoNavType)
	if err1 != nil {
		log.Printf("failed to render error page: %+v", err1)
	}
}

func (v *Views) Error404(c echo.Context) error {
	log.Printf("not found, path: %s, method: %s", c.Path(), c.Request().Method)

	return v.template.RenderTemplate(c.Response().Writer, nil, templates.NotFound404Template, templates.NoNavType)
}

func (v *Views) invalidMethodUsed(c echo.Context) *echo.HTTPError {
	return &echo.HTTPError{
		Code:     http.StatusMethodNotAllowed,
		Message:  "invalid method used",
		Internal: fmt.Errorf("invalid method used, path: %s, method: %s", c.Path(), c.Request().Method),
	}
}
