// Bring people in: birth data from a file (AAF from astro.com or Solar Fire,
// or CSV), or the birthdays in Pocket Contacts.
import { useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, V1 } from '../lib/api'
import { RELATION_NAME, SIGN_GLYPH, signName } from '../lib/astro'
import { fmtDate } from '../lib/format'
import { useApp } from '../lib/app-context'
import { Icon } from './icons'
import { Select } from './Select'
import { CoreHint, Field, Loading, Modal, Tabs, errMsg } from './ui'

type ImportResult = { added: number; skipped: number; problems: string[] }
type Candidate = { id: number; name: string; birthday: string; has_year: boolean; sun: string; added: boolean }
type ContactsRes = { ok: boolean; contacts?: Candidate[]; error?: string; not_connected?: boolean }

export function ImportModal({ initial, onClose, onDone }: { initial: 'file' | 'contacts'; onClose: () => void; onDone: () => void }) {
  const [tab, setTab] = useState<'file' | 'contacts'>(initial)
  return (
    <Modal title="Bring people in" onClose={onClose} size="lg">
      <div className="stack">
        <Tabs tabs={[{ id: 'file', label: 'File or paste', icon: 'upload' }, { id: 'contacts', label: 'Pocket Contacts', icon: 'cake' }]} value={tab} onChange={setTab} />
        {tab === 'file' ? <FileImport onDone={onDone} /> : <ContactsImport onDone={onDone} />}
      </div>
    </Modal>
  )
}

