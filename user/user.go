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
		UserID             int         `db:"user_id" json:"id"`
		Username           string      `db:"username" json:"username" schema:"username"`
		UniversityUsername null.String `db:"university_username" json:"universityUsername"`
		LDAPUsername       null.String `db:"ldap_username" json:"LDAPUsername"`
		LoginType          string      `db:"login_type" json:"loginType"`
		Nickname           string      `db:"nickname" json:"nickname" schema:"nickname"`
		Firstname          string      `db:"first_name" json:"firstName" schema:"firstname"`
		Lastname           string      `db:"last_name" json:"lastName" schema:"lastname"`
		Password           string      `db:"password" json:"-" schema:"password"`
		Salt               string      `db:"salt" json:"-"`
		Avatar             string      `db:"avatar" json:"avatar" schema:"avatar"`
		Email              string      `db:"email" json:"email" schema:"email"`
		LastLogin          null.Time   `db:"last_login"`
		ResetPw            bool        `db:"reset_pw" json:"-"`
		Enabled            bool        `db:"enabled" json:"enabled"`
		CreatedAt          null.Time   `db:"created_at" json:"createdAt"`
		CreatedBy          null.Int    `db:"created_by" json:"createdBy"`
		UpdatedAt          null.Time   `db:"updated_at" json:"updatedAt"`
		UpdatedBy          null.Int    `db:"updated_by" json:"updatedBy"`
		DeletedAt          null.Time   `db:"deleted_at" json:"deletedAt"`
		DeletedBy          null.Int    `db:"deleted_by" json:"deletedBy"`
		UseGravatar        bool        `db:"use_gravatar" json:"useGravatar" schema:"useGravatar"`
		Permissions        []string    `json:"permissions"`
		Roles              []string    `json:"roles"`
		//Pages              int         `db:"pages" json:"pages"`
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

// GetUsersSizePage returns a group of users, used for administration with size and page
func (s *Store) GetUsersSizePage(ctx context.Context, size, page int) ([]User, error) {
	return s.getUsersSizePage(ctx, size, page)
}

// GetUsersSearch returns a group of users, used for administration
func (s *Store) GetUsersSearch(ctx context.Context, search string) ([]User, error) {
	return s.getUsersSearch(ctx, search)
}

// GetUsersSearchSizePage returns a group of users, used for administration with size and page
func (s *Store) GetUsersSearchSizePage(ctx context.Context, search string, size, page int) ([]User, error) {
	return s.getUsersSearchSizePage(ctx, search, size, page)
}

// GetUsersSorted returns a group of users, used for administration with sorting
func (s *Store) GetUsersSorted(ctx context.Context, column, direction string) ([]User, error) {
	switch direction {
	case "asc":
		return s.getUsersOptionsAsc(ctx, column)
	case "desc":
		return s.getUsersOptionsDesc(ctx, column)
	}
	return nil, fmt.Errorf("error in db sorting")
}

// GetUsersSortedSizePage returns a group of users, used for administration with sorting with size and page
func (s *Store) GetUsersSortedSizePage(ctx context.Context, column, direction string, size, page int) ([]User, error) {
	switch direction {
	case "asc":
		return s.getUsersOptionsAscSizePage(ctx, column, size, page)
	case "desc":
		return s.getUsersOptionsDescSizePage(ctx, column, size, page)
	}
	return nil, fmt.Errorf("error in db sorting size page")
}

// GetUsersSortedSearch returns a group of users, used for administration with sorting and searching
func (s *Store) GetUsersSortedSearch(ctx context.Context, column, direction, search string) ([]User, error) {
	switch direction {
	case "asc":
		return s.getUsersSearchOptionsAsc(ctx, search, column)
	case "desc":
		return s.getUsersSearchOptionsDesc(ctx, search, column)
	}
	return nil, fmt.Errorf("error in db sorting")
}

// GetUsersSortedSearchSizePage returns a group of users, used for administration with sorting and searching with pages
func (s *Store) GetUsersSortedSearchSizePage(ctx context.Context, column, direction, search string, size, page int) ([]User, error) {
	switch direction {
	case "asc":
		return s.getUsersSearchOptionsAscSizePage(ctx, search, column, size, page)
	case "desc":
		return s.getUsersSearchOptionsDescSizePage(ctx, search, column, size, page)
	}
	return nil, fmt.Errorf("error in db sorting size page")
}

// VerifyUser will check that the password is correct with provided
// credentials and if verified will return the User object
func (s *Store) VerifyUser(ctx context.Context, u User) (User, bool, error) {
	user, err := s.getUser(ctx, u)
	if err != nil {
		return u, false, fmt.Errorf("failed to get user: %w", err)
	}
	if !user.Enabled {
		return u, false, errors.New("user not enabled, contact Computing Team for help")
	}
	if user.DeletedBy.Valid {
		return u, false, errors.New("user has been deleted, contact Computing Team for help")
	}
	if user.ResetPw {
		u.UserID = user.UserID
		return user, true, errors.New("password reset required")
	}
	if utils.HashPass(user.Salt+u.Password) == user.Password {
		return user, false, nil
	}
	return u, false, errors.New("invalid credentials")
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

//func (s *Store) CheckUserType(ctx context.Context, u User) error {
//	return nil
//}

// GetPermissionsForUser returns all permissions of a user
func (s *Store) GetPermissionsForUser(ctx context.Context, u User) ([]string, error) {
	return s.getPermissionsForUser(ctx, u)
}

// GetRolesForUser returns all roles of a user
func (s *Store) GetRolesForUser(ctx context.Context, u User) ([]role.Role, error) {
	return s.getRolesForUser(ctx, u)
}
