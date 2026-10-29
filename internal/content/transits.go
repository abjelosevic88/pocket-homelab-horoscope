package content

import (
	"fmt"
	"strings"
)

// TransitText is a ready-to-show interpretation of a transit (a planet in
// today's sky touching a point of the birth chart) or a synastry contact.
type TransitText struct {
	Title  string   `json:"title"`
	Text   string   `json:"text"`
	Tone   int      `json:"tone"`
	Areas  []string `json:"areas"`
	Weight int      `json:"weight"`
}

// aspectVerb is the relation in a sentence: "Venus is trining your Sun" reads
// badly, so these are plain-language verbs.
var aspectVerb = map[string]string{
	"conjunction": "is passing over",
	"sextile":     "is making a helpful link to",
	"trine":       "is in easy harmony with",
	"square":      "is at a tense angle to",
	"opposition":  "is opposite",
	"quincunx":    "is at an awkward angle to",
}

// toneOf is the default tone of a transiting body's conjunction.
var conjTone = map[string]int{"sun": 1, "venus": 1, "jupiter": 1, "saturn": -1}

// soft bodies take the edge off hard aspects.
var softBody = map[string]bool{"venus": true, "jupiter": true}

var transitClose = map[string][]string{
	"easy": {
		"It is a good moment to %s.",
		"Use the easy flow to %s.",
		"Take the opening: a small step to %s goes a long way now.",
	},
	"hard": {
		"Slow down before you react; the tension points at something worth adjusting.",
		"Be patient with yourself and others; pushing harder rarely helps under this kind of pressure.",
		"Name what feels off, then pick one small thing you can actually change.",
	},
	"blend": {
		"Notice where your attention goes; this is a moment to %s with intention.",
		"Whatever starts now tends to carry the flavour of this meeting, so make it count.",
	},
	"quincunx": {
		"Small adjustments work better than big decisions while this lasts.",
		"Stay flexible; the fix is usually a tweak, not an overhaul.",
	},
}

// Transit describes a transiting body (sun…pluto) in an aspect to a natal
// body or point. Unknown keys return a zero TransitText.
func Transit(transiting, aspect, natal string) TransitText {
	v, okV := voices[transiting]
	f, okF := frags[natal]
	verb, okA := aspectVerb[aspect]
	if !okV || !okF || !okA {
		return TransitText{}
	}
	tp, _ := PlanetByKey(transiting)
	np, _ := PlanetByKey(natal)
	ai, _ := AspectByKey(aspect)
	nature := ai.Nature

	var effect string
	tone := 0
	switch nature {
	case "easy":
		effect, tone = fmt.Sprintf(v.easy, f.theme), 1
	case "hard":
		effect, tone = fmt.Sprintf(v.hard, f.theme), -1
		if softBody[transiting] {
			tone = 0
		}
	default:
		effect, tone = fmt.Sprintf(v.blend, f.theme), conjTone[transiting]
	}
	subject := tp.Name
	if transiting == "sun" || transiting == "moon" {
		subject = "The " + tp.Name
	}
	s1 := subject + " " + verb + " " + f.your + ". It " + effect + "."

	pool := transitClose[nature]
	if aspect == "quincunx" {
		pool = transitClose["quincunx"]
	}
	s2 := pick(pool, hash(transiting, aspect, natal))
	if strings.Contains(s2, "%s") {
		s2 = fmt.Sprintf(s2, f.act)
	}

	s3 := "This lasts " + v.length + "."
	if v.speed >= 4 {
		s3 = "This is a slow influence that lasts " + v.length + ", so it is worth paying attention to."
	}

	weight := v.speed
	if f.heavy == 3 && v.speed >= 2 {
		weight++
	}
	if weight > 5 {
		weight = 5
	}
	if weight < 1 {
		weight = 1
	}

	return TransitText{
		Title:  tp.Name + " " + strings.ToLower(ai.Name) + " your " + np.Name,
		Text:   join(s1, s2, s3),
		Tone:   tone,
		Areas:  areasOf(transiting, natal),
		Weight: weight,
	}
}

func areasOf(keys ...string) []string {
	seen := map[string]bool{}
	var out []string
	for _, k := range keys {
		a := frags[k].area
		if k == "venus" || k == "dsc" {
			a = "love"
		}
		if a != "" && !seen[a] {
			seen[a] = true
			out = append(out, a)
		}
	}
	return out
}

// ---- synastry ----

