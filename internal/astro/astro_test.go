package astro

import (
	"math"
	"testing"
	"time"
)

func mustTime(t *testing.T, s string) time.Time {
	t.Helper()
	v, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func within(a, b time.Time, d time.Duration) bool {
	x := a.Sub(b)
	return x <= d && x >= -d
}

// Einstein, 1879-03-14 11:30 LMT Ulm (10:50 UT), 48°24′N 10°00′E: Sun
// 23°30′ Pisces, Moon 14°32′ Sagittarius, Asc 11°38′ Cancer, MC 12°51′ Pisces
// (published charts; 1879 is outside DE421 so these are the references).
func TestEinstein(t *testing.T) {
	c := NatalChart(ChartInput{Time: time.Date(1879, 3, 14, 10, 50, 0, 0, time.UTC), TimeKnown: true, Lat: 48.4, Lon: 10.0}, nil)
	want := map[string]float64{"sun": 330 + 23.5, "moon": 240 + 14.53, "asc": 90 + 11.63, "mc": 330 + 12.85}
	for k, w := range want {
		got, ok := c.LonOf(k)
		if !ok || math.Abs(norm180(got-w)) > 3.0/60 {
			t.Errorf("%s = %s, want %s", k, FormatDeg(got), FormatDeg(w))
		}
	}
	if c.Houses == nil || c.Body(Sun).House != 10 || c.Body(Moon).House != 6 {
		t.Errorf("houses: sun %d moon %d", c.Body(Sun).House, c.Body(Moon).House)
	}
	if c.Balance.Dominant == "" || len(c.Aspects) == 0 {
		t.Errorf("balance %+v, %d aspects", c.Balance, len(c.Aspects))
	}
	// without a time: no houses, no angles, only the south node point
	nt := NatalChart(ChartInput{Time: time.Date(1879, 3, 14, 11, 0, 0, 0, time.UTC), Lat: 48.4, Lon: 10.0}, nil)
	if nt.Houses != nil || len(nt.Points) != 1 || nt.Points[0].Key != "southnode" || nt.Body(Sun).House != 0 {
		t.Errorf("time unknown: %+v %+v", nt.Houses, nt.Points)
	}
}

// Placidus cusps must sit at their thirds of the semi-arcs.
func TestPlacidusProperty(t *testing.T) {
	for _, lat := range []float64{-45, -10, 0, 33.5, 51.5, 60} {
		for ramc := 0.0; ramc < 360; ramc += 37 {
			eps := 23.44
			h := housesFrom(ramc, lat, eps, Placidus)
			for _, c := range []struct {
				i     int
				f     float64
				above bool
			}{{10, 1.0 / 3, true}, {11, 2.0 / 3, true}, {1, 2.0 / 3, false}, {2, 1.0 / 3, false}} {
				lon := h.Cusps[c.i]
				dec := math.Asin(math.Sin(eps*deg2rad)*math.Sin(lon*deg2rad)) * rad2deg
				ra := math.Atan2(math.Sin(lon*deg2rad)*math.Cos(eps*deg2rad), math.Cos(lon*deg2rad)) * rad2deg
				ad := ascDiff(dec, lat)
				want := ramc + c.f*(90+ad)
				if !c.above {
					want = ramc + 180 - c.f*(90-ad)
				}
				if d := math.Abs(norm180(ra - want)); d > 1e-6 {
					t.Errorf("lat %v ramc %v cusp %d off by %v", lat, ramc, c.i+1, d)
				}
			}
			// order: every cusp follows the previous one
			for i := 0; i < 12; i++ {
				if span := Norm360(h.Cusps[(i+1)%12] - h.Cusps[i]); span <= 0 || span >= 180 {
					t.Errorf("lat %v ramc %v: house %d spans %v", lat, ramc, i+1, span)
				}
			}
		}
	}
	// Asc is east of the MC, within 180°
	for _, s := range HouseSystems {
		h := ComputeHouses(time.Date(2026, 10, 29, 8, 0, 0, 0, time.UTC), 43.85, 18.36, s, Tropical)
		if d := Norm360(h.Asc - h.MC); d <= 0 || d >= 180 {
			t.Errorf("%s: asc-mc %v", s, d)
		}
		if HouseOf(h.Asc+0.1, h) != 1 || HouseOf(h.MC+0.1, h) != 10 && s != Whole && s != Equal {
			t.Errorf("%s: HouseOf asc/mc", s)
		}
	}
	if h := ComputeHouses(time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC), 69.6, 18.9, Placidus, Tropical); h.Fallback == "" {
		t.Error("Tromsø Placidus should fall back")
	}
}

