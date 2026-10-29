// Find a town by name (Open-Meteo's geocoder through the server): the
// result carries the coordinates and the time zone a chart needs.
import { useEffect, useRef, useState } from 'react'
import { api, V1 } from '../lib/api'
import type { Place } from '../lib/types'
import { Icon } from './icons'
import { Spinner, useDebounce, useDismiss } from './ui'

export function placeLabel(p: { name?: string; place?: string; region?: string; country?: string }): string {
  return [p.name ?? p.place, p.region, p.country].filter(Boolean).join(', ')
}

export function PlaceSearch({ value, onPick, placeholder = 'Search a town or city', autoFocus, disabled }: {
  value: string; onPick: (p: Place) => void; placeholder?: string; autoFocus?: boolean; disabled?: boolean
}) {
  const [q, setQ] = useState(value)
  const [open, setOpen] = useState(false)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const [list, setList] = useState<Place[]>([])
  const [hi, setHi] = useState(0)
  const dq = useDebounce(q, 300)
  const box = useRef<HTMLDivElement>(null)
  const typed = useRef(false)
  useEffect(() => { setQ(value) }, [value])
  useDismiss(box, () => setOpen(false), open)
  useEffect(() => {
    if (!typed.current || dq.trim().length < 2) { setList([]); return }
    let live = true
    setBusy(true); setErr('')
    api.get<Place[]>(`${V1}/places?q=${encodeURIComponent(dq.trim())}`)
      .then(r => { if (live) { setList(r); setHi(0); setOpen(true) } })
      .catch(e => { if (live) setErr((e as Error).message) })
      .finally(() => { if (live) setBusy(false) })
    return () => { live = false }
  }, [dq])
  const pick = (p: Place) => { typed.current = false; setQ(placeLabel(p)); setOpen(false); onPick(p) }
  return (
    <div className="place-search" ref={box}>
      <div className="place-input">
        <Icon name="pin" size={15} />
        <input value={q} disabled={disabled} autoFocus={autoFocus} placeholder={placeholder} aria-autocomplete="list" aria-expanded={open}
          onChange={e => { typed.current = true; setQ(e.target.value) }}
          onFocus={() => list.length && setOpen(true)}
          onKeyDown={e => {
            if (!open || !list.length) return
            if (e.key === 'ArrowDown') { e.preventDefault(); setHi(h => Math.min(list.length - 1, h + 1)) }
            else if (e.key === 'ArrowUp') { e.preventDefault(); setHi(h => Math.max(0, h - 1)) }
            else if (e.key === 'Enter') { e.preventDefault(); pick(list[hi]) }
            else if (e.key === 'Escape') setOpen(false)
          }} />
        <span className="place-spin">{busy && <Spinner size={13} />}</span>
      </div>
      {err && <div className="field-error small">{err}</div>}
      {open && list.length > 0 && (
        <div className="place-pop" role="listbox">
          {list.map((p, i) => (
            <button type="button" key={`${p.lat},${p.lon},${i}`} role="option" aria-selected={i === hi} className={`place-opt ${i === hi ? 'hl' : ''}`}
              onMouseEnter={() => setHi(i)} onClick={() => pick(p)}>
              <b>{p.name}</b><span className="muted small">{[p.region, p.country].filter(Boolean).join(', ')} · {p.tz}</span>
            </button>
          ))}
        </div>
      )}
      {open && !busy && dq.trim().length >= 2 && list.length === 0 && !err && <div className="place-pop"><div className="place-none muted small">No place found. Try the nearest larger town.</div></div>}
    </div>
  )
}
