// Written readings: the ones the AI wrote for you (your day, a chart, a match,
// the year ahead, an answer) — the kept ones stay, the rest go after 60 days.
import { useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api, V1 } from '../lib/api'
import { useApp } from '../lib/app-context'
import { fmtDate, fmtDateTime } from '../lib/format'
import type { Reading, ReadingKind } from '../lib/types'
import { Icon } from '../components/icons'
import { Markdown } from '../components/Markdown'
import { announceChange, useCrumb } from '../components/Layout'
import { Confirm, CoreHint, Empty, Loading, PageHeader, Segmented, Switch, copyText, useAsync, useToast } from '../components/ui'

const KIND_NAME: Record<ReadingKind, string> = { daily: 'Day', natal: 'Birth chart', match: 'Compatibility', year: 'Year ahead', question: 'Question' }
const KIND_ICON: Record<ReadingKind, string> = { daily: 'sun', natal: 'circle', match: 'heart', year: 'calendar', question: 'help' }

/** The first plain sentence of a markdown body, for list previews. */
function preview(body: string): string {
  const text = body.split('\n').filter(l => l.trim() && !/^\s*(#|---)/.test(l)).join(' ')
  return text.replace(/\*\*|\*|_|\[([^\]]+)\]\([^)]+\)/g, (m, a) => a || '').slice(0, 220)
}

