// Add or edit a person: name, how they relate to you, and the birth data
// a chart needs. The Big Three show as you type.
import { useEffect, useState } from 'react'
import { api, V1 } from '../lib/api'
import { RELATION_NAME, SIGN_GLYPH, signName } from '../lib/astro'
import type { Person, Place, Relation } from '../lib/types'
import { Icon } from './icons'
import { PlaceSearch, placeLabel } from './PlaceSearch'
import { Select } from './Select'
import { Field, Modal, Switch, errMsg, useDebounce } from './ui'

const RELATIONS: Relation[] = ['me', 'partner', 'family', 'friend', 'colleague', 'other']
const SWATCHES = ['#6a3fb5', '#be123c', '#e8590c', '#f59f00', '#16a34a', '#0f766e', '#2563eb', '#c026d3', '#64748b']

interface Form {
  name: string; relation: Relation; birth_date: string; birth_time: string; time_unknown: boolean
  place: string; country: string; lat: string; lon: string; tz: string
  color: string; notes: string; favorite: boolean
}

function fromPerson(p?: Person): Form {
  return {
    name: p?.name || '', relation: p?.relation || 'friend', birth_date: p?.birth_date || '', birth_time: p?.birth_time || '', time_unknown: !!p && !p.birth_time,
    place: p?.place || '', country: p?.country || '', lat: p && (p.lat || p.lon) ? String(p.lat) : '', lon: p && (p.lat || p.lon) ? String(p.lon) : '', tz: p?.tz || '',
    color: p?.color || SWATCHES[0], notes: p?.notes || '', favorite: !!p?.favorite,
  }
}

function body(f: Form) {
  return {
    name: f.name.trim(), relation: f.relation, is_me: f.relation === 'me', birth_date: f.birth_date, birth_time: f.time_unknown ? '' : f.birth_time,
    place: f.place.trim(), country: f.country.trim(), lat: Number(f.lat) || 0, lon: Number(f.lon) || 0, tz: f.tz.trim(),
    color: f.color, notes: f.notes, favorite: f.favorite,
  }
}

type Preview = { sun: string; moon: string; rising: string; chart_note?: string }

