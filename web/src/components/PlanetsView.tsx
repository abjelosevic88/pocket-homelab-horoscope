// The Planets tab: a zodiac strip with every planet where it stands, then
// one illustrated card per planet (sign, degree, house, a short reading that
// opens to the full text), grouped from the personal planets outwards.
import { useEffect, useId, useRef, useState, type CSSProperties, type ReactNode } from 'react'
import { BODY_GLYPH, ELEMENT_COLOR, SIGNS, SIGN_ELEMENT, SIGN_GLYPH, fmtDeg, signName } from '../lib/astro'
import type { Placement } from '../lib/types'
import { Icon } from './icons'

const ORD = ['', '1st', '2nd', '3rd', '4th', '5th', '6th', '7th', '8th', '9th', '10th', '11th', '12th']

const GROUPS: { title: string; sub: string; bodies: string[] }[] = [
  { title: 'Personal planets', sub: 'Who you are day to day: your core, your feelings, how you think, love and act.', bodies: ['sun', 'moon', 'mercury', 'venus', 'mars'] },
  { title: 'Social planets', sub: 'How you grow and what you build over the years.', bodies: ['jupiter', 'saturn'] },
  { title: 'Outer planets', sub: 'Slow movers you share with people born around the same years; their house shows where they touch you personally.', bodies: ['uranus', 'neptune', 'pluto'] },
  { title: 'Points', sub: 'Not planets, but sensitive points of the chart.', bodies: ['node', 'southnode', 'asc', 'mc', 'dsc', 'ic', 'fortune'] },
]

// ---- the planets, drawn ----

type Look = { a: string; b: string; extra?: (id: string) => ReactNode; rays?: boolean }
const LOOK: Record<string, Look> = {
  sun: { a: '#fff6c2', b: '#f08c00', rays: true },
  moon: { a: '#f8f9fa', b: '#9aa1a9', extra: () => <><circle cx="21" cy="24" r="4" fill="#868e96" opacity=".35" /><circle cx="34" cy="33" r="5.5" fill="#868e96" opacity=".3" /><circle cx="30" cy="18" r="2.4" fill="#868e96" opacity=".35" /></> },
  mercury: { a: '#dee2e6', b: '#6c6f73', extra: () => <><circle cx="22" cy="30" r="3" fill="#495057" opacity=".3" /><circle cx="33" cy="22" r="2.2" fill="#495057" opacity=".3" /></> },
  venus: { a: '#fff4d6', b: '#d9a441', extra: () => <path d="M10 26c8-4 20 4 36-2M12 33c10-3 18 3 32-1" stroke="#fff" strokeOpacity=".45" strokeWidth="2.4" fill="none" strokeLinecap="round" /> },
  mars: { a: '#ffa07a', b: '#a6321a', extra: () => <><ellipse cx="22" cy="31" rx="7" ry="3.5" fill="#6b1d0e" opacity=".35" /><ellipse cx="34" cy="21" rx="4" ry="2" fill="#6b1d0e" opacity=".3" /><path d="M20 11c4-1 10-1 14 1" stroke="#fff" strokeOpacity=".6" strokeWidth="2.2" strokeLinecap="round" /></> },
  jupiter: { a: '#f6e3c6', b: '#b9784a', extra: () => <><path d="M9 21h38M8 28h40M10 35h36" stroke="#9c5b33" strokeOpacity=".45" strokeWidth="3" /><path d="M11 17h34M12 39h32" stroke="#fff" strokeOpacity=".35" strokeWidth="2" /><ellipse cx="34" cy="32" rx="4.2" ry="2.6" fill="#c0392b" opacity=".75" /></> },
  saturn: { a: '#f7e6b5', b: '#c39a4f', extra: () => <path d="M11 24h34M10 31h36" stroke="#a87b32" strokeOpacity=".35" strokeWidth="2.6" /> },
  uranus: { a: '#d3fbfb', b: '#3ba7b8' },
  neptune: { a: '#a5c3ff', b: '#2846a8', extra: () => <><ellipse cx="22" cy="31" rx="4.5" ry="2.6" fill="#132c73" opacity=".5" /><path d="M14 20c8-2 18-2 28 1" stroke="#fff" strokeOpacity=".35" strokeWidth="2" strokeLinecap="round" fill="none" /></> },
  pluto: { a: '#ead9c2', b: '#7d6250', extra: () => <path d="M27 31c-3-3-6-1-6 2 0 4 6 7 6 7s6-3 6-7c0-3-3-5-6-2z" fill="#fff" opacity=".45" /> },
}

