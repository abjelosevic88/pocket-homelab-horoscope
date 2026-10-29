package content

// Planet describes a planet, the lunar node or a chart point.
type Planet struct {
	Key      string   `json:"key"`
	Name     string   `json:"name"`
	Glyph    string   `json:"glyph"`
	Keywords []string `json:"keywords"`
	Meaning  string   `json:"meaning"`
	Short    string   `json:"short"`
	Cycle    string   `json:"cycle"`
}

var planets = []Planet{
	{Key: "sun", Name: "Sun", Glyph: "☉", Keywords: []string{"identity", "vitality", "purpose"},
		Meaning: "The Sun is your core self: what makes you feel alive, the qualities you are growing into and the way you shine when you are at your best. It is the sign people mean when they ask \"what's your sign?\".",
		Short:   "who you are at heart and what lights you up", Cycle: "one year around the zodiac, about a month in each sign"},
	{Key: "moon", Name: "Moon", Glyph: "☽", Keywords: []string{"feelings", "needs", "comfort"},
		Meaning: "The Moon is your emotional side: how you react, what makes you feel safe and how you look after yourself and others. It shows what you need in order to feel at home.",
		Short:   "how you feel and what you need to feel safe", Cycle: "about 27 days around the zodiac, two to three days in each sign"},
	{Key: "mercury", Name: "Mercury", Glyph: "☿", Keywords: []string{"thinking", "talking", "learning"},
		Meaning: "Mercury is your mind and your voice: how you think, learn, speak and write. It also covers short trips, messages and the everyday business of getting things arranged.",
		Short:   "how you think and communicate", Cycle: "about one year around the zodiac, turning retrograde three or four times"},
	{Key: "venus", Name: "Venus", Glyph: "♀", Keywords: []string{"love", "pleasure", "values"},
		Meaning: "Venus is what you love and how you love: your taste, your way of attracting and relating to people, and what you find beautiful and worth spending on.",
		Short:   "how you love and what you enjoy", Cycle: "about one year around the zodiac, retrograde every 18 months"},
	{Key: "mars", Name: "Mars", Glyph: "♂", Keywords: []string{"drive", "courage", "desire"},
		Meaning: "Mars is your drive: how you go after what you want, how you handle anger and competition, and where you find the energy to act.",
		Short:   "how you take action and assert yourself", Cycle: "about two years around the zodiac"},
	{Key: "jupiter", Name: "Jupiter", Glyph: "♃", Keywords: []string{"growth", "luck", "wisdom"},
		Meaning: "Jupiter is growth and generosity. It shows where life tends to open doors for you, where you are optimistic, and where you might overdo it.",
		Short:   "where you grow and find opportunity", Cycle: "about 12 years around the zodiac, a year in each sign"},
	{Key: "saturn", Name: "Saturn", Glyph: "♄", Keywords: []string{"structure", "responsibility", "time"},
		Meaning: "Saturn is structure and responsibility. It shows where you work hardest, where you feel tested and where, over time, you build real mastery.",
		Short:   "where you take responsibility and build for the long term", Cycle: "about 29 years around the zodiac"},
	{Key: "uranus", Name: "Uranus", Glyph: "♅", Keywords: []string{"change", "freedom", "originality"},
		Meaning: "Uranus is change and originality. It shows where you need freedom, where you break the rules and where life brings sudden surprises.",
		Short:   "where you need freedom and break new ground", Cycle: "about 84 years around the zodiac, seven years in each sign"},
	{Key: "neptune", Name: "Neptune", Glyph: "♆", Keywords: []string{"dreams", "imagination", "compassion"},
		Meaning: "Neptune is imagination, ideals and compassion. It shows where you dream, where you are most sensitive and where things can get a little hazy.",
		Short:   "where you dream and where things blur", Cycle: "about 165 years around the zodiac, 14 years in each sign"},
	{Key: "pluto", Name: "Pluto", Glyph: "♇", Keywords: []string{"power", "depth", "renewal"},
		Meaning: "Pluto is deep change and personal power. It shows where you go through endings and new beginnings, and where you find strength you did not know you had.",
		Short:   "where you transform and find your power", Cycle: "about 248 years around the zodiac, 12 to 30 years in each sign"},
	{Key: "node", Name: "North Node", Glyph: "☊", Keywords: []string{"direction", "growth", "purpose"},
		Meaning: "The North Node is a point, not a planet: where the Moon's path crosses the Sun's. Astrologers read it as the direction you are growing toward in this life, the unfamiliar path that brings the most fulfilment.",
		Short:   "the direction you are growing toward", Cycle: "about 18.6 years around the zodiac, moving backwards"},
}

