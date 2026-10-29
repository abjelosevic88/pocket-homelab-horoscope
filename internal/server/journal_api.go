package server

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/astro"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/store"
)

// ---- the mood journal ----

// skyOf is the sky kept with a journal day: the Moon at local noon, your
// strongest transits and what is retrograde.
func (s *Server) skyOf(ctx context.Context, uid int64, st store.Settings, date string) store.JournalSky {
	_, _, loc, _ := s.homeOf(ctx, uid, st)
	day, err := localDay(date, loc)
	if err != nil {
		return store.JournalSky{Transits: []string{}, Retrograde: []string{}}
	}
	z, _, _ := chartOptions(st)
	at := noonOf(day).UTC()
	ph := astro.MoonPhase(at)
	sky := store.JournalSky{Phase: ph.Name, Illumination: round2(ph.Illumination), Transits: []string{}, Retrograde: []string{}}
	ps := astro.Positions(at, z)
	if mp, ok := findPos(ps, "moon"); ok {
		sky.MoonSign = signKey(mp.Sign)
	}
	for _, p := range ps {
		if p.Retro && p.Body != "node" {
			sky.Retrograde = append(sky.Retrograde, string(p.Body))
		}
	}
	if me, err := s.store.GetMe(ctx, uid); err == nil {
		if natal, err := s.natal(*me, st); err == nil {
			for i, t := range transitsAt(ps, natal) {
				if i >= 3 {
					break
				}
				sky.Transits = append(sky.Transits, t.Title)
			}
		}
	}
	return sky
}

func (s *Server) handleListJournal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	from, to := r.URL.Query().Get("from"), r.URL.Query().Get("to")
	if !validDate(from) || !validDate(to) {
		writeErr(w, http.StatusBadRequest, "from and to are dates (YYYY-MM-DD)")
		return
	}
	limit := queryInt(r, "limit")
	if limit <= 0 || limit > 2000 {
		limit = 400
	}
	js, err := s.store.ListJournal(ctx, actorID(ctx), from, to, limit)
	if err != nil {
		s.fail(w, err, "journal")
		return
	}
	writeJSON(w, http.StatusOK, js)
}

type journalInput struct {
	Date   *string `json:"date"`
	Mood   *int    `json:"mood"`
	Energy *int    `json:"energy"`
	Note   *string `json:"note"`
	Tags   *string `json:"tags"`
}

func (in journalInput) apply(e *store.JournalEntry) string {
	if in.Date != nil {
		e.Date = strings.TrimSpace(*in.Date)
	}
	if in.Mood != nil {
		e.Mood = *in.Mood
	}
	if in.Energy != nil {
		e.Energy = *in.Energy
	}
	if in.Note != nil {
		e.Note = *in.Note
	}
	if in.Tags != nil {
		e.Tags = *in.Tags
	}
	switch {
	case e.Date == "" || !validDate(e.Date):
		return "date looks like 2026-10-29"
	case e.Mood < 0 || e.Mood > 5 || e.Energy < 0 || e.Energy > 5:
		return "mood and energy are 1 to 5 (0 = not set)"
	case e.Mood == 0 && e.Energy == 0 && strings.TrimSpace(e.Note) == "":
		return "add a mood, an energy level or a note"
	case len(e.Note) > 20000:
		return "the note is too long"
	}
	return ""
}

// handleSaveJournal writes the day's check-in (one per day: a second one
// replaces it). A linked app may only add days nobody wrote yet, or change
// the days it wrote itself.
func (s *Server) handleSaveJournal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	st, _ := s.store.GetSettings(ctx, uid)
	var in journalInput
	if err := decode(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if in.Date == nil || *in.Date == "" {
		_, _, loc, _ := s.homeOf(ctx, uid, st)
		d := time.Now().In(loc).Format("2006-01-02")
		in.Date = &d
	}
	e := store.JournalEntry{Source: "app"}
	if cur, err := s.store.GetJournalDay(ctx, uid, *in.Date); err == nil {
		e = *cur
	}
	if app := linkApp(ctx); app != "" {
		if e.ID != 0 && e.Source != app {
			writeErr(w, http.StatusForbidden, "this day was written in Pocket Horoscope; a linked app may not change it")
			return
		}
		e.Source = app
	}
	if msg := in.apply(&e); msg != "" {
		writeErr(w, http.StatusBadRequest, msg)
		return
	}
	e.Sky = s.skyOf(ctx, uid, st, e.Date)
	isNew := e.ID == 0
	if err := s.store.SaveJournal(ctx, uid, &e); err != nil {
		s.fail(w, err, "save check-in")
		return
	}
	if isNew {
		s.publish(ctx, uid, "horoscope.journal_logged", e.Date, map[string]any{"date": e.Date, "mood": e.Mood, "energy": e.Energy, "moon_sign": e.Sky.MoonSign, "phase": e.Sky.Phase})
	}
	writeJSON(w, http.StatusOK, e)
}

