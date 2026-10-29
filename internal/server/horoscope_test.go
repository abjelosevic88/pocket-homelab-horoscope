package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/config"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/store"
)

type app struct {
	t     *testing.T
	url   string
	s     *Server
	st    *store.Store
	alice string
	bob   string
	ids   map[string]int64
}

func newApp(t *testing.T, tweak func(*config.Config)) *app {
	t.Helper()
	cfg := testConfig(t.TempDir())
	cfg.BaseURL = "http://horoscope.test"
	cfg.MaxBodyBytes = 64 << 20
	cfg.GeocoderURL = "off"
	if tweak != nil {
		tweak(cfg)
	}
	st := openTestStore(t, cfg)
	ctx := context.Background()
	a := &app{t: t, st: st, ids: map[string]int64{}}
	for _, u := range []*store.User{{Email: "alice@example.com", Name: "Alice", Role: "admin"}, {Email: "bob@example.com", Name: "Bob", Role: "member"}} {
		if err := st.CreateUser(ctx, u); err != nil {
			t.Fatal(err)
		}
		tok, _, _ := st.CreateAPIToken(ctx, u.ID, "t")
		a.ids[u.Name] = u.ID
		if u.Name == "Alice" {
			a.alice = tok
		} else {
			a.bob = tok
		}
	}
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)
	a.s = New(cfg, st, slog.New(slog.NewTextHandler(io.Discard, nil)), nil)
	if err := a.s.Bootstrap(ctx); err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(a.s.Router())
	t.Cleanup(ts.Close)
	a.url = ts.URL
	return a
}

