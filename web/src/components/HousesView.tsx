// The Houses tab: a dial of the twelve houses as they fall in this chart
// (sized by the real cusps, planets inside), where the planets gather, and a
// card per house with the planets in it and what they mean there.
import { useState, type CSSProperties } from 'react'
import { ELEMENT_COLOR, SIGN_ELEMENT, SIGN_GLYPH, fmtDeg, signName } from '../lib/astro'
import type { HouseView, Placement } from '../lib/types'
import { Icon } from './icons'
import { PlanetOrb } from './PlanetsView'

const ORD = ['', '1st', '2nd', '3rd', '4th', '5th', '6th', '7th', '8th', '9th', '10th', '11th', '12th']
const ANGULAR = new Set([1, 4, 7, 10])

function Dial({ houses, placements, asc, onPick, picked }: { houses: HouseView[]; placements: Placement[]; asc: number; onPick: (n: number) => void; picked: number }) {
  const C = 160, R = 150, r0 = 58
  // the Ascendant at 9 o'clock, the zodiac running counter-clockwise (as on the chart wheel)
  const ang = (lon: number) => ((180 + (lon - asc)) * Math.PI) / 180
  const pt = (lon: number, rad: number): [number, number] => [C + rad * Math.cos(ang(lon)), C - rad * Math.sin(ang(lon))]
  const span = (h: HouseView, i: number) => {
    const next = houses[(i + 1) % 12].cusp
    let d = next - h.cusp
    if (d <= 0) d += 360
    return d
  }
  const wedge = (a: number, d: number, ri: number, ro: number) => {
    const [x1, y1] = pt(a, ro), [x2, y2] = pt(a + d, ro), [x3, y3] = pt(a + d, ri), [x4, y4] = pt(a, ri)
    const large = d > 180 ? 1 : 0
    return `M${x1} ${y1} A${ro} ${ro} 0 ${large} 0 ${x2} ${y2} L${x3} ${y3} A${ri} ${ri} 0 ${large} 1 ${x4} ${y4}Z`
  }
  return (
    <svg viewBox="0 0 320 320" className="hv-dial" role="img" aria-label="The twelve houses of this chart">
      {houses.map((h, i) => {
        const d = span(h, i)
        const el = ELEMENT_COLOR[SIGN_ELEMENT[h.sign]]
        const inside = placements.filter(p => p.house === h.n && p.body !== 'node').sort((a, b) => a.lon - b.lon)
        const [nx, ny] = pt(h.cusp + d / 2, r0 + 16)
        return (
          <g key={h.n} className={`hv-wedge ${picked === h.n ? 'on' : ''}`} onClick={() => onPick(h.n)} style={{ '--el': el } as CSSProperties}>
            <title>{`${ORD[h.n]} house · ${h.name} · ${signName(h.sign)}`}</title>
            <path d={wedge(h.cusp, d, r0, R)} className="hv-wedge-path" />
            <text x={nx} y={ny} textAnchor="middle" dominantBaseline="central" className={`hv-n ${ANGULAR.has(h.n) ? 'ang' : ''}`}>{h.n}</text>
            {inside.map((p, k) => {
              const ringR = r0 + 46 + (k % 2) * 34
              const frac = (k + 1) / (inside.length + 1)
              const [px, py] = pt(h.cusp + d * frac, ringR)
              return <svg key={p.body} x={px - 13} y={py - 13} width={26} height={26} viewBox="0 0 56 56" overflow="visible"><PlanetOrbInner body={p.body} /></svg>
            })}
          </g>
        )
      })}
      {(['AC', 'IC', 'DC', 'MC'] as const).map((lab, i) => {
        const cusp = houses[i * 3]?.cusp ?? 0
        const [x1, y1] = pt(cusp, r0 - 4), [x2, y2] = pt(cusp, R + 4), [lx, ly] = pt(cusp, R - 12)
        return (
          <g key={lab}>
            <line x1={x1} y1={y1} x2={x2} y2={y2} className="hv-axis" />
            <text x={lx} y={ly} textAnchor="middle" dominantBaseline="central" className="hv-axis-l">{lab}</text>
          </g>
        )
      })}
      <circle cx={C} cy={C} r={r0 - 4} className="hv-hub" />
      <text x={C} y={C - 7} textAnchor="middle" className="hv-hub-t">Houses</text>
      <text x={C} y={C + 11} textAnchor="middle" className="hv-hub-s">areas of life</text>
    </svg>
  )
}

// PlanetOrb draws its own <svg>; inside the dial it is nested, which SVG allows.
function PlanetOrbInner({ body }: { body: string }) { return <PlanetOrb body={body} size={56} /> }

