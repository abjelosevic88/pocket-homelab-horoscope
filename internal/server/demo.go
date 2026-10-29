package server

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/astro"
	"github.com/abjelosevic88/pocket-homelab-horoscope/internal/store"
)

// demoPeople are made-up people with plausible birth data.
var demoPeople = []store.Person{
	{Name: "Mara Kovač", Relation: "me", IsMe: true, BirthDate: "1991-04-27", BirthTime: "06:40", Place: "Sarajevo", Country: "Bosnia and Herzegovina", Lat: 43.8486, Lon: 18.3564, TZ: "Europe/Sarajevo", Color: "#6a3fb5"},
	{Name: "Luka Petrović", Relation: "partner", BirthDate: "1989-11-08", BirthTime: "22:15", Place: "Belgrade", Country: "Serbia", Lat: 44.8040, Lon: 20.4651, TZ: "Europe/Belgrade", Color: "#be123c", Favorite: true,
		Notes: "Met in Mostar, summer 2017."},
	{Name: "Ema Kovač", Relation: "family", BirthDate: "2019-07-14", BirthTime: "13:05", Place: "Sarajevo", Country: "Bosnia and Herzegovina", Lat: 43.8486, Lon: 18.3564, TZ: "Europe/Sarajevo", Color: "#f59e0b", Favorite: true},
	{Name: "Jasna Kovač", Relation: "family", BirthDate: "1964-02-02", BirthTime: "04:30", Place: "Tuzla", Country: "Bosnia and Herzegovina", Lat: 44.5384, Lon: 18.6763, TZ: "Europe/Sarajevo", Color: "#0f766e"},
	{Name: "Nina Haller", Relation: "friend", BirthDate: "1992-09-30", BirthTime: "", Place: "Vienna", Country: "Austria", Lat: 48.2085, Lon: 16.3721, TZ: "Europe/Vienna", Color: "#2563eb",
		Notes: "Birth time unknown — ask her mum."},
	{Name: "Tarik Hodžić", Relation: "friend", BirthDate: "1990-01-19", BirthTime: "11:50", Place: "Mostar", Country: "Bosnia and Herzegovina", Lat: 43.3438, Lon: 17.8078, TZ: "Europe/Sarajevo", Color: "#16a34a"},
	{Name: "Sofia Marin", Relation: "colleague", BirthDate: "1987-06-03", BirthTime: "17:20", Place: "Zagreb", Country: "Croatia", Lat: 45.8144, Lon: 15.9780, TZ: "Europe/Zagreb", Color: "#c026d3"},
}

var demoNotes = []string{
	"Slow start, then a good long walk by the river.", "Busy at work, but it went well.", "Felt restless all evening.", "Lovely dinner with Luka.",
	"Ema had a fever, little sleep.", "Finished the report early. Relief.", "Tired and a bit low.", "Coffee with Nina, laughed a lot.",
	"Argued about the holidays, made up later.", "Quiet day at home, read most of the afternoon.", "Productive morning, scattered afternoon.",
	"Mum called, long chat.", "Couldn't focus, too many meetings.", "Swimming, then pizza. Simple and good.", "",
}

var demoTags = []string{"work", "family", "sleep", "friends", "outdoors", "", "", "", "rest", "love"}

// demoReadings are kept readings, as if AI had written them.
var demoReadings = []struct{ kind, title, body string }{
	{"natal", "Mara Kovač's birth chart", "### Who you are\n\nWith the **Sun in Taurus** you are steady, practical and loyal: you like to build things that last and you do not rush. Your **Moon in Aquarius** wants room to think and friends who are also equals, so you can seem calm on the surface while your mind runs ahead. Your **rising sign, Taurus,** makes that steadiness the first thing people meet.\n\n### How you love and connect\n\nVenus in Gemini needs talk, humour and variety in love. You fall for minds before faces.\n\n### Work and drive\n\nMars in Cancer works best for people you care about. Give yourself a cause, not just a task.\n\n### Your growing edge\n\nSaturn square your Sun asks for patience with yourself. Things come to you slowly — and then they stay."},
	{"match", "Mara Kovač & Luka Petrović", "### What draws you together\n\nYour Sun in Taurus sits right opposite Luka's Sun in Scorpio: you are opposites that complete each other, both loyal, both stubborn.\n\n### Where you differ\n\nLuka needs depth and privacy; you need calm and routine. When he goes quiet, it is not a judgement.\n\n### How to make it work\n\nSay the plain thing early. Plan one slow evening a week that belongs to neither of you alone."},
}

// seedDemo fills the owner's account with people, half a year of
// check-ins and a few kept readings (DEMO_DATA=true, for screenshots and
// trying the app). Runs once.
func (s *Server) seedDemo(ctx context.Context, owner int64) error {
	key := fmt.Sprintf("seeded:%d:demo", owner)
	if _, _, done := s.store.CacheGet(ctx, key); done {
		return nil
	}
	if ps, _ := s.store.ListPeople(ctx, owner, ""); len(ps) > 0 {
		_ = s.store.CachePut(ctx, key, "1")
		return nil
	}
	st, _ := s.store.GetSettings(ctx, owner)
	if st.HomeTZ == "" {
		st.HomePlace, st.HomeCountry, st.HomeLat, st.HomeLon, st.HomeTZ = "Sarajevo", "Bosnia and Herzegovina", 43.8486, 18.3564, "Europe/Sarajevo"
		_ = s.store.SaveUserSettings(ctx, owner, st)
	}
	ids := map[string]int64{}
	for _, p := range demoPeople {
		p.Source = "manual"
		if err := s.store.CreatePerson(ctx, owner, &p); err != nil {
			return err
		}
		ids[p.Relation+p.Name] = p.ID
	}
	rng := rand.New(rand.NewSource(42))
	loc := zoneFor(st.HomeTZ)
	today, _ := localDay("", loc)
	for i := 180; i >= 1; i-- {
		if rng.Float64() < 0.22 {
			continue
		}
		day := today.AddDate(0, 0, -i)
		// a gentle pattern so the insights have something to find: brighter
		// around the full Moon, flatter around the new Moon
		ph := astro.MoonPhase(noonOf(day).UTC())
		base := 3.0
		switch ph.Name {
		case "full", "waxing_gibbous":
			base = 3.7
		case "new", "waning_crescent":
			base = 2.6
		}
		mood := int(base + rng.NormFloat64()*0.9 + 0.5)
		energy := int(base + rng.NormFloat64()*1.0 + 0.3)
		e := store.JournalEntry{Date: day.Format("2006-01-02"), Mood: min(5, max(1, mood)), Energy: min(5, max(1, energy)),
			Note: demoNotes[rng.Intn(len(demoNotes))], Tags: demoTags[rng.Intn(len(demoTags))], Source: "app"}
		e.Sky = s.skyOf(ctx, owner, st, e.Date)
		if err := s.store.SaveJournal(ctx, owner, &e); err != nil {
			return err
		}
	}
	me, _ := s.store.GetMe(ctx, owner)
	for _, d := range demoReadings {
		r := store.Reading{Kind: d.kind, Title: d.title, Body: d.body, Model: "demo", Saved: true}
		if me != nil {
			r.PersonID = &me.ID
		}
		if d.kind == "match" {
			for k, id := range ids {
				if len(k) > 7 && k[:7] == "partner" {
					id := id
					r.OtherID = &id
				}
			}
		}
		_ = s.store.CreateReading(ctx, owner, &r)
	}
	s.log.Info("demo data seeded", "people", len(demoPeople))
	return s.store.CachePut(ctx, key, "1")
}
