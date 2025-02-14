package views

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/ystv/web-auth/api"
	mockapi "github.com/ystv/web-auth/api/mocks"
	"github.com/ystv/web-auth/user"
	mockuser "github.com/ystv/web-auth/user/mocks"
)

//func TestSetTokenHandler

func TestValidToken(t *testing.T) {
	tokenID := "testing_token"
	userID := 1234
	validSecret := "secret"
	invalidSecret := "invalid_secret"
	expiredDate := time.Now().Add(-time.Hour)
	validClaim := &JWTClaims{
		UserID:      userID,
		Permissions: []string{"test_permission"},
		RegisteredClaims: jwt.RegisteredClaims{
			ID: tokenID,
		},
	}
	expiredClaim := &JWTClaims{
		UserID:      userID,
		Permissions: []string{"test_permission"},
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			ExpiresAt: &jwt.NumericDate{Time: expiredDate},
		},
	}
	invalidClaim := &struct {
		UserID      string `json:"id"`
		Permissions int    `json:"permissions"`
		jwt.RegisteredClaims
	}{
		UserID:      "abc",
		Permissions: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			ID: tokenID,
		},
	}

	for _, tc := range []struct {
		Name                string
		SigningSecret       string
		ValidatingSecret    string
		SigningMethod       jwt.SigningMethod
		Claim               interface{}
		InvalidExpiration   bool
		ExpectAPICall       bool
		ExpectUserCall      bool
		ExpectedNonNilClaim bool
		ExpectedValid       bool
		ExpectedError       bool
	}{
		{
			Name:                "VALID valid token",
			SigningSecret:       validSecret,
			ValidatingSecret:    validSecret,
			SigningMethod:       jwt.SigningMethodHS512,
			Claim:               validClaim,
			ExpectAPICall:       true,
			ExpectUserCall:      true,
			ExpectedValid:       true,
			ExpectedNonNilClaim: true,
			ExpectedError:       false,
		},
		{
			Name:             "INVALID bad validating secret",
			SigningSecret:    validSecret,
			ValidatingSecret: invalidSecret,
			SigningMethod:    jwt.SigningMethodHS512,
			Claim:            validClaim,
			ExpectedValid:    false,
			ExpectedError:    true,
		},
		{
			Name:             "INVALID invalid signing method",
			SigningSecret:    validSecret,
			ValidatingSecret: validSecret,
			SigningMethod:    jwt.SigningMethodHS256,
			Claim:            validClaim,
			ExpectedValid:    false,
			ExpectedError:    true,
		},
		{
			Name:             "INVALID invalid expiration",
			SigningSecret:    validSecret,
			ValidatingSecret: validSecret,
			SigningMethod:    jwt.SigningMethodHS512,
			Claim:            expiredClaim,
			ExpectedValid:    false,
			ExpectedError:    true,
		},
		{
			Name:             "INVALID invalid claim",
			SigningSecret:    validSecret,
			ValidatingSecret: validSecret,
			SigningMethod:    jwt.SigningMethodHS512,
			Claim:            invalidClaim,
			ExpectedValid:    false,
			ExpectedError:    true,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			ctr := gomock.NewController(t)
			mockAPI := mockapi.NewMockRepo(ctr)
			if tc.ExpectAPICall {
				mockAPI.EXPECT().GetToken(gomock.Any(), api.Token{TokenID: tokenID}).Return(api.Token{TokenID: tokenID}, nil)
			}

			mockUser := mockuser.NewMockRepo(ctr)
			if tc.ExpectUserCall {
				mockUser.EXPECT().GetUserValid(gomock.Any(), gomock.Any()).Return(user.User{}, nil)
			}

			// Declare the token with the algorithm used for signing,
			// and the claims.
			token := jwt.NewWithClaims(tc.SigningMethod, tc.Claim.(jwt.Claims))

			// Create the JWT string
			tokenString, err := token.SignedString([]byte(tc.SigningSecret))
			if err != nil {
				t.Errorf("Error signing token: %v", err)
			}

			v := &Views{
				api:  mockAPI,
				user: mockUser,
				conf: &Config{Security: SecurityConfig{SigningKey: tc.ValidatingSecret}},
			}

			valid, claim, err := v.ValidateToken(tokenString)

			if tc.ExpectedError {
				fmt.Println(err)
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tc.ExpectedValid, valid)
			if tc.ExpectedNonNilClaim {
				assert.Equal(t, userID, claim.UserID)
				assert.Equal(t, []string{"test_permission"}, claim.Permissions)
			}
		})
	}
}