var points = []Planet{
	{Key: "asc", Name: "Ascendant", Glyph: "AC", Keywords: []string{"first impression", "approach", "body"},
		Meaning: "The Ascendant, or rising sign, is the sign that was rising on the eastern horizon when you were born. It shows your first impression, your natural way of meeting the world and the style you lead with. It needs an accurate birth time.",
		Short:   "how you come across and meet the world", Cycle: "moves through all twelve signs every day"},
	{Key: "mc", Name: "Midheaven", Glyph: "MC", Keywords: []string{"career", "reputation", "goals"},
		Meaning: "The Midheaven is the highest point of the sky at your birth. It points to your public life: career, reputation and what you want to be known for.",
		Short:   "your career path and public image", Cycle: "moves through all twelve signs every day"},
	{Key: "dsc", Name: "Descendant", Glyph: "DC", Keywords: []string{"partners", "others", "balance"},
		Meaning: "The Descendant sits opposite the Ascendant. It describes the kind of people you are drawn to as partners and the qualities you often find through other people.",
		Short:   "what you look for in partners", Cycle: "moves through all twelve signs every day"},
	{Key: "ic", Name: "Imum Coeli", Glyph: "IC", Keywords: []string{"home", "roots", "privacy"},
		Meaning: "The IC is the lowest point of the chart, opposite the Midheaven. It describes your roots, your home life and what you need in private to feel settled.",
		Short:   "your roots and private life", Cycle: "moves through all twelve signs every day"},
	{Key: "fortune", Name: "Part of Fortune", Glyph: "⊗", Keywords: []string{"ease", "luck", "contentment"},
		Meaning: "The Part of Fortune is a calculated point that combines your Sun, Moon and Ascendant. It shows where contentment and good fortune come to you most naturally.",
		Short:   "where contentment comes easily", Cycle: "moves through all twelve signs every day"},
	{Key: "southnode", Name: "South Node", Glyph: "☋", Keywords: []string{"comfort zone", "habits", "past"},
		Meaning: "The South Node sits exactly opposite the North Node. It shows the skills and habits that already come naturally to you: useful to draw on, but easy to hide in.",
		Short:   "your comfort zone and old habits", Cycle: "about 18.6 years around the zodiac, moving backwards"},
}

// Planets returns the planets and the North Node in chart order.
func Planets() []Planet { return append([]Planet(nil), planets...) }

// Points returns the chart points (angles, Part of Fortune, South Node).
func Points() []Planet { return append([]Planet(nil), points...) }

// PlanetByKey looks up a planet, the node or a point by key.
func PlanetByKey(k string) (Planet, bool) {
	for _, p := range planets {
		if p.Key == k {
			return p, true
		}
	}
	for _, p := range points {
		if p.Key == k {
			return p, true
		}
	}
	return Planet{}, false
}

// PointByKey looks up a chart point by key.
func PointByKey(k string) (Planet, bool) {
	for _, p := range points {
		if p.Key == k {
			return p, true
		}
	}
	return Planet{}, false
}

// bodyFrag holds hand-written fragments for composing aspect, transit and
// synastry texts naturally.
type bodyFrag struct {
	your  string // "your Sun"
	theme string // noun phrase: "your sense of who you are"
	act   string // what the body does as a verb phrase: "express yourself"
	area  string // main life area for scoring: love|work|mood|energy
	label string // short noun phrase for combinations: "your heart"
	heavy int    // 1–3: how much weight a contact to/from it carries
}

