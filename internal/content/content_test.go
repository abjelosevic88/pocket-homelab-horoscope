package content

import (
	"reflect"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

var (
	bodyKeys   = []string{"sun", "moon", "mercury", "venus", "mars", "jupiter", "saturn", "uranus", "neptune", "pluto", "node"}
	pointKeys  = []string{"asc", "mc", "dsc", "ic", "fortune", "southnode"}
	signKeys   = []string{"aries", "taurus", "gemini", "cancer", "leo", "virgo", "libra", "scorpio", "sagittarius", "capricorn", "aquarius", "pisces"}
	aspectKeys = []string{"conjunction", "sextile", "square", "trine", "opposition", "quincunx"}
	phaseKeys  = []string{"new", "waxing_crescent", "first_quarter", "waxing_gibbous", "full", "waning_gibbous", "last_quarter", "waning_crescent"}
	transiting = []string{"sun", "moon", "mercury", "venus", "mars", "jupiter", "saturn", "uranus", "neptune", "pluto"}
)

// grammar is a sanity check, not a grammar checker: capitalised, single
// spaces, ends with a full stop, no placeholders, no doubled words.
func grammar(t *testing.T, where, s string) {
	t.Helper()
	if s == "" {
		t.Errorf("%s: empty", where)
		return
	}
	r, _ := utf8.DecodeRuneInString(s)
	if !unicode.IsUpper(r) {
		t.Errorf("%s: not capitalised: %q", where, s)
	}
	if !strings.HasSuffix(s, ".") {
		t.Errorf("%s: does not end with a full stop: %q", where, s)
	}
	for _, bad := range []string{"  ", "%s", "%!", "{", "}", " .", " ,", "..", "<nil>", "TODO", "the the ", " a a ", " a overhaul", "your your", "  "} {
		if strings.Contains(s, bad) {
			t.Errorf("%s: contains %q: %q", where, bad, s)
		}
	}
	words := strings.Fields(strings.ToLower(s))
	for i := 1; i < len(words); i++ {
		if words[i] == words[i-1] && len(words[i]) > 2 {
			t.Errorf("%s: doubled word %q: %q", where, words[i], s)
		}
	}
	for _, sentence := range strings.Split(s, ". ") {
		r, _ := utf8.DecodeRuneInString(sentence)
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			t.Errorf("%s: sentence not capitalised: %q", where, sentence)
		}
	}
}

func TestSigns(t *testing.T) {
	if len(Signs()) != 12 {
		t.Fatal("want 12 signs")
	}
	for i, k := range signKeys {
		s, ok := SignByKey(k)
		if !ok || s.Key != k || SignByIndex(i).Key != k {
			t.Fatalf("sign %s", k)
		}
		for _, f := range []string{s.Summary, s.Love, s.Work, s.Friendship, s.Growth} {
			grammar(t, k, f)
		}
		if s.Name == "" || s.Glyph == "" || s.Element == "" || s.Modality == "" || s.Ruler == "" || s.Dates == "" ||
			len(s.Keywords) < 2 || len(s.Strengths) == 0 || len(s.Challenges) == 0 {
			t.Errorf("sign %s incomplete", k)
		}
		if _, ok := PlanetByKey(s.Ruler); !ok {
			t.Errorf("sign %s ruler %q unknown", k, s.Ruler)
		}
	}
	if SignByIndex(-1).Key != "pisces" || SignByIndex(12).Key != "aries" {
		t.Error("SignByIndex wraps")
	}
	if _, ok := SignByKey("ophiuchus"); ok {
		t.Error("unknown sign found")
	}
}

