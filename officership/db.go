package officership

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/ystv/web-auth/user"
	"github.com/ystv/web-auth/utils"
)

func (s *Store) countOfficerships(ctx context.Context) (CountOfficerships, error) {
	var countOfficerships CountOfficerships
	err := s.db.GetContext(ctx, &countOfficerships,
		`SELECT
		(SELECT COUNT(*) FROM people.officerships) as total_officerships,
		(SELECT COUNT(*) FROM people.officerships WHERE is_current = true) as current_officerships,
		(SELECT COUNT(*) FROM people.officership_members) as total_officers,
		(SELECT COUNT(*) FROM people.officership_members WHERE end_date IS NULL) as current_officers;`)
	if err != nil {
		return countOfficerships, fmt.Errorf("failed to count officerships all from db: %w", err)
	}
	return countOfficerships, nil
}

func (s *Store) getOfficerships(ctx context.Context, officershipStatus OfficershipsStatus) ([]Officership, error) {
	var o []Officership
	builder := utils.PSQL().Select("o.*", "COUNT(DISTINCT omc.officership_member_id) AS current_officers", "COUNT(DISTINCT omp.officership_member_id) AS previous_officers", "otm.team_id AS team_id", "ot.name AS team_name").
		From("people.officerships o").
		LeftJoin("people.officership_members omc ON o.officer_id = omc.officer_id AND omc.end_date IS NULL").
		LeftJoin("people.officership_members omp ON o.officer_id = omp.officer_id AND omp.end_date IS NOT NULL").
		LeftJoin("people.officership_team_members otm ON o.officer_id = otm.officer_id").
		LeftJoin("people.officership_teams ot ON ot.team_id = otm.team_id").
		GroupBy("o", "o.officer_id", "o.name", "o.email_alias", "description", "historywiki_url", "role_id", "is_current", "if_unfilled", "otm.team_id", "ot.name")
	switch officershipStatus {
	case Any:
	case Current:
		builder = builder.Where("o.is_current = true")
	case Previous:
		builder = builder.Where("o.is_current = false")
	}
	builder = builder.GroupBy("o", "o.officer_id", "o.name", "o.email_alias", "description", "historywiki_url", "role_id", "is_current", "if_unfilled").
		OrderBy("CASE WHEN o.name = 'Station Director' THEN 0 WHEN o.name LIKE '%Director%' AND o.name NOT LIKE '%Deputy%' AND o.name NOT LIKE '%Assistant%' THEN 1 WHEN o.name LIKE '%Deputy%' THEN 2 WHEN o.name LIKE '%Assistant%' THEN 3 WHEN o.name = 'Head of Welfare and Training' THEN 4 WHEN o.name LIKE '%Head of%' THEN 5 ELSE 6 END, o.name")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getOfficerships: %w", err))
	}
	err = s.db.SelectContext(ctx, &o, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get officerships: %w", err)
	}
	return o, nil
}

func (s *Store) getOfficership(ctx context.Context, o1 Officership) (Officership, error) {
	var o Officership
	builder := utils.PSQL().Select("o.*", "COUNT(DISTINCT omc.officership_member_id) AS current_officers", "COUNT(DISTINCT omp.officership_member_id) AS previous_officers", "otm.team_id AS team_id", "ot.name AS team_name").
		From("people.officerships o").
		LeftJoin("people.officership_members omc ON o.officer_id = omc.officer_id AND omc.end_date IS NULL").
		LeftJoin("people.officership_members omp ON o.officer_id = omp.officer_id AND omp.end_date IS NOT NULL").
		LeftJoin("people.officership_team_members otm ON o.officer_id = otm.officer_id").
		LeftJoin("people.officership_teams ot ON ot.team_id = otm.team_id").
		Where(sq.Or{sq.Eq{"o.officer_id": o1.OfficershipID}, sq.And{sq.Eq{"o.name": o1.Name}, sq.NotEq{"o.name": ""}}}).
		GroupBy("o", "o.officer_id", "o.name", "o.email_alias", "description", "historywiki_url", "role_id", "is_current", "if_unfilled", "otm.team_id", "ot.name").
		OrderBy("CASE WHEN o.name = 'Station Director' THEN 0 WHEN o.name LIKE '%Director%' AND o.name NOT LIKE '%Deputy%' AND o.name NOT LIKE '%Assistant%' THEN 1 WHEN o.name LIKE '%Deputy%' THEN 2 WHEN o.name LIKE '%Assistant%' THEN 3 WHEN o.name = 'Head of Welfare and Training' THEN 4 WHEN o.name LIKE '%Head of%' THEN 5 ELSE 6 END, o.name").
		Limit(1)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getOfficership: %w", err))
	}
	err = s.db.GetContext(ctx, &o, sql, args...)
	if err != nil {
		return Officership{}, fmt.Errorf("failed to get officership: %w", err)
	}
	return o, nil
}

