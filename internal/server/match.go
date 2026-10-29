package server

import (
	"math"
	"net/http"
	"sort"

	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/astro"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/content"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/store"
)

// ---- compatibility (synastry) ----
//
// Two birth charts laid over each other: every aspect between a planet of
// one and a point of the other counts, by tone and weight, into five areas.
// The Sun signs add a short note on how their elements mix.

var matchAreas = []string{"love", "communication", "trust", "fun", "growth"}

type matchAspect struct {
	astro.Aspect
	AName  string   `json:"a_name"`
	BName  string   `json:"b_name"`
	Title  string   `json:"title"`
	Text   string   `json:"text"`
	Tone   int      `json:"tone"`
	Areas  []string `json:"areas"`
	Weight int      `json:"weight"`
}

type matchView struct {
	A        personView            `json:"a"`
	B        personView            `json:"b"`
	Overall  int                   `json:"overall"`
	Scores   map[string]int        `json:"scores"`
	Summary  string                `json:"summary"`
	SunPair  string                `json:"sun_pair"`
	Elements map[string][2]float64 `json:"elements"`
	Aspects  []matchAspect         `json:"aspects"`
	Strong   []matchAspect         `json:"strengths"`
	Hard     []matchAspect         `json:"challenges"`
	Reading  *store.Reading        `json:"reading"`
}

var matchWords = []struct {
	min  int
	text string
}{
	{78, "A rare, easy fit: you understand each other without much effort, and the hard spots are few."},
	{66, "A warm match with real pull. Most things flow; the rough edges are the kind you can talk through."},
	{54, "A good mix of ease and friction. You will learn from each other, sometimes the hard way."},
	{42, "You are wired differently. It can work well with patience and clear words."},
	{0, "Opposite rhythms: this match asks for effort and humour, and gives back a lot of growth."},
}

func (s *Server) matchFor(a, b store.Person, st store.Settings, aV, bV personView) (matchView, error) {
	ca, err := s.natal(a, st)
	if err != nil {
		return matchView{}, err
	}
	cb, err := s.natal(b, st)
	if err != nil {
		return matchView{}, err
	}
	orbs := astro.Orbs{"conjunction": 7, "opposition": 6, "square": 5, "trine": 6, "sextile": 4}
	aPts, bPts := []astro.Point{}, []astro.Point{}
	for _, p := range ca.Points {
		if (p.Key == "asc" || p.Key == "mc") && ca.TimeKnown {
			aPts = append(aPts, p)
		}
	}
	for _, p := range cb.Points {
		if (p.Key == "asc" || p.Key == "mc") && cb.TimeKnown {
			bPts = append(bPts, p)
		}
	}
	// both directions: A's planets to B's chart, then B's to A's points only
	// for the angles (planet–planet pairs would count twice)
	asp := astro.CrossAspects(ca.Positions, aPts, cb.Positions, bPts, orbs)
	m := matchView{A: aV, B: bV, Scores: map[string]int{}, Aspects: []matchAspect{}, Strong: []matchAspect{}, Hard: []matchAspect{}, Elements: map[string][2]float64{}}
	sum := map[string]float64{}
	for _, x := range asp {
		if x.A == "node" && x.B == "node" {
			continue
		}
		tt := content.Synastry(x.A, x.B, string(x.Type))
		if tt.Title == "" {
			continue
		}
		max := orbs[x.Type]
		close := 1.0
		if max > 0 {
			close = 1 - x.Orb/max
		}
		ma := matchAspect{Aspect: x, AName: bodyName(x.A), BName: bodyName(x.B), Title: tt.Title, Text: tt.Text, Tone: tt.Tone, Areas: tt.Areas, Weight: tt.Weight}
		m.Aspects = append(m.Aspects, ma)
		v := float64(tt.Weight) * (0.4 + 0.6*close)
		for _, ar := range tt.Areas {
			switch tt.Tone {
			case 1:
				sum[ar] += v
			case -1:
				sum[ar] -= 0.8 * v
			default:
				sum[ar] += 0.35 * v // a conjunction binds, for better and worse
			}
		}
	}
	sort.SliceStable(m.Aspects, func(i, j int) bool {
		wi := float64(m.Aspects[i].Weight) - m.Aspects[i].Orb/4
		wj := float64(m.Aspects[j].Weight) - m.Aspects[j].Orb/4
		return wi > wj
	})
	for _, x := range m.Aspects {
		if x.Tone > 0 && len(m.Strong) < 5 {
			m.Strong = append(m.Strong, x)
		}
		if x.Tone < 0 && len(m.Hard) < 5 {
			m.Hard = append(m.Hard, x)
		}
	}
	total := 0
	for _, ar := range matchAreas {
		v := 50 + 46*math.Tanh(sum[ar]/9)
		n := int(math.Round(math.Max(10, math.Min(97, v))))
		m.Scores[ar] = n
		total += n
	}
	m.Overall = int(math.Round(float64(total) / float64(len(matchAreas))))
	for _, w := range matchWords {
		if m.Overall >= w.min {
			m.Summary = w.text
			break
		}
	}
	m.SunPair = content.Compatibility(aV.Sun, bV.Sun)
	for _, el := range []string{"fire", "earth", "air", "water"} {
		m.Elements[el] = [2]float64{round2(ca.Balance.Elements[el]), round2(cb.Balance.Elements[el])}
	}
	return m, nil
}

// handleMatch compares two people (?a=&b=; a defaults to you).
func (s *Server) handleMatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := actorID(ctx)
	st, _ := s.store.GetSettings(ctx, uid)
	var a *store.Person
	var err error
	if id := queryInt(r, "a"); id > 0 {
		a, err = s.store.GetPerson(ctx, uid, int64(id))
	} else {
		a, err = s.store.GetMe(ctx, uid)
	}
	if err != nil {
		s.fail(w, err, "person")
		return
	}
	b, err := s.store.GetPerson(ctx, uid, int64(queryInt(r, "b")))
	if err != nil {
		s.fail(w, err, "person")
		return
	}
	if a.ID == b.ID {
		writeErr(w, http.StatusBadRequest, "choose two different people")
		return
	}
	_, _, loc, _ := s.homeOf(ctx, uid, st)
	today, _ := localDay("", loc)
	m, err := s.matchFor(*a, *b, st, s.viewPerson(*a, st, today), s.viewPerson(*b, st, today))
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if rd, err := s.store.FindReading(ctx, uid, "match", a.ID, b.ID, ""); err == nil {
		m.Reading = rd
	}
	writeJSON(w, http.StatusOK, m)
}
