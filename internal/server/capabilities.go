package server

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	core "github.com/abjelosevic88/pocket-homelab-core/sdk"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/store"
)

// What Pocket Search, Pocket Day and Pocket Hub ask Horoscope (the
// capabilities value in links.go): one route each, in the shape every Pocket
// app answers in, as the request's person. They only read what is stored:
// the readings already written and the journal. None of them writes a
// reading, asks the AI or goes out on the network; a day without a written
// reading stays without one.

// moodWords are the moods as the journal names them (1–5).
var moodWords = [...]string{"", "Low", "Meh", "Okay", "Good", "Great"}

// readingKindWords name a reading that has no title.
var readingKindWords = map[string]string{"daily": "Day reading", "natal": "Birth chart reading", "match": "Match reading", "year": "The year ahead", "question": "Question"}

// hasWord reports whether there is a letter or a digit to look for.
func hasWord(q string) bool {
	return strings.IndexFunc(q, func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }) >= 0
}

// dots joins the parts of a line that are there: "Good mood · a long walk".
func dots(parts ...string) string {
	out := parts[:0:0]
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, " · ")
}

// nameList: "Mileva", "Mileva and Hans", "Mileva, Hans and Eduard".
func nameList(names []string) string {
	switch len(names) {
	case 0:
		return ""
	case 1:
		return names[0]
	}
	return strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
}

// cutWords shortens a line to about n characters at a space, with "…".
func cutWords(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	cut := string(r[:n])
	if i := strings.LastIndex(cut, " "); i > n/2 {
		cut = cut[:i]
	}
	return strings.TrimRight(cut, " ,;:") + "…"
}

