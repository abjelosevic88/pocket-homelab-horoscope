package server

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/astro"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/content"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/store"
)

// ---- the sky for everyone: now, the Moon calendar, events ----

type skyEventView struct {
	Kind      string `json:"kind"`
	Time      string `json:"time"`
	Date      string `json:"date"`
	Body      string `json:"body,omitempty"`
	Sign      string `json:"sign"`
	Phase     string `json:"phase,omitempty"`
	Direction string `json:"direction,omitempty"`
	Eclipse   string `json:"eclipse,omitempty"`
	Title     string `json:"title"`
	Text      string `json:"text"`
	Icon      string `json:"icon"`
}

var phaseTitle = map[string]string{"new": "New Moon", "first_quarter": "First quarter Moon", "full": "Full Moon", "last_quarter": "Last quarter Moon"}
var phaseIcon = map[string]string{"new": "🌑", "first_quarter": "🌓", "full": "🌕", "last_quarter": "🌗"}

// viewEvent puts a sky event into words, at local time.
func viewEvent(e astro.SkyEvent, loc *time.Location) skyEventView {
	t := e.Time.In(loc)
	v := skyEventView{Kind: e.Kind, Time: t.Format(time.RFC3339), Date: t.Format("2006-01-02"), Body: string(e.Body), Sign: signKey(e.Sign),
		Phase: e.Phase, Direction: e.Direction, Eclipse: e.Eclipse}
	sn := signName(e.Sign)
	switch e.Kind {
	case "phase":
		v.Title = phaseTitle[e.Phase] + " in " + sn
		v.Icon = phaseIcon[e.Phase]
		v.Text = content.MoonPhaseText(e.Phase)
	case "eclipse":
		kind := "Solar"
		if strings.HasPrefix(e.Eclipse, "lunar") {
			kind = "Lunar"
		}
		sub := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(e.Eclipse, "solar"), "lunar"))
		v.Title = strings.TrimSpace(strings.Title(sub) + " " + strings.ToLower(kind) + " eclipse in " + sn) //nolint:staticcheck // one ASCII word
		if sub == "" {
			v.Title = kind + " eclipse in " + sn
		}
		v.Icon = "🌘"
		if kind == "Solar" {
			v.Icon = "🌑"
		}
		v.Text = content.Eclipse(strings.ToLower(kind), signKey(e.Sign))
	case "ingress":
		v.Title = bodyName(string(e.Body)) + " enters " + sn
		v.Icon = "→"
		v.Text = content.Ingress(string(e.Body), signKey(e.Sign))
	case "station":
		if e.Direction == "retrograde" {
			v.Title = bodyName(string(e.Body)) + " turns retrograde in " + sn
			v.Icon = "℞"
		} else {
			v.Title = bodyName(string(e.Body)) + " turns direct in " + sn
			v.Icon = "D"
		}
		v.Text = content.StationText(string(e.Body), e.Direction)
	}
	return v
}

type skyBody struct {
	Body   string  `json:"body"`
	Name   string  `json:"name"`
	Glyph  string  `json:"glyph"`
	Lon    float64 `json:"lon"`
	Sign   string  `json:"sign"`
	Degree float64 `json:"degree"`
	Where  string  `json:"where"`
	Speed  float64 `json:"speed"`
	Retro  bool    `json:"retro"`
	House  int     `json:"house,omitempty"`
}

// handleSkyNow is the sky right now at home: where everything is, the Moon,
// what is retrograde and what is coming up this week.
func (s *Server) handleSkyNow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	st, _ := s.store.GetSettings(ctx, uid)
	lat, lon, loc, place := s.homeOf(ctx, uid, st)
	z, hs, orbs := chartOptions(st)
	now := time.Now().UTC()
	if v := r.URL.Query().Get("at"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			now = t.UTC()
		}
	}
	hasHome := st.HomeTZ != "" || place != ""
	c := astro.NatalChart(astro.ChartInput{Time: now, TimeKnown: hasHome, Lat: lat, Lon: lon, Zodiac: z, HouseSystem: hs}, orbs)
	bodies := []skyBody{}
	for _, p := range c.Positions {
		pl, _ := content.PlanetByKey(string(p.Body))
		bodies = append(bodies, skyBody{Body: string(p.Body), Name: pl.Name, Glyph: pl.Glyph, Lon: p.Lon, Sign: signKey(p.Sign), Degree: round2(p.Degree),
			Where: fmtLon(p.Lon), Speed: round2(p.Speed), Retro: p.Retro, House: p.House})
	}
	ph := astro.MoonPhase(now)
	moon, _ := position(c, "moon")
	events := []skyEventView{}
	for _, e := range astro.Events(now, now.AddDate(0, 0, 14), z, false) {
		events = append(events, viewEvent(e, loc))
	}
	void := map[string]string{}
	for _, v := range astro.VoidOfCourse(now.Add(-48*time.Hour), now.AddDate(0, 0, 3)) {
		if v.End.After(now) {
			void = map[string]string{"from": v.Start.In(loc).Format(time.RFC3339), "to": v.End.In(loc).Format(time.RFC3339)}
			break
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"at": now.In(loc).Format(time.RFC3339), "place": place, "has_home": hasHome, "zone": loc.String(),
		"chart": c, "bodies": bodies,
		"moon": map[string]any{"sign": signKey(moon.Sign), "degree": round2(moon.Degree), "phase": ph.Name, "illumination": round2(ph.Illumination), "age": round2(ph.Age),
			"text": content.MoonInSign(signKey(moon.Sign)), "phase_text": content.MoonPhaseText(ph.Name), "void": void},
		"events": events,
	})
}