func TestLunationsAndEclipses(t *testing.T) {
	from := mustTime(t, "2025-01-01T00:00:00Z")
	ev := Events(from, from.AddDate(2, 0, 0), Tropical, false)
	// DE421 (skyfield almanac) lunation times
	want := map[string]string{
		"2025-03-14T06:54:39Z": "full", "2025-09-21T19:54:08Z": "new", "2026-02-17T12:01:09Z": "new",
		"2026-03-03T11:37:54Z": "full", "2026-08-12T17:36:45Z": "new", "2026-10-26T04:11:49Z": "full",
		"2026-01-26T04:47:23Z": "first_quarter", "2026-11-01T20:28:27Z": "last_quarter",
	}
	for ts, ph := range want {
		at := mustTime(t, ts)
		ok := false
		for _, e := range ev {
			if e.Kind == "phase" && e.Phase == ph && within(e.Time, at, time.Minute) {
				ok = true
			}
		}
		if !ok {
			t.Errorf("missing %s at %s", ph, ts)
		}
	}
	phases := 0
	eclipses := map[string]string{}
	for _, e := range ev {
		if e.Kind == "phase" {
			phases++
		}
		if e.Kind == "eclipse" {
			eclipses[e.Time.Format("2006-01-02")] = e.Eclipse
		}
	}
	if phases != 99 {
		t.Errorf("%d lunation phases in 2025–2026, want 99", phases)
	}
	// NASA eclipse catalogue, 2025–2026
	wantEcl := map[string]string{
		"2025-03-14": "lunar_total", "2025-03-29": "solar_partial", "2025-09-07": "lunar_total", "2025-09-21": "solar_partial",
		"2026-02-17": "solar_annular", "2026-03-03": "lunar_total", "2026-08-12": "solar_total", "2026-08-28": "lunar_partial",
	}
	if len(eclipses) != len(wantEcl) {
		t.Errorf("eclipses %v", eclipses)
	}
	for d, k := range wantEcl {
		if eclipses[d] != k {
			t.Errorf("eclipse %s = %q, want %q", d, eclipses[d], k)
		}
	}
}

func TestStationsAndIngresses(t *testing.T) {
	from := mustTime(t, "2025-01-01T00:00:00Z")
	ev := Events(from, from.AddDate(2, 0, 0), Tropical, false)
	find := func(kind string, b Body, at time.Time, tol time.Duration, sign int, dir string) {
		t.Helper()
		for _, e := range ev {
			if e.Kind == kind && e.Body == b && within(e.Time, at, tol) && e.Sign == sign && (dir == "" || e.Direction == dir) {
				return
			}
		}
		t.Errorf("no %s %s %s near %s", kind, b, dir, at)
	}
	// DE421 references (stations to ±15 min: the speed is ~0 there)
	find("station", Mercury, mustTime(t, "2026-02-26T06:40:58Z"), 15*time.Minute, 11, "retrograde")
	find("station", Mercury, mustTime(t, "2026-03-20T19:25:38Z"), 15*time.Minute, 11, "direct")
	find("station", Saturn, mustTime(t, "2025-07-13T03:59:31Z"), 30*time.Minute, 0, "retrograde")
	find("ingress", Sun, mustTime(t, "2026-03-20T14:45:57Z"), time.Minute, 0, "")
	find("ingress", Saturn, mustTime(t, "2025-09-01T08:06:01Z"), 10*time.Minute, 11, "") // back into Pisces
	find("ingress", Saturn, mustTime(t, "2026-02-14T00:11:39Z"), 10*time.Minute, 0, "")
	find("ingress", Neptune, mustTime(t, "2026-01-26T17:35:18Z"), time.Hour, 0, "")
	find("ingress", Jupiter, mustTime(t, "2026-06-30T05:52:22Z"), 10*time.Minute, 4, "")

	rs := Retrogrades(mustTime(t, "2026-01-01T00:00:00Z"), mustTime(t, "2026-12-31T00:00:00Z"))
	merc := 0
	for _, r := range rs {
		if r.Body == Mercury {
			merc++
			if !r.ShadowStart.Before(r.Start) || !r.ShadowEnd.After(r.End) {
				t.Errorf("mercury shadow %+v", r)
			}
		}
	}
	if merc != 3 {
		t.Errorf("%d Mercury retrogrades in 2026, want 3", merc)
	}
}