/** A small illustrated planet (or a glyph medallion for chart points). */
export function PlanetOrb({ body, size = 56 }: { body: string; size?: number }) {
  const uid = useId().replace(/:/g, '')
  const look = LOOK[body]
  if (!look) {
    return (
      <svg width={size} height={size} viewBox="0 0 56 56" className="pv-orb" aria-hidden>
        <circle cx="28" cy="28" r="22" fill="var(--accent-soft)" stroke="var(--accent-ring)" />
        <text x="28" y="29" textAnchor="middle" dominantBaseline="central" className="glyph" fontSize={body.length > 3 ? 14 : 20} fill="var(--accent-ink)">{BODY_GLYPH[body] || '·'}</text>
      </svg>
    )
  }
  const g = `pv-g-${uid}`, gl = `pv-gl-${uid}`, cl = `pv-c-${uid}`, glow = `pv-w-${uid}`
  return (
    <svg width={size} height={size} viewBox="0 0 56 56" className="pv-orb" aria-hidden>
      <defs>
        <radialGradient id={g} cx=".35" cy=".3" r=".8"><stop offset="0" stopColor={look.a} /><stop offset="1" stopColor={look.b} /></radialGradient>
        <radialGradient id={gl} cx=".3" cy=".25" r=".45"><stop offset="0" stopColor="#fff" stopOpacity=".55" /><stop offset="1" stopColor="#fff" stopOpacity="0" /></radialGradient>
        <radialGradient id={glow} cx=".5" cy=".5" r=".5"><stop offset=".55" stopColor={look.b} stopOpacity=".45" /><stop offset="1" stopColor={look.b} stopOpacity="0" /></radialGradient>
        <clipPath id={cl}><circle cx="28" cy="28" r="18" /></clipPath>
      </defs>
      {look.rays && <circle cx="28" cy="28" r="28" fill={`url(#${glow})`} />}
      {body === 'saturn' && <ellipse cx="28" cy="28" rx="27" ry="8" fill="none" stroke="#c9ad6f" strokeWidth="3" transform="rotate(-18 28 28)" opacity=".75" />}
      {body === 'uranus' && <ellipse cx="28" cy="28" rx="7" ry="26" fill="none" stroke="#bfeff2" strokeWidth="1.6" transform="rotate(-8 28 28)" opacity=".8" />}
      <circle cx="28" cy="28" r="18" fill={`url(#${g})`} />
      <g clipPath={`url(#${cl})`}>
        {look.extra?.(uid)}
        <circle cx="40" cy="40" r="20" fill="#000" opacity=".12" />
      </g>
      <circle cx="28" cy="28" r="18" fill={`url(#${gl})`} />
      {body === 'saturn' && (
        <g transform="rotate(-18 28 28)">
          <clipPath id={`${cl}-f`}><rect x="0" y="28" width="56" height="28" /></clipPath>
          <ellipse cx="28" cy="28" rx="27" ry="8" fill="none" stroke="#e3cc94" strokeWidth="3" clipPath={`url(#${cl}-f)`} />
        </g>
      )}
    </svg>
  )
}

// ---- the zodiac strip ----

function SkyStrip({ items, onPick }: { items: Placement[]; onPick: (b: string) => void }) {
  const ref = useRef<HTMLDivElement>(null)
  const [width, setWidth] = useState(900)
  useEffect(() => {
    const el = ref.current
    if (!el) return
    const ro = new ResizeObserver(() => setWidth(el.clientWidth || 900))
    ro.observe(el)
    return () => ro.disconnect()
  }, [])
  // planets closer than one orb's width share a column: the later ones drop a lane
  const gap = (40 / Math.max(240, width)) * 360
  const lanes: number[] = []
  const placed = [...items].sort((a, b) => a.lon - b.lon).map(p => {
    let lane = 0
    while (lanes[lane] !== undefined && p.lon - lanes[lane] < gap) lane++
    lanes[lane] = p.lon
    return { p, lane }
  })
  const LANE = 52
  return (
    <div className="pv-strip" ref={ref} style={{ height: 44 + 14 + Math.max(1, lanes.length) * LANE }} role="group" aria-label="Where each planet stands in the zodiac">
      <div className="pv-strip-signs">
        {SIGNS.map(s => (
          <span key={s} title={signName(s)} style={{ '--el': ELEMENT_COLOR[SIGN_ELEMENT[s]] } as CSSProperties}><span className="glyph">{SIGN_GLYPH[s]}</span></span>
        ))}
      </div>
      {placed.map(({ p, lane }) => (
        <button key={p.body} type="button" className="pv-strip-orb" style={{ left: `clamp(26px, ${(p.lon / 360) * 100}%, calc(100% - 26px))`, top: 54 + lane * LANE, '--drop': `${10 + lane * LANE}px` } as CSSProperties}
          title={`${p.name} ${fmtDeg(p.degree)} ${signName(p.sign)}${p.retro ? ', retrograde' : ''}`} onClick={() => onPick(p.body)}>
          <PlanetOrb body={p.body} size={28} />
          <span className="pv-strip-name">{p.name}{p.retro ? ' ℞' : ''}</span>
        </button>
      ))}
    </div>
  )
}

