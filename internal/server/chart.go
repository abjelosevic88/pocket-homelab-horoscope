package server

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/astro"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/content"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/store"
)

// ---- charts ----
//
// A chart is worked out from the birth data with the person's own chart
// settings (zodiac, houses, orbs) and kept in memory for an hour: the same
// birth data and settings always give the same chart.

type chartEntry struct {
	at    time.Time
	chart astro.Chart
}

// zoneFor turns a stored zone into a location: an IANA name, or a fixed
// offset like "UTC+01:00" (what birth-data files carry), else UTC.
func zoneFor(tz string) *time.Location {
	tz = strings.TrimSpace(tz)
	if tz == "" {
		return time.UTC
	}
	if loc, err := time.LoadLocation(tz); err == nil {
		return loc
	}
	if off, ok := parseOffset(tz); ok {
		return time.FixedZone(fmtOffset(off), off)
	}
	return time.UTC
}

// parseOffset reads "UTC+01:00", "+1", "-05:30", "GMT-3" as seconds east of UTC.
func parseOffset(v string) (int, bool) {
	v = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(strings.ToUpper(v), "UTC"), "GMT"))
	if v == "" {
		return 0, true
	}
	sign := 1
	switch v[0] {
	case '+':
		v = v[1:]
	case '-':
		sign, v = -1, v[1:]
	default:
		return 0, false
	}
	h, m := v, "0"
	if i := strings.IndexAny(v, ":."); i >= 0 {
		h, m = v[:i], v[i+1:]
	} else if len(v) == 4 {
		h, m = v[:2], v[2:]
	}
	hh, err1 := strconv.Atoi(h)
	mm, err2 := strconv.Atoi(m)
	if err1 != nil || err2 != nil || hh > 14 || mm >= 60 {
		return 0, false
	}
	return sign * (hh*3600 + mm*60), true
}

func fmtOffset(sec int) string {
	sign := "+"
	if sec < 0 {
		sign, sec = "-", -sec
	}
	return fmt.Sprintf("UTC%s%02d:%02d", sign, sec/3600, sec/60%60)
}

// validZone reports whether a zone name can be used.
func validZone(tz string) bool {
	if tz == "" {
		return true
	}
	if _, err := time.LoadLocation(tz); err == nil {
		return true
	}
	_, ok := parseOffset(tz)
	return ok
}

// birthMoment is the birth as a UTC instant; an unknown time is local noon.
func birthMoment(p store.Person) (time.Time, error) {
	d, err := time.Parse("2006-01-02", p.BirthDate)
	if err != nil {
		return time.Time{}, fmt.Errorf("birth date looks like 1990-04-21")
	}
	h, m := 12, 0
	if p.BirthTime != "" {
		t, err := time.Parse("15:04", p.BirthTime)
		if err != nil {
			return time.Time{}, fmt.Errorf("birth time looks like 07:45")
		}
		h, m = t.Hour(), t.Minute()
	}
	loc := zoneFor(p.TZ)
	return time.Date(d.Year(), d.Month(), d.Day(), h, m, 0, 0, loc).UTC(), nil
}

// chartOptions are a person's chart settings.
func chartOptions(st store.Settings) (astro.Zodiac, astro.HouseSystem, astro.Orbs) {
	z := astro.Zodiac(st.Zodiac)
	if z != "sidereal" {
		z = "tropical"
	}
	hs := astro.HouseSystem(st.HouseSystem)
	switch hs {
	case "placidus", "koch", "whole", "equal", "porphyry":
	default:
		hs = "placidus"
	}
	return z, hs, orbsFor(st)
}

// orbsFor scales the default orbs (tight / normal / wide) and drops the
// quincunx unless minor aspects are on.
func orbsFor(st store.Settings) astro.Orbs {
	k := 1.0
	switch st.Orbs {
	case "tight":
		k = 0.7
	case "wide":
		k = 1.25
	}
	out := astro.Orbs{}
	for t, o := range astro.DefaultOrbs() {
		if t == "quincunx" && !st.MinorAspects {
			continue
		}
		out[t] = o * k
	}
	return out
}

func chartKey(p store.Person, st store.Settings) string {
	return strings.Join([]string{p.BirthDate, p.BirthTime, p.TZ, strconv.FormatFloat(p.Lat, 'f', 4, 64), strconv.FormatFloat(p.Lon, 'f', 4, 64),
		st.Zodiac, st.HouseSystem, st.Orbs, strconv.FormatBool(st.MinorAspects)}, "|")
}

// natal is a person's birth chart (cached).
func (s *Server) natal(p store.Person, st store.Settings) (astro.Chart, error) {
	key := chartKey(p, st)
	s.chartMu.Lock()
	if e, ok := s.charts[key]; ok {
		s.chartMu.Unlock()
		return e.chart, nil
	}
	s.chartMu.Unlock()
	at, err := birthMoment(p)
	if err != nil {
		return astro.Chart{}, err
	}
	z, hs, orbs := chartOptions(st)
	c := astro.NatalChart(astro.ChartInput{Time: at, TimeKnown: p.TimeKnown() && p.HasPlace(), Lat: p.Lat, Lon: p.Lon, Zodiac: z, HouseSystem: hs}, orbs)
	s.chartMu.Lock()
	if len(s.charts) > 2000 {
		s.charts = map[string]chartEntry{}
	}
	s.charts[key] = chartEntry{at: time.Now(), chart: c}
	s.chartMu.Unlock()
	return c, nil
}

