package astro

import (
	"math"
	"sort"
	"time"
)

// Phase describes the Moon's phase.
type Phase struct {
	Angle        float64 `json:"angle"`        // Moon − Sun elongation, 0–360
	Illumination float64 `json:"illumination"` // 0–1 lit fraction
	Name         string  `json:"name"`         // new, waxing_crescent, first_quarter, waxing_gibbous, full, waning_gibbous, last_quarter, waning_crescent
	Age          float64 `json:"age"`          // days since the last new Moon
}

const synodicMonth = 29.530588853

func elongation(jde float64) float64 {
	s := newSky(jde)
	sl, _ := s.sun()
	ml, _, _ := moonPos(jde, s.dpsi)
	return Norm360(ml - sl)
}

// MoonPhase returns the Moon's phase at t.
func MoonPhase(t time.Time) Phase {
	jde := jdeOf(t)
	a := elongation(jde)
	p := Phase{Angle: a, Illumination: (1 - math.Cos(a*deg2rad)) / 2, Name: phaseName(a)}
	// age: time since the elongation last crossed 0
	back := jde - a/12.19 - 1.5
	if prev := findCrossing(func(j float64) float64 { return elongation(j) }, 0, back, jde, 0.5); prev > 0 {
		p.Age = jde - prev
	} else {
		p.Age = a / 360 * synodicMonth
	}
	p.Illumination = math.Round(p.Illumination*1000) / 1000
	p.Age = math.Round(p.Age*10) / 10
	return p
}

func phaseName(a float64) string {
	switch {
	case a < 12 || a >= 348:
		return "new"
	case a < 78:
		return "waxing_crescent"
	case a < 102:
		return "first_quarter"
	case a < 168:
		return "waxing_gibbous"
	case a < 192:
		return "full"
	case a < 258:
		return "waning_gibbous"
	case a < 282:
		return "last_quarter"
	default:
		return "waning_crescent"
	}
}

// bisect finds a root of f in [a, b] where f changes sign, to about a second.
func bisect(f func(float64) float64, a, b float64) float64 { return bisectTol(f, a, b, 1e-5) }

// bisectTol finds the root to a tolerance in days with the Illinois
// variant of regula falsi: bracketing like bisection, but a handful of
// evaluations for the smooth functions used here.
func bisectTol(f func(float64) float64, a, b, tol float64) float64 {
	fa, fb := f(a), f(b)
	if fa == 0 {
		return a
	}
	if fb == 0 || (fa < 0) == (fb < 0) {
		return b
	}
	side := 0
	for i := 0; i < 60; i++ {
		c := (a*fb - b*fa) / (fb - fa)
		if b-a < tol || c <= a || c >= b {
			c = (a + b) / 2
		}
		fc := f(c)
		if fc == 0 {
			return c
		}
		if (fc < 0) == (fb < 0) {
			b, fb = c, fc
			if side == -1 {
				fa /= 2
			}
			side = -1
		} else {
			a, fa = c, fc
			if side == 1 {
				fb /= 2
			}
			side = 1
		}
		if b-a < tol {
			break
		}
		// regula falsi can creep from one side: check a tol-wide bracket around c
		if side == -1 && c-a > tol {
			if fx := f(c - tol); (fx < 0) == (fa < 0) {
				a, fa = c-tol, fx
			}
		} else if side == 1 && b-c > tol {
			if fx := f(c + tol); (fx < 0) == (fb < 0) {
				b, fb = c+tol, fx
			}
		}
	}
	return (a + b) / 2
}

// findCrossing returns the last time in [a, b] when the angle g crosses
// target going forward (0 when none), sampling every step days.
func findCrossing(g func(float64) float64, target, a, b, step float64) float64 {
	f := func(j float64) float64 { return norm180(g(j) - target) }
	last := 0.0
	prev := f(a)
	for t := a + step; t <= b+1e-9; t += step {
		cur := f(t)
		if prev < 0 && cur >= 0 && cur-prev < 90 {
			last = bisect(f, t-step, t)
		}
		prev = cur
	}
	return last
}

// SkyEvent is something happening in the sky.
type SkyEvent struct {
	Kind      string    `json:"kind"` // phase, ingress, station, eclipse
	Time      time.Time `json:"time"`
	Body      Body      `json:"body,omitempty"`
	Sign      int       `json:"sign"`
	Phase     string    `json:"phase,omitempty"`     // new, first_quarter, full, last_quarter
	Direction string    `json:"direction,omitempty"` // retrograde, direct (stations)
	Eclipse   string    `json:"eclipse,omitempty"`   // solar_total, solar_annular, solar_partial, lunar_total, lunar_partial, lunar_penumbral
	Lon       float64   `json:"lon"`
}