function Balance({ placements }: { placements: Placement[] }) {
  const ps = placements.filter(p => p.house > 0 && p.body !== 'node')
  if (!ps.length) return null
  const above = ps.filter(p => p.house >= 7).length
  const east = ps.filter(p => p.house >= 10 || p.house <= 3).length
  const n = ps.length
  const rows: { l: string; r: string; a: number; text: string }[] = [
    { l: 'Below', r: 'Above', a: above, text: above > n / 2 + 1 ? 'More planets above the horizon: much of your life plays out in public, at work and in the world.' : above < n / 2 - 1 ? 'More planets below the horizon: you live a lot inwardly, at home and in private, and grow from there.' : 'Planets spread above and below the horizon: you balance private life and the outside world.' },
    { l: 'Others', r: 'Self', a: east, text: east > n / 2 + 1 ? 'More planets on the eastern side: you tend to steer your own course.' : east < n / 2 - 1 ? 'More planets on the western side: you grow through other people and partnerships.' : 'An even east–west spread: you can lead and you can meet people halfway.' },
  ]
  return (
    <div className="hv-balance">
      {rows.map(r => (
        <div key={r.l} className="hv-bal-row">
          <div className="hv-bal-bar"><span className="muted xs">{r.l} {n - r.a}</span><div><span style={{ width: `${((n - r.a) / n) * 100}%` }} /></div><span className="muted xs">{r.a} {r.r}</span></div>
          <p className="muted small">{r.text}</p>
        </div>
      ))}
    </div>
  )
}

export function HousesView({ houses, placements, asc, system }: { houses: HouseView[]; placements: Placement[]; asc: number; system: string }) {
  const [picked, setPicked] = useState(0)
  const [open, setOpen] = useState<Record<number, boolean>>({})
  const pick = (n: number) => { setPicked(n); setTimeout(() => document.getElementById(`hv-${n}`)?.scrollIntoView({ block: 'center', behavior: 'smooth' }), 0) }
  const counts = houses.map(h => ({ h, n: placements.filter(p => p.house === h.n && p.body !== 'node').length })).filter(x => x.n > 0).sort((a, b) => b.n - a.n)
  return (
    <div className="hv">
      <div className="hv-top">
        <div className="hv-dial-wrap"><Dial houses={houses} placements={placements} asc={asc} onPick={pick} picked={picked} /></div>
        <div className="hv-side">
          <h3>Where your planets gather</h3>
          <p className="muted small">The houses are twelve areas of life. A house with planets in it is where much of your energy goes; an empty house is not a missing part of life, just a quieter one. {system}.</p>
          <div className="hv-top3">
            {counts.slice(0, 3).map(({ h, n }) => (
              <button key={h.n} type="button" className="hv-top3-item" onClick={() => pick(h.n)} style={{ '--el': ELEMENT_COLOR[SIGN_ELEMENT[h.sign]] } as CSSProperties}>
                <span className="hv-badge">{h.n}</span>
                <span className="grow"><b>{h.name}</b><span className="muted xs">{n} {n === 1 ? 'planet' : 'planets'}</span></span>
                <span className="hv-top3-orbs">{placements.filter(p => p.house === h.n && p.body !== 'node').map(p => <PlanetOrb key={p.body} body={p.body} size={22} />)}</span>
              </button>
            ))}
          </div>
          <Balance placements={placements} />
        </div>
      </div>

      <div className="hv-grid">
        {houses.map(h => {
          const inside = placements.filter(p => p.house === h.n)
          const el = ELEMENT_COLOR[SIGN_ELEMENT[h.sign]]
          const isOpen = !!open[h.n]
          return (
            <article key={h.n} id={`hv-${h.n}`} className={`hv-card ${inside.length ? '' : 'hv-empty'} ${picked === h.n ? 'picked' : ''}`} style={{ '--el': el } as CSSProperties}>
              <div className="hv-card-head">
                <span className="hv-badge lg">{h.n}</span>
                <div className="grow">
                  <div className="hv-card-name"><b>{h.name}</b>{ANGULAR.has(h.n) && <span className="hv-ang" title="Angular houses (1, 4, 7 and 10) are the strongest">angle</span>}</div>
                  <span className="pv-sign" style={{ '--el': el } as CSSProperties}><span className="glyph">{SIGN_GLYPH[h.sign]}</span>{signName(h.sign)} <span className="pv-deg">{fmtDeg(h.cusp % 30)}</span></span>
                </div>
              </div>
              <p className="hv-sum">{h.summary}</p>
              {inside.length > 0 ? (
                <>
                  <div className="hv-planets">
                    {inside.map(p => <span key={p.body} className="hv-planet"><PlanetOrb body={p.body} size={22} />{p.name}{p.retro ? ' ℞' : ''}</span>)}
                  </div>
                  {inside.some(p => p.house_text) && (
                    <>
                      {isOpen && <div className="hv-texts">{inside.filter(p => p.house_text).map(p => <p key={p.body}><b>{p.name} here.</b> {p.house_text}</p>)}</div>}
                      <button type="button" className="pv-more" onClick={() => setOpen({ ...open, [h.n]: !isOpen })}>{isOpen ? 'Less' : `What ${inside.length === 1 ? `${inside[0].name} means` : 'they mean'} here`}<Icon name={isOpen ? 'chevron-up' : 'chevron-down'} size={14} /></button>
                    </>
                  )}
                </>
              ) : <p className="muted xs hv-quiet">No planets here. {signName(h.sign)} on its cusp colours how this part of life feels.</p>}
            </article>
          )
        })}
      </div>
    </div>
  )
}
