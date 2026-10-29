package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	core "github.com/abjelosevic88/pocket-homelab-core/sdk"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/ai"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/astro"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/store"
)

// aiFor is the person's AI: Pocket AI through Core, their own
// OpenAI-compatible key, or nil with why.
func (s *Server) aiFor(ctx context.Context, uid int64, st store.Settings) (*ai.Client, string, string) {
	email := ""
	if u, err := s.store.GetUser(ctx, uid); err == nil {
		email = u.Email
	}
	pocket, status := s.core.SelectAI(ctx, email, st.AIProvider, st.OpenAIKey != "")
	if pocket != nil {
		if !status.Ready {
			return nil, "", "Pocket AI: " + status.Detail
		}
		c := ai.New("", "", "", 150*time.Second)
		c.Pocket = pocket.WithModel(st.AIPocketModel)
		return c, "pocket", ""
	}
	if st.OpenAIKey == "" {
		return nil, "", status.Detail
	}
	return ai.New(st.OpenAIBaseURL, st.OpenAIKey, st.OpenAIModel, 150*time.Second), "openai", ""
}

func (s *Server) handleAIStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	st, _ := s.store.GetSettings(ctx, actorID(ctx))
	cl, used, why := s.aiFor(ctx, actorID(ctx), st)
	writeJSON(w, http.StatusOK, map[string]any{"ready": cl != nil, "provider": used, "detail": why})
}

func (s *Server) handleAIModels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	c, status := s.core.SelectAI(ctx, userFrom(ctx).Email, "pocket", false)
	if c == nil || !status.Ready {
		writeErr(w, http.StatusServiceUnavailable, status.Detail)
		return
	}
	models, err := c.Models(ctx)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, models)
}

func validAIProvider(v string) bool { return core.ValidAIProvider(v) }

func validPocketModel(v string) bool { return core.ValidInferenceModel(v) }

var toneWords = map[string]string{
	"gentle":   "Warm and encouraging, never sugary.",
	"straight": "Plain and direct, like a friend who tells it as it is; no softening phrases.",
	"playful":  "Light and witty, with a little humour, still useful.",
}

func aiVoice(st store.Settings) string {
	return `You are a thoughtful astrologer writing for someone who knows no astrology. You get worked-out chart facts (planet positions, houses, aspects, transits with their exact orbs and scores); use only those facts and never invent placements, dates or numbers. Explain any astrology word you use in a few plain words. Astrology here is a mirror for reflection, not fate: no predictions of events, no fear, no medical, legal or financial advice. Write in second person ("you"), short paragraphs, plain markdown (a few ### headings and short lists at most), no emojis. ` + toneWords[firstNonEmpty(st.AITone, "gentle")]
}

// chartFacts is a chart in compact words for the AI.
func chartFacts(name string, c astro.Chart) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s's birth chart (%s zodiac", name, c.Zodiac)
	if c.Houses != nil {
		fmt.Fprintf(&b, ", %s houses", c.Houses.System)
	} else {
		b.WriteString(", birth time unknown: no houses or rising sign")
	}
	b.WriteString("):\n")
	for _, p := range c.Positions {
		fmt.Fprintf(&b, "- %s %s", bodyName(string(p.Body)), fmtLon(p.Lon))
		if p.House > 0 {
			fmt.Fprintf(&b, ", house %d", p.House)
		}
		if p.Retro {
			b.WriteString(", retrograde")
		}
		b.WriteString("\n")
	}
	if c.TimeKnown {
		for _, p := range c.Points {
			if p.Key == "asc" || p.Key == "mc" {
				fmt.Fprintf(&b, "- %s %s\n", bodyName(p.Key), fmtLon(p.Lon))
			}
		}
	}
	b.WriteString("Aspects: ")
	parts := []string{}
	for _, a := range c.Aspects {
		parts = append(parts, fmt.Sprintf("%s %s %s (orb %.1f°)", bodyName(a.A), a.Type, bodyName(a.B), a.Orb))
	}
	b.WriteString(strings.Join(parts, "; ") + ".\n")
	fmt.Fprintf(&b, "Elements: fire %.0f, earth %.0f, air %.0f, water %.0f. Modes: cardinal %.0f, fixed %.0f, mutable %.0f.\n",
		c.Balance.Elements["fire"], c.Balance.Elements["earth"], c.Balance.Elements["air"], c.Balance.Elements["water"],
		c.Balance.Modalities["cardinal"], c.Balance.Modalities["fixed"], c.Balance.Modalities["mutable"])
	return b.String()
}

