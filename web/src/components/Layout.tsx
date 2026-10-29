// The app shell: sidebar, top bar, search palette, shortcuts — the Pocket Calendar layout.
import { createContext, useCallback, useContext, useEffect, useRef, useState } from 'react'
import { Outlet, useLocation, useNavigate, useSearchParams } from 'react-router-dom'
import { api, V1 } from '../lib/api'
import { useApp } from '../lib/app-context'
import type { Dashboard } from '../lib/types'
import { Icon } from './icons'
import { Sidebar } from './Sidebar'
import { CommandPalette, type PaletteCommand } from './CommandPalette'
import { ShortcutsSheet } from './ShortcutsSheet'
import { PersonModal } from './PersonForm'
import { CheckInModal } from './CheckIn'
import { ImportModal } from './ImportPeople'
import { AskModal } from './AskModal'
import { Avatar, type Anchor, type MenuItem, Menu, MenuButton, modalOpen, rectOf, useIsMobile, useMediaQuery } from './ui'
import { AskButton } from './AskButton'

const SIDEBAR_KEY = 'phr-sidebar'

/** Something was added or changed from the shell (or a page): the counts and the pages reload. */
export const announceChange = () => window.dispatchEvent(new CustomEvent('phr:changed'))
export function useOnChange(cb: () => void) {
  const ref = useRef(cb)
  ref.current = cb
  useEffect(() => {
    const h = () => ref.current()
    window.addEventListener('phr:changed', h)
    return () => window.removeEventListener('phr:changed', h)
  }, [])
}

/** The page's own name for the top bar breadcrumb. */
const CrumbCtx = createContext<(t: string) => void>(() => {})
export function useCrumb(title: string | undefined) {
  const set = useContext(CrumbCtx)
  useEffect(() => { set(title || ''); return () => set('') }, [title, set])
}

/** The dashboard numbers (today, streak, counts) for every page that wants them. */
const CountsCtx = createContext<Dashboard | null>(null)
export const useCounts = () => useContext(CountsCtx)

export type CreateKind = 'person' | 'checkin' | 'import' | 'contacts' | 'ask'

/** Open a create action from anywhere (a page's button). */
export const openCreate = (k: CreateKind) => window.dispatchEvent(new CustomEvent('phr:create', { detail: { kind: k } }))

/** The add menu: the sidebar's "Add" button and the phone's + button. */
export function createItems(onCreate: (k: CreateKind) => void): MenuItem[] {
  return [
    { label: 'Add a person…', icon: 'user', hint: 'N', onSelect: () => onCreate('person') },
    { label: "Today's check-in…", icon: 'smile', hint: 'C', onSelect: () => onCreate('checkin') },
    { label: 'Ask about your chart…', icon: 'sparkles', hint: 'A', onSelect: () => onCreate('ask') },
    'sep',
    { label: 'From Pocket Contacts…', icon: 'cake', onSelect: () => onCreate('contacts') },
    { label: 'Import birth data…', icon: 'upload', onSelect: () => onCreate('import') },
  ]
}