func (s *Store) addOfficership(ctx context.Context, o1 Officership) (Officership, error) {
	var o Officership
	stmt, err := s.db.PrepareNamedContext(ctx, "INSERT INTO people.officerships (name, email_alias, description, historywiki_url, role_id, is_current, if_unfilled) VALUES (:name, :email_alias, :description, :historywiki_url, :role_id, :is_current, :if_unfilled) RETURNING officer_id, name, email_alias, description, historywiki_url, role_id, is_current, if_unfilled")
	if err != nil {
		return Officership{}, fmt.Errorf("failed to add officership: %w", err)
	}
	defer stmt.Close()
	err = stmt.Get(&o, o1)
	if err != nil {
		return Officership{}, fmt.Errorf("failed to add officership: %w", err)
	}
	return o, nil
}

func (s *Store) editOfficership(ctx context.Context, o Officership) (Officership, error) {
	builder := utils.PSQL().Update("people.officerships").
		SetMap(map[string]interface{}{
			"name":            o.Name,
			"email_alias":     o.EmailAlias,
			"description":     o.Description,
			"historywiki_url": o.HistoryWikiURL,
			"role_id":         o.RoleID,
			"is_current":      o.IsCurrent,
			"if_unfilled":     o.IfUnfilled,
		}).
		Where(sq.Eq{"officer_id": o.OfficershipID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for editOfficership: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Officership{}, fmt.Errorf("failed to edit officership: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return Officership{}, fmt.Errorf("failed to edit officership: %w", err)
	}
	if rows < 1 {
		return Officership{}, fmt.Errorf("failed to edit officerhip: invalid rows affected: %d", rows)
	}
	return o, nil
}

func (s *Store) deleteOfficership(ctx context.Context, o Officership) error {
	builder := utils.PSQL().Delete("people.officership_members").
		Where(sq.Eq{"officer_id": o.OfficershipID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deleteOfficership: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete officership: %w", err)
	}
	return nil
}

func (s *Store) getOfficershipTeams(ctx context.Context) ([]OfficershipTeam, error) {
	var t []OfficershipTeam
	builder := utils.PSQL().Select("ot.*", "COUNT(DISTINCT otm.officer_id) AS current_officerships", "COUNT(DISTINCT om.officership_member_id) AS current_officers").
		From("people.officership_teams ot").
		LeftJoin("people.officership_team_members otm ON ot.team_id = otm.team_id").
		LeftJoin("people.officerships o ON otm.officer_id = o.officer_id").
		LeftJoin("people.officership_members om ON o.officer_id = om.officer_id AND om.end_date IS NULL").
		Where("o.is_current = true").
		GroupBy("ot", "ot.team_id", "ot.name", "ot.email_alias", "short_description", "full_description")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getOfficershipTeams: %w", err))
	}
	err = s.db.SelectContext(ctx, &t, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get officership teams: %w", err)
	}
	return t, nil
}

func (s *Store) getOfficershipTeam(ctx context.Context, t1 OfficershipTeam) (OfficershipTeam, error) {
	var t OfficershipTeam
	builder := utils.PSQL().Select("ot.*").
		From("people.officership_teams ot").
		Where(sq.Or{sq.Eq{"ot.team_id": t1.TeamID}, sq.And{sq.Eq{"ot.name": t1.Name}, sq.NotEq{"ot.name": ""}}}).
		Limit(1)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getOfficershipTeam: %w", err))
	}
	err = s.db.GetContext(ctx, &t, sql, args...)
	if err != nil {
		return OfficershipTeam{}, fmt.Errorf("failed to get officership team: %w", err)
	}
	return t, nil
}

func (s *Store) addOfficershipTeam(ctx context.Context, t1 OfficershipTeam) (OfficershipTeam, error) {
	var t OfficershipTeam
	stmt, err := s.db.PrepareNamedContext(ctx, "INSERT INTO people.officership_teams (name, email_alias, short_description, full_description) VALUES (:name, :email_alias, :short_description, :full_description) RETURNING team_id, name, email_alias, short_description, full_description")
	if err != nil {
		return OfficershipTeam{}, fmt.Errorf("failed to add officership team: %w", err)
	}
	defer stmt.Close()
	err = stmt.Get(&t, t1)
	if err != nil {
		return OfficershipTeam{}, fmt.Errorf("failed to add officership team: %w", err)
	}
	return t, nil
}

func (s *Store) editOfficershipTeam(ctx context.Context, t OfficershipTeam) (OfficershipTeam, error) {
	builder := utils.PSQL().Update("people.officership_teams").
		SetMap(map[string]interface{}{
			"name":              t.Name,
			"email_alias":       t.EmailAlias,
			"short_description": t.ShortDescription,
			"full_description":  t.FullDescription,
		}).
		Where(sq.Eq{"team_id": t.TeamID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for editOfficershipTeam: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return OfficershipTeam{}, fmt.Errorf("failed to edit officership team: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return OfficershipTeam{}, fmt.Errorf("failed to edit officership team: %w", err)
	}
	if rows < 1 {
		return OfficershipTeam{}, fmt.Errorf("failed to edit officerhip team: invalid rows affected: %d", rows)
	}
	return t, nil
}

func (s *Store) deleteOfficershipTeam(ctx context.Context, t OfficershipTeam) error {
	builder := utils.PSQL().Delete("people.officership_members").
		Where(sq.Eq{"team_id": t.TeamID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deleteOfficershipTeam: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete officership team: %w", err)
	}
	return nil
}

func (s *Store) getOfficershipTeamMembers(ctx context.Context, t1 *OfficershipTeam, officershipStatus OfficershipsStatus) ([]OfficershipTeamMember, error) {
	var m []OfficershipTeamMember
	builder := utils.PSQL().Select("otm.*", "o.name AS officer_name", "COUNT(DISTINCT omc.officership_member_id) AS current_officers", "COUNT(DISTINCT omp.officership_member_id) AS previous_officers").
		From("people.officership_team_members otm").
		LeftJoin("people.officerships o on o.officer_id = otm.officer_id").
		LeftJoin("people.officership_members omc ON o.officer_id = omc.officer_id AND omc.end_date IS NULL").
		LeftJoin("people.officership_members omp ON o.officer_id = omp.officer_id AND omp.end_date IS NOT NULL")
	if t1 != nil {
		builder = builder.Where(sq.Eq{"otm.team_id": t1.TeamID})
	}
	switch officershipStatus {
	case Any:
	case Current:
		builder = builder.Where("o.is_current = true")
	case Previous:
		builder = builder.Where("o.is_current = false")
	}
	builder = builder.OrderBy("CASE WHEN o.name = 'Station Director' THEN 0 WHEN o.name LIKE '%Director%' AND o.name NOT LIKE '%Deputy%' THEN 1 WHEN o.name LIKE '%Deputy%' THEN 2 WHEN o.name = 'Head of Welfare and Training' THEN 3 WHEN o.name LIKE '%Head of%' THEN 4 ELSE 5 END",
		"o.name").
		GroupBy("otm", "otm.officer_id", "otm.team_id", "o.officer_id", "name", "email_alias", "description", "historywiki_url", "role_id", "is_current", "if_unfilled")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getOfficershipTeamMembers: %w", err))
	}
	err = s.db.SelectContext(ctx, &m, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get officership team members: %w", err)
	}
	return m, nil
}

func (s *Store) getOfficershipTeamMember(ctx context.Context, m1 OfficershipTeamMember) (OfficershipTeamMember, error) {
	var m OfficershipTeamMember
	builder := utils.PSQL().Select("otm.*", "o.name AS officer_name", "COUNT(DISTINCT omc.officership_member_id) AS current_officers", "COUNT(DISTINCT omp.officership_member_id) AS previous_officers").
		From("people.officership_team_members otm").
		LeftJoin("people.officerships o on o.officer_id = otm.officer_id").
		LeftJoin("people.officership_members omc ON o.officer_id = omc.officer_id AND omc.end_date IS NULL").
		LeftJoin("people.officership_members omp ON o.officer_id = omp.officer_id AND omp.end_date IS NOT NULL").
		Where(sq.Eq{"otm.team_id": m1.OfficershipTeamMemberID}).
		GroupBy("otm", "otm.officer_id", "otm.team_id", "o.officer_id", "name", "email_alias", "description", "historywiki_url", "role_id", "is_current", "if_unfilled")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getOfficershipTeamMember: %w", err))
	}
	err = s.db.SelectContext(ctx, &m, sql, args...)
	if err != nil {
		return OfficershipTeamMember{}, fmt.Errorf("failed to get officership team member: %w", err)
	}
	return m, nil
}

func (s *Store) addOfficershipTeamMember(ctx context.Context, m1 OfficershipTeamMember) (OfficershipTeamMember, error) {
	var m OfficershipTeamMember
	stmt, err := s.db.PrepareNamedContext(ctx, "INSERT INTO people.officership_team_members (team_id, officer_id, is_leader, is_deputy) VALUES (:team_id, :officer_id, :is_leader, :is_deputy) RETURNING team_id, officer_id, is_leader, is_deputy")
	if err != nil {
		return OfficershipTeamMember{}, fmt.Errorf("failed to add officership team member: %w", err)
	}
	defer stmt.Close()
	err = stmt.Get(&m, m1)
	if err != nil {
		return OfficershipTeamMember{}, fmt.Errorf("failed to add officership team member: %w", err)
	}
	return m, nil
}

func (s *Store) editOfficershipTeamMember(ctx context.Context, m OfficershipTeamMember) (OfficershipTeamMember, error) {
	builder := utils.PSQL().Update("people.officership_team_members").
		SetMap(map[string]interface{}{
			"team_id":    m.OfficershipTeamMemberID,
			"officer_id": m.OfficerID,
			"is_leader":  m.IsLeader,
			"is_deputy":  m.IsDeputy,
		}).
		Where(sq.And{sq.Eq{"team_id": m.OfficershipTeamMemberID}, sq.Eq{"officer_id": m.OfficerID}})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for editOfficershipTeamMember: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return OfficershipTeamMember{}, fmt.Errorf("failed to edit officership team member: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return OfficershipTeamMember{}, fmt.Errorf("failed to edit officership team member: %w", err)
	}
	if rows < 1 {
		return OfficershipTeamMember{}, fmt.Errorf("failed to edit officerhip team member: invalid rows affected: %d", rows)
	}
	return m, nil
}

func (s *Store) deleteOfficershipTeamMember(ctx context.Context, t OfficershipTeamMember) error {
	builder := utils.PSQL().Delete("people.officership_team_members").
		Where(sq.And{sq.Eq{"team_id": t.OfficershipTeamMemberID}, sq.Eq{"officer_id": t.OfficerID}})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deleteOfficershipTeam: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete officership team: %w", err)
	}
	return nil
}

func (s *Store) removeTeamForOfficershipMemberTeams(ctx context.Context, t OfficershipTeam) error {
	builder := utils.PSQL().Delete("people.officership_team_members").
		Where(sq.Eq{"team_id": t.TeamID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for removeTeamForOfficershipMemberTeams: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete removeTeamForOfficershipMemberTeams: %w", err)
	}
	return nil
}

func (s *Store) getOfficershipMembers(ctx context.Context, o1 *Officership, officershipStatus OfficershipsStatus, officershipMemberStatus OfficershipsStatus) ([]OfficershipMember, error) {
	var o []OfficershipMember
	builder := utils.PSQL().Select("om.*", "o.name AS officership_name", "CONCAT(u.first_name, ' ', u.last_name) AS user_name", "otm.team_id AS team_id", "ot.name AS team_name").
		From("people.officership_members om").
		LeftJoin("people.officerships o ON o.officer_id = om.officer_id").
		LeftJoin("people.officership_team_members otm ON otm.officer_id = om.officer_id").
		LeftJoin("people.officership_teams ot ON ot.team_id = otm.team_id").
		LeftJoin("people.users u ON u.user_id = om.user_id")
	if o1 != nil {
		builder = builder.Where(sq.Or{sq.Eq{"o.officer_id": o1.OfficershipID}, sq.And{sq.Eq{"o.name": o1.Name}, sq.NotEq{"o.name": ""}}})
	}
	switch officershipStatus {
	case Any:
	case Current:
		builder = builder.Where("o.is_current = true")
	case Previous:
		builder = builder.Where("o.is_current = false")
	}
	switch officershipMemberStatus {
	case Any:
	case Current:
		builder = builder.Where("om.end_date IS NULL")
	case Previous:
		builder = builder.Where("om.end_date IS NOT NULL")
	}
	builder = builder.OrderBy("CASE WHEN o.name = 'Station Director' THEN 0 WHEN o.name LIKE '%Director%' AND o.name NOT LIKE '%Deputy%' THEN 1 WHEN o.name LIKE '%Deputy%' THEN 2 WHEN o.name = 'Head of Welfare and Training' THEN 3 WHEN o.name LIKE '%Head of%' THEN 4 ELSE 5 END",
		"o.name",
		"om.start_date")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getOfficershipMembers: %w", err))
	}
	err = s.db.SelectContext(ctx, &o, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get officership members: %w", err)
	}
	return o, nil
}

func (s *Store) getOfficershipMember(ctx context.Context, m1 OfficershipMember) (OfficershipMember, error) {
	var m OfficershipMember
	builder := utils.PSQL().Select("om.*", "o.name AS officership_name", "CONCAT(u.first_name, ' ', u.last_name) AS user_name", "otm.team_id AS team_id", "ot.name AS team_name").
		From("people.officership_members om").
		LeftJoin("people.officerships o ON o.officer_id = om.officer_id").
		LeftJoin("people.officership_team_members otm ON otm.officer_id = om.officer_id").
		LeftJoin("people.officership_teams ot ON ot.team_id = otm.team_id").
		LeftJoin("people.users u ON u.user_id = om.user_id").
		Where(sq.Eq{"om.officership_member_id": m1.OfficershipMemberID}).
		Limit(1)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getOfficershipMember: %w", err))
	}
	err = s.db.GetContext(ctx, &m, sql, args...)
	if err != nil {
		return OfficershipMember{}, fmt.Errorf("failed to get officership member: %w", err)
	}
	return m, nil
}

