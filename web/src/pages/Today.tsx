// Today: your day's horoscope, worked out from the sky against your birth
// chart — scores, the Moon, the transits that touch you, tips and the week.
import { useCallback, useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { api, V1 } from '../lib/api'
import { AREA_ICON, AREA_NAME, PHASE_EMOJI, PHASE_NAME, SIGN_GLYPH, bodyName, clockOf, scoreColor, scoreWord, signName } from '../lib/astro'
import { addDays, fmtClock, fmtDate, todayISO, weekday } from '../lib/format'
import type { Area, DayReading, DayTransit, ReadingResult } from '../lib/types'
import { Ring } from '../components/Chart'
import { CheckInModal, EnergyDots, MoodFace, MOOD_WORD, PersonPicker, usePeople } from '../components/CheckIn'
import { Icon } from '../components/icons'
import { announceChange, openCreate, useOnChange } from '../components/Layout'
import { Markdown } from '../components/Markdown'
import { Card, Empty, Loading, Spinner, errMsg, modalOpen, useToast } from '../components/ui'

const AREAS: Area[] = ['love', 'work', 'mood', 'energy']

export function ordinal(n: number): string {
  const s = ['th', 'st', 'nd', 'rd'], v = n % 100
  return n + (s[(v - 20) % 10] || s[v] || s[0])
}

/** The tone of a transit in a word and a colour. */
export function ToneChip({ tone }: { tone: number }) {
  if (tone > 0) return <span className="badge ok">Easy</span>
  if (tone < 0) return <span className="badge warn">Tense</span>
  return <span className="badge">Mixed</span>
}

function orbWords(t: DayTransit): string {
  if (t.orb < 0.3) return 'exact today'
  return t.applying ? 'getting stronger' : 'easing off'
}

export function TransitItem({ t }: { t: DayTransit }) {
  return (
    <li className="td-transit">
      <div className="td-transit-head">
        <b className="grow">{t.title}</b>
        <ToneChip tone={t.tone} />
      </div>
      <p className="td-transit-text">{t.text}</p>
      <div className="td-transit-meta muted small">
        <span>{orbWords(t)}</span>
        {t.areas.length > 0 && <span>· touches {t.areas.map(a => (AREA_NAME[a] || a).toLowerCase()).join(' and ')}</span>}
      </div>
    </li>
  )
}

export default function Today() {
  const [sp, setSp] = useSearchParams()
  const date = sp.get('date') || ''
  const person = sp.get('person') || ''
  const people = usePeople()
  const { toast } = useToast()
  const [data, setData] = useState<DayReading | null>(null)
  const [err, setErr] = useState('')
  const [loading, setLoading] = useState(true)
  const [writing, setWriting] = useState(false)
  const [aiHint, setAiHint] = useState('')
  const [checkin, setCheckin] = useState(false)
  const [showRetro, setShowRetro] = useState(false)

  const load = useCallback(() => {
    const q = new URLSearchParams()
    if (date) q.set('date', date)
    if (person) q.set('person', person)
    setLoading(true)
    api.get<DayReading>(`${V1}/today${q.toString() ? '?' + q : ''}`)
      .then(d => { setData(d); setErr('') })
      .catch(e => setErr(errMsg(e)))
      .finally(() => setLoading(false))
  }, [date, person])
  useEffect(() => { load(); setAiHint('') }, [load])
  useOnChange(load)

  const go = useCallback((d: string) => {
    const n = new URLSearchParams(sp)
    if (!d || d === todayISO()) n.delete('date'); else n.set('date', d)
    setSp(n)
  }, [sp, setSp])
  const shown = data?.date || date || todayISO()

  useEffect(() => {
    const h = (e: KeyboardEvent) => {
      const t = e.target as HTMLElement
      if (e.metaKey || e.ctrlKey || e.altKey || t.closest('input, textarea, [contenteditable="true"]') || modalOpen()) return
      if (e.key === 'ArrowLeft') { e.preventDefault(); go(addDays(shown, -1)) }
      if (e.key === 'ArrowRight') { e.preventDefault(); go(addDays(shown, 1)) }
    }
    window.addEventListener('keydown', h)
    return () => window.removeEventListener('keydown', h)
  }, [go, shown])

  const setPerson = (v: string) => { const n = new URLSearchParams(sp); if (v) n.set('person', v); else n.delete('person'); setSp(n) }

  const write = async () => {
    if (!data) return
    setWriting(true); setAiHint('')
    try {
      const r = await api.post<ReadingResult>(`${V1}/readings`, { kind: 'daily', person_id: data.person_id, date: data.date })
      if (r.ok && r.reading) setData(d => d && { ...d, reading: r.reading! })
      else if (r.not_configured) setAiHint(r.error || 'No AI is set up yet.')
      else toast(r.error || 'The reading could not be written', 'err')
    } catch (e) { toast(errMsg(e), 'err') } finally { setWriting(false) }
  }

  if (!data && loading) return <Loading />
  if (!data) return <Empty title="Today's horoscope did not load" icon="alert">{err}</Empty>
  if (data.needs_me) {
    return (
      <div className="td-welcome">
        <Empty title="Your horoscope starts with your birth chart" icon="sun">
          <p>Add your birth date, and if you know them, the time and the town you were born in. Pocket Horoscope works out where the planets were, and from then on writes your day from the real sky.</p>
          <div className="row td-welcome-actions">
            <button type="button" className="btn primary" onClick={() => openCreate('person')}><Icon name="plus" size={15} />Add your birth details</button>
            <button type="button" className="btn" onClick={() => openCreate('import')}><Icon name="upload" size={15} />Import birth data</button>
          </div>
        </Empty>
      </div>
    )
  }

  const isToday = data.date === todayISO()
  const me = people?.find(p => p.id === data.person_id)
  const isMe = me ? me.is_me : !person
  const first = data.person_name.split(' ')[0]
  return (
    <div className="td">
      <section className="card td-hero">
        <div className="td-nav">
          <button type="button" className="btn ghost icon-only" aria-label="The day before" title="The day before (←)" onClick={() => go(addDays(shown, -1))}><Icon name="chevron-left" size={18} /></button>
          <div className="td-date">
            <b>{isToday ? 'Today' : weekday(data.date, true)}</b>
            <span className="muted small">{isToday ? `${weekday(data.date, true)}, ${fmtDate(data.date)}` : fmtDate(data.date)}</span>
          </div>
          <button type="button" className="btn ghost icon-only" aria-label="The day after" title="The day after (→)" onClick={() => go(addDays(shown, 1))}><Icon name="chevron-right" size={18} /></button>
          <button type="button" className="btn sm td-today-btn" disabled={isToday} onClick={() => go('')}>Today</button>
          <span className="td-spin" aria-hidden>{loading && <Spinner size={13} />}</span>
          <span className="spacer" />
          {people && people.length > 1 && <PersonPicker className="td-person" people={people} value={person} onChange={setPerson} />}
        </div>
        {data.birthday && (
          <div className="callout info td-bday"><Icon name="cake" size={15} /><div><b>Happy birthday{isMe ? '' : `, ${first}`}!</b> The Sun is back where it was when {isMe ? 'you were' : `${first} was`} born: a new personal year starts today.</div></div>
        )}
        <div className="td-hero-main">
          <div className="grow">
            {!isMe && <div className="muted small td-for">{data.person_name}'s day</div>}
            <h1 className="td-headline">{data.headline}</h1>
            <p className="td-summary">{data.summary}</p>
          </div>
          <Ring pct={data.overall} size={84} stroke={8} color={scoreColor(data.overall)}>
            <span className="td-overall"><b>{data.overall}</b><span className="muted">{scoreWord(data.overall)}</span></span>
          </Ring>
        </div>
        <div className="td-scores">
          {AREAS.map(a => {
            const v = data.scores[a]
            return (
              <div key={a} className="td-score" title={`${AREA_NAME[a]}: ${v} of 100`}>
                <div className="td-score-head"><Icon name={AREA_ICON[a]} size={14} /><span>{AREA_NAME[a]}</span><span className="spacer" /><b className="tnum">{v}</b></div>
                <div className="td-bar"><span style={{ width: `${v}%`, background: scoreColor(v) }} /></div>
                <span className="muted xsmall">{scoreWord(v)}</span>
              </div>
            )
          })}
        </div>
      </section>

      <div className="td-grid">
        <div className="td-col">
          <Card title={isMe ? 'Today’s sky for you' : `Today’s sky for ${first}`} sub="The planets in the sky right now, and the points of the birth chart they touch.">
            {data.transits.length === 0
              ? <p className="muted">A quiet day: no planet is close to an important point of the chart. Good for routine and rest.</p>
              : <ul className="td-transits">{data.transits.map(t => <TransitItem key={`${t.transit}-${t.type}-${t.natal}`} t={t} />)}</ul>}
          </Card>

          <div className="td-tips">
            <Card title="Try"><ul className="td-tip-list ok-list">{data.tips.do.map(t => <li key={t}><Icon name="check" size={14} />{t}</li>)}</ul></Card>
            <Card title="Go easy on"><ul className="td-tip-list avoid-list">{data.tips.avoid.map(t => <li key={t}><Icon name="x" size={14} />{t}</li>)}</ul></Card>
          </div>

          <Card title="Your reading" sub="A few paragraphs written by AI from the chart and the day’s sky." className="td-reading">
            {data.reading
              ? <Markdown text={data.reading.body} />
              : (
                <div className="td-reading-empty">
                  <button type="button" className="btn" onClick={write} disabled={writing}>{writing ? <><Spinner size={13} /> Writing…</> : <><Icon name="sparkles" size={15} />Write {isToday ? 'today’s' : 'this day’s'} reading</>}</button>
                  {aiHint && <p className="muted small">{aiHint} <Link to="/settings/ai">Set up AI in Settings → AI</Link>.</p>}
                </div>
              )}
          </Card>
        </div>

        <div className="td-col">
          <Card title="The Moon" className="td-moon">
            <div className="td-moon-head">
              <span className="td-moon-emoji" aria-hidden>{PHASE_EMOJI[data.moon.phase]}</span>
              <div>
                <b>{PHASE_NAME[data.moon.phase]} in {signName(data.moon.sign)} <span className="glyph">{SIGN_GLYPH[data.moon.sign]}</span></b>
                <div className="muted small">{Math.round(data.moon.illumination * 100)}% lit{data.moon.house > 0 && ` · moving through ${isMe ? 'your' : `${first}’s`} ${ordinal(data.moon.house)} house`}</div>
              </div>
            </div>
            <p>{data.moon.text}</p>
            <p className="muted">{data.moon.phase_text}</p>
            {data.moon.void_from && data.moon.void_to && (
              <div className="callout td-void"><Icon name="clock" size={15} /><div>
                <b>Quiet window {fmtClock(clockOf(data.moon.void_from))}–{fmtClock(clockOf(data.moon.void_to))}</b>
                {data.moon.void_from.slice(0, 10) !== data.moon.void_to.slice(0, 10) && <span className="muted"> (until {weekday(data.moon.void_to.slice(0, 10))})</span>}
                <div className="muted small">The Moon is “void of course”: it makes no more contacts before changing sign. Better for finishing things than for starting them.</div>
              </div></div>
            )}
          </Card>

          {isMe && (
            <Card title={isToday ? 'Today’s check-in' : 'Check-in'} actions={<button type="button" className="btn sm" onClick={() => setCheckin(true)}>{data.journal ? 'Edit' : 'Check in'}</button>}>
              {data.journal ? (
                <div className="td-checkin">
                  <div className="row"><MoodFace mood={data.journal.mood} size={26} /><span>{MOOD_WORD[data.journal.mood] || 'Mood not set'}</span><span className="spacer" /><EnergyDots energy={data.journal.energy} /></div>
                  {data.journal.note && <p className="td-checkin-note">{data.journal.note}</p>}
                </div>
              ) : <p className="muted">How {isToday ? 'is' : 'was'} the day? A short check-in lets you see, over time, how the Moon and your days line up.</p>}
            </Card>
          )}

          {data.week.length > 0 && (
            <Card title="The week ahead">
              <div className="td-week">
                {data.week.map(w => (
                  <button key={w.date} type="button" className={`td-day ${w.date === data.date ? 'active' : ''}`} onClick={() => go(w.date)} title={`${fmtDate(w.date)}: ${w.overall}, Moon in ${signName(w.moon)}`}>
                    <span className="muted xsmall">{weekday(w.date)}</span>
                    <b className="tnum" style={{ color: scoreColor(w.overall) }}>{w.overall}</b>
                    <span className="td-day-bar"><span style={{ height: `${w.overall}%`, background: scoreColor(w.overall) }} /></span>
                    <span className="glyph td-day-moon" aria-label={`Moon in ${signName(w.moon)}`}>{SIGN_GLYPH[w.moon]}</span>
                  </button>
                ))}
              </div>
            </Card>
          )}

          {data.events.length > 0 && (
            <Card title="In the sky today">
              <ul className="td-events">
                {data.events.map(e => (
                  <li key={e.kind + e.time + (e.body || '')}>
                    <span className="td-ev-ico glyph" aria-hidden>{e.icon}</span>
                    <div className="grow"><b>{e.title}</b> <span className="muted small">{fmtClock(clockOf(e.time))}</span><div className="muted small">{e.text}</div></div>
                  </li>
                ))}
              </ul>
            </Card>
          )}

          {data.retrograde.length > 0 && (
            <Card title="Retrograde now" actions={<button type="button" className="btn ghost sm" aria-expanded={showRetro} onClick={() => setShowRetro(s => !s)}>{showRetro ? 'Less' : 'What it means'}</button>}>
              <div className="chips">{data.retrograde.map(r => <span key={r.body} className="chip td-chip-static">{r.name} <span className="muted">in {signName(r.sign)}</span></span>)}</div>
              {showRetro && (
                <ul className="td-retro-list">
                  {data.retrograde.map(r => <li key={r.body}><b>{bodyName(r.body)}</b> — <span className="muted">{r.text}</span></li>)}
                </ul>
              )}
            </Card>
          )}
        </div>
      </div>
      {checkin && <CheckInModal date={data.date} onClose={() => setCheckin(false)} onSaved={() => { setCheckin(false); announceChange() }} />}
    </div>
  )
}