func TestPlanetsHousesAspects(t *testing.T) {
	if len(Planets()) != 11 || len(Points()) != 6 {
		t.Fatal("planet/point counts")
	}
	for _, k := range append(append([]string{}, bodyKeys...), pointKeys...) {
		p, ok := PlanetByKey(k)
		if !ok {
			t.Fatalf("planet %s missing", k)
		}
		grammar(t, k, p.Meaning)
		if p.Short == "" || p.Cycle == "" || p.Glyph == "" || len(p.Keywords) == 0 {
			t.Errorf("planet %s incomplete", k)
		}
	}
	for _, k := range pointKeys {
		if _, ok := PointByKey(k); !ok {
			t.Errorf("point %s", k)
		}
	}
	if _, ok := PointByKey("sun"); ok {
		t.Error("sun is not a point")
	}
	if len(Houses()) != 12 {
		t.Fatal("houses")
	}
	for i, h := range Houses() {
		if h.N != i+1 || h.Name == "" || len(h.Keywords) == 0 {
			t.Errorf("house %d", i+1)
		}
		grammar(t, h.Name, h.Summary)
	}
	if len(Aspects()) != 6 {
		t.Fatal("aspects")
	}
	for _, k := range aspectKeys {
		a, ok := AspectByKey(k)
		if !ok || a.Nature == "" || a.Short == "" {
			t.Fatalf("aspect %s", k)
		}
		grammar(t, k, a.Meaning)
	}
}

func TestPlacements(t *testing.T) {
	for _, p := range append(append([]string{}, bodyKeys...), pointKeys...) {
		for _, s := range signKeys {
			grammar(t, p+" in "+s, PlanetInSign(p, s))
		}
	}
	for _, p := range bodyKeys {
		for h := 1; h <= 12; h++ {
			grammar(t, p+" in house", PlanetInHouse(p, h))
		}
	}
	// the detailed ones are longer than the generational ones
	for _, p := range []string{"sun", "moon", "asc", "mercury", "venus", "mars"} {
		for _, s := range signKeys {
			if n := strings.Count(PlanetInSign(p, s), ". "); n < 2 {
				t.Errorf("%s in %s: want 3–4 sentences", p, s)
			}
		}
	}
	// every text is unique
	seen := map[string]string{}
	for p, texts := range inSign {
		for i, s := range texts {
			if prev, ok := seen[s]; ok {
				t.Errorf("duplicate text %s/%d and %s", p, i, prev)
			}
			seen[s] = p
		}
	}
	if PlanetInSign("sun", "ophiuchus") != "" || PlanetInSign("vulcan", "aries") != "" ||
		PlanetInHouse("sun", 0) != "" || PlanetInHouse("sun", 13) != "" || PlanetInHouse("asc", 1) != "" {
		t.Error("unknown keys should return empty")
	}
}

func TestComposedTexts(t *testing.T) {
	all := append(append([]string{}, bodyKeys...), pointKeys...)
	for _, a := range all {
		for _, b := range all {
			for _, asp := range aspectKeys {
				if a != b {
					grammar(t, "natal "+a+" "+asp+" "+b, NatalAspect(a, b, asp))
				}
				s := Synastry(a, b, asp)
				grammar(t, "synastry "+a+" "+asp+" "+b, s.Text)
				if s.Title == "" || len(s.Areas) != 1 || s.Weight < 1 || s.Weight > 5 || s.Tone < -1 || s.Tone > 1 {
					t.Errorf("synastry %s %s %s: %+v", a, asp, b, s)
				}
			}
		}
	}
	for _, tr := range transiting {
		for _, n := range all {
			for _, asp := range aspectKeys {
				x := Transit(tr, asp, n)
				grammar(t, "transit "+tr+" "+asp+" "+n, x.Text)
				if x.Title == "" || len(x.Areas) == 0 || x.Weight < 1 || x.Weight > 5 || x.Tone < -1 || x.Tone > 1 {
					t.Errorf("transit %s %s %s: %+v", tr, asp, n, x)
				}
			}
		}
	}
	if got := Transit("venus", "trine", "sun").Title; got != "Venus trine your Sun" {
		t.Errorf("title %q", got)
	}
	if Transit("venus", "trine", "sun").Tone != 1 || Transit("saturn", "square", "moon").Tone != -1 {
		t.Error("tone")
	}
	if Transit("pluto", "square", "sun").Weight <= Transit("moon", "square", "sun").Weight {
		t.Error("slow transits should weigh more")
	}
	if Transit("node", "trine", "sun").Text != "" || Transit("sun", "biquintile", "moon").Text != "" || NatalAspect("sun", "sun", "trine") != "" {
		t.Error("unknown keys should return empty")
	}
}