func TestMoonIngressAndVoid(t *testing.T) {
	from := mustTime(t, "2026-01-01T00:00:00Z")
	to := from.AddDate(0, 1, 0)
	ev := Events(from, to, Tropical, true)
	n := 0
	for _, e := range ev {
		if e.Kind == "ingress" && e.Body == Moon {
			n++
			if l := PositionOf(Moon, e.Time.Add(time.Minute)).Lon; SignOf(l) != e.Sign {
				t.Errorf("moon ingress %s: in %s", e.Time, FormatDeg(l))
			}
		}
	}
	if n < 12 || n > 15 {
		t.Errorf("%d moon ingresses in a month", n)
	}
	v := VoidOfCourse(from, to)
	if len(v) < 10 {
		t.Fatalf("%d void periods", len(v))
	}
	for _, p := range v {
		if !p.Start.Before(p.End) || p.End.Sub(p.Start) > 3*24*time.Hour {
			t.Errorf("void %+v", p)
		}
		// at the end the Moon leaves the sign it was void in
		if SignOf(PositionOf(Moon, p.End.Add(-time.Minute)).Lon) != p.Sign {
			t.Errorf("void sign %+v", p)
		}
	}
}

func TestMoonPhase(t *testing.T) {
	p := MoonPhase(mustTime(t, "2026-10-29T08:00:00Z")) // full Moon was 10-26 04:12, new 10-10 15:50
	if p.Name != "waning_gibbous" || math.Abs(p.Age-18.7) > 0.1 || p.Illumination < 0.8 || p.Illumination > 0.9 {
		t.Errorf("%+v", p)
	}
	if MoonPhase(mustTime(t, "2026-08-12T17:36:45Z")).Name != "new" || MoonPhase(mustTime(t, "2026-10-26T04:11:49Z")).Name != "full" {
		t.Error("new/full names")
	}
}

func TestSidereal(t *testing.T) {
	at := mustTime(t, "2000-01-01T12:00:00Z")
	if a := Ayanamsa(at); math.Abs(a-23.857) > 0.01 { // Lahiri J2000 ≈ 23°51′25″
		t.Errorf("ayanamsa %v", a)
	}
	tr := Positions(at, Tropical)
	sd := Positions(at, Sidereal)
	for i := range tr {
		if d := norm180(tr[i].Lon - sd[i].Lon - Ayanamsa(at)); math.Abs(d) > 1e-9 {
			t.Errorf("%s sidereal off %v", tr[i].Body, d)
		}
	}
}

