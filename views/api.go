package views

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/rmil/web-auth/db"
	"github.com/rmil/web-auth/sessions"
)

// JWTClaims represents basic identifiable/useful claims
type JWTClaims struct {
	UserID   int    `json:"userID"`
	Username string `json:"username"`
	jwt.StandardClaims
}

var signingKey []byte

func init() {
	signingKey = []byte(os.Getenv("signing_key"))
}

// GetTokenHandler will get a token for the username and password
func GetTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Write([]byte("Method not allowed"))
		return
	}

	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	password = hashPassword(password)

	if username == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid Username or password"))
		return
	}
	if db.ValidUser(username, password) {
		claims := JWTClaims{
			0,
			username,
			jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Hour * 5).Unix(),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// Sign the token with our secret
		tokenString, err := token.SignedString(signingKey)
		if err != nil {
			log.Println("Something went wrong with signing token")
			w.Write([]byte("Authentication failed"))
			return
		}
		// Finally, write the token to the browser window
		w.Write([]byte(tokenString))
	} else {
		w.Write([]byte("Authentication failed"))
	}
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
	expirationTime := time.Now().Add(5 * time.Minute)
	if !sessions.IsLoggedIn(r) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	claims := &JWTClaims{
		UserID:   sessions.GetUserID(r),
		Username: sessions.GetUsername(r),
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
		return
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
}

// RefreshHandler refreshes the JWT token using a stil in date
// JWT instead of using the session.
func RefreshHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tokenString := c.Value
	claims := &JWTClaims{}

	ok, claims := ValidateToken(tokenString)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Will check if token is within 30 seconds of expiry, otherwise bad request
	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// A new token is then generated.
	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = expirationTime.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(signingKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the new token as the users "token" cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
		Path:    "/",
	})
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

		log.Printf("token is valid \"%s\" is logged in", claims.Username)
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
