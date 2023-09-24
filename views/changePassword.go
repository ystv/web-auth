package views

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"
)

func (v *Views) ChangePasswordFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		err := c.Request().ParseForm()
		if err != nil {
			return fmt.Errorf("failed to parse form for changePassword: %w", err)
		}

		oldPassword := c.Request().FormValue("oldPassword")

		var message struct {
			Message string `json:"message"`
			Error   string `json:"error"`
		}

		var status int

		c1.User.Password = null.StringFrom(oldPassword)

		_, _, err = v.user.VerifyUser(c.Request().Context(), c1.User)
		if err != nil {
			message.Error = "old password is not correct"
			return c.JSON(status, message)
		}

		password := c.Request().FormValue("newPassword")
		errString := minRequirementsMet(password)
		if len(errString) > 0 {
			message.Error = fmt.Sprintf("new password doesn't meet the old requirements: %s", errString)
			return c.JSON(status, message)
		}

		if password != c.Request().FormValue("confirmationPassword") {
			message.Error = "new passwords doesn't match"
			return c.JSON(status, message)
		}

		c1.User.Password = null.StringFrom(password)

		err = v.user.EditUserPassword(c.Request().Context(), c1.User)
		if err != nil {
			message.Error = fmt.Sprintf("failed to change password: %+v", err)
			return c.JSON(status, message)
		}

		message.Message = "successfully changed password"
		return c.JSON(status, message)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}
