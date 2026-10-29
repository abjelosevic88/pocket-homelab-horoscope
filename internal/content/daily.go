package content

import (
	"sort"
	"strings"
)

var moonInSign = [12]string{
	"The Moon is in Aries today: feelings run hot and fast, and patience is in short supply. It is a great day to start something or get moving, and a poor one for waiting around. Burn off restlessness with exercise rather than arguments.",
	"The Moon is in Taurus today: the mood slows down and turns toward comfort. Good food, a tidy space and time outdoors feel especially satisfying. It is a steady day for practical work, and a bad one for rushing anyone.",
	"The Moon is in Gemini today: minds are busy and conversations multiply. It is a good day for messages, errands, learning and catching up with people. Your attention may jump around, so keep a list handy.",
	"The Moon is in Cancer today: feelings come closer to the surface and home calls. It is a lovely day for family, cooking, nesting and checking in on people you care about. Be gentle; people are more easily hurt than usual.",
	"The Moon is in Leo today: the mood turns warm, playful and a little dramatic. It is a good day to create, celebrate or treat yourself, and people respond well to generosity. A sincere compliment goes far.",
	"The Moon is in Virgo today: the mood is practical and a little particular. It is a good day for sorting, cleaning, planning and health routines. Try not to let small flaws spoil your mood or anyone else's.",
	"The Moon is in Libra today: people want harmony and good company. It is a good day for relationships, compromise, beauty and social plans. Decisions may take longer than usual, and that is fine.",
	"The Moon is in Scorpio today: feelings run deep and intense. It is a good day for honest conversations, focused work and getting to the bottom of something. Avoid testing people or holding on to small slights.",
	"The Moon is in Sagittarius today: the mood is lighter, more optimistic and restless. It is a good day for learning, planning trips, exercise and big-picture thinking. Watch for promises that are bigger than your calendar.",
	"The Moon is in Capricorn today: the mood is serious, focused and productive. It is a good day to tackle responsibilities, make plans and finish what you started. Remember to rest; not everything has to be earned.",
	"The Moon is in Aquarius today: the mood is friendly but a little detached. It is a good day for friends, groups, new ideas and doing things differently. People need a bit more space than usual.",
	"The Moon is in Pisces today: the mood turns soft, dreamy and intuitive. It is a good day for creativity, rest, music and kindness, and a harder one for detail work. Protect your energy and get some extra sleep if you can.",
}

// MoonInSign is the daily mood text for the Moon in a sign.
func MoonInSign(sign string) string {
	i, ok := signIndex[sign]
	if !ok {
		return ""
	}
	return moonInSign[i]
}

var phaseText = map[string]string{
	"new":             "It is a New Moon, the start of the monthly cycle. A good time to set an intention, start fresh or simply rest and decide what you want next.",
	"waxing_crescent": "The Moon is a waxing crescent, growing a little each night. Take the first small steps toward what you started at the New Moon.",
	"first_quarter":   "It is a First Quarter Moon, the halfway point to full. Expect a small test of your plans; push through and make decisions.",
	"waxing_gibbous":  "The Moon is waxing gibbous, almost full. Refine and adjust; this is a good time to polish what you are working on.",
	"full":            "It is a Full Moon, the peak of the cycle. Feelings and results come to light, so celebrate what has grown and let go of what is finished.",
	"waning_gibbous":  "The Moon is waning gibbous, just past full. A good time to share what you have learned and to say thank you.",
	"last_quarter":    "It is a Last Quarter Moon. Clear out, finish loose ends and let go of what is no longer working.",
	"waning_crescent": "The Moon is a waning crescent, the quiet end of the cycle. Rest, reflect and leave space before the next New Moon.",
}

// MoonPhaseText is a short practical text for a Moon phase.
func MoonPhaseText(phase string) string { return phaseText[phase] }

var retroText = map[string]string{
	"mercury": "Mercury is retrograde: it appears to move backwards for about three weeks. Communication, travel and technology can be glitchy, so double-check messages and plans. It is a great time to review, revise and reconnect with old contacts.",
	"venus":   "Venus is retrograde: about six weeks to rethink what you love and value. Old flames and old questions about money may return. It is a better time to reflect on relationships than to rush into new ones.",
	"mars":    "Mars is retrograde: for about two months, energy turns inward. Projects may stall and old frustrations resurface. Use it to rethink how you go after what you want rather than forcing things.",
	"jupiter": "Jupiter is retrograde for about four months each year. Growth turns inward: a good time to reflect on your beliefs and on what real progress looks like for you.",
	"saturn":  "Saturn is retrograde for about four and a half months each year. It is a time to review your commitments and responsibilities, and quietly strengthen what you are building.",
	"uranus":  "Uranus is retrograde for about five months each year. The urge for change works on the inside first; notice where you feel restless before you act on it.",
	"neptune": "Neptune is retrograde for about five months each year. Illusions can lift and you may see a situation more clearly. A good time for honest reflection on your dreams.",
	"pluto":   "Pluto is retrograde for about five months each year. Deep inner work comes to the surface; a good time to let go of old patterns quietly.",
}

