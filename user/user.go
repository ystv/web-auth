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

		GetUser(ctx context.Context, u User) (User, error)
		GetUsers(ctx context.Context, size, page int, enabled, deleted string) ([]User, error)
		GetUsersSearchNoOrder(ctx context.Context, size, page int, search, enabled, deleted string) ([]User, error)
		GetUsersOrderNoSearch(ctx context.Context, size, page int, sortBy, direction, enabled, deleted string) ([]User, error)
		GetUsersSearchOrder(ctx context.Context, size, page int, search, sortBy, direction, enabled, deleted string) ([]User, error)
		VerifyUser(ctx context.Context, u User) (User, bool, error)
		UpdateUserPassword(ctx context.Context, u User) (User, error)
		UpdateUser(ctx context.Context, u User, userID int) (User, error)
		SetUserLoggedIn(ctx context.Context, u User) (User, error)
		DeleteUser(ctx context.Context, u User, userID int) (User, error)
		GetPermissionsForUser(ctx context.Context, u User) ([]permission.Permission, error)
		GetRolesForUser(ctx context.Context, u User) ([]role.Role, error)
		GetUsersForRole(ctx context.Context, r role.Role) ([]User, error)
		GetRoleUser(ctx context.Context, ru RoleUser) (RoleUser, error)
		GetUsersNotInRole(ctx context.Context, r role.Role) ([]User, error)
		AddRoleUser(ctx context.Context, ru RoleUser) (RoleUser, error)
		RemoveRoleUser(ctx context.Context, ru RoleUser) error
		GetPermissionsForRole(ctx context.Context, r role.Role) ([]permission.Permission, error)
		GetRolesForPermission(ctx context.Context, p permission.Permission) ([]role.Role, error)
		GetRolePermission(ctx context.Context, rp RolePermission) (RolePermission, error)
		GetPermissionsNotInRole(ctx context.Context, r role.Role) ([]permission.Permission, error)
		AddRolePermission(ctx context.Context, rp RolePermission) (RolePermission, error)
		RemoveRolePermission(ctx context.Context, rp RolePermission) error

		newUser(ctx context.Context, u User) error

		countUsers(ctx context.Context) (int, error)
		countUsersActive(ctx context.Context) (int, error)
		countUsers24Hours(ctx context.Context) (int, error)
		countUsersPastYear(ctx context.Context) (int, error)

		updateUser(ctx context.Context, user User) (User, error)
		getUser(ctx context.Context, user User) (User, error)
		getUsers(ctx context.Context, size, page int, enabled, deleted string) ([]User, error)
		getUsersSearchNoOrder(ctx context.Context, size, page int, search, enabled, deleted string) ([]User, error)
		getUsersOrderNoSearch(ctx context.Context, size, page int, sortBy, direction, enabled, deleted string) ([]User, error)
		getUsersSearchOrder(ctx context.Context, size, page int, search, sortBy, direction, enabled, deleted string) ([]User, error)
		getRolesForUser(ctx context.Context, u User) ([]role.Role, error)
		getUsersForRole(ctx context.Context, r role.Role) ([]User, error)
		getRoleUser(ctx context.Context, ru RoleUser) (RoleUser, error)
		getUsersNotInRole(ctx context.Context, r role.Role) ([]User, error)
		addRoleUser(ctx context.Context, ru RoleUser) (RoleUser, error)
		removeRoleUser(ctx context.Context, ru RoleUser) error
		getPermissionsForRole(ctx context.Context, r role.Role) ([]permission.Permission, error)
		getRolesForPermission(ctx context.Context, p permission.Permission) ([]role.Role, error)
		getRolePermission(ctx context.Context, rp RolePermission) (RolePermission, error)
		getPermissionsNotInRole(ctx context.Context, r role.Role) ([]permission.Permission, error)
		addRolePermission(ctx context.Context, rp RolePermission) (RolePermission, error)
		removeRolePermission(ctx context.Context, rp RolePermission) error

		parseDirection(direction string) (string, string, error)
		parseEnabled(enabled string, includeAND bool) string
		parseDeleted(deleted string, includeAND bool) string
		parsePageSize(page, size int) string
	}

	// Store stores the dependencies
	Store struct {
		db    *sqlx.DB
		cloak *gocloaksession.GoCloakSession
	}

	// User represents relevant user fields
	User struct {
		UserID             int                     `db:"user_id" json:"userID"`
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
		LastLogin          null.Time               `db:"last_login" json:"lastLogin"`
		ResetPw            bool                    `db:"reset_pw" json:"resetPw"`
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

	RolePermission struct {
		RoleID       int `db:"role_id" json:"roleID"`
		PermissionID int `db:"permission_id" json:"permissionID"`
	}

	RoleUser struct {
		RoleID int `db:"role_id" json:"roleID"`
		UserID int `db:"user_id" json:"userID"`
	}
)