var frags = map[string]bodyFrag{
	"sun":       {"your Sun", "your sense of who you are", "show who you are", "energy", "your sense of self", 3},
	"moon":      {"your Moon", "your feelings and moods", "look after your feelings", "mood", "your emotions", 3},
	"mercury":   {"your Mercury", "the way you think and talk", "say what you mean", "work", "your mind", 2},
	"venus":     {"your Venus", "your heart and what you enjoy", "enjoy yourself and connect", "love", "your heart", 2},
	"mars":      {"your Mars", "your drive and energy", "go after what you want", "energy", "your drive", 2},
	"jupiter":   {"your Jupiter", "your hopes and sense of possibility", "grow and say yes", "work", "your optimism", 1},
	"saturn":    {"your Saturn", "your sense of duty and your long-term plans", "build something lasting", "work", "your sense of duty", 1},
	"uranus":    {"your Uranus", "your need for freedom", "try something new", "energy", "your independent streak", 1},
	"neptune":   {"your Neptune", "your dreams and ideals", "imagine and let go", "mood", "your imagination", 1},
	"pluto":     {"your Pluto", "your deepest drives", "change something for good", "mood", "your inner intensity", 1},
	"node":      {"your North Node", "your direction in life", "grow in a new direction", "work", "your life direction", 1},
	"asc":       {"your Ascendant", "the way you come across", "present yourself", "energy", "your outward style", 3},
	"mc":        {"your Midheaven", "your career and reputation", "step forward at work", "work", "your ambitions", 3},
	"dsc":       {"your Descendant", "your closest partnerships", "meet others halfway", "love", "your approach to partners", 2},
	"ic":        {"your IC", "your home and private life", "settle in at home", "mood", "your private self", 2},
	"fortune":   {"your Part of Fortune", "your sense of contentment", "enjoy what you have", "mood", "your sense of contentment", 1},
	"southnode": {"your South Node", "your old habits and comfort zone", "lean on what you know", "mood", "your old habits", 1},
}

// transitVoice is how a transiting body acts in a sentence. easy, hard and
// blend are verb phrases with one %s for the natal theme ("your feelings").
type transitVoice struct {
	easy   string
	hard   string
	blend  string
	speed  int    // 1 fast (hours) … 5 very slow (years)
	length string // how long it lasts
}

var voices = map[string]transitVoice{
	"sun":     {"brings a boost of confidence to %s", "puts the spotlight on %s and can stir up a clash of wills", "turns the light on %s and makes it the focus of the day", 2, "a couple of days"},
	"moon":    {"brings an easy, comfortable mood to %s", "can unsettle %s for a few hours", "colours %s and makes your reactions louder", 1, "a few hours"},
	"mercury": {"makes it easier to talk and plan around %s", "can bring mixed messages or second thoughts about %s", "fills your head with thoughts about %s", 2, "a day or two"},
	"venus":   {"brings warmth and a little pleasure to %s", "can bring a small tension or a craving for comfort around %s", "sweetens %s and draws people toward you", 2, "a day or two"},
	"mars":    {"gives energy and courage to %s", "can bring impatience or friction around %s", "fires up %s and wants action now", 3, "a few days"},
	"jupiter": {"opens doors and brings encouragement to %s", "can tempt you to overdo things around %s", "expands %s and brings a sense of possibility", 4, "a few weeks, sometimes in waves"},
	"saturn":  {"helps you build steady, lasting progress with %s", "brings pressure or a reality check to %s", "asks you to get serious about %s", 5, "several weeks to a few months"},
	"uranus":  {"brings a welcome surprise or a fresh idea to %s", "can bring sudden change or restlessness to %s", "shakes up %s and asks for something new", 5, "months, often in several passes"},
	"neptune": {"brings inspiration and a softer mood to %s", "can make things feel unclear around %s for a while", "dissolves old edges around %s and invites you to dream", 5, "months to over a year"},
	"pluto":   {"gives depth and lasting strength to %s", "can bring power struggles or something that has to end around %s", "starts a slow, deep transformation of %s", 5, "a year or more"},
}
