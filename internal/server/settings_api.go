package server

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/ai"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/notify"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/pocket"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/store"
)

// ---- settings ----

func mask(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 8 {
		return "••••"
	}
	return s[:4] + "••••" + s[len(s)-4:]
}

// publicSettings is what one person sees: their settings with the secrets
// masked and, per Pocket app, the mode in effect (<app>_mode), whether it
// works and how (<app>_via: core | direct | "") and why not (<app>_detail).
func (s *Server) publicSettings(ctx context.Context, u *store.User, st store.Settings) map[string]any {
	out := map[string]any{
		"core_configured": s.core.Configured(), "core_registered": s.coreStatus()["registered"],
		"title": st.Title, "date_format": st.DateFormat, "time_format": st.TimeFormat, "week_start": st.WeekStart,
		"zodiac": st.Zodiac, "house_system": st.HouseSystem, "orbs": st.Orbs, "minor_aspects": st.MinorAspects,
		"home_place": st.HomePlace, "home_country": st.HomeCountry, "home_lat": st.HomeLat, "home_lon": st.HomeLon, "home_tz": st.HomeTZ,
		"daily_notify": st.DailyNotify, "daily_time": st.DailyTime, "notify_moon": st.NotifyMoon, "notify_retro": st.NotifyRetro,
		"notify_transits": st.NotifyTransits, "notify_birthdays": st.NotifyBirthdays,
		"notify_url": st.NotifyURL, "notify_format": st.NotifyFormat, "notify_configured": st.NotifyURL != "",
		"contacts_url": st.ContactsURL, "contacts_token": mask(st.ContactsToken), "contacts_public_url": st.ContactsPublicURL,
		"notes_url": st.NotesURL, "notes_token": mask(st.NotesToken), "notes_public_url": st.NotesPublicURL, "notes_folder": st.NotesFolder,
		"health_url": st.HealthURL, "health_token": mask(st.HealthToken), "health_public_url": st.HealthPublicURL, "health_insights": st.HealthInsights,
		"ai_provider": st.AIProvider, "ai_pocket_model": st.AIPocketModel, "openai_api_key": mask(st.OpenAIKey), "openai_model": st.OpenAIModel, "openai_base_url": st.OpenAIBaseURL,
		"ai_daily": st.AIDaily, "ai_tone": st.AITone,
		"setup_complete": st.SetupComplete, "base_url": s.cfg.BaseURL, "geocoder": s.cfg.GeocoderURL != "" && s.cfg.GeocoderURL != "off",
	}
	for _, a := range pocketApps {
		mode, via, detail := s.connStatus(ctx, u, st, a)
		out[a.slug+"_mode"], out[a.slug+"_via"], out[a.slug+"_detail"] = mode, via, detail
		out[a.slug+"_configured"] = via != ""
		_, _, _, public := connSettings(st, a)
		if public == "" && via != "" {
			public = s.resolveSibling(ctx, a.slug)
		}
		out[a.slug+"_open_url"] = public
	}
	return out
}

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	st, err := s.store.GetSettings(r.Context(), actorID(r.Context()))
	if err != nil {
		s.fail(w, err, "settings")
		return
	}
	writeJSON(w, http.StatusOK, s.publicSettings(r.Context(), userFrom(r.Context()), st))
}

var hhmm = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