// skyRange reads ?from&to as local dates (defaults: today and 60 days on), at most two years.
func skyRange(r *http.Request, loc *time.Location, days int) (time.Time, time.Time, error) {
	from, err := localDay(r.URL.Query().Get("from"), loc)
	if err != nil {
		return from, from, err
	}
	to := from.AddDate(0, 0, days)
	if v := r.URL.Query().Get("to"); v != "" {
		t, err := localDay(v, loc)
		if err != nil {
			return from, from, err
		}
		to = t.AddDate(0, 0, 1)
	}
	if to.Before(from) {
		from, to = to, from
	}
	if to.Sub(from) > 731*24*time.Hour {
		to = from.AddDate(2, 0, 0)
	}
	return from, to, nil
}

// handleSkyEvents lists the sky events (?kinds=phase,eclipse,ingress,station, ?moon=1 adds Moon sign changes).
func (s *Server) handleSkyEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	st, _ := s.store.GetSettings(ctx, uid)
	_, _, loc, _ := s.homeOf(ctx, uid, st)
	from, to, err := skyRange(r, loc, 60)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	z, _, _ := chartOptions(st)
	kinds := map[string]bool{}
	for _, k := range strings.Split(r.URL.Query().Get("kinds"), ",") {
		if k = strings.TrimSpace(k); k != "" {
			kinds[k] = true
		}
	}
	out := []skyEventView{}
	for _, e := range astro.Events(from.UTC(), to.UTC(), z, queryBool(r, "moon")) {
		if len(kinds) > 0 && !kinds[e.Kind] {
			continue
		}
		out = append(out, viewEvent(e, loc))
	}
	writeJSON(w, http.StatusOK, map[string]any{"from": from.Format("2006-01-02"), "to": to.AddDate(0, 0, -1).Format("2006-01-02"), "zone": loc.String(), "events": out})
}

type moonDay struct {
	Date         string         `json:"date"`
	Phase        string         `json:"phase"`
	Illumination float64        `json:"illumination"`
	Angle        float64        `json:"angle"`
	Sign         string         `json:"sign"`
	SignChange   string         `json:"sign_change,omitempty"` // the time the Moon enters Sign that day
	Events       []skyEventView `json:"events"`
	Void         []string       `json:"void,omitempty"` // from, to
	Mood         int            `json:"mood,omitempty"`
}

// handleMoonMonth is the Moon calendar of a month (?month=2026-10).
func (s *Server) handleMoonMonth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	st, _ := s.store.GetSettings(ctx, uid)
	_, _, loc, _ := s.homeOf(ctx, uid, st)
	z, _, _ := chartOptions(st)
	month := r.URL.Query().Get("month")
	now := time.Now().In(loc)
	first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	if month != "" {
		t, err := time.ParseInLocation("2006-01", month, loc)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "month looks like 2026-10")
			return
		}
		first = t
	}
	next := first.AddDate(0, 1, 0)
	evs := astro.Events(first.UTC(), next.UTC(), z, true)
	voids := astro.VoidOfCourse(first.Add(-72*time.Hour).UTC(), next.UTC())
	moods := map[string]int{}
	if js, err := s.store.ListJournal(ctx, uid, first.Format("2006-01-02"), next.AddDate(0, 0, -1).Format("2006-01-02"), 0); err == nil {
		for _, j := range js {
			moods[j.Date] = j.Mood
		}
	}
	days := []moonDay{}
	for d := first; d.Before(next); d = d.AddDate(0, 0, 1) {
		at := noonOf(d).UTC()
		ph := astro.MoonPhase(at)
		mp := astro.PositionOf("moon", at)
		if z == "sidereal" {
			for _, p := range astro.Positions(at, z) {
				if p.Body == "moon" {
					mp = p
				}
			}
		}
		md := moonDay{Date: d.Format("2006-01-02"), Phase: ph.Name, Illumination: round2(ph.Illumination), Angle: round2(ph.Angle), Sign: signKey(mp.Sign), Events: []skyEventView{}, Mood: moods[d.Format("2006-01-02")]}
		end := d.AddDate(0, 0, 1)
		for _, e := range evs {
			if e.Time.Before(d) || !e.Time.Before(end) {
				continue
			}
			if e.Kind == "ingress" && e.Body == "moon" {
				md.SignChange = e.Time.In(loc).Format(time.RFC3339)
				md.Sign = signKey(e.Sign)
				continue
			}
			md.Events = append(md.Events, viewEvent(e, loc))
		}
		for _, v := range voids {
			if v.Start.Before(end) && v.End.After(d) {
				md.Void = []string{v.Start.In(loc).Format(time.RFC3339), v.End.In(loc).Format(time.RFC3339)}
			}
		}
		days = append(days, md)
	}
	writeJSON(w, http.StatusOK, map[string]any{"month": first.Format("2006-01"), "zone": loc.String(), "days": days})
}

