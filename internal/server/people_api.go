package server

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/astro"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/content"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/store"
)

// ---- people ----

func (s *Server) handleListPeople(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	st, _ := s.store.GetSettings(ctx, uid)
	ps, err := s.store.ListPeople(ctx, uid, r.URL.Query().Get("q"))
	if err != nil {
		s.fail(w, err, "people")
		return
	}
	_, _, loc, _ := s.homeOf(ctx, uid, st)
	today, _ := localDay("", loc)
	out := make([]personView, 0, len(ps))
	for _, p := range ps {
		out = append(out, s.viewPerson(p, st, today))
	}
	writeJSON(w, http.StatusOK, out)
}

type personInput struct {
	Name      *string  `json:"name"`
	Relation  *string  `json:"relation"`
	IsMe      *bool    `json:"is_me"`
	BirthDate *string  `json:"birth_date"`
	BirthTime *string  `json:"birth_time"`
	Place     *string  `json:"place"`
	Country   *string  `json:"country"`
	Lat       *float64 `json:"lat"`
	Lon       *float64 `json:"lon"`
	TZ        *string  `json:"tz"`
	Color     *string  `json:"color"`
	Notes     *string  `json:"notes"`
	Favorite  *bool    `json:"favorite"`
}

// apply copies the fields that were sent and checks the result.
func (in personInput) apply(p *store.Person) string {
	str := func(dst *string, v *string) {
		if v != nil {
			*dst = strings.TrimSpace(*v)
		}
	}
	str(&p.Name, in.Name)
	str(&p.Relation, in.Relation)
	str(&p.BirthDate, in.BirthDate)
	str(&p.BirthTime, in.BirthTime)
	str(&p.Place, in.Place)
	str(&p.Country, in.Country)
	str(&p.TZ, in.TZ)
	str(&p.Color, in.Color)
	if in.Notes != nil {
		p.Notes = *in.Notes
	}
	if in.Lat != nil {
		p.Lat = *in.Lat
	}
	if in.Lon != nil {
		p.Lon = *in.Lon
	}
	if in.IsMe != nil {
		p.IsMe = *in.IsMe
	}
	if in.Favorite != nil {
		p.Favorite = *in.Favorite
	}
	if p.Relation == "" {
		p.Relation = "friend"
	}
	switch {
	case p.Name == "":
		return "a name is needed"
	case len(p.Name) > 120:
		return "the name is too long"
	case !store.ValidRelation(p.Relation):
		return "relation is one of " + strings.Join(store.Relations, ", ")
	case p.BirthTime != "" && !hhmm.MatchString(p.BirthTime):
		return "birth time looks like 07:45"
	case p.Lat < -90 || p.Lat > 90 || p.Lon < -180 || p.Lon > 180:
		return "the coordinates are off the map"
	case !validZone(p.TZ):
		return "unknown time zone " + p.TZ
	}
	d, err := time.Parse("2006-01-02", p.BirthDate)
	if err != nil {
		return "birth date looks like 1990-04-21"
	}
	if d.Year() < 1800 || d.Year() > 2199 {
		return "birth dates from 1800 to 2199 work"
	}
	return ""
}

