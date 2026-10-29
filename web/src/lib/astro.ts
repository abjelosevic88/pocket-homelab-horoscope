// Names, glyphs and colours for signs, planets, aspects and Moon phases.
export const SIGNS = ['aries', 'taurus', 'gemini', 'cancer', 'leo', 'virgo', 'libra', 'scorpio', 'sagittarius', 'capricorn', 'aquarius', 'pisces'] as const
export const SIGN_NAME: Record<string, string> = { aries: 'Aries', taurus: 'Taurus', gemini: 'Gemini', cancer: 'Cancer', leo: 'Leo', virgo: 'Virgo', libra: 'Libra', scorpio: 'Scorpio', sagittarius: 'Sagittarius', capricorn: 'Capricorn', aquarius: 'Aquarius', pisces: 'Pisces' }
// U+FE0E keeps the glyphs as text, not emoji
export const SIGN_GLYPH: Record<string, string> = { aries: '♈︎', taurus: '♉︎', gemini: '♊︎', cancer: '♋︎', leo: '♌︎', virgo: '♍︎', libra: '♎︎', scorpio: '♏︎', sagittarius: '♐︎', capricorn: '♑︎', aquarius: '♒︎', pisces: '♓︎' }
export const SIGN_ELEMENT: Record<string, Element> = { aries: 'fire', taurus: 'earth', gemini: 'air', cancer: 'water', leo: 'fire', virgo: 'earth', libra: 'air', scorpio: 'water', sagittarius: 'fire', capricorn: 'earth', aquarius: 'air', pisces: 'water' }
export const SIGN_DATES: Record<string, string> = { aries: 'Mar 21 – Apr 19', taurus: 'Apr 20 – May 20', gemini: 'May 21 – Jun 20', cancer: 'Jun 21 – Jul 22', leo: 'Jul 23 – Aug 22', virgo: 'Aug 23 – Sep 22', libra: 'Sep 23 – Oct 22', scorpio: 'Oct 23 – Nov 21', sagittarius: 'Nov 22 – Dec 21', capricorn: 'Dec 22 – Jan 19', aquarius: 'Jan 20 – Feb 18', pisces: 'Feb 19 – Mar 20' }

export type Element = 'fire' | 'earth' | 'air' | 'water'
export const ELEMENTS: Element[] = ['fire', 'earth', 'air', 'water']
/** Element colours (light / dark friendly mid tones). */
export const ELEMENT_COLOR: Record<string, string> = { fire: '#e8590c', earth: '#5c940d', air: '#1c7ed6', water: '#7048e8' }
export const MODALITY_NAME: Record<string, string> = { cardinal: 'Cardinal', fixed: 'Fixed', mutable: 'Mutable' }

export const BODIES = ['sun', 'moon', 'mercury', 'venus', 'mars', 'jupiter', 'saturn', 'uranus', 'neptune', 'pluto', 'node'] as const
export const BODY_NAME: Record<string, string> = {
  sun: 'Sun', moon: 'Moon', mercury: 'Mercury', venus: 'Venus', mars: 'Mars', jupiter: 'Jupiter', saturn: 'Saturn', uranus: 'Uranus', neptune: 'Neptune', pluto: 'Pluto',
  node: 'North Node', southnode: 'South Node', asc: 'Ascendant', mc: 'Midheaven', dsc: 'Descendant', ic: 'IC', fortune: 'Part of Fortune',
}
export const BODY_GLYPH: Record<string, string> = {
  sun: '☉︎', moon: '☽︎', mercury: '☿︎', venus: '♀︎', mars: '♂︎', jupiter: '♃︎', saturn: '♄︎', uranus: '♅︎', neptune: '♆︎', pluto: '♇︎',
  node: '☊︎', southnode: '☋︎', asc: 'AC', mc: 'MC', dsc: 'DC', ic: 'IC', fortune: '⊗︎',
}