func (s *Server) handleUpdateJournal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	e, err := s.store.GetJournal(ctx, uid, idParam(r, "id"))
	if err != nil {
		s.fail(w, err, "check-in")
		return
	}
	if app := linkApp(ctx); app != "" && e.Source != app {
		writeErr(w, http.StatusForbidden, "this day was written in Pocket Horoscope; a linked app may not change it")
		return
	}
	var in journalInput
	if err := decode(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	oldDate := e.Date
	if msg := in.apply(e); msg != "" {
		writeErr(w, http.StatusBadRequest, msg)
		return
	}
	if e.Date != oldDate {
		if _, err := s.store.GetJournalDay(ctx, uid, e.Date); err == nil {
			writeErr(w, http.StatusConflict, "there is already a check-in on "+e.Date)
			return
		}
		_ = s.store.DeleteJournal(ctx, uid, e.ID)
	}
	st, _ := s.store.GetSettings(ctx, uid)
	e.Sky = s.skyOf(ctx, uid, st, e.Date)
	if err := s.store.SaveJournal(ctx, uid, e); err != nil {
		s.fail(w, err, "save check-in")
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (s *Server) handleDeleteJournal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	e, err := s.store.GetJournal(ctx, uid, idParam(r, "id"))
	if err != nil {
		s.fail(w, err, "check-in")
		return
	}
	if app := linkApp(ctx); app != "" && e.Source != app {
		writeErr(w, http.StatusForbidden, "this day was written in Pocket Horoscope; a linked app may not delete it")
		return
	}
	if err := s.store.DeleteJournal(ctx, uid, e.ID); err != nil {
		s.fail(w, err, "delete check-in")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ---- insights: your days next to the Moon ----

type bucket struct {
	Key    string  `json:"key"`
	Mood   float64 `json:"mood"`
	Energy float64 `json:"energy"`
	N      int     `json:"n"`
	Value  float64 `json:"value,omitempty"`
}

var phaseOrder = []string{"new", "waxing_crescent", "first_quarter", "waxing_gibbous", "full", "waning_gibbous", "last_quarter", "waning_crescent"}

var signElement = []string{"fire", "earth", "air", "water"}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	t := 0.0
	for _, x := range xs {
		t += x
	}
	return t / float64(len(xs))
}

func round1(v float64) float64 { return math.Round(v*10) / 10 }

// handleInsights compares mood and energy by Moon phase, Moon element and
// Moon sign (?days=, default a year), and — with Pocket Health linked — sleep
// and resting heart rate by Moon phase.
func (s *Server) handleInsights(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	st, _ := s.store.GetSettings(ctx, uid)
	days := queryInt(r, "days")
	if days <= 0 || days > 3660 {
		days = 365
	}
	_, _, loc, _ := s.homeOf(ctx, uid, st)
	today, _ := localDay("", loc)
	from := today.AddDate(0, 0, -(days - 1)).Format("2006-01-02")
	js, err := s.store.ListJournal(ctx, uid, from, today.Format("2006-01-02"), 0)
	if err != nil {
		s.fail(w, err, "journal")
		return
	}
	type acc struct{ mood, energy []float64 }
	byPhase, byEl, bySign := map[string]*acc{}, map[string]*acc{}, map[string]*acc{}
	add := func(m map[string]*acc, k string, e store.JournalEntry) {
		if k == "" {
			return
		}
		a := m[k]
		if a == nil {
			a = &acc{}
			m[k] = a
		}
		if e.Mood > 0 {
			a.mood = append(a.mood, float64(e.Mood))
		}
		if e.Energy > 0 {
			a.energy = append(a.energy, float64(e.Energy))
		}
	}
	all := acc{}
	for _, e := range js {
		add(byPhase, e.Sky.Phase, e)
		add(bySign, e.Sky.MoonSign, e)
		for i, sk := range astro.Signs {
			if sk == e.Sky.MoonSign {
				add(byEl, signElement[i%4], e)
			}
		}
		if e.Mood > 0 {
			all.mood = append(all.mood, float64(e.Mood))
		}
		if e.Energy > 0 {
			all.energy = append(all.energy, float64(e.Energy))
		}
	}
	list := func(m map[string]*acc, keys []string) []bucket {
		out := []bucket{}
		for _, k := range keys {
			a := m[k]
			if a == nil {
				out = append(out, bucket{Key: k})
				continue
			}
			out = append(out, bucket{Key: k, Mood: round1(mean(a.mood)), Energy: round1(mean(a.energy)), N: max(len(a.mood), len(a.energy))})
		}
		return out
	}
	phases := list(byPhase, phaseOrder)
	findings := []string{}
	avg := mean(all.mood)
	if len(all.mood) >= 14 {
		best, worst := bucket{}, bucket{Mood: 99}
		for _, b := range phases {
			if b.N < 3 || b.Mood == 0 {
				continue
			}
			if b.Mood > best.Mood {
				best = b
			}
			if b.Mood < worst.Mood {
				worst = b
			}
		}
		if best.Key != "" && best.Mood-avg >= 0.3 {
			findings = append(findings, fmt.Sprintf("Your mood is best around the %s Moon (%.1f against your usual %.1f).", phaseWords(best.Key), best.Mood, avg))
		}
		if worst.Key != "" && worst.Key != best.Key && avg-worst.Mood >= 0.3 {
			findings = append(findings, fmt.Sprintf("Around the %s Moon your mood dips (%.1f).", phaseWords(worst.Key), worst.Mood))
		}
		els := list(byEl, signElement)
		sort.SliceStable(els, func(i, j int) bool { return els[i].Mood > els[j].Mood })
		if len(els) > 0 && els[0].N >= 4 && els[0].Mood-avg >= 0.3 {
			findings = append(findings, fmt.Sprintf("Days with the Moon in a %s sign suit you best.", els[0].Key))
		}
		if len(findings) == 0 {
			findings = append(findings, "So far the Moon makes no clear difference to your mood. Keep checking in; patterns show after a few months.")
		}
	}
	out := map[string]any{"from": from, "days": days, "entries": len(js), "mood_avg": round1(avg), "energy_avg": round1(mean(all.energy)),
		"phases": phases, "elements": list(byEl, signElement), "signs": list(bySign, astro.Signs), "findings": findings}
	if st.HealthInsights {
		out["health"] = s.healthByPhase(ctx, uid, today)
	}
	writeJSON(w, http.StatusOK, out)
}

func phaseWords(p string) string { return strings.ReplaceAll(p, "_", " ") }

// healthByPhase reads sleep and resting heart rate from Pocket Health and
// averages them by the Moon's phase (a night counts for the phase of its evening).
func (s *Server) healthByPhase(ctx context.Context, uid int64, today time.Time) map[string]any {
	sib, err := s.clients(ctx, uid).conn(ctx, appHealth, scopeHealthRead)
	if err != nil {
		return map[string]any{"ok": false, "error": err.Error(), "not_connected": isNotConnected(err)}
	}
	res := map[string]any{"ok": true}
	for _, m := range []struct{ key, name, unit string }{{"sleep_asleep", "Time asleep", "min"}, {"resting_hr", "Resting heart rate", "bpm"}} {
		pts, err := sib.client.Metric(ctx, m.key, "y", today.Format("2006-01-02"))
		if err != nil {
			res[m.key] = map[string]any{"error": err.Error()}
			continue
		}
		by := map[string][]float64{}
		var all []float64
		for _, p := range pts {
			d, err := time.ParseInLocation("2006-01-02", p.Date, today.Location())
			if err != nil || p.Value <= 0 {
				continue
			}
			at := noonOf(d)
			if m.key == "sleep_asleep" {
				at = at.AddDate(0, 0, -1).Add(10 * time.Hour) // the evening before the morning it is dated by
			}
			ph := astro.MoonPhase(at.UTC()).Name
			by[ph] = append(by[ph], p.Value)
			all = append(all, p.Value)
		}
		bs := []bucket{}
		for _, k := range phaseOrder {
			bs = append(bs, bucket{Key: k, Value: round1(mean(by[k])), N: len(by[k])})
		}
		res[m.key] = map[string]any{"name": m.name, "unit": m.unit, "avg": round1(mean(all)), "n": len(all), "phases": bs}
	}
	return res
}
