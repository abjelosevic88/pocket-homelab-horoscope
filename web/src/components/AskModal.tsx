// Ask a question about your chart (or someone else's): AI reads the chart and
// today's sky and answers; the answer is kept under Readings.
import { useState } from 'react'
import { Link } from 'react-router-dom'
import { api, V1 } from '../lib/api'
import type { Reading } from '../lib/types'
import { PersonPicker, usePeople } from './CheckIn'
import { Icon } from './icons'
import { Markdown } from './Markdown'
import { Field, Modal, Spinner, errMsg } from './ui'

const SUGGESTIONS = [
  'What are my strengths at work?',
  'How can I handle stress better this month?',
  'What does my Moon sign say about what I need?',
  'Is this a good week to start something new?',
  'How do I show love, and what do I need back?',
]

interface AskResult { ok: boolean; text?: string; error?: string; reading?: Reading }

export function AskModal({ onClose }: { onClose: () => void }) {
  const people = usePeople()
  const [person, setPerson] = useState('')
  const [q, setQ] = useState('')
  const [busy, setBusy] = useState(false)
  const [res, setRes] = useState<AskResult | null>(null)

  const ask = async (question = q) => {
    const text = question.trim()
    if (!text || busy) return
    setQ(text); setBusy(true); setRes(null)
    try {
      setRes(await api.post<AskResult>(`${V1}/ai/ask`, { question: text, person_id: person ? Number(person) : 0 }))
    } catch (e) { setRes({ ok: false, error: errMsg(e) }) } finally { setBusy(false) }
  }
  const aiMissing = !!res && !res.ok && /AI|set up|Settings/i.test(res.error || '')

  return (
    <Modal title="Ask about your chart" onClose={onClose} size="lg" footer={<>
      <span className="grow" />
      <button type="button" className="btn" onClick={onClose}>Close</button>
      <button type="button" className="btn primary" onClick={() => ask()} disabled={busy || !q.trim()}>{busy ? 'Thinking…' : 'Ask'}</button>
    </>}>
      <div className="stack">
        <div className="ask-row">
          <Field label="Your question" className="grow">
            <textarea autoFocus rows={3} value={q} onChange={e => setQ(e.target.value)} placeholder="Ask anything about yourself, your relationships or the days ahead…"
              onKeyDown={e => { if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) { e.preventDefault(); ask() } }} />
          </Field>
          {people && people.length > 1 && <Field label="About" className="ask-person"><PersonPicker people={people} value={person} onChange={setPerson} /></Field>}
        </div>
        {!res && !busy && (
          <div className="chips" aria-label="Suggested questions">
            {SUGGESTIONS.map(s => <button key={s} type="button" className="chip" onClick={() => ask(s)}>{s}</button>)}
          </div>
        )}
        {busy && <div className="ask-wait muted"><Spinner /> Reading the chart and today's sky…</div>}
        {res?.ok && (
          <div className="ask-answer">
            <Markdown text={res.text || ''} />
            {res.reading && <Link className="link-btn small" to={`/readings/${res.reading.id}`} onClick={onClose}><Icon name="bookmark" size={13} /> Kept in Readings</Link>}
          </div>
        )}
        {res && !res.ok && (aiMissing
          ? <div className="callout"><Icon name="info" size={15} /><div><span className="muted">{res.error}</span> <Link to="/settings/ai" onClick={onClose}>Set up AI in Settings → AI</Link>.</div></div>
          : <div className="callout danger"><Icon name="alert" size={15} /><div>{res.error}</div></div>)}
      </div>
    </Modal>
  )
}
