export interface User { id: number; email: string; name: string; role: 'admin' | 'member'; created_at: string }
export interface APIToken { id: number; user_id: number; name: string; prefix: string; last_used_at: string | null; created_at: string }

export type ConnMode = 'core' | 'direct' | 'off'
export type ConnVia = '' | 'core' | 'direct'

export type Zodiac = 'tropical' | 'sidereal'
export type HouseSystem = 'placidus' | 'koch' | 'whole' | 'equal' | 'porphyry'
export type Area = 'love' | 'work' | 'mood' | 'energy'
export type MatchArea = 'love' | 'communication' | 'trust' | 'fun' | 'growth'
export type PhaseName = 'new' | 'waxing_crescent' | 'first_quarter' | 'waxing_gibbous' | 'full' | 'waning_gibbous' | 'last_quarter' | 'waning_crescent'
export type Relation = 'me' | 'partner' | 'family' | 'friend' | 'colleague' | 'other'

export interface Settings {
  core_configured: boolean
  core_registered?: boolean
  title: string; date_format: string; time_format: string; week_start: number

  // how charts are worked out
  zodiac: Zodiac; house_system: HouseSystem; orbs: 'tight' | 'normal' | 'wide'; minor_aspects: boolean
  // where you live (the Moon calendar's times, today's sky); empty = your own birth place
  home_place: string; home_country: string; home_lat: number; home_lon: number; home_tz: string

  daily_notify: boolean; daily_time: string
  notify_moon: boolean; notify_retro: boolean; notify_transits: boolean; notify_birthdays: boolean
  notify_url: string; notify_format: string; notify_configured: boolean

  contacts_mode: ConnMode; contacts_via: ConnVia; contacts_detail: string; contacts_open_url: string; contacts_configured: boolean
  contacts_url: string; contacts_token: string; contacts_public_url: string
  notes_mode: ConnMode; notes_via: ConnVia; notes_detail: string; notes_open_url: string; notes_configured: boolean
  notes_url: string; notes_token: string; notes_public_url: string; notes_folder: string
  health_mode: ConnMode; health_via: ConnVia; health_detail: string; health_open_url: string; health_configured: boolean
  health_url: string; health_token: string; health_public_url: string; health_insights: boolean

  ai_provider: 'auto' | 'pocket' | 'openai'; ai_pocket_model: string
  openai_api_key: string; openai_model: string; openai_base_url: string
  ai_daily: boolean; ai_tone: 'gentle' | 'straight' | 'playful'

  setup_complete: boolean; base_url: string; geocoder: boolean
}

// ---- charts (internal/astro) ----

/** A body's place: ecliptic longitude 0–360 (0 = 0° Aries), sign 0–11, degree within the sign. */
export interface Position { body: string; lon: number; lat: number; speed: number; retro: boolean; sign: number; degree: number; house?: number }
export interface Point { key: string; lon: number; sign: number; degree: number; house?: number }
export interface Houses { system: HouseSystem; cusps: number[]; asc: number; mc: number; fallback?: string }
export interface Aspect { a: string; b: string; type: string; orb: number; applying: boolean; exact: number }
export interface Balance { elements: Record<string, number>; modalities: Record<string, number> }
export interface Chart {
  time: string; time_known: boolean; lat: number; lon: number; zodiac: Zodiac
  positions: Position[]; points: Point[]; houses?: Houses; aspects: Aspect[]; balance: Balance
}

export interface Person {
  id: number; name: string; relation: Relation; is_me: boolean
  birth_date: string; birth_time: string /* '' = unknown */
  place: string; country: string; lat: number; lon: number; tz: string
  color: string; notes: string; favorite: boolean; source: string; external_id: string
  created_at: string; updated_at: string
  // worked out
  sun: string; moon: string; rising: string /* '' without a birth time */
  age: number; next_birthday: string; days_to_birthday: number; chart_note?: string
}

export interface Placement {
  body: string; name: string; glyph: string; lon: number; sign: string; degree: number; where: string
  house: number; retro: boolean; short: string; sign_text: string; house_text: string
}
export interface AspectView extends Aspect { a_name: string; b_name: string; title: string; text: string; nature: 'easy' | 'hard' | 'blend' }
export interface HouseView { n: number; cusp: number; sign: string; name: string; summary: string; bodies: string[] }
export interface ChartView {
  person: Person; chart: Chart
  big_three: Placement[]; placements: Placement[]; points: Placement[]
  aspects: AspectView[]; houses: HouseView[]
  balance: { elements: Record<string, number>; modalities: Record<string, number>; element: string; modality: string; text: string }
  settings: { zodiac: Zodiac; house_system: HouseSystem; orbs: string; minor_aspects: boolean }
}

