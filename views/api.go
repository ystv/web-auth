package views

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
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
)

// SetTokenHandler sets a valid JWT in a cookie instead of returning a string
func (v *Views) SetTokenHandler(c echo.Context) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
	c1 := v.getSessionData(session)

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

	c.Response().Header().Set("Content-Type", "application/json")
	c.Response().WriteHeader(http.StatusCreated)
	_, err = c.Response().Write(tokenByte)
	if err != nil {
		log.Printf("failed to write token to http body: %+v", err)
		data := struct {
			Error error `json:"error"`
		}{
			Error: fmt.Errorf("failed to write token to http body: %w", err),
		}
		return c.JSON(http.StatusInternalServerError, data)
	}
	return nil
}

func (v *Views) newJWT(u user.User) (string, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	perms, err := v.user.GetPermissionsForUser(context.Background(), u)
	if err != nil {
		return "", fmt.Errorf("failed to get user permissions: %w", err)
	}
	p1 := v.removeDuplicate(perms)
	var p2 []string
	for _, p := range p1 {
		p2 = append(p2, p.Name)
	}
	claims := &JWTClaims{
		UserID:      u.UserID,
		Permissions: p2,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: expirationTime},
		},
	}

	// Declare the token with the algorithm used for signing,
	// and the claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
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

		IsTokenValid, claims := v.ValidateToken(token)
		if !IsTokenValid {
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
	return parsedToken.Valid, claims
}
