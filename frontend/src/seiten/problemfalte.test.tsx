// **Die Problemliste zeigt drei und den Rest stückweise** — UXR-06
// (#1914).
//
// # Der Befund
//
// Am Probebestand meldet der Rasterer nichts, und die Messung sah eine leere
// Liste. Ein Rechner ohne die Schriften des Magic-Packs sieht anderes: der
// Rasterer meldet je **Ebene**, deren Schrift fehlt — „Ebene 3 (rule text):
// die Schrift MPlantin ist nicht zu finden" —, und `magic-m15` setzt Name,
// Typ, Regeltext, Stärke, Zeichner und Copyright je in eine eigene. Jede
// Zeile trägt „Zum Feld …", und alle stünden über dem Formular, oben im
// Inspector und damit im Bild: ein Dutzend Knöpfe mehr (R2 aus
// `docs/UIREVIEW.md`), und das erste Feld rutschte unter die Falz.
//
// # Was hier gehalten wird
//
// - **Drei vorn, „3 von 12" und „Weitere 9 zeigen"** — die Falte der Listen.
// - **Der Feldfehler steht vorn** — `problemeAus` stellt ihn voran, und die
//   Falte nimmt die Reihenfolge, wie sie ist.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Problemliste, problemeAus } from './Problemliste'

const worte = kataloge.de.bausteine.listenansicht

const felder = [
  'name',
  'casting cost',
  'super type',
  'sub type',
  'rule text',
  'flavor text',
  'power',
  'toughness',
  'illustrator',
  'copyright',
  'rarity',
]

describe('Die Problemliste mit vielen Befunden', () => {
  it('zeigt drei, den Feldfehler zuerst, und den Rest auf Wunsch', async () => {
    const probleme = problemeAus(
      { feld: 'name', meldung: 'Der Name fehlt.' },
      felder.map(
        (feld, stelle) =>
          `Ebene ${String(stelle + 2)} (${feld}): die Schrift MPlantin ist nicht zu finden`,
      ),
    )
    render(
      <Sprachraum sprache="de">
        <Problemliste probleme={probleme} zumFeld={vi.fn()} />
      </Sprachraum>,
    )

    const liste = screen.getByRole('region', { name: 'Probleme' })
    expect(within(liste).getAllByRole('listitem')).toHaveLength(3)
    expect(within(liste).getAllByRole('listitem')[0]).toHaveTextContent(
      'Der Name fehlt.',
    )
    expect(
      within(liste).getByText(
        worte.zahl.replace('{gezeigt}', '3').replace('{gesamt}', '12'),
      ),
    ).toBeInTheDocument()

    await userEvent.click(
      within(liste).getByRole('button', {
        name: worte.weitere.replace('{anzahl}', '9'),
      }),
    )
    expect(within(liste).getAllByRole('listitem')).toHaveLength(12)
  })
})
