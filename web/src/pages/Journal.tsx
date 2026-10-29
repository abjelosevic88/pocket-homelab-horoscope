// The journal: one check-in a day (mood, energy, a note), each kept with the
// sky of that day.
import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, V1 } from '../lib/api'
import { PHASE_EMOJI, PHASE_NAME, SIGN_GLYPH, signName } from '../lib/astro'
import { addDays, fmtDate, todayISO, weekday } from '../lib/format'
import type { JournalEntry } from '../lib/types'
import { CheckInModal, EnergyDots, MOOD_COLOR, MOOD_WORD, MoodFace } from '../components/CheckIn'
import { Icon } from '../components/icons'
import { announceChange, useOnChange } from '../components/Layout'
import { Confirm, Empty, Loading, PageHeader, errMsg, useToast } from '../components/ui'

const MONTH = (ym: string) => { const [y, m] = ym.split('-').map(Number); return new Date(y, m - 1, 1).toLocaleDateString(undefined, { month: 'long', year: 'numeric' }) }

export default function Journal() {
  const { toast } = useToast()
  const [list, setList] = useState<JournalEntry[] | null>(null)
  const [err, setErr] = useState('')
  const [edit, setEdit] = useState<{ date?: string } | null>(null)
  const [del, setDel] = useState<JournalEntry | null>(null)

  const load = useCallback(() => {
    api.get<JournalEntry[]>(`${V1}/journal?limit=400`).then(l => { setList(l); setErr('') }).catch(e => setErr(errMsg(e)))
  }, [])
  useEffect(load, [load])
  useOnChange(load)

  const byDate = useMemo(() => new Map((list || []).map(e => [e.date, e])), [list])
  const months = useMemo(() => {
    const out = new Map<string, JournalEntry[]>()
    for (const e of list || []) {
      const k = e.date.slice(0, 7)
      if (!out.has(k)) out.set(k, [])
      out.get(k)!.push(e)
    }
    return [...out.entries()]
  }, [list])
  const today = todayISO()
  const strip = Array.from({ length: 30 }, (_, i) => addDays(today, i - 29))

  const remove = async () => {
    if (!del) return
    const e = del
    setDel(null)
    try { await api.del(`${V1}/journal/${e.id}`); toast('Check-in deleted'); announceChange() } catch (x) { toast(errMsg(x), 'err') }
  }

  return (
    <div className="jn">
      <PageHeader title="Journal" sub="One check-in a day. Each one keeps the sky of that day, so you can see how the Moon and your days line up."
        actions={<>
          <Link to="/journal/insights" className="btn"><Icon name="chart" size={15} />You and the Moon</Link>
          <button type="button" className="btn primary" onClick={() => setEdit({})}><Icon name="plus" size={15} />Check in</button>
        </>} />
      {!list && !err && <Loading />}
      {!list && err && <Empty title="The journal did not load" icon="alert">{err}</Empty>}
      {list && list.length === 0 && (
        <Empty title="No check-ins yet" icon="journal">
          <p>Note how your day felt: a mood, your energy and a line or two. After a few weeks you will see whether the Moon has anything to do with it.</p>
          <button type="button" className="btn primary" onClick={() => setEdit({})}><Icon name="plus" size={15} />Your first check-in</button>
        </Empty>
      )}
      {list && list.length > 0 && (
        <>
          <section className="card jn-strip-card" aria-label="The last 30 days">
            <div className="jn-strip-head"><b>The last 30 days</b><span className="muted small">{strip.filter(d => byDate.has(d)).length} check-ins</span></div>
            <div className="jn-strip">
              {strip.map(d => {
                const e = byDate.get(d)
                return (
                  <button key={d} type="button" className={`jn-cell ${d === today ? 'today' : ''}`} style={{ background: e?.mood ? MOOD_COLOR[e.mood] : undefined }}
                    title={`${fmtDate(d)}: ${e ? MOOD_WORD[e.mood] : 'no check-in'}`} aria-label={`${fmtDate(d)}: ${e ? MOOD_WORD[e.mood] : 'no check-in'}`}
                    onClick={() => setEdit({ date: d })} />
                )
              })}
            </div>
            <div className="jn-strip-legend muted xsmall">{[1, 2, 3, 4, 5].map(m => <span key={m}><i style={{ background: MOOD_COLOR[m] }} />{MOOD_WORD[m]}</span>)}</div>
          </section>
          {months.map(([k, es]) => (
            <section key={k} className="card jn-month">
              <h2 className="jn-month-h">{MONTH(k)}<span className="muted small"> · {es.length}</span></h2>
              <ul className="jn-list">
                {es.map(e => (
                  <li key={e.id} className="jn-row">
                    <div className="jn-date"><b>{e.date.slice(8, 10).replace(/^0/, '')}</b><span className="muted xsmall">{weekday(e.date)}</span></div>
                    <div className="jn-mood"><MoodFace mood={e.mood} size={22} /><EnergyDots energy={e.energy} /></div>
                    <div className="grow jn-main">
                      {e.note ? <p className="jn-note">{e.note}</p> : <p className="muted small jn-note">{MOOD_WORD[e.mood]} mood{e.energy ? `, ${e.energy}/5 energy` : ''}</p>}
                      <div className="jn-sky">
                        {e.sky.moon_sign && <span className="jn-sky-moon" title={`${e.sky.phase ? PHASE_NAME[e.sky.phase] : 'Moon'} in ${signName(e.sky.moon_sign)}`}>{e.sky.phase && <span aria-hidden>{PHASE_EMOJI[e.sky.phase]}</span>}<span className="glyph">{SIGN_GLYPH[e.sky.moon_sign]}</span><span className="muted xsmall">Moon in {signName(e.sky.moon_sign)}</span></span>}
                        {e.tags && e.tags.split(',').map(t => t.trim()).filter(Boolean).map(t => <span key={t} className="jn-tag">#{t}</span>)}
                        {e.sky.transits.slice(0, 2).map(t => <span key={t} className="jn-transit">{t}</span>)}
                      </div>
                    </div>
                    <div className="jn-actions">
                      <button type="button" className="btn ghost icon-only sm" aria-label={`Edit the check-in of ${fmtDate(e.date)}`} title="Edit" onClick={() => setEdit({ date: e.date })}><Icon name="edit" size={15} /></button>
                      <button type="button" className="btn ghost icon-only sm" aria-label={`Delete the check-in of ${fmtDate(e.date)}`} title="Delete" onClick={() => setDel(e)}><Icon name="trash" size={15} /></button>
                    </div>
                  </li>
                ))}
              </ul>
            </section>
          ))}
        </>
      )}
      {edit && <CheckInModal date={edit.date} onClose={() => setEdit(null)} onSaved={() => { setEdit(null); announceChange() }} />}
      {del && <Confirm title="Delete this check-in?" message={`The check-in of ${fmtDate(del.date)} goes away for good.`} confirmLabel="Delete" onConfirm={remove} onCancel={() => setDel(null)} />}
    </div>
  )
}