export const ASPECT_NAME: Record<string, string> = { conjunction: 'Conjunction', sextile: 'Sextile', square: 'Square', trine: 'Trine', opposition: 'Opposition', quincunx: 'Quincunx' }
export const ASPECT_GLYPH: Record<string, string> = { conjunction: '☌︎', sextile: '⚹︎', square: '□︎', trine: '△︎', opposition: '☍︎', quincunx: '⚻︎' }
/** easy = blue/green, hard = red, blend = accent. */
export const ASPECT_COLOR: Record<string, string> = { conjunction: 'var(--accent)', sextile: '#0ca678', trine: '#1c7ed6', square: '#e03131', opposition: '#e8590c', quincunx: '#868e96' }
export const ASPECT_NATURE: Record<string, 'easy' | 'hard' | 'blend'> = { conjunction: 'blend', sextile: 'easy', trine: 'easy', square: 'hard', opposition: 'hard', quincunx: 'blend' }

export const PHASE_NAME: Record<string, string> = { new: 'New Moon', waxing_crescent: 'Waxing crescent', first_quarter: 'First quarter', waxing_gibbous: 'Waxing gibbous', full: 'Full Moon', waning_gibbous: 'Waning gibbous', last_quarter: 'Last quarter', waning_crescent: 'Waning crescent' }
export const PHASE_EMOJI: Record<string, string> = { new: '🌑', waxing_crescent: '🌒', first_quarter: '🌓', waxing_gibbous: '🌔', full: '🌕', waning_gibbous: '🌖', last_quarter: '🌗', waning_crescent: '🌘' }
export const PHASES = ['new', 'waxing_crescent', 'first_quarter', 'waxing_gibbous', 'full', 'waning_gibbous', 'last_quarter', 'waning_crescent'] as const

export const AREA_NAME: Record<string, string> = { love: 'Love', work: 'Work', mood: 'Mood', energy: 'Energy', communication: 'Talking', trust: 'Trust', fun: 'Fun', growth: 'Growth' }
export const AREA_ICON: Record<string, string> = { love: 'heart', work: 'target', mood: 'smile', energy: 'zap', communication: 'mail', trust: 'shield', fun: 'sparkles', growth: 'trending-up' }
export const RELATION_NAME: Record<string, string> = { me: 'You', partner: 'Partner', family: 'Family', friend: 'Friend', colleague: 'Colleague', other: 'Other' }

export const signName = (k?: string) => (k ? SIGN_NAME[k] || k : '')
export const bodyName = (k?: string) => (k ? BODY_NAME[k] || k : '')
export const signIndex = (k: string) => SIGNS.indexOf(k as (typeof SIGNS)[number])

/** 14°32′ — the degree within a sign. */
export function fmtDeg(deg: number): string {
  let d = Math.floor(deg)
  let m = Math.round((deg - d) * 60)
  if (m === 60) { d += 1; m = 0 }
  return `${d}°${String(m).padStart(2, '0')}′`
}

/** 14°32′ Sagittarius from an ecliptic longitude. */
export function fmtLon(lon: number, glyph = false): string {
  const l = ((lon % 360) + 360) % 360
  const s = SIGNS[Math.floor(l / 30)]
  return `${fmtDeg(l - Math.floor(l / 30) * 30)} ${glyph ? SIGN_GLYPH[s] : SIGN_NAME[s]}`
}

/** A 0–100 score as a word. */
export function scoreWord(n: number): string {
  if (n >= 80) return 'Excellent'
  if (n >= 65) return 'Good'
  if (n >= 45) return 'Steady'
  if (n >= 30) return 'Mixed'
  return 'Tough'
}

/** The colour of a 0–100 score. */
export function scoreColor(n: number): string {
  if (n >= 65) return 'var(--success)'
  if (n >= 45) return 'var(--accent)'
  if (n >= 30) return 'var(--warning)'
  return 'var(--danger)'
}

/** Sky times come in the home zone ("2026-11-05T14:19:00+01:00"): show that wall clock, not the browser's. */
export const clockOf = (iso: string) => iso.slice(11, 16)
export const dateOf = (iso: string) => iso.slice(0, 10)
