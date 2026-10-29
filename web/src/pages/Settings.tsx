import { AskButton } from '../components/AskButton'
import React, { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api, V1 } from '../lib/api'
import { useApp } from '../lib/app-context'
import type { APIToken, ConnMode, HouseSystem, Settings as SettingsT, User } from '../lib/types'
import { DATE_FORMATS, TIME_FORMATS, fmtBytes, fmtDateTime, type DateFormat, type TimeFormat } from '../lib/format'
import { Avatar, Confirm, CoreHint, Field, Loading, Modal, Segmented, SettingRow, Switch, Tabs, useAsync, useIsMobile, useToast } from '../components/ui'
import { Icon } from '../components/icons'
import { Select } from '../components/Select'
import { ShortcutsSheet } from '../components/ShortcutsSheet'
import { PocketModelSelect } from '../components/PocketModelSelect'
import { PlaceSearch, placeLabel } from '../components/PlaceSearch'

type TabId = 'general' | 'charts' | 'notifications' | 'apps' | 'ai' | 'profile' | 'people' | 'tokens' | 'data' | 'about'

const TABS: { id: TabId; label: string; icon: string; sub: string }[] = [
  { id: 'general', label: 'General', icon: 'settings', sub: 'How dates look and the name of the app.' },
  { id: 'charts', label: 'Charts', icon: 'sliders', sub: 'How charts are worked out, and where you live.' },
  { id: 'notifications', label: 'Notifications', icon: 'bell', sub: 'Your horoscope in the morning, the Moon’s big days, retrogrades and birthdays.' },
  { id: 'apps', label: 'Connected apps', icon: 'layers', sub: 'Pocket Contacts (people by their birthdays), Pocket Notes (keep a reading) and Pocket Health (your sleep next to the Moon).' },
  { id: 'ai', label: 'AI', icon: 'sparkles', sub: 'Who writes your readings and answers questions: Pocket AI through Core, or your own OpenAI key.' },
  { id: 'profile', label: 'Profile', icon: 'user', sub: 'Your name, email and password.' },
  { id: 'people', label: 'Accounts', icon: 'users', sub: 'Everyone who has an account on this box.' },
  { id: 'tokens', label: 'API & shortcuts', icon: 'key', sub: 'Personal tokens for Homepage widgets, Siri Shortcuts and scripts.' },
  { id: 'data', label: 'Backup & data', icon: 'database', sub: 'Download the database, export or import your people, check-ins and readings.' },
  { id: 'about', label: 'About', icon: 'info', sub: 'Version, numbers and the server.' },
]

export default function Settings() {
  const { tab: raw } = useParams()
  const navigate = useNavigate()
  const mobile = useIsMobile()
  const tab: TabId = TABS.find(t => t.id === raw)?.id || 'general'
  const current = TABS.find(t => t.id === tab)!
  const go = (id: TabId) => navigate(`/settings/${id}`, { replace: true })
  useEffect(() => { document.querySelector('.set-content')?.scrollTo({ top: 0 }) }, [tab])
  useEffect(() => {
    const h = (e: KeyboardEvent) => { if (e.key === 'Escape' && !document.querySelector('.modal-backdrop, .menu, .sel-pop, .place-pop')) navigate('/') }
    window.addEventListener('keydown', h)
    return () => window.removeEventListener('keydown', h)
  }, [navigate])
  return (
    <div className="set-page">
      <header className="set-header">
        <button type="button" className="btn ghost icon-only" aria-label="Back" title="Back (esc)" onClick={() => navigate('/')}><Icon name="chevron-left" size={18} /></button>
        <img src="/favicon.svg" alt="" width={24} height={24} className="set-logo" />
        <h1>Settings</h1>
        <span style={{ flex: 1 }} />
        <AskButton />
      </header>
      {mobile && <div className="set-tabs-mobile"><Tabs tabs={TABS.map(t => ({ id: t.id, label: t.label, icon: t.icon }))} value={tab} onChange={go} /></div>}
      <div className="set-body">
        {!mobile && (
          <nav className="set-nav" aria-label="Settings sections">
            {TABS.map(t => (
              <button key={t.id} type="button" className={`set-nav-item ${t.id === tab ? 'active' : ''}`} aria-current={t.id === tab ? 'page' : undefined} onClick={() => go(t.id)}>
                <Icon name={t.icon} size={16} />{t.label}
              </button>
            ))}
          </nav>
        )}
        <main className="set-content">
          <div className="set-inner" key={tab}>
            <div className="set-title"><h2>{current.label}</h2><p>{current.sub}</p></div>
            {tab === 'general' && <General />}
            {tab === 'charts' && <Charts />}
            {tab === 'notifications' && <Notifications />}
            {tab === 'apps' && <Integrations />}
            {tab === 'ai' && <AITab />}
            {tab === 'profile' && <Profile />}
            {tab === 'people' && <Users />}
            {tab === 'tokens' && <Tokens />}
            {tab === 'data' && <Data />}
            {tab === 'about' && <About />}
          </div>
        </main>
      </div>
    </div>
  )
}

function Section({ title, sub, icon, actions, children }: { title: React.ReactNode; sub?: React.ReactNode; icon?: string; actions?: React.ReactNode; children?: React.ReactNode }) {
  return (
    <section className="set-section">
      <div className="set-section-head">
        {icon && <span className="set-app-ico"><Icon name={icon} size={17} /></span>}
        <div className="grow"><h3>{title}</h3>{sub && <p>{sub}</p>}</div>
        {actions && <div className="row">{actions}</div>}
      </div>
      {children && <div className="set-section-body">{children}</div>}
    </section>
  )
}