// handlePutSettings saves the caller's personal settings. The box-wide title
// is only taken from an admin.
func (s *Server) handlePutSettings(w http.ResponseWriter, r *http.Request) {
	me := userFrom(r.Context())
	isAdmin := me.Role == "admin"
	cur, err := s.store.GetSettings(r.Context(), me.ID)
	if err != nil {
		s.fail(w, err, "settings")
		return
	}
	var in store.Settings
	keys, err := decodeKeys(r, &in)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	bad := func(msg string) { writeErr(w, http.StatusBadRequest, msg) }
	if keys["title"] && isAdmin {
		cur.Title = strings.TrimSpace(in.Title)
		if cur.Title == "" {
			cur.Title = store.DefaultTitle
		}
	}
	if keys["date_format"] {
		cur.DateFormat = in.DateFormat
	}
	if keys["time_format"] {
		cur.TimeFormat = in.TimeFormat
	}
	if keys["week_start"] && (in.WeekStart == 1 || in.WeekStart == 7 || in.WeekStart == 6) {
		cur.WeekStart = in.WeekStart
	}
	pick := func(k string, dst *string, v string, allowed []string) bool {
		if !keys[k] {
			return true
		}
		for _, a := range allowed {
			if v == a {
				*dst = v
				return true
			}
		}
		bad(fmt.Sprintf("%s must be one of %s", k, strings.Join(allowed, ", ")))
		return false
	}
	if !pick("zodiac", &cur.Zodiac, in.Zodiac, []string{"tropical", "sidereal"}) ||
		!pick("house_system", &cur.HouseSystem, in.HouseSystem, []string{"placidus", "koch", "whole", "equal", "porphyry"}) ||
		!pick("orbs", &cur.Orbs, in.Orbs, []string{"tight", "normal", "wide"}) ||
		!pick("ai_tone", &cur.AITone, in.AITone, []string{"gentle", "straight", "playful"}) {
		return
	}
	if keys["daily_time"] {
		if !hhmm.MatchString(in.DailyTime) {
			bad("the time looks like 07:30")
			return
		}
		cur.DailyTime = in.DailyTime
	}
	if keys["home_tz"] || keys["home_lat"] || keys["home_lon"] || keys["home_place"] {
		if !validZone(in.HomeTZ) || in.HomeLat < -90 || in.HomeLat > 90 || in.HomeLon < -180 || in.HomeLon > 180 {
			bad("choose a place from the list (or a valid time zone and coordinates)")
			return
		}
		cur.HomePlace, cur.HomeCountry, cur.HomeLat, cur.HomeLon, cur.HomeTZ = strings.TrimSpace(in.HomePlace), strings.TrimSpace(in.HomeCountry), in.HomeLat, in.HomeLon, in.HomeTZ
	}
	secret := func(k string, dst *string, v string) {
		if keys[k] && !strings.Contains(v, "••") {
			*dst = strings.TrimSpace(v)
		}
	}
	url := func(k string, dst *string, v string) {
		if keys[k] {
			*dst = strings.TrimRight(strings.TrimSpace(v), "/")
		}
	}
	text := func(k string, dst *string, v string) {
		if keys[k] {
			*dst = strings.TrimSpace(v)
		}
	}
	boolIn := func(k string, dst *bool, v bool) {
		if keys[k] {
			*dst = v
		}
	}
	boolIn("minor_aspects", &cur.MinorAspects, in.MinorAspects)
	boolIn("daily_notify", &cur.DailyNotify, in.DailyNotify)
	boolIn("notify_moon", &cur.NotifyMoon, in.NotifyMoon)
	boolIn("notify_retro", &cur.NotifyRetro, in.NotifyRetro)
	boolIn("notify_transits", &cur.NotifyTransits, in.NotifyTransits)
	boolIn("notify_birthdays", &cur.NotifyBirthdays, in.NotifyBirthdays)
	boolIn("health_insights", &cur.HealthInsights, in.HealthInsights)
	boolIn("ai_daily", &cur.AIDaily, in.AIDaily)
	text("notify_url", &cur.NotifyURL, in.NotifyURL)
	if keys["notify_format"] && in.NotifyFormat != "" {
		cur.NotifyFormat = in.NotifyFormat
	}
	url("contacts_url", &cur.ContactsURL, in.ContactsURL)
	secret("contacts_token", &cur.ContactsToken, in.ContactsToken)
	url("contacts_public_url", &cur.ContactsPublicURL, in.ContactsPublicURL)
	url("notes_url", &cur.NotesURL, in.NotesURL)
	secret("notes_token", &cur.NotesToken, in.NotesToken)
	url("notes_public_url", &cur.NotesPublicURL, in.NotesPublicURL)
	if keys["notes_folder"] {
		cur.NotesFolder = strings.Trim(strings.TrimSpace(in.NotesFolder), "/")
	}
	url("health_url", &cur.HealthURL, in.HealthURL)
	secret("health_token", &cur.HealthToken, in.HealthToken)
	url("health_public_url", &cur.HealthPublicURL, in.HealthPublicURL)
	if keys["ai_provider"] {
		if !validAIProvider(in.AIProvider) {
			bad("AI is automatic, Pocket AI or OpenAI")
			return
		}
		cur.AIProvider = in.AIProvider
	}
	if keys["ai_pocket_model"] {
		if !validPocketModel(in.AIPocketModel) {
			bad("unknown model")
			return
		}
		cur.AIPocketModel = in.AIPocketModel
	}
	secret("openai_api_key", &cur.OpenAIKey, in.OpenAIKey)
	text("openai_model", &cur.OpenAIModel, in.OpenAIModel)
	url("openai_base_url", &cur.OpenAIBaseURL, in.OpenAIBaseURL)
	mode := func(k string, dst *string, v string) {
		if keys[k] && (v == modeCore || v == modeDirect || v == modeOff) {
			*dst = v
		}
	}
	mode("contacts_mode", &cur.ContactsMode, in.ContactsMode)
	mode("notes_mode", &cur.NotesMode, in.NotesMode)
	mode("health_mode", &cur.HealthMode, in.HealthMode)
	if err := s.store.SaveUserSettings(r.Context(), me.ID, cur); err != nil {
		s.fail(w, err, "save settings")
		return
	}
	if isAdmin && keys["title"] {
		if err := s.store.SaveBoxSettings(r.Context(), cur); err != nil {
			s.fail(w, err, "save settings")
			return
		}
	}
	for _, k := range modeKeys {
		if keys[k] {
			s.reportModesSoon()
			break
		}
	}
	writeJSON(w, http.StatusOK, s.publicSettings(r.Context(), me, cur))
}

