package astro

import (
	"math"
	"sort"
	"time"
)

// TransitHit is one transit of a moving body to a natal body or point:
// when it comes within orb, the exact passes (up to three with retrogrades)
// and when it leaves.
type TransitHit struct {
	Transit Body        `json:"transit"`
	Natal   string      `json:"natal"` // body or point key
	Type    AspectType  `json:"type"`
	Start   time.Time   `json:"start"`
	Exact   []time.Time `json:"exact"`
	End     time.Time   `json:"end"`
}

// transitTargets are the natal objects transits are measured to.
var transitTargets = map[string]bool{"sun": true, "moon": true, "mercury": true, "venus": true, "mars": true,
	"jupiter": true, "saturn": true, "uranus": true, "neptune": true, "pluto": true, "node": true, "asc": true, "mc": true}

// mergeGap is how long a transit may drop out of orb and still count as the
// same transit (a retrograde loop).
func mergeGap(b Body) float64 {
	switch b {
	case Moon, Sun:
		return 0
	case Mercury, Venus:
		return 60
	case Mars:
		return 90
	default:
		return 260
	}
}

func samplingStep(b Body) float64 {
	switch b {
	case Moon:
		return 0.1
	case Sun, Mercury, Venus, Mars:
		return 0.5
	case Jupiter, Saturn:
		return 1
	default:
		return 2
	}
}

// lookaround is how far outside the window to look so that Start and End
// of transits overlapping the window are real.
func lookaround(b Body) float64 {
	switch b {
	case Moon:
		return 1
	case Sun:
		return 3
	case Mercury, Venus:
		return 80
	case Mars:
		return 120
	default:
		return 400
	}
}

// TransitHits finds the Ptolemaic aspects the given bodies make to the natal
// chart's bodies and angles between from and to (orb in degrees, ~1).
// Results are sorted by their first exact pass (or start).
func TransitHits(natal Chart, from, to time.Time, bodies []Body, orb float64) []TransitHit {
	if orb <= 0 {
		orb = 1
	}
	type target struct {
		key string
		lon float64
	}
	var targets []target
	for _, p := range natal.Positions {
		if transitTargets[string(p.Body)] {
			targets = append(targets, target{string(p.Body), p.Lon})
		}
	}
	for _, p := range natal.Points {
		if transitTargets[p.Key] {
			targets = append(targets, target{p.Key, p.Lon})
		}
	}
	a, b := jdeOf(from), jdeOf(to)
	var out []TransitHit
	for _, bd := range bodies {
		bd := bd
		step := samplingStep(bd)
		pad := lookaround(bd)
		lonAt := func(j float64) float64 { return withZodiac(bodyLon(bd, j), natal.Zodiac, j) }
		coarse := func(j float64) float64 {
			if bd == Moon || bd == Node {
				return lonAt(j)
			}
			l, _ := newSkyCut(j, coarseCutoff).at(bd)
			return withZodiac(l, natal.Zodiac, j)
		}
		start, end := a-pad, b+pad
		n := int(math.Ceil((end-start)/step)) + 1
		lons := make([]float64, n)
		for k := range lons {
			lons[k] = coarse(start + float64(k)*step)
		}
		for _, tg := range targets {
			for _, asp := range Ptolemaic {
				angs := []float64{AspectAngles[asp]}
				if asp != Conjunction && asp != Opposition {
					angs = append(angs, -AspectAngles[asp])
				}
				// scan the coarse samples first (indices only), refine later
				// just the transits that touch the window
				type run struct {
					ang    float64
					sk, ek int   // first sample in orb, first sample out (-1: still in)
					xk     []int // samples just after an exact pass
				}
				var runs []run
				for _, ang := range angs {
					off := func(l float64) float64 { return norm180(l - tg.lon - ang) }
					in := false
					var cur run
					for k := 0; k < n; k++ {
						d := off(lons[k])
						within := math.Abs(d) <= orb
						if within && !in {
							in = true
							cur = run{ang: ang, sk: k, ek: -1}
						}
						if k > 0 && in {
							p := off(lons[k-1])
							if (p < 0) != (d < 0) && math.Abs(d-p) < 90 {
								cur.xk = append(cur.xk, k)
							}
						}
						if !within && in {
							in = false
							cur.ek = k
							runs = append(runs, cur)
						}
					}
					if in {
						runs = append(runs, cur)
					}
				}
				at := func(k int) float64 { return start + float64(k)*step }
				endOf := func(r run) float64 {
					if r.ek < 0 {
						return end
					}
					return at(r.ek)
				}
				sort.Slice(runs, func(i, j int) bool { return runs[i].sk < runs[j].sk })
				// merge the passes of one retrograde loop
				var groups [][]run
				for _, r := range runs {
					if g := len(groups); g > 0 {
						last := groups[g-1][len(groups[g-1])-1]
						if at(r.sk)-endOf(last) <= mergeGap(bd) {
							groups[g-1] = append(groups[g-1], r)
							continue
						}
					}
					groups = append(groups, []run{r})
				}
				for _, g := range groups {
					first, last := g[0], g[len(g)-1]
					if endOf(last) < a || at(first.sk)-step > b {
						continue
					}
					fo := func(ang float64) func(float64) float64 {
						return func(j float64) float64 { return norm180(lonAt(j) - tg.lon - ang) }
					}
					edge := func(ang float64) func(float64) float64 {
						f := fo(ang)
						return func(j float64) float64 { return math.Abs(f(j)) - orb }
					}
					s0 := at(first.sk)
					if first.sk > 0 {
						s0 = bisectTol(edge(first.ang), s0-step, s0, 0.01)
					}
					e0 := end
					if last.ek >= 0 {
						e0 = bisectTol(edge(last.ang), at(last.ek)-step, at(last.ek), 0.01)
					}
					if e0 < a || s0 > b {
						continue
					}
					h := TransitHit{Transit: bd, Natal: tg.key, Type: asp, Start: timeOfJDE(s0), End: timeOfJDE(e0)}
					var ex []float64
					for _, r := range g {
						for _, k := range r.xk {
							ex = append(ex, bisectTol(fo(r.ang), at(k)-step, at(k), 3e-4))
						}
					}
					sort.Float64s(ex)
					for _, x := range ex {
						h.Exact = append(h.Exact, timeOfJDE(x))
					}
					out = append(out, h)
				}
			}
		}
	}
	key := func(h TransitHit) time.Time {
		if len(h.Exact) > 0 {
			return h.Exact[0]
		}
		return h.Start
	}
	sort.SliceStable(out, func(i, j int) bool { return key(out[i]).Before(key(out[j])) })
	return out
}
