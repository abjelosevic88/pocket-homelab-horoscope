package server

import (
	"context"
	"hash/fnv"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/astro"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/content"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/store"
)

// ---- the day's horoscope ----
//
// A day's reading is worked out, not written by hand: the sky at the
// person's local noon against their birth chart. Each transit (a moving
// planet in aspect to a point of the birth chart) counts by how heavy it is
// (content.Transit's weight), how close it is (the orb) and its tone, into
// the four areas it touches. The Moon sets the mood of the day. The same
// person, day and settings always give the same reading.

var areas = []string{"love", "work", "mood", "energy"}

// dailyOrb is how far from exact a transiting body still counts on a day.
var dailyOrb = map[string]float64{"moon": 6, "sun": 3, "mercury": 3, "venus": 3, "mars": 3, "jupiter": 2.5, "saturn": 2.5, "uranus": 2, "neptune": 2, "pluto": 2}

type dayTransit struct {
	Transit  string   `json:"transit"`
	Natal    string   `json:"natal"`
	Type     string   `json:"type"`
	Orb      float64  `json:"orb"`
	Applying bool     `json:"applying"`
	Title    string   `json:"title"`
	Text     string   `json:"text"`
	Tone     int      `json:"tone"`
	Areas    []string `json:"areas"`
	Weight   int      `json:"weight"`
	Strength float64  `json:"strength"`
}

type dayMoon struct {
	Sign         string  `json:"sign"`
	Degree       float64 `json:"degree"`
	Phase        string  `json:"phase"`
	Illumination float64 `json:"illumination"`
	Age          float64 `json:"age"`
	House        int     `json:"house"`
	Text         string  `json:"text"`
	PhaseText    string  `json:"phase_text"`
	VoidFrom     string  `json:"void_from,omitempty"`
	VoidTo       string  `json:"void_to,omitempty"`
}

type dayScore struct {
	Date    string         `json:"date"`
	Overall int            `json:"overall"`
	Scores  map[string]int `json:"scores"`
	Tone    int            `json:"tone"`
	Moon    string         `json:"moon"`
	Phase   string         `json:"phase"`
}

type retroView struct {
	Body string `json:"body"`
	Name string `json:"name"`
	Sign string `json:"sign"`
	Text string `json:"text"`
}

type dayReading struct {
	Date       string              `json:"date"`
	PersonID   int64               `json:"person_id"`
	PersonName string              `json:"person_name"`
	Headline   string              `json:"headline"`
	Summary    string              `json:"summary"`
	Tone       int                 `json:"tone"`
	Overall    int                 `json:"overall"`
	Scores     map[string]int      `json:"scores"`
	Moon       dayMoon             `json:"moon"`
	Transits   []dayTransit        `json:"transits"`
	Tips       content.DayTips     `json:"tips"`
	Retrograde []retroView         `json:"retrograde"`
	Events     []skyEventView      `json:"events"`
	Week       []dayScore          `json:"week"`
	Reading    *store.Reading      `json:"reading"`
	Journal    *store.JournalEntry `json:"journal"`
	Birthday   bool                `json:"birthday"`
}

// natalTargets are the birth-chart points transits are measured against.
func natalTargets(c astro.Chart) ([]astro.Position, []astro.Point) {
	pts := []astro.Point{}
	for _, p := range c.Points {
		if p.Key == "asc" || p.Key == "mc" {
			pts = append(pts, p)
		}
	}
	return c.Positions, pts
}

// transitsAt are the transits at an instant, strongest first.
func transitsAt(sky []astro.Position, natal astro.Chart) []dayTransit {
	nPos, nPts := natalTargets(natal)
	wide := astro.Orbs{"conjunction": 6, "opposition": 6, "square": 6, "trine": 6, "sextile": 6}
	asp := astro.CrossAspects(sky, nil, nPos, nPts, wide)
	out := []dayTransit{}
	for _, a := range asp {
		if a.B == "node" && a.A == "moon" {
			continue
		}
		max := dailyOrb[a.A]
		if max == 0 || a.Orb > max {
			continue
		}
		tt := content.Transit(a.A, string(a.Type), a.B)
		if tt.Title == "" {
			continue
		}
		closeness := 1 - a.Orb/max
		if a.Applying {
			closeness = math.Min(1, closeness+0.1)
		}
		out = append(out, dayTransit{Transit: a.A, Natal: a.B, Type: string(a.Type), Orb: round2(a.Orb), Applying: a.Applying, Title: tt.Title, Text: tt.Text,
			Tone: tt.Tone, Areas: tt.Areas, Weight: tt.Weight, Strength: round2(float64(tt.Weight) * (0.35 + 0.65*closeness))})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Strength > out[j].Strength })
	return out
}