// handleIntegrationTest checks a connection with the caller's own saved
// settings (or the URL/key typed into the form).
func (s *Server) handleIntegrationTest(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Which  string `json:"which"`
		URL    string `json:"url"`
		Key    string `json:"key"`
		User   string `json:"user"`
		Format string `json:"format"`
	}
	if err := decode(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	uid := actorID(r.Context())
	st, _ := s.store.GetSettings(r.Context(), uid)
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	keyOr := func(saved string) string {
		if strings.Contains(in.Key, "••") || in.Key == "" {
			return saved
		}
		return in.Key
	}
	urlOr := func(saved string) string {
		if strings.TrimSpace(in.URL) == "" {
			return saved
		}
		return strings.TrimRight(strings.TrimSpace(in.URL), "/")
	}
	var msg string
	var err error
	if a, ok := appBySlug(in.Which); ok {
		mode, savedURL, savedTok, _ := connSettings(st, a)
		typed := in.Key != "" && !strings.Contains(in.Key, "••")
		var conn *siblingConn
		if tok := keyOr(savedTok); tok != "" && (typed || s.connMode(mode, savedTok) == modeDirect) {
			base := urlOr(savedURL)
			if base == "" {
				base = s.resolveSibling(ctx, a.slug)
			}
			conn = &siblingConn{client: pocket.New(base, tok, 15*time.Second), via: modeDirect}
		} else {
			conn, err = s.clients(ctx, uid).conn(ctx, a, a.scopes...)
		}
		if err == nil {
			how := "direct connection"
			if conn.via == modeCore {
				how = "through Pocket Core"
			}
			var v string
			if v, err = conn.client.Health(ctx); err == nil {
				switch a.slug {
				case "contacts":
					var cs []pocket.Contact
					cs, err = conn.client.Contacts(ctx)
					n := 0
					for _, c := range cs {
						if c.Birthday != "" {
							n++
						}
					}
					msg = fmt.Sprintf("Pocket Contacts %s, %s, %d contacts with a birthday", v, how, n)
				case "notes":
					msg = fmt.Sprintf("Pocket Notes %s, %s", v, how)
				case "health":
					var pts []pocket.DayValue
					pts, err = conn.client.Metric(ctx, "sleep_asleep", "m", "")
					msg = fmt.Sprintf("Pocket Health %s, %s, %d nights of sleep this month", v, how, len(pts))
				}
			}
		}
	} else if in.Which == "openai" {
		c := ai.New(urlOr(st.OpenAIBaseURL), keyOr(st.OpenAIKey), st.OpenAIModel, 20*time.Second)
		var ids []string
		if ids, err = c.Models(ctx); err == nil {
			msg = fmt.Sprintf("The key works (%d models)", len(ids))
		}
	} else if in.Which == "notify" {
		f := in.Format
		if f == "" {
			f = st.NotifyFormat
		}
		n := notify.New(urlOr(st.NotifyURL), f)
		err = n.Send(ctx, notify.Message{Title: "Pocket Horoscope", Body: "Notifications work. Your daily horoscope and the Moon's big days will arrive here.", Priority: "info", Tags: []string{"crystal_ball"}})
		msg = "Test message sent via " + n.Format
	} else {
		writeErr(w, http.StatusBadRequest, "unknown integration")
		return
	}
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": msg})
}