type retroPeriod struct {
	Body  string `json:"body"`
	Name  string `json:"name"`
	Start string `json:"start"`
	End   string `json:"end"`
	Sign  string `json:"sign"`
	Text  string `json:"text"`
}

// handleRetrogrades lists the retrograde periods touching a year (?year=2026).
func (s *Server) handleRetrogrades(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	st, _ := s.store.GetSettings(ctx, uid)
	_, _, loc, _ := s.homeOf(ctx, uid, st)
	z, _, _ := chartOptions(st)
	year := queryInt(r, "year")
	if year < 1800 || year > 2199 {
		year = time.Now().In(loc).Year()
	}
	from := time.Date(year, 1, 1, 0, 0, 0, 0, loc)
	to := from.AddDate(1, 0, 0)
	// look half a year either side so periods crossing the year are whole
	evs := astro.Events(from.AddDate(0, -6, 0).UTC(), to.AddDate(0, 6, 0).UTC(), z, false)
	open := map[astro.Body]astro.SkyEvent{}
	out := []retroPeriod{}
	for _, e := range evs {
		if e.Kind != "station" {
			continue
		}
		if e.Direction == "retrograde" {
			open[e.Body] = e
			continue
		}
		st, ok := open[e.Body]
		if !ok {
			continue
		}
		delete(open, e.Body)
		if e.Time.Before(from) || !st.Time.Before(to) {
			continue
		}
		out = append(out, retroPeriod{Body: string(e.Body), Name: bodyName(string(e.Body)), Start: st.Time.In(loc).Format(time.RFC3339), End: e.Time.In(loc).Format(time.RFC3339),
			Sign: signKey(st.Sign), Text: content.Retrograde(string(e.Body))})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Start < out[j].Start })
	writeJSON(w, http.StatusOK, map[string]any{"year": year, "periods": out})
}

// ---- for Pocket Calendar and Pocket Day ----

var overlayColors = map[string]string{"phase": "#6a3fb5", "eclipse": "#b45309", "station": "#be123c", "ingress": "#0f766e", "transit": "#7c3aed"}

