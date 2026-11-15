// **Die Kennzahlen im Fenster** — Werkstück U10 von Phase 5.1
// (#1421).
import { render, screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Deckkennzahlen } from './Deckkennzahlen'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

import type { Statistikposten } from './deckstatistik'

const texte = kataloge.de.seiten.decks

function zeige(posten: Statistikposten[]) {
  render(
    <Sprachraum sprache="de">
      <Deckkennzahlen posten={posten} />
    </Sprachraum>,
  )
}

describe('die Kennzahlen eines Decks', () => {
  it('zeigt Typen, Länder, Durchschnitt und Farbquellen', () => {
    zeige([
      { typzeile: 'Creature — Elf Druid', kosten: '{1}{G}', menge: 4 },
      { typzeile: 'Instant', kosten: '{R}', menge: 2 },
      { typzeile: 'Basic Land — Forest', kosten: '', menge: 9 },
      { typzeile: 'Land — Forest Island', kosten: '', menge: 3 },
    ])

    const typen = screen.getByLabelText(texte.typen)
    expect(within(typen).getByText('Creature')).toBeInTheDocument()
    expect(within(typen).getByText('4')).toBeInTheDocument()

    // **Zwölf Länder** (9 + 3) und **(4×2 + 2×1) / 6 = 1,67** im Schnitt.
    //
    // **Im Kasten der Kennzahlen gesucht**: die Zwölf steht auch bei den
    // Quellen, und `getByText` über die ganze Seite fände beide.
    const kennzahlen = screen.getByLabelText(texte.kennzahlen)
    expect(within(kennzahlen).getByText('12')).toBeInTheDocument()
    expect(within(kennzahlen).getByText('1.67')).toBeInTheDocument()

    // **Grün aus zwölf, Blau aus drei** — das Zweifarbenland zählt in beiden.
    const quellen = screen.getByLabelText(texte.quellen)
    expect(within(quellen).getByText(texte.farbeGruen)).toBeInTheDocument()
    expect(within(quellen).getByText('12')).toBeInTheDocument()
    expect(within(quellen).getByText('3')).toBeInTheDocument()
  })

  it('sagt es mit einem Strich, wenn es keinen Durchschnitt gibt', () => {
    zeige([{ typzeile: 'Basic Land — Forest', kosten: '', menge: 20 }])

    expect(screen.getByText(texte.keineZahl)).toBeInTheDocument()
  })

  it('bleibt ganz aus, wenn keine Karte eine Typzeile hat', () => {
    zeige([{ typzeile: '', kosten: '{G}', menge: 4 }])

    expect(screen.queryByLabelText(texte.typen)).not.toBeInTheDocument()
  })

  it('nimmt den nachgelesenen Manawert einer zweiseitigen Karte', () => {
    // **Leere Kosten, Manawert 3** — gerechnet wäre es null.
    zeige([
      { typzeile: 'Creature — Werewolf', kosten: '', menge: 1, manawert: 3 },
    ])

    expect(screen.getByText('3.00')).toBeInTheDocument()
  })

  it('zählt ein Land ohne Grundtyp nicht als Quelle', () => {
    // **Ein Land mit Regeltext** — es zu deuten wäre geraten (§10).
    zeige([
      { typzeile: 'Land', kosten: '', menge: 4 },
      { typzeile: 'Instant', kosten: '{R}', menge: 1 },
    ])

    expect(screen.queryByLabelText(texte.quellen)).not.toBeInTheDocument()
  })
})