func (s *Server) handleCreatePerson(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	var in personInput
	if err := decode(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	p := store.Person{Source: "manual"}
	if msg := in.apply(&p); msg != "" {
		writeErr(w, http.StatusBadRequest, msg)
		return
	}
	if p.Relation == "me" {
		p.IsMe = true
	}
	if err := s.store.CreatePerson(ctx, uid, &p); err != nil {
		s.fail(w, err, "add person")
		return
	}
	s.publish(ctx, uid, "horoscope.person_added", fmt.Sprint(p.ID), map[string]any{"name": p.Name, "relation": p.Relation})
	st, _ := s.store.GetSettings(ctx, uid)
	_, _, loc, _ := s.homeOf(ctx, uid, st)
	today, _ := localDay("", loc)
	writeJSON(w, http.StatusCreated, s.viewPerson(p, st, today))
}

func (s *Server) handleGetPerson(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	p, err := s.store.GetPerson(ctx, uid, idParam(r, "id"))
	if err != nil {
		s.fail(w, err, "person")
		return
	}
	st, _ := s.store.GetSettings(ctx, uid)
	_, _, loc, _ := s.homeOf(ctx, uid, st)
	today, _ := localDay("", loc)
	writeJSON(w, http.StatusOK, s.viewPerson(*p, st, today))
}

func (s *Server) handleUpdatePerson(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	p, err := s.store.GetPerson(ctx, uid, idParam(r, "id"))
	if err != nil {
		s.fail(w, err, "person")
		return
	}
	var in personInput
	if err := decode(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if msg := in.apply(p); msg != "" {
		writeErr(w, http.StatusBadRequest, msg)
		return
	}
	if p.Relation == "me" {
		p.IsMe = true
	}
	if err := s.store.UpdatePerson(ctx, uid, p); err != nil {
		s.fail(w, err, "save person")
		return
	}
	st, _ := s.store.GetSettings(ctx, uid)
	_, _, loc, _ := s.homeOf(ctx, uid, st)
	today, _ := localDay("", loc)
	writeJSON(w, http.StatusOK, s.viewPerson(*p, st, today))
}

func (s *Server) handleDeletePerson(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := s.store.DeletePerson(ctx, actorID(ctx), idParam(r, "id")); err != nil {
		s.fail(w, err, "delete person")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ---- the birth chart, with its reading ----

type placementView struct {
	Body      string  `json:"body"`
	Name      string  `json:"name"`
	Glyph     string  `json:"glyph"`
	Lon       float64 `json:"lon"`
	Sign      string  `json:"sign"`
	Degree    float64 `json:"degree"`
	Where     string  `json:"where"`
	House     int     `json:"house"`
	Retro     bool    `json:"retro"`
	Short     string  `json:"short"`
	SignText  string  `json:"sign_text"`
	HouseText string  `json:"house_text"`
}

type aspectView struct {
	astro.Aspect
	AName string `json:"a_name"`
	BName string `json:"b_name"`
	Title string `json:"title"`
	Text  string `json:"text"`
	Tone  string `json:"nature"`
}

type chartView struct {
	Person     personView      `json:"person"`
	Chart      astro.Chart     `json:"chart"`
	BigThree   []placementView `json:"big_three"`
	Placements []placementView `json:"placements"`
	Points     []placementView `json:"points"`
	Aspects    []aspectView    `json:"aspects"`
	Houses     []houseView     `json:"houses"`
	Balance    balanceView     `json:"balance"`
	Settings   map[string]any  `json:"settings"`
}

type houseView struct {
	N       int      `json:"n"`
	Cusp    float64  `json:"cusp"`
	Sign    string   `json:"sign"`
	Name    string   `json:"name"`
	Summary string   `json:"summary"`
	Bodies  []string `json:"bodies"`
}

type balanceView struct {
	Elements   map[string]float64 `json:"elements"`
	Modalities map[string]float64 `json:"modalities"`
	Element    string             `json:"element"`
	Modality   string             `json:"modality"`
	Text       string             `json:"text"`
}

var elementWords = map[string]string{
	"fire":  "Fire leads: you warm up fast, act on enthusiasm and need things to move.",
	"earth": "Earth leads: you trust what is real and useful, and build slowly and well.",
	"air":   "Air leads: you live through ideas, talk and people, and need room to think.",
	"water": "Water leads: you feel your way, read moods quickly and need closeness that is real.",
}

var modalityWords = map[string]string{
	"cardinal": "You are a starter: you like to begin things and set the direction.",
	"fixed":    "You are a keeper: you stay the course and finish what matters to you.",
	"mutable":  "You are an adapter: you change course easily and see many sides.",
}

func top(m map[string]float64) string {
	best, bv := "", -1.0
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if m[k] > bv {
			best, bv = k, m[k]
		}
	}
	return best
}

func (s *Server) placement(key string, lon float64, sign int, deg float64, house int, retro bool) placementView {
	pl, _ := content.PlanetByKey(key)
	v := placementView{Body: key, Name: firstNonEmpty(pl.Name, bodyName(key)), Glyph: pl.Glyph, Lon: lon, Sign: signKey(sign), Degree: deg,
		Where: fmtLon(lon), House: house, Retro: retro, Short: pl.Short}
	v.SignText = content.PlanetInSign(key, v.Sign)
	if house > 0 {
		v.HouseText = content.PlanetInHouse(key, house)
	}
	return v
}

// chartFor builds the full chart view of a person.
func (s *Server) chartFor(p store.Person, st store.Settings, today time.Time) (chartView, error) {
	c, err := s.natal(p, st)
	if err != nil {
		return chartView{}, err
	}
	v := chartView{Person: s.viewPerson(p, st, today), Chart: c, Placements: []placementView{}, Points: []placementView{}, Aspects: []aspectView{}, Houses: []houseView{},
		Settings: map[string]any{"zodiac": st.Zodiac, "house_system": st.HouseSystem, "orbs": st.Orbs, "minor_aspects": st.MinorAspects}}
	for _, ps := range c.Positions {
		pv := s.placement(string(ps.Body), ps.Lon, ps.Sign, ps.Degree, ps.House, ps.Retro)
		v.Placements = append(v.Placements, pv)
		if ps.Body == "sun" || ps.Body == "moon" {
			v.BigThree = append(v.BigThree, pv)
		}
	}
	for _, pt := range c.Points {
		pv := s.placement(pt.Key, pt.Lon, pt.Sign, pt.Degree, pt.House, false)
		v.Points = append(v.Points, pv)
		if pt.Key == "asc" && c.TimeKnown {
			v.BigThree = append(v.BigThree, pv)
		}
	}
	for _, a := range c.Aspects {
		info, _ := content.AspectByKey(string(a.Type))
		v.Aspects = append(v.Aspects, aspectView{Aspect: a, AName: bodyName(a.A), BName: bodyName(a.B),
			Title: bodyName(a.A) + " " + strings.ToLower(info.Name) + " " + bodyName(a.B), Text: content.NatalAspect(a.A, a.B, string(a.Type)), Tone: info.Nature})
	}
	if c.Houses != nil {
		hs := content.Houses()
		for i, cusp := range c.Houses.Cusps {
			hv := houseView{N: i + 1, Cusp: cusp, Sign: signKey(int(cusp / 30)), Bodies: []string{}}
			if i < len(hs) {
				hv.Name, hv.Summary = hs[i].Name, hs[i].Summary
			}
			for _, ps := range c.Positions {
				if ps.House == i+1 {
					hv.Bodies = append(hv.Bodies, string(ps.Body))
				}
			}
			v.Houses = append(v.Houses, hv)
		}
	}
	v.Balance = balanceView{Elements: c.Balance.Elements, Modalities: c.Balance.Modalities}
	v.Balance.Element, v.Balance.Modality = top(c.Balance.Elements), top(c.Balance.Modalities)
	v.Balance.Text = strings.TrimSpace(elementWords[v.Balance.Element] + " " + modalityWords[v.Balance.Modality])
	return v, nil
}

func (s *Server) handlePersonChart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	p, err := s.store.GetPerson(ctx, uid, idParam(r, "id"))
	if err != nil {
		s.fail(w, err, "person")
		return
	}
	st, _ := s.store.GetSettings(ctx, uid)
	_, _, loc, _ := s.homeOf(ctx, uid, st)
	today, _ := localDay("", loc)
	v, err := s.chartFor(*p, st, today)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// handlePreviewChart works out a chart from birth data that is not saved
// (the "add a person" form shows the Big Three as you type).
func (s *Server) handlePreviewChart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	var in personInput
	if err := decode(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	p := store.Person{Name: "Preview"}
	if in.Name == nil || strings.TrimSpace(*in.Name) == "" {
		in.Name = &p.Name
	}
	if msg := in.apply(&p); msg != "" {
		writeErr(w, http.StatusBadRequest, msg)
		return
	}
	st, _ := s.store.GetSettings(ctx, uid)
	_, _, loc, _ := s.homeOf(ctx, uid, st)
	today, _ := localDay("", loc)
	writeJSON(w, http.StatusOK, s.viewPerson(p, st, today))
}

// ---- place search ----

type place struct {
	Name       string  `json:"name"`
	Region     string  `json:"region"`
	Country    string  `json:"country"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	TZ         string  `json:"tz"`
	Population int     `json:"population"`
}

func (s *Server) handlePlaces(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len([]rune(q)) < 2 {
		writeJSON(w, http.StatusOK, []place{})
		return
	}
	if s.cfg.GeocoderURL == "" || s.cfg.GeocoderURL == "off" {
		writeJSON(w, http.StatusOK, []place{})
		return
	}
	ctx := r.Context()
	key := "places:" + strings.ToLower(q)
	if v, at, ok := s.store.CacheGet(ctx, key); ok {
		if t, err := time.Parse(time.RFC3339Nano, at); err == nil && time.Since(t) < 30*24*time.Hour {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_, _ = io.WriteString(w, v)
			return
		}
	}
	out, err := s.geocode(ctx, q)
	if err != nil {
		writeErr(w, http.StatusBadGateway, "place search did not answer: "+err.Error())
		return
	}
	b, _ := json.Marshal(out)
	_ = s.store.CachePut(ctx, key, string(b))
	writeJSON(w, http.StatusOK, out)
}

// geocode asks Open-Meteo's geocoder (no key; it returns the time zone too).
func (s *Server) geocode(ctx context.Context, q string) ([]place, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	u := s.cfg.GeocoderURL + "?" + url.Values{"name": {q}, "count": {"10"}, "language": {"en"}, "format": {"json"}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", s.cfg.UserAgent)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", res.StatusCode)
	}
	var body struct {
		Results []struct {
			Name       string  `json:"name"`
			Admin1     string  `json:"admin1"`
			Country    string  `json:"country"`
			Latitude   float64 `json:"latitude"`
			Longitude  float64 `json:"longitude"`
			Timezone   string  `json:"timezone"`
			Population int     `json:"population"`
		} `json:"results"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&body); err != nil {
		return nil, err
	}
	out := []place{}
	for _, x := range body.Results {
		out = append(out, place{Name: x.Name, Region: x.Admin1, Country: x.Country, Lat: x.Latitude, Lon: x.Longitude, TZ: x.Timezone, Population: x.Population})
	}
	return out, nil
}

// ---- importing birth data ----

// handleImportPeople reads birth data pasted or uploaded as text: Astrodienst
// AAF records (#A93/#B93, what astro.com, Solar Fire and most astrology
// programs export) or CSV with a header (name, date, time, place, lat, lon,
// tz, relation). Someone already here with the same name and birth date is skipped.
func (s *Server) handleImportPeople(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	var in struct {
		Text string `json:"text"`
	}
	if err := decode(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	var people []store.Person
	var problems []string
	if strings.Contains(in.Text, "#A93") {
		people, problems = parseAAF(in.Text)
	} else {
		people, problems = parseCSVPeople(in.Text)
	}
	have, _ := s.store.ListPeople(ctx, uid, "")
	seen := map[string]bool{}
	for _, p := range have {
		seen[strings.ToLower(p.Name)+"|"+p.BirthDate] = true
	}
	added, skipped := 0, 0
	for _, p := range people {
		k := strings.ToLower(p.Name) + "|" + p.BirthDate
		if seen[k] {
			skipped++
			continue
		}
		in := personInput{}
		if msg := in.apply(&p); msg != "" {
			problems = append(problems, p.Name+": "+msg)
			continue
		}
		p.IsMe, p.Source = false, "import"
		if p.Relation == "me" {
			p.Relation = "other"
		}
		if err := s.store.CreatePerson(ctx, uid, &p); err != nil {
			problems = append(problems, p.Name+": "+err.Error())
			continue
		}
		seen[k] = true
		added++
	}
	if problems == nil {
		problems = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"added": added, "skipped": skipped, "problems": problems})
}

// parseAAF reads Astrodienst AAF: "#A93:Name,First,*,14.3.1879,11:30,Ulm,D"
// then "#B93:2408153.98,48n24,10e00,1hE00,0" (the offset at birth, "1hE00" =
// UTC+1, west negative). An unknown time is written "12:00" with "*" in
// many exports; AAF has no flag for it, so times are kept as given.
func parseAAF(text string) ([]store.Person, []string) {
	var out []store.Person
	var problems []string
	var cur *store.Person
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r", ""), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "#A93:"):
			f := strings.Split(strings.TrimPrefix(line, "#A93:"), ",")
			if len(f) < 6 {
				problems = append(problems, "unreadable line: "+line)
				cur = nil
				continue
			}
			name := strings.TrimSpace(strings.Trim(f[1], "*") + " " + strings.Trim(f[0], "*"))
			p := store.Person{Name: strings.TrimSpace(name), Relation: "other", Place: strings.TrimSpace(f[5])}
			if len(f) > 6 {
				p.Country = strings.TrimSpace(f[6])
			}
			d := strings.Split(strings.TrimSpace(f[3]), ".")
			if len(d) == 3 {
				dd, _ := strconv.Atoi(d[0])
				mm, _ := strconv.Atoi(d[1])
				yy, _ := strconv.Atoi(d[2])
				p.BirthDate = fmt.Sprintf("%04d-%02d-%02d", yy, mm, dd)
			}
			p.BirthTime = normTime(f[4])
			out = append(out, p)
			cur = &out[len(out)-1]
		case strings.HasPrefix(line, "#B93:") && cur != nil:
			f := strings.Split(strings.TrimPrefix(line, "#B93:"), ",")
			if len(f) < 4 {
				continue
			}
			if lat, ok := aafCoord(f[1], "n", "s"); ok {
				cur.Lat = lat
			}
			if lon, ok := aafCoord(f[2], "e", "w"); ok {
				cur.Lon = lon
			}
			if off, ok := aafOffset(f[3]); ok {
				cur.TZ = fmtOffset(off)
			}
			cur = nil
		}
	}
	return out, problems
}