// Retrograde explains a planet's retrograde period (mercury…pluto).
func Retrograde(planet string) string { return retroText[planet] }

// StationText describes a planet stationing ("retrograde" or "direct").
func StationText(planet, direction string) string {
	p, ok := PlanetByKey(planet)
	if !ok || retroText[planet] == "" {
		return ""
	}
	switch direction {
	case "retrograde":
		return p.Name + " turns retrograde today and seems to stand still in the sky. Its themes (" + p.Short + ") feel slower and more inward for a while, so review before you push ahead."
	case "direct":
		return p.Name + " turns direct today after its retrograde. Things linked to " + p.Short + " start moving forward again, though it takes a few days to pick up speed."
	}
	return ""
}

// ingressFocus is, per planet, how long a sign change colours things and
// what it colours.
var ingressFocus = map[string][2]string{
	"sun":     {"For the next month", "the general mood"},
	"moon":    {"For the next two or three days", "the mood"},
	"mercury": {"For the next few weeks", "how people think and talk"},
	"venus":   {"For the next few weeks", "love, taste and pleasure"},
	"mars":    {"For the next six weeks or so", "energy and drive"},
	"jupiter": {"For about a year", "growth and opportunity"},
	"saturn":  {"For the next two to three years", "responsibility and hard lessons"},
	"uranus":  {"For about seven years", "sudden change and invention"},
	"neptune": {"For over a decade", "collective dreams and ideals"},
	"pluto":   {"For many years", "deep collective change"},
	"node":    {"For about a year and a half", "the collective sense of direction"},
}

// Ingress describes a planet entering a sign.
func Ingress(planet, sign string) string {
	p, ok := PlanetByKey(planet)
	s, okS := SignByKey(sign)
	f, okF := ingressFocus[planet]
	if !ok || !okS || !okF {
		return ""
	}
	return p.Name + " enters " + s.Name + ". " + f[0] + ", the focus for " + f[1] + " shifts toward " + signQuality[signIndex[sign]] + "."
}

// Eclipse describes a solar or lunar eclipse in a sign. kind may be "solar",
// "lunar" or a more specific variant such as "solar_total" or "lunar_partial".
func Eclipse(kind, sign string) string {
	s, ok := SignByKey(sign)
	if !ok {
		return ""
	}
	q := signQuality[signIndex[sign]]
	switch {
	case strings.HasPrefix(kind, "solar"):
		return "A Solar Eclipse in " + s.Name + " is a supercharged New Moon: a door opens toward new beginnings around " + q + ". Changes started now may take months to unfold, so do not force them."
	case strings.HasPrefix(kind, "lunar"):
		return "A Lunar Eclipse in " + s.Name + " is a supercharged Full Moon: something reaches a turning point, especially around " + q + ". Let go of what has run its course."
	}
	return ""
}

// VoidOfCourse explains a void-of-course Moon.
func VoidOfCourse() string {
	return "The Moon is void of course: it has made its last major aspect before changing signs. It is a good window for routine tasks, rest and reflection, and a weaker one for starting important things."
}

// ---- tips ----

// DayTips are a few practical suggestions for the day.
type DayTips struct {
	Do    []string `json:"do"`
	Avoid []string `json:"avoid"`
}

var tipDo = map[string]map[string][]string{
	"love": {
		"high": {"Tell someone what you appreciate about them.", "Plan something nice with someone you love.", "Say yes to a social invitation."},
		"mid":  {"Send a quick message to someone you have not heard from.", "Listen more than you talk today."},
		"low":  {"Give the people around you a little extra patience.", "Keep plans with others simple and low-pressure."},
	},
	"work": {
		"high": {"Pitch the idea you have been sitting on.", "Tackle the task that needs your full focus.", "Ask for what you need at work."},
		"mid":  {"Clear a few small tasks off your list.", "Review your plans for the week."},
		"low":  {"Double-check important messages before sending.", "Stick to routine work and leave big decisions for later."},
	},
	"mood": {
		"high": {"Do something purely because you enjoy it.", "Share your good mood with someone."},
		"mid":  {"Take a short walk without your phone.", "Notice one good thing before bed."},
		"low":  {"Go easy on yourself and lower the bar for today.", "Make time for rest, a bath or an early night."},
	},
	"energy": {
		"high": {"Use the extra energy for exercise or a physical job.", "Start the thing you have been putting off."},
		"mid":  {"Pace yourself with short breaks.", "Get some fresh air in the middle of the day."},
		"low":  {"Save your energy for what really matters.", "Drink water, eat well and stretch."},
	},
}

