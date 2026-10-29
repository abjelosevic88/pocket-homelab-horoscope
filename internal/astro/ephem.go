package astro

import "math"

const (
	deg2rad = math.Pi / 180
	rad2deg = 180 / math.Pi
	// light time for one AU, in days
	lightDay = 0.0057755183
)

// vec is a rectangular J2000 ecliptic vector (AU).
type vec struct{ x, y, z float64 }

func (a vec) sub(b vec) vec { return vec{a.x - b.x, a.y - b.y, a.z - b.z} }
func (a vec) len() float64  { return math.Sqrt(a.x*a.x + a.y*a.y + a.z*a.z) }

func spherical(l, b, r float64) vec {
	sl, cl := math.Sincos(l)
	sb, cb := math.Sincos(b)
	return vec{r * cb * cl, r * cb * sl, r * sb}
}

// helio returns a planet's heliocentric J2000 ecliptic position (AU).
// cutoff > 0 drops the smaller VSOP87 terms (coarse sampling, speeds).
func helio(b Body, jde, cutoff float64) vec {
	if b == Pluto {
		return plutoHelio(jde)
	}
	p := vsopData[string(b)]
	l, bb, r := vsopHelio(&p, (jde-j2000)/365250, cutoff)
	return spherical(l, bb, r)
}

func earthHelio(jde, cutoff float64) vec {
	p := vsopData["earth"]
	l, b, r := vsopHelio(&p, (jde-j2000)/365250, cutoff)
	return spherical(l, b, r)
}

// precess moves J2000 ecliptic coordinates (radians) to the ecliptic and
// mean equinox of date (Meeus 21.5 with T = 0).
func precess(lon, lat, jde float64) (float64, float64) {
	t := (jde - j2000) / 36525
	eta := (47.0029*t - 0.03302*t*t + 0.000060*t*t*t) / 3600 * deg2rad
	pi := 174.876384*deg2rad - (869.8089*t-0.03536*t*t)/3600*deg2rad
	p := (5029.0966*t + 1.11113*t*t - 0.000006*t*t*t) / 3600 * deg2rad
	se, ce := math.Sincos(eta)
	sb, cb := math.Sincos(lat)
	sd, cd := math.Sincos(pi - lon)
	a := ce*cb*sd - se*sb
	bb := cb * cd
	c := ce*sb + se*cb*sd
	return p + pi - math.Atan2(a, bb), math.Asin(c)
}

// nutation returns Δψ and Δε in degrees (IAU 1980, Meeus table 22.A).
func nutation(jde float64) (dpsi, deps float64) {
	T := (jde - j2000) / 36525
	D := (297.85036 + 445267.11148*T - 0.0019142*T*T + T*T*T/189474) * deg2rad
	M := (357.52772 + 35999.050340*T - 0.0001603*T*T - T*T*T/300000) * deg2rad
	N := (134.96298 + 477198.867398*T + 0.0086972*T*T + T*T*T/56250) * deg2rad
	F := (93.27191 + 483202.017538*T - 0.0036825*T*T + T*T*T/327270) * deg2rad
	O := (125.04452 - 1934.136261*T + 0.0020708*T*T + T*T*T/450000) * deg2rad
	var sp, se float64
	for i := len(table22A) - 1; i >= 0; i-- {
		r := table22A[i]
		s, c := math.Sincos(r.d*D + r.m*M + r.n*N + r.f*F + r.ω*O)
		sp += s * (r.s0 + r.s1*T)
		se += c * (r.c0 + r.c1*T)
	}
	return sp * 0.0001 / 3600, se * 0.0001 / 3600
}

// meanObliquity returns ε₀ in degrees (Laskar).
func meanObliquity(jde float64) float64 {
	u := (jde - j2000) / 3652500
	c := []float64{84381.448, -4680.93, -1.55, 1999.25, -51.38, -249.67, -39.05, 7.12, 27.87, 5.79, 2.45}
	s, p := 0.0, 1.0
	for _, k := range c {
		s += k * p
		p *= u
	}
	return s / 3600
}

// trueObliquity returns ε = ε₀ + Δε in degrees.
func trueObliquity(jde float64) float64 {
	_, de := nutation(jde)
	return meanObliquity(jde) + de
}

// sky caches what every body at one instant shares.
type sky struct {
	jde     float64
	earth   vec
	dpsi    float64 // degrees
	sunLon  float64 // true geometric longitude of date, degrees
	ecc, pe float64 // Earth's eccentricity and perihelion longitude (deg)
	cutoff  float64
}