function FileImport({ onDone }: { onDone: () => void }) {
  const [text, setText] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const [res, setRes] = useState<ImportResult | null>(null)
  const file = useRef<HTMLInputElement>(null)
  const read = async (fl: File | undefined) => { if (fl) setText(await fl.text()) }
  const run = async () => {
    setBusy(true); setErr(''); setRes(null)
    try { setRes(await api.post<ImportResult>(`${V1}/people/import`, { text })) } catch (e) { setErr(errMsg(e)) } finally { setBusy(false) }
  }
  return (
    <div className="stack">
      <p className="muted small">
        <b>AAF</b> is what astro.com, Solar Fire and most astrology programs export (lines starting <code>#A93:</code> and <code>#B93:</code>).
        A <b>CSV</b> needs a header row with <code>name, date, time, place, lat, lon, tz, relation</code> — date as 1990-04-21, time as 07:45, tz as Europe/Sarajevo or UTC+01:00. Only name and date are required.
      </p>
      <div className="row wrap">
        <button type="button" className="btn" onClick={() => file.current?.click()}><Icon name="upload" size={15} />Choose a file</button>
        <input ref={file} type="file" accept=".aaf,.csv,.txt,text/plain,text/csv" hidden onChange={e => read(e.target.files?.[0])} />
        <span className="muted small">or paste below</span>
      </div>
      <textarea className="imp-text" rows={8} value={text} onChange={e => setText(e.target.value)} spellCheck={false}
        placeholder={'name,date,time,place,lat,lon,tz,relation\nAna Horvat,1990-04-21,07:45,Zagreb,45.81,15.98,Europe/Zagreb,friend'} />
      {err && <div className="callout danger"><Icon name="alert" size={15} /><div>{err}</div></div>}
      {res && (
        <div className={`callout ${res.added ? 'ok' : 'info'}`}>
          <Icon name={res.added ? 'check-circle' : 'info'} size={15} />
          <div>
            <b>{res.added} added</b>{res.skipped ? `, ${res.skipped} already here` : ''}.
            {res.problems.length > 0 && <ul className="imp-problems">{res.problems.slice(0, 8).map((p, i) => <li key={i}>{p}</li>)}{res.problems.length > 8 && <li>…and {res.problems.length - 8} more</li>}</ul>}
          </div>
        </div>
      )}
      <div className="form-actions">
        {res && res.added > 0 && <button type="button" className="btn" onClick={onDone}>See people</button>}
        <button type="button" className="btn primary" disabled={busy || !text.trim()} onClick={run}>Import</button>
      </div>
    </div>
  )
}

function ContactsImport({ onDone }: { onDone: () => void }) {
  const { settings } = useApp()
  const [data, setData] = useState<ContactsRes | null>(null)
  const [pick, setPick] = useState<Set<number>>(new Set())
  const [relation, setRelation] = useState('friend')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const [q, setQ] = useState('')
  useEffect(() => { api.get<ContactsRes>(`${V1}/people/contacts`).then(setData).catch(e => setData({ ok: false, error: errMsg(e) })) }, [])
  if (!data) return <Loading />
  if (!data.ok) {
    const coreOn = !!settings?.core_configured && settings.core_registered !== false
    return (
      <div className="stack">
        <div className="callout warn"><Icon name="alert" size={15} /><div>{data.error || 'Pocket Contacts is not connected.'}</div></div>
        {data.not_connected && <p className="muted small">Connect Pocket Contacts in <Link to="/settings/apps">Settings → Connected apps</Link>; through Pocket Core there is nothing to paste.</p>}
        {data.not_connected && !coreOn && <CoreHint />}
      </div>
    )
  }
  const list = (data.contacts || []).filter(c => !q.trim() || c.name.toLowerCase().includes(q.trim().toLowerCase()))
  const pickable = (c: Candidate) => c.has_year && !c.added
  const toggle = (id: number) => setPick(s => { const n = new Set(s); n.has(id) ? n.delete(id) : n.add(id); return n })
  const all = list.filter(pickable)
  const add = async () => {
    setBusy(true); setErr('')
    try {
      const r = await api.post<{ ok: boolean; added?: number; skipped?: number; error?: string }>(`${V1}/people/contacts`, { ids: [...pick], relation })
      if (!r.ok) setErr(r.error || 'Could not add them')
      else onDone()
    } catch (e) { setErr(errMsg(e)) } finally { setBusy(false) }
  }
  if (!data.contacts?.length) return <p className="muted">None of your contacts has a birthday yet. Add birthdays in Pocket Contacts and they show up here.</p>
  return (
    <div className="stack">
      <p className="muted small">Contacts with a birthday. A chart needs the year; the birth time and place can be added later on each person's page.</p>
      <div className="row wrap">
        <div className="search-box grow"><Icon name="search" size={15} /><input value={q} onChange={e => setQ(e.target.value)} placeholder="Filter contacts" /></div>
        <button type="button" className="btn sm" onClick={() => setPick(new Set(pick.size === all.length ? [] : all.map(c => c.id)))} disabled={!all.length}>{pick.size === all.length && all.length ? 'Select none' : 'Select all'}</button>
      </div>
      <div className="imp-list" role="list">
        {list.map(c => {
          const why = c.added ? 'Already added' : !c.has_year ? 'No birth year' : ''
          return (
            <label key={c.id} role="listitem" className={`imp-row ${why ? 'off' : ''}`} title={why || undefined}>
              <input type="checkbox" checked={pick.has(c.id)} disabled={!!why} onChange={() => toggle(c.id)} />
              <span className="imp-glyph glyph">{c.sun ? SIGN_GLYPH[c.sun] : '·'}</span>
              <span className="grow ellipsis">{c.name}</span>
              <span className="muted small imp-when">{why || `${c.has_year ? fmtDate(c.birthday) : c.birthday.replace(/^--/, '')} · ${signName(c.sun)}`}</span>
            </label>
          )
        })}
      </div>
      <div className="row wrap imp-foot">
        <Field label="Add them as" group className="imp-rel">
          <Select value={relation} onChange={setRelation} ariaLabel="Relation" options={['partner', 'family', 'friend', 'colleague', 'other'].map(r => ({ value: r, label: RELATION_NAME[r] }))} />
        </Field>
        <span className="spacer" />
        <button type="button" className="btn primary" disabled={busy || !pick.size} onClick={add}>Add {pick.size || ''} {pick.size === 1 ? 'person' : 'people'}</button>
      </div>
      {err && <div className="callout danger"><Icon name="alert" size={15} /><div>{err}</div></div>}
    </div>
  )
}