export function PersonModal({ person, onClose, onSaved }: { person?: Person; onClose: () => void; onSaved: (p: Person) => void }) {
  const [f, setF] = useState<Form>(() => fromPerson(person))
  const [byHand, setByHand] = useState(false)
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  const [preview, setPreview] = useState<Preview | null>(null)
  const set = (p: Partial<Form>) => setF(x => ({ ...x, ...p }))

  // the Big Three as you type
  const key = JSON.stringify([f.birth_date, f.time_unknown ? '' : f.birth_time, f.lat, f.lon, f.tz])
  const dkey = useDebounce(key, 400)
  useEffect(() => {
    if (!/^\d{4}-\d{2}-\d{2}$/.test(f.birth_date)) { setPreview(null); return }
    let live = true
    api.post<Preview>(`${V1}/people/preview`, { ...body(f), name: f.name.trim() || 'Preview' })
      .then(r => { if (live) setPreview(r) })
      .catch(() => { if (live) setPreview(null) })
    return () => { live = false }
  }, [dkey]) // eslint-disable-line react-hooks/exhaustive-deps

  const pickPlace = (p: Place) => set({ place: p.name, country: p.country, lat: String(p.lat), lon: String(p.lon), tz: p.tz })

  const save = async () => {
    setErr('')
    if (!f.name.trim()) { setErr('Add a name.'); return }
    if (!f.birth_date) { setErr('Add the birth date.'); return }
    setBusy(true)
    try {
      const saved = person ? await api.put<Person>(`${V1}/people/${person.id}`, body(f)) : await api.post<Person>(`${V1}/people`, body(f))
      onSaved(saved)
    } catch (e) {
      setErr(errMsg(e))
    } finally {
      setBusy(false)
    }
  }

  const hasPlace = !!f.tz && (!!Number(f.lat) || !!Number(f.lon))
  return (
    <Modal title={person ? `Edit ${person.name}` : 'Add a person'} onClose={onClose} size="lg" footer={<>
      <button type="button" className="btn" onClick={onClose}>Cancel</button>
      <button type="button" className="btn primary" onClick={save} disabled={busy}>{person ? 'Save' : 'Add'}</button>
    </>}>
      <form className="stack pf" onSubmit={e => { e.preventDefault(); save() }}>
        <div className="pf-big3" aria-live="polite">
          {(['sun', 'moon', 'rising'] as const).map(k => {
            const v = preview?.[k] || ''
            return (
              <div key={k} className="pf-big3-item">
                <span className="pf-big3-glyph glyph">{v ? SIGN_GLYPH[v] : '–'}</span>
                <span className="pf-big3-sign">{v ? signName(v) : k === 'rising' && preview ? 'Needs the time' : '…'}</span>
                <span className="muted xs">{k === 'sun' ? 'Sun' : k === 'moon' ? 'Moon' : 'Rising'}</span>
              </div>
            )
          })}
        </div>
        <div className="two-col">
          <Field label="Name"><input value={f.name} onChange={e => set({ name: e.target.value })} autoFocus placeholder="Ana Horvat" maxLength={120} /></Field>
          <Field label="Who they are to you" group>
            <Select value={f.relation} onChange={v => set({ relation: v as Relation })} ariaLabel="Relation"
              options={RELATIONS.map(r => ({ value: r, label: r === 'me' ? 'This is me' : RELATION_NAME[r] }))} />
          </Field>
        </div>
        <div className="two-col">
          <Field label="Birth date"><input type="date" value={f.birth_date} min="1800-01-01" max="2199-12-31" onChange={e => set({ birth_date: e.target.value })} /></Field>
          <Field label="Birth time" help={f.time_unknown ? 'Without the time there is no rising sign or houses.' : 'As on the birth certificate, local time.'}>
            <div className="row">
              <input type="time" className="grow" value={f.time_unknown ? '' : f.birth_time} disabled={f.time_unknown} onChange={e => set({ birth_time: e.target.value })} />
              <label className="row pf-unknown"><Switch size="sm" checked={f.time_unknown} onChange={v => set({ time_unknown: v })} label="I don't know the time" /><span className="small">Don't know</span></label>
            </div>
          </Field>
        </div>
        <Field label="Birth place" group help={hasPlace ? <>{placeLabel({ place: f.place, country: f.country })} · {Number(f.lat).toFixed(2)}°, {Number(f.lon).toFixed(2)}° · {f.tz}</> : 'The town is enough; the time zone and coordinates come with it.'}>
          <PlaceSearch value={placeLabel({ place: f.place, country: f.country })} onPick={pickPlace} />
        </Field>
        <div>
          <button type="button" className="link-btn small" onClick={() => setByHand(b => !b)} aria-expanded={byHand}>
            <Icon name={byHand ? 'chevron-down' : 'chevron-right'} size={13} /> Enter coordinates by hand
          </button>
          {byHand && (
            <div className="pf-hand">
              <Field label="Latitude" help="North is positive"><input inputMode="decimal" value={f.lat} onChange={e => set({ lat: e.target.value })} placeholder="43.85" /></Field>
              <Field label="Longitude" help="East is positive"><input inputMode="decimal" value={f.lon} onChange={e => set({ lon: e.target.value })} placeholder="18.36" /></Field>
              <Field label="Time zone" help="Europe/Sarajevo or UTC+01:00"><input value={f.tz} onChange={e => set({ tz: e.target.value })} placeholder="Europe/Sarajevo" /></Field>
              <Field label="Place name"><input value={f.place} onChange={e => set({ place: e.target.value })} /></Field>
            </div>
          )}
        </div>
        <div className="two-col">
          <Field label="Colour" group>
            <div className="pf-swatches" role="radiogroup" aria-label="Colour">
              {SWATCHES.map(c => <button key={c} type="button" role="radio" aria-checked={f.color === c} aria-label={c} className={`pf-swatch ${f.color === c ? 'on' : ''}`} style={{ background: c }} onClick={() => set({ color: c })} />)}
            </div>
          </Field>
          <Field label="Favourite" group>
            <label className="row"><Switch checked={f.favorite} onChange={v => set({ favorite: v })} label="Favourite" /><span className="small muted">Shown near the top</span></label>
          </Field>
        </div>
        <Field label="Notes"><textarea rows={3} value={f.notes} onChange={e => set({ notes: e.target.value })} placeholder="Where the birth time came from, anything to remember" /></Field>
        {err && <div className="callout danger"><Icon name="alert" size={15} /><div>{err}</div></div>}
        <button type="submit" hidden />
      </form>
    </Modal>
  )
}