func (s *Store) addOfficershipMember(ctx context.Context, m1 OfficershipMember) (OfficershipMember, error) {
	var m OfficershipMember
	stmt, err := s.db.PrepareNamedContext(ctx, "INSERT INTO people.officership_members (user_id, officer_id, start_date, end_date) VALUES (:user_id, :officer_id, :start_date, :end_date) RETURNING officership_member_id, user_id, officer_id, start_date, end_date")
	if err != nil {
		return OfficershipMember{}, fmt.Errorf("failed to add officership member: %w", err)
	}
	defer stmt.Close()
	err = stmt.Get(&m, m1)
	if err != nil {
		return OfficershipMember{}, fmt.Errorf("failed to add officership member: %w", err)
	}
	return m, nil
}

func (s *Store) editOfficershipMember(ctx context.Context, m OfficershipMember) (OfficershipMember, error) {
	builder := utils.PSQL().Update("people.officership_members").
		SetMap(map[string]interface{}{
			"user_id":    m.UserID,
			"officer_id": m.OfficerID,
			"start_date": m.StartDate,
			"end_date":   m.EndDate,
		}).
		Where(sq.Eq{"officership_member_id": m.OfficershipMemberID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for editOfficershipMember: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return OfficershipMember{}, fmt.Errorf("failed to edit officership member: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return OfficershipMember{}, fmt.Errorf("failed to edit officership member: %w", err)
	}
	if rows < 1 {
		return OfficershipMember{}, fmt.Errorf("failed to edit officerhip member: invalid rows affected: %d", rows)
	}
	return m, nil
}

func (s *Store) deleteOfficershipMember(ctx context.Context, m OfficershipMember) error {
	builder := utils.PSQL().Delete("people.officership_members").
		Where(sq.Eq{"officership_member_id": m.OfficershipMemberID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deleteOfficershipMember: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete officership member: %w", err)
	}
	return nil
}

func (s *Store) removeOfficershipForOfficershipMembers(ctx context.Context, o Officership) error {
	builder := utils.PSQL().Delete("people.officership_members").
		Where(sq.Eq{"officer_id": o.OfficershipID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for removeOfficershipForOfficershipMembers: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete removeOfficershipForOfficershipMembers: %w", err)
	}
	return nil
}

func (s *Store) removeUserForOfficershipMembers(ctx context.Context, u user.User) error {
	builder := utils.PSQL().Delete("people.officership_members").
		Where(sq.Eq{"user_id": u.UserID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for removeUsersForOfficershipMembers: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete removeUsersForOfficershipMembers: %w", err)
	}
	return nil
}
