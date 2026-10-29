// You and the Moon: your check-ins (and, with Pocket Health, your sleep and
// resting heart rate) by the Moon's phase, element and sign.
import { type ReactNode, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, V1 } from '../lib/api'
import { ELEMENT_COLOR, PHASE_EMOJI, PHASE_NAME, SIGN_GLYPH, signName } from '../lib/astro'
import type { Bucket, Insights as InsightsData } from '../lib/types'
import { Icon } from '../components/icons'
import { useOnChange } from '../components/Layout'
import { Card, Empty, Loading, PageHeader, Segmented, Spinner, errMsg } from '../components/ui'

const ELEMENT_NAME: Record<string, string> = { fire: 'Fire', earth: 'Earth', air: 'Air', water: 'Water' }

/** Mood and energy side by side per bucket, on the 1–5 scale. */
function PairBars({ buckets, label, color }: { buckets: Bucket[]; label: (k: string) => ReactNode; color?: (k: string) => string }) {
  return (
    <div className="in-bars" role="table" aria-label="Mood and energy">
      {buckets.map(b => (
        <div key={b.key} className="in-col" role="row" title={b.n ? `${b.key.replace(/_/g, ' ')}: mood ${b.mood}, energy ${b.energy} (${b.n} days)` : `${b.key.replace(/_/g, ' ')}: no check-ins yet`}>
          <div className="in-pair">
            <span className="in-bar mood" style={{ height: `${(b.mood / 5) * 100}%`, background: color?.(b.key) }} aria-label={`mood ${b.mood}`} />
            <span className="in-bar energy" style={{ height: `${(b.energy / 5) * 100}%` }} aria-label={`energy ${b.energy}`} />
          </div>
          <span className="in-lbl">{label(b.key)}</span>
          <span className="in-n muted">{b.n ? b.n : '–'}</span>
        </div>
      ))}
    </div>
  )
}

/** One value per bucket (sleep minutes, heart rate). */
function ValueBars({ buckets, fmt }: { buckets: Bucket[]; fmt: (v: number) => string }) {
  const vals = buckets.map(b => b.value || 0).filter(Boolean)
  const lo = vals.length ? Math.min(...vals) * 0.9 : 0, hi = vals.length ? Math.max(...vals) : 1
  return (
    <div className="in-bars">
      {buckets.map(b => (
        <div key={b.key} className="in-col" title={b.n ? `${PHASE_NAME[b.key]}: ${fmt(b.value || 0)} (${b.n} days)` : `${PHASE_NAME[b.key]}: no data`}>
          <span className="in-val tnum">{b.value ? fmt(b.value) : ''}</span>
          <div className="in-pair"><span className="in-bar single" style={{ height: b.value ? `${Math.max(4, ((b.value - lo) / Math.max(1, hi - lo)) * 100)}%` : 0 }} /></div>
          <span className="in-lbl" aria-hidden>{PHASE_EMOJI[b.key]}</span>
          <span className="in-n muted">{b.n || '–'}</span>
        </div>
      ))}
    </div>
  )
}

const hm = (min: number) => `${Math.floor(min / 60)}h ${String(Math.round(min % 60)).padStart(2, '0')}`

interface HealthMetric { name?: string; unit?: string; avg?: number; n?: number; phases?: Bucket[]; error?: string }

