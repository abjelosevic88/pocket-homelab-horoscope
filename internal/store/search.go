package store

import (
	"context"
	"strings"
)

// likeArg is a LIKE pattern that finds w anywhere, with %, _ and \ taken as
// they are (use with ESCAPE '\').
func likeArg(w string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + r.Replace(w) + "%"
}

// wordsWhere is "every word is in one of these columns".
func wordsWhere(words, cols []string) (string, []any) {
	var where []string
	var args []any
	for _, w := range words {
		ors := make([]string, len(cols))
		for i, c := range cols {
			ors[i] = c + ` LIKE ? ESCAPE '\'`
			args = append(args, likeArg(w))
		}
		where = append(where, "("+strings.Join(ors, " OR ")+")")
	}
	return strings.Join(where, " AND "), args
}

// SearchReadings finds the readings a person kept whose title, text or
// people have every word, newest first.
func (s *Store) SearchReadings(ctx context.Context, uid int64, words []string, limit int) ([]Reading, error) {
	if len(words) == 0 {
		return []Reading{}, nil
	}
	where, args := wordsWhere(words, []string{"r.title", "r.body", "COALESCE(p.name, '')", "COALESCE(o.name, '')"})
	rows, err := s.DB.QueryContext(ctx, readingSelect+` WHERE r.user_id = ? AND r.saved = 1 AND `+where+` ORDER BY r.created_at DESC, r.id DESC LIMIT ?`,
		append(append([]any{uid}, args...), limit)...)
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

// SearchJournal finds the check-ins whose note or tags have every word, newest first.
func (s *Store) SearchJournal(ctx context.Context, uid int64, words []string, limit int) ([]JournalEntry, error) {
	if len(words) == 0 {
		return []JournalEntry{}, nil
	}
	where, args := wordsWhere(words, []string{"note", "tags"})
	rows, err := s.DB.QueryContext(ctx, `SELECT `+journalCols+` FROM journal WHERE user_id = ? AND `+where+` ORDER BY date DESC LIMIT ?`,
		append(append([]any{uid}, args...), limit)...)
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

// DailyReadings are the day readings stored for a date (YYYY-MM-DD), the
// latest one per person. Unsaved ones count too: they are the day's reading
// until they are pruned, as on the Today page.
func (s *Store) DailyReadings(ctx context.Context, uid int64, date string) ([]Reading, error) {
	rows, err := s.DB.QueryContext(ctx, readingSelect+` WHERE r.user_id = ? AND r.kind = 'daily' AND r.ref = ? AND r.other_id IS NULL
		AND r.id = (SELECT MAX(x.id) FROM readings x WHERE x.user_id = r.user_id AND x.kind = 'daily' AND x.ref = r.ref AND x.other_id IS NULL
			AND COALESCE(x.person_id, 0) = COALESCE(r.person_id, 0))
		ORDER BY r.id`, uid, date)
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