function useSettingsForm() {
  const { settings, setSettings, user } = useApp()
  const { toast } = useToast()
  const [f, setF] = useState<SettingsT | null>(settings)
  const [busy, setBusy] = useState(false)
  useEffect(() => { if (!f && settings) setF(settings) }, [settings]) // eslint-disable-line react-hooks/exhaustive-deps
  const save = async (patch?: Partial<SettingsT>) => {
    if (!f) return
    setBusy(true)
    try { const s = await api.put<SettingsT>(`${V1}/settings`, { ...f, ...patch }); setSettings(s); setF(s); toast('Saved') } catch (e) { toast((e as Error).message, 'err') } finally { setBusy(false) }
  }
  return { f, setF, save, busy, isAdmin: user?.role === 'admin' }
}

function useTester() {
  const [test, setTest] = useState<Record<string, string>>({})
  const run = async (which: string, url: string, key: string, extra: Record<string, string> = {}) => {
    setTest(t => ({ ...t, [which]: 'testing…' }))
    try { const r = await api.post<{ ok: boolean; message?: string; error?: string }>(`${V1}/integrations/test`, { which, url, key, ...extra }); setTest(t => ({ ...t, [which]: r.ok ? `✓ ${r.message}` : `✕ ${r.error}` })) } catch (e) { setTest(t => ({ ...t, [which]: `✕ ${(e as Error).message}` })) }
  }
  const status = (which: string) => <span className={`small ${test[which]?.startsWith('✓') ? 'ok' : test[which]?.startsWith('✕') ? 'bad' : 'muted'}`}>{test[which]}</span>
  return { run, status }
}

// every half hour of the day, for the morning horoscope
const HALF_HOURS = Array.from({ length: 48 }, (_, i) => { const v = `${String(Math.floor(i / 2)).padStart(2, '0')}:${i % 2 ? '30' : '00'}`; return { value: v, label: v } })

function General() {
  const { f, setF, save, busy, isAdmin } = useSettingsForm()
  if (!f) return <Loading />
  return (
    <div className="stack loose">
      <Section title="Dates" icon="calendar">
        <SettingRow title="Week starts on" help="For the Moon calendar."><Segmented size="sm" value={f.week_start} onChange={v => setF({ ...f, week_start: v })} options={[{ value: 1, label: 'Monday' }, { value: 6, label: 'Saturday' }, { value: 7, label: 'Sunday' }]} /></SettingRow>
        <SettingRow title="Date format"><Select ariaLabel="Date format" value={f.date_format} onChange={v => setF({ ...f, date_format: v as DateFormat })} options={DATE_FORMATS.map(d => ({ value: d.id, label: d.label }))} /></SettingRow>
        <SettingRow title="Time format"><Select ariaLabel="Time format" value={f.time_format} onChange={v => setF({ ...f, time_format: v as TimeFormat })} options={TIME_FORMATS.map(d => ({ value: d.id, label: d.label }))} /></SettingRow>
        {isAdmin && <SettingRow title="App title" help="Shown in the sidebar for everyone on this box."><input value={f.title} onChange={e => setF({ ...f, title: e.target.value })} /></SettingRow>}
      </Section>
      <div className="row"><button type="button" className="btn primary" onClick={() => save()} disabled={busy}>Save changes</button></div>
    </div>
  )
}

const HOUSES: { value: HouseSystem; label: string; hint: string }[] = [
  { value: 'placidus', label: 'Placidus', hint: 'The most used today: houses by the time it takes each degree to rise.' },
  { value: 'koch', label: 'Koch', hint: 'Like Placidus, from the time of birth at the birth place; popular in Germany.' },
  { value: 'whole', label: 'Whole sign', hint: 'The oldest way: each house is one whole sign, starting with your rising sign.' },
  { value: 'equal', label: 'Equal', hint: 'Twelve equal houses of 30°, starting at your rising degree.' },
  { value: 'porphyry', label: 'Porphyry', hint: 'Each quarter of the sky split in three; works everywhere, even near the poles.' },
]

