package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	core "github.com/abjelosevic88/pocket-homelab-core/sdk"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/config"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/store"
)

// fakeCapCore is a Pocket Core that vouches for link tokens
// pcx_<alice|bob>_<search|day|widget|capture|read|write> from Pocket Hub
// and remembers the manifest Horoscope registers with.
type fakeCapCore struct {
	*httptest.Server
	mu       sync.Mutex
	manifest map[string]any
}

func newCapApp(t *testing.T) (*app, *fakeCapCore) {
	t.Helper()
	fc := &fakeCapCore{}
	scopes := map[string]string{"search": core.ScopeSearch, "day": core.ScopeDay, "widget": core.ScopeWidget, "capture": core.ScopeCapture,
		"read": "horoscope:read", "write": "horoscope:write"}
	fc.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/apps/register":
			var m map[string]any
			_ = json.NewDecoder(r.Body).Decode(&m)
			fc.mu.Lock()
			fc.manifest = m
			fc.mu.Unlock()
			io.WriteString(w, `{}`)
		case "/api/v1/auth/introspect":
			var in struct{ Token string }
			_ = json.NewDecoder(r.Body).Decode(&in)
			parts := strings.Split(in.Token, "_")
			if len(parts) != 3 || parts[0] != "pcx" || scopes[parts[2]] == "" {
				io.WriteString(w, `{"active":false}`)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"active": true, "kind": "delegated", "audience": "horoscope", "scope": scopes[parts[2]],
				"user": map[string]any{"id": 1, "sub": parts[1] + "@example.com", "email": parts[1] + "@example.com", "name": parts[1], "role": "member"},
				"app":  map[string]any{"id": 9, "slug": "hub", "name": "Pocket Hub"}})
		default:
			io.WriteString(w, `{}`)
		}
	}))
	t.Cleanup(fc.Close)
	return newApp(t, func(c *config.Config) { c.CoreURL, c.CoreRegistrationKey = fc.URL, "key" }), fc
}

// keepReading stores a written reading as the AI would have left it.
func (a *app) keepReading(who string, kind string, person, other int64, ref, title, body string, saved bool) int64 {
	a.t.Helper()
	r := &store.Reading{Kind: kind, Ref: ref, Title: title, Body: body, Model: "openai", Saved: saved}
	if person > 0 {
		r.PersonID = &person
	}
	if other > 0 {
		r.OtherID = &other
	}
	if err := a.st.CreateReading(context.Background(), a.ids[who], r); err != nil {
		a.t.Fatal(err)
	}
	return r.ID
}

// readingCount is how many readings a person has, kept or not.
func (a *app) readingCount(who string) int {
	a.t.Helper()
	rs, err := a.st.ListReadings(context.Background(), a.ids[who], store.ReadingFilter{})
	if err != nil {
		a.t.Fatal(err)
	}
	return len(rs)
}

// withAI points alice's AI at a server that counts every call, and asks for
// the day's reading to be written by itself: the capability routes must
// still never call it.
func (a *app) withAI() *atomic.Int32 {
	a.t.Helper()
	var calls atomic.Int32
	ai := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "no", http.StatusTeapot)
	}))
	a.t.Cleanup(ai.Close)
	a.ok(a.alice, "PUT", "/api/v1/settings", map[string]any{"ai_provider": "openai", "openai_api_key": "sk-testtesttest", "openai_base_url": ai.URL, "ai_daily": true})
	return &calls
}

// todayOf is today in a person's home zone, as the Today page counts it.
func (a *app) todayOf(who string) string {
	a.t.Helper()
	ctx := context.Background()
	st, _ := a.st.GetSettings(ctx, a.ids[who])
	_, _, loc, _ := a.s.homeOf(ctx, a.ids[who], st)
	d, _ := localDay("", loc)
	return d.Format("2006-01-02")
}

// items are the rows of a capability answer.
func items(t *testing.T, m map[string]any) []map[string]any {
	t.Helper()
	l, ok := m["items"].([]any)
	if !ok {
		t.Fatalf("no items list in %v", m)
	}
	out := []map[string]any{}
	for _, v := range l {
		out = append(out, v.(map[string]any))
	}
	return out
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }

