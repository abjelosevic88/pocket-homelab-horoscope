package store

import (
	"context"
	"encoding/json"
)

const journalCols = `id, user_id, date, mood, energy, note, tags, sky, source, created_at, updated_at`

func scanJournal(row interface{ Scan(...any) error }) (*JournalEntry, error) {
	var e JournalEntry
	var sky string
	if err := row.Scan(&e.ID, &e.UserID, &e.Date, &e.Mood, &e.Energy, &e.Note, &e.Tags, &sky, &e.Source, &e.CreatedAt, &e.UpdatedAt); err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(sky), &e.Sky)
	if e.Sky.Transits == nil {
		e.Sky.Transits = []string{}
	}
	if e.Sky.Retrograde == nil {
		e.Sky.Retrograde = []string{}
	}
	return &e, nil
}

// ListJournal returns check-ins between two dates (inclusive, "" = open), newest first.
func (s *Store) ListJournal(ctx context.Context, uid int64, from, to string, limit int) ([]JournalEntry, error) {
	where, args := `user_id = ?`, []any{uid}
	if from != "" {
		where += ` AND date >= ?`
		args = append(args, from)
	}
	if to != "" {
		where += ` AND date <= ?`
		args = append(args, to)
	}
	q := `SELECT ` + journalCols + ` FROM journal WHERE ` + where + ` ORDER BY date DESC`
	if limit > 0 {
		q += ` LIMIT ?`
		args = append(args, limit)
	}
	rows, err := s.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []JournalEntry{}
	for rows.Next() {
		e, err := scanJournal(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *e)
	}
	return out, rows.Err()
}

// GetJournal loads one check-in by id.
func (s *Store) GetJournal(ctx context.Context, uid, id int64) (*JournalEntry, error) {
	e, err := scanJournal(s.DB.QueryRowContext(ctx, `SELECT `+journalCols+` FROM journal WHERE user_id = ? AND id = ?`, uid, id))
	return e, notFound(err)
}

// GetJournalDay loads the check-in of a date.
func (s *Store) GetJournalDay(ctx context.Context, uid int64, date string) (*JournalEntry, error) {
	e, err := scanJournal(s.DB.QueryRowContext(ctx, `SELECT `+journalCols+` FROM journal WHERE user_id = ? AND date = ?`, uid, date))
	return e, notFound(err)
}

// SaveJournal creates the day's check-in or replaces it (one per day).
func (s *Store) SaveJournal(ctx context.Context, uid int64, e *JournalEntry) error {
	sky, _ := json.Marshal(e.Sky)
	now := Now()
	e.Tags = JoinTags(e.Tags)
	if e.Source == "" {
		e.Source = "app"
	}
	err := s.DB.QueryRowContext(ctx, `INSERT INTO journal (user_id, date, mood, energy, note, tags, sky, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, date) DO UPDATE SET mood = excluded.mood, energy = excluded.energy, note = excluded.note, tags = excluded.tags,
			sky = excluded.sky, updated_at = excluded.updated_at
		RETURNING id, created_at, source`, uid, e.Date, e.Mood, e.Energy, e.Note, e.Tags, string(sky), e.Source, now, now).Scan(&e.ID, &e.CreatedAt, &e.Source)
	if err != nil {
		return err
	}
	e.UserID, e.UpdatedAt = uid, now
	return nil
}

// DeleteJournal removes a check-in.
func (s *Store) DeleteJournal(ctx context.Context, uid, id int64) error {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM journal WHERE user_id = ? AND id = ?`, uid, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}
