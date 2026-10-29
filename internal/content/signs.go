// Package content holds every piece of interpretation text Pocket Horoscope
// shows: signs, planets, houses, aspects, placements, transits, synastry and
// the daily texts. It works on plain string keys ("sun", "aries", "trine")
// and has no dependency on the astronomy code.
package content

// Sign describes one zodiac sign.
type Sign struct {
	Key        string   `json:"key"`
	Name       string   `json:"name"`
	Glyph      string   `json:"glyph"`
	Element    string   `json:"element"`
	Modality   string   `json:"modality"`
	Ruler      string   `json:"ruler"`
	Dates      string   `json:"dates"`
	Keywords   []string `json:"keywords"`
	Strengths  []string `json:"strengths"`
	Challenges []string `json:"challenges"`
	Summary    string   `json:"summary"`
	Love       string   `json:"love"`
	Work       string   `json:"work"`
	Friendship string   `json:"friendship"`
	Growth     string   `json:"growth"`
}

var signs = []Sign{
	{
		Key: "aries", Name: "Aries", Glyph: "♈", Element: "fire", Modality: "cardinal", Ruler: "mars", Dates: "Mar 21 – Apr 19",
		Keywords:   []string{"courage", "action", "independence", "beginnings"},
		Strengths:  []string{"Brave", "Direct", "Energetic", "Quick to start"},
		Challenges: []string{"Impatient", "Hot-tempered", "Loses interest fast"},
		Summary:    "Aries is the first sign of the zodiac and it likes to go first. It is the spark that gets things moving: direct, honest and happiest when there is something to start or someone to stand up for. Aries energy is quick to act and quick to forgive. Its lesson is learning that patience can be a kind of strength too.",
		Love:       "In love, Aries is warm, open and easy to read. It likes to chase a little, enjoys playful competition and needs a partner who can keep up and speak their mind.",
		Work:       "Aries does best with clear goals, real responsibility and room to move fast. It is a natural starter of projects and does well when someone else enjoys finishing the details.",
		Friendship: "As a friend, Aries is loyal and protective and will be the first to suggest a plan. It may forget a birthday, but it will show up when you need help moving house.",
		Growth:     "Growth for Aries means slowing down long enough to listen, and finishing what it starts before chasing the next new thing.",
	},
	{
		Key: "taurus", Name: "Taurus", Glyph: "♉", Element: "earth", Modality: "fixed", Ruler: "venus", Dates: "Apr 20 – May 20",
		Keywords:   []string{"stability", "comfort", "patience", "the senses"},
		Strengths:  []string{"Reliable", "Patient", "Practical", "Loyal"},
		Challenges: []string{"Stubborn", "Slow to change", "Possessive"},
		Summary:    "Taurus is steady, calm and grounded. It values comfort, good food, nice things and people it can count on, and it is willing to work patiently to build a secure life. Once Taurus decides something, it rarely changes its mind. Its gift is staying power; its lesson is learning to welcome change instead of resisting it.",
		Love:       "Taurus loves through presence and touch: shared meals, time together and small physical comforts. It is slow to commit but very loyal once it does.",
		Work:       "Taurus is the colleague who quietly gets things done. It prefers steady progress, clear expectations and work that produces something real and lasting.",
		Friendship: "Taurus friends are dependable and generous hosts. They remember how you take your coffee and will stick with you for decades.",
		Growth:     "Growth for Taurus comes from trying something new before it feels completely safe, and from letting go of things it has outgrown.",
	},
	{
		Key: "gemini", Name: "Gemini", Glyph: "♊", Element: "air", Modality: "mutable", Ruler: "mercury", Dates: "May 21 – Jun 20",
		Keywords:   []string{"curiosity", "conversation", "variety", "learning"},
		Strengths:  []string{"Curious", "Quick-witted", "Adaptable", "Sociable"},
		Challenges: []string{"Restless", "Scattered", "Easily bored"},
		Summary:    "Gemini is the sign of curiosity. It wants to know a little about everything, talk to everyone and see a situation from every side. Gemini energy is quick, funny and flexible, and it keeps life interesting. Its lesson is learning to go deep as well as wide.",
		Love:       "Gemini falls for minds first. It needs conversation, laughter and a partner who keeps surprising it, and it shows love by sharing ideas, links and stories.",
		Work:       "Gemini shines in work full of people, information and variety: writing, teaching, sales, anything with many moving parts. Routine drains it quickly.",
		Friendship: "Gemini friends know everyone and connect people to each other. They are great company and always have something new to tell you.",
		Growth:     "Growth for Gemini means picking a few things to commit to and seeing them through, even after the novelty wears off.",
	},
	{
		Key: "cancer", Name: "Cancer", Glyph: "♋", Element: "water", Modality: "cardinal", Ruler: "moon", Dates: "Jun 21 – Jul 22",
		Keywords:   []string{"home", "family", "care", "feelings"},
		Strengths:  []string{"Caring", "Intuitive", "Protective", "Loyal"},
		Challenges: []string{"Moody", "Holds on too tight", "Easily hurt"},
		Summary:    "Cancer is the sign of home and belonging. It is sensitive, caring and deeply loyal to the people it calls its own. Cancer energy notices how others feel and wants everyone to feel safe and fed. Its lesson is caring for itself as well as it cares for others.",
		Love:       "Cancer loves with its whole heart and wants a relationship that feels like home. It needs reassurance and tenderness, and gives the same in return.",
		Work:       "Cancer does well in work that looks after people or builds something lasting, and it often turns a team into a kind of family.",
		Friendship: "Cancer friends remember everything you told them and check in when you are low. They are the ones who bring soup.",
		Growth:     "Growth for Cancer comes from saying what it needs directly instead of hoping others will guess, and from letting the past rest.",
	},
	{
		Key: "leo", Name: "Leo", Glyph: "♌", Element: "fire", Modality: "fixed", Ruler: "sun", Dates: "Jul 23 – Aug 22",
		Keywords:   []string{"self-expression", "warmth", "creativity", "pride"},
		Strengths:  []string{"Generous", "Warm-hearted", "Confident", "Creative"},
		Challenges: []string{"Needs attention", "Proud", "Dramatic"},
		Summary:    "Leo is warm, bright and generous, like the Sun that rules it. It loves to create, to play and to make people feel special. Leo energy wants to be seen and appreciated, and it gives its own encouragement freely. Its lesson is knowing it is worthy even when nobody is clapping.",
		Love:       "Leo loves big: grand gestures, loyalty and a lot of affection. It needs a partner who admires it openly and enjoys a bit of romance.",
		Work:       "Leo does best where it can lead, perform or create, and where good work gets noticed. It brings energy and morale to any team.",
		Friendship: "Leo friends are fun, loyal and protective. They throw the party, cheer you on and will defend you in public.",
		Growth:     "Growth for Leo means sharing the spotlight, taking honest feedback without feeling smaller, and creating for the joy of it.",
	},
	{
		Key: "virgo", Name: "Virgo", Glyph: "♍", Element: "earth", Modality: "mutable", Ruler: "mercury", Dates: "Aug 23 – Sep 22",
		Keywords:   []string{"care", "detail", "usefulness", "improvement"},
		Strengths:  []string{"Thoughtful", "Practical", "Helpful", "Precise"},
		Challenges: []string{"Self-critical", "Worries", "Perfectionist"},
		Summary:    "Virgo notices what others miss and quietly makes things better. It is practical, modest and happiest when it is useful to someone. Virgo energy likes order, healthy routines and doing a job properly. Its lesson is that good enough often is, and that it deserves the same kindness it gives.",
		Love:       "Virgo shows love through small, practical acts: fixing things, remembering details, helping out. It needs time to open up and appreciates a partner who is reliable.",
		Work:       "Virgo excels at work that needs care and accuracy, from editing to health care to systems. It is the person who makes the process actually work.",
		Friendship: "Virgo friends give good advice and real help. They may be reserved at first, but they are some of the most dependable people you will meet.",
		Growth:     "Growth for Virgo comes from easing up on self-criticism, accepting help and leaving room for mess and fun.",
	},
	{
		Key: "libra", Name: "Libra", Glyph: "♎", Element: "air", Modality: "cardinal", Ruler: "venus", Dates: "Sep 23 – Oct 22",
		Keywords:   []string{"balance", "partnership", "fairness", "beauty"},
		Strengths:  []string{"Diplomatic", "Fair-minded", "Charming", "Cooperative"},
		Challenges: []string{"Indecisive", "Avoids conflict", "People-pleasing"},
		Summary:    "Libra is the sign of relationships and fairness. It sees every side of a question, has a good eye for beauty and works hard to keep the peace. Libra energy is social, gracious and at its best in partnership. Its lesson is making decisions for itself, even when someone might be disappointed.",
		Love:       "Libra is a natural partner who enjoys romance, good manners and shared plans. It needs harmony and a relationship of equals.",
		Work:       "Libra does well in work that involves people, negotiation, design or justice. It is good at bringing different sides together.",
		Friendship: "Libra friends are easy to be around and great at bringing people together. They make every gathering a little more pleasant.",
		Growth:     "Growth for Libra means saying what it really thinks, accepting that some conflict is healthy, and finding balance within itself.",
	},
	{
		Key: "scorpio", Name: "Scorpio", Glyph: "♏", Element: "water", Modality: "fixed", Ruler: "pluto", Dates: "Oct 23 – Nov 21",
		Keywords:   []string{"depth", "intensity", "trust", "transformation"},
		Strengths:  []string{"Loyal", "Determined", "Perceptive", "Resilient"},
		Challenges: []string{"Secretive", "Jealous", "Holds grudges"},
		Summary:    "Scorpio goes deep. It wants the truth behind the surface and feels everything strongly, even when it shows little. Scorpio energy is focused, loyal and remarkably resilient: it can rebuild after almost anything. Its lesson is learning to trust and to let go.",
		Love:       "Scorpio wants closeness that is real and total. It is intensely loyal and needs honesty above all; once trust is broken it is hard to win back.",
		Work:       "Scorpio is great at research, strategy, crisis work and anything that needs focus and discretion. It does not give up easily.",
		Friendship: "Scorpio friends are few but lifelong. They keep your secrets and will tell you the truth when others will not.",
		Growth:     "Growth for Scorpio comes from letting people in, forgiving where it can, and trusting that being open is not the same as being weak.",
	},
	{
		Key: "sagittarius", Name: "Sagittarius", Glyph: "♐", Element: "fire", Modality: "mutable", Ruler: "jupiter", Dates: "Nov 22 – Dec 21",
		Keywords:   []string{"adventure", "meaning", "freedom", "optimism"},
		Strengths:  []string{"Optimistic", "Honest", "Adventurous", "Open-minded"},
		Challenges: []string{"Restless", "Tactless", "Overpromises"},
		Summary:    "Sagittarius is the explorer of the zodiac. It wants to travel, learn, ask big questions and find meaning in life. Sagittarius energy is cheerful, frank and always looking at the horizon. Its lesson is that freedom and commitment can live together.",
		Love:       "Sagittarius needs a partner who is also a fellow traveller: someone to laugh with, learn with and explore with, without feeling fenced in.",
		Work:       "Sagittarius does best in work with variety, learning and a sense of purpose: teaching, travel, publishing, anything that broadens horizons.",
		Friendship: "Sagittarius friends are great fun and refreshingly honest. They will drag you on adventures and make you laugh at yourself.",
		Growth:     "Growth for Sagittarius means following through, thinking before speaking, and finding adventure in staying as well as in going.",
	},
	{
		Key: "capricorn", Name: "Capricorn", Glyph: "♑", Element: "earth", Modality: "cardinal", Ruler: "saturn", Dates: "Dec 22 – Jan 19",
		Keywords:   []string{"ambition", "responsibility", "structure", "patience"},
		Strengths:  []string{"Disciplined", "Responsible", "Patient", "Ambitious"},
		Challenges: []string{"Too serious", "Workaholic", "Guarded"},
		Summary:    "Capricorn climbs steadily toward its goals. It is responsible, patient and practical, and it takes its commitments seriously. Capricorn energy builds things that last and often gets better with age. Its lesson is that rest and play are not a reward to be earned but part of a good life.",
		Love:       "Capricorn is careful in love and shows it through commitment and effort over time. Behind the reserve is a dry sense of humour and a deeply loyal heart.",
		Work:       "Capricorn is a natural manager and planner. It does well with long-term goals, clear structure and work that earns respect.",
		Friendship: "Capricorn friends are steady and wise. They give practical help and are still there years later.",
		Growth:     "Growth for Capricorn comes from letting itself be supported, showing its softer side and measuring success by more than achievement.",
	},
	{
		Key: "aquarius", Name: "Aquarius", Glyph: "♒", Element: "air", Modality: "fixed", Ruler: "uranus", Dates: "Jan 20 – Feb 18",
		Keywords:   []string{"originality", "community", "ideas", "freedom"},
		Strengths:  []string{"Original", "Independent", "Humanitarian", "Open-minded"},
		Challenges: []string{"Detached", "Contrary", "Stubborn about ideas"},
		Summary:    "Aquarius thinks differently. It cares about people as a group, about fairness and the future, and it is not afraid to break with tradition. Aquarius energy is friendly, inventive and independent. Its lesson is connecting with feelings, including its own, as well as with ideas.",
		Love:       "Aquarius wants a partner who is also a friend and gives it room to be itself. It shows love through loyalty and honest conversation more than grand romance.",
		Work:       "Aquarius thrives in work that is new, technical or socially useful, with freedom to try unusual approaches.",
		Friendship: "Friendship is where Aquarius shines. It has friends from every walk of life and treats them as equals.",
		Growth:     "Growth for Aquarius means letting people get close, admitting feelings out loud and being open to being wrong.",
	},
	{
		Key: "pisces", Name: "Pisces", Glyph: "♓", Element: "water", Modality: "mutable", Ruler: "neptune", Dates: "Feb 19 – Mar 20",
		Keywords:   []string{"imagination", "compassion", "intuition", "dreams"},
		Strengths:  []string{"Compassionate", "Imaginative", "Gentle", "Intuitive"},
		Challenges: []string{"Escapist", "Unclear boundaries", "Easily overwhelmed"},
		Summary:    "Pisces is the dreamer of the zodiac. It is gentle, imaginative and deeply compassionate, often sensing what others feel without being told. Pisces energy flows easily between people and moods, and it is drawn to art, music and anything that lifts life above the everyday. Its lesson is protecting its own energy and keeping its feet on the ground.",
		Love:       "Pisces loves romantically and selflessly. It needs a kind, steady partner who values its sensitivity and helps it keep healthy boundaries.",
		Work:       "Pisces does well in creative, caring or healing work, and in places where imagination matters more than rigid rules.",
		Friendship: "Pisces friends listen without judging and make you feel understood. They are generous with time and comfort.",
		Growth:     "Growth for Pisces comes from saying no when it needs to, facing practical matters directly and trusting its own worth.",
	},
}

var signIndex = func() map[string]int {
	m := make(map[string]int, len(signs))
	for i, s := range signs {
		m[s.Key] = i
	}
	return m
}()

// Signs returns the twelve signs in zodiac order (Aries first).
func Signs() []Sign { return append([]Sign(nil), signs...) }

// SignByKey looks a sign up by key ("aries").
func SignByKey(k string) (Sign, bool) {
	i, ok := signIndex[k]
	if !ok {
		return Sign{}, false
	}
	return signs[i], true
}

// SignByIndex returns the sign at zodiac index 0–11 (wrapping around).
func SignByIndex(i int) Sign {
	i %= 12
	if i < 0 {
		i += 12
	}
	return signs[i]
}

func signName(k string) string {
	if s, ok := SignByKey(k); ok {
		return s.Name
	}
	return ""
}

func signElement(k string) string {
	if s, ok := SignByKey(k); ok {
		return s.Element
	}
	return ""
}
