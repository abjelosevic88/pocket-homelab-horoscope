package astro

import "math"

// vsopTerm is one A·cos(B + C·t) term.
type vsopTerm [3]float64

// vsopPlanet holds the L, B, R series, each split by power of t.
type vsopPlanet [3][][]vsopTerm

// vsopHelio returns heliocentric ecliptic longitude, latitude (radians, J2000
// ecliptic and equinox) and distance (AU) at t Julian millennia (TDB) from J2000.
// Terms within a series are sorted by decreasing amplitude, so a cutoff > 0
// stops early (used for speeds and coarse sampling).
func vsopHelio(p *vsopPlanet, t, cutoff float64) (l, b, r float64) {
	var out [3]float64
	for v := 0; v < 3; v++ {
		tp := 1.0
		for _, series := range p[v] {
			s := 0.0
			for _, term := range series {
				if term[0] < cutoff {
					break
				}
				s += term[0] * math.Cos(term[1]+term[2]*t)
			}
			out[v] += s * tp
			tp *= t
		}
	}
	return out[0], out[1], out[2]
}
