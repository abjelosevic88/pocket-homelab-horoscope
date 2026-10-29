// Coming up: the transits of the next weeks or months — when a moving planet
// reaches an important point of the birth chart, and how long it lasts.
import { useEffect, useMemo, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { api, V1 } from '../lib/api'
import { AREA_NAME } from '../lib/astro'
import { addDays, fmtDate, fmtRange, todayISO } from '../lib/format'
import type { TransitHit, TransitList } from '../lib/types'
import { PersonPicker, usePeople } from '../components/CheckIn'
import { Icon } from '../components/icons'
import { openCreate } from '../components/Layout'
import { Empty, Loading, PageHeader, Segmented, Spinner, Switch, errMsg } from '../components/ui'
import { ToneChip } from './Today'

const MONTH = (ym: string) => { const [y, m] = ym.split('-').map(Number); return new Date(y, m - 1, 1).toLocaleDateString(undefined, { month: 'long', year: 'numeric' }) }
const day = (iso: string) => iso.slice(0, 10)
const ms = (d: string) => Date.parse(d.length === 10 ? d + 'T00:00:00Z' : d)

function TransitRow({ h, from, to }: { h: TransitHit; from: string; to: string }) {
  const [open, setOpen] = useState(false)
  const a = ms(from), b = ms(to), span = Math.max(1, b - a)
  const pct = (iso: string) => Math.max(0, Math.min(100, ((ms(iso) - a) / span) * 100))
  const left = pct(h.start), right = pct(h.end)
  const exacts = h.exact.map(day)
  const when = exacts.length
    ? `Exact ${exacts.map((d, i) => fmtDate(d, i === exacts.length - 1 || d.slice(0, 4) !== exacts[exacts.length - 1].slice(0, 4))).join(', ')}`
    : ''
  return (
    <li className={`tr-row ${open ? 'open' : ''}`}>
      <button type="button" className="tr-head" aria-expanded={open} onClick={() => setOpen(o => !o)}>
        <span className="grow tr-title">
          <b>{h.title}</b>
          <span className="muted small">{fmtRange(day(h.start), day(h.end))}{when && ` · ${when}`}</span>
        </span>
        {h.active && <span className="badge accent">Now</span>}
        <ToneChip tone={h.tone} />
        <Icon name={open ? 'chevron-up' : 'chevron-down'} size={15} className="muted" />
      </button>
      <div className="tr-bar" title={`${fmtDate(day(h.start))} – ${fmtDate(day(h.end))}`}>
        <span className="tr-span" style={{ left: `${left}%`, width: `${Math.max(1, right - left)}%` }} />
        {h.exact.filter(e => ms(e) >= a && ms(e) <= b).map(e => <span key={e} className="tr-exact" style={{ left: `${pct(e)}%` }} title={`Exact ${fmtDate(day(e))}`} />)}
      </div>
      {open && (
        <div className="tr-body">
          <p>{h.text}</p>
          {h.areas.length > 0 && <p className="muted small">Touches {h.areas.map(x => (AREA_NAME[x] || x).toLowerCase()).join(' and ')}. {h.weight >= 4 ? 'A slow, strong influence.' : h.weight <= 2 ? 'A quick, light influence.' : ''}</p>}
        </div>
      )}
    </li>
  )
}

export default function Transits() {
  const [sp, setSp] = useSearchParams()
  const person = sp.get('person') || ''
  const people = usePeople()
  const [days, setDays] = useState(90)
  const [slow, setSlow] = useState(false)
  const [data, setData] = useState<TransitList | null>(null)
  const [loading, setLoading] = useState(true)
  const [err, setErr] = useState('')

  useEffect(() => {
    const q = new URLSearchParams({ days: String(days) })
    if (person) q.set('person', person)
    if (slow) q.set('slow', '1')
    let live = true
    setLoading(true)
    api.get<TransitList>(`${V1}/transits?${q}`).then(d => { if (live) { setData(d); setErr('') } }).catch(e => { if (live) setErr(errMsg(e)) }).finally(() => { if (live) setLoading(false) })
    return () => { live = false }
  }, [person, days, slow])

  const from = data?.from || todayISO()
  const to = addDays(from, data?.days || days)
  const groups = useMemo(() => {
    const out = new Map<string, TransitHit[]>()
    for (const h of data?.transits || []) {
      const next = (h.exact || []).find(e => day(e) >= from)
      const key = next ? day(next).slice(0, 7) : h.exact.length ? 'now' : day(h.start) >= from ? day(h.start).slice(0, 7) : 'now'
      if (!out.has(key)) out.set(key, [])
      out.get(key)!.push(h)
    }
    return [...out.entries()].sort(([a], [b]) => (a === 'now' ? -1 : b === 'now' ? 1 : a.localeCompare(b)))
  }, [data, from])

  const setPerson = (v: string) => { const n = new URLSearchParams(sp); if (v) n.set('person', v); else n.delete('person'); setSp(n) }
  return (
    <div className="tr">
      <PageHeader title="Coming up" sub="When a moving planet reaches an important point of the birth chart. The slow planets (Jupiter and beyond) shape months; the fast ones, a few days." />
      <div className="tr-toolbar">
        {people && people.length > 1 && <PersonPicker className="tr-person" people={people} value={person} onChange={setPerson} />}
        <Segmented label="How far ahead" value={days} onChange={setDays} options={[{ value: 30, label: '30 days' }, { value: 90, label: '3 months' }, { value: 365, label: 'A year' }]} />
        <label className="tr-slow"><Switch checked={slow} onChange={setSlow} label="Only slow planets" size="sm" /><span className="small">Only slow planets</span></label>
        <span className="tr-spin" aria-hidden>{loading && data && <Spinner size={13} />}</span>
      </div>
      {!data && loading && <Loading />}
      {!data && !loading && err && <Empty title="Could not load the transits" icon="alert">{err}</Empty>}
      {data?.needs_me && (
        <Empty title="Add your birth details first" icon="sun">
          <p>Transits are measured against a birth chart.</p>
          <button type="button" className="btn primary" onClick={() => openCreate('person')}><Icon name="plus" size={15} />Add your birth details</button>
        </Empty>
      )}
      {data && !data.needs_me && groups.length === 0 && <Empty title="Nothing big in this stretch" icon="moon">No planet reaches an important point of the chart in these days. Try a longer stretch.</Empty>}
      {data && !data.needs_me && groups.length > 0 && (
        <div className="stack">
          <div className="tr-legend muted xsmall"><span>{fmtDate(from)}</span><span className="tr-legend-key"><span className="tr-exact-key" /> exact</span><span>{fmtDate(to)}</span></div>
          {groups.map(([k, hs]) => (
            <section key={k} className="card tr-group">
              <h2 className="tr-group-h">{k === 'now' ? 'Already under way' : MONTH(k)}</h2>
              <ul className="tr-list">{hs.map(h => <TransitRow key={`${h.transit}-${h.type}-${h.natal}-${h.start}`} h={h} from={from} to={to} />)}</ul>
            </section>
          ))}
        </div>
      )}
    </div>
  )
}
