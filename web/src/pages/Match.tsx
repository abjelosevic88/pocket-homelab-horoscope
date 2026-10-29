// Compatibility: two birth charts laid over each other.
import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { api, V1 } from '../lib/api'
import { AREA_ICON, AREA_NAME, ASPECT_COLOR, ASPECT_GLYPH, ELEMENTS, ELEMENT_COLOR, SIGN_GLYPH, scoreColor, scoreWord, signName } from '../lib/astro'
import type { Chart, ChartView, Match, MatchAspect, Person, Position, ReadingResult } from '../lib/types'
import { Ring } from '../components/Chart'
import { ChartWheel } from '../components/ChartWheel'
import { Icon } from '../components/icons'
import { openCreate } from '../components/Layout'
import { Markdown } from '../components/Markdown'
import { Select } from '../components/Select'
import { Card, Empty, Loading, PageHeader, Spinner, errMsg, useToast } from '../components/ui'
import { Big3 } from './People'

const AREAS = ['love', 'communication', 'trust', 'fun', 'growth'] as const

export default function MatchPage() {
  const { toast } = useToast()
  const [sp, setSp] = useSearchParams()
  const [people, setPeople] = useState<Person[] | null>(null)
  const [m, setM] = useState<Match | null>(null)
  const [outer, setOuter] = useState<Position[] | undefined>()
  const [chartA, setChartA] = useState<Chart | null>(null)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const [writing, setWriting] = useState(false)
  const [aiNote, setAiNote] = useState('')
  const [showAll, setShowAll] = useState(false)

  useEffect(() => { api.get<Person[]>(`${V1}/people`).then(setPeople).catch(e => { setPeople([]); toast(errMsg(e), 'err') }) }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const me = people?.find(p => p.is_me)
  const a = Number(sp.get('a')) || me?.id || people?.[0]?.id || 0
  const bRaw = Number(sp.get('b')) || 0
  const b = bRaw && bRaw !== a ? bRaw : people?.find(p => p.id !== a && p.relation === 'partner')?.id || people?.find(p => p.id !== a)?.id || 0
  const setPair = (na: number, nb: number) => {
    const n = new URLSearchParams()
    if (na && na !== me?.id) n.set('a', String(na))
    if (nb) n.set('b', String(nb))
    setSp(n, { replace: true })
  }

  useEffect(() => {
    if (!a || !b || a === b) return
    let live = true
    setBusy(true); setErr(''); setAiNote('')
    Promise.all([api.get<Match>(`${V1}/match?a=${a}&b=${b}`), api.get<ChartView>(`${V1}/people/${a}/chart`), api.get<ChartView>(`${V1}/people/${b}/chart`)])
      .then(([mm, ca, cb]) => { if (live) { setM(mm); setChartA(ca.chart); setOuter(cb.chart.positions) } })
      .catch(e => { if (live) setErr(errMsg(e)) })
      .finally(() => { if (live) setBusy(false) })
    return () => { live = false }
  }, [a, b])

  const write = async () => {
    setWriting(true); setAiNote('')
    try {
      const r = await api.post<ReadingResult>(`${V1}/readings`, { kind: 'match', person_id: a, other_id: b })
      if (r.ok && r.reading) setM(x => (x ? { ...x, reading: r.reading! } : x))
      else if (r.not_configured) setAiNote(r.error || 'No AI is set up.')
      else toast(r.error || 'The reading could not be written', 'err')
    } catch (e) { toast(errMsg(e), 'err') } finally { setWriting(false) }
  }

  if (!people) return <Loading />
  if (people.length < 2) return (
    <div className="stack loose">
      <PageHeader title="Compatibility" />
      <Empty title="Add two people to compare" icon="heart">
        <p>Compatibility lays two birth charts over each other: where you click, and where you rub.</p>
        <button type="button" className="btn primary" onClick={() => openCreate('person')}>Add a person</button>
      </Empty>
    </div>
  )
  const opts = people.map(p => ({ value: String(p.id), label: p.is_me ? `${p.name} (you)` : p.name, hint: signName(p.sun) }))
  const pa = people.find(p => p.id === a), pb = people.find(p => p.id === b)
  const ready = m && m.a.id === a && m.b.id === b
  return (
    <div className="stack loose">
      <PageHeader title="Compatibility" sub="Two birth charts laid over each other." />
      <div className="mt-pick">
        <Select value={String(a)} onChange={v => setPair(Number(v), b === Number(v) ? 0 : b)} options={opts} ariaLabel="First person" />
        <span className="mt-amp" aria-hidden><Icon name="heart" size={16} /></span>
        <Select value={String(b)} onChange={v => setPair(a, Number(v))} options={opts.filter(o => o.value !== String(a))} ariaLabel="Second person" />
      </div>
      {err && <div className="callout danger"><Icon name="alert" size={15} /><div>{err}</div></div>}
      {!ready ? <Loading /> : (
        <div className={`stack loose ${busy ? 'mt-busy' : ''}`}>
          <Card className="mt-hero">
            <div className="mt-hero-in">
              <Ring pct={m.overall} size={108} stroke={9} color={scoreColor(m.overall)}><b>{m.overall}</b><span>{scoreWord(m.overall)}</span></Ring>
              <div className="grow stack tight">
                <div className="mt-names">
                  <span><b>{pa?.name}</b> <Big3 p={m.a} compact /></span>
                  <span><b>{pb?.name}</b> <Big3 p={m.b} compact /></span>
                </div>
                <p className="mt-summary">{m.summary}</p>
              </div>
            </div>
            <div className="mt-areas">
              {AREAS.map(k => (
                <div key={k} className="mt-area">
                  <span className="mt-area-l"><Icon name={AREA_ICON[k]} size={14} />{AREA_NAME[k]}</span>
                  <span className="mt-area-track"><span style={{ width: `${m.scores[k]}%`, background: scoreColor(m.scores[k]) }} /></span>
                  <span className="tnum small">{m.scores[k]}</span>
                </div>
              ))}
            </div>
          </Card>

          <div className="mt-grid">
            <Card title="The two charts" sub={`${pa?.name}'s chart inside, ${pb?.name}'s planets on the outer ring.`}>
              {chartA && <ChartWheel chart={chartA} outer={outer} size={480} />}
            </Card>
            <div className="stack">
              {m.sun_pair && (
                <Card title={<span className="row"><span className="glyph">{SIGN_GLYPH[m.a.sun]}</span> {signName(m.a.sun)} & <span className="glyph">{SIGN_GLYPH[m.b.sun]}</span> {signName(m.b.sun)}</span>}>
                  <p className="mt-p">{m.sun_pair}</p>
                </Card>
              )}
              <Card title="Elements" sub="How much fire, earth, air and water each chart has.">
                <div className="mt-els">
                  {ELEMENTS.map(el => {
                    const [x, y] = m.elements[el] || [0, 0]
                    return (
                      <div key={el} className="mt-el">
                        <span className="mt-el-l">{el[0].toUpperCase() + el.slice(1)}</span>
                        <span className="mt-el-bars">
                          <span className="mt-el-bar a" style={{ width: `${x}%`, background: ELEMENT_COLOR[el] }} title={`${pa?.name}: ${Math.round(x)}%`} />
                          <span className="mt-el-bar b" style={{ width: `${y}%`, background: ELEMENT_COLOR[el] }} title={`${pb?.name}: ${Math.round(y)}%`} />
                        </span>
                      </div>
                    )
                  })}
                  <div className="mt-el-key muted xs"><span><i className="mt-key a" />{pa?.name}</span><span><i className="mt-key b" />{pb?.name}</span></div>
                </div>
              </Card>
            </div>
          </div>

          <div className="two-col mt-sc">
            <Card title="What flows" sub="The easiest links between you.">
              {m.strengths.length ? <AspectList list={m.strengths} /> : <p className="muted small">No strong easy links — your bond is built more than given.</p>}
            </Card>
            <Card title="Where you rub" sub="The links that ask for patience.">
              {m.challenges.length ? <AspectList list={m.challenges} /> : <p className="muted small">No strong tense links between your charts.</p>}
            </Card>
          </div>

          <Card title="A written reading" actions={!m.reading ? <button type="button" className="btn primary sm" onClick={write} disabled={writing}>{writing ? <Spinner /> : <Icon name="sparkles" size={14} />}Write it</button> : <Link to={`/readings/${m.reading.id}`} className="btn sm">Open</Link>}>
            {m.reading ? <Markdown text={m.reading.body} /> : <p className="muted small">{aiNote ? <>{aiNote} Set it up in <Link to="/settings/ai">Settings → AI</Link>.</> : 'Put the whole comparison into a few paragraphs of plain words.'}</p>}
          </Card>

          <Card title={`All ${m.aspects.length} links`} actions={<button type="button" className="btn sm" onClick={() => setShowAll(s => !s)} aria-expanded={showAll}>{showAll ? 'Hide' : 'Show'}</button>}>
            {showAll ? <AspectList list={m.aspects} /> : <p className="muted small">Every aspect between {pa?.name}'s planets and {pb?.name}'s chart, closest first.</p>}
          </Card>
        </div>
      )}
    </div>
  )
}

function AspectList({ list }: { list: MatchAspect[] }) {
  const [open, setOpen] = useState(-1)
  return (
    <div className="mt-as">
      {list.map((x, i) => (
        <div key={i} className={`mt-as-row ${open === i ? 'open' : ''}`}>
          <button type="button" className="mt-as-head" onClick={() => setOpen(open === i ? -1 : i)} aria-expanded={open === i}>
            <span className="glyph mt-as-glyph" style={{ color: ASPECT_COLOR[x.type] }}>{ASPECT_GLYPH[x.type]}</span>
            <span className="grow"><b>{x.title}</b><span className="muted xs"> · {x.orb.toFixed(1)}°</span></span>
            <Icon name={open === i ? 'chevron-up' : 'chevron-down'} size={14} className="muted" />
          </button>
          {open === i && <p className="mt-as-text">{x.text}</p>}
        </div>
      ))}
    </div>
  )
}