function Charts() {
  const { f, setF, save, busy } = useSettingsForm()
  if (!f) return <Loading />
  const hs = HOUSES.find(h => h.value === f.house_system) || HOUSES[0]
  const home = f.home_tz ? placeLabel({ place: f.home_place, country: f.home_country }) : ''
  return (
    <div className="stack loose">
      <Section title="The zodiac" icon="sun" sub="Both are real astrology traditions; they differ by about 24° today, so most planets fall one sign earlier in the sidereal zodiac.">
        <SettingRow title="Zodiac" help={f.zodiac === 'sidereal' ? 'Sidereal (Lahiri): the signs follow the stars, as in Vedic astrology.' : 'Tropical: the signs follow the seasons — the usual Western horoscope.'}>
          <Segmented size="sm" value={f.zodiac} onChange={v => setF({ ...f, zodiac: v })} options={[{ value: 'tropical', label: 'Tropical' }, { value: 'sidereal', label: 'Sidereal' }]} />
        </SettingRow>
      </Section>
      <Section title="Houses" icon="grid" sub="The twelve areas of life in a chart. They need a birth time; without one a chart shows no houses.">
        <SettingRow title="House system" help={hs.hint}>
          <Select ariaLabel="House system" value={f.house_system} onChange={v => setF({ ...f, house_system: v as HouseSystem })} options={HOUSES.map(h => ({ value: h.value, label: h.label }))} />
        </SettingRow>
      </Section>
      <Section title="Aspects" icon="target" sub="An aspect is an angle between two planets. The orb is how far from exact it may be and still count.">
        <SettingRow title="Orbs" help={f.orbs === 'tight' ? 'Tight: only the strongest aspects, a calmer chart.' : f.orbs === 'wide' ? 'Wide: more aspects, including the faint ones.' : 'Normal: the usual allowance.'}>
          <Segmented size="sm" value={f.orbs} onChange={v => setF({ ...f, orbs: v })} options={[{ value: 'tight', label: 'Tight' }, { value: 'normal', label: 'Normal' }, { value: 'wide', label: 'Wide' }]} />
        </SettingRow>
        <SettingRow title="Minor aspects" help="Also show the quincunx (150°), an awkward angle that asks for adjustment."><Switch checked={f.minor_aspects} onChange={v => setF({ ...f, minor_aspects: v })} label="Minor aspects" /></SettingRow>
      </Section>
      <Section title="Where you live" icon="pin" sub="Times in the Moon calendar and today’s sky are shown for this place. Empty = your own birth place.">
        <div className="stack">
          <PlaceSearch value={home} onPick={p => setF({ ...f, home_place: p.name, home_country: p.country, home_lat: p.lat, home_lon: p.lon, home_tz: p.tz })} placeholder="Search your town or city" disabled={!f.geocoder && !f.home_tz} />
          {!f.geocoder && <p className="muted small">Place search is switched off on this server (GEOCODER_URL=off).</p>}
          {f.home_tz
            ? <div className="row wrap"><span className="muted small">{f.home_lat.toFixed(2)}°, {f.home_lon.toFixed(2)}° · {f.home_tz}</span><button type="button" className="btn ghost sm" onClick={() => setF({ ...f, home_place: '', home_country: '', home_lat: 0, home_lon: 0, home_tz: '' })}>Clear</button></div>
            : <p className="muted small">Not set: your own birth place is used.</p>}
        </div>
      </Section>
      <div className="row"><button type="button" className="btn primary" onClick={() => save()} disabled={busy}>Save changes</button></div>
    </div>
  )
}

function Notifications() {
  const { f, setF, save, busy } = useSettingsForm()
  const { toast } = useToast()
  const { run, status } = useTester()
  const [sending, setSending] = useState(false)
  if (!f) return <Loading />
  const sendNow = async () => {
    setSending(true)
    try { const r = await api.post<{ ok: boolean; sent?: number; error?: string }>(`${V1}/notify/today`); if (r.ok) toast(`Sent ${r.sent} ${r.sent === 1 ? 'message' : 'messages'}`); else toast(r.error || 'failed', 'err') } catch (e) { toast((e as Error).message, 'err') } finally { setSending(false) }
  }
  const time = HALF_HOURS.some(h => h.value === f.daily_time) ? HALF_HOURS : [...HALF_HOURS, { value: f.daily_time, label: f.daily_time }].sort((a, b) => a.value.localeCompare(b.value))
  return (
    <div className="stack loose">
      <Section title="Where messages go" icon="bell" actions={<span className={`badge ${f.notify_configured ? 'ok' : ''}`}>{f.notify_configured ? 'on' : 'not set'}</span>}
        sub={<>An Apprise gateway (<code>…/notify/&lt;key&gt;</code>), an ntfy topic URL, or any webhook that takes JSON.</>}>
        <SettingRow title="URL" stack><input value={f.notify_url} onChange={e => setF({ ...f, notify_url: e.target.value })} placeholder="http://apprise:8000/notify/horoscope" /></SettingRow>
        <SettingRow title="Format"><Segmented value={f.notify_format} onChange={v => setF({ ...f, notify_format: v })} options={[{ value: 'apprise', label: 'Apprise' }, { value: 'ntfy', label: 'ntfy' }, { value: 'json', label: 'JSON' }]} size="sm" /></SettingRow>
      </Section>
      <Section title="Every morning" icon="sunrise" sub="Everything below arrives once a day at this time, where you live.">
        <SettingRow title="At"><Select ariaLabel="Time" value={f.daily_time} onChange={v => setF({ ...f, daily_time: v })} options={time} /></SettingRow>
        <SettingRow title="Your horoscope" help="The day’s headline, your scores and any big transit that is exact today."><Switch checked={f.daily_notify} onChange={v => setF({ ...f, daily_notify: v })} label="Daily horoscope" /></SettingRow>
        <SettingRow title="New and full Moons, eclipses" help="On the day, with the time."><Switch checked={f.notify_moon} onChange={v => setF({ ...f, notify_moon: v })} label="Moon days" /></SettingRow>
        <SettingRow title="Retrogrades" help="When a planet turns retrograde or direct."><Switch checked={f.notify_retro} onChange={v => setF({ ...f, notify_retro: v })} label="Retrogrades" /></SettingRow>
        <SettingRow title="Big transits" help="Mentioned in your horoscope on the day a slow planet’s transit to your chart is exact."><Switch checked={f.notify_transits} onChange={v => setF({ ...f, notify_transits: v })} label="Big transits" /></SettingRow>
        <SettingRow title="Birthdays" help="The people you keep here, on their birthday."><Switch checked={f.notify_birthdays} onChange={v => setF({ ...f, notify_birthdays: v })} label="Birthdays" /></SettingRow>
      </Section>
      <div className="row wrap">
        <button type="button" className="btn primary" onClick={() => save()} disabled={busy}>Save changes</button>
        <button type="button" className="btn" onClick={() => run('notify', f.notify_url, '', { format: f.notify_format })} disabled={!f.notify_url}><Icon name="mail" size={15} />Send a test</button>
        <button type="button" className="btn" onClick={sendNow} disabled={!f.notify_configured || sending}><Icon name="sun" size={15} />Send today’s now</button>
        {status('notify')}
      </div>
    </div>
  )
}