export default function Readings() {
  const navigate = useNavigate()
  const [kind, setKind] = useState<'' | ReadingKind>('')
  const [kept, setKept] = useState(false)
  const { data, error, loading } = useAsync(() => api.get<Reading[]>(`${V1}/readings?kind=${kind}&saved=${kept ? 1 : 0}`), [kind, kept])
  return (
    <div className="page readings-page">
      <PageHeader title="Readings" sub="What the AI wrote for you. Keep the ones you want; the others are cleared after 60 days." />
      <div className="readings-toolbar">
        <Segmented size="sm" className="readings-kinds" value={kind} onChange={setKind} label="Kind" options={[
          { value: '', label: 'All' }, { value: 'daily', label: 'Days' }, { value: 'natal', label: 'Charts' }, { value: 'match', label: 'Matches' }, { value: 'year', label: 'Years' }, { value: 'question', label: 'Questions' }]} />
        <label className="readings-kept"><Switch checked={kept} onChange={setKept} label="Kept only" size="sm" /><span className="small">Kept only</span></label>
      </div>
      {error && <div className="callout danger"><Icon name="alert" size={15} /><span>{error}</span></div>}
      {!data && loading && <Loading />}
      {data && data.length === 0 && (
        <Empty title={kept ? 'No kept readings' : 'No readings yet'} icon="book-open">
          Ask for one on <Link to="/">Today</Link>, on someone’s chart or on a <Link to="/match">compatibility</Link> page — or with Add → Ask about your chart.
        </Empty>
      )}
      {data && data.length > 0 && (
        <div className="reading-list">
          {data.map(r => (
            <button type="button" key={r.id} className="reading-row" onClick={() => navigate(`/readings/${r.id}`)}>
              <span className="reading-ico"><Icon name={KIND_ICON[r.kind] || 'book-open'} size={16} /></span>
              <span className="reading-main">
                <span className="reading-title">
                  <span className="ellipsis">{r.title || KIND_NAME[r.kind]}</span>
                  {r.saved && <Icon name="bookmark" size={13} className="reading-kept" title="Kept" />}
                </span>
                <span className="reading-meta muted small">
                  <span className="chip-static">{KIND_NAME[r.kind] || r.kind}</span>
                  {[r.person_name, r.other_name].filter(Boolean).join(' & ')}
                  {(r.person_name || r.other_name) && ' · '}{fmtDate(r.created_at.slice(0, 10))}
                </span>
                <span className="reading-preview muted small">{preview(r.body)}</span>
              </span>
              <Icon name="chevron-right" size={16} className="muted" />
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

export function ReadingPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { toast } = useToast()
  const { settings } = useApp()
  const { data: r, setData, error, loading } = useAsync(() => api.get<Reading>(`${V1}/readings/${id}`), [id])
  const [del, setDel] = useState(false)
  const [busy, setBusy] = useState<'' | 'keep' | 'note'>('')
  const [notesErr, setNotesErr] = useState<{ msg: string; notConnected: boolean } | null>(null)
  useCrumb(r?.title)
  if (error) return <div className="page"><div className="callout danger"><Icon name="alert" size={15} /><span>{error === 'not found' ? 'This reading is gone.' : error}</span></div><p className="mt"><Link to="/readings">All readings</Link></p></div>
  if (!r || loading && !r) return <Loading />
  const coreOn = !!settings && settings.core_configured && settings.core_registered !== false

  const keep = async () => {
    setBusy('keep')
    try { const n = await api.put<Reading>(`${V1}/readings/${r.id}`, { saved: !r.saved }); setData(n); toast(n.saved ? 'Kept' : 'No longer kept'); announceChange() }
    catch (e) { toast((e as Error).message, 'err') } finally { setBusy('') }
  }
  const toNotes = async () => {
    setBusy('note'); setNotesErr(null)
    try {
      const res = await api.post<{ ok: boolean; url?: string; error?: string; not_connected?: boolean }>(`${V1}/readings/${r.id}/note`)
      if (res.ok) {
        setData({ ...r, saved: true })
        toast(res.url ? <>Saved to Pocket Notes · <a href={res.url} target="_blank" rel="noreferrer">Open</a></> : 'Saved to Pocket Notes')
      } else setNotesErr({ msg: res.error || 'failed', notConnected: !!res.not_connected })
    } catch (e) { toast((e as Error).message, 'err') } finally { setBusy('') }
  }
  const copy = async () => { if (await copyText(`${r.title}\n\n${r.body}`)) toast('Copied'); else toast('Could not copy', 'err') }
  const remove = async () => {
    setDel(false)
    try { await api.del(`${V1}/readings/${r.id}`); announceChange(); toast('Reading deleted'); navigate('/readings') } catch (e) { toast((e as Error).message, 'err') }
  }

  const who = [r.person_name, r.other_name].filter(Boolean).join(' & ')
  return (
    <div className="page reading-page">
      <div className="reading-head">
        <span className="chip-static"><Icon name={KIND_ICON[r.kind] || 'book-open'} size={13} />{KIND_NAME[r.kind] || r.kind}</span>
        <h1>{r.title}</h1>
        <p className="muted small">
          {who && <>{r.person_id ? <Link to={`/people/${r.person_id}`}>{r.person_name}</Link> : r.person_name}{r.other_name && <> & {r.other_id ? <Link to={`/people/${r.other_id}`}>{r.other_name}</Link> : r.other_name}</>} · </>}
          Written {fmtDateTime(r.created_at)}{r.model && r.model !== 'demo' ? ` · ${r.model === 'pocket' ? 'Pocket AI' : 'OpenAI'}` : ''}
        </p>
      </div>
      <div className="reading-actions">
        <button type="button" className={`btn sm ${r.saved ? 'active-soft' : ''}`} onClick={keep} disabled={!!busy}><Icon name="bookmark" size={14} />{r.saved ? 'Kept' : 'Keep'}</button>
        <button type="button" className="btn sm" onClick={toNotes} disabled={!!busy}><Icon name="note" size={14} />{busy === 'note' ? 'Saving…' : 'Save to Pocket Notes'}</button>
        <button type="button" className="btn sm" onClick={copy}><Icon name="copy" size={14} />Copy</button>
        <span className="grow" />
        <button type="button" className="btn sm ghost danger-text" onClick={() => setDel(true)}><Icon name="trash" size={14} />Delete</button>
      </div>
      {notesErr && (
        <div className="callout warn reading-notes-err">
          <Icon name="alert" size={15} />
          <div>
            {notesErr.msg}.{notesErr.notConnected && <> Connect Pocket Notes in <Link to="/settings/apps">Settings → Connected apps</Link>.</>}
            {notesErr.notConnected && !coreOn && <CoreHint />}
          </div>
        </div>
      )}
      <article className="reading-body card"><Markdown text={r.body} /></article>
      <p className="muted small reading-foot">Astrology is a mirror for reflection, not a forecast of events.</p>
      {del && <Confirm title="Delete this reading?" message="It is removed from here; a copy saved to Pocket Notes stays there." confirmLabel="Delete" onCancel={() => setDel(false)} onConfirm={remove} />}
    </div>
  )
}