// Events lists lunations, eclipses, sign ingresses and stations between from
// and to, sorted by time. Moon ingresses (every ~2.5 days) only on request.
func Events(from, to time.Time, z Zodiac, includeMoonIngress bool) []SkyEvent {
	a, b := jdeOf(from), jdeOf(to)
	if b <= a {
		return nil
	}
	var out []SkyEvent
	out = append(out, lunations(a, b, z)...)
	out = append(out, planetEvents(a, b, z)...)
	if includeMoonIngress {
		out = append(out, moonIngresses(a, b, z)...)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Time.Before(out[j].Time) })
	return out
}

var phaseTargets = []struct {
	angle float64
	name  string
}{{0, "new"}, {90, "first_quarter"}, {180, "full"}, {270, "last_quarter"}}

func lunations(a, b float64, z Zodiac) []SkyEvent {
	var out []SkyEvent
	const step = 0.5
	prev := elongation(a)
	for t := a; t < b; t += step {
		next := elongation(math.Min(t+step, b))
		end := math.Min(t+step, b)
		for _, pt := range phaseTargets {
			f0 := norm180(prev - pt.angle)
			f1 := norm180(next - pt.angle)
			if f0 < 0 && f1 >= 0 && f1-f0 < 90 {
				target := pt.angle
				j := bisect(func(j float64) float64 { return norm180(elongation(j) - target) }, t, end)
				moon := bodyLon(Moon, j)
				lon := withZodiac(moon, z, j)
				ev := SkyEvent{Kind: "phase", Time: timeOfJDE(j), Body: Moon, Sign: SignOf(lon), Phase: pt.name, Lon: lon}
				out = append(out, ev)
				if pt.name == "new" || pt.name == "full" {
					if e := eclipseAt(j, pt.name == "new"); e != "" {
						ec := ev
						ec.Kind, ec.Eclipse, ec.Phase = "eclipse", e, pt.name
						if pt.name == "new" {
							ec.Body = Sun
						}
						out = append(out, ec)
					}
				}
			}
		}
		prev = next
	}
	return out
}

// eclipseAt classifies an eclipse at a new (solar) or full (lunar) Moon, or
// returns "" when the Moon is too far from the node. Geometry is geocentric
// and approximate: good enough to name an eclipse and its kind.
func eclipseAt(jde float64, solar bool) string {
	s := newSky(jde)
	_, beta, dist := moonPos(jde, s.dpsi)
	r := s.earth.len()
	const earthR = 6378.14
	moonSD := 358473400 / dist / 3600           // degrees
	sunSD := 959.63 / r / 3600                  // degrees
	moonPar := math.Asin(earthR/dist) * rad2deg // degrees
	sunPar := 8.794 / r / 3600
	// closest approach of the Moon to the Sun/shadow axis is β·cos i
	sep := math.Abs(beta) * math.Cos(5.145*deg2rad)
	if solar {
		// gamma: the shadow axis' distance from Earth's centre in Earth radii
		gamma := math.Sin(sep*deg2rad) * dist / earthR
		switch {
		case gamma > 1.55:
			return ""
		case gamma < 0.997:
			if moonSD*1.005 >= sunSD { // a little topocentric enlargement
				return "solar_total"
			}
			return "solar_annular"
		default:
			if sep > moonPar+sunSD+moonSD {
				return ""
			}
			return "solar_partial"
		}
	}
	umbra := 1.02 * (moonPar + sunPar - sunSD)
	penumbra := 1.02 * (moonPar + sunPar + sunSD)
	switch {
	case sep+moonSD <= umbra:
		return "lunar_total"
	case sep-moonSD < umbra:
		return "lunar_partial"
	case sep-moonSD < penumbra:
		return "lunar_penumbral"
	}
	return ""
}

var stationBodies = []Body{Mercury, Venus, Mars, Jupiter, Saturn, Uranus, Neptune, Pluto}

