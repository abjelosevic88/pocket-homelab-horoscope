// Package astro computes astrological charts from real planet positions:
// apparent geocentric longitudes (VSOP87 for Mercury–Neptune and the Sun,
// Meeus' lunar and Pluto theories), house systems, aspects, the Moon's
// phases, sky events (lunations, ingresses, stations, eclipses, void-of-course
// Moons) and transits to a natal chart. Pure Go, no data files.
//
// Accuracy against JPL DE421 (1900–2049): Sun–Neptune within ~3″,
// the Moon within ~25″, Pluto a few arcseconds (1885–2099; outside that a
// Keplerian fallback good to about a tenth of a degree).
package astro

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Body is a moving chart object.
type Body string

const (
	Sun     Body = "sun"
	Moon    Body = "moon"
	Mercury Body = "mercury"
	Venus   Body = "venus"
	Mars    Body = "mars"
	Jupiter Body = "jupiter"
	Saturn  Body = "saturn"
	Uranus  Body = "uranus"
	Neptune Body = "neptune"
	Pluto   Body = "pluto"
	Node    Body = "node" // true north (ascending) lunar node
)

// Bodies lists every body in chart order.
var Bodies = []Body{Sun, Moon, Mercury, Venus, Mars, Jupiter, Saturn, Uranus, Neptune, Pluto, Node}

// Signs are the zodiac sign keys, index 0 (Aries) to 11 (Pisces).
var Signs = []string{"aries", "taurus", "gemini", "cancer", "leo", "virgo", "libra", "scorpio", "sagittarius", "capricorn", "aquarius", "pisces"}

// Zodiac selects tropical or sidereal (Lahiri) longitudes.
type Zodiac string

const (
	Tropical Zodiac = "tropical"
	Sidereal Zodiac = "sidereal"
)

// Position is a body's place in the zodiac.
type Position struct {
	Body   Body    `json:"body"`
	Lon    float64 `json:"lon"`    // ecliptic longitude, degrees 0–360
	Lat    float64 `json:"lat"`    // ecliptic latitude, degrees
	Speed  float64 `json:"speed"`  // degrees per day
	Retro  bool    `json:"retro"`  // moving backwards
	Sign   int     `json:"sign"`   // 0–11
	Degree float64 `json:"degree"` // 0–30 within the sign
	House  int     `json:"house,omitempty"`
}

// Norm360 folds an angle into [0, 360).
func Norm360(d float64) float64 {
	d = math.Mod(d, 360)
	if d < 0 {
		d += 360
	}
	return d
}

// norm180 folds an angle into [-180, 180).
func norm180(d float64) float64 {
	d = Norm360(d)
	if d >= 180 {
		d -= 360
	}
	return d
}

// SignOf returns the sign index (0–11) of a longitude.
func SignOf(lon float64) int {
	s := int(Norm360(lon) / 30)
	if s > 11 {
		s = 11
	}
	return s
}

// SignName returns the capitalised sign name for an index.
func SignName(i int) string {
	if i < 0 || i > 11 {
		return ""
	}
	return strings.ToUpper(Signs[i][:1]) + Signs[i][1:]
}

// FormatDeg renders a longitude as "14°32′ Sagittarius".
func FormatDeg(lon float64) string {
	lon = Norm360(lon)
	s := SignOf(lon)
	d := lon - float64(s*30)
	deg := int(d)
	min := int(math.Round((d - float64(deg)) * 60))
	if min == 60 {
		deg, min = deg+1, 0
	}
	if deg == 30 { // rounding at the very end of a sign
		deg, s = 0, (s+1)%12
	}
	return fmt.Sprintf("%d°%02d′ %s", deg, min, SignName(s))
}

// Ayanamsa returns the Lahiri (Chitrapaksha) ayanamsa in degrees: 23°15′00.658″
// on 1956-03-21 0h TT, moved by general precession (IAU 2006).
func Ayanamsa(t time.Time) float64 {
	return ayanamsaJDE(jdeOf(t))
}

func ayanamsaJDE(jde float64) float64 {
	const t0 = 2435553.5
	p := func(jde float64) float64 {
		T := (jde - j2000) / 36525
		return (5028.796195*T + 1.1054348*T*T) / 3600
	}
	return 23.245524743 + p(jde) - p(t0)
}

// withZodiac converts a tropical longitude.
func withZodiac(lon float64, z Zodiac, jde float64) float64 {
	if z == Sidereal {
		return Norm360(lon - ayanamsaJDE(jde))
	}
	return lon
}

func makePosition(b Body, lon, lat, speed float64) Position {
	lon = Norm360(lon)
	s := SignOf(lon)
	return Position{Body: b, Lon: lon, Lat: lat, Speed: speed, Retro: speed < 0, Sign: s, Degree: lon - float64(s*30)}
}

// PositionOf returns a body's apparent tropical position at t.
func PositionOf(b Body, t time.Time) Position {
	jde := jdeOf(t)
	lon, lat := bodyLonLat(b, jde)
	return makePosition(b, lon, lat, bodySpeed(b, jde))
}

// Positions returns every body at t in the chosen zodiac.
func Positions(t time.Time, z Zodiac) []Position {
	jde := jdeOf(t)
	const h = 0.25 // days either side for the planets' speed
	s0, sm, sp := newSky(jde), newSkyCut(jde-h, coarseCutoff), newSkyCut(jde+h, coarseCutoff)
	out := make([]Position, len(Bodies))
	for i, b := range Bodies {
		lon, lat := s0.at(b)
		var speed float64
		if b == Moon {
			const hm = 0.05
			lo, _, _ := moonPos(jde-hm, s0.dpsi)
			hi, _, _ := moonPos(jde+hm, s0.dpsi)
			speed = norm180(hi-lo) / (2 * hm)
		} else {
			lo, _ := sm.at(b)
			hi, _ := sp.at(b)
			speed = norm180(hi-lo) / (2 * h)
		}
		out[i] = makePosition(b, withZodiac(lon, z, jde), lat, speed)
	}
	return out
}

func speedStep(b Body) float64 {
	if b == Moon {
		return 0.05
	}
	return 0.25
}

func bodySpeed(b Body, jde float64) float64 {
	h := speedStep(b)
	return norm180(bodyLon(b, jde+h)-bodyLon(b, jde-h)) / (2 * h)
}

// MeanNode returns the mean north node's tropical longitude at t.
func MeanNode(t time.Time) float64 {
	return meanNode(jdeOf(t))
}
