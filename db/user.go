package db

import "log"

//CreateUser will create a new user, take as input the parameters and
//insert it into database
func CreateUser(username, name, nickname, password string) error {
	err := Query("insert into user(username, name, nickname, password) values(?,?,?,?)", username, name, nickname, password)
	return err
}

//ValidateUser will check if the user exists in db and if exists if the username password
//combination is valid
func ValidateUser(username, password string) bool {
	var passwordFromDB string
	userSQL := "select password from user where username=?"
	log.Print("validating user ", username)
	rows := database.query(userSQL, username)

	defer rows.Close()
	if rows.Next() {
		err := rows.Scan(&passwordFromDB)
		if err != nil {
			return false
		}
	}
	//If the password matches, return true
	if password == passwordFromDB {
		return true
	}
	//by default return false
	return false
}

//GetUserID will get the user's ID from the database
func GetUserID(username string) (int, error) {
	var userID int
	userSQL := "select user_id from user where username=?"
	rows := database.query(userSQL, username)

	defer rows.Close()
	if rows.Next() {
		err := rows.Scan(&userID)
		if err != nil {
			return -1, err
		}
	}
	return userID, nil
}
