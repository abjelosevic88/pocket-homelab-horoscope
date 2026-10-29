// The Aspects tab: how the planets talk to each other. A summary of flowing,
// tense and blending links, filters by kind and by planet, a tinted aspect
// grid (desktop) and one card per aspect: the two planets, the link between
// them, how strong it is and what it means.
import { useState, type CSSProperties } from 'react'
import { ASPECT_COLOR, ASPECT_GLYPH, ASPECT_NAME, BODY_GLYPH } from '../lib/astro'
import type { AspectView, Placement } from '../lib/types'
import { Icon } from './icons'
import { PlanetOrb } from './PlanetsView'
import { Empty, useIsMobile } from './ui'

type Nature = 'easy' | 'hard' | 'blend'
const NATURE: Record<Nature, { label: string; word: string; color: string; text: string }> = {
  easy: { label: 'Flowing', word: 'flowing', color: '#1c7ed6', text: 'trines and sextiles: talents that work without effort' },
  hard: { label: 'Tense', word: 'tense', color: '#e03131', text: 'squares and oppositions: friction that pushes you to grow' },
  blend: { label: 'Blending', word: 'blending', color: 'var(--accent)', text: 'conjunctions: two sides of you fused into one' },
}

function strength(orb: number): { pct: number; word: string } {
  const pct = Math.max(6, Math.round((1 - Math.min(orb, 9) / 9) * 100))
  return { pct, word: orb < 1 ? 'very strong' : orb < 3 ? 'strong' : orb < 5.5 ? 'moderate' : 'mild' }
}

function summary(easy: number, hard: number): string {
  if (!easy && !hard) return 'Few links between your planets: each part of you works fairly independently.'
  if (easy >= hard * 2) return 'Many more flowing links than tense ones: much comes naturally to you. Growth comes from choosing a challenge on purpose.'
  if (hard >= easy * 2) return 'More tense links than flowing ones: life pushes you, and that friction is where your strength and drive come from.'
  return 'A good mix of flowing and tense links: natural gifts, and enough friction to keep you growing.'
}