// ---- readings ----

type readingRequest struct {
	Kind     string `json:"kind"` // daily | natal | match | year | question
	PersonID int64  `json:"person_id"`
	OtherID  int64  `json:"other_id"`
	Date     string `json:"date"`
	Question string `json:"question"`
	Fresh    bool   `json:"fresh"` // write a new one even if there is one
}

// writeReading writes (or returns the kept) reading of a kind. ok=false with
// a message when no AI is set up.
func (s *Server) writeReading(ctx context.Context, uid int64, in readingRequest) (*store.Reading, string, error) {
	st, _ := s.store.GetSettings(ctx, uid)
	var p *store.Person
	var err error
	if in.PersonID > 0 {
		p, err = s.store.GetPerson(ctx, uid, in.PersonID)
	} else {
		p, err = s.store.GetMe(ctx, uid)
	}
	if err != nil {
		if in.PersonID == 0 && err == store.ErrNotFound {
			return nil, "add your own birth details first (People → You)", nil
		}
		return nil, "", err
	}
	_, _, loc, _ := s.homeOf(ctx, uid, st)
	today, _ := localDay("", loc)
	ref, title := "", ""
	var other *store.Person
	switch in.Kind {
	case "daily":
		day, err := localDay(in.Date, loc)
		if err != nil {
			return nil, err.Error(), nil
		}
		ref = day.Format("2006-01-02")
		title = "Your day, " + day.Format("Monday 2 January")
		if !p.IsMe {
			title = p.Name + "'s day, " + day.Format("Monday 2 January")
		}
	case "natal":
		title = p.Name + "'s birth chart"
	case "year":
		ref = strconv.Itoa(today.Year())
		if in.Date != "" && len(in.Date) >= 4 {
			ref = in.Date[:4]
		}
		title = "The year ahead, " + ref
	case "match":
		if other, err = s.store.GetPerson(ctx, uid, in.OtherID); err != nil {
			return nil, "", err
		}
		if other.ID == p.ID {
			return nil, "choose two different people", nil
		}
		title = p.Name + " & " + other.Name
	case "question":
		in.Question = strings.TrimSpace(in.Question)
		if in.Question == "" {
			return nil, "ask a question", nil
		}
		if len(in.Question) > 1000 {
			in.Question = in.Question[:1000]
		}
		ref, title = in.Question, in.Question
		in.Fresh = true
	default:
		return nil, "kind is daily, natal, match, year or question", nil
	}
	otherID := int64(0)
	if other != nil {
		otherID = other.ID
	}
	if !in.Fresh {
		if r, err := s.store.FindReading(ctx, uid, in.Kind, p.ID, otherID, ref); err == nil {
			return r, "", nil
		}
	}
	key := fmt.Sprintf("%d:%s:%d:%d:%s", uid, in.Kind, p.ID, otherID, ref)
	s.aiMu.Lock()
	if s.aiActive[key] {
		s.aiMu.Unlock()
		return nil, "this reading is being written already", nil
	}
	s.aiActive[key] = true
	s.aiMu.Unlock()
	defer func() {
		s.aiMu.Lock()
		delete(s.aiActive, key)
		s.aiMu.Unlock()
	}()
	cl, used, why := s.aiFor(ctx, uid, st)
	if cl == nil {
		return nil, firstNonEmpty(why, "no AI is set up (Settings → AI)"), nil
	}
	natal, err := s.natal(*p, st)
	if err != nil {
		return nil, err.Error(), nil
	}
	facts := chartFacts(p.Name, natal)
	var ask string
	switch in.Kind {
	case "daily":
		day, _ := localDay(ref, loc)
		rd, err := s.readingFor(ctx, uid, *p, st, day, false)
		if err != nil {
			return nil, err.Error(), nil
		}
		var b strings.Builder
		fmt.Fprintf(&b, "The day: %s. Moon in %s (%s, %.0f%% lit)", day.Format("Monday 2 January 2006"), signName(signIndex(rd.Moon.Sign)), phaseWords(rd.Moon.Phase), rd.Moon.Illumination*100)
		if rd.Moon.House > 0 {
			fmt.Fprintf(&b, ", in their house %d", rd.Moon.House)
		}
		b.WriteString(".\nTransits today (strongest first):\n")
		for _, t := range rd.Transits {
			fmt.Fprintf(&b, "- %s (orb %.1f°, %s)\n", t.Title, t.Orb, map[bool]string{true: "applying", false: "separating"}[t.Applying])
		}
		fmt.Fprintf(&b, "Scores 0-100: love %d, work %d, mood %d, energy %d.\n", rd.Scores["love"], rd.Scores["work"], rd.Scores["mood"], rd.Scores["energy"])
		for _, r := range rd.Retrograde {
			fmt.Fprintf(&b, "%s is retrograde.\n", r.Name)
		}
		ask = "Write their horoscope for this day: a one-line summary, then 3 short paragraphs (the mood of the day; love and people; work and energy), then 'Try:' with two concrete suggestions. 180-260 words.\n\n" + b.String() + "\n" + facts
	case "natal":
		ask = "Write a reading of this birth chart: ### Who you are (Sun, Moon, rising), ### How you love and connect, ### Work and drive, ### Your growing edge (the hardest aspect, as a strength in the making). 450-650 words. Mention the placements you base each point on.\n\n" + facts
	case "year":
		y, _ := strconv.Atoi(ref)
		from := time.Date(y, 1, 1, 0, 0, 0, 0, loc)
		hits := s.transitHits(natal, from.UTC(), from.AddDate(1, 0, 0).UTC(), []astro.Body{"jupiter", "saturn", "uranus", "neptune", "pluto"})
		var b strings.Builder
		for _, h := range hits {
			ex := []string{}
			for _, e := range h.Exact {
				ex = append(ex, e.In(loc).Format("2 Jan"))
			}
			fmt.Fprintf(&b, "- %s %s natal %s: from %s to %s, exact %s\n", bodyName(string(h.Transit)), h.Type, bodyName(h.Natal), h.Start.In(loc).Format("2 Jan 2006"), h.End.In(loc).Format("2 Jan 2006"), strings.Join(ex, ", "))
		}
		if b.Len() == 0 {
			b.WriteString("- no slow-planet transits within 1° this year\n")
		}
		ask = "Write their year ahead for " + ref + " from the slow transits below: ### The theme of the year, then one ### section per season (Winter, Spring, Summer, Autumn) naming the transits and their dates, and ### What to make of it. 400-600 words.\n\nSlow transits in " + ref + ":\n" + b.String() + "\n" + facts
	case "match":
		otherChart, err := s.natal(*other, st)
		if err != nil {
			return nil, err.Error(), nil
		}
		m, err := s.matchFor(*p, *other, st, s.viewPerson(*p, st, today), s.viewPerson(*other, st, today))
		if err != nil {
			return nil, err.Error(), nil
		}
		var b strings.Builder
		fmt.Fprintf(&b, "Scores 0-100: overall %d, love %d, communication %d, trust %d, fun %d, growth %d.\nAspects between them (%s's planet first):\n",
			m.Overall, m.Scores["love"], m.Scores["communication"], m.Scores["trust"], m.Scores["fun"], m.Scores["growth"], p.Name)
		for i, x := range m.Aspects {
			if i >= 14 {
				break
			}
			fmt.Fprintf(&b, "- %s's %s %s %s's %s (orb %.1f°)\n", p.Name, x.AName, x.Type, other.Name, x.BName, x.Orb)
		}
		rel := other.Relation
		ask = fmt.Sprintf("Write a compatibility reading for %s and %s (%s is %s's %s): ### What draws you together, ### Where you differ, ### How to make it work. 350-500 words; address %s as \"you\".\n\n%s\n%s\n%s",
			p.Name, other.Name, other.Name, p.Name, rel, p.Name, b.String(), facts, chartFacts(other.Name, otherChart))
	case "question":
		day, _ := localDay("", loc)
		rd, _ := s.readingFor(ctx, uid, *p, st, day, false)
		var b strings.Builder
		if rd != nil {
			fmt.Fprintf(&b, "Today (%s): Moon in %s, %s. Transits: ", day.Format("2 January 2006"), rd.Moon.Sign, phaseWords(rd.Moon.Phase))
			for _, t := range rd.Transits {
				b.WriteString(t.Title + "; ")
			}
			b.WriteString("\n")
		}
		ask = "Question: " + in.Question + "\n\nAnswer in a few short paragraphs from the chart and today's sky below. If astrology cannot answer it, say so kindly and answer as a sensible friend would.\n\n" + b.String() + facts
	}
	text, err := cl.Complete(ctx, aiVoice(st), ask, false)
	if err != nil {
		return nil, "", fmt.Errorf("AI: %w", err)
	}
	rd := &store.Reading{PersonID: &p.ID, Kind: in.Kind, Ref: ref, Title: title, Body: strings.TrimSpace(text), Model: used, Saved: in.Kind == "question"}
	if other != nil {
		rd.OtherID = &other.ID
	}
	if err := s.store.CreateReading(ctx, uid, rd); err != nil {
		return nil, "", err
	}
	rd.PersonName = p.Name
	if other != nil {
		rd.OtherName = other.Name
	}
	return rd, "", nil
}