// aafCoord reads "48n24" / "10e00" / "73w59".
func aafCoord(v, pos, neg string) (float64, bool) {
	v = strings.ToLower(strings.TrimSpace(v))
	for _, sep := range []string{pos, neg} {
		if i := strings.Index(v, sep); i > 0 {
			d, err1 := strconv.Atoi(v[:i])
			m, err2 := strconv.Atoi(v[i+1:])
			if err1 != nil || err2 != nil {
				return 0, false
			}
			x := float64(d) + float64(m)/60
			if sep == neg {
				x = -x
			}
			return x, true
		}
	}
	return 0, false
}

// aafOffset reads "1hE00" (UTC+1), "5hW00" (UTC-5), "0h00".
func aafOffset(v string) (int, bool) {
	v = strings.ToUpper(strings.TrimSpace(v))
	i := strings.Index(v, "H")
	if i < 1 {
		return 0, false
	}
	h, err := strconv.Atoi(v[:i])
	if err != nil {
		return 0, false
	}
	rest, sign := v[i+1:], 1
	if strings.HasPrefix(rest, "W") {
		sign = -1
	}
	rest = strings.TrimLeft(rest, "EW")
	m, _ := strconv.Atoi(rest)
	return sign * (h*3600 + m*60), true
}

// normTime turns "9:30" / "09:30" / "09:30:00" into "09:30" ("" when unreadable).
func normTime(v string) string {
	v = strings.TrimSpace(v)
	for _, f := range []string{"15:04", "15:04:05", "15.04"} {
		if t, err := time.Parse(f, v); err == nil {
			return t.Format("15:04")
		}
	}
	return ""
}