func (a *app) call(auth, method, path string, body any) (int, map[string]any, []any) {
	a.t.Helper()
	var rd io.Reader
	if body != nil {
		j, _ := json.Marshal(body)
		rd = bytes.NewReader(j)
	}
	req, _ := http.NewRequest(method, a.url+path, rd)
	req.Header.Set("Content-Type", "application/json")
	if auth != "" {
		req.Header.Set("Authorization", "Bearer "+auth)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		a.t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	var m map[string]any
	var l []any
	if json.Unmarshal(raw, &m) != nil {
		_ = json.Unmarshal(raw, &l)
	}
	return res.StatusCode, m, l
}

func (a *app) ok(auth, method, path string, body any) map[string]any {
	a.t.Helper()
	code, m, _ := a.call(auth, method, path, body)
	if code >= 300 {
		a.t.Fatalf("%s %s → %d %v", method, path, code, m)
	}
	return m
}

func (a *app) list(auth, path string) []any {
	a.t.Helper()
	code, m, l := a.call(auth, "GET", path, nil)
	if code >= 300 {
		a.t.Fatalf("GET %s → %d %v", path, code, m)
	}
	return l
}

func (a *app) expect(want int, auth, method, path string, body any) map[string]any {
	a.t.Helper()
	code, m, _ := a.call(auth, method, path, body)
	if code != want {
		a.t.Fatalf("%s %s → %d, want %d (%v)", method, path, code, want, m)
	}
	return m
}

func num(m map[string]any, k string) float64 { v, _ := m[k].(float64); return v }

var einstein = map[string]any{"name": "Albert", "relation": "me", "birth_date": "1879-03-14", "birth_time": "11:30", "place": "Ulm",
	"lat": 48.4, "lon": 10.0, "tz": "UTC+00:40"}

func TestPeoplePerPersonAndChart(t *testing.T) {
	a := newApp(t, nil)
	me := a.expect(201, a.alice, "POST", "/api/v1/people", einstein)
	if me["sun"] != "pisces" || me["moon"] != "sagittarius" || me["rising"] != "cancer" || me["is_me"] != true {
		t.Fatalf("big three = %v %v %v", me["sun"], me["moon"], me["rising"])
	}
	id := int64(num(me, "id"))
	ch := a.ok(a.alice, "GET", fmt.Sprintf("/api/v1/people/%d/chart", id), nil)
	if bt := ch["big_three"].([]any); len(bt) != 3 {
		t.Fatalf("big three = %v", bt)
	}
	if hs := ch["houses"].([]any); len(hs) != 12 {
		t.Fatalf("houses = %d", len(hs))
	}
	for _, p := range ch["placements"].([]any) {
		pm := p.(map[string]any)
		if pm["sign_text"] == "" {
			t.Fatalf("no text for %v", pm["body"])
		}
	}
	// someone else's person is not there
	a.expect(404, a.bob, "GET", fmt.Sprintf("/api/v1/people/%d", id), nil)
	a.expect(404, a.bob, "GET", fmt.Sprintf("/api/v1/people/%d/chart", id), nil)
	a.expect(404, a.bob, "PUT", fmt.Sprintf("/api/v1/people/%d", id), map[string]any{"name": "X"})
	a.expect(404, a.bob, "DELETE", fmt.Sprintf("/api/v1/people/%d", id), nil)
	if l := a.list(a.bob, "/api/v1/people"); len(l) != 0 {
		t.Fatalf("bob sees %d people", len(l))
	}
	if m := a.ok(a.bob, "GET", "/api/v1/today", nil); m["needs_me"] != true {
		t.Fatalf("bob's today = %v", m)
	}
	// a second "me" takes the mark
	other := a.expect(201, a.alice, "POST", "/api/v1/people", map[string]any{"name": "Mileva", "relation": "partner", "birth_date": "1875-12-19"})
	if other["rising"] != "" || other["chart_note"] == "" {
		t.Fatalf("no birth time: rising %v, note %v", other["rising"], other["chart_note"])
	}
	a.expect(400, a.alice, "POST", "/api/v1/people", map[string]any{"name": "Bad", "birth_date": "1990-02-31"})
	a.expect(400, a.alice, "POST", "/api/v1/people", map[string]any{"name": "Bad", "birth_date": "1990-02-01", "birth_time": "25:00"})
	a.expect(400, a.alice, "POST", "/api/v1/people", map[string]any{"name": "Bad", "birth_date": "1990-02-01", "tz": "Mars/Olympus"})
}

func TestTodayMatchAndSky(t *testing.T) {
	a := newApp(t, nil)
	a.expect(201, a.alice, "POST", "/api/v1/people", einstein)
	b := a.expect(201, a.alice, "POST", "/api/v1/people", map[string]any{"name": "Mileva", "relation": "partner", "birth_date": "1875-12-19", "birth_time": "08:00",
		"lat": 45.25, "lon": 19.0, "tz": "UTC+01:16"})
	d := a.ok(a.alice, "GET", "/api/v1/today?date=2026-03-03", nil)
	if d["headline"] == "" || d["summary"] == "" || len(d["week"].([]any)) != 7 {
		t.Fatalf("today = %v", d)
	}
	sc := d["scores"].(map[string]any)
	for _, k := range []string{"love", "work", "mood", "energy"} {
		if v := num(sc, k); v < 8 || v > 96 {
			t.Fatalf("score %s = %v", k, v)
		}
	}
	// 3 March 2026: a total lunar eclipse
	found := false
	for _, e := range d["events"].([]any) {
		if e.(map[string]any)["kind"] == "eclipse" {
			found = true
		}
	}
	if !found {
		t.Fatalf("no eclipse on 2026-03-03: %v", d["events"])
	}
	m := a.ok(a.alice, "GET", fmt.Sprintf("/api/v1/match?b=%d", int64(num(b, "id"))), nil)
	if num(m, "overall") < 10 || len(m["aspects"].([]any)) == 0 || m["sun_pair"] == "" {
		t.Fatalf("match = %v", m)
	}
	a.expect(404, a.bob, "GET", fmt.Sprintf("/api/v1/match?a=%d&b=%d", 1, int64(num(b, "id"))), nil)
	tr := a.ok(a.alice, "GET", "/api/v1/transits?from=2026-01-01&days=365&slow=1", nil)
	if len(tr["transits"].([]any)) == 0 {
		t.Fatal("no slow transits in a year")
	}
	mm := a.ok(a.alice, "GET", "/api/v1/sky/moon?month=2026-08", nil)
	if days := mm["days"].([]any); len(days) != 31 {
		t.Fatalf("august has %d days", len(days))
	}
	ov := a.ok(a.alice, "GET", "/api/v1/overlay?from=2026-08-01&to=2026-08-31", nil)
	solar := false
	for _, it := range ov["items"].([]any) {
		if strings.Contains(it.(map[string]any)["title"].(string), "olar eclipse") {
			solar = true
		}
	}
	if !solar {
		t.Fatal("overlay misses the 12 August 2026 solar eclipse")
	}
	rg := a.ok(a.alice, "GET", "/api/v1/sky/retrogrades?year=2026", nil)
	merc := 0
	for _, p := range rg["periods"].([]any) {
		if p.(map[string]any)["body"] == "mercury" {
			merc++
		}
	}
	if merc != 3 {
		t.Fatalf("Mercury retrograde %d times in 2026, want 3", merc)
	}
	a.ok(a.alice, "GET", "/api/v1/sky/now", nil)
	day := a.ok(a.alice, "GET", "/api/v1/day/2026-03-03", nil)
	if day["headline"] == "" || day["moon"] == nil {
		t.Fatalf("day = %v", day)
	}
}

func TestJournalPerPersonAndInsights(t *testing.T) {
	a := newApp(t, nil)
	e := a.ok(a.alice, "POST", "/api/v1/journal", map[string]any{"date": "2026-10-01", "mood": 4, "energy": 3, "note": "fine", "tags": "Work, work, sleep"})
	if e["tags"] != "work, sleep" || e["sky"].(map[string]any)["moon_sign"] == "" {
		t.Fatalf("entry = %v", e)
	}
	// one per day: a second save replaces it
	e2 := a.ok(a.alice, "POST", "/api/v1/journal", map[string]any{"date": "2026-10-01", "mood": 2})
	if e2["id"] != e["id"] || num(e2, "mood") != 2 {
		t.Fatalf("replace = %v", e2)
	}
	id := int64(num(e, "id"))
	a.expect(404, a.bob, "PUT", fmt.Sprintf("/api/v1/journal/%d", id), map[string]any{"mood": 5})
	a.expect(404, a.bob, "DELETE", fmt.Sprintf("/api/v1/journal/%d", id), nil)
	if l := a.list(a.bob, "/api/v1/journal"); len(l) != 0 {
		t.Fatalf("bob sees %d entries", len(l))
	}
	a.expect(400, a.alice, "POST", "/api/v1/journal", map[string]any{"date": "2026-10-02", "mood": 9})
	a.expect(400, a.alice, "POST", "/api/v1/journal", map[string]any{"date": "2026-10-02"})
	ins := a.ok(a.alice, "GET", "/api/v1/journal/insights?days=3660", nil)
	if num(ins, "entries") != 1 || len(ins["phases"].([]any)) != 8 {
		t.Fatalf("insights = %v", ins)
	}
}

func TestReadingsNeedAI(t *testing.T) {
	a := newApp(t, nil)
	a.expect(201, a.alice, "POST", "/api/v1/people", einstein)
	r := a.ok(a.alice, "POST", "/api/v1/readings", map[string]any{"kind": "daily"})
	if r["ok"] != false || r["not_configured"] != true {
		t.Fatalf("without AI = %v", r)
	}
	// a kept reading is per person too
	rd := store.Reading{Kind: "natal", Title: "Albert", Body: "### Hi\n\nText", Saved: true}
	if err := a.st.CreateReading(context.Background(), a.ids["Alice"], &rd); err != nil {
		t.Fatal(err)
	}
	a.expect(404, a.bob, "GET", fmt.Sprintf("/api/v1/readings/%d", rd.ID), nil)
	a.expect(404, a.bob, "DELETE", fmt.Sprintf("/api/v1/readings/%d", rd.ID), nil)
	if l := a.list(a.alice, "/api/v1/readings?saved=1"); len(l) != 1 {
		t.Fatalf("alice readings = %d", len(l))
	}
}

func TestImportAAFAndCSV(t *testing.T) {
	a := newApp(t, nil)
	aaf := "#A93:Einstein,Albert,*,14.3.1879,11:30,Ulm,D\n#B93:2408153.98,48n24,10e00,0hE40,0\n#A93:Doe,Jane,*,1.2.1990,7:05,London,UK\n#B93:*,51n30,0w07,0hE00,0\n"
	r := a.ok(a.alice, "POST", "/api/v1/people/import", map[string]any{"text": aaf})
	if num(r, "added") != 2 {
		t.Fatalf("aaf = %v", r)
	}
	ps := a.list(a.alice, "/api/v1/people")
	var albert map[string]any
	for _, p := range ps {
		if p.(map[string]any)["name"] == "Albert Einstein" {
			albert = p.(map[string]any)
		}
	}
	if albert == nil || albert["rising"] != "cancer" || albert["tz"] != "UTC+00:40" || albert["birth_time"] != "11:30" {
		t.Fatalf("albert = %v", albert)
	}
	r = a.ok(a.alice, "POST", "/api/v1/people/import", map[string]any{"text": aaf})
	if num(r, "added") != 0 || num(r, "skipped") != 2 {
		t.Fatalf("second import = %v", r)
	}
	csv := "name,date,time,place,lat,lon,tz,relation\nAna,1995-06-01,9:15,Sarajevo,43.85,18.36,Europe/Sarajevo,friend\nNo Date,,,,,,,\n"
	r = a.ok(a.alice, "POST", "/api/v1/people/import", map[string]any{"text": csv})
	if num(r, "added") != 1 || len(r["problems"].([]any)) != 1 {
		t.Fatalf("csv = %v", r)
	}
	if l := a.list(a.bob, "/api/v1/people"); len(l) != 0 {
		t.Fatal("import leaked to bob")
	}
}

func TestExportImportIsIdempotent(t *testing.T) {
	a := newApp(t, nil)
	a.expect(201, a.alice, "POST", "/api/v1/people", einstein)
	a.expect(201, a.alice, "POST", "/api/v1/people", map[string]any{"name": "Mileva", "relation": "partner", "birth_date": "1875-12-19"})
	a.ok(a.alice, "POST", "/api/v1/journal", map[string]any{"date": "2026-10-01", "mood": 4})
	_, doc, _ := a.call(a.alice, "GET", "/api/v1/export.json", nil)
	r1 := a.ok(a.bob, "POST", "/api/v1/import", doc)
	if num(r1, "people") != 2 || num(r1, "journal") != 1 {
		t.Fatalf("import = %v", r1)
	}
	a.ok(a.bob, "POST", "/api/v1/import", doc)
	if l := a.list(a.bob, "/api/v1/people"); len(l) != 2 {
		t.Fatalf("second import doubled: %d", len(l))
	}
	if l := a.list(a.alice, "/api/v1/people"); len(l) != 2 {
		t.Fatalf("alice changed: %d", len(l))
	}
	if m := a.ok(a.bob, "GET", "/api/v1/today", nil); m["needs_me"] == true {
		t.Fatal("the imported me should be bob's me")
	}
}

func TestSettingsPerPerson(t *testing.T) {
	a := newApp(t, nil)
	a.ok(a.alice, "PUT", "/api/v1/settings", map[string]any{"zodiac": "sidereal", "notes_token": "pn_secretsecret", "title": "Family Stars", "daily_time": "06:45",
		"home_place": "Sarajevo", "home_lat": 43.85, "home_lon": 18.36, "home_tz": "Europe/Sarajevo"})
	bob := a.ok(a.bob, "GET", "/api/v1/settings", nil)
	if bob["zodiac"] != "tropical" || bob["notes_token"] != "" || bob["title"] != "Family Stars" || bob["home_tz"] != "" {
		t.Fatalf("bob's settings leak: %v", bob)
	}
	a.ok(a.bob, "PUT", "/api/v1/settings", map[string]any{"title": "Bob's"})
	al := a.ok(a.alice, "GET", "/api/v1/settings", nil)
	if al["title"] != "Family Stars" || al["notes_token"] == "" || strings.Contains(al["notes_token"].(string), "secret") || al["daily_time"] != "06:45" {
		t.Fatalf("alice settings = %v", al)
	}
	a.expect(400, a.alice, "PUT", "/api/v1/settings", map[string]any{"daily_time": "25:00"})
	a.expect(400, a.alice, "PUT", "/api/v1/settings", map[string]any{"house_system": "astrodice"})
	a.expect(400, a.alice, "PUT", "/api/v1/settings", map[string]any{"home_tz": "Nowhere/Land"})
}

func TestLinkScopes(t *testing.T) {
	get := func(p string) *http.Request { r, _ := http.NewRequest("GET", p, nil); return r }
	post := func(p string) *http.Request { r, _ := http.NewRequest("POST", p, nil); return r }
	read := []string{"horoscope:read"}
	if !linkAllows(read, get("/api/v1/day/2026-10-23")) || !linkAllows(read, get("/api/v1/overlay")) || !linkAllows(read, get("/api/v1/today")) {
		t.Fatal("read should allow day, overlay and today")
	}
	if linkAllows(read, get("/api/v1/people")) || linkAllows(read, get("/api/v1/people/1/chart")) || linkAllows(read, get("/api/v1/match")) {
		t.Fatal("links never see other people's birth data")
	}
	if linkAllows(read, post("/api/v1/journal")) {
		t.Fatal("read must not write")
	}
	if linkAllows(read, get("/api/v1/settings")) || linkAllows(read, get("/api/v1/export.json")) {
		t.Fatal("links never see settings or exports")
	}
	if !linkAllows([]string{"horoscope:write"}, post("/api/v1/journal")) || linkAllows([]string{"horoscope:write"}, get("/api/v1/journal")) {
		t.Fatal("write may add check-ins but not read")
	}
}

func TestZonesAndOffsets(t *testing.T) {
	for in, want := range map[string]int{"UTC+01:00": 3600, "+0530": 19800, "-3": -10800, "GMT-03:30": -12600, "UTC": 0} {
		if got, ok := parseOffset(in); !ok || got != want {
			t.Fatalf("parseOffset(%q) = %d %v, want %d", in, got, ok, want)
		}
	}
	if _, ok := parseOffset("Europe/Paris"); ok {
		t.Fatal("a zone name is not an offset")
	}
	if !validZone("Europe/Sarajevo") || !validZone("UTC+00:40") || validZone("Atlantis/Main") {
		t.Fatal("validZone")
	}
	at, err := birthMoment(store.Person{BirthDate: "1879-03-14", BirthTime: "11:30", TZ: "UTC+00:40"})
	if err != nil || !at.Equal(time.Date(1879, 3, 14, 10, 50, 0, 0, time.UTC)) {
		t.Fatalf("birth moment = %v %v", at, err)
	}
	if sunSignOf(1, 19) != 9 || sunSignOf(1, 20) != 10 || sunSignOf(12, 22) != 9 || sunSignOf(3, 21) != 0 {
		t.Fatal("sunSignOf")
	}
	if off, ok := aafOffset("5hW00"); !ok || off != -18000 {
		t.Fatalf("aafOffset = %d", off)
	}
}

func TestWithinMinutes(t *testing.T) {
	if !withinMinutes("08:02", "08:00", 4) || withinMinutes("07:59", "08:00", 4) || withinMinutes("08:05", "08:00", 4) {
		t.Fatal("withinMinutes")
	}
}