var _ Repo = &Store{}

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

// GetUsers returns a group of users, used for administration with size and page
func (s *Store) GetUsers(ctx context.Context, size, page int, enabled, deleted string) ([]User, error) {
	return s.getUsers(ctx, size, page, enabled, deleted)
}

func (s *Store) GetUsersSearchNoOrder(ctx context.Context, size, page int, search, enabled, deleted string) ([]User, error) {
	return s.getUsersSearchNoOrder(ctx, size, page, search, enabled, deleted)
}

func (s *Store) GetUsersOrderNoSearch(ctx context.Context, size, page int, sortBy, direction, enabled, deleted string) ([]User, error) {
	return s.getUsersOrderNoSearch(ctx, size, page, sortBy, direction, enabled, deleted)
}

func (s *Store) GetUsersSearchOrder(ctx context.Context, size, page int, search, sortBy, direction, enabled, deleted string) ([]User, error) {
	return s.getUsersSearchOrder(ctx, size, page, search, sortBy, direction, enabled, deleted)
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
	user, err = s.updateUser(ctx, user)
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
	user, err = s.updateUser(ctx, user)
	if err != nil {
		return u, fmt.Errorf("failed to update user: %w", err)
	}
	return user, nil
}

// SetUserLoggedIn will set the last login date to now
func (s *Store) SetUserLoggedIn(ctx context.Context, u User) (User, error) {
	u.LastLogin = null.TimeFrom(time.Now())
	return s.updateUser(ctx, u)
}

// DeleteUser will delete a user
func (s *Store) DeleteUser(ctx context.Context, u User, userID int) (User, error) {
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

func (s *Store) GetRoleUser(ctx context.Context, ru RoleUser) (RoleUser, error) {
	return s.getRoleUser(ctx, ru)
}

func (s *Store) GetUsersNotInRole(ctx context.Context, r role.Role) ([]User, error) {
	return s.getUsersNotInRole(ctx, r)
}

func (s *Store) AddRoleUser(ctx context.Context, ru RoleUser) (RoleUser, error) {
	return s.addRoleUser(ctx, ru)
}

func (s *Store) RemoveRoleUser(ctx context.Context, ru RoleUser) error {
	return s.removeRoleUser(ctx, ru)
}

// GetPermissionsForRole returns all permissions for role
func (s *Store) GetPermissionsForRole(ctx context.Context, r role.Role) ([]permission.Permission, error) {
	return s.getPermissionsForRole(ctx, r)
}

// GetRolesForPermission returns all roles where a permission is used
func (s *Store) GetRolesForPermission(ctx context.Context, p permission.Permission) ([]role.Role, error) {
	return s.getRolesForPermission(ctx, p)
}

// GetRolePermission returns the role permission
func (s *Store) GetRolePermission(ctx context.Context, rp RolePermission) (RolePermission, error) {
	return s.getRolePermission(ctx, rp)
}

func (s *Store) GetPermissionsNotInRole(ctx context.Context, r role.Role) ([]permission.Permission, error) {
	return s.getPermissionsNotInRole(ctx, r)
}

func (s *Store) AddRolePermission(ctx context.Context, rp RolePermission) (RolePermission, error) {
	return s.addRolePermission(ctx, rp)
}

func (s *Store) RemoveRolePermission(ctx context.Context, rp RolePermission) error {
	return s.removeRolePermission(ctx, rp)
}