// parseCSVPeople reads a CSV with a header row.
func parseCSVPeople(text string) ([]store.Person, []string) {
	rd := csv.NewReader(strings.NewReader(text))
	rd.FieldsPerRecord = -1
	rd.TrimLeadingSpace = true
	rows, err := rd.ReadAll()
	if err != nil || len(rows) < 2 {
		return nil, []string{"expected a CSV with a header row (name, date, time, place, lat, lon, tz, relation) or AAF records"}
	}
	col := map[string]int{}
	for i, h := range rows[0] {
		h = strings.ToLower(strings.TrimSpace(h))
		switch h {
		case "birth_date", "birthday", "dob":
			h = "date"
		case "birth_time":
			h = "time"
		case "latitude":
			h = "lat"
		case "longitude", "lng":
			h = "lon"
		case "timezone", "zone":
			h = "tz"
		case "city", "birth_place":
			h = "place"
		}
		col[h] = i
	}
	get := func(row []string, k string) string {
		if i, ok := col[k]; ok && i < len(row) {
			return strings.TrimSpace(row[i])
		}
		return ""
	}
	if _, ok := col["name"]; !ok {
		return nil, []string{"the CSV needs a name column"}
	}
	if _, ok := col["date"]; !ok {
		return nil, []string{"the CSV needs a date column (YYYY-MM-DD)"}
	}
	var out []store.Person
	for _, row := range rows[1:] {
		p := store.Person{Name: get(row, "name"), BirthDate: get(row, "date"), BirthTime: get(row, "time"), Place: get(row, "place"),
			Country: get(row, "country"), TZ: get(row, "tz"), Relation: strings.ToLower(get(row, "relation")), Notes: get(row, "notes")}
		if p.Name == "" {
			continue
		}
		p.Lat, _ = strconv.ParseFloat(get(row, "lat"), 64)
		p.Lon, _ = strconv.ParseFloat(get(row, "lon"), 64)
		p.BirthTime = normTime(p.BirthTime)
		if !store.ValidRelation(p.Relation) {
			p.Relation = "friend"
		}
		out = append(out, p)
	}
	return out, nil
}