func signIndex(k string) int {
	for i, s := range astro.Signs {
		if s == k {
			return i
		}
	}
	return 0
}

// handleWriteReading writes a reading with AI (POST /readings).
func (s *Server) handleWriteReading(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var in readingRequest
	if err := decode(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	rd, msg, err := s.writeReading(ctx, actorID(ctx), in)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, http.StatusNotFound, "not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	if rd == nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": msg, "not_configured": strings.Contains(msg, "AI") || strings.Contains(msg, "Pocket AI")})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "reading": rd})
}

func (s *Server) handleListReadings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	f := store.ReadingFilter{Kind: r.URL.Query().Get("kind"), PersonID: int64(queryInt(r, "person")), SavedOnly: queryBool(r, "saved"), Limit: queryInt(r, "limit")}
	if f.Limit <= 0 || f.Limit > 500 {
		f.Limit = 200
	}
	rs, err := s.store.ListReadings(ctx, actorID(ctx), f)
	if err != nil {
		s.fail(w, err, "readings")
		return
	}
	writeJSON(w, http.StatusOK, rs)
}

func (s *Server) handleGetReading(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rd, err := s.store.GetReading(ctx, actorID(ctx), idParam(r, "id"))
	if err != nil {
		s.fail(w, err, "reading")
		return
	}
	writeJSON(w, http.StatusOK, rd)
}

