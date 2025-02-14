package crowd

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"

	"github.com/ystv/web-auth/utils"
)

//go:generate mockgen -destination mocks/mock_crowd.go -package mock_crowd github.com/ystv/web-auth/crowd Repo

type (
	Repo interface {
		GetCrowdApp(context.Context, CrowdApp) (CrowdApp, error)
		GetCrowdApps(context.Context, CrowdAppStatus) ([]CrowdApp, error)
		VerifyCrowd(context.Context, CrowdApp) (CrowdApp, error)
	}

	// Store stores the dependencies
	Store struct {
		db *sqlx.DB
	}

	//nolint:revive
	CrowdApp struct {
		AppID       int         `db:"app_id" json:"appID"`
		Name        string      `db:"name" json:"name"`
		Description null.String `db:"description" json:"description,omitempty"`
		Active      bool        `db:"active" json:"active"`
		Password    null.String `db:"password" json:"-"`
		Salt        null.String `db:"salt" json:"-"`
	}

	// CrowdAppStatus indicates the state desired for a database get of crowd apps
	//nolint:revive
	CrowdAppStatus int
)

const (
	Any CrowdAppStatus = iota
	Inactive
	Active
)

var _ Repo = &Store{}

// NewCrowdRepo stores our dependency
func NewCrowdRepo(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) GetCrowdApp(ctx context.Context, c CrowdApp) (CrowdApp, error) {
	return s.getCrowdApp(ctx, c)
}

func (s *Store) GetCrowdApps(ctx context.Context, crowdAppStatus CrowdAppStatus) ([]CrowdApp, error) {
	return s.getCrowdApps(ctx, crowdAppStatus)
}

// VerifyCrowd will check that the password is correct with provided
// credentials and if verified will return the CrowdApp object
// returned is the user object
func (s *Store) VerifyCrowd(ctx context.Context, c CrowdApp) (CrowdApp, error) {
	crowd, err := s.GetCrowdApp(ctx, c)
	if err != nil {
		return c, fmt.Errorf("failed to get crowd app: %w", err)
	}

	if !crowd.Active {
		return c, errors.New("crowd app not active")
	}

	if utils.HashPass(crowd.Salt.String+c.Password.String) == crowd.Password.String {
		return crowd, nil
	}

	return c, errors.New("invalid credentials")
}