// ---- people from Pocket Contacts ----

type contactCandidate struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Birthday string `json:"birthday"`
	HasYear  bool   `json:"has_year"`
	Sun      string `json:"sun"`
	Added    bool   `json:"added"`
}

// handleContactBirthdays lists the person's contacts that have a birthday.
func (s *Server) handleContactBirthdays(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	sib, err := s.clients(ctx, uid).conn(ctx, appContacts, scopeContactsRead)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error(), "not_connected": isNotConnected(err)})
		return
	}
	cs, err := sib.client.Contacts(ctx)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": "Pocket Contacts: " + err.Error()})
		return
	}
	have, _ := s.store.ListPeople(ctx, uid, "")
	added := map[string]bool{}
	for _, p := range have {
		if p.Source == "contacts" {
			added[p.ExternalID] = true
		}
	}
	out := []contactCandidate{}
	for _, c := range cs {
		if c.Birthday == "" || c.Kind == "org" {
			continue
		}
		cc := contactCandidate{ID: c.ID, Name: contactName(c.FN, c.Given, c.Family, c.Nickname), Birthday: c.Birthday, Added: added[fmt.Sprint(c.ID)]}
		if _, err := time.Parse("2006-01-02", c.Birthday); err == nil {
			cc.HasYear = true
		}
		if m, d, ok := monthDay(c.Birthday); ok {
			cc.Sun = astro.Signs[sunSignOf(m, d)]
		}
		out = append(out, cc)
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) })
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "contacts": out})
}