export interface Place { name: string; region: string; country: string; lat: number; lon: number; tz: string; population: number }

// ---- the day ----

export interface DayTransit {
  transit: string; natal: string; type: string; orb: number; applying: boolean
  title: string; text: string; tone: -1 | 0 | 1; areas: string[]; weight: number; strength: number
}
export interface DayScore { date: string; overall: number; scores: Record<Area, number>; tone: number; moon: string; phase: PhaseName }
export interface SkyEvent {
  kind: 'phase' | 'ingress' | 'station' | 'eclipse'; time: string; date: string
  body?: string; sign: string; phase?: string; direction?: 'retrograde' | 'direct'; eclipse?: string
  title: string; text: string; icon: string
}
export interface DayReading {
  needs_me?: boolean
  date: string; person_id: number; person_name: string
  headline: string; summary: string; tone: number; overall: number; scores: Record<Area, number>
  moon: { sign: string; degree: number; phase: PhaseName; illumination: number; age: number; house: number; text: string; phase_text: string; void_from?: string; void_to?: string }
  transits: DayTransit[]
  tips: { do: string[]; avoid: string[] }
  retrograde: { body: string; name: string; sign: string; text: string }[]
  events: SkyEvent[]
  week: DayScore[]
  reading: Reading | null
  journal: JournalEntry | null
  birthday: boolean
}

export interface TransitHit {
  transit: string; natal: string; type: string; start: string; exact: string[]; end: string
  title: string; text: string; tone: number; areas: string[]; weight: number; active: boolean
}
export interface TransitList { needs_me?: boolean; person_id: number; person_name: string; from: string; days: number; transits: TransitHit[] }

// ---- the sky ----

export interface SkyBody { body: string; name: string; glyph: string; lon: number; sign: string; degree: number; where: string; speed: number; retro: boolean; house?: number }
export interface SkyNow {
  at: string; place: string; has_home: boolean; zone: string
  chart: Chart; bodies: SkyBody[]
  moon: { sign: string; degree: number; phase: PhaseName; illumination: number; age: number; text: string; phase_text: string; void: { from?: string; to?: string } }
  events: SkyEvent[]
}
export interface MoonDay { date: string; phase: PhaseName; illumination: number; angle: number; sign: string; sign_change?: string; events: SkyEvent[]; void?: string[]; mood?: number }
export interface MoonMonth { month: string; zone: string; days: MoonDay[] }
export interface RetroPeriod { body: string; name: string; start: string; end: string; sign: string; text: string }

// ---- compatibility ----

export interface MatchAspect extends Aspect { a_name: string; b_name: string; title: string; text: string; tone: number; areas: string[]; weight: number }
export interface Match {
  a: Person; b: Person; overall: number; scores: Record<MatchArea, number>; summary: string; sun_pair: string
  elements: Record<string, [number, number]>
  aspects: MatchAspect[]; strengths: MatchAspect[]; challenges: MatchAspect[]
  reading: Reading | null
}

// ---- journal & readings ----

export interface JournalSky { moon_sign: string; phase: PhaseName | ''; illumination: number; transits: string[]; retrograde: string[] }
export interface JournalEntry { id: number; date: string; mood: number; energy: number; note: string; tags: string; sky: JournalSky; source: string; created_at: string; updated_at: string }
export interface Bucket { key: string; mood: number; energy: number; n: number; value?: number }
export interface Insights {
  from: string; days: number; entries: number; mood_avg: number; energy_avg: number
  phases: Bucket[]; elements: Bucket[]; signs: Bucket[]; findings: string[]
  health?: { ok: boolean; error?: string; not_connected?: boolean } & Record<string, unknown>
}

export type ReadingKind = 'daily' | 'natal' | 'match' | 'year' | 'question'
export interface Reading {
  id: number; person_id: number | null; other_id: number | null; kind: ReadingKind; ref: string
  title: string; body: string; model: string; saved: boolean; created_at: string; person_name: string; other_name: string
}
export interface ReadingResult { ok: boolean; error?: string; not_configured?: boolean; reading?: Reading }

export interface Dashboard {
  date: string
  moon: { sign: string; phase: PhaseName; illumination: number }
  next: SkyEvent[]
  headline?: string; scores?: Record<Area, number>; overall?: number; tone?: number
  me?: { name: string; sun: string; moon: string; rising: string }
  people: number
}
