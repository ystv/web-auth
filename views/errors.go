package views

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/ystv/web-auth/templates"
)

func (v *Views) CustomHTTPErrorHandler(err error, c echo.Context) {
	log.Print(err)
	data := struct {
		Error string
	}{
		Error: err.Error(),
	}
	err1 := v.template.RenderNoNavsTemplate(c.Response().Writer, data, templates.ErrorTemplate)
	if err1 != nil {
		log.Printf("failed to render error page: %+v", err1)
	}
}

func (v *Views) Error404(c echo.Context) error {
	return v.template.RenderNoNavsTemplate(c.Response().Writer, nil, templates.NotFound404Template)
}
