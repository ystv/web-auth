package views

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/ystv/web-auth/api"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
)

type (
	// JWTClaims represents basic identifiable/useful claims
	JWTClaims struct {
		UserID      int      `json:"id"`
		Permissions []string `json:"perms"`
		jwt.RegisteredClaims
	}
	// statusStruct used for test API as the return JSON
	statusStruct struct {
		StatusCode int    `json:"status_code"`
		Message    string `json:"message"`
	}

	// ManageAPITemplate returns the data to the front end
	ManageAPITemplate struct {
		Tokens   []api.Token
		AddedJWT string
		TemplateHelper
	}
)

// ManageAPIFunc is the main home page for API management
func (v *Views) ManageAPIFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	tokens, err := v.api.GetTokens(c.Request().Context(), c1.User.UserID)
	if err != nil {
		return fmt.Errorf("failed to get tokens for manageAPI: %w", err)
	}

	p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
	if err != nil {
		return fmt.Errorf("failed to get user permissions for users: %w", err)
	}

	data := ManageAPITemplate{
		Tokens: tokens,
		TemplateHelper: TemplateHelper{
			UserPermissions: p1,
			ActivePage:      "apiManage",
			Assumed:         c1.Assumed,
		},
	}

	return v.template.RenderTemplate(c.Response(), data, templates.ManageAPITemplate, templates.RegularType)
}

// ManageAPIFunc is the main home page for API management internal
func (v *Views) manageAPIFunc(c echo.Context, addedJWT string) error {
	c1 := v.getSessionData(c)

	tokens, err := v.api.GetTokens(c.Request().Context(), c1.User.UserID)
	if err != nil {
		return fmt.Errorf("failed to get tokens for manageAPI: %w", err)
	}

	p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
	if err != nil {
		return fmt.Errorf("failed to get user permissions for users: %w", err)
	}

	data := ManageAPITemplate{
		Tokens:   tokens,
		AddedJWT: addedJWT,
		TemplateHelper: TemplateHelper{
			UserPermissions: p1,
			ActivePage:      "apiManage",
			Assumed:         c1.Assumed,
		},
	}

	return v.template.RenderTemplate(c.Response(), data, templates.ManageAPITemplate, templates.RegularType)
}

// TokenAddFunc adds a token to be used by the user
func (v *Views) TokenAddFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		err := c.Request().ParseForm()
		if err != nil {
			return fmt.Errorf("failed to parse form for tokenAdd: %w", err)
		}

		name := c.Request().FormValue("name")
		description := c.Request().FormValue("description")
		expiry := c.Request().FormValue("expiry")

		if len(name) < 2 {
			return fmt.Errorf("token name too short")
		}

		id := uuid.NewString()

		parse, err := time.Parse("02/01/2006", expiry)
		if err != nil {
			return fmt.Errorf("failed to parse expiry: %w", err)
		}

		diff := time.Now().Add(2 * time.Hour * 24).Compare(parse)
		if diff != -1 {
			return fmt.Errorf("expiry date must be more than 2 days away")
		}

		t := api.Token{
			TokenID:     id,
			Name:        name,
			Description: description,
			Expiry:      null.TimeFrom(parse),
			UserID:      c1.User.UserID,
		}

		t1, err := v.api.GetToken(c.Request().Context(), t)
		if err == nil && len(t1.TokenID) > 0 {
			return fmt.Errorf("token with id \"%s\" already exists", id)
		}

		addedJWT, err := v.newJWTCustom(c1.User, parse, id)
		if err != nil {
			return fmt.Errorf("failed to generate jwt for tokenAdd: %w", err)
		}

		_, err = v.api.AddToken(c.Request().Context(), t)
		if err != nil {
			return fmt.Errorf("error adding token for addToken: %w", err)
		}
		return v.manageAPIFunc(c, addedJWT)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}