func itemTitles(t *testing.T, m map[string]any) []string {
	t.Helper()
	out := []string{}
	for _, it := range items(t, m) {
		out = append(out, it["title"].(string))
	}
	return out
}

const dayBody = "### Your day\n\n**A soft day to slow down.** The Moon sits in your fourth house, so home pulls at you.\n\nLove and people: be gentle.\n\nTry: an early night."

func TestPocketSearch(t *testing.T) {
	a := newApp(t, nil)
	calls := a.withAI()
	me := int64(num(a.expect(201, a.alice, "POST", "/api/v1/people", einstein), "id"))
	mileva := int64(num(a.expect(201, a.alice, "POST", "/api/v1/people", map[string]any{"name": "Mileva", "relation": "partner", "birth_date": "1875-12-19"}), "id"))
	natal := a.keepReading("Alice", "natal", me, 0, "", "Albert's birth chart", "### Who you are\n\nYour **Venus in Aries** gives a quick heart.\n\n### Work\n\nYou dig deep.", true)
	ask := a.keepReading("Alice", "question", me, 0, "Will I move to Bern?", "Will I move to Bern?", "Jupiter says it is a good year for a move.", true)
	match := a.keepReading("Alice", "match", me, mileva, "", "Albert & Mileva", "You think alike.", true)
	a.keepReading("Alice", "daily", me, 0, "2026-11-20", "Your day, Friday 20 November", "A river of feelings today.", false) // not kept: a cache
	a.ok(a.alice, "POST", "/api/v1/journal", map[string]any{"date": "2026-11-21", "mood": 4, "energy": 3, "note": "Long walk by the river", "tags": "walk, sun"})
	a.ok(a.alice, "POST", "/api/v1/journal", map[string]any{"date": "2026-11-22", "mood": 2, "tags": "rain"})

	res := a.ok(a.alice, "GET", "/api/v1/pocket/search?q=river", nil)
	got := items(t, res)
	if len(got) != 1 || got[0]["id"] != "journal:1" || got[0]["kind"] != "check-in" || got[0]["title"] != "Check-in, Saturday 21 November 2026" ||
		got[0]["snippet"] != "Good mood, 3/5 energy · Long walk by the river" || got[0]["url"] != "/?date=2026-11-21" || got[0]["date"] != "2026-11-21" {
		t.Fatalf("river finds the check-in, not the reading that was never kept: %v", got)
	}
	got = items(t, a.ok(a.alice, "GET", "/api/v1/pocket/search?q=venus", nil))
	if len(got) != 1 || got[0]["id"] != "reading:"+itoa(natal) || got[0]["kind"] != "reading" || got[0]["title"] != "Albert's birth chart" ||
		got[0]["url"] != "/readings/"+itoa(natal) || got[0]["snippet"] != "Your Venus in Aries gives a quick heart. You dig deep." {
		t.Fatalf("venus: %v", got)
	}
	if len(got[0]["date"].(string)) != 10 {
		t.Fatalf("a reading without a day carries the day it was written: %v", got[0])
	}
	// every word has to be there; a title with the words comes first
	if got := items(t, a.ok(a.alice, "GET", "/api/v1/pocket/search?q=quick+heart", nil)); len(got) != 1 || got[0]["id"] != "reading:"+itoa(natal) {
		t.Fatalf("two words: %v", got)
	}
	if got := items(t, a.ok(a.alice, "GET", "/api/v1/pocket/search?q=quick+river", nil)); len(got) != 0 {
		t.Fatalf("words in different things: %v", got)
	}
	a.keepReading("Alice", "year", me, 0, "2027", "The year ahead, 2027", "A year to move house, maybe to Bern.", true)
	got = items(t, a.ok(a.alice, "GET", "/api/v1/pocket/search?q=bern", nil))
	if len(got) != 2 || got[0]["id"] != "reading:"+itoa(ask) || got[0]["date"] == "" || got[1]["title"] != "The year ahead, 2027" {
		t.Fatalf("bern: %v", got)
	}
	// readings about a person by their name; check-ins by their tags
	if got := items(t, a.ok(a.alice, "GET", "/api/v1/pocket/search?q=mileva", nil)); len(got) != 1 || got[0]["id"] != "reading:"+itoa(match) {
		t.Fatalf("mileva: %v", got)
	}
	if got := items(t, a.ok(a.alice, "GET", "/api/v1/pocket/search?q=rain", nil)); len(got) != 1 || got[0]["snippet"] != "Meh mood · Tags: rain" {
		t.Fatalf("by tag: %v", got)
	}
	if got := items(t, a.ok(a.alice, "GET", "/api/v1/pocket/search?q=o&limit=1", nil)); len(got) != 1 {
		t.Fatalf("limit 1: %v", got)
	}
	for _, q := range []string{"", "%3F%3F", "*", "%25", "_"} {
		if got := items(t, a.ok(a.alice, "GET", "/api/v1/pocket/search?q="+q, nil)); len(got) != 0 {
			t.Fatalf("%q finds %v", q, got)
		}
	}
	// another person's readings and journal never appear
	for _, q := range []string{"river", "venus", "bern", "mileva", "rain"} {
		if got := items(t, a.ok(a.bob, "GET", "/api/v1/pocket/search?q="+q, nil)); len(got) != 0 {
			t.Fatalf("bob finds alice's %q: %v", q, got)
		}
	}
	a.expect(401, "", "GET", "/api/v1/pocket/search?q=river", nil)
	if calls.Load() != 0 {
		t.Fatalf("search asked the AI %d times", calls.Load())
	}
}