var tipAvoid = map[string]map[string][]string{
	"love": {
		"high": {"Do not overpromise in the glow of a good mood."},
		"mid":  {"Avoid reading too much into one message."},
		"low":  {"Avoid picking fights over small things.", "Do not make relationship decisions in a bad mood."},
	},
	"work": {
		"high": {"Do not take on more than you can finish."},
		"mid":  {"Avoid multitasking on things that need care."},
		"low":  {"Avoid signing or sending anything important in a hurry.", "Do not take criticism too personally today."},
	},
	"mood": {
		"high": {"Do not let a small setback spoil a good day."},
		"mid":  {"Avoid doom-scrolling before bed."},
		"low":  {"Avoid comparing yourself to others today.", "Do not bottle things up; talk to someone you trust."},
	},
	"energy": {
		"high": {"Avoid pushing so hard that you crash tomorrow."},
		"mid":  {"Avoid long stretches without a break."},
		"low":  {"Avoid overbooking your day.", "Do not start a big project on an empty tank."},
	},
}

var elementTips = map[string][2]string{
	"fire":  {"Move your body and take initiative.", "Avoid snap reactions and impulse buys."},
	"earth": {"Get practical things done and enjoy a simple pleasure.", "Avoid being stubborn when plans change."},
	"air":   {"Talk things through and catch up with friends.", "Avoid overthinking or scattering your attention."},
	"water": {"Trust your gut and make time for people who feel safe.", "Avoid taking other people's moods on as your own."},
}

var phaseTips = map[string][2]string{
	"new":             {"Write down one intention for the coming month.", "Avoid launching big plans before you have rested."},
	"waxing_crescent": {"Take one small step toward a new goal.", "Avoid giving up on something just because it started slowly."},
	"first_quarter":   {"Make the decision you have been putting off.", "Avoid backing away at the first obstacle."},
	"waxing_gibbous":  {"Fine-tune and improve what you are working on.", "Avoid starting from scratch when a tweak will do."},
	"full":            {"Celebrate something that has gone well.", "Avoid big confrontations while feelings run high."},
	"waning_gibbous":  {"Share what you have learned or say thank you.", "Avoid holding on to credit or to grudges."},
	"last_quarter":    {"Clear out clutter, finish loose ends.", "Avoid starting new commitments for now."},
	"waning_crescent": {"Rest and leave some empty space in your day.", "Avoid pushing through exhaustion."},
}

func band(score int) string {
	switch {
	case score >= 65:
		return "high"
	case score <= 40:
		return "low"
	}
	return "mid"
}

// Tips picks two or three things to do and to avoid, from the day's area
// scores (0–100), the Moon's sign and the Moon's phase. Deterministic.
func Tips(areaScores map[string]int, moonSign string, phase string) DayTips {
	areas := make([]string, 0, len(areaScores))
	for a := range areaScores {
		if _, ok := tipDo[a]; ok {
			areas = append(areas, a)
		}
	}
	// best first, worst last; ties by name so the order never changes
	sort.Slice(areas, func(i, j int) bool {
		if areaScores[areas[i]] != areaScores[areas[j]] {
			return areaScores[areas[i]] > areaScores[areas[j]]
		}
		return areas[i] < areas[j]
	})
	seed := hash(moonSign, phase)
	for _, a := range areas {
		seed += int64(areaScores[a]) * int64(len(a))
	}
	var t DayTips
	add := func(dst *[]string, s string) {
		if s == "" {
			return
		}
		for _, x := range *dst {
			if x == s {
				return
			}
		}
		*dst = append(*dst, s)
	}
	if len(areas) > 0 {
		best, worst := areas[0], areas[len(areas)-1]
		add(&t.Do, pick(tipDo[best][band(areaScores[best])], seed))
		add(&t.Avoid, pick(tipAvoid[worst][band(areaScores[worst])], seed))
		if len(areas) > 1 {
			add(&t.Do, pick(tipDo[worst][band(areaScores[worst])], seed+1))
		}
	}
	if et, ok := elementTips[signElement(moonSign)]; ok {
		add(&t.Do, et[0])
		add(&t.Avoid, et[1])
	}
	if pt, ok := phaseTips[phase]; ok {
		if len(t.Do) < 3 {
			add(&t.Do, pt[0])
		}
		add(&t.Avoid, pt[1])
	}
	if len(t.Do) > 3 {
		t.Do = t.Do[:3]
	}
	if len(t.Avoid) > 3 {
		t.Avoid = t.Avoid[:3]
	}
	return t
}