func (s *Server) handleNotifyTest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	st, _ := s.store.GetSettings(ctx, actorID(ctx))
	n := notify.New(st.NotifyURL, st.NotifyFormat)
	if err := n.Send(ctx, notify.Message{Title: "Pocket Horoscope", Body: "Notifications work.", Priority: "info", Tags: []string{"crystal_ball"}}); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// ---- system / backup / export ----

func (s *Server) handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	var size int64
	if fi, err := os.Stat(s.cfg.DBPath); err == nil {
		size = fi.Size()
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"version": Version, "go": runtime.Version(), "os": runtime.GOOS, "arch": runtime.GOARCH,
		"core":      s.coreStatus(),
		"auth_mode": s.cfg.AuthMode, "data_dir": s.cfg.DataDir, "db_path": s.cfg.DBPath, "db_bytes": size,
		"uptime": time.Since(s.started).Round(time.Second).String(), "timezone": time.Now().Location().String(),
		"counts": s.store.CountsFor(r.Context(), actorID(r.Context())), "saved": s.saved.Load(),
	})
}

// handleBackup downloads the whole database, an admin-only box operation.
func (s *Server) handleBackup(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	dir := filepath.Join(s.cfg.DataDir, "backups")
	name := fmt.Sprintf("pocket-horoscope-%s.db", time.Now().Format("20060102-150405"))
	dst := filepath.Join(dir, name)
	if err := s.store.Backup(r.Context(), dst); err != nil {
		s.fail(w, err, "backup")
		return
	}
	w.Header().Set("Content-Type", "application/vnd.sqlite3")
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	http.ServeFile(w, r, dst)
}

// exportDoc is the JSON export of one person's data.
type exportDoc struct {
	App      string               `json:"app"`
	Version  string               `json:"version"`
	Exported string               `json:"exported_at"`
	People   []store.Person       `json:"people"`
	Journal  []store.JournalEntry `json:"journal"`
	Readings []store.Reading      `json:"readings"`
}

func (s *Server) handleExportJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	doc := exportDoc{App: "pocket-horoscope", Version: Version, Exported: store.Now()}
	var err error
	if doc.People, err = s.store.ListPeople(ctx, uid, ""); err != nil {
		s.fail(w, err, "export")
		return
	}
	doc.Journal, _ = s.store.ListJournal(ctx, uid, "", "", 0)
	doc.Readings, _ = s.store.ListReadings(ctx, uid, store.ReadingFilter{SavedOnly: true})
	w.Header().Set("Content-Disposition", `attachment; filename="pocket-horoscope-export.json"`)
	writeJSON(w, http.StatusOK, doc)
}