// TokenDeleteFunc deletes a token
func (v *Views) TokenDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		tokenID := c.Param("tokenid")
		if len(tokenID) != 36 {
			return fmt.Errorf("failed to parse tokenid for tokenDelete: tokenid is the incorrect length")
		}

		token1, err := v.api.GetToken(c.Request().Context(), api.Token{TokenID: tokenID})
		if err != nil {
			return fmt.Errorf("failed to get token in tokenDelete: %w", err)
		}

		if token1.UserID != c1.User.UserID {
			return fmt.Errorf("failed to get token in tokenDelete: unauthorized")
		}

		err = v.api.DeleteToken(c.Request().Context(), token1)
		if err != nil {
			return fmt.Errorf("failed to delete token in tokenDelete: %w", err)
		}
		return v.ManageAPIFunc(c)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}

// SetTokenHandler sets a valid JWT in a cookie instead of returning a string
func (v *Views) SetTokenHandler(c echo.Context) error {
	c1 := v.getSessionData(c)

	tokenString, err := v.newJWT(c1.User)
	if err != nil {
		log.Printf("failed to set cookie: %+v", err)
		data := struct {
			Error error `json:"error"`
		}{
			Error: fmt.Errorf("failed to set cookie: %w", err),
		}
		return c.JSON(http.StatusInternalServerError, data)
	}
	token := struct {
		Token string `json:"token"`
	}{Token: tokenString}
	tokenByte, err := json.Marshal(token)
	if err != nil {
		log.Printf("failed to marshal json: %+v", err)
		data := struct {
			Error error `json:"error"`
		}{
			Error: fmt.Errorf("failed to marshal json: %w", err),
		}
		return c.JSON(http.StatusInternalServerError, data)
	}

	_ = tokenByte

	callback := ""
	callbackURL, err := url.Parse(c.QueryParam("callback"))
	fmt.Println(callbackURL.String(), err)
	if err == nil /*&& strings.HasSuffix(callbackURL.Host, v.conf.BaseDomainName)*/ && callbackURL.String() != "" {
		callback = callbackURL.String()
	}
	//c.Response().Header().Set("Content-Type", "application/json")
	c.Response().Header().Set("Authorization", "Bearer "+tokenString)
	cookie := new(http.Cookie)
	cookie.Name = "token"
	cookie.Expires = time.Now().Add(30 * time.Second)
	cookie.Value = tokenString
	cookie.Secure = false
	cookie.HttpOnly = false
	cookie.Domain = "localhost"
	c.SetCookie(cookie)
	http.SetCookie(c.Response().Writer, cookie)
	c.Response().Committed = false
	//c.Response().Write(tokenByte)
	//_, err = c.Response().Write(tokenByte)
	//if err != nil {
	//	log.Printf("failed to write token to http body: %+v", err)
	//	data := struct {
	//		Error error `json:"error"`
	//	}{
	//		Error: fmt.Errorf("failed to write token to http body: %w", err),
	//	}
	//	return c.JSON(http.StatusInternalServerError, data)
	//}
	if len(callback) > 0 {
		//c.Response().Header().Set("Location", callback)
		//c.Response().WriteHeader(http.StatusFound)
		return c.Redirect(http.StatusFound, callback+"?token="+tokenString)
		//c.Redirect()
	}
	return nil
}

// newJWT generates a new jwt token
func (v *Views) newJWT(u user.User) (string, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	perms, err := v.user.GetPermissionsForUser(context.Background(), u)
	if err != nil {
		return "", fmt.Errorf("failed to get user permissions: %w", err)
	}
	p1 := removeDuplicate(perms)
	p2 := make([]string, 0, len(p1))
	for _, p := range p1 {
		p2 = append(p2, p.Name)
	}
	claims := &JWTClaims{
		UserID:      u.UserID,
		Permissions: p2,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  &jwt.NumericDate{Time: time.Now()},
			ExpiresAt: &jwt.NumericDate{Time: expirationTime},
		},
	}

	// Declare the token with the algorithm used for signing,
	// and the claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	// Create the JWT string
	tokenString, err := token.SignedString([]byte(v.conf.Security.SigningKey))
	if err != nil {
		// If there is an error in creating the JWT
		return "", fmt.Errorf("failed to make jwt string: %w", err)
	}
	return tokenString, nil
}