function Integrations() {
  const { f, setF, save, busy } = useSettingsForm()
  const { run, status } = useTester()
  if (!f) return <Loading />
  const coreOn = f.core_configured && f.core_registered !== false
  const eff = (m: ConnMode): ConnMode => !coreOn && m === 'core' ? 'off' : m
  return (
    <div className="stack loose">
      <div className="callout info"><Icon name="info" size={15} /><div>These connections are yours. Through <b>Pocket Core</b> Horoscope acts as you in your own Pocket apps with nothing to paste (switch it off for yourself in Pocket Core → Settings → Apps acting for me); <b>Direct</b> uses a URL and an API token you enter.{!coreOn && <CoreHint />}</div></div>
      <Section title="Pocket Contacts — people by their birthdays" icon="cake" actions={connBadge(eff(f.contacts_mode), f.contacts_via)}
        sub="Add the people in your contacts who have a birthday, so their charts and your compatibility are a tap away. Add → From Pocket Contacts.">
        <Connection app="Pocket Contacts" mode={f.contacts_mode} via={f.contacts_via} detail={f.contacts_detail} core={coreOn} hasToken={!!f.contacts_token}
          onMode={m => save({ contacts_mode: m })} onDisconnect={() => save({ contacts_url: '', contacts_token: '' })}
          test={<><button type="button" className="btn sm" onClick={() => run('contacts', f.contacts_url, f.contacts_token)}>Test connection</button>{status('contacts')}</>}
          direct={<div className="two-col tight">
            <Field label="URL" help="Inside Docker use http://host.docker.internal:PORT or the service name."><input value={f.contacts_url} onChange={e => setF({ ...f, contacts_url: e.target.value })} placeholder="http://host.docker.internal:8191" /></Field>
            <Field label="API token" help="Pocket Contacts → Settings → API tokens"><input value={f.contacts_token} onChange={e => setF({ ...f, contacts_token: e.target.value })} autoComplete="off" /></Field>
          </div>} />
      </Section>
      <Section title="Pocket Notes — keep a reading" icon="note" actions={connBadge(eff(f.notes_mode), f.notes_via)}
        sub="“Save to Pocket Notes” on a reading writes it as a note in the folder you choose.">
        <Connection app="Pocket Notes" mode={f.notes_mode} via={f.notes_via} detail={f.notes_detail} core={coreOn} hasToken={!!f.notes_token}
          onMode={m => save({ notes_mode: m })} onDisconnect={() => save({ notes_url: '', notes_token: '' })}
          test={<><button type="button" className="btn sm" onClick={() => run('notes', f.notes_url, f.notes_token)}>Test connection</button>{status('notes')}</>}
          direct={<div className="two-col tight">
            <Field label="URL" help="Inside Docker use http://host.docker.internal:PORT or the service name."><input value={f.notes_url} onChange={e => setF({ ...f, notes_url: e.target.value })} placeholder="http://host.docker.internal:8182" /></Field>
            <Field label="API token" help="Pocket Notes → Settings → API tokens"><input value={f.notes_token} onChange={e => setF({ ...f, notes_token: e.target.value })} autoComplete="off" /></Field>
          </div>} />
        {eff(f.notes_mode) !== 'off' && <SettingRow title="Folder" help="Where readings go in Pocket Notes."><input value={f.notes_folder} onChange={e => setF({ ...f, notes_folder: e.target.value })} placeholder="Horoscope" /></SettingRow>}
      </Section>
      <Section title="Pocket Health — your sleep and the Moon" icon="bed" actions={connBadge(eff(f.health_mode), f.health_via)}
        sub="Insights compare your sleep and resting heart rate with the phases of the Moon. Only read, never changed.">
        <Connection app="Pocket Health" mode={f.health_mode} via={f.health_via} detail={f.health_detail} core={coreOn} hasToken={!!f.health_token}
          onMode={m => save({ health_mode: m })} onDisconnect={() => save({ health_url: '', health_token: '' })}
          test={<><button type="button" className="btn sm" onClick={() => run('health', f.health_url, f.health_token)}>Test connection</button>{status('health')}</>}
          direct={<div className="two-col tight">
            <Field label="URL" help="Inside Docker use http://host.docker.internal:PORT or the service name."><input value={f.health_url} onChange={e => setF({ ...f, health_url: e.target.value })} placeholder="http://host.docker.internal:8199" /></Field>
            <Field label="API token" help="Pocket Health → Settings → API tokens"><input value={f.health_token} onChange={e => setF({ ...f, health_token: e.target.value })} autoComplete="off" /></Field>
          </div>} />
        {eff(f.health_mode) !== 'off' && <SettingRow title="Compare my sleep with the Moon" help="Shown under Journal → Insights."><Switch checked={f.health_insights} onChange={v => save({ health_insights: v })} label="Compare my sleep with the Moon" /></SettingRow>}
      </Section>
      <Section title="Other Pocket apps" icon="layers" sub="Horoscope offers your day to the rest of the box through Pocket Core: Pocket Day writes it into your journal and Pocket Calendar shows the Moon and your big transits. The birth data of the people you keep never leaves this app.">
        <dl className="set-dl">
          <div><dt>horoscope:read</dt><dd className="muted">Your day, the Moon and the sky, the transits coming up and your check-ins — for Pocket Day, Calendar and dashboards. Never other people’s birth data.</dd></div>
          <div><dt>horoscope:write</dt><dd className="muted">Add a day’s mood check-in from another app; it can change only what it added itself.</dd></div>
        </dl>
      </Section>
      <div className="row"><button type="button" className="btn primary" onClick={() => save()} disabled={busy}>Save changes</button></div>
    </div>
  )
}