func TestAspects(t *testing.T) {
	ps := []Position{
		makePosition(Sun, 10, 0, 1), makePosition(Moon, 101, 0, 13), // square, 1° orb, applying? moon moves away
		makePosition(Mars, 250, 0, 0.5), makePosition(Venus, 128, 0, 1.2), // venus–sun trine 2°… 118, venus-mars 122 trine
	}
	as := FindAspects(ps, nil, nil)
	got := map[string]Aspect{}
	for _, a := range as {
		got[a.A+"-"+a.B] = a
	}
	if a, ok := got["sun-moon"]; !ok || a.Type != Square || a.Orb != 1 || a.Applying {
		t.Errorf("sun-moon %+v", a)
	}
	if a, ok := got["sun-venus"]; !ok || a.Type != Trine || a.Orb != 2 {
		t.Errorf("sun-venus %+v", a)
	}
	if a, ok := got["mars-venus"]; !ok || a.Type != Trine || !a.Applying { // venus 128 → mars 250 = 122 shrinking? venus faster closes to 120
		t.Errorf("mars-venus %+v", a)
	}
	x := CrossAspects([]Position{makePosition(Saturn, 0.5, 0, 0.1)}, nil, []Position{makePosition(Sun, 1, 0, 1)}, []Point{makePoint("asc", 90.2)}, nil)
	// sorted by orb: Saturn square Asc (0.3°, moving away from 0.2°), Saturn on the Sun (0.5°, closing)
	if len(x) != 2 || x[0].B != "asc" || x[0].Type != Square || x[0].Applying || x[1].B != "sun" || x[1].Type != Conjunction || !x[1].Applying {
		t.Errorf("cross %+v", x)
	}
}

// Saturn crossed 1° Aries three times: 2025 (direct), 2025 (retrograde), 2026 (direct).
func TestTransitThreePasses(t *testing.T) {
	natal := Chart{Zodiac: Tropical, Positions: []Position{makePosition(Sun, 1.0, 0, 1)}}
	hits := TransitHits(natal, mustTime(t, "2025-01-01T00:00:00Z"), mustTime(t, "2026-12-31T00:00:00Z"), []Body{Saturn}, 1)
	var conj *TransitHit
	for i := range hits {
		if hits[i].Natal == "sun" && hits[i].Type == Conjunction {
			conj = &hits[i]
		}
	}
	if conj == nil || len(conj.Exact) != 3 {
		t.Fatalf("saturn conjunct sun: %+v", hits)
	}
	for _, x := range conj.Exact {
		if d := math.Abs(norm180(PositionOf(Saturn, x).Lon - 1.0)); d > 1e-3 {
			t.Errorf("exact pass %s off by %v°", x, d)
		}
	}
	if !conj.Start.Before(conj.Exact[0]) || !conj.End.After(conj.Exact[2]) {
		t.Errorf("bounds %+v", conj)
	}
	// a fast transit: the Sun squares the natal Sun once in each direction a year
	hs := TransitHits(natal, mustTime(t, "2026-01-01T00:00:00Z"), mustTime(t, "2026-12-31T00:00:00Z"), []Body{Sun}, 1)
	squares := 0
	for _, h := range hs {
		if h.Type == Square && len(h.Exact) == 1 {
			squares++
		}
	}
	if squares != 2 {
		t.Errorf("%d sun squares: %+v", squares, hs)
	}
}

func TestFormatDeg(t *testing.T) {
	cases := map[float64]string{0: "0°00′ Aries", 254.5333: "14°32′ Sagittarius", 359.9999: "0°00′ Aries", 29.999: "0°00′ Taurus"}
	for in, want := range cases {
		if got := FormatDeg(in); got != want {
			t.Errorf("FormatDeg(%v) = %q, want %q", in, got, want)
		}
	}
}

func BenchmarkPositions(b *testing.B) {
	at := time.Date(2026, 10, 29, 8, 0, 0, 0, time.UTC)
	for i := 0; i < b.N; i++ {
		Positions(at, Tropical)
	}
}

func BenchmarkEventsYear(b *testing.B) {
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < b.N; i++ {
		Events(at, at.AddDate(1, 0, 0), Tropical, true)
	}
}

func BenchmarkTransits90Days(b *testing.B) {
	natal := NatalChart(ChartInput{Time: time.Date(1988, 5, 3, 7, 20, 0, 0, time.UTC), TimeKnown: true, Lat: 43.85, Lon: 18.36}, nil)
	at := time.Date(2026, 10, 29, 0, 0, 0, 0, time.UTC)
	bodies := []Body{Sun, Mercury, Venus, Mars, Jupiter, Saturn, Uranus, Neptune, Pluto}
	for i := 0; i < b.N; i++ {
		TransitHits(natal, at, at.AddDate(0, 0, 90), bodies, 1)
	}
}
