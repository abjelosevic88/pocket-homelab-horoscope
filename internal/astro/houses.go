package astro

import (
	"math"
	"time"
)

// HouseSystem names a way of dividing the chart into twelve houses.
type HouseSystem string

const (
	Placidus HouseSystem = "placidus"
	Koch     HouseSystem = "koch"
	Whole    HouseSystem = "whole"
	Equal    HouseSystem = "equal"
	Porphyry HouseSystem = "porphyry"
)

// HouseSystems lists the supported systems.
var HouseSystems = []HouseSystem{Placidus, Koch, Whole, Equal, Porphyry}

// Houses are the twelve cusps plus the angles, in degrees.
type Houses struct {
	System   HouseSystem `json:"system"`
	Cusps    [12]float64 `json:"cusps"` // Cusps[0] is the 1st house
	Asc      float64     `json:"asc"`
	MC       float64     `json:"mc"`
	Fallback string      `json:"fallback,omitempty"` // why another system was used
}

// ComputeHouses returns the houses for a moment and place (lon east positive).
// Placidus and Koch are undefined near the poles; above 66° latitude they fall
// back to Porphyry and say so in Fallback.
func ComputeHouses(t time.Time, lat, lon float64, sys HouseSystem, z Zodiac) Houses {
	jde := jdeOf(t)
	ramc := Norm360(siderealDeg(t) + lon)
	eps := trueObliquity(jde)
	h := housesFrom(ramc, lat, eps, sys)
	if z == Sidereal {
		a := ayanamsaJDE(jde)
		h.Asc = Norm360(h.Asc - a)
		h.MC = Norm360(h.MC - a)
		for i := range h.Cusps {
			h.Cusps[i] = Norm360(h.Cusps[i] - a)
		}
		if h.System == Whole { // whole signs follow the sidereal sign of the Ascendant
			start := float64(SignOf(h.Asc) * 30)
			for i := range h.Cusps {
				h.Cusps[i] = Norm360(start + float64(i*30))
			}
		}
	}
	return h
}

// siderealDeg returns Greenwich apparent sidereal time in degrees.
func siderealDeg(t time.Time) float64 {
	jd := jdUT(t)
	T := (jd - j2000) / 36525
	gmst := 280.46061837 + 360.98564736629*(jd-j2000) + 0.000387933*T*T - T*T*T/38710000
	jde := jdeOf(t)
	dpsi, deps := nutation(jde)
	eps := meanObliquity(jde) + deps
	return Norm360(gmst + dpsi*math.Cos(eps*deg2rad))
}

// mcOf returns the ecliptic longitude culminating at a given RAMC.
func mcOf(ramc, eps float64) float64 {
	r := ramc * deg2rad
	return Norm360(math.Atan2(math.Sin(r), math.Cos(r)*math.Cos(eps*deg2rad)) * rad2deg)
}

// ascOf returns the ecliptic longitude rising at a given RAMC and latitude.
func ascOf(ramc, lat, eps float64) float64 {
	r := ramc * deg2rad
	e := eps * deg2rad
	return Norm360(math.Atan2(math.Cos(r), -(math.Sin(r)*math.Cos(e)+math.Tan(lat*deg2rad)*math.Sin(e))) * rad2deg)
}

// lonOfRA returns the ecliptic longitude (latitude 0) with right ascension ra.
func lonOfRA(ra, eps float64) float64 {
	r := ra * deg2rad
	return Norm360(math.Atan2(math.Sin(r), math.Cos(r)*math.Cos(eps*deg2rad)) * rad2deg)
}

func housesFrom(ramc, lat, eps float64, sys HouseSystem) Houses {
	h := Houses{System: sys, MC: mcOf(ramc, eps), Asc: ascOf(ramc, lat, eps)}
	if (sys == Placidus || sys == Koch) && math.Abs(lat) > 66 {
		h.Fallback = "Placidus and Koch houses can't be drawn this close to the poles; using Porphyry."
		sys = Porphyry
	}
	c := &h.Cusps
	switch sys {
	case Whole:
		start := float64(SignOf(h.Asc) * 30)
		for i := range c {
			c[i] = Norm360(start + float64(i*30))
		}
		return h
	case Equal:
		for i := range c {
			c[i] = Norm360(h.Asc + float64(i*30))
		}
		return h
	case Placidus:
		c[10] = placidusCusp(ramc, lat, eps, 1.0/3, true)
		c[11] = placidusCusp(ramc, lat, eps, 2.0/3, true)
		c[1] = placidusCusp(ramc, lat, eps, 2.0/3, false)
		c[2] = placidusCusp(ramc, lat, eps, 1.0/3, false)
	case Koch:
		dmc := math.Asin(math.Sin(eps*deg2rad)*math.Sin(h.MC*deg2rad)) * rad2deg
		ad := ascDiff(dmc, lat) / 3
		asc := func(oa float64) float64 { return ascOf(oa-90, lat, eps) } // point rising at oblique ascension oa
		// the MC degree's diurnal semi-arc in thirds of time (Swiss Ephemeris' form)
		c[10] = asc(ramc + 30 - 2*ad)
		c[11] = asc(ramc + 60 - ad)
		c[1] = asc(ramc + 120 + ad)
		c[2] = asc(ramc + 150 + 2*ad)
	default: // Porphyry: trisect each quadrant
		q1 := Norm360(h.Asc - h.MC) // MC → Asc
		q2 := 180 - q1              // Asc → IC
		c[10] = Norm360(h.MC + q1/3)
		c[11] = Norm360(h.MC + 2*q1/3)
		c[1] = Norm360(h.Asc + q2/3)
		c[2] = Norm360(h.Asc + 2*q2/3)
	}
	c[0], c[9] = h.Asc, h.MC
	c[6], c[3] = Norm360(h.Asc+180), Norm360(h.MC+180)
	c[4], c[5] = Norm360(c[10]+180), Norm360(c[11]+180)
	c[7], c[8] = Norm360(c[1]+180), Norm360(c[2]+180)
	return h
}

// ascDiff returns the ascensional difference (degrees) for a declination.
func ascDiff(dec, lat float64) float64 {
	x := math.Tan(dec*deg2rad) * math.Tan(lat*deg2rad)
	return math.Asin(math.Max(-1, math.Min(1, x))) * rad2deg
}

// placidusCusp iterates a Placidus cusp: f is the fraction of the diurnal
// (above, cusps 11/12) or nocturnal (below, cusps 2/3) semi-arc.
func placidusCusp(ramc, lat, eps, f float64, above bool) float64 {
	ra := ramc + f*90
	if !above {
		ra = ramc + 180 - f*90
	}
	lon := lonOfRA(ra, eps)
	for i := 0; i < 50; i++ {
		dec := math.Asin(math.Sin(eps*deg2rad)*math.Sin(lon*deg2rad)) * rad2deg
		ad := ascDiff(dec, lat)
		if above {
			ra = ramc + f*(90+ad)
		} else {
			ra = ramc + 180 - f*(90-ad)
		}
		next := lonOfRA(ra, eps)
		if math.Abs(norm180(next-lon)) < 1e-9 {
			return next
		}
		lon = next
	}
	return lon
}

// HouseOf returns the house (1–12) a longitude falls in.
func HouseOf(lon float64, h Houses) int {
	lon = Norm360(lon)
	for i := 0; i < 12; i++ {
		a, b := h.Cusps[i], h.Cusps[(i+1)%12]
		span := Norm360(b - a)
		if Norm360(lon-a) < span {
			return i + 1
		}
	}
	return 1
}
