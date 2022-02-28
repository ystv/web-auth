package views

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ystv/web-auth/helpers"
	"github.com/ystv/web-auth/user"
)

type (
	// JWTClaims represents basic identifiable/useful claims
	JWTClaims struct {
		UserID      int      `json:"id"`
		Permissions []string `json:"perms"`
		jwt.StandardClaims
	}
	// statusStruct used for test API as the return JSON
	statusStruct struct {
		StatusCode int    `json:"status_code"`
		Message    string `json:"message"`
	}
)

// SetTokenHandler sets a valid JWT in a cookie instead of returning a string
func (v *Views) SetTokenHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := v.cookie.Get(r, "session")
	u := helpers.GetUser(session)

	tokenString, err := v.newJWT(u)
	if err != nil {
		err = fmt.Errorf("failed to set cookie: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	token := struct {
		Token string `json:"token"`
	}{Token: tokenString}
	tokenByte, err := json.Marshal(token)
	if err != nil {
		err = fmt.Errorf("failed to marshal jwt", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(tokenByte)
	if err != nil {
		err = fmt.Errorf("failed to write token to http body", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (v *Views) newJWT(u user.User) (string, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	perms, err := v.user.GetPermissions(context.Background(), u)
	if err != nil {
		return "", fmt.Errorf("failed to get user permissions: %w", err)
	}
	claims := &JWTClaims{
		UserID:      u.UserID,
		Permissions: perms,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
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

// TestAPI returns a JSON object with a valid JWT
func (v *Views) TestAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		token := r.Header.Get("Authorization")
		splitToken := strings.Split(token, "Bearer ")
		token = splitToken[1]

		if token == "" {
			http.Error(w, "no bearer token provided", http.StatusBadRequest)
			return
		}

		IsTokenValid, claims := v.ValidateToken(token)
		if !IsTokenValid {
			status := statusStruct{
				StatusCode: http.StatusBadRequest,
				Message:    "invalid token",
			}
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(status)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		log.Printf("token is valid \"%d\" is logged in", claims.UserID)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		status := statusStruct{
			StatusCode: http.StatusOK,
			Message:    "valid token",
		}

		err := json.NewEncoder(w).Encode(status)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
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
