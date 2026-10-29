import { Modal } from './ui'

const GROUPS: { title: string; keys: [string[], string][] }[] = [
  { title: 'Go to', keys: [[['t'], 'Today'], [['p'], 'People & charts'], [['m'], 'Compatibility'], [['u'], 'Coming up'], [['s'], 'Sky & Moon'], [['j'], 'Journal'], [['r'], 'Readings'], [[','], 'Settings'], [['/'], 'Search people & commands'], [['⌘', 'K'], 'Search people & commands']] },
  { title: 'Add', keys: [[['n'], 'Add a person'], [['c'], "Today's check-in"], [['a'], 'Ask about your chart']] },
  { title: 'Today', keys: [[['←', '→'], 'The day before, the day after']] },
  { title: 'Window', keys: [[['['], 'Show or hide the sidebar'], [['esc'], 'Close a dialog or menu'], [['?'], 'This sheet']] },
]

export function ShortcutsSheet({ onClose }: { onClose: () => void }) {
  return (
    <Modal title="Keyboard shortcuts" onClose={onClose} size="lg">
      <div className="shortcuts">
        {GROUPS.map(g => (
          <section key={g.title}>
            <h4 className="sc-h">{g.title}</h4>
            {g.keys.map(([keys, label]) => (
              <div key={label + keys.join('')} className="sc-row"><span>{label}</span><span className="sc-keys">{keys.map(k => <kbd key={k}>{k}</kbd>)}</span></div>
            ))}
          </section>
        ))}
      </div>
    </Modal>
  )
}
