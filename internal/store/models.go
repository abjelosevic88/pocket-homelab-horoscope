package store

import "strings"

// User is an account on this box.
type User struct {
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"` // admin | member
	CreatedAt    string `json:"created_at"`
}

// APIToken is a personal access token (hash stored, plaintext shown once).
type APIToken struct {
	ID         int64   `json:"id"`
	UserID     int64   `json:"user_id"`
	Name       string  `json:"name"`
	Prefix     string  `json:"prefix"`
	LastUsedAt *string `json:"last_used_at"`
	CreatedAt  string  `json:"created_at"`
}

// Person is someone whose chart you keep (you included).
type Person struct {
	ID         int64   `json:"id"`
	UserID     int64   `json:"-"`
	Name       string  `json:"name"`
	Relation   string  `json:"relation"`
	IsMe       bool    `json:"is_me"`
	BirthDate  string  `json:"birth_date"`
	BirthTime  string  `json:"birth_time"`
	Place      string  `json:"place"`
	Country    string  `json:"country"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	TZ         string  `json:"tz"`
	Color      string  `json:"color"`
	Notes      string  `json:"notes"`
	Favorite   bool    `json:"favorite"`
	Source     string  `json:"source"`
	ExternalID string  `json:"external_id"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

// TimeKnown reports whether the birth time was given.
func (p Person) TimeKnown() bool { return p.BirthTime != "" }

// HasPlace reports whether a birth place (coordinates and zone) was given.
func (p Person) HasPlace() bool { return p.TZ != "" && (p.Lat != 0 || p.Lon != 0) }

// Relations a person can have to you.
var Relations = []string{"me", "partner", "family", "friend", "colleague", "other"}

// ValidRelation reports whether r is a known relation.
func ValidRelation(r string) bool {
	for _, x := range Relations {
		if x == r {
			return true
		}
	}
	return false
}

// JournalSky is the sky of a journal day, kept with the entry.
type JournalSky struct {
	MoonSign     string   `json:"moon_sign"`
	Phase        string   `json:"phase"`
	Illumination float64  `json:"illumination"`
	Transits     []string `json:"transits"`
	Retrograde   []string `json:"retrograde"`
}

// JournalEntry is one day's check-in.
type JournalEntry struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"-"`
	Date      string     `json:"date"`
	Mood      int        `json:"mood"`
	Energy    int        `json:"energy"`
	Note      string     `json:"note"`
	Tags      string     `json:"tags"`
	Sky       JournalSky `json:"sky"`
	Source    string     `json:"source"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
}

// Reading is a written reading (AI or composed).
type Reading struct {
	ID         int64  `json:"id"`
	UserID     int64  `json:"-"`
	PersonID   *int64 `json:"person_id"`
	OtherID    *int64 `json:"other_id"`
	Kind       string `json:"kind"`
	Ref        string `json:"ref"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	Model      string `json:"model"`
	Saved      bool   `json:"saved"`
	CreatedAt  string `json:"created_at"`
	PersonName string `json:"person_name"`
	OtherName  string `json:"other_name"`
}

// Settings are the box-wide values plus one person's preferences and their
// own connections. Connection modes: "core" (an app link through Pocket
// Core), "direct" (URL + token) or "off"; "" = not chosen yet.
type Settings struct {
	Title         string `json:"title"`
	SetupComplete bool   `json:"setup_complete"`

	DateFormat string `json:"date_format"`
	TimeFormat string `json:"time_format"`
	WeekStart  int    `json:"week_start"` // 1 = Monday, 7 = Sunday

	// How charts are drawn.
	Zodiac       string `json:"zodiac"`       // tropical | sidereal
	HouseSystem  string `json:"house_system"` // placidus | koch | whole | equal | porphyry
	Orbs         string `json:"orbs"`         // tight | normal | wide
	MinorAspects bool   `json:"minor_aspects"`

	// Where you live: the Moon calendar's times and today's sky.
	HomePlace   string  `json:"home_place"`
	HomeCountry string  `json:"home_country"`
	HomeLat     float64 `json:"home_lat"`
	HomeLon     float64 `json:"home_lon"`
	HomeTZ      string  `json:"home_tz"`

	// Notifications.
	DailyNotify     bool   `json:"daily_notify"`
	DailyTime       string `json:"daily_time"` // "07:30"
	NotifyMoon      bool   `json:"notify_moon"`
	NotifyRetro     bool   `json:"notify_retro"`
	NotifyTransits  bool   `json:"notify_transits"`
	NotifyBirthdays bool   `json:"notify_birthdays"`
	NotifyURL       string `json:"notify_url"`
	NotifyFormat    string `json:"notify_format"` // apprise | ntfy | json

	ContactsMode      string `json:"contacts_mode"`
	ContactsURL       string `json:"contacts_url"`
	ContactsToken     string `json:"contacts_token"`
	ContactsPublicURL string `json:"contacts_public_url"`

	NotesMode      string `json:"notes_mode"`
	NotesURL       string `json:"notes_url"`
	NotesToken     string `json:"notes_token"`
	NotesPublicURL string `json:"notes_public_url"`
	NotesFolder    string `json:"notes_folder"`

	HealthMode      string `json:"health_mode"`
	HealthURL       string `json:"health_url"`
	HealthToken     string `json:"health_token"`
	HealthPublicURL string `json:"health_public_url"`
	HealthInsights  bool   `json:"health_insights"`

	// AI for written readings and questions: Pocket AI through Core, or an
	// OpenAI-compatible API with the person's own key.
	AIProvider    string `json:"ai_provider"` // auto | pocket | openai
	AIPocketModel string `json:"ai_pocket_model"`
	OpenAIKey     string `json:"openai_api_key"`
	OpenAIModel   string `json:"openai_model"`
	OpenAIBaseURL string `json:"openai_base_url"`
	AIDaily       bool   `json:"ai_daily"` // write the day's reading with AI by itself
	AITone        string `json:"ai_tone"`  // gentle | straight | playful
}

// SplitTags turns "a, b ,c" into clean lowercase tags.
func SplitTags(s string) []string {
	var out []string
	seen := map[string]bool{}
	for _, t := range strings.Split(s, ",") {
		t = strings.ToLower(strings.TrimSpace(t))
		if t != "" && !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	return out
}

// JoinTags is the stored form of tags.
func JoinTags(s string) string { return strings.Join(SplitTags(s), ", ") }