func newSky(jde float64) *sky { return newSkyCut(jde, 0) }

// coarseCutoff keeps VSOP87 terms above 1e-6 (a few arcseconds): plenty for
// speeds and for sampling before an exact search.
const coarseCutoff = 1e-6

func newSkyCut(jde, cutoff float64) *sky {
	s := &sky{jde: jde, earth: earthHelio(jde, cutoff), cutoff: cutoff}
	s.dpsi, _ = nutation(jde)
	e := s.earth
	l := math.Atan2(-e.y, -e.x)
	b := math.Atan2(-e.z, math.Hypot(e.x, e.y))
	l, _ = precess(l, b, jde)
	s.sunLon = Norm360(l * rad2deg)
	T := (jde - j2000) / 36525
	s.ecc = 0.016708634 - 0.000042037*T
	s.pe = 102.93735 + 1.71946*T
	return s
}

// planet returns the apparent longitude and latitude (degrees, ecliptic and
// true equinox of date) of a planet (Mercury–Pluto).
func (s *sky) planet(b Body) (float64, float64) {
	g := helio(b, s.jde, s.cutoff).sub(s.earth)
	tau := lightDay * g.len()
	g = helio(b, s.jde-tau, s.cutoff).sub(s.earth)
	lon := math.Atan2(g.y, g.x)
	lat := math.Atan2(g.z, math.Hypot(g.x, g.y))
	lon, lat = precess(lon, lat, s.jde)
	lond, latd := lon*rad2deg, lat*rad2deg
	// annual aberration (Meeus 23.2)
	const k = 20.49552 / 3600
	cb := math.Cos(lat)
	dl := (-k*math.Cos((s.sunLon-lond)*deg2rad) + s.ecc*k*math.Cos((s.pe-lond)*deg2rad)) / cb
	db := -k * math.Sin(lat) * (math.Sin((s.sunLon-lond)*deg2rad) - s.ecc*math.Sin((s.pe-lond)*deg2rad))
	return Norm360(lond + dl + s.dpsi), latd + db
}

// sun returns the Sun's apparent longitude and latitude in degrees.
func (s *sky) sun() (float64, float64) {
	e := s.earth
	r := e.len()
	l := math.Atan2(-e.y, -e.x)
	b := math.Atan2(-e.z, math.Hypot(e.x, e.y))
	l, b = precess(l, b, s.jde)
	return Norm360(l*rad2deg + s.dpsi - 20.4898/3600/r), b * rad2deg
}

// moon returns the Moon's apparent longitude, latitude (degrees) and distance (km).
func moonPos(jde, dpsi float64) (float64, float64, float64) {
	T := (jde - j2000) / 36525
	const p = deg2rad
	horner := func(x float64, c ...float64) float64 {
		s := 0.0
		for i := len(c) - 1; i >= 0; i-- {
			s = s*x + c[i]
		}
		return s
	}
	Lp := horner(T, 218.3164477*p, 481267.88123421*p, -.0015786*p, p/538841, -p/65194000)
	D, M, Mp, F := moonArgs(T)
	A1 := 119.75*p + 131.849*p*T
	A2 := 53.09*p + 479264.29*p*T
	A3 := 313.45*p + 481266.484*p*T
	E := horner(T, 1, -.002516, -.0000074)
	E2 := E * E
	sl := 3958*math.Sin(A1) + 1962*math.Sin(Lp-F) + 318*math.Sin(A2)
	sr := 0.
	sb := -2235*math.Sin(Lp) + 382*math.Sin(A3) + 175*math.Sin(A1-F) +
		175*math.Sin(A1+F) + 127*math.Sin(Lp-Mp) - 115*math.Sin(Lp+Mp)
	for i := range ta {
		r := &ta[i]
		sa, ca := math.Sincos(D*r.D + M*r.M + Mp*r.Mp + F*r.F)
		f := 1.0
		switch r.M {
		case 1, -1:
			f = E
		case 2, -2:
			f = E2
		}
		sl += r.Σl * sa * f
		sr += r.Σr * ca * f
	}
	for i := range tb {
		r := &tb[i]
		f := 1.0
		switch r.M {
		case 1, -1:
			f = E
		case 2, -2:
			f = E2
		}
		sb += r.Σb * math.Sin(D*r.D+M*r.M+Mp*r.Mp+F*r.F) * f
	}
	lon := Norm360(Lp*rad2deg + sl*1e-6 + dpsi)
	return lon, sb * 1e-6, 385000.56 + sr*1e-3
}

