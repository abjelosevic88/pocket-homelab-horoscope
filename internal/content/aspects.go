package content

// AspectInfo describes an aspect: a meaningful angle between two planets.
type AspectInfo struct {
	Key     string  `json:"key"`
	Name    string  `json:"name"`
	Glyph   string  `json:"glyph"`
	Angle   float64 `json:"angle"`
	Nature  string  `json:"nature"`
	Meaning string  `json:"meaning"`
	Short   string  `json:"short"`
}

var aspects = []AspectInfo{
	{"conjunction", "Conjunction", "☌", 0, "blend",
		"A conjunction is when two planets sit together in the same part of the sky. Their energies merge and act as one, which makes the combination strong and hard to separate.",
		"work as one"},
	{"sextile", "Sextile", "⚹", 60, "easy",
		"A sextile is a 60° angle. It is a friendly, helpful link: an opportunity that works well when you take a small step to use it.",
		"help each other"},
	{"square", "Square", "□", 90, "hard",
		"A square is a 90° angle. The two planets pull in different directions, which creates tension, but also the push to grow and get things done.",
		"pull against each other"},
	{"trine", "Trine", "△", 120, "easy",
		"A trine is a 120° angle. It is the easiest, most natural link: the two planets support each other and talents come without much effort.",
		"flow together"},
	{"opposition", "Opposition", "☍", 180, "hard",
		"An opposition is when two planets face each other across the chart. It is a see-saw: you swing between the two, or meet one side through other people, until you find balance.",
		"face each other"},
	{"quincunx", "Quincunx", "⚻", 150, "hard",
		"A quincunx (or inconjunct) is a 150° angle. The two planets have little in common, so it feels like an awkward fit that needs regular small adjustments.",
		"need constant adjusting"},
}

// Aspects returns the aspects the app uses.
func Aspects() []AspectInfo { return append([]AspectInfo(nil), aspects...) }

// AspectByKey looks an aspect up by key.
func AspectByKey(k string) (AspectInfo, bool) {
	for _, a := range aspects {
		if a.Key == k {
			return a, true
		}
	}
	return AspectInfo{}, false
}

func aspectNature(k string) string {
	a, _ := AspectByKey(k)
	return a.Nature
}

// natalLink says, for each aspect, how the two themes relate in a birth chart.
var natalLink = map[string]string{
	"conjunction": "are fused together in you, so one rarely shows up without the other",
	"sextile":     "support each other easily once you make a small effort",
	"trine":       "work together naturally, often so smoothly you barely notice",
	"square":      "pull against each other, which can feel like an inner tug-of-war",
	"opposition":  "sit on opposite ends of a see-saw, and you may swing between them",
	"quincunx":    "speak different languages, so they need regular small adjustments",
}

// natalAdvice is the closing sentence per aspect nature.
var natalAdvice = map[string]string{
	"conjunction": "Used well, this is a concentrated strength; the work is making sure it does not run the whole show.",
	"sextile":     "Lean into it on purpose: these are skills that grow quickly when you practise them.",
	"trine":       "It is a real gift, though easy to take for granted, so put it to conscious use.",
	"square":      "The friction is uncomfortable but productive: much of your drive and growth comes from working it out.",
	"opposition":  "Other people often mirror this back to you; the goal is to honour both sides instead of picking one.",
	"quincunx":    "Small, regular tweaks work better here than big decisions; flexibility is the skill this builds.",
}

// NatalAspect describes an aspect between two bodies or points in a birth chart.
func NatalAspect(a, b, aspect string) string {
	fa, okA := frags[a]
	fb, okB := frags[b]
	link, okL := natalLink[aspect]
	if !okA || !okB || !okL || a == b {
		return ""
	}
	pa, _ := PlanetByKey(a)
	pb, _ := PlanetByKey(b)
	ai, _ := AspectByKey(aspect)
	first := "Your " + pa.Name + " and " + pb.Name + " form " + article(ai.Key) + " " + ai.Key + "."
	second := capitalize(fa.label) + " and " + fb.label + " " + link + "."
	return first + " " + second + " " + natalAdvice[aspect]
}
