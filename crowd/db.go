package crowd

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/ystv/web-auth/utils"
)

func (s *Store) getCrowdApp(ctx context.Context, c1 CrowdApp) (CrowdApp, error) {
	var c CrowdApp

	builder := utils.PSQL().Select("*").
		From("web_auth.crowd_apps").
		Where(sq.Or{
			sq.Eq{"app_id": c1.AppID},
			sq.Eq{"name": c1.Name}}).
		Limit(1)

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getCrowdApp: %w", err))
	}

	//nolint:musttag
	err = s.db.GetContext(ctx, &c, sql, args...)
	if err != nil {
		return c, fmt.Errorf("failed to get crowd app from db: %w", err)
	}

	return c, nil
}

func (s *Store) getCrowdApps(ctx context.Context, crowdAppStatus CrowdAppStatus) ([]CrowdApp, error) {
	var c []CrowdApp

	builder := utils.PSQL().Select("app_id", "name", "description", "active").
		From("web_auth.crowd_apps")

	switch crowdAppStatus {
	case Any:
	case Active:
		builder = builder.Where("active = true")
	case Inactive:
		builder = builder.Where("active = false")
	}

	builder = builder.OrderBy("name")

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getCrowdApps: %w", err))
	}

	err = s.db.SelectContext(ctx, &c, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get crowd apps: %w", err)
	}

	return c, nil
}