function AITab() {
  const { f, setF, save, busy } = useSettingsForm()
  const { run, status } = useTester()
  const { data: ai } = useAsync(() => api.get<{ ready: boolean; provider: string; detail: string }>(`${V1}/ai/status`), [])
  if (!f) return <Loading />
  const coreOn = f.core_configured && f.core_registered !== false
  const shown = !coreOn && f.ai_provider === 'pocket' ? 'openai' : (f.ai_provider || 'auto')
  return (
    <div className="stack loose">
      <Section title="What the AI does" icon="sparkles" sub="Writes readings in your own words — your day, a birth chart, the year ahead, a compatibility — and answers questions, from the worked-out chart only. Nothing is sent unless you ask, or switch on the morning reading below.">
        {ai && (ai.ready ? <p className="muted small row" style={{ gap: 6 }}><Icon name="sparkles" size={13} />{ai.provider === 'pocket' ? 'Pocket AI' : 'Your OpenAI key'} is ready.</p>
          : <div className="callout info"><Icon name="info" size={15} /><div>No AI connected{ai.detail ? ` (${ai.detail})` : ''}. Everything else works without it.{!coreOn && <CoreHint />}</div></div>)}
      </Section>
      <Section title="Provider" icon="sparkles">
        <div className="grid" style={{ gap: 12 }}>
          <Field group label="Who writes" help="Automatic prefers your enabled Pocket AI connection; without one it uses your OpenAI settings.">
            <Select ariaLabel="AI provider" value={shown} onChange={v => setF({ ...f, ai_provider: v as SettingsT['ai_provider'] })} options={[{ value: 'auto', label: coreOn ? 'Automatic — prefer Pocket AI' : 'Automatic' }, { value: 'pocket', label: 'Pocket AI', disabled: !coreOn, hint: coreOn ? undefined : 'needs Pocket Core' }, { value: 'openai', label: 'OpenAI / compatible API' }]} />
          </Field>
          {!coreOn && <CoreHint />}
          {coreOn && shown !== 'openai' && <PocketModelSelect value={f.ai_pocket_model || ''} onChange={v => setF({ ...f, ai_pocket_model: v })} />}
          {(shown === 'openai' || (!coreOn && shown === 'auto')) && <>
            <Field label="OpenAI API key" help="platform.openai.com → API keys. Stored on the server; shown masked."><input value={f.openai_api_key} onChange={e => setF({ ...f, openai_api_key: e.target.value })} autoComplete="off" placeholder="sk-…" /></Field>
            <div className="two-col tight">
              <Field label="Model" help="Default gpt-4o-mini."><input value={f.openai_model} onChange={e => setF({ ...f, openai_model: e.target.value })} placeholder="gpt-4o-mini" /></Field>
              <Field label="Base URL (optional)" help="Ollama, LM Studio, OpenRouter…"><input value={f.openai_base_url} onChange={e => setF({ ...f, openai_base_url: e.target.value })} placeholder="https://api.openai.com/v1" /></Field>
            </div>
          </>}
        </div>
      </Section>
      <Section title="How it writes" icon="pen">
        <SettingRow title="Tone" help={f.ai_tone === 'straight' ? 'Plain and direct, no softening.' : f.ai_tone === 'playful' ? 'Light, with a little humour.' : 'Warm and encouraging.'}>
          <Segmented size="sm" value={f.ai_tone} onChange={v => setF({ ...f, ai_tone: v })} options={[{ value: 'gentle', label: 'Gentle' }, { value: 'straight', label: 'Straight' }, { value: 'playful', label: 'Playful' }]} />
        </SettingRow>
        <SettingRow title="Write my daily reading every morning" help="At your notification time, so it is ready when you open the app (and in the morning message)."><Switch checked={f.ai_daily} onChange={v => setF({ ...f, ai_daily: v })} label="Daily AI reading" /></SettingRow>
      </Section>
      <div className="row wrap">
        <button type="button" className="btn primary" onClick={() => save()} disabled={busy}>Save changes</button>
        {(shown === 'openai' || (!coreOn && shown === 'auto')) && <><button type="button" className="btn" onClick={() => run('openai', f.openai_base_url, f.openai_api_key)} disabled={!f.openai_api_key}>Test the OpenAI key</button>{status('openai')}</>}
      </div>
    </div>
  )
}

