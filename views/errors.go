package views

import (
	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/templates"
)

func (v *Views) Error404(c echo.Context) error {
	return v.template.RenderNoNavsTemplate(c.Response().Writer, nil, templates.NotFound404Template)
}

func (v *Views) Error500(c echo.Context) error {
	return v.template.RenderNoNavsTemplate(c.Response().Writer, nil, templates.Forbidden500Template)
}
