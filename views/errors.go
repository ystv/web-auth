package views

import (
	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/templates"
)

// Error404 handles 404 errors
func (v *Views) Error404(c echo.Context) error {
	return v.template.RenderNoNavsTemplate(c.Response().Writer, nil, templates.NotFound404Template)
}

// Error500 handles 500 errors
func (v *Views) Error500(c echo.Context) error {
	return v.template.RenderNoNavsTemplate(c.Response().Writer, nil, templates.Forbidden500Template)
}
