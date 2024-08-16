package officership

import (
	"context"

	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"

	"github.com/ystv/web-auth/user"
)

type (
	// Store stores the dependencies
	Store struct {
		db *sqlx.DB
	}

	// Officership represents relevant officership fields
	Officership struct {
		OfficershipID    int         `db:"officer_id" json:"officershipID"`
		Name             string      `db:"name" json:"name"`
		EmailAlias       string      `db:"email_alias" json:"emailAlias"`
		Description      string      `db:"description" json:"description"`
		HistoryWikiURL   string      `db:"historywiki_url" json:"historyWikiURL"`
		RoleID           null.Int    `db:"role_id" json:"roleID,omitempty"`
		IsCurrent        bool        `db:"is_current" json:"isCurrent"`
		IfUnfilled       null.Bool   `db:"if_unfilled" json:"ifUnfilled,omitempty"`
		CurrentOfficers  int         `db:"current_officers" json:"currentOfficers,omitempty"`
		PreviousOfficers int         `db:"previous_officers" json:"previousOfficers,omitempty"`
		TeamID           null.Int    `db:"team_id" json:"teamID"`
		TeamName         null.String `db:"team_name" json:"teamName"`
		IsTeamLeader     null.Bool   `db:"is_team_leader" json:"isTeamLeader"`
		IsTeamDeputy     null.Bool   `db:"is_team_deputy" json:"isTeamDeputy"`
	}

	// OfficershipsStatus indicates the state desired for a database get of officers
	OfficershipsStatus int

	// OfficershipTeam represents relevant officership team fields
	//
	//nolint:revive
	OfficershipTeam struct {
		TeamID              int    `db:"team_id" json:"teamID"`
		Name                string `db:"name" json:"name"`
		EmailAlias          string `db:"email_alias" json:"emailAlias"`
		ShortDescription    string `db:"short_description" json:"shortDescription"`
		FullDescription     string `db:"full_description" json:"fullDescription"`
		CurrentOfficerships int    `db:"current_officerships" json:"currentOfficerships"`
		CurrentOfficers     int    `db:"current_officers" json:"currentOfficers"`
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

	// OfficershipTeamMember represents relevant officership team member fields
	//
	//nolint:revive
	OfficershipTeamMember struct {
		TeamID           int    `db:"team_id" json:"officershipTeamMemberID"`
		OfficerID        int    `db:"officer_id" json:"officerID"`
		IsLeader         bool   `db:"is_leader" json:"isLeader"`
		IsDeputy         bool   `db:"is_deputy" json:"isDeputy"`
		IsCurrent        bool   `db:"is_current" json:"isCurrent"`
		OfficerName      string `db:"officer_name" json:"officerName"`
		CurrentOfficers  int    `db:"current_officers" json:"currentOfficers"`
		PreviousOfficers int    `db:"previous_officers" json:"previousOfficers"`
	}

	CountOfficerships struct {
		TotalOfficerships   int `db:"total_officerships" json:"totalOfficerships"`
		CurrentOfficerships int `db:"current_officerships" json:"currentOfficerships"`
		TotalOfficers       int `db:"total_officers" json:"totalOfficers"`
		CurrentOfficers     int `db:"current_officers" json:"currentOfficers"`
	}
)

const (
	Any OfficershipsStatus = iota
	Retired
	Current
)

// NewOfficershipRepo stores our dependency
func NewOfficershipRepo(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) CountOfficerships(ctx context.Context) (CountOfficerships, error) {
	return s.countOfficerships(ctx)
}

func (s *Store) GetOfficerships(ctx context.Context, officershipStatus OfficershipsStatus) ([]Officership, error) {
	return s.getOfficerships(ctx, officershipStatus)
}

func (s *Store) GetOfficership(ctx context.Context, o Officership) (Officership, error) {
	return s.getOfficership(ctx, o)
}

func (s *Store) AddOfficership(ctx context.Context, o Officership) (Officership, error) {
	return s.addOfficership(ctx, o)
}

func (s *Store) EditOfficership(ctx context.Context, o Officership) (Officership, error) {
	return s.editOfficership(ctx, o)
}

func (s *Store) DeleteOfficership(ctx context.Context, o Officership) error {
	return s.deleteOfficership(ctx, o)
}

func (s *Store) GetOfficershipTeams(ctx context.Context) ([]OfficershipTeam, error) {
	return s.getOfficershipTeams(ctx)
}

func (s *Store) GetOfficershipTeam(ctx context.Context, t OfficershipTeam) (OfficershipTeam, error) {
	return s.getOfficershipTeam(ctx, t)
}

func (s *Store) AddOfficershipTeam(ctx context.Context, t OfficershipTeam) (OfficershipTeam, error) {
	return s.addOfficershipTeam(ctx, t)
}

func (s *Store) EditOfficershipTeam(ctx context.Context, t OfficershipTeam) (OfficershipTeam, error) {
	return s.editOfficershipTeam(ctx, t)
}

func (s *Store) DeleteOfficershipTeam(ctx context.Context, t OfficershipTeam) error {
	return s.deleteOfficershipTeam(ctx, t)
}

func (s *Store) GetOfficershipTeamMembers(ctx context.Context, t *OfficershipTeam,
	officershipStatus OfficershipsStatus) ([]OfficershipTeamMember, error) {
	return s.getOfficershipTeamMembers(ctx, t, officershipStatus)
}

func (s *Store) GetOfficershipTeamMember(ctx context.Context, m OfficershipTeamMember) (OfficershipTeamMember, error) {
	return s.getOfficershipTeamMember(ctx, m)
}

func (s *Store) AddOfficershipTeamMember(ctx context.Context, m OfficershipTeamMember) (OfficershipTeamMember, error) {
	return s.addOfficershipTeamMember(ctx, m)
}

func (s *Store) DeleteOfficershipTeamMember(ctx context.Context, m OfficershipTeamMember) error {
	return s.deleteOfficershipTeamMember(ctx, m)
}

func (s *Store) RemoveTeamForOfficershipTeamMembers(ctx context.Context, t OfficershipTeam) error {
	return s.removeTeamForOfficershipTeamMembers(ctx, t)
}

func (s *Store) GetOfficershipMembers(ctx context.Context, o *Officership, u *user.User, officershipStatus,
	officershipMemberStatus OfficershipsStatus, orderByOfficerName bool) ([]OfficershipMember, error) {
	return s.getOfficershipMembers(ctx, o, u, officershipStatus, officershipMemberStatus, orderByOfficerName)
}

func (s *Store) GetOfficershipMember(ctx context.Context, m OfficershipMember) (OfficershipMember, error) {
	return s.getOfficershipMember(ctx, m)
}

func (s *Store) AddOfficershipMember(ctx context.Context, m OfficershipMember) (OfficershipMember, error) {
	return s.addOfficershipMember(ctx, m)
}

func (s *Store) EditOfficershipMember(ctx context.Context, m OfficershipMember) (OfficershipMember, error) {
	return s.editOfficershipMember(ctx, m)
}

func (s *Store) DeleteOfficershipMember(ctx context.Context, m OfficershipMember) error {
	return s.deleteOfficershipMember(ctx, m)
}

func (s *Store) RemoveOfficershipForOfficershipMembers(ctx context.Context, o Officership) error {
	return s.removeOfficershipForOfficershipMembers(ctx, o)
}

func (s *Store) RemoveUserForOfficershipMembers(ctx context.Context, u user.User) error {
	return s.removeUserForOfficershipMembers(ctx, u)
}