// areaBias evens out the areas: more transits touch work than mood, so
// without it work days would read better than they feel.
var areaBias = map[string]float64{"love": -2, "work": -6, "mood": 3, "energy": 1}

// scoreDay turns transits and the Moon into area scores (8–96) and a tone.
func scoreDay(ts []dayTransit, phase string, moonHouse int) (map[string]int, int, int) {
	sum := map[string]float64{}
	tone := 0.0
	for _, t := range ts {
		v := float64(t.Tone) * t.Strength
		tone += v
		for _, a := range t.Areas {
			sum[a] += v
		}
		if t.Tone == 0 {
			for _, a := range t.Areas {
				sum[a] += 0.15 * t.Strength // a blend leans slightly lively
			}
		}
	}
	// the Moon's phase: waxing builds energy, the full Moon stirs mood
	switch phase {
	case "new":
		sum["energy"] -= 0.8
		sum["mood"] += 0.3
	case "waxing_crescent", "first_quarter", "waxing_gibbous":
		sum["energy"] += 0.8
		sum["work"] += 0.4
	case "full":
		sum["energy"] += 1
		sum["mood"] -= 0.4
		sum["love"] += 0.6
	case "waning_gibbous", "last_quarter", "waning_crescent":
		sum["energy"] -= 0.4
		sum["mood"] += 0.4
	}
	// the house the Moon walks through today
	switch moonHouse {
	case 5, 7:
		sum["love"] += 1
	case 6, 10:
		sum["work"] += 1
	case 4, 12:
		sum["mood"] += 0.6
		sum["energy"] -= 0.5
	case 1:
		sum["energy"] += 0.8
	}
	scores := map[string]int{}
	total := 0
	for _, a := range areas {
		v := 52 + 3.6*sum[a] + areaBias[a]
		v = 50 + 46*math.Tanh((v-50)/46) // soft edges instead of a wall at 0/100
		n := int(math.Round(math.Max(8, math.Min(96, v))))
		scores[a] = n
		total += n
	}
	t := 0
	if tone > 1.2 {
		t = 1
	} else if tone < -1.2 {
		t = -1
	}
	return scores, int(math.Round(float64(total) / float64(len(areas)))), t
}