// handleExportCSV writes the people (birth data) as CSV, the same columns
// the importer reads.
func (s *Server) handleExportCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ps, err := s.store.ListPeople(ctx, actorID(ctx), "")
	if err != nil {
		s.fail(w, err, "export")
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="pocket-horoscope-people.csv"`)
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"name", "relation", "date", "time", "place", "country", "lat", "lon", "tz", "notes"})
	for _, p := range ps {
		_ = cw.Write([]string{p.Name, p.Relation, p.BirthDate, p.BirthTime, p.Place, p.Country, strconv.FormatFloat(p.Lat, 'f', 5, 64),
			strconv.FormatFloat(p.Lon, 'f', 5, 64), p.TZ, p.Notes})
	}
	cw.Flush()
}

// handleImportJSON merges an export into the caller's own data. People are
// matched by name and birth date, check-ins by day and readings by their
// kind, subject and time, so importing the same file twice adds nothing.
func (s *Server) handleImportJSON(w http.ResponseWriter, r *http.Request) {
	var doc exportDoc
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, s.cfg.MaxBodyBytes))
	if err := dec.Decode(&doc); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if doc.App != "" && doc.App != "pocket-horoscope" {
		writeErr(w, http.StatusBadRequest, "this is not a Pocket Horoscope export")
		return
	}
	ctx := r.Context()
	uid := actorID(ctx)
	counts := map[string]int{"people": 0, "journal": 0, "readings": 0, "skipped": 0}
	have, _ := s.store.ListPeople(ctx, uid, "")
	byKey := map[string]int64{}
	hasMe := false
	for _, p := range have {
		byKey[strings.ToLower(p.Name)+"|"+p.BirthDate] = p.ID
		hasMe = hasMe || p.IsMe
	}
	ids := map[int64]int64{}
	for _, p := range doc.People {
		k := strings.ToLower(p.Name) + "|" + p.BirthDate
		if id, ok := byKey[k]; ok {
			ids[p.ID] = id
			continue
		}
		old := p.ID
		np := p
		np.ID, np.ExternalID = 0, ""
		if np.Source != "manual" {
			np.Source = "import"
		}
		if np.IsMe && hasMe {
			np.IsMe, np.Relation = false, "other"
		}
		if msg := (personInput{}).apply(&np); msg != "" {
			counts["skipped"]++
			continue
		}
		if err := s.store.CreatePerson(ctx, uid, &np); err != nil {
			counts["skipped"]++
			continue
		}
		hasMe = hasMe || np.IsMe
		byKey[k], ids[old] = np.ID, np.ID
		counts["people"]++
	}
	for _, e := range doc.Journal {
		if _, err := s.store.GetJournalDay(ctx, uid, e.Date); err == nil {
			continue
		}
		e.ID, e.Source = 0, "import"
		if msg := (journalInput{}).apply(&e); msg != "" {
			counts["skipped"]++
			continue
		}
		if err := s.store.SaveJournal(ctx, uid, &e); err != nil {
			counts["skipped"]++
			continue
		}
		counts["journal"]++
	}
	kept, _ := s.store.ListReadings(ctx, uid, store.ReadingFilter{})
	seen := map[string]bool{}
	for _, x := range kept {
		seen[x.Kind+"|"+x.Ref+"|"+x.Title+"|"+x.Body] = true
	}
	for _, x := range doc.Readings {
		if seen[x.Kind+"|"+x.Ref+"|"+x.Title+"|"+x.Body] {
			continue
		}
		mapID := func(id *int64) *int64 {
			if id == nil {
				return nil
			}
			if n, ok := ids[*id]; ok {
				return &n
			}
			return nil
		}
		x.PersonID, x.OtherID, x.Saved = mapID(x.PersonID), mapID(x.OtherID), true
		if err := s.store.CreateReading(ctx, uid, &x); err != nil {
			counts["skipped"]++
			continue
		}
		counts["readings"]++
	}
	writeJSON(w, http.StatusOK, counts)
}
