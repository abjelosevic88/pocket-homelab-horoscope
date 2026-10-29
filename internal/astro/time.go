package astro

import (
	"math"
	"time"
)

const j2000 = 2451545.0

// jdUT returns the Julian day (UT) of t.
func jdUT(t time.Time) float64 {
	return float64(t.UnixNano())/86400e9 + 2440587.5
}

// jdeOf returns the Julian Ephemeris Day (TT) of t.
func jdeOf(t time.Time) float64 {
	jd := jdUT(t)
	return jd + deltaT(yearOf(jd))/86400
}

// timeOfJDE converts a Julian Ephemeris Day back to UTC, rounded to the second.
func timeOfJDE(jde float64) time.Time {
	jd := jde - deltaT(yearOf(jde))/86400
	ns := (jd - 2440587.5) * 86400e9
	return time.Unix(0, int64(ns)).UTC().Round(time.Second)
}

func yearOf(jd float64) float64 {
	return 2000 + (jd-j2000)/365.25
}

// deltaT returns TT − UT in seconds (Espenak & Meeus polynomials, with the
// observed values from 2000 on and a gentle extrapolation after 2025).
func deltaT(y float64) float64 {
	switch {
	case y < 1700:
		u := (y - 1820) / 100
		return -20 + 32*u*u
	case y < 1800:
		t := y - 1700
		return 8.83 + 0.1603*t - 0.0059285*t*t + 0.00013336*t*t*t - t*t*t*t/1174000
	case y < 1860:
		t := y - 1800
		return 13.72 - 0.332447*t + 0.0068612*t*t + 0.0041116*math.Pow(t, 3) - 0.00037436*math.Pow(t, 4) +
			0.0000121272*math.Pow(t, 5) - 0.0000001699*math.Pow(t, 6) + 0.000000000875*math.Pow(t, 7)
	case y < 1900:
		t := y - 1860
		return 7.62 + 0.5737*t - 0.251754*t*t + 0.01680668*math.Pow(t, 3) - 0.0004473624*math.Pow(t, 4) + math.Pow(t, 5)/233174
	case y < 1920:
		t := y - 1900
		return -2.79 + 1.494119*t - 0.0598939*t*t + 0.0061966*math.Pow(t, 3) - 0.000197*math.Pow(t, 4)
	case y < 1941:
		t := y - 1920
		return 21.20 + 0.84493*t - 0.076100*t*t + 0.0020936*math.Pow(t, 3)
	case y < 1961:
		t := y - 1950
		return 29.07 + 0.407*t - t*t/233 + math.Pow(t, 3)/2547
	case y < 1986:
		t := y - 1975
		return 45.45 + 1.067*t - t*t/260 - math.Pow(t, 3)/718
	case y < 2000:
		t := y - 2000
		return 63.86 + 0.3345*t - 0.060374*t*t + 0.0017275*math.Pow(t, 3) + 0.000651814*math.Pow(t, 4) + 0.00002373599*math.Pow(t, 5)
	case y < 2025:
		// observed (IERS), linear between 5-year marks
		obs := []float64{63.83, 64.69, 66.07, 67.64, 69.36, 69.20}
		i := int((y - 2000) / 5)
		f := (y - 2000 - float64(i*5)) / 5
		return obs[i] + (obs[i+1]-obs[i])*f
	case y < 2050:
		return 69.20 + 0.3*(y-2025)
	case y < 2150:
		u := (y - 1820) / 100
		return -20 + 32*u*u - 0.5628*(2150-y)
	default:
		u := (y - 1820) / 100
		return -20 + 32*u*u
	}
}