function Connection({ app, mode: saved, via, detail, core, onMode, onDisconnect, hasToken, test, direct }: {
  app: string; mode: ConnMode; via: string; detail: string; core: boolean; onMode: (m: ConnMode) => void; onDisconnect: () => void
  hasToken: boolean; test: React.ReactNode; direct: React.ReactNode
}) {
  const [advanced, setAdvanced] = useState(saved === 'direct')
  const mode: ConnMode = !core && saved === 'core' ? 'off' : saved
  const ok = via !== '' && mode !== 'off'
  const showDirect = !core || mode === 'direct' || advanced
  const line = mode === 'off' ? `Off: Horoscope does not talk to ${app}.`
    : mode === 'core' ? (ok ? `Connected through Pocket Core, as you${detail ? ` (${detail})` : ''}. Nothing to fill in; switch it off for yourself in Pocket Core → Settings → Apps acting for me.` : `Not connected: ${detail}.`)
    : (ok ? `Direct connection with your own ${app} URL and token; use Test connection to check it.` : `Not connected: ${detail || 'add the URL and a token'}.`)
  return (
    <div className="conn-block">
      <div className="conn-mode">
        <Segmented size="sm" value={mode} onChange={m => onMode(m as ConnMode)} options={[
          { value: 'core', label: 'Pocket Core', title: core ? 'Through Pocket Core: per person, nothing to paste' : 'Needs Pocket Core: there is none on this box', disabled: !core },
          ...(showDirect ? [{ value: 'direct', label: 'Direct', title: 'Your own URL and API token' }] : []),
          { value: 'off', label: 'Off' },
        ]} />
        {!showDirect && <button type="button" className="conn-advanced" onClick={() => setAdvanced(true)}>Advanced: connect directly with a URL and key</button>}
      </div>
      <div className={`conn-status ${mode === 'off' ? '' : ok ? 'ok' : 'bad'}`}><span className="dot" />{line}</div>
      {mode === 'direct' && <>
        {direct}
        <div className="row wrap">{test}{hasToken && <button type="button" className="btn sm ghost danger-text" onClick={onDisconnect} title="Remove the saved URL and token">Disconnect</button>}</div>
      </>}
      {mode === 'core' && <div className="row wrap">{test}</div>}
    </div>
  )
}

const connBadge = (mode: ConnMode, via: string) => mode === 'off' ? <span className="badge">off</span>
  : via ? <span className="badge ok">{via === 'core' ? 'Pocket Core' : 'direct'}</span>
  : <span className="badge warn">not connected</span>

function Users() {
  const { user, users, reloadMeta, status } = useApp()
  const { toast } = useToast()
  const [edit, setEdit] = useState<Partial<User> | null>(null)
  const [del, setDel] = useState<User | null>(null)
  const isAdmin = user?.role === 'admin'
  const run = async (fn: () => Promise<unknown>, msg: string) => { try { await fn(); await reloadMeta(); toast(msg) } catch (e) { toast((e as Error).message, 'err') } }
  return (
    <Section title="People on this box" actions={isAdmin && <button type="button" className="btn primary sm" onClick={() => setEdit({ role: 'member' })}><Icon name="plus" size={14} />Add account</button>}
      sub="Everyone signs in with their own account and has their own charts, people, journal and connections. Admins manage accounts; they do not see anyone else’s.">
      <div className="set-people">
        {users.map(u => (
          <div key={u.id} className="set-person">
            <Avatar name={u.name} email={u.email} size={34} />
            <div className="grow">
              <div className="set-person-name">{u.name || <span className="muted">—</span>}{u.id === user?.id && <span className="badge">you</span>}{u.role === 'admin' && <span className="badge accent">admin</span>}</div>
              <div className="muted small ellipsis">{u.email}</div>
            </div>
            {isAdmin && <span className="row" style={{ justifyContent: 'flex-end' }}><button type="button" className="btn ghost sm" onClick={() => setEdit(u)}>Edit</button>{u.id !== user?.id && <button type="button" className="btn ghost sm danger-text" onClick={() => setDel(u)}>Remove</button>}</span>}
          </div>
        ))}
      </div>
      {status?.auth_mode !== 'local' && <p className="muted small mt">Auth mode is <code>{status?.auth_mode}</code>: users are created automatically on first visit; passwords are not used here.</p>}
      {edit && <UserModal user={edit} local={status?.auth_mode === 'local'} onClose={() => setEdit(null)} onSaved={() => { setEdit(null); reloadMeta() }} />}
      {del && <Confirm confirmLabel="Remove" title={`Remove ${del.name || del.email}?`} message="Their account is deleted together with their people, journal and readings." onCancel={() => setDel(null)} onConfirm={() => { const d = del; setDel(null); run(() => api.del(`${V1}/users/${d.id}`), 'User removed') }} />}
    </Section>
  )
}

function UserModal({ user, local, onClose, onSaved }: { user: Partial<User>; local: boolean; onClose: () => void; onSaved: () => void }) {
  const { toast } = useToast()
  const [f, setF] = useState({ name: user.name || '', email: user.email || '', password: '', role: user.role || 'member' })
  const [busy, setBusy] = useState(false)
  const save = async (e: React.FormEvent) => {
    e.preventDefault(); setBusy(true)
    try { if (user.id) await api.put(`${V1}/users/${user.id}`, f); else await api.post(`${V1}/users`, f); toast(user.id ? 'User saved' : 'User added'); onSaved() } catch (e) { toast((e as Error).message, 'err') } finally { setBusy(false) }
  }
  return (
    <Modal title={user.id ? 'Edit account' : 'Add an account'} onClose={onClose} size="sm">
      <form onSubmit={save} className="grid" style={{ gap: 14 }}>
        <Field label="Name"><input value={f.name} onChange={e => setF({ ...f, name: e.target.value })} autoFocus required maxLength={60} /></Field>
        <Field label="Email" help="Used to sign in."><input type="email" value={f.email} onChange={e => setF({ ...f, email: e.target.value })} required /></Field>
        {local && <Field label={user.id ? 'New password (leave empty to keep)' : 'Password'} help="At least 8 characters."><input type="password" value={f.password} onChange={e => setF({ ...f, password: e.target.value })} minLength={user.id ? 0 : 8} required={!user.id} autoComplete="new-password" /></Field>}
        <Field group label="Role"><Segmented value={f.role} onChange={v => setF({ ...f, role: v })} options={[{ value: 'member', label: 'Member' }, { value: 'admin', label: 'Admin' }]} /></Field>
        <div className="form-actions"><button type="button" className="btn" onClick={onClose}>Cancel</button><button className="btn primary" disabled={busy}>{user.id ? 'Save' : 'Add account'}</button></div>
      </form>
    </Modal>
  )
}