func TestPocketDay(t *testing.T) {
	a := newApp(t, nil)
	calls := a.withAI()
	const d = "2026-11-20"
	// before anything is written, the day is empty: nothing is worked out or written here
	if day := a.ok(a.alice, "GET", "/api/v1/pocket/day/"+d, nil); day["date"] != d || day["summary"] != nil || len(items(t, day)) != 0 || day["metrics"] != nil {
		t.Fatalf("a day without a reading: %v", day)
	}
	me := int64(num(a.expect(201, a.alice, "POST", "/api/v1/people", einstein), "id"))
	mileva := int64(num(a.expect(201, a.alice, "POST", "/api/v1/people", map[string]any{"name": "Mileva", "relation": "partner", "birth_date": "1875-12-19"}), "id"))
	if day := a.ok(a.alice, "GET", "/api/v1/pocket/day/"+d, nil); day["summary"] != nil || len(items(t, day)) != 0 {
		t.Fatalf("with a chart but no reading: %v", day)
	}
	if n := a.readingCount("Alice"); n != 0 {
		t.Fatalf("the day route wrote %d readings", n)
	}

	a.keepReading("Alice", "daily", me, 0, d, "Your day, Friday 20 November", "An older reading. Replaced.", false)
	mine := a.keepReading("Alice", "daily", me, 0, d, "Your day, Friday 20 November", dayBody, false)
	a.keepReading("Alice", "daily", mileva, 0, d, "Mileva's day, Friday 20 November", "Summary: Talk it through! Then rest.", true)
	a.keepReading("Alice", "daily", me, 0, "2026-11-19", "Your day, Thursday 19 November", "Yesterday's reading.", false)
	a.keepReading("Alice", "match", me, mileva, d, "Albert & Mileva", "Not a day reading.", true)
	a.ok(a.alice, "POST", "/api/v1/journal", map[string]any{"date": d, "mood": 4, "energy": 3, "note": "Walked by the river"})

	day := a.ok(a.alice, "GET", "/api/v1/pocket/day/"+d, nil)
	if day["summary"] != "A soft day to slow down." {
		t.Fatalf("summary: %v", day)
	}
	rows := items(t, day)
	if len(rows) != 2 || rows[0]["id"] != itoa(mine) || rows[0]["title"] != "A soft day to slow down." || rows[0]["kind"] != "reading" ||
		rows[0]["note"] != "Walked by the river" || rows[0]["url"] != "/?date="+d {
		t.Fatalf("yours first, with your note: %v", rows)
	}
	if rows[1]["title"] != "Mileva: Talk it through!" || rows[1]["url"] != "/?date="+d+"&person="+itoa(mileva) || rows[1]["note"] != nil {
		t.Fatalf("the people you keep: %v", rows[1])
	}
	ms := day["metrics"].([]any)
	if len(ms) != 2 || ms[0].(map[string]any)["label"] != "Mood" || ms[0].(map[string]any)["value"] != "Good" || ms[1].(map[string]any)["value"] != "3/5" {
		t.Fatalf("metrics: %v", ms)
	}
	// a check-in without a reading of yours
	a.ok(a.alice, "POST", "/api/v1/journal", map[string]any{"date": "2026-11-21", "mood": 3, "energy": 2, "note": "Tired\nbut fine"})
	if day := a.ok(a.alice, "GET", "/api/v1/pocket/day/2026-11-21", nil); day["summary"] != "Okay mood, 2/5 energy · Tired" || len(items(t, day)) != 0 || len(day["metrics"].([]any)) != 2 {
		t.Fatalf("only a check-in: %v", day)
	}
	// only someone else's reading
	a.keepReading("Alice", "daily", mileva, 0, "2026-11-23", "Mileva's day", "Calm.", false)
	if day := a.ok(a.alice, "GET", "/api/v1/pocket/day/2026-11-23", nil); day["summary"] != "A reading for Mileva" || len(items(t, day)) != 1 {
		t.Fatalf("only Mileva's: %v", day)
	}
	// somebody else's day, a bad date, the old route as before
	if day := a.ok(a.bob, "GET", "/api/v1/pocket/day/"+d+"?tz=Europe/Sarajevo", nil); day["summary"] != nil || len(items(t, day)) != 0 {
		t.Fatalf("bob sees alice's day: %v", day)
	}
	a.expect(400, a.alice, "GET", "/api/v1/pocket/day/yesterday", nil)
	if old := a.ok(a.alice, "GET", "/api/v1/day/"+d, nil); old["headline"] == nil || old["moon"] == nil || old["reading"] != dayBody || old["journal"] == nil {
		t.Fatalf("/day changed: %v", old)
	}
	if calls.Load() != 0 {
		t.Fatalf("the day asked the AI %d times", calls.Load())
	}
}

