package astro

import (
	"math"
	"sort"
	"time"
)

// Point is a calculated chart point: "asc", "mc", "dsc", "ic", "fortune",
// "southnode".
type Point struct {
	Key    string  `json:"key"`
	Lon    float64 `json:"lon"`
	Sign   int     `json:"sign"`
	Degree float64 `json:"degree"`
	House  int     `json:"house,omitempty"`
}

func makePoint(key string, lon float64) Point {
	lon = Norm360(lon)
	s := SignOf(lon)
	return Point{Key: key, Lon: lon, Sign: s, Degree: lon - float64(s*30)}
}

// ChartInput describes a birth (or any) moment and place.
type ChartInput struct {
	Time        time.Time // UTC instant; local noon when the time is unknown
	TimeKnown   bool
	Lat, Lon    float64 // degrees, north and east positive
	Zodiac      Zodiac
	HouseSystem HouseSystem
}

// Chart is a computed chart.
type Chart struct {
	Time      time.Time  `json:"time"`
	TimeKnown bool       `json:"time_known"`
	Lat       float64    `json:"lat"`
	Lon       float64    `json:"lon"`
	Zodiac    Zodiac     `json:"zodiac"`
	Positions []Position `json:"positions"`
	Points    []Point    `json:"points"`
	Houses    *Houses    `json:"houses,omitempty"`
	Aspects   []Aspect   `json:"aspects"`
	Balance   Balance    `json:"balance"`
}

// Body returns the chart's position for b (zero value when missing).
func (c Chart) Body(b Body) Position {
	for _, p := range c.Positions {
		if p.Body == b {
			return p
		}
	}
	return Position{}
}

// Point returns the chart point with the key, if present.
func (c Chart) Point(key string) (Point, bool) {
	for _, p := range c.Points {
		if p.Key == key {
			return p, true
		}
	}
	return Point{}, false
}

// LonOf returns the longitude of a body or point key, if present.
func (c Chart) LonOf(key string) (float64, bool) {
	for _, p := range c.Positions {
		if string(p.Body) == key {
			return p.Lon, true
		}
	}
	if p, ok := c.Point(key); ok {
		return p.Lon, true
	}
	return 0, false
}

// NatalChart computes positions, points, houses, aspects and the balance.
// Without a known time there are no houses, angles or Part of Fortune.
func NatalChart(in ChartInput, orbs Orbs) Chart {
	if in.Zodiac == "" {
		in.Zodiac = Tropical
	}
	if in.HouseSystem == "" {
		in.HouseSystem = Placidus
	}
	if orbs == nil {
		orbs = DefaultOrbs()
	}
	c := Chart{Time: in.Time.UTC(), TimeKnown: in.TimeKnown, Lat: in.Lat, Lon: in.Lon, Zodiac: in.Zodiac}
	c.Positions = Positions(in.Time, in.Zodiac)
	node := c.Body(Node)
	if in.TimeKnown {
		h := ComputeHouses(in.Time, in.Lat, in.Lon, in.HouseSystem, in.Zodiac)
		c.Houses = &h
		for i := range c.Positions {
			c.Positions[i].House = HouseOf(c.Positions[i].Lon, h)
		}
		sun, moon := c.Body(Sun), c.Body(Moon)
		fortune := h.Asc + moon.Lon - sun.Lon
		if sun.House >= 1 && sun.House <= 6 { // Sun below the horizon: night chart
			fortune = h.Asc + sun.Lon - moon.Lon
		}
		c.Points = []Point{
			makePoint("asc", h.Asc), makePoint("mc", h.MC),
			makePoint("dsc", h.Asc+180), makePoint("ic", h.MC+180),
			makePoint("fortune", fortune),
		}
	}
	c.Points = append(c.Points, makePoint("southnode", node.Lon+180))
	if c.Houses != nil {
		for i := range c.Points {
			c.Points[i].House = HouseOf(c.Points[i].Lon, *c.Houses)
		}
	}
	var angles []Point
	for _, p := range c.Points {
		if p.Key == "asc" || p.Key == "mc" {
			angles = append(angles, p)
		}
	}
	c.Aspects = FindAspects(c.Positions, angles, orbs)
	c.Balance = balanceOf(c)
	return c
}

