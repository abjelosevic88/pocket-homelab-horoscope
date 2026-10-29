package store

import (
	"context"
	"database/sql"
)

const readingSelect = `SELECT r.id, r.user_id, r.person_id, r.other_id, r.kind, r.ref, r.title, r.body, r.model, r.saved, r.created_at,
	COALESCE(p.name, ''), COALESCE(o.name, '')
	FROM readings r LEFT JOIN people p ON p.id = r.person_id LEFT JOIN people o ON o.id = r.other_id`

func scanReading(row interface{ Scan(...any) error }) (*Reading, error) {
	var r Reading
	var pid, oid sql.NullInt64
	var saved int
	if err := row.Scan(&r.ID, &r.UserID, &pid, &oid, &r.Kind, &r.Ref, &r.Title, &r.Body, &r.Model, &saved, &r.CreatedAt, &r.PersonName, &r.OtherName); err != nil {
		return nil, err
	}
	r.PersonID, r.OtherID, r.Saved = idPtr(pid), idPtr(oid), saved == 1
	return &r, nil
}

// ReadingFilter narrows ListReadings.
type ReadingFilter struct {
	Kind      string
	PersonID  int64
	SavedOnly bool
	Limit     int
}

// ListReadings returns readings, newest first.
func (s *Store) ListReadings(ctx context.Context, uid int64, f ReadingFilter) ([]Reading, error) {
	where, args := `r.user_id = ?`, []any{uid}
	if f.Kind != "" {
		where += ` AND r.kind = ?`
		args = append(args, f.Kind)
	}
	if f.PersonID > 0 {
		where += ` AND (r.person_id = ? OR r.other_id = ?)`
		args = append(args, f.PersonID, f.PersonID)
	}
	if f.SavedOnly {
		where += ` AND r.saved = 1`
	}
	q := readingSelect + ` WHERE ` + where + ` ORDER BY r.created_at DESC, r.id DESC`
	if f.Limit > 0 {
		q += ` LIMIT ?`
		args = append(args, f.Limit)
	}
	rows, err := s.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Reading{}
	for rows.Next() {
		r, err := scanReading(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *r)
	}
	return out, rows.Err()
}

// GetReading loads one reading.
func (s *Store) GetReading(ctx context.Context, uid, id int64) (*Reading, error) {
	r, err := scanReading(s.DB.QueryRowContext(ctx, readingSelect+` WHERE r.user_id = ? AND r.id = ?`, uid, id))
	return r, notFound(err)
}

// FindReading is the latest reading of a kind for a person (and partner) and ref.
func (s *Store) FindReading(ctx context.Context, uid int64, kind string, personID, otherID int64, ref string) (*Reading, error) {
	r, err := scanReading(s.DB.QueryRowContext(ctx, readingSelect+` WHERE r.user_id = ? AND r.kind = ? AND COALESCE(r.person_id, 0) = ? AND COALESCE(r.other_id, 0) = ? AND r.ref = ?
		ORDER BY r.id DESC LIMIT 1`, uid, kind, personID, otherID, ref))
	return r, notFound(err)
}

// CreateReading stores a reading.
func (s *Store) CreateReading(ctx context.Context, uid int64, r *Reading) error {
	now := Now()
	res, err := s.DB.ExecContext(ctx, `INSERT INTO readings (user_id, person_id, other_id, kind, ref, title, body, model, saved, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		uid, nullI(r.PersonID), nullI(r.OtherID), r.Kind, r.Ref, r.Title, r.Body, r.Model, boolInt(r.Saved), now)
	if err != nil {
		return err
	}
	r.ID, _ = res.LastInsertId()
	r.UserID, r.CreatedAt = uid, now
	return nil
}

// SetReadingSaved keeps (or un-keeps) a reading.
func (s *Store) SetReadingSaved(ctx context.Context, uid, id int64, saved bool) error {
	res, err := s.DB.ExecContext(ctx, `UPDATE readings SET saved = ? WHERE user_id = ? AND id = ?`, boolInt(saved), uid, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteReading removes a reading.
func (s *Store) DeleteReading(ctx context.Context, uid, id int64) error {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM readings WHERE user_id = ? AND id = ?`, uid, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// PruneReadings drops unsaved readings older than the date (they are a cache).
func (s *Store) PruneReadings(ctx context.Context, before string) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM readings WHERE saved = 0 AND created_at < ?`, before)
	return err
}