func TestPocketWidget(t *testing.T) {
	a := newApp(t, nil)
	calls := a.withAI()
	empty := func(tok, why string) {
		t.Helper()
		if w := a.ok(tok, "GET", "/api/v1/pocket/widget", nil); w["line"] != nil || w["metrics"] != nil || len(items(t, w)) != 0 {
			t.Fatalf("%s: %v", why, w)
		}
	}
	empty(a.alice, "without your birth details")
	me := int64(num(a.expect(201, a.alice, "POST", "/api/v1/people", einstein), "id"))
	empty(a.alice, "before today's reading is written")
	today := a.todayOf("Alice")
	a.keepReading("Alice", "daily", me, 0, store.AddDays(today, -1), "Yesterday", "Yesterday's news.", true)
	empty(a.alice, "with only yesterday's reading")
	if n := a.readingCount("Alice"); n != 1 {
		t.Fatalf("the widget wrote a reading: %d", n)
	}
	a.keepReading("Alice", "daily", me, 0, today, "Your day", dayBody, false)
	w := a.ok(a.alice, "GET", "/api/v1/pocket/widget", nil)
	ms, _ := w["metrics"].([]any)
	if w["line"] != "A soft day to slow down." || w["url"] != "/" || len(ms) != 1 || ms[0].(map[string]any)["label"] != "Sign" || ms[0].(map[string]any)["value"] != "Pisces" {
		t.Fatalf("widget: %v", w)
	}
	empty(a.bob, "bob's widget")
	if calls.Load() != 0 {
		t.Fatalf("the widget asked the AI %d times", calls.Load())
	}
}

func TestReadingLine(t *testing.T) {
	for _, c := range [][2]string{
		{dayBody, "A soft day to slow down."},
		{"A plain day. With more.", "A plain day."},
		{"# Your day, Friday\n\n*Go slowly?* Yes.", "Go slowly?"},
		{"---\n\n> One-line summary: Rest wins today.\n", "Rest wins today."},
		{"- First point. Second.", "First point."},
		{"### Only a heading", "Only a heading"},
		{"", ""},
		{strings.Repeat("word ", 80), strings.TrimSpace(strings.Repeat("word ", 40)) + "…"},
	} {
		if got := readingLine(c[0]); got != c[1] {
			t.Errorf("readingLine(%q) = %q, want %q", c[0], got, c[1])
		}
	}
}

