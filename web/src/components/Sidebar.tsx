// Left column: the add button, sections, today's Moon and your Big Three — the Pocket Calendar sidebar.
import { useState } from 'react'
import { Link, NavLink } from 'react-router-dom'
import { useApp } from '../lib/app-context'
import { PHASE_EMOJI, PHASE_NAME, SIGN_GLYPH, signName } from '../lib/astro'
import type { Dashboard } from '../lib/types'
import { Icon } from './icons'
import { createItems, type CreateKind } from './Layout'
import { type Anchor, Menu, rectOf } from './ui'

export function Sidebar({ onCreate, onNavigate, counts }: { onCreate: (k: CreateKind) => void; onNavigate: () => void; counts: Dashboard | null }) {
  const { status, settings } = useApp()
  const [menu, setMenu] = useState<Anchor | null>(null)
  const title = settings?.title || status?.title || 'Pocket Horoscope'
  const nav: { to: string; label: string; icon: string; end?: boolean; count?: string }[] = [
    { to: '/', label: 'Today', icon: 'sun', end: true, count: counts?.overall != null ? String(counts.overall) : undefined },
    { to: '/people', label: 'People & charts', icon: 'users', count: counts?.people ? String(counts.people) : undefined },
    { to: '/match', label: 'Compatibility', icon: 'heart' },
    { to: '/transits', label: 'Coming up', icon: 'trending-up' },
    { to: '/sky', label: 'Sky & Moon', icon: 'moon' },
    { to: '/journal', label: 'Journal', icon: 'journal' },
    { to: '/readings', label: 'Readings', icon: 'book-open' },
  ]
  return (
    <aside className="sidebar" aria-label="Pocket Horoscope">
      <div className="sb-brand">
        <img src="/favicon.svg" alt="" className="sb-logo" />
        <span className="sb-title">{title !== 'Pocket Horoscope' ? title : <>Pocket <b>Horoscope</b></>}</span>
        {status?.version && <span className="sb-version" title={`Version ${status.version}`}>{/^v?\d+\.\d+/.test(status.version) ? 'v' + status.version.replace(/^v/, '').replace(/[-+].*$/, '') : status.version}</span>}
      </div>
      <button type="button" className="btn primary create-btn" aria-haspopup="menu" aria-expanded={!!menu} onClick={e => setMenu(menu ? null : rectOf(e.currentTarget))}>
        <Icon name="plus" size={15} strokeWidth={2.2} />Add<Icon name="chevron-down" size={15} className="create-caret" />
      </button>
      {menu && <Menu anchor={menu} items={createItems(onCreate)} onClose={() => setMenu(null)} width={256} />}

      <nav className="sb-nav" aria-label="Sections">
        {nav.map(n => (
          <NavLink key={n.to} to={n.to} end={n.end} onClick={onNavigate} className={({ isActive }) => `sb-nav-item ${isActive ? 'active' : ''}`}>
            <Icon name={n.icon} size={16} />{n.label}
            {!!n.count && <span className="sb-count">{n.count}</span>}
          </NavLink>
        ))}
      </nav>

      {counts && (
        <section className="sb-sec sb-moon" aria-label="The Moon today">
          <Link to="/sky/moon" className="sb-moon-row" onClick={onNavigate}>
            <span className="sb-moon-emoji" aria-hidden>{PHASE_EMOJI[counts.moon.phase]}</span>
            <span className="sb-moon-text">
              <b>{PHASE_NAME[counts.moon.phase]}</b>
              <span className="muted small">in {signName(counts.moon.sign)} · {Math.round(counts.moon.illumination * 100)}% lit</span>
            </span>
          </Link>
        </section>
      )}

      {counts?.me && (
        <section className="sb-sec">
          <h4 className="sb-h"><span className="sb-h-btn">{counts.me.name.split(' ')[0]}'s Big Three</span></h4>
          <div className="sb-big3">
            {([['Sun', counts.me.sun], ['Moon', counts.me.moon], ['Rising', counts.me.rising]] as const).map(([k, v]) => (
              <div key={k} className="sb-big3-item" title={v ? `${k} in ${signName(v)}` : `${k}: needs a birth time`}>
                <span className="sb-big3-glyph">{v ? SIGN_GLYPH[v] : '–'}</span>
                <span className="muted small">{k}</span>
              </div>
            ))}
          </div>
        </section>
      )}

      <div className="sb-foot">
        <Link to="/settings" className="sb-foot-link" onClick={onNavigate}><Icon name="settings" size={15} />Settings</Link>
        <button type="button" className="sb-foot-link" onClick={() => { onNavigate(); window.dispatchEvent(new CustomEvent('phr:shortcuts')) }}><Icon name="keyboard" size={15} />Shortcuts</button>
      </div>
    </aside>
  )
}