// ---- aspects ----

// AspectType names an aspect.
type AspectType string

const (
	Conjunction AspectType = "conjunction"
	Sextile     AspectType = "sextile"
	Square      AspectType = "square"
	Trine       AspectType = "trine"
	Opposition  AspectType = "opposition"
	Quincunx    AspectType = "quincunx"
)

// AspectAngles maps each aspect to its angle.
var AspectAngles = map[AspectType]float64{Conjunction: 0, Sextile: 60, Square: 90, Trine: 120, Opposition: 180, Quincunx: 150}

var aspectOrder = []AspectType{Conjunction, Opposition, Trine, Square, Sextile, Quincunx}

// Ptolemaic are the five major aspects.
var Ptolemaic = []AspectType{Conjunction, Sextile, Square, Trine, Opposition}

// Aspect is an angle between two chart objects.
type Aspect struct {
	A        string     `json:"a"`
	B        string     `json:"b"`
	Type     AspectType `json:"type"`
	Orb      float64    `json:"orb"`      // degrees off exact
	Applying bool       `json:"applying"` // getting closer
	Exact    float64    `json:"exact"`    // the aspect's angle
}

// Orbs is the allowed orb per aspect type, in degrees.
type Orbs map[AspectType]float64

// DefaultOrbs are moderate orbs; the Sun and Moon get 2° more on the
// major aspects, the node and angles less.
func DefaultOrbs() Orbs {
	return Orbs{Conjunction: 8, Opposition: 8, Trine: 7, Square: 7, Sextile: 5, Quincunx: 2.5}
}

func orbFor(o Orbs, t AspectType, a, b string) float64 {
	base := o[t]
	if base <= 0 {
		return 0
	}
	lum := a == "sun" || a == "moon" || b == "sun" || b == "moon"
	if lum && t != Quincunx {
		base += 2
	}
	minor := func(k string) bool { return k == "node" || k == "southnode" || k == "fortune" }
	if minor(a) || minor(b) {
		base = math.Max(1.5, base-3)
	}
	return base
}

type obj struct {
	key   string
	lon   float64
	speed float64
}

func objects(ps []Position, pts []Point) []obj {
	out := make([]obj, 0, len(ps)+len(pts))
	for _, p := range ps {
		out = append(out, obj{string(p.Body), p.Lon, p.Speed})
	}
	for _, p := range pts {
		out = append(out, obj{p.Key, p.Lon, 0})
	}
	return out
}

// aspectBetween finds the tightest aspect within orb (b.speed 0 = B is fixed).
func aspectBetween(a, b obj, orbs Orbs) (Aspect, bool) {
	sep := math.Abs(norm180(a.lon - b.lon))
	best := Aspect{}
	found := false
	for _, t := range aspectOrder {
		lim := orbFor(orbs, t, a.key, b.key)
		if lim == 0 {
			continue
		}
		off := math.Abs(sep - AspectAngles[t])
		if off <= lim && (!found || off < best.Orb) {
			best = Aspect{A: a.key, B: b.key, Type: t, Orb: off, Exact: AspectAngles[t]}
			found = true
		}
	}
	if !found {
		return best, false
	}
	// applying when the orb shrinks a moment later
	const dt = 0.01
	later := math.Abs(math.Abs(norm180(a.lon+a.speed*dt-(b.lon+b.speed*dt))) - best.Exact)
	best.Applying = later < best.Orb
	best.Orb = math.Round(best.Orb*100) / 100
	return best, true
}

