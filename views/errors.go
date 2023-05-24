package views

import (
	"fmt"
	"github.com/ystv/web-auth/public/templates"
	"net/http"
)

func (v *Views) Error404(w http.ResponseWriter, _ *http.Request) {
	err := v.template.RenderNoNavsTemplate(w, nil, templates.Error404Template)
	if err != nil {
		fmt.Println(err)
	}
}
