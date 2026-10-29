// The day's check-in (mood, energy, a note) and the small pieces the journal
// pages share: mood faces, energy dots and the person picker.
import { type ReactNode, useEffect, useState } from 'react'
import { api, V1 } from '../lib/api'
import { fmtDate, todayISO, weekday } from '../lib/format'
import { RELATION_NAME } from '../lib/astro'
import type { JournalEntry, Person } from '../lib/types'
import { Select } from './Select'
import { Field, Modal, Spinner, errMsg, useToast } from './ui'

export const MOOD_FACE = ['', '😞', '😕', '😐', '🙂', '😄']
export const MOOD_WORD = ['Not set', 'Low', 'Meh', 'Okay', 'Good', 'Great']
export const ENERGY_WORD = ['Not set', 'Drained', 'Low', 'Okay', 'Good', 'Buzzing']
export const MOOD_COLOR = ['var(--surface-2)', '#e03131', '#f08c00', '#adb5bd', '#37b24d', '#2f9e44']

/** A mood as a face (0 = nothing). */
export function MoodFace({ mood, size = 18 }: { mood: number; size?: number }) {
  if (!mood) return <span className="ci-face-none" style={{ width: size, height: size }} aria-label="Mood not set" />
  return <span className="ci-face" style={{ fontSize: size }} role="img" aria-label={`Mood: ${MOOD_WORD[mood]}`} title={`Mood: ${MOOD_WORD[mood]}`}>{MOOD_FACE[mood]}</span>
}

/** Energy as five small dots. */
export function EnergyDots({ energy }: { energy: number }) {
  return (
    <span className="ci-dots" role="img" aria-label={`Energy: ${ENERGY_WORD[energy] || 'not set'}`} title={`Energy: ${ENERGY_WORD[energy] || 'not set'}`}>
      {[1, 2, 3, 4, 5].map(i => <span key={i} className={i <= energy ? 'on' : ''} />)}
    </span>
  )
}

/** Everyone you keep a chart for (you first). */
export function usePeople() {
  const [people, setPeople] = useState<Person[] | null>(null)
  useEffect(() => { api.get<Person[]>(`${V1}/people`).then(setPeople).catch(() => setPeople([])) }, [])
  return people
}

/** Pick a person ('' = you). */
export function PersonPicker({ people, value, onChange, className = '' }: { people: Person[]; value: string; onChange: (v: string) => void; className?: string }) {
  const me = people.find(p => p.is_me)
  const opts = people.map(p => ({ value: p.is_me ? '' : String(p.id), label: p.is_me ? `${p.name} (you)` : p.name, hint: p.is_me ? undefined : RELATION_NAME[p.relation] }))
  if (!me) opts.unshift({ value: '', label: 'You', hint: undefined })
  return <Select className={className} value={value} onChange={onChange} options={opts} ariaLabel="Whose day" />
}

function Scale({ label, value, onChange, words, render }: { label: string; value: number; onChange: (v: number) => void; words: string[]; render: (n: number) => ReactNode }) {
  return (
    <div className="ci-scale" role="radiogroup" aria-label={label}>
      {[1, 2, 3, 4, 5].map(n => (
        <button key={n} type="button" role="radio" aria-checked={value === n} aria-label={words[n]} className={`ci-opt ${value === n ? 'active' : ''}`}
          onClick={() => onChange(value === n ? 0 : n)}>
          <span className="ci-opt-ico">{render(n)}</span>
          <span className="ci-opt-word">{words[n]}</span>
        </button>
      ))}
    </div>
  )
}

export function CheckInModal({ onClose, onSaved, date }: { onClose: () => void; onSaved: (e: JournalEntry) => void; date?: string }) {
  const { toast } = useToast()
  const [day, setDay] = useState(date || todayISO())
  const [mood, setMood] = useState(0)
  const [energy, setEnergy] = useState(0)
  const [note, setNote] = useState('')
  const [tags, setTags] = useState('')
  const [existing, setExisting] = useState<JournalEntry | null>(null)
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')

  useEffect(() => {
    let live = true
    setLoading(true)
    api.get<JournalEntry[]>(`${V1}/journal?from=${day}&to=${day}`).then(list => {
      if (!live) return
      const e = list[0] || null
      setExisting(e)
      setMood(e?.mood || 0); setEnergy(e?.energy || 0); setNote(e?.note || ''); setTags(e?.tags || '')
    }).catch(() => {}).finally(() => { if (live) setLoading(false) })
    return () => { live = false }
  }, [day])

  const save = async () => {
    if (!mood && !energy && !note.trim()) { setErr('Add a mood, an energy level or a note.'); return }
    setBusy(true); setErr('')
    try {
      const e = await api.post<JournalEntry>(`${V1}/journal`, { date: day, mood, energy, note, tags })
      toast(existing ? 'Check-in updated' : 'Check-in saved')
      onSaved(e)
    } catch (e) { setErr(errMsg(e)) } finally { setBusy(false) }
  }

  const isToday = day === todayISO()
  return (
    <Modal title={isToday ? "Today's check-in" : `Check-in for ${weekday(day, true)}, ${fmtDate(day)}`} onClose={onClose} footer={<>
      <span className="grow muted small">{loading ? <Spinner size={12} /> : existing ? 'You already checked in on this day; saving replaces it.' : ''}</span>
      <button type="button" className="btn" onClick={onClose}>Cancel</button>
      <button type="button" className="btn primary" onClick={save} disabled={busy || loading}>{busy ? 'Saving…' : 'Save'}</button>
    </>}>
      <div className="stack">
        <p className="muted small ci-intro">How did the day feel? Over time Pocket Horoscope compares your days with the Moon and shows you any pattern.</p>
        <Field label="Mood" group><Scale label="Mood" value={mood} onChange={setMood} words={MOOD_WORD} render={n => <span className="ci-big-face">{MOOD_FACE[n]}</span>} /></Field>
        <Field label="Energy" group><Scale label="Energy" value={energy} onChange={setEnergy} words={ENERGY_WORD} render={n => <EnergyDots energy={n} />} /></Field>
        <Field label="Note"><textarea value={note} onChange={e => setNote(e.target.value)} rows={3} placeholder="What happened, what you noticed…" /></Field>
        <div className="two-col">
          <Field label="Tags" help="Separate with commas, e.g. work, family"><input value={tags} onChange={e => setTags(e.target.value)} placeholder="work, sleep" /></Field>
          <Field label="Day"><input type="date" value={day} max={todayISO()} onChange={e => e.target.value && setDay(e.target.value)} /></Field>
        </div>
        {err && <div className="callout danger">{err}</div>}
      </div>
    </Modal>
  )
}
