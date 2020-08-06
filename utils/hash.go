package utils

import (
	"encoding/hex"

	whirl "github.com/balacode/zr-whirl"
)

// HashPass hashes a password using a Whirlpool hash.
// Passwords are presumed to be hashed.
func HashPass(password string) string {
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
