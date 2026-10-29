// The chart wheel: the zodiac around the edge, the houses, the planets at
// their degrees and the aspects between them. The Ascendant sits at nine
// o'clock (0° Aries when there is no birth time); degrees run anticlockwise.
// An optional outer ring shows a second set of planets (today's sky, or a
// partner's chart) against this one.
import { useMemo } from 'react'
import { ASPECT_COLOR, ASPECT_NATURE, BODY_GLYPH, ELEMENT_COLOR, SIGNS, SIGN_ELEMENT, SIGN_GLYPH, bodyName, fmtLon } from '../lib/astro'
import type { Chart, Position } from '../lib/types'

const VB = 440 // viewBox side; the wheel is drawn around (C, C)
const C = VB / 2
const ORD = ['', '1st', '2nd', '3rd', '4th', '5th', '6th', '7th', '8th', '9th', '10th', '11th', '12th']

interface Mark { key: string; lon: number; retro: boolean; house?: number; at: number }

const norm = (d: number) => ((d % 360) + 360) % 360

/** Spread marks apart so glyphs don't overlap; each keeps its true degree. */
function spread(items: { key: string; lon: number; retro: boolean; house?: number }[], minSep: number): Mark[] {
  const ms: Mark[] = items.map(i => ({ ...i, at: norm(i.lon) })).sort((a, b) => a.at - b.at)
  const n = ms.length
  if (n < 2) return ms
  for (let iter = 0; iter < 60; iter++) {
    let moved = false
    for (let i = 0; i < n; i++) {
      const a = ms[i], b = ms[(i + 1) % n]
      let gap = b.at - a.at
      if (i === n - 1) gap += 360
      if (gap < minSep) {
        const push = (minSep - gap) / 2 + 0.01
        a.at = norm(a.at - push)
        b.at = norm(b.at + push)
        moved = true
      }
    }
    if (!moved) break
    ms.sort((a, b) => a.at - b.at)
  }
  return ms
}

