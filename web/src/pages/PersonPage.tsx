// One person's birth chart: the wheel, the Big Three, planets, houses,
// aspects and the readings written about them.
import { useEffect, useRef, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api, ApiError, V1 } from '../lib/api'
import { ASPECT_COLOR, ASPECT_GLYPH, ASPECT_NAME, BODY_GLYPH, ELEMENTS, ELEMENT_COLOR, MODALITY_NAME, RELATION_NAME, SIGN_GLYPH, fmtDeg, signName } from '../lib/astro'
import { fmtDate } from '../lib/format'
import type { AspectView, ChartView, Placement, Reading, ReadingResult } from '../lib/types'
import { ChartWheel } from '../components/ChartWheel'
import { Icon } from '../components/icons'
import { announceChange, useCrumb } from '../components/Layout'
import { Markdown } from '../components/Markdown'
import { PersonModal } from '../components/PersonForm'
import { PlanetsView } from '../components/PlanetsView'
import { HousesView } from '../components/HousesView'
import { AspectsView } from '../components/AspectsView'
import { Card, Confirm, Empty, Loading, Spinner, Tabs, errMsg, useIsMobile, useToast } from '../components/ui'

type Tab = 'overview' | 'planets' | 'houses' | 'aspects' | 'readings'
const TABS: { id: Tab; label: string }[] = [{ id: 'overview', label: 'Overview' }, { id: 'planets', label: 'Planets' }, { id: 'houses', label: 'Houses' }, { id: 'aspects', label: 'Aspects' }, { id: 'readings', label: 'Readings' }]
const ORD = ['', '1st', '2nd', '3rd', '4th', '5th', '6th', '7th', '8th', '9th', '10th', '11th', '12th']
const PCT = (n: number) => `${Math.round(n)}%`

