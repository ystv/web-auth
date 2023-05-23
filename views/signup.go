package views

import (
	"fmt"
	"github.com/ystv/web-auth/public/templates"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/ystv/web-auth/user"
)

// UserSignup represents the HTML form
type UserSignup struct {
	Firstname       string `db:"first_name" schema:"firstname" validate:"required,gte=3"`
	Lastname        string `db:"last_name" schema:"lastname" validate:"required,gte=3"`
	Email           string `db:"email" schema:"email" validate:"required,email"`
	Password        string `db:"password" schema:"password" validate:"required,gte=8"`
	ConfirmPassword string `schema:"confirmpassword" validate:"required,eqfield=Password,gte=8"`
}

// SignUpFunc will enable new users to sign up to our service
func (v *Views) SignUpFunc(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		fmt.Println("DEBUG - SIGNUP POST")
		// Parsing form to struct
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		uSignup := UserSignup{}
		err = decoder.Decode(&uSignup, r.PostForm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		uSignup.Email += "@york.ac.uk"
		err = v.validate.Struct(uSignup)
		if err != nil {
			if _, ok := err.(*validator.ValidationErrors); ok {
				err = fmt.Errorf("failed to validate: %w", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			issues := ""
			for _, err := range err.(validator.ValidationErrors) {
				issues += " " + err.Error()
			}
			log.Println(issues)
			v.signupTmplExec(w, issues)
			return
		}

		uNormal := user.User{
			Email: uSignup.Email,
		}

		_, err = v.user.GetUser(r.Context(), uNormal)
		if err == nil {
			v.signupTmplExec(w, "Account already exists")
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)

	case "GET":
		fmt.Println("DEBUG - SIGNUP POST")
		v.signupTmplExec(w, "")
	}
}

func (v *Views) signupTmplExec(w http.ResponseWriter, msg string) {
	err := v.template.RenderNoNavsTemplate(w, msg, templates.SignupTemplate)
	//err := v.tpl.ExecuteTemplate(w, "signup.tmpl", msg)
	if err != nil {
		err = fmt.Errorf("signup template exec failed: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

//TODO: Implement signup holding page
