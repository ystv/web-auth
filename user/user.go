package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Clarilab/gocloaksession"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"

	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/role"
	"github.com/ystv/web-auth/utils"
)

//go:generate mockgen -destination mocks/mock_user.go -package mock_user github.com/ystv/web-auth/user Repo

type (
	Repo interface {
		CountUsersAll(context.Context) (CountUsers, error)
		GetUser(context.Context, User) (User, error)
		GetUserValid(context.Context, User) (User, error)
		GetUsers(context.Context, int, int, string, string, string, string, string) ([]User, int, error)
		VerifyUser(context.Context, User) (User, bool, error)
		AddUser(context.Context, User, int) (User, error)
		EditUserPassword(context.Context, User) error
		EditUser(context.Context, User, int) error
		SetUserLoggedIn(context.Context, User) error
		EditUserAvatar(context.Context, User) error
		DeleteUser(context.Context, User, int) error
		GetPermissionsForUser(context.Context, User) ([]permission.Permission, error)
		GetRolesForUser(context.Context, User) ([]role.Role, error)
		GetUsersForRole(context.Context, role.Role) ([]User, error)
		GetRoleUser(context.Context, RoleUser) (RoleUser, error)
		GetUsersNotInRole(context.Context, role.Role) ([]User, error)
		AddRoleUser(context.Context, RoleUser) (RoleUser, error)
		RemoveRoleUser(context.Context, RoleUser) error
		RemoveUserForRoles(context.Context, User) error
		GetPermissionsForRole(context.Context, role.Role) ([]permission.Permission, error)
		GetRolesForPermission(context.Context, permission.Permission) ([]role.Role, error)
		GetRolePermission(context.Context, RolePermission) (RolePermission, error)
		GetPermissionsNotInRole(context.Context, role.Role) ([]permission.Permission, error)
		AddRolePermission(context.Context, RolePermission) (RolePermission, error)
		RemoveRolePermission(context.Context, RolePermission) error
	}

	// Store stores the dependencies
	Store struct {
		db          *sqlx.DB
		cdnEndpoint string
		cloak       *gocloaksession.GoCloakSession
	}

	// User represents relevant user fields
	//
	//nolint:musttag
	User struct {
		UserID             int                     `db:"user_id" json:"userID"`
		Username           string                  `db:"username" json:"username" schema:"username"`
		UniversityUsername string                  `db:"university_username" json:"universityUsername"`
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
		Authenticated      bool                    `json:"authenticated"`
		AssumedUser        *User                   `json:"assumedUser"`
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

	// DetailedUser is the user object in full for the front end
	DetailedUser struct {
		UserID             int                     `json:"id"`
		Username           string                  `json:"username"`
		UniversityUsername string                  `json:"universityUsername"`
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
		Officers           []OfficershipMember     `json:"officers"`
	}

	// OfficershipMember represents relevant officership member fields
	//
	//nolint:revive
	OfficershipMember struct {
		OfficershipMemberID int         `db:"officership_member_id" json:"officershipMemberID"`
		UserID              int         `db:"user_id" json:"userID"`
		OfficerID           int         `db:"officer_id" json:"officerID"`
		StartDate           null.Time   `db:"start_date" json:"startDate"`
		EndDate             null.Time   `db:"end_date" json:"endDate"`
		OfficershipName     string      `db:"officership_name" json:"officershipName"`
		UserName            string      `db:"user_name" json:"userName"`
		TeamID              null.Int    `db:"team_id" json:"teamID"`
		TeamName            null.String `db:"team_name" json:"teamName"`
	}

	CountUsers struct {
		TotalUsers             int `db:"total_users" json:"totalUsers"`
		ActiveUsers            int `db:"active_users" json:"activeUsers"`
		ActiveUsersPast24Hours int `db:"active_users_past_24_hours" json:"activeUsersPast24Hours"`
		ActiveUsersPastYear    int `db:"active_users_past_year" json:"activeUsersPastYear"`
	}

	RoleTemplate struct {
		RoleID      int
		Name        string
		Description string
		Permissions []permission.Permission
		Users       []User
	}

	// PermissionTemplate is for the front end of permission
	PermissionTemplate struct {
		PermissionID int
		Name         string
		Description  string
		Roles        []role.Role
	}

	// RolePermission symbolises a link between a role.Role and permission.Permission
	RolePermission struct {
		RoleID       int `db:"role_id" json:"roleID"`
		PermissionID int `db:"permission_id" json:"permissionID"`
	}

	// RoleUser symbolises a link between a role.Role and User
	RoleUser struct {
		RoleID int `db:"role_id" json:"roleID"`
		UserID int `db:"user_id" json:"userID"`
	}
)

var _ Repo = &Store{}

// NewUserRepo stores our dependency
func NewUserRepo(db *sqlx.DB, cdnEndpoint string) *Store {
	return &Store{
		db:          db,
		cloak:       nil,
		cdnEndpoint: cdnEndpoint,
	}
}

// CountUsersAll returns the number of users, active users, active users in the last 24 hours and past year
func (s *Store) CountUsersAll(ctx context.Context) (CountUsers, error) {
	return s.countUsersAll(ctx)
}

// GetUser returns a user using any unique identity fields
func (s *Store) GetUser(ctx context.Context, u User) (User, error) {
	return s.getUser(ctx, u)
}

// GetUserValid returns a user using any unique identity fields which is enabled and not deleted
func (s *Store) GetUserValid(ctx context.Context, u User) (User, error) {
	user, err := s.GetUser(ctx, u)
	if err != nil {
		return u, fmt.Errorf("failed to get user: %w", err)
	}

	if !user.Enabled {
		return u, errors.New("user not enabled, contact Computing Team for help")
	}

	if user.DeletedBy.Valid {
		return u, errors.New("user has been deleted, contact Computing Team for help")
	}

	if user.ResetPw {
		u.UserID = user.UserID

		return u, errors.New("password reset required")
	}

	return user, nil
}

func (s *Store) GetUsers(ctx context.Context, size, page int, search, sortBy, direction, enabled,
	deleted string) ([]User, int, error) {
	return s.getUsers(ctx, size, page, search, sortBy, direction, enabled, deleted)
}

// VerifyUser will check that the password is correct with provided
// credentials and if verified will return the User object
// returned is the user object, bool of if the password is forced to be changed and any errors encountered
func (s *Store) VerifyUser(ctx context.Context, u User) (User, bool, error) {
	user, err := s.GetUser(ctx, u)
	if err != nil {
		return u, false, fmt.Errorf("failed to get user: %w", err)
	}

	if !user.Enabled {
		return u, false, errors.New("user not enabled, contact Computing Team for help")
	}

	if user.DeletedBy.Valid {
		return u, false, errors.New("user has been deleted, contact Computing Team for help")
	}

	if utils.HashPass(user.Salt.String+u.Password.String) == user.Password.String {
		if user.ResetPw {
			u.UserID = user.UserID

			return user, true, errors.New("password reset required")
		}

		return user, false, nil
	}

	u.UseGravatar = user.UseGravatar
	u.Avatar = user.Avatar

	return u, false, errors.New("invalid credentials")
}

// AddUser adds a new User
func (s *Store) AddUser(ctx context.Context, u User, userID int) (User, error) {
	_, err := s.GetUser(ctx, u)
	if err == nil {
		return User{}, errors.New("failed to add user for addUser: user already exists")
	}

	u.Password = null.StringFrom(utils.HashPass(u.Salt.String + u.Password.String))
	u.ResetPw = true
	u.CreatedBy = null.IntFrom(int64(userID))
	u.CreatedAt = null.TimeFrom(time.Now())

	u, err = s.addUser(ctx, u)
	if err != nil {
		return User{}, fmt.Errorf("failed to add user for addUser: %w", err)
	}

	return u, nil
}

// EditUserPassword will edit the password and set the reset_pw to false
func (s *Store) EditUserPassword(ctx context.Context, u User) error {
	user, err := s.GetUser(ctx, u)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.Password = null.StringFrom(utils.HashPass(user.Salt.String + u.Password.String))
	user.ResetPw = false
	user.UpdatedBy = null.IntFrom(int64(user.UserID))
	user.UpdatedAt = null.TimeFrom(time.Now())

	err = s.editUser(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to edit user for editUserPassword: %w", err)
	}

	return nil
}

// EditUser will edit the user
func (s *Store) EditUser(ctx context.Context, u User, userID int) error {
	user, err := s.GetUser(ctx, u)
	if err != nil {
		return fmt.Errorf("failed to get user for editUser: %w", err)
	}

	if len(u.Username) > 0 {
		user.Username = u.Username
	}

	if len(u.UniversityUsername) > 0 {
		user.UniversityUsername = u.UniversityUsername
	}

	if len(u.LDAPUsername.String) > 0 {
		user.LDAPUsername = u.LDAPUsername
	}

	if len(u.LoginType) > 0 {
		user.LoginType = u.LoginType
	}

	if len(u.Nickname) > 0 {
		user.Nickname = u.Nickname
	}

	if len(u.Firstname) > 0 {
		user.Firstname = u.Firstname
	}

	if len(u.Lastname) > 0 {
		user.Lastname = u.Lastname
	}

	if len(u.Avatar) > 0 {
		user.Avatar = u.Avatar
	}

	if len(u.Email) > 0 {
		user.Email = u.Email
	}

	user.ResetPw = u.ResetPw
	user.Enabled = u.Enabled
	user.UseGravatar = u.UseGravatar
	user.UpdatedBy = null.IntFrom(int64(userID))
	user.UpdatedAt = null.TimeFrom(time.Now())

	err = s.editUser(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to edit user: %w", err)
	}

	return nil
}

// SetUserLoggedIn will set the last login date to now
func (s *Store) SetUserLoggedIn(ctx context.Context, u User) error {
	u.LastLogin = null.TimeFrom(time.Now())

	return s.editUser(ctx, u)
}

func (s *Store) EditUserAvatar(ctx context.Context, userParam User) error {
	user, err := s.getUser(ctx, userParam)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	user.UseGravatar = userParam.UseGravatar
	user.Avatar = userParam.Avatar
	err = s.editUser(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to edit user for edit user password: %w", err)
	}
	return nil
}

// DeleteUser will soft delete a user
func (s *Store) DeleteUser(ctx context.Context, u User, userID int) error {
	now := null.TimeFrom(time.Now())
	id := null.IntFrom(int64(userID))
	blank := null.NewString("", true)

	u.Username = fmt.Sprintf("deleted-%d", u.UserID)
	u.Firstname = "*Deleted*"
	u.Nickname = ""
	u.Lastname = "*User*"
	u.LDAPUsername = null.NewString("", false)
	u.Email = fmt.Sprintf("noreply+%d@ystv.co.uk", u.UserID)
	u.UniversityUsername = ""
	u.Enabled = false
	u.Avatar = ""
	u.UseGravatar = false
	u.Password = blank
	u.Salt = blank
	u.UpdatedBy = id
	u.UpdatedAt = now
	u.DeletedBy = id
	u.DeletedAt = now

	return s.editUser(ctx, u)
}

// GetPermissionsForUser returns all permissions of a user
func (s *Store) GetPermissionsForUser(ctx context.Context, u User) ([]permission.Permission, error) {
	return s.getPermissionsForUser(ctx, u)
}

// GetRolesForUser returns all roles of a user
func (s *Store) GetRolesForUser(ctx context.Context, u User) ([]role.Role, error) {
	return s.getRolesForUser(ctx, u)
}

// GetUsersForRole returns all the Users that are linked to a role.Role
func (s *Store) GetUsersForRole(ctx context.Context, r role.Role) ([]User, error) {
	return s.getUsersForRole(ctx, r)
}

// GetRoleUser returns a single link between a role.Role and User
func (s *Store) GetRoleUser(ctx context.Context, ru RoleUser) (RoleUser, error) {
	return s.getRoleUser(ctx, ru)
}

// GetUsersNotInRole returns all the users not linked to a role.Role
func (s *Store) GetUsersNotInRole(ctx context.Context, r role.Role) ([]User, error) {
	return s.getUsersNotInRole(ctx, r)
}

// AddRoleUser adds a link between a role.Role and User
func (s *Store) AddRoleUser(ctx context.Context, ru RoleUser) (RoleUser, error) {
	return s.addRoleUser(ctx, ru)
}

// RemoveRoleUser removes a link between a role.Role and User
func (s *Store) RemoveRoleUser(ctx context.Context, ru RoleUser) error {
	return s.removeRoleUser(ctx, ru)
}

// RemoveUserForRoles removes links between a User and Roles
func (s *Store) RemoveUserForRoles(ctx context.Context, u User) error {
	return s.removeUserForRoles(ctx, u)
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

// GetPermissionsNotInRole returns all the permission.Permission not in a role.Role
func (s *Store) GetPermissionsNotInRole(ctx context.Context, r role.Role) ([]permission.Permission, error) {
	return s.getPermissionsNotInRole(ctx, r)
}

// AddRolePermission creates a link between a role.Role and permission.Permission
func (s *Store) AddRolePermission(ctx context.Context, rp RolePermission) (RolePermission, error) {
	return s.addRolePermission(ctx, rp)
}

// RemoveRolePermission removes a link between a role.Role and permission.Permission
func (s *Store) RemoveRolePermission(ctx context.Context, rp RolePermission) error {
	return s.removeRolePermission(ctx, rp)
}