// handleOverlay lists sky events (and the exact days of the big transits to
// you) as calendar items for Pocket Calendar: ?from&to (YYYY-MM-DD).
func (s *Server) handleOverlay(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	st, _ := s.store.GetSettings(ctx, uid)
	_, _, loc, _ := s.homeOf(ctx, uid, st)
	fromS, toS := r.URL.Query().Get("from"), r.URL.Query().Get("to")
	if len(fromS) > 10 {
		fromS = fromS[:10]
	}
	if len(toS) > 10 {
		toS = toS[:10]
	}
	if fromS == "" || toS == "" || !validDate(fromS) || !validDate(toS) {
		writeErr(w, http.StatusBadRequest, "from and to are dates (YYYY-MM-DD)")
		return
	}
	from, _ := localDay(fromS, loc)
	to, _ := localDay(toS, loc)
	to = to.AddDate(0, 0, 1)
	if to.Sub(from) > 400*24*time.Hour {
		to = from.AddDate(0, 0, 400)
	}
	z, _, _ := chartOptions(st)
	items := []map[string]any{}
	for _, e := range astro.Events(from.UTC(), to.UTC(), z, false) {
		if e.Kind == "ingress" && (e.Body == "moon" || e.Body == "node") {
			continue
		}
		v := viewEvent(e, loc)
		t := e.Time.In(loc)
		items = append(items, map[string]any{"id": fmt.Sprintf("sky-%s-%s-%d", e.Kind, e.Body+astro.Body(e.Phase), e.Time.Unix()), "kind": e.Kind,
			"title": strings.TrimSpace(v.Icon + " " + v.Title), "start": t.Format(time.RFC3339), "end": t.Add(30 * time.Minute).Format(time.RFC3339),
			"all_day": false, "color": overlayColors[e.Kind], "url": s.cfg.BaseURL + "/sky"})
	}
	if me, err := s.store.GetMe(ctx, uid); err == nil {
		if natal, err := s.natal(*me, st); err == nil {
			for _, h := range s.transitHits(natal, from.UTC(), to.UTC(), []astro.Body{"jupiter", "saturn", "uranus", "neptune", "pluto"}) {
				tt := content.Transit(string(h.Transit), string(h.Type), h.Natal)
				for _, ex := range h.Exact {
					if ex.Before(from) || !ex.Before(to) {
						continue
					}
					d := ex.In(loc).Format("2006-01-02")
					items = append(items, map[string]any{"id": fmt.Sprintf("transit-%s-%s-%s-%d", h.Transit, h.Type, h.Natal, ex.Unix()), "kind": "transit",
						"title": "✦ " + tt.Title, "start": d, "end": d, "all_day": true, "color": overlayColors["transit"], "url": s.cfg.BaseURL + "/transits"})
				}
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

// handleDay is one day for Pocket Day's journal: the Moon, your day's
// headline and scores, the sky events and your check-in.
func (s *Server) handleDay(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	st, _ := s.store.GetSettings(ctx, uid)
	_, _, loc, _ := s.homeOf(ctx, uid, st)
	day, err := localDay(chiParam(r, "date"), loc)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	z, _, _ := chartOptions(st)
	at := noonOf(day).UTC()
	ph := astro.MoonPhase(at)
	mp := astro.PositionOf("moon", at)
	out := map[string]any{"date": day.Format("2006-01-02"), "url": s.cfg.BaseURL + "/?date=" + day.Format("2006-01-02"),
		"moon": map[string]any{"sign": signKey(mp.Sign), "phase": ph.Name, "illumination": round2(ph.Illumination)}}
	evs := []skyEventView{}
	for _, e := range astro.Events(day.UTC(), day.AddDate(0, 0, 1).UTC(), z, false) {
		evs = append(evs, viewEvent(e, loc))
	}
	out["events"] = evs
	if me, err := s.store.GetMe(ctx, uid); err == nil {
		if rd, err := s.readingFor(ctx, uid, *me, st, day, false); err == nil {
			titles := []string{}
			for _, t := range rd.Transits {
				titles = append(titles, t.Title)
			}
			out["headline"], out["summary"], out["scores"], out["overall"], out["transits"] = rd.Headline, rd.Summary, rd.Scores, rd.Overall, titles
			if rd.Reading != nil {
				out["reading"] = rd.Reading.Body
			}
		}
	}
	if j, err := s.store.GetJournalDay(ctx, uid, day.Format("2006-01-02")); err == nil {
		out["journal"] = map[string]any{"mood": j.Mood, "energy": j.Energy, "note": j.Note, "tags": store.SplitTags(j.Tags)}
	}
	writeJSON(w, http.StatusOK, out)
}

// handleDashboard is the short overview for Hub cards and siblings.
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	st, _ := s.store.GetSettings(ctx, uid)
	_, _, loc, _ := s.homeOf(ctx, uid, st)
	day, _ := localDay("", loc)
	now := time.Now().UTC()
	ph := astro.MoonPhase(now)
	mp := astro.PositionOf("moon", now)
	z, _, _ := chartOptions(st)
	out := map[string]any{"date": day.Format("2006-01-02"), "moon": map[string]any{"sign": signKey(mp.Sign), "phase": ph.Name, "illumination": round2(ph.Illumination)}}
	next := []skyEventView{}
	for _, e := range astro.Events(now, now.AddDate(0, 0, 30), z, false) {
		if e.Kind == "ingress" && e.Body == "moon" {
			continue
		}
		next = append(next, viewEvent(e, loc))
		if len(next) >= 5 {
			break
		}
	}
	out["next"] = next
	if me, err := s.store.GetMe(ctx, uid); err == nil {
		if rd, err := s.readingFor(ctx, uid, *me, st, day, false); err == nil {
			out["headline"], out["scores"], out["overall"], out["tone"] = rd.Headline, rd.Scores, rd.Overall, rd.Tone
		}
		v := s.viewPerson(*me, st, day)
		out["me"] = map[string]any{"name": me.Name, "sun": v.Sun, "moon": v.Moon, "rising": v.Rising}
	}
	ps, _ := s.store.ListPeople(ctx, uid, "")
	out["people"] = len(ps)
	writeJSON(w, http.StatusOK, out)
}