// planetEvents finds sign ingresses (Sun to Pluto and the node) and stations
// (Mercury to Pluto) by sampling once a day.
func planetEvents(a, b float64, z Zodiac) []SkyEvent {
	bodies := []Body{Sun, Mercury, Venus, Mars, Jupiter, Saturn, Uranus, Neptune, Pluto, Node}
	const step = 1.0
	n := int(math.Ceil((b-a)/step)) + 3
	// samples at a-step … b+step so stations at the edges can be seen
	lons := make([][]float64, len(bodies))
	for i := range lons {
		lons[i] = make([]float64, n)
	}
	for k := 0; k < n; k++ {
		j := a + float64(k-1)*step
		s := newSkyCut(j, coarseCutoff)
		for i, bd := range bodies {
			l, _ := s.at(bd)
			lons[i][k] = withZodiac(l, z, j)
		}
	}
	var out []SkyEvent
	for i, bd := range bodies {
		bd := bd
		lonZ := func(j float64) float64 { return withZodiac(bodyLon(bd, j), z, j) }
		for k := 1; k < n-1; k++ {
			t0 := a + float64(k-1)*step
			// ingress between samples k and k+1
			s0, s1 := SignOf(lons[i][k]), SignOf(lons[i][k+1])
			if s0 != s1 {
				d := norm180(lons[i][k+1] - lons[i][k])
				boundary := float64(s1 * 30) // moving forward into s1
				if d < 0 {
					boundary = float64(s0 * 30) // moving back out of s0
				}
				f := func(j float64) float64 { return norm180(lonZ(j) - boundary) }
				j := bisect(f, t0, t0+step)
				if j >= a && j <= b {
					out = append(out, SkyEvent{Kind: "ingress", Time: timeOfJDE(j), Body: bd, Sign: s1, Lon: lonZ(j + 1e-4)})
				}
			}
			if bd == Sun || bd == Node {
				continue
			}
			// station: daily motion changes sign around sample k
			d0 := norm180(lons[i][k] - lons[i][k-1])
			d1 := norm180(lons[i][k+1] - lons[i][k])
			if (d0 < 0) != (d1 < 0) {
				sp := func(j float64) float64 { return bodySpeed(bd, j) }
				lo, hi := t0-step, t0+step
				if (sp(lo) < 0) == (sp(hi) < 0) {
					continue
				}
				j := bisect(sp, lo, hi)
				if j < a || j > b {
					continue
				}
				dir := "direct"
				if d0 > 0 {
					dir = "retrograde"
				}
				l := lonZ(j)
				out = append(out, SkyEvent{Kind: "station", Time: timeOfJDE(j), Body: bd, Sign: SignOf(l), Direction: dir, Lon: l})
			}
		}
	}
	return dedupeEvents(out)
}

// dedupeEvents drops a station found twice from neighbouring samples.
func dedupeEvents(ev []SkyEvent) []SkyEvent {
	sort.SliceStable(ev, func(i, j int) bool { return ev[i].Time.Before(ev[j].Time) })
	out := ev[:0]
	for _, e := range ev {
		dup := false
		for k := len(out) - 1; k >= 0 && e.Time.Sub(out[k].Time) < 2*24*time.Hour; k-- {
			o := out[k]
			if o.Kind == e.Kind && o.Body == e.Body && o.Direction == e.Direction && o.Sign == e.Sign && e.Time.Sub(o.Time) < time.Hour {
				dup = true
				break
			}
		}
		if !dup {
			out = append(out, e)
		}
	}
	return out
}

func moonLonZ(j float64, z Zodiac) float64 {
	dpsi, _ := nutation(j)
	l, _, _ := moonPos(j, dpsi)
	return withZodiac(l, z, j)
}

func moonIngresses(a, b float64, z Zodiac) []SkyEvent {
	var out []SkyEvent
	const step = 0.25
	prev := moonLonZ(a, z)
	for t := a; t < b; t += step {
		end := math.Min(t+step, b)
		next := moonLonZ(end, z)
		if s1 := SignOf(next); s1 != SignOf(prev) {
			boundary := float64(s1 * 30)
			j := bisect(func(j float64) float64 { return norm180(moonLonZ(j, z) - boundary) }, t, end)
			out = append(out, SkyEvent{Kind: "ingress", Time: timeOfJDE(j), Body: Moon, Sign: s1, Lon: boundary})
		}
		prev = next
	}
	return out
}

// ---- void-of-course Moon ----

// VoidPeriod is a stretch when the Moon makes no more major aspects before
// leaving its sign.
type VoidPeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Sign  int       `json:"sign"` // the sign the Moon is leaving
}

var vocBodies = []Body{Sun, Mercury, Venus, Mars, Jupiter, Saturn, Uranus, Neptune, Pluto}