func contactName(fn, given, family, nick string) string {
	if fn != "" {
		return fn
	}
	if n := strings.TrimSpace(given + " " + family); n != "" {
		return n
	}
	return nick
}

// monthDay reads "1990-04-21" or "--04-21".
func monthDay(b string) (int, int, bool) {
	b = strings.TrimPrefix(b, "--")
	parts := strings.Split(b, "-")
	if len(parts) < 2 {
		return 0, 0, false
	}
	m, err1 := strconv.Atoi(parts[len(parts)-2])
	d, err2 := strconv.Atoi(parts[len(parts)-1])
	if err1 != nil || err2 != nil || m < 1 || m > 12 || d < 1 || d > 31 {
		return 0, 0, false
	}
	return m, d, true
}

// sunSignOf is the usual tropical Sun sign of a calendar day (for a list
// before a chart exists; the chart itself uses the real Sun).
func sunSignOf(m, d int) int {
	starts := []int{20, 19, 21, 20, 21, 21, 23, 23, 23, 23, 22, 22} // day the next sign starts, by month (Jan: Aquarius from the 20th)
	sign := (m + 8) % 12                                            // Jan → Capricorn (9)
	if d >= starts[m-1] {
		sign = (sign + 1) % 12
	}
	return sign
}

// handleAddFromContacts adds the chosen contacts as people (birth date only;
// time and place can be added later).
func (s *Server) handleAddFromContacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	var in struct {
		IDs      []int64 `json:"ids"`
		Relation string  `json:"relation"`
	}
	if err := decode(r, &in); err != nil || len(in.IDs) == 0 {
		writeErr(w, http.StatusBadRequest, "choose at least one contact")
		return
	}
	if !store.ValidRelation(in.Relation) || in.Relation == "me" {
		in.Relation = "friend"
	}
	sib, err := s.clients(ctx, uid).conn(ctx, appContacts, scopeContactsRead)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	cs, err := sib.client.Contacts(ctx)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": "Pocket Contacts: " + err.Error()})
		return
	}
	want := map[int64]bool{}
	for _, id := range in.IDs {
		want[id] = true
	}
	added, skipped := 0, 0
	for _, c := range cs {
		if !want[c.ID] {
			continue
		}
		if _, err := time.Parse("2006-01-02", c.Birthday); err != nil {
			skipped++ // no year: no chart
			continue
		}
		if _, err := s.store.FindPersonByExternal(ctx, uid, "contacts", fmt.Sprint(c.ID)); err == nil {
			skipped++
			continue
		}
		p := store.Person{Name: contactName(c.FN, c.Given, c.Family, c.Nickname), Relation: in.Relation, BirthDate: c.Birthday, Color: c.Color,
			Source: "contacts", ExternalID: fmt.Sprint(c.ID)}
		if err := s.store.CreatePerson(ctx, uid, &p); err != nil {
			skipped++
			continue
		}
		added++
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "added": added, "skipped": skipped})
}