// FindAspects finds the aspects among one chart's bodies and between its
// bodies and the given points (Ascendant, MC). The node is never paired
// with the south node.
func FindAspects(ps []Position, pts []Point, orbs Orbs) []Aspect {
	if orbs == nil {
		orbs = DefaultOrbs()
	}
	bodies := objects(ps, nil)
	var out []Aspect
	for i := 0; i < len(bodies); i++ {
		for j := i + 1; j < len(bodies); j++ {
			if a, ok := aspectBetween(bodies[i], bodies[j], orbs); ok {
				out = append(out, a)
			}
		}
		for _, p := range objects(nil, pts) {
			if bodies[i].key == "node" && p.key == "southnode" {
				continue
			}
			if a, ok := aspectBetween(bodies[i], p, orbs); ok {
				out = append(out, a)
			}
		}
	}
	sortAspects(out)
	return out
}

// CrossAspects finds aspects from the first set (transiting planets or the
// first person) to the second (natal chart or the second person). Applying
// uses only the first set's motion: the second chart is fixed.
func CrossAspects(a []Position, aPts []Point, b []Position, bPts []Point, orbs Orbs) []Aspect {
	if orbs == nil {
		orbs = DefaultOrbs()
	}
	as, bs := objects(a, aPts), objects(b, bPts)
	var out []Aspect
	for _, x := range as {
		for _, y := range bs {
			y.speed = 0
			if asp, ok := aspectBetween(x, y, orbs); ok {
				out = append(out, asp)
			}
		}
	}
	sortAspects(out)
	return out
}

func sortAspects(as []Aspect) {
	sort.SliceStable(as, func(i, j int) bool { return as[i].Orb < as[j].Orb })
}

// ---- balance ----

// Balance weighs elements and modalities in a chart (percentages).
type Balance struct {
	Elements         map[string]float64 `json:"elements"`          // fire, earth, air, water
	Modalities       map[string]float64 `json:"modalities"`        // cardinal, fixed, mutable
	Dominant         string             `json:"dominant"`          // element
	DominantModality string             `json:"dominant_modality"` // modality
	Lacking          []string           `json:"lacking,omitempty"` // elements under 10%
}

// ElementOf returns the element of a sign index.
func ElementOf(sign int) string {
	return [4]string{"fire", "earth", "air", "water"}[((sign%12)+12)%12%4]
}

// ModalityOf returns the modality of a sign index.
func ModalityOf(sign int) string {
	return [3]string{"cardinal", "fixed", "mutable"}[((sign%12)+12)%12%3]
}

var balanceWeight = map[string]float64{
	"sun": 3, "moon": 3, "asc": 3, "mercury": 2, "venus": 2, "mars": 2,
	"jupiter": 1.5, "saturn": 1.5, "uranus": 1, "neptune": 1, "pluto": 1, "mc": 1,
}

func balanceOf(c Chart) Balance {
	el := map[string]float64{"fire": 0, "earth": 0, "air": 0, "water": 0}
	mo := map[string]float64{"cardinal": 0, "fixed": 0, "mutable": 0}
	total := 0.0
	add := func(key string, sign int) {
		w := balanceWeight[key]
		if w == 0 {
			return
		}
		el[ElementOf(sign)] += w
		mo[ModalityOf(sign)] += w
		total += w
	}
	for _, p := range c.Positions {
		add(string(p.Body), p.Sign)
	}
	for _, p := range c.Points {
		add(p.Key, p.Sign)
	}
	b := Balance{Elements: el, Modalities: mo}
	if total == 0 {
		return b
	}
	pick := func(m map[string]float64, order []string) string {
		best := order[0]
		for _, k := range order {
			if m[k] > m[best] {
				best = k
			}
		}
		return best
	}
	for k := range el {
		el[k] = math.Round(el[k]/total*1000) / 10
	}
	for k := range mo {
		mo[k] = math.Round(mo[k]/total*1000) / 10
	}
	b.Dominant = pick(el, []string{"fire", "earth", "air", "water"})
	b.DominantModality = pick(mo, []string{"cardinal", "fixed", "mutable"})
	for _, k := range []string{"fire", "earth", "air", "water"} {
		if el[k] < 10 {
			b.Lacking = append(b.Lacking, k)
		}
	}
	return b
}
