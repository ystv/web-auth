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
			sq.Eq{"username": c1.Username}}).
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

	builder := utils.PSQL().Select("app_id", "name", "username", "description", "active").
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

func (s *Store) addCrowdApp(ctx context.Context, c CrowdApp) (CrowdApp, error) {
	builder := utils.PSQL().Insert("web_auth.crowd_apps").
		Columns("name", "username", "description", "active", "password", "salt").
		Values(c.Name, c.Username, c.Description, c.Active, c.Password, c.Salt).
		Suffix("RETURNING app_id")

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for addCrowdApp: %w", err))
	}

	stmt, err := s.db.PrepareContext(ctx, sql)
	if err != nil {
		return CrowdApp{}, fmt.Errorf("failed to add crowd app: %w", err)
	}

	defer stmt.Close()

	err = stmt.QueryRow(args...).Scan(&c.AppID)
	if err != nil {
		return CrowdApp{}, fmt.Errorf("failed to add crowd app: %w", err)
	}

	return c, nil
}

func (s *Store) editCrowdApp(ctx context.Context, c CrowdApp) (CrowdApp, error) {
	builder := utils.PSQL().Update("web_auth.crowd_apps").
		SetMap(map[string]interface{}{
			"name":        c.Name,
			"description": c.Description,
			"active":      c.Active,
		}).
		Where(sq.Eq{"app_id": c.AppID})

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for editCrowdApp: %w", err))
	}

	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return CrowdApp{}, fmt.Errorf("failed to edit crowd app: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return CrowdApp{}, fmt.Errorf("failed to edit crowd app: %w", err)
	}

	if rows < 1 {
		return CrowdApp{}, fmt.Errorf("failed to edit crowd app: invalid rows affected: %d", rows)
	}

	return c, nil
}

func (s *Store) deleteCrowdApp(ctx context.Context, c CrowdApp) error {
	builder := utils.PSQL().Delete("web_auth.crowd_apps").
		Where(sq.Eq{"app_id": c.AppID})

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deleteCrowdApp: %w", err))
	}

	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete crowd app: %w", err)
	}

	return nil
}
