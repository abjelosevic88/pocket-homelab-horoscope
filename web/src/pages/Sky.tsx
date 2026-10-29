// The sky for everyone: where the planets are now, the Moon calendar, the
// year's retrogrades and the events coming up. Times are the home zone's.
import { type ReactNode, useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { api, V1 } from '../lib/api'
import { useApp } from '../lib/app-context'
import { BODY_GLYPH, BODY_NAME, PHASE_EMOJI, PHASE_NAME, SIGN_GLYPH, clockOf, dateOf, fmtLon, signName } from '../lib/astro'
import { addDays, fmtClock, fmtDate, todayISO } from '../lib/format'
import type { MoonDay, MoonMonth, RetroPeriod, SkyEvent, SkyNow } from '../lib/types'
import { Icon } from '../components/icons'
import { ChartWheel } from '../components/ChartWheel'
import { useCrumb } from '../components/Layout'
import { Card, Empty, Loading, PageHeader, Tabs, useAsync, useIsMobile } from '../components/ui'

type View = 'now' | 'moon' | 'retrogrades' | 'events'
const VIEWS: { id: View; label: string; icon: string }[] = [
  { id: 'now', label: 'Now', icon: 'sun' },
  { id: 'moon', label: 'Moon calendar', icon: 'moon' },
  { id: 'retrogrades', label: 'Retrogrades', icon: 'refresh' },
  { id: 'events', label: 'Coming up', icon: 'list' },
]

export default function Sky() {
  const { view: raw } = useParams()
  const navigate = useNavigate()
  const view: View = VIEWS.find(v => v.id === raw)?.id || 'now'
  useCrumb(view === 'now' ? undefined : VIEWS.find(v => v.id === view)!.label)
  return (
    <div className="page sky-page">
      <PageHeader title="Sky & Moon" sub="The same sky for everyone: where the planets are, the Moon’s month and what is coming." />
      <div className="sky-tabs"><Tabs tabs={VIEWS} value={view} onChange={v => navigate(v === 'now' ? '/sky' : `/sky/${v}`, { replace: true })} /></div>
      {view === 'now' && <NowView />}
      {view === 'moon' && <MoonView />}
      {view === 'retrogrades' && <RetroView />}
      {view === 'events' && <EventsView />}
    </div>
  )
}

/** "today", "tomorrow" or the date, for a home-zone timestamp relative to a home-zone day. */
function dayWord(iso: string, today: string): string {
  const d = dateOf(iso)
  if (d === today) return 'today'
  if (d === addDays(today, 1)) return 'tomorrow'
  if (d === addDays(today, -1)) return 'yesterday'
  return fmtDate(d, false)
}

// ---------- now ----------

/** Average daily motion, to say whether a planet is fast, slow or nearly still. */
const AVG_SPEED: Record<string, number> = { sun: 0.986, moon: 13.18, mercury: 1.2, venus: 1.1, mars: 0.52, jupiter: 0.083, saturn: 0.034, uranus: 0.012, neptune: 0.006, pluto: 0.004, node: 0.053 }

function speedWord(body: string, speed: number): string {
  if (body === 'node') return speed < 0 ? 'moves backwards, as usual' : 'moves forwards'
  const avg = AVG_SPEED[body] || 1
  const r = Math.abs(speed) / avg
  const dir = speed < 0 ? 'backwards' : ''
  if (r < 0.15) return speed < 0 ? 'almost still — about to turn direct' : 'almost still — about to turn'
  if (r < 0.7) return `slow${dir ? ', ' + dir : ''}`
  if (r > 1.3) return `fast${dir ? ', ' + dir : ''}`
  return dir || 'normal pace'
}

function NowView() {
  const { data, error } = useAsync(() => api.get<SkyNow>(`${V1}/sky/now`), [])
  const mobile = useIsMobile()
  const [pick, setPick] = useState('')
  if (error) return <div className="callout danger"><Icon name="alert" size={15} /><span>{error}</span></div>
  if (!data) return <Loading />
  const today = dateOf(data.at)
  const speeds = new Map(data.chart.positions.map(p => [p.body, p.speed]))
  const v = data.moon.void
  return (
    <div className="stack loose">
      {!data.has_home && (
        <div className="callout info"><Icon name="pin" size={15} /><div>Times and houses need a place. <Link to="/settings/charts">Set where you live</Link> (or add your own birth details) to see the sky for your town.</div></div>
      )}
      <div className="sky-now">
        <Card className="sky-wheel-card" title="The sky right now" sub={<>{data.place ? `For ${data.place}, ` : ''}{fmtDate(today)} at {fmtClock(clockOf(data.at))}{data.chart.houses ? ' · houses for this place and time' : ''}</>}>
          <div className="sky-wheel">
            <ChartWheel chart={data.chart} size={mobile ? 320 : 420} onPick={k => setPick(p => p === k ? '' : k)} highlight={pick || undefined} />
          </div>
        </Card>
        <div className="stack">
          <Card className="sky-moon-card">
            <div className="sky-moon-top">
              <MoonIcon illumination={data.moon.illumination} waxing={data.moon.age < 14.77} size={56} />
              <div>
                <div className="sky-moon-title">{PHASE_NAME[data.moon.phase]} in {signName(data.moon.sign)}</div>
                <div className="muted small">{Math.round(data.moon.illumination * 100)}% lit · {Math.round(data.moon.age * 10) / 10} days since the new Moon</div>
              </div>
            </div>
            <p className="sky-text">{data.moon.text}</p>
            <p className="sky-text muted">{data.moon.phase_text}</p>
            {v.from && v.to && (
              <div className="sky-void">
                <Icon name="clock" size={14} />
                <div>
                  <b>Void of course</b> from {fmtClock(clockOf(v.from))} {dayWord(v.from, today)} to {fmtClock(clockOf(v.to))} {dayWord(v.to, today)}.
                  <span className="muted"> The Moon makes no more big aspects before it changes sign: a quiet stretch, better for finishing than for starting.</span>
                </div>
              </div>
            )}
          </Card>
          <Card title="Where the planets are" flush>
            <table className="table compact sky-table">
              <thead><tr><th aria-label="Glyph" /><th>Planet</th><th>Where</th>{!mobile && <th>Moving</th>}</tr></thead>
              <tbody>
                {data.bodies.map(b => {
                  const sp = speeds.get(b.body) ?? b.speed
                  return (
                    <tr key={b.body} className={pick === b.body ? 'sky-row-hl' : ''} onClick={() => setPick(p => p === b.body ? '' : b.body)}>
                      <td className="glyph sky-glyph">{BODY_GLYPH[b.body] || b.glyph}</td>
                      <td>{BODY_NAME[b.body] || b.name}{b.retro && b.body !== 'node' && <span className="sky-retro" title="Retrograde: seems to move backwards">℞</span>}{mobile && <div className="muted small">{speedWord(b.body, sp)}</div>}</td>
                      <td className="tnum">{fmtLon(b.lon)}{b.house ? <span className="muted small"> · house {b.house}</span> : null}</td>
                      {!mobile && <td className="muted small">{speedWord(b.body, sp)}</td>}
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </Card>
        </div>
      </div>
      <Card title="Coming up in the next two weeks" actions={<Link to="/sky/events" className="btn ghost sm">More<Icon name="chevron-right" size={14} /></Link>}>
        {data.events.length === 0 ? <p className="muted">A quiet fortnight: no new or full Moons, sign changes or retrogrades.</p> : <EventList events={data.events} today={today} />}
      </Card>
    </div>
  )
}

/** A Moon drawn as lit and dark parts (lit on the right while waxing, as seen from the north). */
export function MoonIcon({ illumination, waxing, size = 24 }: { illumination: number; waxing: boolean; size?: number }) {
  const r = size / 2 - 1
  const c = size / 2
  const k = Math.max(0, Math.min(1, illumination))
  let lit: ReactNode = null
  if (k >= 0.985) lit = <circle cx={c} cy={c} r={r} className="moon-lit" />
  else if (k > 0.015) {
    const rx = r * Math.abs(2 * k - 1)
    const outer = waxing ? 1 : 0
    const term = waxing ? (k < 0.5 ? 0 : 1) : (k < 0.5 ? 1 : 0)
    lit = <path className="moon-lit" d={`M ${c} ${c - r} A ${r} ${r} 0 0 ${outer} ${c} ${c + r} A ${rx} ${r} 0 0 ${term} ${c} ${c - r} Z`} />
  }
  return (
    <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`} className="moon-svg" aria-hidden>
      <circle cx={c} cy={c} r={r} className="moon-dark" />
      {lit}
    </svg>
  )
}

function EventList({ events, today }: { events: SkyEvent[]; today: string }) {
  const [open, setOpen] = useState<string | null>(null)
  return (
    <div className="sky-events">
      {events.map(e => {
        const key = `${e.kind}-${e.time}-${e.body || ''}`
        const isOpen = open === key
        return (
          <div key={key} className={`sky-event ${isOpen ? 'open' : ''}`}>
            <button type="button" className="sky-event-row" aria-expanded={isOpen} onClick={() => setOpen(isOpen ? null : key)}>
              <span className={`sky-event-ico k-${e.kind}`}>{e.icon}</span>
              <span className="sky-event-main">
                <span className="sky-event-title">{e.title}</span>
                <span className="muted small">{dayWord(e.time, today) === 'today' || dayWord(e.time, today) === 'tomorrow' ? dayWord(e.time, today) : fmtDate(e.date, false)} · {fmtClock(clockOf(e.time))}</span>
              </span>
              <Icon name="chevron-down" size={15} className="sky-event-caret muted" />
            </button>
            {isOpen && e.text && <p className="sky-event-text">{e.text}</p>}
          </div>
        )
      })}
    </div>
  )
}

// ---------- the Moon calendar ----------

const WEEKDAYS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']

function monthLabel(m: string): string {
  const [y, mo] = m.split('-').map(Number)
  return new Date(y, mo - 1, 1).toLocaleDateString(undefined, { month: 'long', year: 'numeric' })
}
function shiftMonth(m: string, n: number): string {
  const [y, mo] = m.split('-').map(Number)
  const d = new Date(y, mo - 1 + n, 1)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
}

const MOOD_COLOR = ['', 'var(--danger)', 'var(--warning)', 'var(--muted)', 'var(--accent)', 'var(--success)']

function MoonView() {
  const { settings } = useApp()
  const [sp, setSp] = useSearchParams()
  const today = todayISO()
  const month = /^\d{4}-\d{2}$/.test(sp.get('month') || '') ? sp.get('month')! : today.slice(0, 7)
  const setMonth = (m: string) => { const n = new URLSearchParams(sp); if (m === today.slice(0, 7)) n.delete('month'); else n.set('month', m); setSp(n, { replace: true }) }
  const { data, error } = useAsync(() => api.get<MoonMonth>(`${V1}/sky/moon?month=${month}`), [month])
  const [sel, setSel] = useState<string>('')
  useEffect(() => { setSel(today.slice(0, 7) === month ? today : '') }, [month, today])
  const ws = settings?.week_start || 1
  const shown = data && data.month === month ? data : null
  const lead = useMemo(() => {
    const [y, m] = month.split('-').map(Number)
    const iso = new Date(y, m - 1, 1).getDay() || 7
    return (iso - ws + 7) % 7
  }, [month, ws])
  const heads = useMemo(() => Array.from({ length: 7 }, (_, i) => WEEKDAYS[(ws - 1 + i) % 7]), [ws])
  const day = shown?.days.find(d => d.date === sel)
  return (
    <div className="stack loose">
      <div className="sky-nav">
        <button type="button" className="btn ghost icon-only" aria-label="Previous month" onClick={() => setMonth(shiftMonth(month, -1))}><Icon name="chevron-left" size={18} /></button>
        <h2 className="sky-nav-title">{monthLabel(month)}</h2>
        <button type="button" className="btn ghost icon-only" aria-label="Next month" onClick={() => setMonth(shiftMonth(month, 1))}><Icon name="chevron-right" size={18} /></button>
        <button type="button" className="btn sm" onClick={() => setMonth(today.slice(0, 7))} disabled={month === today.slice(0, 7)}>This month</button>
      </div>
      {error && <div className="callout danger"><Icon name="alert" size={15} /><span>{error}</span></div>}
      <div className="moon-cal" aria-busy={!shown}>
        {heads.map(h => <div key={h} className="moon-head">{h}</div>)}
        {Array.from({ length: lead }, (_, i) => <div key={`l${i}`} className="moon-cell blank" />)}
        {(shown?.days || placeholderDays(month)).map(d => (
          <MoonCell key={d.date} d={d} today={today} selected={sel === d.date} onClick={() => setSel(sel === d.date ? '' : d.date)} loading={!shown} />
        ))}
      </div>
      <div className="moon-legend muted small">
        <span><span className="moon-badge b-ecl">E</span> eclipse</span>
        <span><span className="moon-badge b-st">℞</span> retrograde / <span className="moon-badge b-st">D</span> direct</span>
        <span><span className="moon-vc">v/c</span> void of course</span>
        <span><span className="moon-mood" style={{ background: 'var(--accent)' }} /> your mood</span>
      </div>
      {day ? <MoonDayDetail d={day} today={today} /> : <p className="muted small">Tap a day for its Moon and events.</p>}
    </div>
  )
}

/** Empty cells while the month loads, so the grid keeps its size. */
function placeholderDays(month: string): MoonDay[] {
  const [y, m] = month.split('-').map(Number)
  const n = new Date(y, m, 0).getDate()
  return Array.from({ length: n }, (_, i) => ({ date: `${month}-${String(i + 1).padStart(2, '0')}`, phase: 'new', illumination: 0, angle: 0, sign: '', events: [] }))
}

function MoonCell({ d, today, selected, onClick, loading }: { d: MoonDay; today: string; selected: boolean; onClick: () => void; loading: boolean }) {
  const big = d.events.find(e => e.kind === 'phase')
  const ecl = d.events.some(e => e.kind === 'eclipse')
  const stations = d.events.filter(e => e.kind === 'station')
  return (
    <button type="button" className={`moon-cell ${selected ? 'sel' : ''} ${d.date === today ? 'today' : ''} ${big ? 'big' : ''}`} onClick={onClick} disabled={loading}
      aria-label={loading ? d.date : `${fmtDate(d.date)}: ${PHASE_NAME[d.phase]}, Moon in ${signName(d.sign)}`}>
      <span className="moon-num">{Number(d.date.slice(8))}</span>
      {!loading && <>
        <span className="moon-icon"><MoonIcon illumination={d.illumination} waxing={d.angle < 180} size={22} /></span>
        <span className="moon-sign glyph" title={`Moon in ${signName(d.sign)}`}>{SIGN_GLYPH[d.sign]}{d.sign_change && <span className="moon-ing">{fmtClock(clockOf(d.sign_change))}</span>}</span>
        <span className="moon-badges">
          {big && <span className="moon-badge b-phase" title={big.title}>{big.icon}</span>}
          {ecl && <span className="moon-badge b-ecl" title="Eclipse">E</span>}
          {stations.map(s => <span key={s.time} className="moon-badge b-st" title={s.title}>{s.direction === 'retrograde' ? '℞' : 'D'}</span>)}
          {d.void && <span className="moon-vc" title="Void of course">v/c</span>}
        </span>
        {!!d.mood && <span className="moon-mood" style={{ background: MOOD_COLOR[d.mood] }} title={`Your mood: ${d.mood}/5`} />}
      </>}
    </button>
  )
}

function MoonDayDetail({ d, today }: { d: MoonDay; today: string }) {
  return (
    <Card className="moon-detail" title={fmtDate(d.date)} sub={`${PHASE_EMOJI[d.phase]} ${PHASE_NAME[d.phase]} · ${Math.round(d.illumination * 100)}% lit`}>
      <div className="stack">
        <p className="sky-text">
          The Moon is in <b>{signName(d.sign)}</b>{d.sign_change && <> from {fmtClock(clockOf(d.sign_change))}</>}.
          {d.void && <> Void of course from {fmtClock(clockOf(d.void[0]))} {dayWord(d.void[0], d.date)} to {fmtClock(clockOf(d.void[1]))} {dayWord(d.void[1], d.date)}.</>}
        </p>
        {!!d.mood && <p className="muted small">Your check-in that day: mood {d.mood}/5.</p>}
        {d.events.length > 0 && <EventList events={d.events} today={today} />}
      </div>
    </Card>
  )
}

// ---------- retrogrades ----------

const RETRO_ORDER = ['mercury', 'venus', 'mars', 'jupiter', 'saturn', 'uranus', 'neptune', 'pluto']
const MONTH_LETTERS = ['J', 'F', 'M', 'A', 'M', 'J', 'J', 'A', 'S', 'O', 'N', 'D']

function RetroView() {
  const thisYear = Number(todayISO().slice(0, 4))
  const [year, setYear] = useState(thisYear)
  const { data, error } = useAsync(() => api.get<{ year: number; periods: RetroPeriod[] }>(`${V1}/sky/retrogrades?year=${year}`), [year])
  const shown = data && data.year === year ? data.periods : null
  const y0 = new Date(year, 0, 1).getTime()
  const y1 = new Date(year + 1, 0, 1).getTime()
  const pct = (iso: string) => Math.max(0, Math.min(100, ((new Date(iso).getTime() - y0) / (y1 - y0)) * 100))
  const now = Date.now()
  const todayPct = now >= y0 && now < y1 ? ((now - y0) / (y1 - y0)) * 100 : null
  const bodies = RETRO_ORDER.filter(b => shown?.some(p => p.body === b))
  return (
    <div className="stack loose">
      <div className="sky-nav">
        <button type="button" className="btn ghost icon-only" aria-label="Previous year" onClick={() => setYear(year - 1)}><Icon name="chevron-left" size={18} /></button>
        <h2 className="sky-nav-title">{year}</h2>
        <button type="button" className="btn ghost icon-only" aria-label="Next year" onClick={() => setYear(year + 1)}><Icon name="chevron-right" size={18} /></button>
        <button type="button" className="btn sm" onClick={() => setYear(thisYear)} disabled={year === thisYear}>This year</button>
      </div>
      {error && <div className="callout danger"><Icon name="alert" size={15} /><span>{error}</span></div>}
      <p className="muted small">A planet is retrograde when, seen from Earth, it seems to move backwards for a while. Astrologers read it as a time to review rather than to start.</p>
      {!shown ? <Loading /> : shown.length === 0 ? <Empty title="No retrogrades" icon="refresh" /> : <>
        <Card className="retro-card">
          <div className="retro-grid">
            <div className="retro-months"><span />{MONTH_LETTERS.map((m, i) => <span key={i}>{m}</span>)}</div>
            {bodies.map(b => (
              <div key={b} className="retro-lane">
                <span className="retro-name"><span className="glyph">{BODY_GLYPH[b]}</span>{BODY_NAME[b]}</span>
                <span className="retro-track">
                  {shown.filter(p => p.body === b).map(p => {
                    const a = pct(p.start)
                    const e = pct(p.end)
                    return <span key={p.start} className={`retro-bar r-${b}`} style={{ left: `${a}%`, width: `${Math.max(0.8, e - a)}%` }} title={`${BODY_NAME[b]} retrograde ${fmtDate(dateOf(p.start))} – ${fmtDate(dateOf(p.end))}`} />
                  })}
                  {todayPct != null && <span className="retro-today" style={{ left: `${todayPct}%` }} />}
                </span>
              </div>
            ))}
          </div>
        </Card>
        <div className="retro-list">
          {bodies.map(b => {
            const ps = shown.filter(p => p.body === b)
            return (
              <Card key={b} className="retro-item" title={<span className="retro-item-title"><span className="glyph">{BODY_GLYPH[b]}</span>{BODY_NAME[b]} retrograde</span>}>
                <ul className="retro-dates">
                  {ps.map(p => {
                    const active = now >= new Date(p.start).getTime() && now <= new Date(p.end).getTime()
                    return <li key={p.start}><span className="tnum">{fmtDate(dateOf(p.start))} – {fmtDate(dateOf(p.end))}</span><span className="muted small"> · starts in {signName(p.sign)}</span>{active && <span className="badge accent retro-now">now</span>}</li>
                  })}
                </ul>
                <p className="sky-text muted">{ps[0].text}</p>
              </Card>
            )
          })}
        </div>
      </>}
    </div>
  )
}

// ---------- coming up ----------

const KINDS: { id: SkyEvent['kind']; label: string }[] = [
  { id: 'phase', label: 'Moon phases' }, { id: 'eclipse', label: 'Eclipses' }, { id: 'ingress', label: 'Sign changes' }, { id: 'station', label: 'Retrogrades' },
]

function EventsView() {
  const today = todayISO()
  const to = addDays(today, 89)
  const { data, error } = useAsync(() => api.get<{ events: SkyEvent[] }>(`${V1}/sky/events?from=${today}&to=${to}`), [today])
  const [kinds, setKinds] = useState<Set<string>>(() => new Set(KINDS.map(k => k.id)))
  const toggle = (k: string) => setKinds(s => { const n = new Set(s); if (n.has(k)) n.delete(k); else n.add(k); return n })
  const groups = useMemo(() => {
    const out: [string, SkyEvent[]][] = []
    for (const e of data?.events || []) {
      if (!kinds.has(e.kind)) continue
      const m = e.date.slice(0, 7)
      if (!out.length || out[out.length - 1][0] !== m) out.push([m, []])
      out[out.length - 1][1].push(e)
    }
    return out
  }, [data, kinds])
  return (
    <div className="stack loose">
      <div className="chips" role="group" aria-label="Show">
        {KINDS.map(k => <button key={k.id} type="button" className={`chip ${kinds.has(k.id) ? 'active' : ''}`} aria-pressed={kinds.has(k.id)} onClick={() => toggle(k.id)}>{k.label}</button>)}
      </div>
      <p className="muted small">The next three months, {fmtDate(today, false)} to {fmtDate(to)}.</p>
      {error && <div className="callout danger"><Icon name="alert" size={15} /><span>{error}</span></div>}
      {!data ? <Loading /> : groups.length === 0 ? <Empty title="Nothing to show" icon="list">Switch on more kinds above.</Empty> : groups.map(([m, es]) => (
        <section key={m} className="sky-month">
          <h3 className="sky-month-h">{monthLabel(m)}</h3>
          <Card flush><EventList events={es} today={today} /></Card>
        </section>
      ))}
    </div>
  )
}