var synLink = map[string]string{
	"conjunction": "meet head-on and blend together",
	"sextile":     "get along well and spark each other",
	"trine":       "flow together easily and naturally",
	"square":      "rub against each other",
	"opposition":  "pull in opposite directions yet attract",
	"quincunx":    "need patience to understand each other",
}

// synPair holds hand-written openers for the classic contacts (keys sorted).
var synPair = map[string]string{
	"moon|sun":        "Sun–Moon contacts are one of the classic signs of a strong bond: one person's core self meets the other's emotional needs.",
	"mars|venus":      "Venus–Mars contacts are about attraction and chemistry, the pull between wanting and being wanted.",
	"moon|moon":       "Moon contacts show how easily you feel at home together and read each other's moods.",
	"mercury|mercury": "Mercury contacts show how easily you talk, joke and understand each other.",
	"sun|sun":         "Sun contacts show how your basic natures and styles fit together.",
	"venus|venus":     "Venus contacts show whether you enjoy the same things and show affection in the same way.",
	"mars|mars":       "Mars contacts show how your energies combine, including how you handle conflict.",
	"moon|venus":      "Moon–Venus contacts bring tenderness and a feeling of being cared for.",
	"sun|venus":       "Sun–Venus contacts bring affection, admiration and simply liking each other.",
	"jupiter|venus":   "Venus–Jupiter contacts bring generosity, fun and good will between you.",
	"jupiter|sun":     "Sun–Jupiter contacts bring encouragement: you tend to bring out the best in each other.",
	"mercury|moon":    "Moon–Mercury contacts help you talk about feelings without getting lost in them.",
	"mars|sun":        "Sun–Mars contacts bring energy and drive, and sometimes competition.",
	"mars|moon":       "Moon–Mars contacts bring passion and protectiveness, and sometimes quick tempers.",
	"asc|venus":       "Venus on the Ascendant is a classic sign of finding each other attractive.",
	"asc|sun":         "Sun–Ascendant contacts mean you notice and recognise each other easily.",
	"asc|moon":        "Moon–Ascendant contacts bring an easy, familiar feeling, as if you have met before.",
	"dsc|sun":         "A Sun on the other person's Descendant is a classic partnership signature.",
	"dsc|venus":       "Venus on the other person's Descendant often feels like the partner they had in mind.",
	"node|sun":        "North Node contacts often feel meant to be, as if you help each other grow.",
	"moon|node":       "Moon–Node contacts bring a deep, familiar emotional connection.",
	"node|venus":      "Venus–Node contacts often bring a sense of meeting someone important.",
	"mercury|sun":     "Sun–Mercury contacts make it easy to share ideas and understand each other.",
	"mercury|venus":   "Mercury–Venus contacts bring sweet words and enjoyable conversation.",
	"mars|mercury":    "Mercury–Mars contacts bring lively debate and quick decisions, sometimes sharp words.",
	"jupiter|moon":    "Moon–Jupiter contacts bring warmth, generosity and emotional support.",
	"mc|sun":          "Sun–Midheaven contacts mean one of you can support the other's ambitions.",
	"jupiter|jupiter": "Jupiter contacts show whether you share hopes, beliefs and a sense of humour.",
	"saturn|saturn":   "Saturn contacts show whether your sense of duty and timing line up.",
	"jupiter|mercury": "Mercury–Jupiter contacts bring big conversations and shared plans.",
	"pluto|sun":       "Sun–Pluto contacts are intense and life-changing, for better and sometimes for worse.",
	"pluto|venus":     "Venus–Pluto contacts bring deep, magnetic attraction that is hard to walk away from.",
	"neptune|venus":   "Venus–Neptune contacts bring romance and a sense of the magical.",
	"uranus|venus":    "Venus–Uranus contacts bring excitement and sudden attraction, with a need for freedom.",
	"moon|saturn":     "Moon–Saturn contacts bring loyalty and staying power, though they can feel heavy at times.",
	"saturn|sun":      "Sun–Saturn contacts add commitment and seriousness; they can feel like glue or like weight.",
	"saturn|venus":    "Venus–Saturn contacts are a classic sign of long-term commitment, steady rather than giddy.",
	"mars|pluto":      "Mars–Pluto contacts bring powerful energy that needs a shared goal.",
	"mars|uranus":     "Mars–Uranus contacts are exciting and unpredictable.",
	"neptune|sun":     "Sun–Neptune contacts bring inspiration and idealism; keep your eyes open.",
	"moon|neptune":    "Moon–Neptune contacts bring deep empathy and sometimes confusion about needs.",
	"moon|pluto":      "Moon–Pluto contacts bring emotional intensity and a strong bond.",
	"moon|uranus":     "Moon–Uranus contacts bring excitement but can make emotional security tricky.",
	"sun|uranus":      "Sun–Uranus contacts are stimulating: you wake each other up.",
	"asc|mars":        "Mars–Ascendant contacts bring physical attraction and energy.",
	"mars|saturn":     "Mars–Saturn contacts can be frustrating unless you learn each other's pace.",
	"mercury|saturn":  "Mercury–Saturn contacts bring serious, practical conversations.",
	"mercury|neptune": "Mercury–Neptune contacts bring imaginative talk and a risk of misunderstandings.",
	"mercury|uranus":  "Mercury–Uranus contacts bring exciting, surprising conversations.",
	"jupiter|saturn":  "Jupiter–Saturn contacts balance optimism with realism, which is good for shared plans.",
	"jupiter|mars":    "Mars–Jupiter contacts bring enthusiasm for doing things together.",
	"mars|neptune":    "Mars–Neptune contacts can be romantic, but motives may be unclear.",
	"jupiter|pluto":   "Jupiter–Pluto contacts bring big shared ambitions.",
}