// ---- the cards ----

function PlanetCard({ pl, open, onToggle }: { pl: Placement; open: boolean; onToggle: () => void }) {
  const el = SIGN_ELEMENT[pl.sign]
  const color = ELEMENT_COLOR[el] || 'var(--accent)'
  const first = (pl.sign_text || '').split(/(?<=\.)\s/)[0]
  return (
    <article id={`pl-${pl.body}`} className={`pv-card ${open ? 'open' : ''}`} style={{ '--el': color } as CSSProperties}>
      <button type="button" className="pv-card-head" onClick={onToggle} aria-expanded={open}>
        <PlanetOrb body={pl.body} />
        <div className="pv-card-title">
          <div className="pv-card-name"><b>{pl.name}</b>{pl.retro && <span className="pv-retro" title="Retrograde at birth">℞ retrograde</span>}</div>
          <div className="muted small">{pl.short}</div>
          <div className="pv-chips">
            <span className="pv-sign"><span className="glyph">{SIGN_GLYPH[pl.sign]}</span>{signName(pl.sign)} <span className="pv-deg">{fmtDeg(pl.degree)}</span></span>
            {pl.house > 0 && <span className="pv-house">{ORD[pl.house]} house</span>}
          </div>
        </div>
      </button>
      <div className="pv-progress" title={`${fmtDeg(pl.degree)} of the 30° of ${signName(pl.sign)}`}><span style={{ width: `${Math.min(100, (pl.degree / 30) * 100)}%` }} /></div>
      <div className="pv-text">
        {open ? (
          <>
            {pl.sign_text && <p><b>In {signName(pl.sign)}.</b> {pl.sign_text}</p>}
            {pl.house_text && <p><b>In the {ORD[pl.house]} house.</b> {pl.house_text}</p>}
            {pl.retro && <p className="muted small">Retrograde at birth: this side of you tends to work inwardly first, and shows later.</p>}
          </>
        ) : <p>{first}</p>}
        {(pl.house_text || (pl.sign_text && pl.sign_text !== first)) && (
          <button type="button" className="pv-more" onClick={onToggle}>{open ? 'Less' : 'Read more'}<Icon name={open ? 'chevron-up' : 'chevron-down'} size={14} /></button>
        )}
      </div>
    </article>
  )
}

export function PlanetsView({ placements, points, open, onOpen }: { placements: Placement[]; points: Placement[]; open: string; onOpen: (b: string) => void }) {
  const all = [...placements, ...points]
  const byBody = new Map(all.map(p => [p.body, p]))
  const pick = (b: string) => { onOpen(b); setTimeout(() => document.getElementById(`pl-${b}`)?.scrollIntoView({ block: 'center', behavior: 'smooth' }), 0) }
  return (
    <div className="pv">
      <SkyStrip items={placements.filter(p => p.body !== 'node')} onPick={pick} />
      {GROUPS.map(g => {
        const items = g.bodies.map(b => byBody.get(b)).filter((p): p is Placement => !!p)
        if (!items.length) return null
        return (
          <section key={g.title} className="pv-group">
            <div className="pv-group-head"><h3>{g.title}</h3><p className="muted small">{g.sub}</p></div>
            <div className="pv-grid">
              {items.map(pl => <PlanetCard key={pl.body} pl={pl} open={open === pl.body} onToggle={() => onOpen(open === pl.body ? '' : pl.body)} />)}
            </div>
          </section>
        )
      })}
    </div>
  )
}