func seedOf(parts ...string) int64 {
	h := fnv.New64a()
	for _, p := range parts {
		_, _ = h.Write([]byte(p))
		_, _ = h.Write([]byte{0})
	}
	return int64(h.Sum64() & 0x7fffffffffffffff)
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

func firstSentence(s string) string {
	if i := strings.Index(s, ". "); i > 0 {
		return s[:i+1]
	}
	return s
}

// noonOf is local noon of a day.
func noonOf(day time.Time) time.Time {
	return time.Date(day.Year(), day.Month(), day.Day(), 12, 0, 0, 0, day.Location())
}

// quickDay scores one day (for the week strip and the overlay).
func (s *Server) quickDay(natal astro.Chart, day time.Time, z astro.Zodiac) dayScore {
	at := noonOf(day).UTC()
	sky := astro.Positions(at, z)
	ts := transitsAt(sky, natal)
	ph := astro.MoonPhase(at)
	house := 0
	if mp, ok := findPos(sky, "moon"); ok && natal.Houses != nil {
		house = astro.HouseOf(mp.Lon, *natal.Houses)
	}
	scores, overall, tone := scoreDay(ts, ph.Name, house)
	ds := dayScore{Date: day.Format("2006-01-02"), Overall: overall, Scores: scores, Tone: tone, Phase: ph.Name}
	if mp, ok := findPos(sky, "moon"); ok {
		ds.Moon = signKey(mp.Sign)
	}
	return ds
}

func findPos(ps []astro.Position, b string) (astro.Position, bool) {
	for _, p := range ps {
		if string(p.Body) == b {
			return p, true
		}
	}
	return astro.Position{}, false
}

// readingFor works out a person's day.
func (s *Server) readingFor(ctx context.Context, uid int64, p store.Person, st store.Settings, day time.Time, week bool) (*dayReading, error) {
	natal, err := s.natal(p, st)
	if err != nil {
		return nil, err
	}
	z, _, _ := chartOptions(st)
	at := noonOf(day).UTC()
	sky := astro.Positions(at, z)
	ts := transitsAt(sky, natal)
	ph := astro.MoonPhase(at)
	rd := &dayReading{Date: day.Format("2006-01-02"), PersonID: p.ID, PersonName: p.Name, Transits: []dayTransit{}, Retrograde: []retroView{}, Events: []skyEventView{}, Week: []dayScore{}}
	if mp, ok := findPos(sky, "moon"); ok {
		rd.Moon = dayMoon{Sign: signKey(mp.Sign), Degree: round2(mp.Degree), Phase: ph.Name, Illumination: round2(ph.Illumination), Age: round2(ph.Age),
			Text: content.MoonInSign(signKey(mp.Sign)), PhaseText: content.MoonPhaseText(ph.Name)}
		if natal.Houses != nil {
			rd.Moon.House = astro.HouseOf(mp.Lon, *natal.Houses)
		}
	}
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	for _, v := range astro.VoidOfCourse(dayStart.UTC(), dayStart.AddDate(0, 0, 1).UTC()) {
		rd.Moon.VoidFrom, rd.Moon.VoidTo = v.Start.In(day.Location()).Format(time.RFC3339), v.End.In(day.Location()).Format(time.RFC3339)
		break
	}
	rd.Scores, rd.Overall, rd.Tone = scoreDay(ts, ph.Name, rd.Moon.House)

	// what to show: the strongest few, at most two of them the Moon's
	moons := 0
	for _, t := range ts {
		if len(rd.Transits) >= 6 {
			break
		}
		if t.Transit == "moon" {
			if moons >= 2 {
				continue
			}
			moons++
		}
		rd.Transits = append(rd.Transits, t)
	}
	topArea := "mood"
	best := -1
	for _, a := range areas {
		d := rd.Scores[a] - 50
		if d < 0 {
			d = -d
		}
		if d > best {
			best, topArea = d, a
		}
	}
	// the day number walks through the headline pool, so neighbouring days
	// with the same mood still read differently
	rd.Headline = content.Headline(rd.Tone, topArea, day.Unix()/86400+seedOf(strconv.FormatInt(p.ID, 10))%7)
	parts := []string{}
	if len(rd.Transits) > 0 {
		parts = append(parts, firstSentence(rd.Transits[0].Text))
	}
	if rd.Moon.Text != "" {
		parts = append(parts, firstSentence(rd.Moon.Text))
	}
	rd.Summary = strings.Join(parts, " ")
	rd.Tips = content.Tips(rd.Scores, rd.Moon.Sign, rd.Moon.Phase)
	for _, sp := range sky {
		if sp.Retro && sp.Body != "node" && sp.Body != "sun" && sp.Body != "moon" {
			rd.Retrograde = append(rd.Retrograde, retroView{Body: string(sp.Body), Name: bodyName(string(sp.Body)), Sign: signKey(sp.Sign), Text: content.Retrograde(string(sp.Body))})
		}
	}
	for _, e := range astro.Events(dayStart.UTC(), dayStart.AddDate(0, 0, 1).UTC(), z, false) {
		rd.Events = append(rd.Events, viewEvent(e, day.Location()))
	}
	if week {
		for i := 0; i < 7; i++ {
			rd.Week = append(rd.Week, s.quickDay(natal, dayStart.AddDate(0, 0, i), z))
		}
	}
	if bd, err := time.Parse("2006-01-02", p.BirthDate); err == nil && bd.Month() == day.Month() && bd.Day() == day.Day() {
		rd.Birthday = true
	}
	if r, err := s.store.FindReading(ctx, uid, "daily", p.ID, 0, rd.Date); err == nil {
		rd.Reading = r
	}
	if p.IsMe {
		if j, err := s.store.GetJournalDay(ctx, uid, rd.Date); err == nil {
			rd.Journal = j
		}
	}
	return rd, nil
}

// personParam is ?person= (default: you). The second value says why there is none.
func (s *Server) personParam(ctx context.Context, uid int64, r *http.Request) (*store.Person, string) {
	if id := queryInt(r, "person"); id > 0 {
		p, err := s.store.GetPerson(ctx, uid, int64(id))
		if err != nil {
			return nil, "not found"
		}
		return p, ""
	}
	p, err := s.store.GetMe(ctx, uid)
	if err != nil {
		return nil, "no_me"
	}
	return p, ""
}

// handleToday is the day's reading for you (or ?person=), ?date= for another day.
func (s *Server) handleToday(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	st, _ := s.store.GetSettings(ctx, uid)
	p, why := s.personParam(ctx, uid, r)
	if p == nil {
		if why == "no_me" {
			writeJSON(w, http.StatusOK, map[string]any{"needs_me": true})
			return
		}
		writeErr(w, http.StatusNotFound, why)
		return
	}
	_, _, loc, _ := s.homeOf(ctx, uid, st)
	day, err := localDay(r.URL.Query().Get("date"), loc)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	rd, err := s.readingFor(ctx, uid, *p, st, day, true)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rd)
}

