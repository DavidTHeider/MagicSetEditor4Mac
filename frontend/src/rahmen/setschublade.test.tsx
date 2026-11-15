// **Die Laschen einer Schublade** — UX-08.AK-01
// (#1680).
//
// **Was hier steht und was in `setschubladen.test.tsx`**: dort die Naht durch
// die ganze Anwendung — Knopf, Seitenwechsel, Schublade über der Fläche; hier
// die Laschenleiste der Schublade für sich, mit demselben Tastenmuster wie
// die Leiste des Arbeitsplatzes.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Setschublade } from './Setschublade'

import type { Seitenschluessel, Setschluessel } from '../navigation'
import type { Schublade } from '../schubladen'

const texte = kataloge.de

// eigenschaften ist eine Schublade mit vier Laschen — genug, dass ein Pfeil
// nach links am Anfang wirklich umläuft und nicht zufällig beim Nachbarn
// landet. **Hier aufgeschrieben und nicht aus `schubladen.ts` gelesen**: der
// Test gilt dem Bauteil, und die Einteilung hält `laschengruppen.test.ts`.
const eigenschaften: Schublade = {
  schluessel: 'eigenschaften',
  laschen: ['stil', 'setangaben', 'symbol', 'schluesselwoerter'],
}

function zeige(
  aktuell: Setschluessel,
  wechseln: (seite: Seitenschluessel) => void,
) {
  render(
    <Sprachraum sprache="de">
      <Setschublade
        schublade={eigenschaften}
        aktuell={aktuell}
        wechseln={wechseln}
        schliessen={vi.fn()}
      >
        <p>Inhalt</p>
      </Setschublade>
    </Sprachraum>,
  )
}

describe('Setschublade', () => {
  it('führt ihre Laschen als benannte Leiste, die gewählte im Tabweg', () => {
    zeige('setangaben', vi.fn())

    const leiste = screen.getByRole('tablist', {
      name: texte.allgemein.arbeitsplatz.schubladen.eigenschaften.laschen,
    })
    expect(leiste).toBeInTheDocument()

    const gewaehlt = screen.getByRole('tab', { selected: true })
    expect(gewaehlt).toHaveTextContent(texte.seiten.setangaben.titel)
    expect(gewaehlt).toHaveAttribute('tabindex', '0')

    // **Feld und Lasche zeigen aufeinander**, wie in der Hauptleiste
    // (#1553).
    const feld = screen.getByRole('tabpanel')
    expect(gewaehlt).toHaveAttribute('aria-controls', feld.id)
    expect(feld).toHaveAttribute('aria-labelledby', gewaehlt.id)
    expect(feld).toHaveTextContent('Inhalt')

    // **Und der Satz darunter sagt, warum es hier steht** — wie im Entwurf.
    expect(
      screen.getByText(
        texte.allgemein.arbeitsplatz.schubladen.eigenschaften.hinweis,
      ),
    ).toBeInTheDocument()
  })

  it('wechselt über eine Lasche', async () => {
    const gewechselt = vi.fn()
    zeige('stil', gewechselt)

    await userEvent.click(
      screen.getByRole('tab', { name: texte.seiten.symbol.titel }),
    )

    expect(gewechselt).toHaveBeenCalledWith('symbol')
  })

  // **Die Pfeile wandern, und der Fokus wandert mit** — ohne ihn zählte die
  // nächste Taste wieder von der alten Lasche, und man käme nie weiter als
  // eins.
  it('wandert mit den Pfeiltasten und nimmt den Fokus mit', async () => {
    const gewechselt = vi.fn()
    zeige('stil', gewechselt)

    screen.getByRole('tab', { name: texte.seiten.stil.titel }).focus()

    await userEvent.keyboard('{ArrowRight}')
    expect(gewechselt).toHaveBeenLastCalledWith('setangaben')
    expect(
      screen.getByRole('tab', { name: texte.seiten.setangaben.titel }),
    ).toHaveFocus()
  })

  // **Zyklisch** — von der ersten links heraus steht man auf der letzten.
  it('läuft am Anfang auf die letzte Lasche um', async () => {
    const gewechselt = vi.fn()
    zeige('stil', gewechselt)

    screen.getByRole('tab', { name: texte.seiten.stil.titel }).focus()

    await userEvent.keyboard('{ArrowLeft}')
    expect(gewechselt).toHaveBeenLastCalledWith('schluesselwoerter')
  })

  // **Eine andere Taste tut nichts** — Escape muss bis zur Schublade
  // durchkommen, und ⌘K bis zur Palette.
  it('lässt andere Tasten durch', async () => {
    const gewechselt = vi.fn()
    zeige('stil', gewechselt)

    screen.getByRole('tab', { name: texte.seiten.stil.titel }).focus()

    await userEvent.keyboard('{ArrowDown}')

    expect(gewechselt).not.toHaveBeenCalled()
  })
})
