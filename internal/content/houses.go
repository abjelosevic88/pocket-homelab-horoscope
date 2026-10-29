package content

// House describes one of the twelve houses (areas of life) of a chart.
type House struct {
	N        int      `json:"n"`
	Name     string   `json:"name"`
	Keywords []string `json:"keywords"`
	Summary  string   `json:"summary"`
}

var houses = []House{
	{1, "Self & appearance", []string{"identity", "body", "first impressions"},
		"The first house is about you: your body, your style and the way you start things. Planets here are easy for others to see in you."},
	{2, "Money & values", []string{"income", "possessions", "self-worth"},
		"The second house covers what you own and earn, and what you value, including how much you value yourself. It shows how you build security."},
	{3, "Communication & neighbourhood", []string{"talking", "learning", "siblings", "short trips"},
		"The third house is your everyday mind: conversations, messages, learning, siblings and the streets you walk every day."},
	{4, "Home & family", []string{"home", "family", "roots"},
		"The fourth house is home, family and your roots. It describes where you come from and what you need in private to feel settled."},
	{5, "Joy & creativity", []string{"fun", "romance", "children", "creativity"},
		"The fifth house is play: hobbies, romance, creativity, children and anything you do purely because you enjoy it."},
	{6, "Work & health", []string{"routines", "daily work", "health", "service"},
		"The sixth house is your daily routine: the work you do, the habits that keep you well, and how you help others."},
	{7, "Partnerships", []string{"marriage", "close partners", "contracts"},
		"The seventh house covers one-to-one relationships: partners, spouses, close collaborators and open rivals. It shows what you seek in another person."},
	{8, "Intimacy & shared resources", []string{"trust", "shared money", "change", "intimacy"},
		"The eighth house is about deep bonds and what you share: intimacy, trust, joint money, inheritance, and the big changes that transform you."},
	{9, "Travel & beliefs", []string{"travel", "study", "philosophy", "faith"},
		"The ninth house is the wider world: long journeys, higher learning, beliefs and the search for meaning."},
	{10, "Career & reputation", []string{"career", "status", "public life"},
		"The tenth house is your public life: career, reputation, ambitions and the mark you want to leave."},
	{11, "Friends & community", []string{"friends", "groups", "hopes"},
		"The eleventh house covers friends, groups, communities and your hopes for the future. It shows where you find your people."},
	{12, "Rest & inner life", []string{"solitude", "rest", "dreams", "the unconscious"},
		"The twelfth house is the quiet, hidden part of life: rest, solitude, dreams, intuition and the things you keep to yourself."},
}

// Houses returns the twelve houses in order.
func Houses() []House { return append([]House(nil), houses...) }

// houseArea is a short noun phrase for each house, used in composed texts.
var houseArea = [13]string{"",
	"your self-image and the way you come across",
	"money and what you value",
	"conversations, learning and your daily errands",
	"home and family",
	"fun, romance and creative projects",
	"daily work, routines and health habits",
	"your partnerships",
	"intimacy, trust and shared money",
	"travel, study and big-picture beliefs",
	"your career and reputation",
	"friends and the groups you belong to",
	"rest, solitude and your inner life",
}