export function AspectsView({ aspects, placements }: { aspects: AspectView[]; placements: Placement[] }) {
  const mobile = useIsMobile()
  const [open, setOpen] = useState('')
  const [nature, setNature] = useState<Nature | ''>('')
  const [body, setBody] = useState('')
  if (!aspects.length) return <Empty title="No aspects" icon="activity">No planets are at a meaningful angle to each other within the orbs you chose.</Empty>
  const keyOf = (a: AspectView) => `${a.a}-${a.b}-${a.type}`
  const counts = { easy: 0, hard: 0, blend: 0 } as Record<Nature, number>
  for (const a of aspects) counts[a.nature as Nature]++
  const bodies = placements.map(p => p.body)
  const inChart = new Set(aspects.flatMap(a => [a.a, a.b]))
  const find = (x: string, y: string) => aspects.find(a => (a.a === x && a.b === y) || (a.a === y && a.b === x))
  const shown = aspects.filter(a => (!nature || a.nature === nature) && (!body || a.a === body || a.b === body))
  const openAspect = (a: AspectView) => {
    setOpen(keyOf(a))
    setTimeout(() => document.getElementById(`as-${keyOf(a)}`)?.scrollIntoView({ block: 'center', behavior: 'smooth' }), 0)
  }
  const nameOf = (k: string) => placements.find(p => p.body === k)?.name || aspects.find(a => a.a === k)?.a_name || aspects.find(a => a.b === k)?.b_name || k

  return (
    <div className="av">
      <div className="av-sum">
        {(Object.keys(NATURE) as Nature[]).map(n => (
          <button key={n} type="button" className={`av-stat ${nature === n ? 'on' : ''}`} style={{ '--c': NATURE[n].color } as CSSProperties} onClick={() => setNature(nature === n ? '' : n)} aria-pressed={nature === n}>
            <span className="av-stat-n">{counts[n]}</span>
            <span className="av-stat-l"><b>{NATURE[n].label}</b><span className="muted xs">{NATURE[n].text}</span></span>
          </button>
        ))}
      </div>
      <p className="av-summary muted">{summary(counts.easy, counts.hard)}</p>

      <div className="av-filter" role="group" aria-label="Show the aspects of one planet">
        <button type="button" className={`av-chip ${!body ? 'on' : ''}`} onClick={() => setBody('')}>All planets</button>
        {bodies.filter(b => inChart.has(b)).map(b => (
          <button key={b} type="button" className={`av-chip ${body === b ? 'on' : ''}`} onClick={() => setBody(body === b ? '' : b)}>
            <PlanetOrb body={b} size={18} />{nameOf(b)}
          </button>
        ))}
      </div>

      <div className="av-main">
        {!mobile && (
          <div className="av-grid-card">
            <h3>Aspect grid</h3>
            <p className="muted xs">Each square is where two planets meet. Tap one to read it.</p>
            <table className="av-grid" aria-label="Aspect grid">
              <tbody>
                {bodies.map((row, i) => (
                  <tr key={row}>
                    <th scope="row" className="glyph" title={nameOf(row)}>{BODY_GLYPH[row]}</th>
                    {bodies.slice(0, i).map(col => {
                      const a = find(row, col)
                      const dim = a && ((nature && a.nature !== nature) || (body && a.a !== body && a.b !== body))
                      return (
                        <td key={col}>
                          {a && (
                            <button type="button" className={`av-cell glyph ${open === keyOf(a) ? 'on' : ''} ${dim ? 'dim' : ''}`} style={{ '--c': ASPECT_COLOR[a.type] } as CSSProperties}
                              title={`${a.title}, ${a.orb.toFixed(1)}°`} onClick={() => openAspect(a)}>{ASPECT_GLYPH[a.type]}</button>
                          )}
                        </td>
                      )
                    })}
                    <td className="av-diag glyph" title={nameOf(row)}>{BODY_GLYPH[row]}</td>
                  </tr>
                ))}
              </tbody>
            </table>
            <div className="av-legend">
              {Object.keys(ASPECT_NAME).filter(t => aspects.some(a => a.type === t)).map(t => (
                <span key={t}><span className="glyph" style={{ color: ASPECT_COLOR[t] }}>{ASPECT_GLYPH[t]}</span>{ASPECT_NAME[t]}</span>
              ))}
            </div>
          </div>
        )}

        <div className="av-list">
          {shown.length === 0 && <p className="muted small">No aspects match. <button type="button" className="link-btn" onClick={() => { setNature(''); setBody('') }}>Show all</button></p>}
          {shown.map(a => {
            const k = keyOf(a)
            const isOpen = open === k
            const s = strength(a.orb)
            const c = ASPECT_COLOR[a.type]
            return (
              <article key={k} id={`as-${k}`} className={`av-card ${isOpen ? 'open' : ''}`} style={{ '--c': c } as CSSProperties}>
                <button type="button" className="av-card-head" onClick={() => setOpen(isOpen ? '' : k)} aria-expanded={isOpen}>
                  <span className="av-pair">
                    <PlanetOrb body={a.a} size={34} />
                    <span className={`av-link ${a.nature}`}><span className="av-link-g glyph">{ASPECT_GLYPH[a.type]}</span></span>
                    <PlanetOrb body={a.b} size={34} />
                  </span>
                  <span className="av-card-title">
                    <b>{a.title}</b>
                    <span className="av-meta"><span className="av-nature">{NATURE[a.nature as Nature]?.word}</span><span className="muted xs">{a.orb.toFixed(1)}° from exact · {s.word}</span></span>
                    <span className="av-strength" title={`${s.word}: the closer to exact, the stronger`}><span style={{ width: `${s.pct}%` }} /></span>
                  </span>
                  <Icon name={isOpen ? 'chevron-up' : 'chevron-down'} size={15} className="muted" />
                </button>
                {isOpen && (
                  <div className="av-text">
                    <p>{a.text}</p>
                    <p className="muted small">{ASPECT_NAME[a.type]} ({Math.round(a.exact)}°): {NATURE[a.nature as Nature]?.text}. {a.applying ? 'Still tightening at birth, so it tends to grow in importance over life.' : 'Already easing at birth, so it is often felt most in early life.'}</p>
                    <p className="small"><button type="button" className="link-btn" onClick={() => setBody(a.a)}>All aspects of {a.a_name}</button> · <button type="button" className="link-btn" onClick={() => setBody(a.b)}>All aspects of {a.b_name}</button></p>
                  </div>
                )}
              </article>
            )
          })}
        </div>
      </div>
    </div>
  )
}
