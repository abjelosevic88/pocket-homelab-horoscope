package store

import (
	"context"
	"database/sql"
	"strings"
)

const personCols = `id, user_id, name, relation, is_me, birth_date, birth_time, place, country, lat, lon, tz, color, notes, favorite, source, external_id, created_at, updated_at`

func scanPerson(row interface{ Scan(...any) error }) (*Person, error) {
	var p Person
	var me, fav int
	if err := row.Scan(&p.ID, &p.UserID, &p.Name, &p.Relation, &me, &p.BirthDate, &p.BirthTime, &p.Place, &p.Country, &p.Lat, &p.Lon, &p.TZ,
		&p.Color, &p.Notes, &fav, &p.Source, &p.ExternalID, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	p.IsMe, p.Favorite = me == 1, fav == 1
	return &p, nil
}

// ListPeople returns one person's people: you first, then favourites, then by name.
func (s *Store) ListPeople(ctx context.Context, uid int64, q string) ([]Person, error) {
	where, args := `user_id = ?`, []any{uid}
	if q = strings.TrimSpace(q); q != "" {
		where += ` AND (name LIKE ? OR place LIKE ? OR notes LIKE ?)`
		like := "%" + q + "%"
		args = append(args, like, like, like)
	}
	rows, err := s.DB.QueryContext(ctx, `SELECT `+personCols+` FROM people WHERE `+where+` ORDER BY is_me DESC, favorite DESC, name COLLATE NOCASE`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Person{}
	for rows.Next() {
		p, err := scanPerson(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

// GetPerson loads one of a person's people (someone else's id = ErrNotFound).
func (s *Store) GetPerson(ctx context.Context, uid, id int64) (*Person, error) {
	p, err := scanPerson(s.DB.QueryRowContext(ctx, `SELECT `+personCols+` FROM people WHERE user_id = ? AND id = ?`, uid, id))
	return p, notFound(err)
}

// GetMe is the account's own chart (ErrNotFound until it is added).
func (s *Store) GetMe(ctx context.Context, uid int64) (*Person, error) {
	p, err := scanPerson(s.DB.QueryRowContext(ctx, `SELECT `+personCols+` FROM people WHERE user_id = ? AND is_me = 1`, uid))
	return p, notFound(err)
}

// FindPersonByExternal finds a person brought in from elsewhere.
func (s *Store) FindPersonByExternal(ctx context.Context, uid int64, source, ext string) (*Person, error) {
	p, err := scanPerson(s.DB.QueryRowContext(ctx, `SELECT `+personCols+` FROM people WHERE user_id = ? AND source = ? AND external_id = ?`, uid, source, ext))
	return p, notFound(err)
}

// CreatePerson adds a person; marking them as "me" moves the mark from anyone else.
func (s *Store) CreatePerson(ctx context.Context, uid int64, p *Person) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if p.IsMe {
		if _, err := tx.ExecContext(ctx, `UPDATE people SET is_me = 0, relation = CASE WHEN relation = 'me' THEN 'other' ELSE relation END WHERE user_id = ? AND is_me = 1`, uid); err != nil {
			return err
		}
		p.Relation = "me"
	}
	now := Now()
	res, err := tx.ExecContext(ctx, `INSERT INTO people (user_id, name, relation, is_me, birth_date, birth_time, place, country, lat, lon, tz, color, notes, favorite, source, external_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		uid, p.Name, p.Relation, boolInt(p.IsMe), p.BirthDate, p.BirthTime, p.Place, p.Country, p.Lat, p.Lon, p.TZ, p.Color, p.Notes, boolInt(p.Favorite),
		firstOr(p.Source, "manual"), p.ExternalID, now, now)
	if err != nil {
		return conflict(err)
	}
	p.ID, _ = res.LastInsertId()
	p.UserID, p.CreatedAt, p.UpdatedAt = uid, now, now
	p.Source = firstOr(p.Source, "manual")
	return tx.Commit()
}

// UpdatePerson saves a person.
func (s *Store) UpdatePerson(ctx context.Context, uid int64, p *Person) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if p.IsMe {
		if _, err := tx.ExecContext(ctx, `UPDATE people SET is_me = 0, relation = CASE WHEN relation = 'me' THEN 'other' ELSE relation END WHERE user_id = ? AND is_me = 1 AND id <> ?`, uid, p.ID); err != nil {
			return err
		}
		p.Relation = "me"
	} else if p.Relation == "me" {
		p.Relation = "other"
	}
	p.UpdatedAt = Now()
	res, err := tx.ExecContext(ctx, `UPDATE people SET name = ?, relation = ?, is_me = ?, birth_date = ?, birth_time = ?, place = ?, country = ?, lat = ?, lon = ?, tz = ?,
		color = ?, notes = ?, favorite = ?, updated_at = ? WHERE user_id = ? AND id = ?`,
		p.Name, p.Relation, boolInt(p.IsMe), p.BirthDate, p.BirthTime, p.Place, p.Country, p.Lat, p.Lon, p.TZ, p.Color, p.Notes, boolInt(p.Favorite), p.UpdatedAt, uid, p.ID)
	if err != nil {
		return conflict(err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return tx.Commit()
}

// DeletePerson removes a person (and the readings about them).
func (s *Store) DeletePerson(ctx context.Context, uid, id int64) error {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM people WHERE user_id = ? AND id = ?`, uid, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func firstOr(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func idPtr(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	v := n.Int64
	return &v
}