// synCategory groups a contact into a life area.
func synCategory(a, b string) string {
	has := func(k string) bool { return a == k || b == k }
	switch {
	case has("venus") || has("mars") || has("dsc"):
		return "love"
	case has("mercury"):
		return "communication"
	case has("saturn") || has("pluto") || (has("moon") && !has("jupiter")):
		return "trust"
	case has("jupiter") || has("sun") && has("asc"):
		return "fun"
	default:
		return "growth"
	}
}

var synClose = map[string]map[string]string{
	"easy": {
		"love":          "This is a sweet, natural part of your connection; enjoy it and do not take it for granted.",
		"communication": "Conversation is likely to be one of the easy pleasures between you.",
		"trust":         "This gives the relationship a steady base you can lean on.",
		"fun":           "This makes time together light and enjoyable.",
		"growth":        "You quietly help each other become more yourselves.",
	},
	"hard": {
		"love":          "The spark is real, and so is the friction; talking openly keeps it from turning into point-scoring.",
		"communication": "Misunderstandings can happen here, so check what the other person meant before reacting.",
		"trust":         "This asks for patience and clear agreements, but it can make the bond much stronger over time.",
		"fun":           "You may want different things from your free time; take turns choosing.",
		"growth":        "It can be uncomfortable, but this is where you learn the most from each other.",
	},
	"blend": {
		"love":          "This is a strong, noticeable part of your connection that both of you will feel.",
		"communication": "You think alike in some ways, which makes talking feel natural.",
		"trust":         "This ties you together strongly; handle it with care.",
		"fun":           "You bring out each other's enthusiasm.",
		"growth":        "This contact is one of the reasons you met.",
	},
}

// Synastry describes a contact between person 1's body a and person 2's body
// b. Areas are love, communication, trust, fun or growth.
func Synastry(a, b, aspect string) TransitText {
	fa, okA := frags[a]
	fb, okB := frags[b]
	link, okL := synLink[aspect]
	if !okA || !okB || !okL {
		return TransitText{}
	}
	pa, _ := PlanetByKey(a)
	pb, _ := PlanetByKey(b)
	ai, _ := AspectByKey(aspect)
	nature := ai.Nature

	k := a + "|" + b
	if b < a {
		k = b + "|" + a
	}
	opener := synPair[k]
	if opener == "" {
		opener = "Your " + pa.Name + " and their " + pb.Name + " form " + article(ai.Key) + " " + ai.Key + "."
	}
	middle := capitalize(fa.label + " and their " + strings.TrimPrefix(fb.label, "your ") + " " + link + ".")
	cat := synCategory(a, b)
	closing := synClose[nature][cat]

	tone := 0
	switch nature {
	case "easy":
		tone = 1
	case "hard":
		tone = -1
		if softBody[a] && softBody[b] {
			tone = 0
		}
	default:
		if softBody[a] || softBody[b] || a == "sun" && b == "moon" || a == "moon" && b == "sun" {
			tone = 1
		}
		if a == "saturn" || b == "saturn" {
			tone = 0
		}
	}
	weight := fa.heavy + fb.heavy - 1
	if weight > 5 {
		weight = 5
	}
	return TransitText{
		Title:  "Your " + pa.Name + " " + strings.ToLower(ai.Name) + " their " + pb.Name,
		Text:   join(opener, middle, closing),
		Tone:   tone,
		Areas:  []string{cat},
		Weight: weight,
	}
}
