package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/astro"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/notify"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/store"
)

// pushes runs every minute and sends, at each person's chosen time (in
// their home zone): the day's horoscope, the Moon's big days (new and full
// Moons, eclipses), planets turning retrograde or direct, and birthdays.
// Each goes out once a day.
func (s *Server) pushes(ctx context.Context) {
	ids, err := s.store.SettingsUsers(ctx)
	if err != nil {
		return
	}
	for _, uid := range ids {
		st, err := s.store.GetSettings(ctx, uid)
		if err != nil || st.NotifyURL == "" {
			continue
		}
		_, _, loc, _ := s.homeOf(ctx, uid, st)
		now := time.Now().In(loc)
		if !withinMinutes(now.Format("15:04"), firstNonEmpty(st.DailyTime, "07:30"), 4) {
			continue
		}
		_, _ = s.sendDay(ctx, uid, st, now, false)
	}
}

// sendDay sends what the day has for one person (force = the "send now" button).
func (s *Server) sendDay(ctx context.Context, uid int64, st store.Settings, now time.Time, force bool) (int, error) {
	n := notify.New(st.NotifyURL, st.NotifyFormat)
	today := now.Format("2006-01-02")
	once := func(key string) bool {
		if force {
			return true
		}
		k := fmt.Sprintf("push:%d:%s:%s", uid, today, key)
		if _, _, done := s.store.CacheGet(ctx, k); done {
			return false
		}
		_ = s.store.CachePut(ctx, k, "1")
		return true
	}
	sent := 0
	var firstErr error
	send := func(m notify.Message) {
		if err := n.Send(ctx, m); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			s.log.Debug("notify failed", "user", uid, "err", err)
			return
		}
		sent++
	}
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	me, meErr := s.store.GetMe(ctx, uid)
	if st.DailyNotify && meErr == nil && once("daily") {
		if rd, err := s.readingFor(ctx, uid, *me, st, day, false); err == nil {
			body := rd.Summary
			if st.AIDaily {
				if r, _, err := s.writeReading(ctx, uid, readingRequest{Kind: "daily", Date: today}); err == nil && r != nil {
					body = firstSentence(strings.TrimLeft(r.Body, "# "))
				}
			}
			body += fmt.Sprintf("\nLove %d · Work %d · Mood %d · Energy %d", rd.Scores["love"], rd.Scores["work"], rd.Scores["mood"], rd.Scores["energy"])
			for _, t := range rd.Transits {
				if t.Weight >= 4 && t.Orb < 0.5 {
					body += "\nExact today: " + t.Title + "."
				}
			}
			send(notify.Message{Title: rd.Headline, Body: body, Priority: "info", Tags: []string{"crystal_ball"}, Link: s.cfg.BaseURL + "/"})
			s.publish(ctx, uid, "horoscope.daily_ready", today, map[string]any{"date": today, "headline": rd.Headline, "scores": rd.Scores, "overall": rd.Overall,
				"moon_sign": rd.Moon.Sign, "phase": rd.Moon.Phase})
		}
	}
	z, _, _ := chartOptions(st)
	loc := now.Location()
	for _, e := range astro.Events(day.UTC(), day.AddDate(0, 0, 1).UTC(), z, false) {
		v := viewEvent(e, loc)
		at := e.Time.In(loc).Format("15:04")
		switch {
		case (e.Kind == "phase" && (e.Phase == "new" || e.Phase == "full")) || e.Kind == "eclipse":
			if st.NotifyMoon && once("sky:"+e.Kind+e.Phase) {
				send(notify.Message{Title: v.Title + " at " + at, Body: firstSentence(v.Text), Priority: "info", Tags: []string{"new_moon"}, Link: s.cfg.BaseURL + "/sky"})
			}
		case e.Kind == "station":
			if st.NotifyRetro && once("station:"+string(e.Body)) {
				send(notify.Message{Title: v.Title, Body: firstSentence(v.Text), Priority: "info", Tags: []string{"arrows_counterclockwise"}, Link: s.cfg.BaseURL + "/sky"})
			}
		}
	}
	if st.NotifyBirthdays && once("birthdays") {
		ps, _ := s.store.ListPeople(ctx, uid, "")
		for _, p := range ps {
			if p.IsMe {
				continue
			}
			if v := s.viewPerson(p, st, day); v.DaysToBD == 0 && v.NextBDay == today {
				send(notify.Message{Title: p.Name + "'s birthday", Body: fmt.Sprintf("%s turns %d today. A good day to look at their year ahead together.", p.Name, v.Age),
					Priority: "info", Tags: []string{"birthday"}, Link: fmt.Sprintf("%s/people/%d", s.cfg.BaseURL, p.ID)})
			}
		}
	}
	return sent, firstErr
}

// withinMinutes reports whether now (HH:MM) is at or up to n minutes after t (HH:MM).
func withinMinutes(now, t string, n int) bool {
	a, err1 := time.Parse("15:04", now)
	b, err2 := time.Parse("15:04", t)
	if err1 != nil || err2 != nil {
		return false
	}
	d := a.Sub(b)
	return d >= 0 && d <= time.Duration(n)*time.Minute
}

// handleSendToday sends today's notifications now (Settings → Notifications).
func (s *Server) handleSendToday(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	st, _ := s.store.GetSettings(ctx, uid)
	if st.NotifyURL == "" {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": "set a notification address first"})
		return
	}
	_, _, loc, _ := s.homeOf(ctx, uid, st)
	st.DailyNotify = true
	n, err := s.sendDay(ctx, uid, st, time.Now().In(loc), true)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	if n == 0 {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": "nothing to send: add your own birth details first"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "sent": n})
}
