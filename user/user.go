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
		CountUsers(ctx context.Context) (int, error)
		CountUsersActive(ctx context.Context) (int, error)
		CountUsers24Hours(ctx context.Context) (int, error)
		CountUsersPastYear(ctx context.Context) (int, error)

		GetUsers(ctx context.Context) ([]User, error)
		GetUser(ctx context.Context, u User) (User, error)
		GetUsersSizePage(ctx context.Context, size, page int) ([]User, error)
		GetUsersSearch(ctx context.Context, search string) ([]User, error)
		GetUsersSearchSizePage(ctx context.Context, search string, size, page int) ([]User, error)
		GetUsersSorted(ctx context.Context, column, direction string) ([]User, error)
		GetUsersSortedSizePage(ctx context.Context, column, direction string, size, page int) ([]User, error)
		GetUsersSortedSearch(ctx context.Context, column, direction, search string) ([]User, error)
		GetUsersSortedSearchSizePage(ctx context.Context, column, direction, search string, size, page int) ([]User, error)
		VerifyUser(ctx context.Context, u User) (User, bool, error)
		UpdateUserPassword(ctx context.Context, u User) (User, error)
		UpdateUser(ctx context.Context, u User, userID int) (User, error)
		SetUserLoggedIn(ctx context.Context, u User) error
		DeleteUser(ctx context.Context, u User, userID int) error
		GetPermissionsForUser(ctx context.Context, u User) ([]permission.Permission, error)
		GetRolesForUser(ctx context.Context, u User) ([]role.Role, error)
		GetUsersForRole(ctx context.Context, r role.Role) ([]User, error)
		GetPermissionsForRole(ctx context.Context, r role.Role) ([]permission.Permission, error)
		GetRolesForPermission(ctx context.Context, p permission.Permission) ([]role.Role, error)
		newUser(ctx context.Context, u User) error

		countUsers(ctx context.Context) (int, error)
		countUsersActive(ctx context.Context) (int, error)
		countUsers24Hours(ctx context.Context) (int, error)
		countUsersPastYear(ctx context.Context) (int, error)
		updateUser(ctx context.Context, user User) error
		getUser(ctx context.Context, user User) (User, error)
		getUsers(ctx context.Context) ([]User, error)
		getUsersSizePage(ctx context.Context, size, page int) ([]User, error)
		getUsersSearch(ctx context.Context, search string) ([]User, error)
		getUsersSearchSizePage(ctx context.Context, search string, size, page int) ([]User, error)
		getUsersOptionsAsc(ctx context.Context, sortBy string) ([]User, error)
		getUsersOptionsAscSizePage(ctx context.Context, sortBy string, size, page int) ([]User, error)
		getUsersSearchOptionsAsc(ctx context.Context, search, sortBy string) ([]User, error)
		getUsersSearchOptionsAscSizePage(ctx context.Context, search, sortBy string, size, page int) ([]User, error)
		getUsersOptionsDesc(ctx context.Context, sortBy string) ([]User, error)
		getUsersOptionsDescSizePage(ctx context.Context, sortBy string, size, page int) ([]User, error)
		getUsersSearchOptionsDesc(ctx context.Context, search, sortBy string) ([]User, error)
		getUsersSearchOptionsDescSizePage(ctx context.Context, search, sortBy string, size, page int) ([]User, error)
		getRolesForUser(ctx context.Context, u User) ([]role.Role, error)
		getUsersForRole(ctx context.Context, r role.Role) ([]User, error)
		getPermissionsForRole(ctx context.Context, r role.Role) ([]permission.Permission, error)
		getRolesForPermission(ctx context.Context, p permission.Permission) ([]role.Role, error)
	}

	// Store stores the dependencies
	Store struct {
		db    *sqlx.DB
		cloak *gocloaksession.GoCloakSession
	}

	// User represents relevant user fields
	User struct {
		UserID             int                     `db:"user_id" json:"id"`
		Username           string                  `db:"username" json:"username" schema:"username"`
		UniversityUsername null.String             `db:"university_username" json:"universityUsername"`
		LDAPUsername       null.String             `db:"ldap_username" json:"LDAPUsername"`
		LoginType          string                  `db:"login_type" json:"loginType"`
		Nickname           string                  `db:"nickname" json:"nickname" schema:"nickname"`
		Firstname          string                  `db:"first_name" json:"firstName" schema:"firstname"`
		Lastname           string                  `db:"last_name" json:"lastName" schema:"lastname"`
		Password           null.String             `db:"password" json:"-" schema:"password"`
		Salt               null.String             `db:"salt" json:"-"`
		Avatar             string                  `db:"avatar" json:"avatar" schema:"avatar"`
		Email              string                  `db:"email" json:"email" schema:"email"`
		LastLogin          null.Time               `db:"last_login"`
		ResetPw            bool                    `db:"reset_pw" json:"-"`
		Enabled            bool                    `db:"enabled" json:"enabled"`
		CreatedAt          null.Time               `db:"created_at" json:"createdAt"`
		CreatedBy          null.Int                `db:"created_by" json:"createdBy"`
		UpdatedAt          null.Time               `db:"updated_at" json:"updatedAt"`
		UpdatedBy          null.Int                `db:"updated_by" json:"updatedBy"`
		DeletedAt          null.Time               `db:"deleted_at" json:"deletedAt"`
		DeletedBy          null.Int                `db:"deleted_by" json:"deletedBy"`
		UseGravatar        bool                    `db:"use_gravatar" json:"useGravatar" schema:"useGravatar"`
		Permissions        []permission.Permission `json:"permissions"`
		Roles              []role.Role             `json:"roles"`
		Authenticated      bool
	}

	// StrippedUser represents user information, an administrator can view
	StrippedUser struct {
		UserID    int
		Username  string
		Name      string
		Email     string
		LastLogin string
		Enabled   bool
		Deleted   bool
	}

	DetailedUser struct {
		UserID             int                     `json:"id"`
		Username           string                  `json:"username"`
		UniversityUsername null.String             `json:"universityUsername"`
		LDAPUsername       null.String             `json:"LDAPUsername"`
		LoginType          string                  `json:"loginType"`
		Nickname           string                  `json:"nickname"`
		Firstname          string                  `json:"firstName"`
		Lastname           string                  `json:"lastName"`
		Avatar             string                  `json:"avatar"`
		UseGravatar        bool                    `json:"useGravatar"`
		Email              string                  `json:"email"`
		LastLogin          null.String             `json:"lastLogin"`
		ResetPw            bool                    `json:"resetPw"`
		Enabled            bool                    `json:"enabled"`
		CreatedAt          null.String             `json:"createdAt"`
		CreatedBy          User                    `json:"createdBy"`
		UpdatedAt          null.String             `json:"updatedAt"`
		UpdatedBy          User                    `json:"updatedBy"`
		DeletedAt          null.String             `json:"deletedAt"`
		DeletedBy          User                    `json:"deletedBy"`
		Gravatar           null.String             `json:"gravatar"`
		Permissions        []permission.Permission `json:"permissions"`
		Roles              []role.Role             `json:"roles"`
	}

	RoleTemplate struct {
		RoleID      int
		Name        string
		Description string
		Permissions []permission.Permission
		Users       []User
	}

	PermissionTemplate struct {
		PermissionID int
		Name         string
		Description  string
		Roles        []role.Role
	}
)