// A link token with one pocket:* scope reaches its one route and nothing else.
func TestPocketLinkTokens(t *testing.T) {
	a, _ := newCapApp(t)
	me := int64(num(a.expect(201, a.alice, "POST", "/api/v1/people", einstein), "id"))
	today := a.todayOf("Alice")
	a.keepReading("Alice", "daily", me, 0, today, "Your day", dayBody, true)
	a.ok(a.alice, "POST", "/api/v1/journal", map[string]any{"date": today, "mood": 5, "note": "A river walk"})
	line := map[string]any{"request_id": "req-0000000000000001", "text": "x"}
	for _, c := range []struct {
		tok, method, path string
		body              any
		want              int
	}{
		{"pcx_alice_search", "GET", "/api/v1/pocket/search?q=river", nil, 200},
		{"pcx_alice_day", "GET", "/api/v1/pocket/day/" + today, nil, 200},
		{"pcx_alice_widget", "GET", "/api/v1/pocket/widget", nil, 200},
		{"pcx_alice_widget", "GET", "/api/v1/auth/me", nil, 200},
		// another capability's route; capture is not offered, so there is no such route
		{"pcx_alice_search", "GET", "/api/v1/pocket/day/" + today, nil, 403},
		{"pcx_alice_search", "GET", "/api/v1/pocket/widget", nil, 403},
		{"pcx_alice_day", "GET", "/api/v1/pocket/search?q=river", nil, 403},
		{"pcx_alice_day", "GET", "/api/v1/pocket/widget", nil, 403},
		{"pcx_alice_widget", "GET", "/api/v1/pocket/search?q=river", nil, 403},
		{"pcx_alice_widget", "GET", "/api/v1/pocket/day/" + today, nil, 403},
		{"pcx_alice_capture", "POST", "/api/v1/pocket/capture", line, 404},
		{"pcx_alice_capture", "GET", "/api/v1/pocket/widget", nil, 403},
		// the rest of the app
		{"pcx_alice_widget", "GET", "/api/v1/dashboard", nil, 403},
		{"pcx_alice_widget", "GET", "/api/v1/today", nil, 403},
		{"pcx_alice_search", "GET", "/api/v1/readings", nil, 403},
		{"pcx_alice_search", "GET", "/api/v1/journal", nil, 403},
		{"pcx_alice_search", "GET", "/api/v1/people", nil, 403},
		{"pcx_alice_day", "GET", "/api/v1/day/" + today, nil, 403},
		{"pcx_alice_day", "POST", "/api/v1/readings", map[string]any{"kind": "daily"}, 403},
		{"pcx_alice_widget", "POST", "/api/v1/journal", map[string]any{"mood": 3}, 403},
		{"pcx_alice_widget", "GET", "/api/v1/settings", nil, 403},
		// the app's own scopes do not open the new routes
		{"pcx_alice_read", "GET", "/api/v1/pocket/search?q=river", nil, 403},
		{"pcx_alice_read", "GET", "/api/v1/pocket/day/" + today, nil, 403},
		{"pcx_alice_read", "GET", "/api/v1/pocket/widget", nil, 403},
		{"pcx_alice_write", "POST", "/api/v1/pocket/capture", line, 404},
		{"pcx_alice_read", "GET", "/api/v1/day/" + today, nil, 200},
		{"pcx_nobody", "GET", "/api/v1/pocket/widget", nil, 401},
	} {
		if code, m, _ := a.call(c.tok, c.method, c.path, c.body); code != c.want {
			t.Errorf("%s %s %s → %d, want %d (%v)", c.tok, c.method, c.path, code, c.want, m)
		}
	}
	// the token acts for its person only
	if got := items(t, a.ok("pcx_alice_search", "GET", "/api/v1/pocket/search?q=river", nil)); len(got) != 1 || got[0]["kind"] != "check-in" {
		t.Fatalf("alice through the link: %v", got)
	}
	if got := itemTitles(t, a.ok("pcx_alice_day", "GET", "/api/v1/pocket/day/"+today, nil)); len(got) != 1 || got[0] != "A soft day to slow down." {
		t.Fatalf("alice's day through the link: %v", got)
	}
	if w := a.ok("pcx_alice_widget", "GET", "/api/v1/pocket/widget", nil); w["line"] != "A soft day to slow down." {
		t.Fatalf("alice's widget through the link: %v", w)
	}
	if got := items(t, a.ok("pcx_bob_search", "GET", "/api/v1/pocket/search?q=river", nil)); len(got) != 0 {
		t.Fatalf("bob through the link finds alice's check-in: %v", got)
	}
	if d := a.ok("pcx_bob_day", "GET", "/api/v1/pocket/day/"+today, nil); len(items(t, d)) != 0 || d["summary"] != nil {
		t.Fatalf("bob's day through the link: %v", d)
	}
	if w := a.ok("pcx_bob_widget", "GET", "/api/v1/pocket/widget", nil); w["line"] != nil {
		t.Fatalf("bob's widget through the link: %v", w)
	}
}