export function ChartWheel({ chart, outer, size = 420, onPick, highlight }: { chart: Chart; outer?: Position[]; size?: number; onPick?: (key: string) => void; highlight?: string }) {
  const houses = chart.time_known ? chart.houses : undefined
  const asc = houses ? houses.asc : 0
  const hasOuter = !!outer && outer.length > 0

  // radii, outside in
  const R0 = 206           // edge of the zodiac ring
  const RZ = 176           // inside of the zodiac ring
  const ROUT = 158         // outer planets (when shown)
  const RSEP = hasOuter ? 141 : RZ // the band the natal planets sit in starts here
  const RP = hasOuter ? 123 : 148  // natal planets
  const RA = hasOuter ? 86 : 100   // the aspect circle
  const RH = RA + 11               // house numbers

  // math angle (degrees, y up) of an ecliptic longitude
  const ang = (lon: number) => 180 + (lon - asc)
  const xy = (lon: number, r: number): [number, number] => {
    const t = (ang(lon) * Math.PI) / 180
    return [C + r * Math.cos(t), C - r * Math.sin(t)]
  }
  const f = (n: number) => Math.round(n * 100) / 100

  /** An annulus sector from lon a to lon b (a < b, anticlockwise). */
  const sector = (a: number, b: number, r1: number, r2: number) => {
    const [x1, y1] = xy(a, r2), [x2, y2] = xy(b, r2), [x3, y3] = xy(b, r1), [x4, y4] = xy(a, r1)
    const large = b - a > 180 ? 1 : 0
    // anticlockwise in the y-up wheel = sweep 0 in SVG's y-down space
    return `M${f(x1)},${f(y1)} A${r2},${r2} 0 ${large} 0 ${f(x2)},${f(y2)} L${f(x3)},${f(y3)} A${r1},${r1} 0 ${large} 1 ${f(x4)},${f(y4)} Z`
  }

  const natal = useMemo(() => {
    const items = chart.positions.map(p => ({ key: p.body, lon: p.lon, retro: p.retro, house: p.house }))
    if (houses) for (const pt of chart.points) if (pt.key === 'fortune') items.push({ key: 'fortune', lon: pt.lon, retro: false, house: pt.house })
    return spread(items, hasOuter ? 7.5 : 6.6)
  }, [chart, houses, hasOuter])
  const outerMarks = useMemo(() => (hasOuter ? spread(outer!.map(p => ({ key: p.body, lon: p.lon, retro: p.retro })), 6) : []), [outer, hasOuter])

  const lonOf: Record<string, number> = {}
  for (const p of chart.positions) lonOf[p.body] = p.lon
  for (const pt of chart.points) lonOf[pt.key] = pt.lon

  const aspects = useMemo(() => {
    const list = chart.aspects.filter(a => a.type !== 'conjunction' && lonOf[a.a] != null && lonOf[a.b] != null)
    const bodies = list.filter(a => !['asc', 'mc', 'dsc', 'ic', 'fortune', 'southnode'].includes(a.b) && !['asc', 'mc', 'dsc', 'ic', 'fortune', 'southnode'].includes(a.a))
    return list.length > 28 ? bodies : list.filter(a => !['dsc', 'ic', 'southnode', 'fortune'].includes(a.a) && !['dsc', 'ic', 'southnode', 'fortune'].includes(a.b))
  }, [chart]) // eslint-disable-line react-hooks/exhaustive-deps

  const tip = (key: string, lon: number, house?: number, retro?: boolean, prefix = '') =>
    `${prefix}${bodyName(key)} ${fmtLon(lon)}${house ? `, ${ORD[house]} house` : ''}${retro ? ', retrograde' : ''}`

  const planet = (m: Mark, r: number, bandOuter: number, prefix: string, small: boolean) => {
    const [gx, gy] = xy(m.at, r)
    const [t1x, t1y] = xy(m.lon, bandOuter)
    const [t2x, t2y] = xy(m.lon, bandOuter - 5)
    const [cx, cy] = xy(m.at, r + (small ? 9 : 11))
    const [dx, dy] = xy(m.at, r - (small ? 13 : 15))
    const moved = Math.abs(norm(m.at - m.lon + 180) - 180) > 0.6
    const deg = Math.floor(norm(m.lon) % 30)
    const hl = highlight === m.key
    const fs = small ? 14 : 16
    const act = onPick ? () => onPick(m.key) : undefined
    return (
      <g key={prefix + m.key} className={`cw-planet ${onPick ? 'pickable' : ''} ${hl ? 'hl' : ''} ${small ? 'outer' : ''}`}
        role={onPick ? 'button' : undefined} tabIndex={onPick ? 0 : undefined} aria-label={tip(m.key, m.lon, m.house, m.retro, prefix ? 'Outer: ' : '')}
        onClick={act} onKeyDown={act ? e => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); act() } } : undefined}>
        <title>{tip(m.key, m.lon, m.house, m.retro, prefix ? (prefix === 'o' ? 'Outer ring: ' : '') : '')}</title>
        <line className="cw-tick-p" x1={f(t1x)} y1={f(t1y)} x2={f(t2x)} y2={f(t2y)} />
        {moved && <line className="cw-lead" x1={f(t2x)} y1={f(t2y)} x2={f(cx)} y2={f(cy)} />}
        <circle className="cw-hit" cx={f(gx)} cy={f(gy)} r={small ? 10 : 12} />
        <text className="cw-glyph glyph" x={f(gx)} y={f(gy)} fontSize={fs} textAnchor="middle" dominantBaseline="central">{BODY_GLYPH[m.key] || m.key.slice(0, 2)}</text>
        <text className="cw-deg" x={f(dx)} y={f(dy)} fontSize={small ? 8 : 8.5} textAnchor="middle" dominantBaseline="central">{deg}°{m.retro ? '℞' : ''}</text>
      </g>
    )
  }

  const axes = houses ? [
    { key: 'AC', lon: houses.asc }, { key: 'DC', lon: houses.asc + 180 }, { key: 'MC', lon: houses.mc }, { key: 'IC', lon: houses.mc + 180 },
  ] : []

  return (
    <svg className="chart-wheel" viewBox={`-14 -14 ${VB + 28} ${VB + 28}`} style={{ maxWidth: size }} role="img" aria-label="Chart wheel">
      {/* zodiac ring */}
      <circle className="cw-face" cx={C} cy={C} r={R0} />
      {SIGNS.map((s, i) => {
        const [gx, gy] = xy(i * 30 + 15, (R0 + RZ) / 2)
        return (
          <g key={s}>
            <path d={sector(i * 30, i * 30 + 30, RZ, R0)} fill={ELEMENT_COLOR[SIGN_ELEMENT[s]]} className={`cw-sign ${i % 2 ? 'odd' : ''}`} />
            <text className="cw-sign-glyph glyph" x={f(gx)} y={f(gy)} fontSize={15} textAnchor="middle" dominantBaseline="central" fill={ELEMENT_COLOR[SIGN_ELEMENT[s]]}>
              <title>{s[0].toUpperCase() + s.slice(1)}</title>{SIGN_GLYPH[s]}
            </text>
          </g>
        )
      })}
      {Array.from({ length: 12 }, (_, i) => { const [x1, y1] = xy(i * 30, RZ), [x2, y2] = xy(i * 30, R0); return <line key={i} className="cw-line" x1={f(x1)} y1={f(y1)} x2={f(x2)} y2={f(y2)} /> })}
      {Array.from({ length: 72 }, (_, i) => {
        const d = i * 5
        if (d % 30 === 0) return null
        const [x1, y1] = xy(d, RZ), [x2, y2] = xy(d, RZ + (d % 10 === 0 ? 6 : 4))
        return <line key={d} className="cw-tick" x1={f(x1)} y1={f(y1)} x2={f(x2)} y2={f(y2)} />
      })}
      <circle className="cw-line" cx={C} cy={C} r={RZ} />
      {hasOuter && <circle className="cw-line soft" cx={C} cy={C} r={RSEP} />}
      <circle className="cw-inner" cx={C} cy={C} r={RA} />

      {/* houses */}
      {houses && houses.cusps.map((cu, i) => {
        const next = houses.cusps[(i + 1) % 12]
        const mid = cu + norm(next - cu) / 2
        const [x1, y1] = xy(cu, RA), [x2, y2] = xy(cu, RSEP)
        const [nx, ny] = xy(mid, RH)
        const axis = i === 0 || i === 3 || i === 6 || i === 9
        return (
          <g key={i}>
            <line className={`cw-cusp ${axis ? 'axis' : ''}`} x1={f(x1)} y1={f(y1)} x2={f(x2)} y2={f(y2)} />
            <text className="cw-house-n" x={f(nx)} y={f(ny)} fontSize={8.5} textAnchor="middle" dominantBaseline="central">{i + 1}</text>
          </g>
        )
      })}
      {axes.map(a => {
        const [x1, y1] = xy(a.lon, RZ), [x2, y2] = xy(a.lon, R0 + 4)
        const [lx, ly] = xy(a.lon, R0 + 11)
        return (
          <g key={a.key}>
            <line className="cw-axis" x1={f(x1)} y1={f(y1)} x2={f(x2)} y2={f(y2)} />
            <text className="cw-axis-label" x={f(lx)} y={f(ly)} fontSize={9} textAnchor="middle" dominantBaseline="central">{a.key}</text>
          </g>
        )
      })}

      {/* aspects */}
      {aspects.map((a, i) => {
        const [x1, y1] = xy(lonOf[a.a], RA), [x2, y2] = xy(lonOf[a.b], RA)
        const w = Math.max(0.5, 1.6 * (1 - a.orb / 9))
        const blend = ASPECT_NATURE[a.type] === 'blend'
        const dim = highlight && a.a !== highlight && a.b !== highlight
        return (
          <line key={i} className={`cw-aspect ${dim ? 'dim' : ''}`} x1={f(x1)} y1={f(y1)} x2={f(x2)} y2={f(y2)} stroke={ASPECT_COLOR[a.type] || 'var(--muted)'}
            strokeWidth={f(w)} strokeDasharray={blend ? '3 3' : undefined}>
            <title>{`${bodyName(a.a)} ${a.type} ${bodyName(a.b)}, ${a.orb.toFixed(1)}° from exact`}</title>
          </line>
        )
      })}
      {Object.entries(lonOf).filter(([k]) => chart.positions.some(p => p.body === k)).map(([k, lon]) => {
        const [x, y] = xy(lon, RA)
        return <circle key={k} className="cw-dot" cx={f(x)} cy={f(y)} r={1.6} />
      })}

      {/* planets */}
      {outerMarks.map(m => planet(m, ROUT, RZ, 'o', true))}
      {natal.map(m => planet(m, RP, RSEP, '', false))}
    </svg>
  )
}
