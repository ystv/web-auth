package user

import (
	"context"

	"github.com/Nerzal/gocloak/v8"
)

//nolint:unused
func (s *Store) newUser(ctx context.Context, u User) error {
	client := gocloak.NewClient("https://sso2.ystv.co.uk")
	token, err := client.LoginAdmin(ctx, "user", "pass", "realmName")
	if err != nil {
		return err
	}
	user := gocloak.User{
		FirstName:     gocloak.StringP(u.Firstname),
		LastName:      gocloak.StringP(u.Lastname),
		Email:         gocloak.StringP(u.Email),
		EmailVerified: gocloak.BoolP(true),
		Enabled:       gocloak.BoolP(true),
		Username:      gocloak.StringP(u.Username),
	}
	_, err = client.CreateUser(ctx, token.AccessToken, "realName", user)
	return err
}