// VoidOfCourse lists void-of-course Moon periods overlapping [from, to]
// (tropical signs, Ptolemaic aspects to the Sun through Pluto).
func VoidOfCourse(from, to time.Time) []VoidPeriod {
	a, b := jdeOf(from), jdeOf(to)
	ings := moonIngresses(a-3, b+3, Tropical)
	var out []VoidPeriod
	for i := 1; i < len(ings); i++ {
		in, outT := jdeOf(ings[i-1].Time), jdeOf(ings[i].Time)
		if outT < a || in > b {
			continue
		}
		start := lastLunarAspect(in, outT)
		if start == 0 {
			start = in
		}
		if outT-start < 1.0/1440 {
			continue
		}
		out = append(out, VoidPeriod{Start: timeOfJDE(start), End: ings[i].Time, Sign: ings[i-1].Sign})
	}
	return out
}

// lastLunarAspect returns the time of the Moon's last exact Ptolemaic aspect
// to a planet between in and out (0 when none). Planets are interpolated
// quadratically across the stretch; the Moon is computed exactly.
func lastLunarAspect(in, out float64) float64 {
	mid := (in + out) / 2
	s0, s1, s2 := newSky(in), newSky(mid), newSky(out)
	last := 0.0
	for _, bd := range vocBodies {
		l0, _ := s0.at(bd)
		l1, _ := s1.at(bd)
		l2, _ := s2.at(bd)
		d1, d2 := norm180(l1-l0), norm180(l2-l0)
		h := (out - in) / 2
		planet := func(j float64) float64 { // Lagrange quadratic through the three samples
			x := (j - in) / h
			return l0 + d1*x*(2-x) + d2*x*(x-1)/2
		}
		for _, asp := range Ptolemaic {
			for _, sgn := range []float64{1, -1} {
				ang := AspectAngles[asp] * sgn
				if sgn < 0 && (ang == 0 || ang == -180) {
					continue
				}
				f := func(j float64) float64 {
					dpsi, _ := nutation(j)
					m, _, _ := moonPos(j, dpsi)
					return norm180(m - planet(j) - ang)
				}
				const step = 1.0 / 12
				prev := f(in)
				for t := in + step; t <= out+1e-9; t += step {
					tt := math.Min(t, out)
					cur := f(tt)
					if prev < 0 && cur >= 0 && cur-prev < 90 {
						if j := bisect(f, tt-step, tt); j > last {
							last = j
						}
					}
					prev = cur
				}
			}
		}
	}
	return last
}

// ---- retrogrades ----

// RetroPeriod is one retrograde loop of a planet.
type RetroPeriod struct {
	Body        Body      `json:"body"`
	Start       time.Time `json:"start"` // station retrograde
	End         time.Time `json:"end"`   // station direct
	StartLon    float64   `json:"start_lon"`
	EndLon      float64   `json:"end_lon"`
	ShadowStart time.Time `json:"shadow_start"` // first reaches the degree where it will turn direct
	ShadowEnd   time.Time `json:"shadow_end"`   // back at the degree where it turned retrograde
}

// Retrogrades lists retrograde periods (Mercury to Pluto) overlapping
// [from, to], tropical longitudes.
func Retrogrades(from, to time.Time) []RetroPeriod {
	pad := 200 * 24 * time.Hour
	ev := planetEvents(jdeOf(from.Add(-pad)), jdeOf(to.Add(pad)), Tropical)
	var out []RetroPeriod
	for _, bd := range stationBodies {
		var open *SkyEvent
		for i := range ev {
			e := ev[i]
			if e.Kind != "station" || e.Body != bd {
				continue
			}
			if e.Direction == "retrograde" {
				open = &ev[i]
				continue
			}
			if open == nil {
				continue
			}
			p := RetroPeriod{Body: bd, Start: open.Time, End: e.Time, StartLon: open.Lon, EndLon: e.Lon}
			open = nil
			if p.End.Before(from) || p.Start.After(to) {
				continue
			}
			bd := bd
			lon := func(j float64) float64 { return bodyLon(bd, j) }
			s, en := jdeOf(p.Start), jdeOf(p.End)
			if j := findCrossing(lon, p.EndLon, s-200, s, 1); j > 0 {
				p.ShadowStart = timeOfJDE(j)
			}
			if j := firstCrossing(lon, p.StartLon, en, en+200, 1); j > 0 {
				p.ShadowEnd = timeOfJDE(j)
			}
			out = append(out, p)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Start.Before(out[j].Start) })
	return out
}

// firstCrossing is findCrossing's twin returning the first forward crossing.
func firstCrossing(g func(float64) float64, target, a, b, step float64) float64 {
	f := func(j float64) float64 { return norm180(g(j) - target) }
	prev := f(a)
	for t := a + step; t <= b+1e-9; t += step {
		cur := f(t)
		if prev < 0 && cur >= 0 && cur-prev < 90 {
			return bisect(f, t-step, t)
		}
		prev = cur
	}
	return 0
}