// dropCharts forgets charts not used for an hour.
func (s *Server) dropCharts() {
	s.chartMu.Lock()
	defer s.chartMu.Unlock()
	for k, e := range s.charts {
		if time.Since(e.at) > time.Hour {
			delete(s.charts, k)
		}
	}
}

// ---- small helpers used everywhere ----

func signKey(i int) string {
	if i < 0 || i > 11 {
		return ""
	}
	return astro.Signs[i]
}

func signName(i int) string { return content.SignByIndex(((i % 12) + 12) % 12).Name }

func bodyName(k string) string {
	if p, ok := content.PlanetByKey(k); ok {
		return p.Name
	}
	return strings.ToUpper(k[:1]) + k[1:]
}

// fmtLon is "14°32′ Sagittarius".
func fmtLon(lon float64) string {
	lon = math.Mod(math.Mod(lon, 360)+360, 360)
	sign := int(lon / 30)
	d := lon - float64(sign*30)
	deg := int(d)
	min := int(math.Round((d - float64(deg)) * 60))
	if min == 60 {
		deg, min = deg+1, 0
	}
	return fmt.Sprintf("%d°%02d′ %s", deg, min, signName(sign))
}

// position finds a body in a chart.
func position(c astro.Chart, b string) (astro.Position, bool) {
	for _, p := range c.Positions {
		if string(p.Body) == b {
			return p, true
		}
	}
	return astro.Position{}, false
}

// point finds a chart point (asc, mc, …).
func point(c astro.Chart, k string) (astro.Point, bool) {
	for _, p := range c.Points {
		if p.Key == k {
			return p, true
		}
	}
	return astro.Point{}, false
}

// bigThree are a chart's Sun, Moon and rising sign keys ("" when unknown).
func bigThree(c astro.Chart) (sun, moon, rising string) {
	if p, ok := position(c, "sun"); ok {
		sun = signKey(p.Sign)
	}
	if p, ok := position(c, "moon"); ok {
		moon = signKey(p.Sign)
	}
	if p, ok := point(c, "asc"); ok && c.TimeKnown {
		rising = signKey(p.Sign)
	}
	return
}

// personView is a person with their Big Three, for lists.
type personView struct {
	store.Person
	Sun       string `json:"sun"`
	Moon      string `json:"moon"`
	Rising    string `json:"rising"`
	Age       int    `json:"age"`
	NextBDay  string `json:"next_birthday"`
	DaysToBD  int    `json:"days_to_birthday"`
	ChartNote string `json:"chart_note,omitempty"`
}

func (s *Server) viewPerson(p store.Person, st store.Settings, today time.Time) personView {
	v := personView{Person: p}
	if c, err := s.natal(p, st); err == nil {
		v.Sun, v.Moon, v.Rising = bigThree(c)
		if !c.TimeKnown {
			v.ChartNote = "No birth time or place: no rising sign or houses, and the Moon may be a sign off."
			if mp, ok := position(c, "moon"); ok {
				deg := mp.Degree
				if deg > 6 && deg < 24 {
					v.ChartNote = "No birth time or place: no rising sign or houses."
				}
			}
		}
	}
	if d, err := time.Parse("2006-01-02", p.BirthDate); err == nil {
		v.Age = today.Year() - d.Year()
		next := time.Date(today.Year(), d.Month(), d.Day(), 0, 0, 0, 0, today.Location())
		if d.Month() == 2 && d.Day() == 29 && next.Month() != 2 {
			next = time.Date(today.Year(), 2, 28, 0, 0, 0, 0, today.Location())
		}
		t0 := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
		if next.Before(t0) {
			next = next.AddDate(1, 0, 0)
		} else {
			v.Age--
		}
		if next.Equal(t0) {
			v.Age++
		}
		v.NextBDay = next.Format("2006-01-02")
		v.DaysToBD = int(next.Sub(t0).Hours()/24 + 0.5)
	}
	return v
}

// homeOf is where the person lives (Settings), else their own birth place,
// else the server's zone at 0°/0°.
func (s *Server) homeOf(ctx context.Context, uid int64, st store.Settings) (lat, lon float64, loc *time.Location, place string) {
	if st.HomeTZ != "" {
		return st.HomeLat, st.HomeLon, zoneFor(st.HomeTZ), st.HomePlace
	}
	if me, err := s.store.GetMe(ctx, uid); err == nil && me.HasPlace() {
		return me.Lat, me.Lon, zoneFor(me.TZ), me.Place
	}
	return 0, 0, time.Local, ""
}

// localDay is the date in a zone (today when d is empty).
func localDay(d string, loc *time.Location) (time.Time, error) {
	if d == "" {
		n := time.Now().In(loc)
		return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, loc), nil
	}
	t, err := time.ParseInLocation("2006-01-02", d, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("date looks like 2026-10-29")
	}
	return t, nil
}