var (
	_ Repo = &Store{}
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

// CountUsersActive returns the number of active users
func (s *Store) CountUsersActive(ctx context.Context) (int, error) {
	return s.countUsersActive(ctx)
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
	if utils.HashPass(user.Salt.String+u.Password.String) == user.Password.String {
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
	user.Password = null.StringFrom(utils.HashPass(user.Salt.String + u.Password.String))
	user.ResetPw = false
	err = s.updateUser(ctx, user)
	if err != nil {
		return u, fmt.Errorf("failed to update user: %w", err)
	}
	return user, nil
}

// UpdateUser will update the user
func (s *Store) UpdateUser(ctx context.Context, u User, userID int) (User, error) {
	user, err := s.GetUser(ctx, u)
	if err != nil {
		return u, fmt.Errorf("failed to get user: %w", err)
	}
	if u.Username != user.Username && len(u.Username) > 0 {
		user.Username = u.Username
	}
	if u.UniversityUsername.String != user.UniversityUsername.String && len(u.UniversityUsername.String) > 0 {
		user.UniversityUsername = u.UniversityUsername
	}
	if u.LDAPUsername.String != user.LDAPUsername.String && len(u.LDAPUsername.String) > 0 {
		user.LDAPUsername = u.LDAPUsername
	}
	if u.LoginType != user.LoginType && len(u.LoginType) > 0 {
		user.LoginType = u.LoginType
	}
	if u.Nickname != user.Nickname && len(u.Nickname) > 0 {
		user.Nickname = u.Nickname
	}
	if u.Firstname != user.Firstname && len(u.Firstname) > 0 {
		user.Firstname = u.Firstname
	}
	if u.Lastname != user.Lastname && len(u.Lastname) > 0 {
		user.Lastname = u.Lastname
	}
	if u.Avatar != user.Avatar && len(u.Avatar) > 0 {
		user.Avatar = u.Avatar
	}
	if u.Email != user.Email && len(u.Email) > 0 {
		user.Email = u.Email
	}
	if u.ResetPw != user.ResetPw {
		user.ResetPw = u.ResetPw
	}
	if u.Enabled != user.Enabled {
		user.Enabled = u.Enabled
	}
	if u.UseGravatar != user.UseGravatar {
		user.UseGravatar = u.UseGravatar
	}
	user.UpdatedBy = null.IntFrom(int64(userID))
	user.UpdatedAt = null.TimeFrom(time.Now())
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

// DeleteUser will delete a user
func (s *Store) DeleteUser(ctx context.Context, u User, userID int) error {
	now := null.TimeFrom(time.Now())
	u.Enabled = false
	u.Password = null.NewString("", false)
	u.Salt = null.NewString("", false)
	u.UpdatedBy = null.IntFrom(int64(userID))
	u.UpdatedAt = now
	u.DeletedBy = null.IntFrom(int64(userID))
	u.DeletedAt = now
	return s.updateUser(ctx, u)
}

// GetPermissionsForUser returns all permissions of a user
func (s *Store) GetPermissionsForUser(ctx context.Context, u User) ([]permission.Permission, error) {
	return s.getPermissionsForUser(ctx, u)
}

// GetRolesForUser returns all roles of a user
func (s *Store) GetRolesForUser(ctx context.Context, u User) ([]role.Role, error) {
	return s.getRolesForUser(ctx, u)
}

func (s *Store) GetUsersForRole(ctx context.Context, r role.Role) ([]User, error) {
	return s.getUsersForRole(ctx, r)
}

// GetPermissionsForRole returns all permissions for role
func (s *Store) GetPermissionsForRole(ctx context.Context, r role.Role) ([]permission.Permission, error) {
	return s.getPermissionsForRole(ctx, r)
}

// GetRolesForPermission returns all roles where a permission is used
func (s *Store) GetRolesForPermission(ctx context.Context, p permission.Permission) ([]role.Role, error) {
	return s.getRolesForPermission(ctx, p)
}
