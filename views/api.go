package views

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ystv/web-auth/helpers"
)

// JWTClaims represents basic identifiable/useful claims
type JWTClaims struct {
	UserID int `json:"userID"`
	jwt.StandardClaims
}

var signingKey []byte

func init() {
	signingKey = []byte(os.Getenv("signing_key"))
}

// ValidateToken will validate the token
func ValidateToken(myToken string) (bool, *JWTClaims) {
	token, err := jwt.ParseWithClaims(myToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return signingKey, nil
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
func SetTokenHandler(w http.ResponseWriter, r *http.Request) {
	w = getJWTCookie(w, r)
}

// TestAPI returns a JSON object with a valid JWT
func TestAPI(w http.ResponseWriter, r *http.Request) {
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

		IsTokenValid, claims := ValidateToken(token.Value)
		// When the token is not valid show
		// the default error JOSN document
		if !IsTokenValid {
			status = statusStruct{
				StatusCode: http.StatusInternalServerError,
				Message:    message,
			}
			w.WriteHeader(http.StatusInternalServerError)
			// the following statmeent will write the
			// JSON document to the HTTP ReponseWriter object
			err = json.NewEncoder(w).Encode(status)
			if err != nil {
				panic(err)
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
			panic(err)
		}
	}
}

// statusStruct used for test API as the return JSON
type statusStruct struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func getJWTCookie(w http.ResponseWriter, r *http.Request) http.ResponseWriter {
	// Probably would be nice to handle this error
	session, _ := cStore.Get(r, "session")

	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &JWTClaims{
		UserID: helpers.GetUser(session).UserID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Declare the token with the algorithm used for signing,
	// and the claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		// If there is an error in creating the JWT
		w.WriteHeader(http.StatusInternalServerError)
		return w
	}
	// Finally, we set the client cooke for the "token" as the JWT
	// we generated, also setting the expiry time as the same
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
		Domain:  ".ystv.co.uk",
		Path:    "/",
	})
	w.Header().Set("Access-Control-Allow-Origin", "creator.ystv.co.uk:3000")
	return w
}