export default function Insights() {
  const [days, setDays] = useState(365)
  const [data, setData] = useState<InsightsData | null>(null)
  const [loading, setLoading] = useState(true)
  const [err, setErr] = useState('')
  const [tick, setTick] = useState(0)
  useOnChange(() => setTick(t => t + 1))

  useEffect(() => {
    let live = true
    setLoading(true)
    api.get<InsightsData>(`${V1}/journal/insights?days=${days}`).then(d => { if (live) { setData(d); setErr('') } }).catch(e => { if (live) setErr(errMsg(e)) }).finally(() => { if (live) setLoading(false) })
    return () => { live = false }
  }, [days, tick])

  const legend = <div className="in-legend muted xsmall"><span><i className="mood" />Mood</span><span><i className="energy" />Energy</span><span>numbers below: days</span></div>
  const health = data?.health
  return (
    <div className="in">
      <PageHeader title="You and the Moon" sub={<>Your check-ins next to the Moon, from the <Link to="/journal">journal</Link>. Averages on a 1 to 5 scale.</>}
        actions={<><span className="tr-spin" aria-hidden>{loading && data && <Spinner size={13} />}</span><Segmented label="Period" value={days} onChange={setDays} options={[{ value: 90, label: '3 months' }, { value: 365, label: 'A year' }, { value: 3660, label: 'All' }]} /></>} />
      {!data && loading && <Loading />}
      {!data && !loading && err && <Empty title="The insights did not load" icon="alert">{err}</Empty>}
      {data && (
        <div className="stack">
          {data.entries < 14 ? (
            <div className="callout info"><Icon name="info" size={15} /><div><b>{data.entries ? `${data.entries} check-in${data.entries === 1 ? '' : 's'} so far.` : 'No check-ins in this period yet.'}</b> Patterns need a few weeks of check-ins (at least 14, better a few months) before they mean anything. <Link to="/journal">Check in</Link> on most days and come back.</div></div>
          ) : (
            <div className="stack tight">
              {data.findings.map(f => <div key={f} className="callout info"><Icon name="sparkles" size={15} /><div>{f}</div></div>)}
              <p className="muted small">From {data.entries} check-ins: mood {data.mood_avg} and energy {data.energy_avg} on average. A pattern is a hint, not a rule — sleep, work and people matter more than the sky.</p>
            </div>
          )}

          <Card title="By the Moon’s phase" sub="From new Moon (dark) through full Moon and back.">
            {legend}
            <PairBars buckets={data.phases} label={k => <span aria-label={PHASE_NAME[k]} title={PHASE_NAME[k]}>{PHASE_EMOJI[k]}</span>} />
          </Card>

          <div className="two-col">
            <Card title="By the Moon’s element" sub="The four kinds of sign the Moon passes through.">
              {legend}
              <PairBars buckets={data.elements} label={k => ELEMENT_NAME[k] || k} color={k => ELEMENT_COLOR[k]} />
            </Card>
            <Card title="By the Moon’s sign" sub="Each sign holds the Moon for about two and a half days.">
              {legend}
              <PairBars buckets={data.signs} label={k => <span className="glyph" aria-label={signName(k)} title={signName(k)}>{SIGN_GLYPH[k]}</span>} />
            </Card>
          </div>

          {health && (
            <Card title="Sleep and heart" sub="From Pocket Health, by the Moon’s phase on each night’s evening.">
              {!health.ok ? (
                <div className="callout"><Icon name="info" size={15} /><div>
                  <span className="muted">{health.error || 'Pocket Health is not connected.'}</span>{' '}
                  {health.not_connected && <Link to="/settings/apps">Connect Pocket Health in Settings → Connected apps</Link>}
                </div></div>
              ) : (
                <div className="two-col">
                  {(['sleep_asleep', 'resting_hr'] as const).map(k => {
                    const m = health[k] as HealthMetric | undefined
                    if (!m) return null
                    if (m.error || !m.phases) return <div key={k} className="callout"><Icon name="info" size={15} /><div className="muted">{k === 'sleep_asleep' ? 'Sleep' : 'Resting heart rate'}: {m.error || 'no data'}</div></div>
                    const fmt = k === 'sleep_asleep' ? hm : (v: number) => `${Math.round(v)}`
                    return (
                      <div key={k} className="in-health">
                        <div className="row"><b className="grow">{m.name}</b><span className="muted small">{m.n ? `avg ${fmt(m.avg || 0)}${k === 'resting_hr' ? ' bpm' : ''} · ${m.n} days` : 'no data yet'}</span></div>
                        <ValueBars buckets={m.phases} fmt={fmt} />
                      </div>
                    )
                  })}
                </div>
              )}
            </Card>
          )}
        </div>
      )}
    </div>
  )
}