// ---- transits coming up ----

type transitHitView struct {
	astro.TransitHit
	Title  string   `json:"title"`
	Text   string   `json:"text"`
	Tone   int      `json:"tone"`
	Areas  []string `json:"areas"`
	Weight int      `json:"weight"`
	Active bool     `json:"active"`
}

// handleTransits lists the transits of the coming months (?from, ?days ≤ 730, ?slow=1 for Jupiter and beyond only).
func (s *Server) handleTransits(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	st, _ := s.store.GetSettings(ctx, uid)
	p, why := s.personParam(ctx, uid, r)
	if p == nil {
		if why == "no_me" {
			writeJSON(w, http.StatusOK, map[string]any{"needs_me": true})
			return
		}
		writeErr(w, http.StatusNotFound, why)
		return
	}
	natal, err := s.natal(*p, st)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	_, _, loc, _ := s.homeOf(ctx, uid, st)
	from, err := localDay(r.URL.Query().Get("from"), loc)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	days := queryInt(r, "days")
	if days <= 0 {
		days = 90
	}
	days = min(days, 730)
	bodies := []astro.Body{"sun", "mercury", "venus", "mars", "jupiter", "saturn", "uranus", "neptune", "pluto"}
	if queryBool(r, "slow") {
		bodies = bodies[4:]
	}
	hits := s.transitHits(natal, from.UTC(), from.AddDate(0, 0, days).UTC(), bodies)
	now := time.Now()
	out := make([]transitHitView, 0, len(hits))
	for _, h := range hits {
		tt := content.Transit(string(h.Transit), string(h.Type), h.Natal)
		if tt.Title == "" {
			continue
		}
		if h.Exact == nil {
			h.Exact = []time.Time{} // it comes within orb but turns before exact
		}
		out = append(out, transitHitView{TransitHit: h, Title: tt.Title, Text: tt.Text, Tone: tt.Tone, Areas: tt.Areas, Weight: tt.Weight,
			Active: !h.Start.After(now) && !h.End.Before(now)})
	}
	writeJSON(w, http.StatusOK, map[string]any{"person_id": p.ID, "person_name": p.Name, "from": from.Format("2006-01-02"), "days": days, "transits": out})
}

// transitHits is astro.TransitHits against the birth points that matter.
func (s *Server) transitHits(natal astro.Chart, from, to time.Time, bodies []astro.Body) []astro.TransitHit {
	return astro.TransitHits(natal, from, to, bodies, 1)
}