function Profile() {
  const { user, setUser, status } = useApp()
  const { toast } = useToast()
  const [f, setF] = useState({ name: user?.name || '', email: user?.email || '', current_password: '', new_password: '' })
  const [busy, setBusy] = useState(false)
  const save = async (e: React.FormEvent) => {
    e.preventDefault(); setBusy(true)
    try { setUser(await api.put(`${V1}/auth/me`, f)); setF(x => ({ ...x, current_password: '', new_password: '' })); toast('Profile saved') } catch (e) { toast((e as Error).message, 'err') } finally { setBusy(false) }
  }
  return (
    <form onSubmit={save} className="stack loose">
      <div className="set-profile">
        <Avatar name={user?.name} email={user?.email} size={56} />
        <div className="grow"><h3>{user?.name || user?.email}</h3><p className="muted small">{user?.email} · {user?.role}</p></div>
      </div>
      <Section title="You">
        <div className="stack">
          <Field label="Name"><input value={f.name} onChange={e => setF({ ...f, name: e.target.value })} /></Field>
          <Field label="Email"><input type="email" value={f.email} onChange={e => setF({ ...f, email: e.target.value })} /></Field>
        </div>
      </Section>
      {status?.auth_mode === 'local' && <Section title="Password" sub="Leave empty to keep the current one."><div className="stack">
        <Field label="Current password"><input type="password" value={f.current_password} onChange={e => setF({ ...f, current_password: e.target.value })} autoComplete="current-password" /></Field>
        <Field label="New password" help="At least 8 characters."><input type="password" value={f.new_password} onChange={e => setF({ ...f, new_password: e.target.value })} autoComplete="new-password" /></Field>
      </div></Section>}
      <div className="row"><button className="btn primary" disabled={busy}>Save changes</button></div>
    </form>
  )
}

function Tokens() {
  const { toast } = useToast()
  const { settings } = useApp()
  const { data, reload } = useAsync(() => api.get<APIToken[]>(`${V1}/tokens`), [])
  const [name, setName] = useState('')
  const [fresh, setFresh] = useState('')
  const [del, setDel] = useState<APIToken | null>(null)
  const create = async (e: React.FormEvent) => {
    e.preventDefault()
    try { const r = await api.post<{ token: string }>(`${V1}/tokens`, { name: name || 'token' }); setFresh(r.token); setName(''); reload() } catch (e) { toast((e as Error).message, 'err') }
  }
  const base = settings?.base_url || window.location.origin
  return (
    <div className="stack loose">
      <Section title="Personal tokens" icon="key" sub={<>Use as <code>Authorization: Bearer &lt;token&gt;</code>. A token acts as you: it can read and change everything, feed a Homepage widget (<code>GET /api/v1/dashboard</code>) or log a check-in from a Shortcut. See <code>docs/api.md</code>.</>}>
        <form onSubmit={create} className="row mb"><input value={name} onChange={e => setName(e.target.value)} placeholder="Token name (e.g. Homepage)" /><button className="btn primary"><Icon name="plus" size={15} />Create</button></form>
        {fresh && <div className="set-fresh mb"><div className="row"><Icon name="check-circle" size={16} /><b>Copy it now — it is shown once.</b></div><pre className="mono">{fresh}</pre></div>}
        {data && data.length > 0 ? (
          <table className="table compact">
            <thead><tr><th>Name</th><th>Prefix</th><th>Last used</th><th>Created</th><th /></tr></thead>
            <tbody>{data.map(t => <tr key={t.id}><td>{t.name}</td><td className="mono">{t.prefix}…</td><td className="muted">{t.last_used_at ? fmtDateTime(t.last_used_at) : 'never'}</td><td className="muted">{fmtDateTime(t.created_at)}</td><td className="right"><button type="button" className="btn ghost sm danger-text" onClick={() => setDel(t)}>Revoke</button></td></tr>)}</tbody>
          </table>
        ) : <p className="muted">No tokens yet.</p>}
        {del && <Confirm confirmLabel="Revoke" title={`Revoke “${del.name}”?`} message="Anything using this token stops working immediately." onCancel={() => setDel(null)} onConfirm={async () => { await api.del(`${V1}/tokens/${del.id}`); setDel(null); reload() }} />}
      </Section>
      <Section title="Endpoints" icon="code" sub={<>At <code>{base}/api/v1</code>.</>}>
        <dl className="set-dl">
          <div><dt><code>GET /today</code></dt><dd className="muted">Your day: headline, scores, the Moon and transits (<code>?date=YYYY-MM-DD</code>, <code>?person=ID</code>).</dd></div>
          <div><dt><code>POST /journal</code></dt><dd className="muted"><code>{'{"mood":4,"energy":3,"note":"Good day"}'}</code> — today’s check-in (a Siri Shortcut).</dd></div>
          <div><dt><code>GET /dashboard</code></dt><dd className="muted">The Moon, your day’s headline and scores, what is coming — for Homepage.</dd></div>
          <div><dt><code>GET /sky/now</code></dt><dd className="muted">Where every planet is right now.</dd></div>
          <div><dt><code>GET /day/YYYY-MM-DD</code></dt><dd className="muted">One day — what Pocket Day reads.</dd></div>
          <div><dt><code>GET /overlay?from=&to=</code></dt><dd className="muted">Moon phases, eclipses, retrogrades and your big transits — for Pocket Calendar.</dd></div>
        </dl>
      </Section>
    </div>
  )
}

