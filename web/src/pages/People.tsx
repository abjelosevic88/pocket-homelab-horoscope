// Everyone whose chart you keep: you first, then by how they relate to you.
import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, V1 } from '../lib/api'
import { RELATION_NAME, SIGN_GLYPH, signName } from '../lib/astro'
import { fmtDate } from '../lib/format'
import type { Person } from '../lib/types'
import { Icon } from '../components/icons'
import { openCreate, useOnChange } from '../components/Layout'
import { Empty, Loading, PageHeader, errMsg, useToast } from '../components/ui'

const GROUPS: [string, string][] = [['partner', 'Partner'], ['family', 'Family'], ['friend', 'Friends'], ['colleague', 'Colleagues'], ['other', 'Others']]

export function Big3({ p, compact }: { p: Pick<Person, 'sun' | 'moon' | 'rising'>; compact?: boolean }) {
  const items: [string, string][] = [['Sun', p.sun], ['Moon', p.moon], ['Rising', p.rising]]
  return (
    <span className={`big3 ${compact ? 'compact' : ''}`}>
      {items.map(([k, v]) => (
        <span key={k} className="big3-item" title={v ? `${k} in ${signName(v)}` : `${k}: needs a birth time`}>
          <span className="glyph big3-glyph">{v ? SIGN_GLYPH[v] : '–'}</span>
          {!compact && <span className="big3-text"><b>{v ? signName(v) : '—'}</b><span className="muted xs">{k}</span></span>}
        </span>
      ))}
    </span>
  )
}

function birthdayText(p: Person): string {
  if (p.days_to_birthday === 0) return 'Birthday today'
  if (p.days_to_birthday === 1) return 'Birthday tomorrow'
  if (p.days_to_birthday <= 30) return `Birthday in ${p.days_to_birthday} days`
  return ''
}

function PersonRow({ p, onFav }: { p: Person; onFav: (p: Person) => void }) {
  const bd = birthdayText(p)
  return (
    <div className="pp-row">
      <Link to={`/people/${p.id}`} className="pp-row-main">
        <span className="pp-avatar" style={{ background: p.color || 'var(--accent)' }}><span className="glyph">{p.sun ? SIGN_GLYPH[p.sun] : '·'}</span></span>
        <span className="pp-text">
          <span className="pp-name ellipsis">{p.name}</span>
          <span className="muted small ellipsis">{p.age >= 0 ? `${p.age} · ` : ''}{fmtDate(p.birth_date)}{p.place ? ` · ${p.place}` : ''}{bd && <span className="pp-bd"> · {bd}</span>}</span>
        </span>
        <Big3 p={p} compact />
      </Link>
      <button type="button" className={`btn ghost icon-only sm fav ${p.favorite ? 'on' : ''}`} onClick={() => onFav(p)} aria-label={p.favorite ? 'Remove from favourites' : 'Add to favourites'} title="Favourite"><Icon name="star" size={15} /></button>
    </div>
  )
}

export default function People() {
  const { toast } = useToast()
  const [people, setPeople] = useState<Person[] | null>(null)
  const [q, setQ] = useState('')
  const load = () => api.get<Person[]>(`${V1}/people`).then(setPeople).catch(e => { setPeople([]); toast(errMsg(e), 'err') })
  useEffect(() => { load() }, []) // eslint-disable-line react-hooks/exhaustive-deps
  useOnChange(load)

  const fav = async (p: Person) => {
    setPeople(ps => ps?.map(x => (x.id === p.id ? { ...x, favorite: !p.favorite } : x)) || null)
    try { await api.put(`${V1}/people/${p.id}`, { favorite: !p.favorite }) } catch (e) { toast(errMsg(e), 'err'); load() }
  }

  const shown = useMemo(() => {
    const s = q.trim().toLowerCase()
    return (people || []).filter(p => !s || `${p.name} ${p.place} ${signName(p.sun)} ${RELATION_NAME[p.relation]}`.toLowerCase().includes(s))
  }, [people, q])
  const me = shown.find(p => p.is_me)
  const others = shown.filter(p => !p.is_me)
  const favs = others.filter(p => p.favorite)

  const actions = (
    <>
      <button type="button" className="btn" onClick={() => openCreate('import')}><Icon name="upload" size={15} />Import</button>
      <button type="button" className="btn" onClick={() => openCreate('contacts')}><Icon name="cake" size={15} />From Contacts</button>
      <button type="button" className="btn primary" onClick={() => openCreate('person')}><Icon name="plus" size={15} />Add person</button>
    </>
  )
  if (!people) return <Loading />
  return (
    <div className="stack loose">
      <PageHeader title="People & charts" sub="Birth charts for you and the people you care about." actions={actions} />
      {people.length === 0 ? (
        <Empty title="No one here yet" icon="users">
          <p>Start with yourself: your birth date, and the time and place if you know them.</p>
          <div className="row wrap pp-empty-actions"><button type="button" className="btn primary" onClick={() => openCreate('person')}>Add yourself</button><button type="button" className="btn" onClick={() => openCreate('contacts')}>From Pocket Contacts</button></div>
        </Empty>
      ) : (
        <>
          <div className="search-box pp-search"><Icon name="search" size={15} /><input value={q} onChange={e => setQ(e.target.value)} placeholder="Search by name, place or sign" aria-label="Search people" /></div>
          {!q && !people.some(p => p.is_me) && (
            <button type="button" className="pp-me-cta" onClick={() => openCreate('person')}>
              <Icon name="user" size={18} /><span><b>Add yourself</b><span className="muted small">Your own chart makes Today, Coming up and Compatibility personal.</span></span><Icon name="chevron-right" size={16} />
            </button>
          )}
          {me && (
            <Link to={`/people/${me.id}`} className="pp-me">
              <span className="pp-avatar lg" style={{ background: me.color || 'var(--accent)' }}><span className="glyph">{me.sun ? SIGN_GLYPH[me.sun] : '·'}</span></span>
              <span className="pp-text">
                <span className="muted xs pp-you">You</span>
                <span className="pp-name">{me.name}</span>
                <span className="muted small">{fmtDate(me.birth_date)}{me.birth_time ? ` at ${me.birth_time}` : ''}{me.place ? ` · ${me.place}` : ''}</span>
              </span>
              <Big3 p={me} />
            </Link>
          )}
          {favs.length > 0 && !q && (
            <section className="pp-group"><h3 className="pp-h">Favourites</h3><div className="pp-list">{favs.map(p => <PersonRow key={p.id} p={p} onFav={fav} />)}</div></section>
          )}
          {GROUPS.map(([rel, label]) => {
            const ps = others.filter(p => p.relation === rel && (q || !p.favorite))
            if (!ps.length) return null
            return <section key={rel} className="pp-group"><h3 className="pp-h">{label} <span className="muted">{ps.length}</span></h3><div className="pp-list">{ps.map(p => <PersonRow key={p.id} p={p} onFav={fav} />)}</div></section>
          })}
          {shown.length === 0 && <p className="muted">No one matches “{q}”.</p>}
        </>
      )}
    </div>
  )
}
