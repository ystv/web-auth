package db

import (
	"context"
	"encoding/hex"
	"errors"

	whirl "github.com/balacode/zr-whirl"
	"github.com/ystv/web-auth/types"
)

// UpdateUser will update a user record by ID
func (store *DB) UpdateUser(ctx context.Context, user *types.User) error {
	_, err := store.Exec(ctx,
		`UPDATE people.users
		SET password = $1,
			salt = $2,
			email = $3,
			reset_pw = $4
		WHERE user_id = $5;`, user.Password, user.Salt, user.Email, user.ResetPw)
	if err != nil {
		return err
	}
	return nil
}

// VerifyUser will verify the identity of a user using any of the identity fields and password
func (store *DB) VerifyUser(ctx context.Context, user *types.User) error {
	plaintext := user.Password
	err := store.QueryRow(ctx,
		`SELECT password, salt
		FROM people.users
		WHERE username = $1;`, user.Username).Scan(&user.Password, &user.Salt)
	if err != nil {
		return errors.New("Invalid username/password")
	}
	if hashPass(user.Salt+plaintext) == user.Password {
		return nil
	}
	return errors.New("Invalid username/password")

}

func hashPass(password string) string {
	iter := 1000
	var next string
	for i := 0; i < iter; i++ {
		next += password
		tmp := whirl.HashOfBytes([]byte(next), []byte(""))
		next = hex.EncodeToString(tmp)
	}
	return next
}

// func hashPass(pass []byte) ([]byte, error) {
// 	pass, err := bcrypt.GenerateFromPassword(pass, 10)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return pass, nil
// }

// func checkPassHash(hash, pass []byte) error {
// 	return bcrypt.CompareHashAndPassword(hash, pass)
// }
