package views

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ystv/web-auth/helpers"
	"github.com/ystv/web-auth/user"
)

type (
	// JWTClaims represents basic identifiable/useful claims
	JWTClaims struct {
		UserID      int               `json:"id"`
		Permissions []user.Permission `json:"perms"`
		jwt.StandardClaims
	}
	// statusStruct used for test API as the return JSON
	statusStruct struct {
		StatusCode int    `json:"status_code"`
		Message    string `json:"message"`
	}
)

// ValidateToken will validate the token
func (v *Views) ValidateToken(myToken string) (bool, *JWTClaims) {
	token, err := jwt.ParseWithClaims(myToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(v.conf.Security.SigningKey), nil
	})

	if err != nil {
		return false, nil
	}

	if !token.Valid {
		return false, nil
	}

	claims := token.Claims.(*JWTClaims)
	return token.Valid, claims
}

// SetTokenHandler sets a valid JWT in a cookie instead of returning a string
func (v *Views) SetTokenHandler(w http.ResponseWriter, r *http.Request) {
	w, err := v.getJWTCookie(w, r)
	if err != nil {
		err = fmt.Errorf("login: failed to set cookie: %w", err)
		log.Printf("%+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// TestAPI returns a JSON object with a valid JWT
func (v *Views) TestAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		var err error
		var message string
		var status statusStruct
		token, err := r.Cookie("token")
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		IsTokenValid, claims := v.ValidateToken(token.Value)
		// When the token is not valid show
		// the default error JOSN document
		if !IsTokenValid {
			status = statusStruct{
				StatusCode: http.StatusBadRequest,
				Message:    message,
			}
			w.WriteHeader(http.StatusInternalServerError)
			// the following statment will write the
			// JSON document to the HTTP ReponseWriter object
			err = json.NewEncoder(w).Encode(status)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		log.Printf("token is valid \"%d\" is logged in", claims.UserID)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		status = statusStruct{
			StatusCode: http.StatusOK,
			Message:    "Good",
		}

		err = json.NewEncoder(w).Encode(status)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (v *Views) getJWTCookie(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, error) {
	// Probably would be nice to handle this error
	session, _ := v.cookie.Get(r, "session")

	expirationTime := time.Now().Add(5 * time.Minute)
	u := helpers.GetUser(session)
	perms, err := v.user.GetPermissions(context.Background(), u)
	if err != nil {
		log.Printf("getJWTCookie failed: %+v", err)
		return nil, err
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
		return nil, err
	}
	// Finally, we set the client cooke for the "token" as the JWT
	// we generated, also setting the expiry time as the same
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
		Domain:  v.conf.DomainName,
		Path:    "/",
	})
	return w, nil
}