function Data() {
  const { toast } = useToast()
  const { user } = useApp()
  const [busy, setBusy] = useState(false)
  const importJSON = async (file: File) => {
    setBusy(true)
    try { const r = await api.post<Record<string, number>>(`${V1}/import`, JSON.parse(await file.text())); toast('Imported: ' + (Object.entries(r).map(([k, v]) => `${v} ${k}`).join(', ') || 'nothing new')); window.dispatchEvent(new CustomEvent('phr:changed')) } catch (e) { toast((e as Error).message, 'err') } finally { setBusy(false) }
  }
  return (
    <div className="stack loose">
      {user?.role === 'admin' && <Section title="Database backup" icon="database" sub="The whole box, every account — admins only.">
        <p className="muted">A consistent copy of the SQLite database (<code>VACUUM INTO</code>). Restore by stopping the app and replacing <code>pocket-horoscope.db</code> in the data directory.</p>
        <a className="btn primary" href={`${V1}/backup`}><Icon name="download" size={15} />Download database</a>
      </Section>}
      <Section title="Export & import" icon="archive" sub="Everything of yours as one JSON file (your people with their birth data, your check-ins and the readings you kept), or your people as a CSV that most astrology apps read. Importing the same file twice adds nothing.">
        <div className="row wrap"><a className="btn" href={`${V1}/export.json`}><Icon name="download" size={15} />Export JSON</a><a className="btn" href={`${V1}/export.csv`}><Icon name="download" size={15} />People CSV</a><label className="btn"><Icon name="upload" size={15} />{busy ? 'Importing…' : 'Import JSON…'}<input type="file" accept="application/json,.json" hidden onChange={e => e.target.files?.[0] && importJSON(e.target.files[0])} /></label></div>
      </Section>
    </div>
  )
}

function About() {
  const { data, error } = useAsync(() => api.get<Record<string, unknown>>(`${V1}/settings/system`), [])
  const [keys, setKeys] = useState(false)
  if (error) return <div className="callout danger"><Icon name="alert" size={15} /><span>{error}</span></div>
  if (!data) return <Loading />
  const counts = (data.counts || {}) as Record<string, number>
  const core = (data.core || {}) as { configured?: boolean; registered?: boolean; url?: string; error?: string }
  const tiles: [string, number | undefined, string][] = [['People', counts.people, 'users'], ['Check-ins', counts.journal, 'journal'], ['Readings', counts.readings, 'book-open'], ['Tokens', counts.api_tokens, 'key']]
  const rows: [string, React.ReactNode][] = [
    ['Version', <code key="v">{String(data.version)}</code>],
    ['Uptime', String(data.uptime)],
    ['Runtime', `${String(data.go)} · ${String(data.os)}/${String(data.arch)}`],
    ['Time zone', String(data.timezone)],
    ['Sign-in', String(data.auth_mode)],
    ['Pocket Core', core.configured ? (core.registered ? <span key="c" className="ok">connected · {core.url}</span> : <span key="c" className="bad">{core.error || 'not registered yet'}</span>) : <span key="c" className="muted">not configured</span>],
    ['Database', <span key="db"><code>{String(data.db_path)}</code> · {fmtBytes(Number(data.db_bytes))}</span>],
    ['Accounts', `${counts.users ?? 0} on this box`],
  ]
  return (
    <div className="stack loose">
      <div className="set-about-hero">
        <img src="/favicon.svg" alt="" width={48} height={48} />
        <div>
          <h3>Pocket Horoscope</h3>
          <p className="muted small">Astrology worked out from the real sky: birth charts for you and the people you care about, a personal daily horoscope, compatibility, the Moon calendar with retrogrades and eclipses, the transits coming up and a mood journal next to the Moon. One binary, one SQLite file, one volume.</p>
        </div>
      </div>
      <div className="set-stats">
        {tiles.map(([label, n, icon]) => (
          <div key={label} className="set-stat"><Icon name={icon} size={16} /><span className="set-stat-n tnum">{(n ?? 0).toLocaleString()}</span><span className="set-stat-l">{label}</span></div>
        ))}
      </div>
      <Section title="This server">
        <dl className="set-dl">
          {rows.map(([k, v]) => <div key={k}><dt>{k}</dt><dd>{v}</dd></div>)}
        </dl>
      </Section>
      <div className="row wrap">
        <button type="button" className="btn" onClick={() => setKeys(true)}><Icon name="keyboard" size={15} />Keyboard shortcuts</button>
        <Link className="btn" to="/sky/moon"><Icon name="moon" size={15} />Moon calendar</Link>
        <a className="btn" href="https://github.com/abjelosevic88/pocket-homelab-horoscope" target="_blank" rel="noreferrer"><Icon name="external" size={15} />Project on GitHub</a>
      </div>
      <p className="muted small">Open source (MIT). Part of the Pocket apps: Pocket Core connects them — Day, Calendar, Contacts, Notes, Health and more. Planet positions are worked out from VSOP87 and ELP (within a few arc-seconds of NASA JPL). Astrology is for reflection, not a forecast of events or advice on health, money or law.</p>
      {keys && <ShortcutsSheet onClose={() => setKeys(false)} />}
    </div>
  )
}