func TestDaily(t *testing.T) {
	for _, s := range signKeys {
		grammar(t, "moon in "+s, MoonInSign(s))
		for _, s2 := range signKeys {
			grammar(t, "compat "+s+" "+s2, Compatibility(s, s2))
		}
		for _, p := range append(append([]string{}, transiting...), "node") {
			grammar(t, "ingress "+p+" "+s, Ingress(p, s))
		}
		for _, k := range []string{"solar", "lunar", "solar_total", "lunar_partial", "solar_annular"} {
			grammar(t, "eclipse "+k, Eclipse(k, s))
		}
	}
	for _, ph := range phaseKeys {
		grammar(t, "phase "+ph, MoonPhaseText(ph))
	}
	for _, p := range transiting[2:] {
		grammar(t, "retro "+p, Retrograde(p))
		grammar(t, "station r "+p, StationText(p, "retrograde"))
		grammar(t, "station d "+p, StationText(p, "direct"))
	}
	grammar(t, "voc", VoidOfCourse())
	if Retrograde("sun") != "" || StationText("moon", "direct") != "" || StationText("mars", "sideways") != "" ||
		MoonInSign("x") != "" || MoonPhaseText("x") != "" || Eclipse("hybrid", "aries") != "" || Compatibility("x", "leo") != "" {
		t.Error("unknown keys should return empty")
	}
	if Compatibility("aries", "leo") == Compatibility("aries", "taurus") {
		t.Error("compatibility should differ by element")
	}
}

func TestTipsAndHeadline(t *testing.T) {
	scores := map[string]int{"love": 80, "work": 30, "mood": 55, "energy": 60}
	for _, s := range signKeys {
		for _, ph := range phaseKeys {
			a := Tips(scores, s, ph)
			b := Tips(map[string]int{"energy": 60, "mood": 55, "work": 30, "love": 80}, s, ph)
			if !reflect.DeepEqual(a, b) {
				t.Fatalf("tips not deterministic for %s %s", s, ph)
			}
			if len(a.Do) < 2 || len(a.Do) > 3 || len(a.Avoid) < 2 || len(a.Avoid) > 3 {
				t.Errorf("tips %s %s: %+v", s, ph, a)
			}
			for _, x := range append(append([]string{}, a.Do...), a.Avoid...) {
				grammar(t, "tip", x)
			}
		}
	}
	if len(Tips(nil, "", "").Do) != 0 {
		t.Error("no input, no tips")
	}
	for tone := -2; tone <= 2; tone++ {
		for _, area := range []string{"love", "work", "mood", "energy", "unknown"} {
			for seed := int64(-3); seed < 5; seed++ {
				h := Headline(tone, area, seed)
				if h == "" || h != Headline(tone, area, seed) {
					t.Fatalf("headline %d %s %d", tone, area, seed)
				}
			}
		}
	}
}

func TestSamples(t *testing.T) {
	if testing.Verbose() {
		t.Log(NatalAspect("sun", "moon", "square"))
		t.Log(Transit("mars", "square", "moon").Text)
		t.Log(Transit("jupiter", "conjunction", "mc").Text)
		t.Log(Transit("moon", "trine", "venus").Text)
		t.Log(Synastry("venus", "mars", "trine").Text)
		t.Log(Synastry("asc", "saturn", "quincunx").Text)
		t.Log(Ingress("venus", "libra"))
		t.Log(Compatibility("aries", "cancer"))
		t.Log(Tips(map[string]int{"love": 80, "work": 30, "mood": 55, "energy": 60}, "scorpio", "full"))
		t.Log(PlanetInSign("fortune", "virgo"))
	}
}
