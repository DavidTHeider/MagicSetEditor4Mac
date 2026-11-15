// **Ein Anker im Portal ist beim ersten Lauf noch leer** — N19 aus dem
// Nachlauf über Phase 6.3 (#1664), verlagert in Phase 6.11 (#1672).
//
// **Der Verdacht**: Mantines Portal legt seinen Behälter in einem
// Layout-Effekt an und zeichnet die Kinder erst danach. Merkte sich
// `useRueckkehr` den Knoten beim ersten Effekt als `null`, hielte das
// Aufräumen beim Verbergen jede Schublade für verschwunden und gäbe den
// Fokus zurück, auch wenn er längst in der Leiste stünde.
//
// **Widerlegt am 2026-10-04** (Mantine 9.6.1, React 19): der Knoten ist
// beim ersten Effekt gesetzt. Die Gegenprobe — `knoten` im Haken auf `null`
// gezwungen — macht diesen Test rot. Er bleibt als Wächter für die nächste
// Fassung von Mantine.
import { render, screen } from '@testing-library/react'
import { Activity } from 'react'
import { afterEach, expect, it } from 'vitest'

import { Schublade } from '../bausteine/Schublade'
import { Sprachraum } from '../sprache/kontext'
import { effekteNachholen } from '../test/effekte'

afterEach(() => {
  document.body.replaceChildren()
})

function Seite({ sichtbar, offen }: { sichtbar: boolean; offen: boolean }) {
  return (
    <Sprachraum sprache="de">
      <button type="button">Leiste</button>
      <Activity mode={sichtbar ? 'visible' : 'hidden'}>
        <button type="button">Auslöser</button>
        {offen && (
          <Schublade
            titel="Eigenschaften"
            schliessenBezeichnung="Schließen"
            schliessen={() => undefined}
          >
            <p>Inhalt</p>
          </Schublade>
        )}
      </Activity>
    </Sprachraum>
  )
}

it('lässt den Fokus in der Leiste, wenn die Seite mit offener Schublade verborgen wird', async () => {
  // **Der Auslöser hat den Fokus, bevor sie aufgeht** — sonst merkt sich
  // der Haken kein Rückkehrziel, und ein falsches Zurückgeben fiele nicht
  // auf (die erste Fassung dieses Tests blieb so bei der Gegenprobe grün).
  const { rerender } = render(<Seite sichtbar offen={false} />)
  screen.getByRole('button', { name: 'Auslöser' }).focus()
  rerender(<Seite sichtbar offen />)
  await screen.findByRole('dialog', { name: 'Eigenschaften' })
  await effekteNachholen()

  const leiste = screen.getByRole('button', { name: 'Leiste' })
  leiste.focus()
  expect(leiste).toHaveFocus()

  rerender(<Seite sichtbar={false} offen />)
  await effekteNachholen()

  expect(leiste).toHaveFocus()
})