export default function PersonPage() {
  const { id, tab: tabParam } = useParams()
  const navigate = useNavigate()
  const { toast } = useToast()
  const [view, setView] = useState<ChartView | null>(null)
  const [missing, setMissing] = useState(false)
  const [edit, setEdit] = useState(false)
  const [del, setDel] = useState(false)
  const [open, setOpen] = useState<string>('')
  const [writing, setWriting] = useState<'' | 'natal' | 'year'>('')
  const [aiNote, setAiNote] = useState('')
  const tab: Tab = (TABS.some(t => t.id === tabParam) ? tabParam : 'overview') as Tab
  const setTab = (t: Tab) => navigate(`/people/${id}${t === 'overview' ? '' : '/' + t}`, { replace: true })

  const load = () => api.get<ChartView>(`${V1}/people/${id}/chart`).then(v => { setView(v); setMissing(false) }).catch(e => { if (e instanceof ApiError && e.status === 404) setMissing(true); else toast(errMsg(e), 'err') })
  useEffect(() => { setView(null); load() }, [id]) // eslint-disable-line react-hooks/exhaustive-deps
  useCrumb(view?.person.name)

  const pick = (key: string) => { setOpen(key); setTab('planets') }
  useEffect(() => {
    if (tab === 'planets' && open) document.getElementById(`pl-${open}`)?.scrollIntoView({ block: 'center', behavior: 'smooth' })
  }, [tab, open])

  const write = async (kind: 'natal' | 'year') => {
    setWriting(kind); setAiNote('')
    try {
      const r = await api.post<ReadingResult>(`${V1}/readings`, { kind, person_id: Number(id) })
      if (r.ok && r.reading) navigate(`/readings/${r.reading.id}`)
      else if (r.not_configured) setAiNote(r.error || 'No AI is set up.')
      else toast(r.error || 'The reading could not be written', 'err')
    } catch (e) { toast(errMsg(e), 'err') } finally { setWriting('') }
  }
  const remove = async () => {
    try { await api.del(`${V1}/people/${id}`); announceChange(); navigate('/people') } catch (e) { toast(errMsg(e), 'err') }
  }

  if (missing) return <Empty title="This person is not here" icon="user"><Link to="/people">Back to people</Link></Empty>
  if (!view) return <Loading />
  const p = view.person
  const c = view.chart
  return (
    <div className="stack loose">
      <div className="pc-head">
        <span className="pp-avatar lg" style={{ background: p.color || 'var(--accent)' }}><span className="glyph">{p.sun ? SIGN_GLYPH[p.sun] : '·'}</span></span>
        <div className="grow">
          <h1 className="pc-name">{p.name}</h1>
          <div className="muted small">
            {RELATION_NAME[p.relation]} · born {fmtDate(p.birth_date)}{p.birth_time ? ` at ${p.birth_time}` : ', time unknown'}{p.place ? ` in ${p.place}${p.country ? `, ${p.country}` : ''}` : ''}
          </div>
          {p.chart_note && <div className="small pc-note"><Icon name="info" size={13} /> {p.chart_note}</div>}
        </div>
      </div>
      <div className="pc-actions">
        <button type="button" className="btn" onClick={() => setEdit(true)}><Icon name="edit" size={15} />Edit</button>
        {!p.is_me && <Link className="btn" to={`/match?b=${p.id}`}><Icon name="heart" size={15} />Compare</Link>}
        <button type="button" className="btn" onClick={() => write('year')} disabled={!!writing}>{writing === 'year' ? <Spinner /> : <Icon name="calendar" size={15} />}Year ahead</button>
        <button type="button" className="btn primary" onClick={() => write('natal')} disabled={!!writing}>{writing === 'natal' ? <Spinner /> : <Icon name="sparkles" size={15} />}Write a reading</button>
        <button type="button" className="btn ghost icon-only danger-text" onClick={() => setDel(true)} aria-label="Delete" title="Delete"><Icon name="trash" size={16} /></button>
      </div>
      {aiNote && <p className="muted small pc-ai-note"><Icon name="info" size={13} /> {aiNote} Set it up in <Link to="/settings/ai">Settings → AI</Link>.</p>}
      <Tabs tabs={TABS} value={tab} onChange={setTab} />

      {tab === 'overview' && (
        <div className="pc-overview">
          <div className="pc-wheel"><ChartWheel chart={c} onPick={pick} size={520} /></div>
          <div className="stack">
            {view.big_three.map(b => (
              <Card key={b.body} className="pc-big">
                <div className="pc-big-head">
                  <span className="glyph pc-big-glyph">{SIGN_GLYPH[b.sign]}</span>
                  <div><div className="muted xs">{b.body === 'sun' ? 'Sun sign — who you are' : b.body === 'moon' ? 'Moon sign — how you feel' : 'Rising sign — how you come across'}</div><b>{signName(b.sign)}</b> <span className="muted small">{fmtDeg(b.degree)}{b.house ? ` · ${ORD[b.house]} house` : ''}</span></div>
                </div>
                <p className="pc-text">{b.sign_text}</p>
              </Card>
            ))}
            {!c.time_known && <p className="muted small">Add a birth time and place to see the rising sign and the houses.</p>}
            <Card title="Elements and modes" sub={view.balance.text}>
              <div className="pc-balance">
                {ELEMENTS.map(el => (
                  <div key={el} className="pc-bal-row"><span className="pc-bal-l">{el[0].toUpperCase() + el.slice(1)}</span><span className="pc-bal-track"><span style={{ width: PCT(view.balance.elements[el] || 0), background: ELEMENT_COLOR[el] }} /></span><span className="tnum small muted">{PCT(view.balance.elements[el] || 0)}</span></div>
                ))}
                {['cardinal', 'fixed', 'mutable'].map(m => (
                  <div key={m} className="pc-bal-row"><span className="pc-bal-l">{MODALITY_NAME[m]}</span><span className="pc-bal-track"><span style={{ width: PCT(view.balance.modalities[m] || 0), background: 'var(--accent)' }} /></span><span className="tnum small muted">{PCT(view.balance.modalities[m] || 0)}</span></div>
                ))}
              </div>
            </Card>
          </div>
        </div>
      )}

      {tab === 'planets' && (
        <PlanetsView placements={view.placements} points={view.points.filter(x => c.time_known || x.body === 'southnode')} open={open} onOpen={setOpen} />
      )}

      {tab === 'houses' && (c.houses ? (
        <HousesView houses={view.houses} placements={view.placements} asc={c.houses.asc} system={`${c.houses.system[0].toUpperCase() + c.houses.system.slice(1)} houses${c.houses.fallback ? ` (${c.houses.fallback})` : ''}`} />
      ) : <Empty title="No houses without a birth time" icon="clock">The houses depend on the exact time and place of birth. <button type="button" className="link-btn" onClick={() => setEdit(true)}>Add them</button>.</Empty>)}

      {tab === 'aspects' && <AspectsView aspects={view.aspects} placements={view.placements} />}

      {tab === 'readings' && <PersonReadings id={p.id} onWrite={() => write('natal')} writing={writing === 'natal'} />}

      {edit && <PersonModal person={p} onClose={() => setEdit(false)} onSaved={() => { setEdit(false); announceChange(); load() }} />}
      {del && <Confirm title={`Delete ${p.name}?`} message="Their chart and the readings about them are removed. This cannot be undone." confirmLabel="Delete" onConfirm={remove} onCancel={() => setDel(false)} />}
    </div>
  )
}

function PersonReadings({ id, onWrite, writing }: { id: number; onWrite: () => void; writing: boolean }) {
  const [list, setList] = useState<Reading[] | null>(null)
  useEffect(() => { api.get<Reading[]>(`${V1}/readings?person=${id}`).then(setList).catch(() => setList([])) }, [id])
  if (!list) return <Loading />
  if (!list.length) return (
    <Empty title="No readings yet" icon="book-open">
      <p>A written reading puts the whole chart into a few pages of plain words.</p>
      <button type="button" className="btn primary" onClick={onWrite} disabled={writing}>{writing ? <Spinner /> : <Icon name="sparkles" size={15} />}Write a reading</button>
    </Empty>
  )
  return (
    <div className="stack">
      {list.map(r => (
        <Link key={r.id} to={`/readings/${r.id}`} className="pc-reading">
          <div className="row"><b className="grow ellipsis">{r.title}</b>{r.saved && <Icon name="bookmark" size={14} className="muted" />}<span className="muted small">{fmtDate(r.created_at.slice(0, 10))}</span></div>
          <div className="pc-reading-prev"><Markdown text={r.body.slice(0, 600)} /></div>
        </Link>
      ))}
    </div>
  )
}
