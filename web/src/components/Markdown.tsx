// A small, safe markdown renderer for readings: ### headings, paragraphs,
// - / 1. lists, **bold**, *italic*, [links](https://…) and --- rules. No HTML passes through.
import React from 'react'

function inline(text: string, key: string): React.ReactNode[] {
  const out: React.ReactNode[] = []
  const re = /(\*\*([^*]+)\*\*|\*([^*]+)\*|_([^_]+)_|\[([^\]]+)\]\((https?:\/\/[^)\s]+|\/[^)\s]*)\))/g
  let last = 0
  let m: RegExpExecArray | null
  let i = 0
  while ((m = re.exec(text))) {
    if (m.index > last) out.push(text.slice(last, m.index))
    const k = `${key}-${i++}`
    if (m[2]) out.push(<strong key={k}>{m[2]}</strong>)
    else if (m[3] || m[4]) out.push(<em key={k}>{m[3] || m[4]}</em>)
    else if (m[5]) out.push(<a key={k} href={m[6]} target={m[6].startsWith('/') ? undefined : '_blank'} rel="noreferrer">{m[5]}</a>)
    last = m.index + m[0].length
  }
  if (last < text.length) out.push(text.slice(last))
  return out
}

export function Markdown({ text, className = '' }: { text: string; className?: string }) {
  const lines = (text || '').replace(/\r/g, '').split('\n')
  const blocks: React.ReactNode[] = []
  let para: string[] = []
  let list: { ordered: boolean; items: string[] } | null = null
  const flush = () => {
    if (para.length) { blocks.push(<p key={`p${blocks.length}`}>{inline(para.join(' '), `p${blocks.length}`)}</p>); para = [] }
    if (list) {
      const items = list.items.map((it, i) => <li key={i}>{inline(it, `l${blocks.length}-${i}`)}</li>)
      blocks.push(list.ordered ? <ol key={`o${blocks.length}`}>{items}</ol> : <ul key={`u${blocks.length}`}>{items}</ul>)
      list = null
    }
  }
  for (const raw of lines) {
    const line = raw.trimEnd()
    const h = /^(#{1,4})\s+(.*)$/.exec(line)
    const ul = /^\s*[-*•]\s+(.*)$/.exec(line)
    const ol = /^\s*\d+[.)]\s+(.*)$/.exec(line)
    if (!line.trim()) { flush(); continue }
    if (/^\s*(-{3,}|\*{3,})\s*$/.test(line)) { flush(); blocks.push(<hr key={`hr${blocks.length}`} />); continue }
    if (h) {
      flush()
      const lvl = Math.min(4, Math.max(3, h[1].length + 1))
      blocks.push(React.createElement(`h${lvl}`, { key: `h${blocks.length}` }, inline(h[2], `h${blocks.length}`)))
      continue
    }
    if (ul || ol) {
      if (para.length) { const p = para; para = []; blocks.push(<p key={`p${blocks.length}`}>{inline(p.join(' '), `p${blocks.length}`)}</p>) }
      const ordered = !!ol
      if (list && list.ordered !== ordered) flush()
      if (!list) list = { ordered, items: [] }
      list.items.push((ul || ol)![1])
      continue
    }
    if (list) flush()
    para.push(line.trim())
  }
  flush()
  return <div className={`md ${className}`}>{blocks}</div>
}
