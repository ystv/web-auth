package views

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/api"
	"github.com/ystv/web-auth/templates"
	"gopkg.in/guregu/null.v4"
	"log"
	"net/http"
	"strings"
	"time"

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

	ManageAPITemplate struct {
		Tokens     []api.Token
		UserID     int
		AddedJWT   string
		ActivePage string
	}
)

func (v *Views) ManageAPIFunc(c echo.Context) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

	c1 := v.getData(session)

	tokens, err := v.api.GetTokens(c.Request().Context(), c1.User.UserID)
	if err != nil {
		log.Printf("failed to get tokens for manageAPI: %+v", err)
		if !v.conf.Debug {
			return v.errorHandle(c, fmt.Errorf("failed to get tokens for manageAPI: %+v", err))
		}
	}

	data := ManageAPITemplate{
		Tokens:     tokens,
		UserID:     c1.User.UserID,
		ActivePage: "apiManage",
	}

	return v.template.RenderTemplate(c.Response(), data, templates.ManageAPITemplate)
}

func (v *Views) manageAPIFunc(c echo.Context, addedJWT string) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

	c1 := v.getData(session)

	tokens, err := v.api.GetTokens(c.Request().Context(), c1.User.UserID)
	if err != nil {
		log.Printf("failed to get tokens for manageAPI: %+v", err)
		if !v.conf.Debug {
			return v.errorHandle(c, fmt.Errorf("failed to get tokens for manageAPI: %+v", err))
		}
	}

	data := ManageAPITemplate{
		Tokens:     tokens,
		UserID:     c1.User.UserID,
		AddedJWT:   addedJWT,
		ActivePage: "apiManage",
	}

	return v.template.RenderTemplate(c.Response(), data, templates.ManageAPITemplate)
}

func (v *Views) TokenAddFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

		c1 := v.getData(session)

		err := c.Request().ParseForm()
		if err != nil {
			return v.errorHandle(c, fmt.Errorf("failed to parse form for tokenAdd: %+v", err))
		}

		name := c.Request().FormValue("name")
		description := c.Request().FormValue("description")
		expiry := c.Request().FormValue("expiry")

		if len(name) < 2 {
			return v.errorHandle(c, fmt.Errorf("token name too short"))
		}

		id := uuid.NewString()

		parse, err := time.Parse("02/01/2006", expiry)
		if err != nil {
			return v.errorHandle(c, fmt.Errorf("failed to parse expiry: %w", err))
		}

		diff := time.Now().Add(2 * time.Hour * 24).Compare(parse)
		if diff != -1 {
			return v.errorHandle(c, fmt.Errorf("expiry date must be more than 2 days away"))
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
			return v.errorHandle(c, fmt.Errorf("token with id \"%s\" already exists", id))
		}

		addedJWT, err := v.newJWTCustom(c1.User, parse, id)
		if err != nil {
			return v.errorHandle(c, fmt.Errorf("failed to generate jwt for tokenAdd: %w", err))
		}

		_, err = v.api.AddToken(c.Request().Context(), t)
		if err != nil {
			return v.errorHandle(c, fmt.Errorf("error adding token for addToken: %w", err))
		}
		return v.manageAPIFunc(c, addedJWT)
	} else {
		return v.errorHandle(c, fmt.Errorf("invalid method used"))
	}
}

func (v *Views) TokenDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

		c1 := v.getData(session)

		tokenID := c.Param("tokenid")
		if len(tokenID) != 36 {
			return v.errorHandle(c, fmt.Errorf("failed to parse tokenid for tokenDelete: tokenid is the incorrect length"))
		}

		token1, err := v.api.GetToken(c.Request().Context(), api.Token{TokenID: tokenID})
		if err != nil {
			return v.errorHandle(c, fmt.Errorf("failed to get token in tokenDelete: %w", err))
		}

		if token1.UserID != c1.User.UserID {
			return v.errorHandle(c, fmt.Errorf("failed to get token in tokenDelete: unauthorized"))
		}

		err = v.api.DeleteToken(c.Request().Context(), token1)
		if err != nil {
			return v.errorHandle(c, fmt.Errorf("failed to delete token in tokenDelete: %w", err))
		}
		return v.ManageAPIFunc(c)
	} else {
		return v.errorHandle(c, fmt.Errorf("invalid method used"))
	}
}

// SetTokenHandler sets a valid JWT in a cookie instead of returning a string
func (v *Views) SetTokenHandler(c echo.Context) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

	c1 := v.getData(session)

	tokenString, err := v.newJWT(c1.User)
	if err != nil {
		err = fmt.Errorf("failed to set cookie: %w", err)
		http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
		return err
	}
	token := struct {
		Token string `json:"token"`
	}{Token: tokenString}
	tokenByte, err := json.Marshal(token)
	if err != nil {
		err = fmt.Errorf("failed to marshal jwt: %w", err)
		http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
		return err
	}

	c.Response().Header().Set("Content-Type", "application/json")
	c.Response().WriteHeader(http.StatusCreated)
	_, err = c.Response().Write(tokenByte)
	if err != nil {
		err = fmt.Errorf("failed to write token to http body: %w", err)
		http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

func (v *Views) newJWT(u user.User) (string, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	perms, err := v.user.GetPermissionsForUser(context.Background(), u)
	if err != nil {
		return "", fmt.Errorf("failed to get user permissions: %w", err)
	}
	p1 := removeDuplicate(perms)
	var p2 []string
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
	var p2 []string
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

// TestAPI returns a JSON object with a valid JWT
func (v *Views) TestAPI(c echo.Context) error {
	if c.Request().Method == "GET" {
		token := c.Request().Header.Get("Authorization")
		splitToken := strings.Split(token, "Bearer ")
		if len(splitToken) <= 1 {
			return &echo.HTTPError{
				Code:     http.StatusBadRequest,
				Message:  fmt.Sprintf("inalid bearer token provided"),
				Internal: fmt.Errorf("invalid bearer token provided"),
			}
		}
		token = splitToken[1]

		if token == "" {
			http.Error(c.Response(), "no bearer token provided", http.StatusBadRequest)
			return fmt.Errorf("no bearer token provided")
		}

		IsTokenValid, claims := v.ValidateToken(token)
		if !IsTokenValid {
			status := statusStruct{
				StatusCode: http.StatusBadRequest,
				Message:    "invalid token",
			}
			c.Response().WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(c.Response()).Encode(status)
			if err != nil {
				http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
			}
			return err
		}

		log.Printf("token is valid \"%d\" is logged in", claims.UserID)
		c.Response().Header().Set("Content-Type", "application/json; charset=UTF-8")
		c.Response().WriteHeader(http.StatusOK)

		status := statusStruct{
			StatusCode: http.StatusOK,
			Message:    "valid token",
		}

		err := json.NewEncoder(c.Response()).Encode(status)
		if err != nil {
			http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
		}
	}
	return nil
}

// ValidateToken will validate the token
func (v *Views) ValidateToken(token string) (bool, *JWTClaims) {
	parsedToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(v.conf.Security.SigningKey), nil
	})

	if err != nil {
		return false, nil
	}

	if !parsedToken.Valid {
		return false, nil
	}

	claims := parsedToken.Claims.(*JWTClaims)

	if len(claims.ID) > 0 {
		_, err = v.api.GetToken(context.Background(), api.Token{TokenID: claims.ID})
		if err != nil {
			return false, nil
		}
	}

	_, err = v.user.GetUserValid(context.Background(), user.User{UserID: claims.UserID})
	if err != nil {
		return false, nil
	}
	return parsedToken.Valid, claims
}