func TestPocketLinkScopes(t *testing.T) {
	req := func(m, p string) *http.Request { return httptest.NewRequest(m, p, nil) }
	for _, c := range []struct {
		scope string
		r     *http.Request
		want  bool
	}{
		{core.ScopeSearch, req("GET", "/api/v1/pocket/search?q=x"), true},
		{core.ScopeDay, req("GET", "/api/v1/pocket/day/2026-01-02"), true},
		{core.ScopeWidget, req("GET", "/api/v1/pocket/widget"), true},
		{core.ScopeWidget, req("GET", "/api/v1/auth/me"), true},
		{core.ScopeCapture, req("POST", "/api/v1/pocket/capture"), false},
		{core.ScopeSearch, req("POST", "/api/v1/pocket/search"), false},
		{core.ScopeDay, req("GET", "/api/v1/pocket/day/today"), false},
		{core.ScopeDay, req("GET", "/api/v1/day/2026-01-02"), false},
		{core.ScopeWidget, req("GET", "/api/v1/dashboard"), false},
		{core.ScopeSearch, req("GET", "/api/v1/readings"), false},
		{"horoscope:read", req("GET", "/api/v1/pocket/widget"), false},
		{"horoscope:read", req("GET", "/api/v1/pocket/day/2026-01-02"), false},
		{"horoscope:write", req("POST", "/api/v1/pocket/capture"), false},
	} {
		if got := linkAllows([]string{c.scope}, c.r); got != c.want {
			t.Errorf("%s %s %s = %v, want %v", c.scope, c.r.Method, c.r.URL.Path, got, c.want)
		}
	}
}

// Horoscope tells Core what it can do, so Search, Day and Hub find it on their own.
func TestManifestCapabilities(t *testing.T) {
	a, fc := newCapApp(t)
	if err := a.s.registerWithCore(context.Background()); err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	for i := 0; i < 50; i++ {
		fc.mu.Lock()
		m = fc.manifest
		fc.mu.Unlock()
		if m != nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if m["slug"] != "horoscope" || m["category"] != "Everyday" {
		t.Fatalf("manifest: %v %v", m["slug"], m["category"])
	}
	search, _ := m["search"].(map[string]any)
	day, _ := m["day"].(map[string]any)
	widget, _ := m["widget"].(map[string]any)
	if search == nil || day == nil || widget == nil || m["capture"] != nil {
		t.Fatalf("capabilities: %v %v %v %v", search, day, widget, m["capture"])
	}
	if b := search["bangs"].([]any); len(b) != 2 || b[0] != "horoscope" || b[1] != "stars" {
		t.Fatalf("bangs: %v", b)
	}
	if day["label"] != "Horoscope" || day["one"] != "reading" || day["many"] != "readings" || widget["title"] != "Today's stars" {
		t.Fatalf("day %v widget %v", day, widget)
	}
	offered := map[string]bool{}
	for _, o := range m["offers"].([]any) {
		offered[o.(map[string]any)["scope"].(string)] = true
	}
	for _, sc := range []string{"horoscope:read", "horoscope:write", core.ScopeSearch, core.ScopeDay, core.ScopeWidget} {
		if !offered[sc] {
			t.Errorf("offer %s missing: %v", sc, m["offers"])
		}
	}
	if offered[core.ScopeCapture] {
		t.Errorf("capture is offered: %v", m["offers"])
	}
}