func moonArgs(T float64) (D, M, Mp, F float64) {
	const p = deg2rad
	h := func(c ...float64) float64 {
		s := 0.0
		for i := len(c) - 1; i >= 0; i-- {
			s = s*T + c[i]
		}
		return s
	}
	D = h(297.8501921*p, 445267.1114034*p, -.0018819*p, p/545868, -p/113065000)
	M = h(357.5291092*p, 35999.0502909*p, -.0001535*p, p/24490000)
	Mp = h(134.9633964*p, 477198.8675055*p, .0087414*p, p/69699, -p/14712000)
	F = h(93.272095*p, 483202.0175233*p, -.0036539*p, -p/3526000, p/863310000)
	return
}

func meanNode(jde float64) float64 {
	T := (jde - j2000) / 36525
	return Norm360(125.0445479 - 1934.1362891*T + .0020754*T*T + T*T*T/467441 - T*T*T*T/60616000)
}

// trueNode returns the true node's apparent longitude in degrees (Meeus).
func trueNode(jde, dpsi float64) float64 {
	D, M, Mp, F := moonArgs((jde - j2000) / 36525)
	return Norm360(meanNode(jde) - 1.4979*math.Sin(2*(D-F)) - .15*math.Sin(M) -
		.1226*math.Sin(2*D) + .1176*math.Sin(2*F) - .0801*math.Sin(2*(Mp-F)) + dpsi)
}

// plutoHelio returns Pluto's heliocentric J2000 ecliptic position: Meeus
// chapter 37 between 1885 and 2099, Keplerian mean elements (JPL, Standish)
// outside it.
func plutoHelio(jde float64) vec {
	y := yearOf(jde)
	if y < 1885 || y > 2099 {
		return keplerPluto(jde)
	}
	T := (jde - j2000) / 36525
	J := (34.35 + 3034.9057*T) * deg2rad
	S := (50.08 + 1222.1138*T) * deg2rad
	P := (238.96 + 144.96*T) * deg2rad
	var l, b, r float64
	for i := range t37 {
		t := &t37[i]
		sa, ca := math.Sincos(J*t.i + S*t.j + P*t.k)
		l += t.lA*sa + t.lB*ca
		b += t.bA*sa + t.bB*ca
		r += t.rA*sa + t.rB*ca
	}
	l = (l + 238.958116 + 144.96*T) * deg2rad
	b = (b - 3.908239) * deg2rad
	r += 40.7241346
	return spherical(l, b, r)
}

func keplerPluto(jde float64) vec {
	T := (jde - j2000) / 36525
	a := 39.48211675 - 0.00031596*T
	e := 0.24882730 + 0.00005170*T
	I := (17.14001206 + 0.00004818*T) * deg2rad
	L := 238.92903833 + 145.20780515*T
	w := 224.06891629 - 0.04062942*T // longitude of perihelion
	O := (110.30393684 - 0.01183482*T) * deg2rad
	M := Norm360(L-w) * deg2rad
	E := M
	for i := 0; i < 20; i++ {
		E -= (E - e*math.Sin(E) - M) / (1 - e*math.Cos(E))
	}
	xp := a * (math.Cos(E) - e)
	yp := a * math.Sqrt(1-e*e) * math.Sin(E)
	om := w*deg2rad - O // argument of perihelion
	so, co := math.Sincos(om)
	sO, cO := math.Sincos(O)
	sI, cI := math.Sincos(I)
	return vec{
		(co*cO-so*sO*cI)*xp + (-so*cO-co*sO*cI)*yp,
		(co*sO+so*cO*cI)*xp + (-so*sO+co*cO*cI)*yp,
		(so*sI)*xp + (co*sI)*yp,
	}
}

// at returns a body's apparent tropical longitude and latitude at jde.
func (s *sky) at(b Body) (float64, float64) {
	switch b {
	case Sun:
		return s.sun()
	case Moon:
		lon, lat, _ := moonPos(s.jde, s.dpsi)
		return lon, lat
	case Node:
		return trueNode(s.jde, s.dpsi), 0
	default:
		return s.planet(b)
	}
}

func bodyLonLat(b Body, jde float64) (float64, float64) {
	if b == Moon || b == Node {
		dpsi, _ := nutation(jde)
		if b == Moon {
			lon, lat, _ := moonPos(jde, dpsi)
			return lon, lat
		}
		return trueNode(jde, dpsi), 0
	}
	return newSky(jde).at(b)
}

func bodyLon(b Body, jde float64) float64 {
	l, _ := bodyLonLat(b, jde)
	return l
}