// handleUpdateReading keeps or un-keeps a reading ({"saved": true}).
func (s *Server) handleUpdateReading(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	var in struct {
		Saved bool `json:"saved"`
	}
	if err := decode(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := s.store.SetReadingSaved(ctx, uid, idParam(r, "id"), in.Saved); err != nil {
		s.fail(w, err, "reading")
		return
	}
	rd, _ := s.store.GetReading(ctx, uid, idParam(r, "id"))
	writeJSON(w, http.StatusOK, rd)
}

func (s *Server) handleDeleteReading(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := s.store.DeleteReading(ctx, actorID(ctx), idParam(r, "id")); err != nil {
		s.fail(w, err, "delete reading")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// handleReadingToNotes keeps a reading as a note in Pocket Notes.
func (s *Server) handleReadingToNotes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	rd, err := s.store.GetReading(ctx, uid, idParam(r, "id"))
	if err != nil {
		s.fail(w, err, "reading")
		return
	}
	c := s.clients(ctx, uid)
	sib, err := c.conn(ctx, appNotes, scopeNotesWrite)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error(), "not_connected": isNotConnected(err)})
		return
	}
	body := rd.Body + "\n\n---\n*From Pocket Horoscope, " + rd.CreatedAt[:10] + ".*"
	if s.cfg.BaseURL != "" {
		body += " [Open](" + s.cfg.BaseURL + "/readings/" + strconv.FormatInt(rd.ID, 10) + ")"
	}
	n, err := sib.client.CreateNote(ctx, rd.Title, firstNonEmpty(c.st.NotesFolder, "Horoscope"), body)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": "Pocket Notes: " + err.Error()})
		return
	}
	_ = s.store.SetReadingSaved(ctx, uid, rd.ID, true)
	link := ""
	if sib.public != "" && n.ID > 0 {
		link = strings.TrimRight(sib.public, "/") + "/notes/" + strconv.FormatInt(n.ID, 10)
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "note_id": n.ID, "url": link})
}

// handleAsk answers a question from the person's chart and today's sky.
func (s *Server) handleAsk(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var in struct {
		Question string `json:"question"`
		PersonID int64  `json:"person_id"`
	}
	if err := decode(r, &in); err != nil || strings.TrimSpace(in.Question) == "" {
		writeErr(w, http.StatusBadRequest, "question is required")
		return
	}
	rd, msg, err := s.writeReading(ctx, actorID(ctx), readingRequest{Kind: "question", PersonID: in.PersonID, Question: in.Question})
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	if rd == nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": msg, "not_configured": strings.Contains(msg, "AI") || strings.Contains(msg, "Pocket AI")})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "provider": rd.Model, "text": rd.Body, "reading": rd})
}
