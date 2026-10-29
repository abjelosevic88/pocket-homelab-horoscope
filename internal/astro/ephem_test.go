package astro

import (
	"math"
	"testing"
	"time"
)

// TestAgainstDE421 checks apparent longitudes against JPL DE421.
func TestAgainstDE421(t *testing.T) {
	tol := map[Body]float64{Moon: 30, Pluto: 120} // arcseconds; others 60
	worst := map[Body]float64{}
	for _, r := range refDE421 {
		at, _ := time.Parse(time.RFC3339, r.t)
		for b, want := range r.pos {
			p := PositionOf(b, at)
			d := math.Abs(norm180(p.Lon-want[0])) * 3600
			dl := math.Abs(p.Lat-want[1]) * 3600
			if d > worst[b] {
				worst[b] = d
			}
			lim := tol[b]
			if lim == 0 {
				lim = 60
			}
			if d > lim || dl > lim {
				t.Errorf("%s %s: lon %.5f want %.5f (%.1f″), lat %.5f want %.5f (%.1f″)", r.t, b, p.Lon, want[0], d, p.Lat, want[1], dl)
			}
		}
	}
	t.Logf("worst longitude error (arcsec): %v", worst)
}