export default function Layout() {
  const navigate = useNavigate()
  const loc = useLocation()

  const [sp, setSp] = useSearchParams()
  const mobile = useIsMobile()
  const narrow = useMediaQuery('(max-width: 1099px)')
  const drawer = mobile || narrow
  const [sidebarOpen, setSidebarOpen] = useState(() => !drawer && localStorage.getItem(SIDEBAR_KEY) !== '0')
  useEffect(() => { setSidebarOpen(!drawer && localStorage.getItem(SIDEBAR_KEY) !== '0') }, [drawer])
  const drawerOpenedAt = useRef(0)
  const toggleSidebar = () => setSidebarOpen(o => { drawerOpenedAt.current = performance.now(); const n = !o; if (!drawer) localStorage.setItem(SIDEBAR_KEY, n ? '1' : '0'); return n })
  const closeDrawer = () => { if (drawer) setSidebarOpen(false) }
  useEffect(() => { document.querySelector('.content')?.scrollTo({ top: 0 }) }, [loc.pathname])

  const [crumb, setCrumbRaw] = useState('')
  const setCrumb = useCallback((t: string) => setCrumbRaw(t), [])

  const [counts, setCounts] = useState<Dashboard | null>(null)
  const load = useCallback(() => { api.get<Dashboard>(`${V1}/dashboard`).then(setCounts).catch(() => {}) }, [])
  useEffect(() => { load() }, [loc.pathname, load])
  useOnChange(load)

  const [palette, setPalette] = useState(false)
  const [shortcuts, setShortcuts] = useState(false)
  const [modal, setModal] = useState<{ kind: 'person' } | { kind: 'checkin' } | { kind: 'import'; tab: 'file' | 'contacts' } | { kind: 'ask' } | null>(null)
  const [fabMenu, setFabMenu] = useState<Anchor | null>(null)

  // /?checkin=1 (the PWA shortcut) opens the check-in, /?add=1 the person form
  useEffect(() => {
    if (sp.get('checkin')) { setModal({ kind: 'checkin' }); sp.delete('checkin'); setSp(sp, { replace: true }) }
    else if (sp.get('add')) { setModal({ kind: 'person' }); sp.delete('add'); setSp(sp, { replace: true }) }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const create = (k: CreateKind) => {
    closeDrawer()
    switch (k) {
      case 'person': setModal({ kind: 'person' }); break
      case 'checkin': setModal({ kind: 'checkin' }); break
      case 'import': setModal({ kind: 'import', tab: 'file' }); break
      case 'contacts': setModal({ kind: 'import', tab: 'contacts' }); break
      case 'ask': setModal({ kind: 'ask' }); break
    }
  }
  const createRef = useRef(create)
  createRef.current = create

  useEffect(() => {
    const open = () => setShortcuts(true)
    const make = (e: Event) => { const d = (e as CustomEvent<{ kind: CreateKind }>).detail; createRef.current(d.kind) }
    window.addEventListener('phr:shortcuts', open)
    window.addEventListener('phr:create', make)
    return () => { window.removeEventListener('phr:shortcuts', open); window.removeEventListener('phr:create', make) }
  }, [])

  useEffect(() => {
    const h = (e: KeyboardEvent) => {
      const t = e.target as HTMLElement
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') { e.preventDefault(); setPalette(p => !p); return }
      if (e.metaKey || e.ctrlKey || e.altKey) return
      if (t.closest('input, textarea, select, [contenteditable="true"]') || modalOpen() || document.querySelector('.menu, .popover, .sel-pop')) return
      const go = (to: string) => { e.preventDefault(); navigate(to) }
      const mk = (k: CreateKind) => { e.preventDefault(); createRef.current(k) }
      switch (e.key) {
        case '/': e.preventDefault(); setPalette(true); break
        case '?': e.preventDefault(); setShortcuts(true); break
        case 'n': mk('person'); break
        case 'c': mk('checkin'); break
        case 'a': mk('ask'); break
        case 't': go('/'); break
        case 'p': go('/people'); break
        case 'm': go('/match'); break
        case 'u': go('/transits'); break
        case 's': go('/sky'); break
        case 'j': go('/journal'); break
        case 'r': go('/readings'); break
        case ',': go('/settings'); break
        case '[': e.preventDefault(); toggleSidebar(); break
      }
    }
    window.addEventListener('keydown', h)
    return () => window.removeEventListener('keydown', h)
  }) // eslint-disable-line react-hooks/exhaustive-deps

  const commands: PaletteCommand[] = [
    { id: 'person', label: 'Add a person', icon: 'user', hint: 'N', keywords: 'new birth chart profile partner friend', run: () => create('person') },
    { id: 'checkin', label: "Today's check-in", icon: 'smile', hint: 'C', keywords: 'mood energy journal log', run: () => create('checkin') },
    { id: 'ask', label: 'Ask about your chart', icon: 'sparkles', hint: 'A', keywords: 'ai question reading', run: () => create('ask') },
    { id: 'today', label: 'Today', icon: 'sun', hint: 'T', keywords: 'horoscope daily day', run: () => navigate('/') },
    { id: 'people', label: 'People & charts', icon: 'users', hint: 'P', keywords: 'birth chart natal', run: () => navigate('/people') },
    { id: 'match', label: 'Compatibility', icon: 'heart', hint: 'M', keywords: 'synastry love match partner', run: () => navigate('/match') },
    { id: 'transits', label: 'Coming up (transits)', icon: 'trending-up', hint: 'U', keywords: 'future year ahead saturn return', run: () => navigate('/transits') },
    { id: 'sky', label: 'Sky now', icon: 'moon', hint: 'S', keywords: 'planets current retrograde', run: () => navigate('/sky') },
    { id: 'moon', label: 'Moon calendar', icon: 'calendar', keywords: 'phases full new eclipse void', run: () => navigate('/sky/moon') },
    { id: 'retro', label: 'Retrogrades this year', icon: 'refresh', keywords: 'mercury retrograde', run: () => navigate('/sky/retrogrades') },
    { id: 'journal', label: 'Journal', icon: 'journal', hint: 'J', keywords: 'mood diary check-in', run: () => navigate('/journal') },
    { id: 'insights', label: 'Insights: you and the Moon', icon: 'chart', keywords: 'mood statistics patterns sleep', run: () => navigate('/journal/insights') },
    { id: 'readings', label: 'Readings', icon: 'book-open', hint: 'R', keywords: 'ai saved kept', run: () => navigate('/readings') },
    { id: 'import', label: 'Import birth data', icon: 'upload', keywords: 'aaf csv astro.com solar fire', run: () => create('import') },
    { id: 'contacts', label: 'Add people from Pocket Contacts', icon: 'cake', keywords: 'birthdays contacts', run: () => create('contacts') },
    { id: 'chart-settings', label: 'Settings: charts', icon: 'sliders', keywords: 'zodiac sidereal houses placidus orbs home', run: () => navigate('/settings/charts') },
    { id: 'apps', label: 'Settings: connected apps', icon: 'layers', keywords: 'contacts notes health core integrations', run: () => navigate('/settings/apps') },
    { id: 'settings', label: 'Settings', icon: 'settings', hint: ',', run: () => navigate('/settings') },
    { id: 'shortcuts', label: 'Keyboard shortcuts', icon: 'keyboard', hint: '?', run: () => setShortcuts(true) },
  ]

  const closeModal = () => setModal(null)
  return (
    <CrumbCtx.Provider value={setCrumb}>
      <CountsCtx.Provider value={counts}>
        <div className={`app ${sidebarOpen ? 'with-sidebar' : 'no-sidebar'} ${mobile ? 'is-mobile' : ''} ${drawer ? 'drawer-mode' : ''}`}>
          {drawer && sidebarOpen && <div className="drawer-backdrop" onClick={e => { if (e.timeStamp - drawerOpenedAt.current > 400) setSidebarOpen(false) }} />}
          <div className={`sidebar-wrap ${sidebarOpen ? 'open' : ''}`}>
            <Sidebar onCreate={create} onNavigate={closeDrawer} counts={counts} />
          </div>
          <div className="main">
            <TopBar onMenu={toggleSidebar} onPalette={() => setPalette(true)} mobile={mobile} compact={narrow} sidebarOpen={sidebarOpen} crumb={crumb} />
            <main className="content"><div className="content-inner"><Outlet /></div></main>
          </div>
          {mobile && <button type="button" className="fab" aria-label="Add" title="Add" aria-haspopup="menu" aria-expanded={!!fabMenu} onClick={e => setFabMenu(fabMenu ? null : rectOf(e.currentTarget))}><Icon name="plus" size={24} strokeWidth={2.2} /></button>}
          {fabMenu && <Menu anchor={fabMenu} align="right" items={createItems(create)} onClose={() => setFabMenu(null)} width={250} />}
          {palette && <CommandPalette onClose={() => setPalette(false)} commands={commands} />}
          {shortcuts && <ShortcutsSheet onClose={() => setShortcuts(false)} />}
          {modal?.kind === 'person' && <PersonModal onClose={closeModal} onSaved={p => { closeModal(); announceChange(); navigate(`/people/${p.id}`) }} />}
          {modal?.kind === 'checkin' && <CheckInModal onClose={closeModal} onSaved={() => { closeModal(); announceChange() }} />}
          {modal?.kind === 'import' && <ImportModal initial={modal.tab} onClose={closeModal} onDone={() => { closeModal(); announceChange(); navigate('/people') }} />}
          {modal?.kind === 'ask' && <AskModal onClose={closeModal} />}
        </div>
      </CountsCtx.Provider>
    </CrumbCtx.Provider>
  )
}

const SECTION: Record<string, [string, string]> = { '': ['Today', '/'], people: ['People', '/people'], match: ['Compatibility', '/match'], transits: ['Coming up', '/transits'], sky: ['Sky & Moon', '/sky'], journal: ['Journal', '/journal'], readings: ['Readings', '/readings'] }

function TopBar({ onMenu, onPalette, mobile, compact, sidebarOpen, crumb }: { onMenu: () => void; onPalette: () => void; mobile: boolean; compact: boolean; sidebarOpen: boolean; crumb: string }) {
  const { user, themeMode, setThemeMode, status, setUser } = useApp()
  const navigate = useNavigate()
  const loc = useLocation()
  const [section, sectionTo] = SECTION[loc.pathname.split('/')[1] || ''] || ['', '/']
  const signOut = async () => { await api.post(`${V1}/auth/logout`).catch(() => {}); setUser(null); navigate('/login') }
  const avatarMenu = (): MenuItem[] => [
    { heading: user?.name || user?.email || '' },
    { label: 'Profile', icon: 'user', onSelect: () => navigate('/settings/profile') },
    { label: 'Settings', icon: 'settings', onSelect: () => navigate('/settings'), hint: ',' },
    { label: 'Keyboard shortcuts', icon: 'keyboard', onSelect: () => window.dispatchEvent(new CustomEvent('phr:shortcuts')), hint: '?' },
    'sep',
    { heading: 'Appearance' },
    { label: 'System', checked: themeMode === 'system', onSelect: () => setThemeMode('system') },
    { label: 'Light', checked: themeMode === 'light', onSelect: () => setThemeMode('light') },
    { label: 'Dark', checked: themeMode === 'dark', onSelect: () => setThemeMode('dark') },
    ...(status?.auth_mode === 'local' ? ['sep' as const, { label: 'Sign out', icon: 'log-out', onSelect: signOut }] : []),
  ]
  return (
    <header className="topbar">
      <button type="button" className="btn ghost icon-only" onClick={onMenu} aria-label={sidebarOpen ? 'Hide the sidebar' : 'Show the sidebar'} title="Menu ([)"><Icon name={mobile || compact ? 'menu' : 'sidebar'} size={18} /></button>
      <h1 className="range-title">
        {crumb ? <><button type="button" className="crumb" onClick={() => navigate(sectionTo)}>{section}</button><Icon name="chevron-right" size={16} className="crumb-sep" /><span className="ellipsis">{crumb}</span></> : section}
      </h1>
      <span className="spacer" />
      <button type="button" className="btn ghost icon-only" onClick={onPalette} aria-label="Search" title="Search (⌘K or /)"><Icon name="search" size={18} /></button>
      <AskButton />
      <MenuButton label="Account" className="avatar-btn" align="right" width={220} items={avatarMenu}>
        <Avatar name={user?.name} email={user?.email} size={30} />
      </MenuButton>
    </header>
  )
}