// ---- headlines ----

var headlines = map[int]map[string][]string{
	1: {
		"love":   {"A warm day for the heart", "Connection comes easily today", "Good company, good feelings"},
		"work":   {"Your plans have the wind behind them", "A productive, clear-headed day", "A good day to make progress"},
		"mood":   {"A lighter, brighter mood", "Room to breathe and enjoy", "An easy, good-natured day"},
		"energy": {"Energy is on your side", "A day to get moving", "Plenty of fuel in the tank"},
	},
	0: {
		"love":   {"Mixed signals, open heart", "Give and take in your relationships", "A day for honest conversations"},
		"work":   {"Steady work, small wins", "Keep it simple and keep going", "A day for sorting and planning"},
		"mood":   {"Ups and downs, nothing you cannot handle", "Check in with yourself today", "A reflective kind of day"},
		"energy": {"Pace yourself", "Energy comes in waves today", "Balance effort with rest"},
	},
	-1: {
		"love":   {"Handle feelings with care", "Patience pays off with people today", "Go gently with the ones you love"},
		"work":   {"Slow down and double-check", "Not the day to force things", "Expect a few bumps at work"},
		"mood":   {"Be kind to yourself today", "A tender day; lower the bar", "Feelings run close to the surface"},
		"energy": {"Conserve your energy", "Rest is productive today", "A day to recharge, not to push"},
	},
}

// Headline returns a short daily headline for a tone (-1, 0, 1) and the
// day's most important area. Deterministic per seed.
func Headline(tone int, topArea string, seed int64) string {
	if tone > 1 {
		tone = 1
	}
	if tone < -1 {
		tone = -1
	}
	pool := headlines[tone][topArea]
	if len(pool) == 0 {
		pool = headlines[tone]["mood"]
	}
	return pick(pool, seed)
}

// ---- sun-sign compatibility ----

// elementPair holds hand-written texts per unordered element pair.
var elementPair = map[string]string{
	"fire|fire":   "Two fire signs make a lively, passionate match full of energy and laughter. You inspire each other, but you may both want to lead, so take turns.",
	"earth|earth": "Two earth signs build something solid together. You share practical values and a love of comfort; just make sure routine leaves room for romance.",
	"air|air":     "Two air signs never run out of things to talk about. You give each other freedom and ideas; remember to share feelings as well as thoughts.",
	"water|water": "Two water signs understand each other's feelings almost without words. The bond can be deep and tender; keep some outside air so moods do not swirl.",
	"air|fire":    "Fire and air feed each other: air brings ideas and fire brings action. This is usually an exciting, social and easy-going pairing.",
	"earth|water": "Earth and water nourish each other: water brings feeling and earth brings stability. This tends to be a caring, loyal and lasting match.",
	"earth|fire":  "Fire and earth move at different speeds: one wants to leap, the other to plan. It works well when you respect each other's pace and combine vision with follow-through.",
	"fire|water":  "Fire and water are an intense mix: passion meets deep feeling. It can be magnetic, but one may feel the other is too much; kindness and patience keep the steam pleasant.",
	"air|earth":   "Air and earth think differently: one lives in ideas, the other in facts. You can learn a lot from each other if you value both the plan and the dream.",
	"air|water":   "Air and water meet head and heart. One wants to talk it through, the other to feel it through; patience turns that difference into real understanding.",
}

// Compatibility is a short sun-sign pair blurb based on how their elements
// relate. Order of the two signs does not matter.
func Compatibility(sign1, sign2 string) string {
	a, okA := SignByKey(sign1)
	b, okB := SignByKey(sign2)
	if !okA || !okB {
		return ""
	}
	e1, e2 := a.Element, b.Element
	if e2 < e1 {
		e1, e2 = e2, e1
	}
	body := elementPair[e1+"|"+e2]
	var intro string
	switch d := (signIndex[sign2] - signIndex[sign1] + 12) % 12; {
	case d == 0:
		intro = "Two " + a.Name + " suns recognise each other instantly, strengths and quirks alike."
	case d == 6:
		intro = a.Name + " and " + b.Name + " sit opposite each other in the zodiac, so each has what the other is missing."
	default:
		intro = a.Name + " is " + article(a.Element) + " " + a.Element + " sign and " + b.Name + " is " + article(b.Element) + " " + b.Element + " sign."
		if a.Element == b.Element {
			intro = a.Name + " and " + b.Name + " are both " + a.Element + " signs."
		}
	}
	return intro + " " + body
}

func article(word string) string {
	if word != "" && strings.ContainsRune("aeiou", rune(word[0])) {
		return "an"
	}
	return "a"
}
