package user

import (
	"context"
	"errors"
	"fmt"
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/role"
	"time"

	"github.com/Clarilab/gocloaksession"
	"github.com/jmoiron/sqlx"
	"github.com/ystv/web-auth/utils"
	"gopkg.in/guregu/null.v4"
)

type (
	// Repo where all user data is stored
	Repo interface {
		GetUser(ctx context.Context, u User) error
		GetUsers(ctx context.Context, u User) error
		UpdateUser(ctx context.Context, u User) error
		GetPermissions(ctx context.Context, u User) error
		CheckUserType(ctx context.Context, u User) error
		GetPermissionsForUser(ctx context.Context, u User) ([]permission.Permission, error)
		GetRolesForUser(ctx context.Context, u User) ([]role.Role, error)
	}
	// Store stores the dependencies
	Store struct {
		db    *sqlx.DB
		cloak *gocloaksession.GoCloakSession
	}
	// User represents relevant user fields
	User struct {
		UserID        int                     `db:"user_id" json:"id"`
		Username      string                  `db:"username" json:"username" schema:"username"`
		Nickname      string                  `db:"nickname" json:"nickname" schema:"nickname"`
		Firstname     string                  `db:"first_name" json:"firstName" schema:"firstname"`
		Lastname      string                  `db:"last_name" json:"lastName" schema:"lastname"`
		Password      string                  `db:"password" json:"-" schema:"password"`
		Salt          string                  `db:"salt" json:"-"`
		Avatar        string                  `db:"avatar" json:"avatar" schema:"avatar"`
		Email         string                  `db:"email" json:"email" schema:"email"`
		LastLogin     null.Time               `db:"last_login"`
		ResetPw       bool                    `db:"reset_pw" json:"-"`
		Permissions   []permission.Permission `json:"permissions"`
		UseGravatar   bool                    `db:"use_gravatar" json:"useGravatar" schema:"useGravatar"`
		Authenticated bool
	}
)

// NewUserRepo stores our dependency
func NewUserRepo(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

// CountUsers returns the number of users
func (s *Store) CountUsers(ctx context.Context) (int, error) {
	return s.countUsers(ctx)
}

// CountUsers24Hours returns the number of users who logged in the past 24 hours
func (s *Store) CountUsers24Hours(ctx context.Context) (int, error) {
	return s.countUsers24Hours(ctx)
}

// CountUsersPastYear returns the number of users who logged in the past 24 hours
func (s *Store) CountUsersPastYear(ctx context.Context) (int, error) {
	return s.countUsersPastYear(ctx)
}

// GetUser returns a user using any unique identity fields
func (s *Store) GetUser(ctx context.Context, u User) (User, error) {
	return s.getUser(ctx, u)
}

// GetUsers returns a group of users, used for administration
func (s *Store) GetUsers(ctx context.Context) ([]User, error) {
	return s.getUsers(ctx)
}

// GetUsersSorted returns a group of users, used for administration with sorting
func (s *Store) GetUsersSorted(ctx context.Context, column, direction string) ([]User, error) {
	switch column {
	case "userId":
		switch direction {
		case "asc":
			return s.getUsersIDA(ctx)
		case "desc":
			return s.getUsersIDD(ctx)
		}
	case "name":
		switch direction {
		case "asc":
			return s.getUsersFLNA(ctx)
		case "desc":
			return s.getUsersFLND(ctx)
		}
	case "username":
		switch direction {
		case "asc":
			return s.getUsersUA(ctx)
		case "desc":
			return s.getUsersUD(ctx)
		}
	case "email":
		switch direction {
		case "asc":
			return s.getUsersEA(ctx)
		case "desc":
			return s.getUsersED(ctx)
		}
	case "lastLogin":
		switch direction {
		case "asc":
			return s.getUsersLLA(ctx)
		case "desc":
			return s.getUsersLLD(ctx)
		}
	}
	return nil, fmt.Errorf("error in db sorting")
}

// VerifyUser will check that the password is correct with provided
// credentials and if verified will return the User object
func (s *Store) VerifyUser(ctx context.Context, u User) (User, error) {
	user, err := s.getUser(ctx, u)
	if err != nil {
		return u, fmt.Errorf("failed to get user: %w", err)
	}
	if utils.HashPass(user.Salt+u.Password) == user.Password {
		return user, nil
	}
	return u, errors.New("invalid credentials")
}

// UpdateUserPassword will update the password and set the reset_pw to false
func (s *Store) UpdateUserPassword(ctx context.Context, u User) (User, error) {
	user, err := s.GetUser(ctx, u)
	if err != nil {
		return u, fmt.Errorf("failed to get user: %w", err)
	}
	user.Password = utils.HashPass(user.Salt + u.Password)
	user.ResetPw = false
	err = s.updateUser(ctx, user)
	if err != nil {
		return u, fmt.Errorf("failed to update user: %w", err)
	}
	return user, nil
}

// SetUserLoggedIn will set the last login date to now
func (s *Store) SetUserLoggedIn(ctx context.Context, u User) error {
	u.LastLogin = null.TimeFrom(time.Now())
	return s.updateUser(ctx, u)
}

func (s *Store) CheckUserType(ctx context.Context, u User) error {
	return nil
}

// GetPermissionsForUser returns all permissions of a user
func (s *Store) GetPermissionsForUser(ctx context.Context, u User) ([]permission.Permission, error) {
	return s.getPermissionsForUser(ctx, u)
}

// GetRolesForUser returns all roles of a user
func (s *Store) GetRolesForUser(ctx context.Context, u User) ([]role.Role, error) {
	return s.getRolesForUser(ctx, u)
}