// plainLine is one line of a reading without its markdown: no heading,
// quote or list marks, no bold, italics or code.
func plainLine(s string) string {
	s = strings.TrimLeft(strings.TrimSpace(s), "#>-*+ ")
	s = strings.NewReplacer("**", "", "__", "", "*", "", "`", "").Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

// plainText is a whole reading as plain words, headings left out.
func plainText(body string) string {
	var parts []string
	for _, ln := range strings.Split(body, "\n") {
		if t := strings.TrimSpace(ln); t != "" && !strings.HasPrefix(t, "#") {
			if p := plainLine(t); p != "" {
				parts = append(parts, p)
			}
		}
	}
	return strings.Join(parts, " ")
}

// sentenceOf is the first sentence of a line, at most about 200 characters.
func sentenceOf(s string) string {
	end := -1
	for _, mark := range []string{". ", "! ", "? "} {
		if i := strings.Index(s, mark); i > 0 && (end < 0 || i < end) {
			end = i
		}
	}
	if end > 0 {
		s = s[:end+1]
	}
	return cutWords(s, 200)
}

// readingLine is a written reading in one sentence: the first sentence of
// its first line of text (the AI starts a day's reading with a one-line
// summary), a leading "Summary:" dropped. Only headings: the first of them.
func readingLine(body string) string {
	heading := ""
	for _, ln := range strings.Split(body, "\n") {
		t := strings.TrimSpace(ln)
		if t == "" || strings.Trim(t, "-*_= ") == "" {
			continue
		}
		if strings.HasPrefix(t, "#") {
			if heading == "" {
				heading = plainLine(t)
			}
			continue
		}
		p := plainLine(t)
		for _, label := range []string{"one-line summary:", "summary:", "in short:"} {
			if len(p) > len(label) && strings.EqualFold(p[:len(label)], label) {
				p = strings.TrimSpace(p[len(label):])
			}
		}
		if p != "" {
			return sentenceOf(p)
		}
	}
	return sentenceOf(heading)
}

// snippetAround is a short piece of text around the first of the words.
func snippetAround(text string, words []string) string {
	lower := strings.ToLower(text)
	at := -1
	for _, w := range words {
		if i := strings.Index(lower, strings.ToLower(w)); i >= 0 && (at < 0 || i < at) {
			at = i
		}
	}
	if at < 0 || len(lower) != len(text) { // not found, or lower case changed the length
		return cutWords(text, 160)
	}
	prefix := ""
	if start := at - 60; start > 0 {
		if i := strings.IndexByte(text[start:at], ' '); i >= 0 {
			text, prefix = text[start+i+1:], "…"
		}
	}
	return prefix + cutWords(text, 160)
}

// checkinWords is a check-in's mood and energy as the journal says them:
// "Good mood, 4/5 energy"; nothing when neither was set.
func checkinWords(j store.JournalEntry) string {
	var parts []string
	if j.Mood >= 1 && j.Mood <= 5 {
		parts = append(parts, moodWords[j.Mood]+" mood")
	}
	if j.Energy >= 1 && j.Energy <= 5 {
		parts = append(parts, strconv.Itoa(j.Energy)+"/5 energy")
	}
	return strings.Join(parts, ", ")
}

// longDay writes a date as the readings do: "Friday 3 October 2026".
func longDay(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.Format("Monday 2 January 2006")
}

// dayURL opens a day on the Today page (yours, or a person's with their id).
func dayURL(date string, person int64) string {
	if person > 0 {
		return "/?date=" + date + "&person=" + strconv.FormatInt(person, 10)
	}
	return "/?date=" + date
}

// ---------- search ----------

// pocketSearch answers Pocket Search: the readings the person kept (title,
// text and the people they are about) and their journal check-ins (note and
// tags). Every word has to be there. A reading whose title has the words
// comes first, then the check-ins, then the readings with the words in their
// text. Readings that were only written and not kept are a cache and are not
// found; words without a letter or a digit find nothing.
func (s *Server) pocketSearch(r *http.Request, q string, limit int) ([]core.SearchItem, error) {
	var words []string
	for _, w := range strings.Fields(q) {
		if hasWord(w) {
			words = append(words, w)
		}
	}
	if len(words) == 0 {
		return nil, nil
	}
	ctx := r.Context()
	uid := actorID(ctx)
	readings, err := s.store.SearchReadings(ctx, uid, words, limit)
	if err != nil {
		return nil, err
	}
	journal, err := s.store.SearchJournal(ctx, uid, words, limit)
	if err != nil {
		return nil, err
	}
	hasAll := func(v string) bool {
		v = strings.ToLower(v)
		for _, w := range words {
			if !strings.Contains(v, strings.ToLower(w)) {
				return false
			}
		}
		return true
	}
	out := []core.SearchItem{}
	for _, rd := range readings {
		date := rd.Ref
		if len(date) != 10 || !validDate(date) {
			date = rd.CreatedAt[:min(10, len(rd.CreatedAt))]
		}
		it := core.SearchItem{ID: "reading:" + strconv.FormatInt(rd.ID, 10), Title: firstNonEmpty(strings.TrimSpace(rd.Title), readingKindWords[rd.Kind], "Reading"),
			Snippet: snippetAround(plainText(rd.Body), words), Kind: "reading", Date: date, URL: "/readings/" + strconv.FormatInt(rd.ID, 10), Score: 1}
		if hasAll(it.Title) {
			it.Score = 3
		}
		out = append(out, it)
	}
	for _, j := range journal {
		snippet := snippetAround(strings.Join(strings.Fields(j.Note), " "), words)
		if j.Note == "" || !hasAll(j.Note) && hasAll(j.Tags) {
			snippet = dots(snippet, "Tags: "+j.Tags)
		}
		out = append(out, core.SearchItem{ID: "journal:" + strconv.FormatInt(j.ID, 10), Title: "Check-in, " + longDay(j.Date),
			Snippet: dots(checkinWords(j), snippet), Kind: "check-in", Date: j.Date, URL: dayURL(j.Date, 0), Score: 2})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// ---------- day ----------

// pocketDay answers Pocket Day: the day's readings already written (yours
// first, its first sentence as the day's line, then the people you keep),
// with your check-in's note on yours and its mood and energy as numbers.
// Nothing is worked out or written here: a day nobody wrote a reading for
// has only the check-in, if there is one. Days are plain dates here (the
// readings' and the journal's), so the time zone is not used.
func (s *Server) pocketDay(r *http.Request, date string, _ *time.Location) (*core.Day, error) {
	ctx := r.Context()
	uid := actorID(ctx)
	readings, err := s.store.DailyReadings(ctx, uid, date)
	if err != nil {
		return nil, err
	}
	var meID int64
	if me, err := s.store.GetMe(ctx, uid); err == nil {
		meID = me.ID
	} else if !errors.Is(err, store.ErrNotFound) {
		return nil, err
	}
	var check *store.JournalEntry
	if j, err := s.store.GetJournalDay(ctx, uid, date); err == nil {
		check = j
	} else if !errors.Is(err, store.ErrNotFound) {
		return nil, err
	}
	day := &core.Day{Date: date}
	var others []string
	for _, rd := range readings {
		line := readingLine(rd.Body)
		if line == "" {
			continue
		}
		it := core.DayItem{ID: strconv.FormatInt(rd.ID, 10), Title: line, Kind: "reading"}
		if meID > 0 && rd.PersonID != nil && *rd.PersonID == meID {
			it.URL = dayURL(date, 0)
			if check != nil {
				it.Note = strings.TrimSpace(check.Note)
			}
			day.Summary = line
			day.Items = append([]core.DayItem{it}, day.Items...)
			continue
		}
		if rd.PersonID != nil {
			it.URL = dayURL(date, *rd.PersonID)
		} else {
			it.URL = "/readings/" + it.ID
		}
		if rd.PersonName != "" {
			it.Title = rd.PersonName + ": " + line
			others = append(others, rd.PersonName)
		}
		day.Items = append(day.Items, it)
	}
	if check != nil {
		if check.Mood >= 1 && check.Mood <= 5 {
			day.Metrics = append(day.Metrics, core.Metric{Label: "Mood", Value: moodWords[check.Mood]})
		}
		if check.Energy >= 1 && check.Energy <= 5 {
			day.Metrics = append(day.Metrics, core.Metric{Label: "Energy", Value: strconv.Itoa(check.Energy) + "/5"})
		}
	}
	switch {
	case day.Summary != "":
	case check != nil:
		// no reading of yours that day: the check-in is the day's line
		note := strings.TrimSpace(strings.SplitN(strings.TrimSpace(check.Note), "\n", 2)[0])
		day.Summary = firstNonEmpty(dots(checkinWords(*check), note), "Checked in")
	case len(others) == 1:
		day.Summary = "A reading for " + others[0]
	case len(others) > 1:
		day.Summary = "Readings for " + nameList(others)
	case len(day.Items) > 0:
		day.Summary = day.Items[0].Title
	}
	if day.Summary == "" {
		return nil, nil
	}
	return day, nil
}

// ---------- widget ----------

// pocketWidget answers Pocket Hub: the first sentence of today's reading (in
// your home time zone, as the Today page counts days) and your sign. Until
// today's reading is written there is nothing to show; the widget never
// writes it.
func (s *Server) pocketWidget(r *http.Request) (*core.WidgetData, error) {
	ctx := r.Context()
	uid := actorID(ctx)
	me, err := s.store.GetMe(ctx, uid)
	if errors.Is(err, store.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	st, _ := s.store.GetSettings(ctx, uid)
	_, _, loc, _ := s.homeOf(ctx, uid, st)
	today, _ := localDay("", loc)
	rd, err := s.store.FindReading(ctx, uid, "daily", me.ID, 0, today.Format("2006-01-02"))
	if errors.Is(err, store.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	line := readingLine(rd.Body)
	if line == "" {
		return nil, nil
	}
	w := &core.WidgetData{Line: line, URL: "/"}
	if c, err := s.natal(*me, st); err == nil {
		if sun, _, _ := bigThree(c); sun != "" {
			w.Metrics = []core.Metric{{Label: "Sign", Value: signName(signIndex(sun))}}
		}
	}
	return w, nil
}