// newJWTCustom generates a new jwt token for the user
func (v *Views) newJWTCustom(u user.User, expiry time.Time, tokenID string) (string, error) {
	compare := expiry.Compare(time.Now().AddDate(1, 0, 0))
	if compare == 1 {
		return "", fmt.Errorf("expiration date is more than a year away, can only have a maximum of 1 year")
	}

	perms, err := v.user.GetPermissionsForUser(context.Background(), u)
	if err != nil {
		return "", fmt.Errorf("failed to get user permissions: %w", err)
	}

	p1 := removeDuplicate(perms)

	p2 := make([]string, 0, len(p1))
	for _, p := range p1 {
		p2 = append(p2, p.Name)
	}
	claims := &JWTClaims{
		UserID:      u.UserID,
		Permissions: p2,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			IssuedAt:  &jwt.NumericDate{Time: time.Now()},
			ExpiresAt: &jwt.NumericDate{Time: expiry},
		},
	}

	// Declare the token with the algorithm used for signing,
	// and the claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	// Create the JWT string
	tokenString, err := token.SignedString([]byte(v.conf.Security.SigningKey))
	if err != nil {
		// If there is an error in creating the JWT
		return "", fmt.Errorf("failed to make jwt string: %w", err)
	}
	return tokenString, nil
}

// TestAPITokenFunc returns a JSON object if the JWT in the Authorization header is valid.
func (v *Views) TestAPITokenFunc(c echo.Context) error {
	if c.Request().Method == "GET" {
		token := c.Request().Header.Get("Authorization")
		splitToken := strings.Split(token, "Bearer ")
		if len(splitToken) <= 1 {
			log.Println("invalid bearer token provided")
			data := struct {
				Error string `json:"error"`
			}{
				Error: "invalid bearer token provided",
			}
			return c.JSON(http.StatusBadRequest, data)
		}
		token = splitToken[1]

		if token == "" {
			log.Println("no bearer token provided")
			data := struct {
				Error string `json:"error"`
			}{
				Error: "no bearer token provided",
			}
			return c.JSON(http.StatusBadRequest, data)
		}

		valid, claims, err := v.ValidateToken(token)
		log.Printf("valid: %t - claims: %+v - error: %+v", valid, claims, err)
		if !valid {
			status := statusStruct{
				StatusCode: http.StatusBadRequest,
				Message:    "invalid token",
			}
			c.Response().WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(c.Response()).Encode(status)
			if err != nil {
				log.Printf("failed to encode json: %+v", err)
				data := struct {
					Error error `json:"error"`
				}{
					Error: fmt.Errorf("failed to encode json: %w", err),
				}
				return c.JSON(http.StatusInternalServerError, data)
			}
			return c.JSON(http.StatusBadRequest, status)
		}

		log.Printf("token is valid \"%d\" is logged in", claims.UserID)
		c.Response().Header().Set("Content-Type", "application/json; charset=UTF-8")
		c.Response().WriteHeader(http.StatusOK)

		status := statusStruct{
			StatusCode: http.StatusOK,
			Message:    "valid token",
		}

		err = json.NewEncoder(c.Response()).Encode(status)
		if err != nil {
			log.Printf("failed to encode json: %+v", err)
			data := struct {
				Error error `json:"error"`
			}{
				Error: fmt.Errorf("failed to encode json: %w", err),
			}
			return c.JSON(http.StatusInternalServerError, data)
		}
	}
	return nil
}

// ValidateToken will validate the token
func (v *Views) ValidateToken(token string) (bool, *JWTClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(v.conf.Security.SigningKey), nil
	})
	if err != nil {
		return false, nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !parsedToken.Valid {
		return false, nil, fmt.Errorf("failed to validate token: invalid token")
	}

	claims := parsedToken.Claims.(*JWTClaims)

	if len(claims.ID) > 0 {
		_, err = v.api.GetToken(context.Background(), api.Token{TokenID: claims.ID})
		if err != nil {
			return false, nil, fmt.Errorf("failed to get token: %w", err)
		}
	}

	_, err = v.user.GetUserValid(context.Background(), user.User{UserID: claims.UserID})
	if err != nil {
		return false, nil, fmt.Errorf("failed to get valid user: %w", err)
	}
	return parsedToken.Valid, claims, nil
}
